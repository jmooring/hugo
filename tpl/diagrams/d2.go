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
	"github.com/gohugoio/hugo/markup/markup_config"
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

// d2Diagram implements the SVGDiagram interface for D2 diagrams.
type d2Diagram struct {
	d *d2SVG
}

// d2SVG represents a D2 SVG diagram. D2 generates an svg element that is
// subsequently wrapped within another svg element. This struct stores the
// inner svg element and the metadata from the outer svg element.
type d2SVG struct {
	// The inner svg element.
	Body string `xml:",innerxml"`

	// The width attribute of the outer svg element.
	Width int `xml:"width,attr"`

	// The height attribute of the outer svg element.
	Height int `xml:"height,attr"`

	// The viewBox attribute of the outer svg element.
	ViewBox string `xml:"viewBox,attr"`

	// The preserveAspectRatio attribute of the outer svg element.
	PreserveAspectRatio string `xml:"preserveAspectRatio,attr"`
}

// Wrapped returns the inner svg element wrapped within an outer svg element,
// effectively reconstructing the original SVG diagram created by D2.
func (d d2Diagram) Wrapped() template.HTML {
	return template.HTML(d.d.String())
}

// Inner returns the inner svg element.
func (d d2Diagram) Inner() template.HTML {
	return template.HTML(d.d.Body)
}

// Width returns the width attribute of the outer svg element, which may differ
// from the inner svg element's width if scaled during rendering.
func (d d2Diagram) Width() int {
	return d.d.Width
}

// Height returns the height attribute of the outer svg element, which may
// differ from the inner svg element's height if scaled during rendering.
func (d d2Diagram) Height() int {
	return d.d.Height
}

// ViewBox returns the viewBox attribute of the outer svg element. The viewBox
// coordinates are not affected by scaling.
func (d d2Diagram) ViewBox() string {
	return d.d.ViewBox
}

// PreserveAspectRatio returns the preserveAspectRatio attribute of the outer
// svg element.
func (d d2Diagram) PreserveAspectRatio() string {
	return d.d.PreserveAspectRatio
}

// String returns a string representation of the D2 diagram, consisting of the
// inner svg element wrapped within an outer svg element.
func (d d2SVG) String() string {
	return fmt.Sprintf(`<svg xmlns=%q xmlns:xlink=%q viewBox=%q width="%d" height="%d" preserveAspectRatio=%q>%s</svg>`,
		"http://www.w3.org/2000/svg",
		"http://www.w3.org/1999/xlink",
		d.ViewBox,
		d.Width,
		d.Height,
		d.PreserveAspectRatio,
		d.Body,
	)
}

type d2Options struct {
	// Whether to center the diagram within the viewport, applicable only when
	// the viewport's aspect ratio is different than that of the SVG viewBox
	// attribute. When true, sets the preserveAspectRatio attribute to xMidYMid
	// meet. When false, sets the preserveAspectRatio attribute to xMinYMin
	// meet.
	Center bool

	// The D2 theme to use if the system is in dark mode. This value is
	// case-insensitive. See https://d2lang.com/tour/themes.
	DarkTheme string

	// The D2 layout engine to use when automatically arranging diagram
	// elements. See https://d2lang.com/tour/layouts.
	LayoutEngine string

	// The D2 theme to use if the system is in light mode or has no preference.
	// This value is case-insensitive. See https://d2lang.com/tour/themes.
	LightTheme string

	// Whether to minify the SVG markup.
	Minify bool

	// The number of pixels with which to pad each side of the diagram. This
	// value must be within the bounds of 0 and 1000, inclusive.
	Padding uint16

	// A salt value used to generate a unique ID, preventing conflicts when
	// embedding multiple identical diagrams in the same HTML document.
	Salt string

	// How much to reduce or enlarge the diagram. Values less than 1 reduce the
	// diagram, while values greater than 1 enlarge the diagram. This value
	// must be greater than 0 and less than or equal to 100.
	Scale float32

	// Whether to render the diagram as if sketched by hand.
	Sketch bool
}

// D2 returns an SVGDiagram object created from the given D2 markup and options.
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
	c := ns.deps.Conf.GetConfigSection("markup").(markup_config.Config).Diagrams.D2

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
		Center:      go2.Pointer(bool(opts.Center)),
		DarkThemeID: &darkThemeID,
		NoXMLTag:    go2.Pointer(true),
		OmitVersion: go2.Pointer(true),
		Pad:         go2.Pointer(int64(opts.Padding)),
		Salt:        go2.Pointer(opts.Salt),
		Scale:       go2.Pointer(float64(opts.Scale)),
		Sketch:      go2.Pointer(bool(opts.Sketch)),
		ThemeID:     &lightThemeID,
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

	// D2 produces svg output where the content is wrapped within an additional
	// svg element. In the above, svgBytes is a byte slice of the
	// double-wrapped SVG diagram rendered by D2. We need to:
	//
	// 	1. Extract metadata from the outer element
	//		- width
	//		- height
	// 		- viewBox
	//		- preserveAspectRatio
	//	2. Extract the inner svg element
	//	3. Discard the outer wrapper
	//
	// The xml.Unmarshal function handles all of this for us provided we have
	// properly tagged the fields of the d2SVG struct. We cache the resulting
	// d2SVG, so we only need to unmarshal and optionally minify when the cache
	// is cold or when a diagram is changed.
	d2SVG := &d2SVG{}
	err = xml.Unmarshal(svgBytes, &d2SVG)
	if err != nil {
		return nil, err
	}

	if opts.Minify {
		minifiedBody, err := minify.SVG(d2SVG.Body)
		if err != nil {
			return nil, err
		}
		d2SVG.Body = minifiedBody
	}

	return d2SVG, nil
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
