// Copyright 2024 The Hugo Authors. All rights reserved.
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
	"html/template"

	"github.com/gohugoio/hugo/cache/dynacache"
	"github.com/gohugoio/hugo/deps"
)

// New returns a new instance of the diagrams-namespaced template functions.
func New(deps *deps.Deps) *Namespace {
	if deps.MemCache == nil {
		panic("must provide MemCache")
	}

	return &Namespace{
		deps: deps,
		cacheD2: dynacache.GetOrCreatePartition[string, []byte](
			deps.MemCache,
			"/tmpl/diagrams/d2",
			dynacache.OptionsPartition{Weight: 30, ClearWhen: dynacache.ClearNever},
		),
	}
}

// Namespace provides template functions for the diagrams namespace.
type Namespace struct {
	deps    *deps.Deps
	cacheD2 *dynacache.Partition[string, []byte]
}

type SVGDiagram interface {
	// Wrapped returns the diagram as an SVG, including the <svg> container.
	Wrapped() template.HTML

	// Inner returns the inner markup of the SVG.
	// This allows for the <svg> container to be created manually.
	Inner() template.HTML

	// Width returns the width attribute of the SVG.
	Width() int

	// Height returns the height attribute of the SVG.
	Height() int

	// ViewBox returns the viewBox attribute of the SVG.
	ViewBox() string

	// PreserveAspectRatio returns the preserveAspectRatio attribute of the SVG.
	PreserveAspectRatio() string
}
