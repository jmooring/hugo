// Copyright 2025 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diagrams

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/gohugoio/hugo/common/hashing"
	"github.com/gohugoio/hugo/common/hugio"
	"github.com/gohugoio/hugo/tpl/diagrams/diagrams_config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/tdewolff/minify/v2/minify"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

const cacheKeyPrefix = "diagrams/d2/"

// A d2Diagram extends the functionality of d2SVG by implementing the
// SVGDiagram interface.
type d2Diagram struct {
	d *d2SVG
}

// d2SVG represents an SVG diagram created by D2, containing its rendered
// content and metadata. D2 generates an `svg` element that is subsequently
// wrapped within another `svg` element. This struct stores the content of the
// inner `svg` element.
type d2SVG struct {
	Body    string `xml:",innerxml"`
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	ViewBox string `xml:"viewBox,attr"`
}

// Wrapped returns the inner `svg` element wrapped within an outer `svg`
// element, effectively reconstructing the original SVG diagram created by D2.
func (d d2Diagram) Wrapped() template.HTML {
	return template.HTML(d.d.String())
}

// Inner returns the inner `svg` element.
func (d d2Diagram) Inner() template.HTML {
	return template.HTML(d.d.Body)
}

// Width returns the width of the outer `svg` element, which may differ from
// the inner `svg` element's width if scaled during rendering.
func (d d2Diagram) Width() int {
	return d.d.Width
}

// Height returns the height of the outer `svg` element, which may differ from
// the inner `svg` element's height if scaled during rendering.
func (d d2Diagram) Height() int {
	return d.d.Height
}

// ViewBox returns the ViewBox of the outer `svg` element. The ViewBox
// coordinates are not affected by rendering scaling.
func (d d2Diagram) ViewBox() string {
	return d.d.ViewBox
}

// String returns a string representation of the D2 diagram, consisting of the
// inner `svg` element wrapped within an outer `svg` element.
func (d d2SVG) String() string {
	return fmt.Sprintf(`<svg xmlns=%q xmlns:xlink=%q iewBox=%q width="%d" height="%d">%s</svg>`,
		"http://www.w3.org/2000/svg",
		"http://www.w3.org/1999/xlink",
		d.ViewBox,
		d.Width,
		d.Height,
		d.Body,
	)
}

type d2Options struct {
	// The D2 theme to use if the system is in dark mode. This value is
	// case-insensitive. See https://d2lang.com/tour/themes.
	DarkTheme string

	// The layout engine to use when automatically arranging diagram elements.
	// See https://d2lang.com/tour/layouts.
	LayoutEngine string

	// The D2 theme to use if the system is in light mode or has no preference.
	// This value is case-insensitive. See https://d2lang.com/tour/themes.
	LightTheme string

	// Whether to minify the SVG element.
	Minify bool

	// The number of pixels with which to pad each side of the diagram. This
	// value must be within the bounds of 0 and 1000, inclusive.
	Padding uint16

	// How much to reduce or enlarge the diagram. Values less than 1 reduce the
	// diagram, while values greater than 1 enlarge the diagram. This value
	// must be greater than 0 and less than or equal to 100.
	Scale float32

	// Whether to render the diagram as if sketched by hand.
	Sketch bool
}

// D2 returns an SVG diagram created from the given D2 markup and options.
func (ns *Namespace) D2(args ...any) (SVGDiagram, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errors.New("requires 1 or 2 arguments")
	}

	// Get and validate the D2 markup.
	markup, err := cast.ToStringE(args[0])
	if err != nil {
		return nil, errors.New("first argument must be a string")
	}
	if markup == "" {
		return nil, errors.New("invalid markup (empty string)")
	}

	opts := &d2Options{}

	// Get parameters from site configuration.
	c := ns.deps.Conf.GetConfigSection(name).(diagrams_config.Config).D2
	err = mapstructure.WeakDecode(c, &opts)
	if err != nil {
		return nil, err
	}

	// Merge the given options, if any.
	if len(args) == 2 {
		err := mapstructure.WeakDecode(args[1], &opts)
		if err != nil {
			return nil, err
		}
	}

	err = validateOptions(opts)
	if err != nil {
		return nil, err
	}

	d2SVG, err := ns.getOrCreateD2SVG(markup, opts)
	if err != nil {
		return nil, err
	}

	return d2Diagram{
		d: d2SVG,
	}, nil
}

// getOrCreateD2SVG gets or creates a d2SVG from the given markup and options.
// It first checks the dynamic cache for a matching key. If not found, it
// checks the file cache. If the diagram is not found in either cache, it
// creates a new d2SVG from the given markup and options, caching the result.
func (ns *Namespace) getOrCreateD2SVG(markup string, opts *d2Options) (*d2SVG, error) {
	s := hashing.HashString(markup, opts)
	key := cacheKeyPrefix + s[:2] + "/" + s[2:]

	b, err := ns.cacheD2.GetOrCreate(key, func(string) ([]byte, error) {
		fileCache := ns.deps.ResourceSpec.FileCaches.MiscCache()

		_, r, err := fileCache.GetOrCreate(key, func() (io.ReadCloser, error) {
			d2SVG, err := createD2SVG(markup, opts)
			if err != nil {
				return nil, err
			}

			// Encode the d2SVG struct to a gob and then cache it.
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err = enc.Encode(d2SVG)
			if err != nil {
				return nil, err
			}

			return hugio.NewReadSeekerNoOpCloserFromBytes(buf.Bytes()), nil
		})
		if err != nil {
			return nil, err
		}

		defer r.Close()

		return io.ReadAll(r)
	})
	if err != nil {
		return nil, err
	}

	// Decode the gob to a d2SVG struct.
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	var d2SVG d2SVG
	err = dec.Decode(&d2SVG)
	if err != nil {
		return nil, err
	}

	return &d2SVG, nil
}

// getOrCreateD2SVG gets or creates a d2SVG from the given markup and options.
func createD2SVG(markup string, opts *d2Options) (*d2SVG, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, err
	}

	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		switch strings.ToLower(opts.LayoutEngine) {
		case "dagre":
			return d2dagrelayout.DefaultLayout, nil
		case "elk":
			return d2elklayout.DefaultLayout, nil
		default:
			return nil, errors.New("layout engine must be elk or dagre")
		}
	}

	lightThemeID, err := getThemeID(opts.LightTheme)
	if err != nil {
		return nil, err
	}

	darkThemeID, err := getThemeID(opts.DarkTheme)
	if err != nil {
		return nil, err
	}

	renderOpts := &d2svg.RenderOpts{
		ThemeID:     &lightThemeID,
		DarkThemeID: &darkThemeID,
		Pad:         go2.Pointer(int64(opts.Padding)),
		Scale:       go2.Pointer(float64(opts.Scale)),
		Sketch:      go2.Pointer(bool(opts.Sketch)),
		NoXMLTag:    go2.Pointer(true),
	}

	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}

	ctx := log.WithDefault(context.Background())

	diagram, _, err := d2lib.Compile(ctx, markup, compileOpts, renderOpts)
	if err != nil {
		return nil, err
	}

	svgBytes, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, err
	}

	svgBytes = sanitizeSVG(svgBytes)

	if opts.Minify {
		svgString, err := minify.SVG(string(svgBytes))
		if err != nil {
			return nil, err
		}
		svgBytes = []byte(svgString)
	}

	// D2 produces `svg` output where the content is wrapped within an
	// additional `svg` element. In the above, `svgBytes` is a sanitized and
	// minified byte slice of the double-wrapped SVG diagram rendered by D2.
	// We need to:
	//
	//  - Extract metadata (height, width, viewBox) from the outer element
	//  - Extract the inner `svg` element
	//  - Discard the outer wrapper
	//
	// The xml.Unmarshal function handles all of this for us provided we have
	// properly tagged the fields of the d2SVG struct. We cache the resulting
	// d2SVG, so we only need to do this when the cache is cold or when a
	// diagram is changed.
	d2SVG := &d2SVG{}
	err = xml.Unmarshal(svgBytes, &d2SVG)
	if err != nil {
		return d2SVG, err
	}

	return d2SVG, nil
}

// sanitizeSVG removes an attribute that triggers a validation error.
// See https://github.com/terrastruct/d2/issues/2273
func sanitizeSVG(svgBytes []byte) []byte {
	return bytes.Replace(svgBytes, []byte(` id="d2-svg"`), []byte(""), 1)
}

// getThemeID returns the theme ID corresponding to the given theme name. The
// lookup is case-insensitive.
func getThemeID(themeName string) (int64, error) {
	if themeName == "" {
		return 0, errors.New("cannot resolve an empty string to a theme ID")
	}
	for _, theme := range d2themescatalog.LightCatalog {
		if strings.EqualFold(theme.Name, themeName) {
			return theme.ID, nil
		}
	}
	for _, theme := range d2themescatalog.DarkCatalog {
		if strings.EqualFold(theme.Name, themeName) {
			return theme.ID, nil
		}
	}

	return 0, fmt.Errorf("theme does not exist: %s", themeName)
}

// validateOptions validates the options used to create D2 diagrams.
func validateOptions(opts *d2Options) error {
	if opts.DarkTheme == "" {
		return errors.New("invalid dark theme (empty string)")
	}
	if opts.LayoutEngine == "" {
		return errors.New("invalid layout engine (empty string)")
	}
	if opts.LayoutEngine != "dagre" && opts.LayoutEngine != "elk" {
		return errors.New("layout engine must be elk or dagre")
	}
	if opts.LightTheme == "" {
		return errors.New("invalid light theme (empty string)")
	}
	if opts.Padding > 1000 {
		return errors.New("padding must be an integer between 0 and 1000 inclusive")
	}
	if opts.Scale <= 0 || opts.Scale > 100 {
		return errors.New("scale must be greater than 0 and less than or equal to 100")
	}

	return nil
}
