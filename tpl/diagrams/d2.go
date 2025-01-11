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
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/tdewolff/minify/v2/minify"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

type d2Diagram struct {
	d struct {
		inner   string
		wrapped string
		width   int
		height  int
		viewBox string
	}
}

func (d d2Diagram) Inner() template.HTML {
	return template.HTML(d.d.inner)
}

func (d d2Diagram) Wrapped() template.HTML {
	return template.HTML(d.d.wrapped)
}

func (d d2Diagram) Width() int {
	return d.d.width
}

func (d d2Diagram) Height() int {
	return d.d.height
}

func (d d2Diagram) ViewBox() string {
	return d.d.viewBox
}

type d2Options struct {
	// The D2 theme to use if the system is in dark mode. This value is
	// case-insensitive. See https://d2lang.com/tour/themes.
	DarkTheme string

	// The D2 theme to use if the system is in light mode or has no preference.
	// This value is case-insensitive. See https://d2lang.com/tour/themes.
	LightTheme string

	// Whether to minify the SVG elements.
	Minify bool

	// The number of pixels with which to pad each side of the diagram. This
	// value must be within the bounds of 0 and 1000, inclusive.
	Padding uint16

	// How much to reduce or enlarge the diagram. Values less than 1 reduce the
	// diagram, while values greater than 1 enlarge the diagram. This value
	// must be greater than 0 and less than or equal to 100.
	Scale float32

	// Whether to render the diagram as a sketch.
	Sketch bool
}

const (
	d2DefaultDarkTheme  string  = "Dark Flagship Terrastruct"
	d2DefaultLightTheme string  = "Neutral Default"
	d2DefaultMinify     bool    = true
	d2DefaultPadding    uint16  = 0
	d2DefaultScale      float32 = 1
	d2DefaultSketch     bool    = false
)

// D2 returns an SVG diagram object from the given D2 markup using the
// specified options.
func (d *Namespace) D2(args ...any) (SVGDiagram, error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.New("requires 1 or 2 arguments")
	}

	markup, err := cast.ToStringE(args[0])
	if err != nil {
		return nil, err
	}

	if markup == "" {
		return nil, errors.New("cannot create diagram from an empty string")
	}

	opts := &d2Options{
		DarkTheme:  d2DefaultDarkTheme,
		LightTheme: d2DefaultLightTheme,
		Minify:     d2DefaultMinify,
		Padding:    d2DefaultPadding,
		Scale:      d2DefaultScale,
		Sketch:     d2DefaultSketch,
	}

	if len(args) == 2 {
		err := mapstructure.WeakDecode(args[1], &opts)
		if err != nil {
			return nil, err
		}
	}

	if opts.Padding > 1000 {
		return nil, errors.New("padding must be an integer between 0 and 1000 inclusive")
	}

	if opts.Scale <= 0 || opts.Scale > 100 {
		return nil, errors.New("scale must be greater than 0 and less than or equal to 100")
	}

	return createD2Diagram(markup, opts)
}

func createD2Diagram(markup string, opts *d2Options) (d2Diagram, error) {
	d2 := d2Diagram{}

	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return d2, err
	}

	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}

	lightThemeID, err := getThemeID(opts.LightTheme)
	if err != nil {
		return d2, err
	}

	darkThemeID, err := getThemeID(opts.DarkTheme)
	if err != nil {
		return d2, err
	}

	renderOpts := &d2svg.RenderOpts{
		ThemeID:     &lightThemeID,
		DarkThemeID: &darkThemeID,
		Pad:         go2.Pointer(int64(opts.Padding)),
		Scale:       go2.Pointer(float64(opts.Scale)),
		Sketch:      go2.Pointer(bool(opts.Sketch)),
	}

	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}

	ctx := log.WithDefault(context.Background())

	diagram, _, err := d2lib.Compile(ctx, markup, compileOpts, renderOpts)
	if err != nil {
		return d2, err
	}

	svgb, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return d2, err
	}

	wrapped := string(svgb)
	if opts.Minify {
		wrapped, err = minify.SVG(wrapped)
		if err != nil {
			return d2, nil
		}
	}

	type d2SVG struct {
		Inner   []byte `xml:",innerxml"`
		Width   int    `xml:"width,attr"`
		Height  int    `xml:"height,attr"`
		ViewBox string `xml:"viewBox,attr"`
	}

	var svg d2SVG
	err = xml.Unmarshal(svgb, &svg)
	if err != nil {
		return d2, err
	}

	inner := string(svg.Inner)
	if opts.Minify {
		inner, err = minify.SVG(inner)
		if err != nil {
			return d2, nil
		}
	}

	d2.d.inner = inner
	d2.d.wrapped = wrapped
	d2.d.width = svg.Width
	d2.d.height = svg.Height
	d2.d.viewBox = svg.ViewBox

	return d2, nil
}

// getTheme return the theme ID corresponding to the given theme name.
func getThemeID(themeName string) (int64, error) {
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
