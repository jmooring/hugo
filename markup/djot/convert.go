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

// Package djot converts djot markup to HTML using an external helper.
package djot

import (
	"github.com/gohugoio/hugo/common/hexec"
	"github.com/gohugoio/hugo/htesting"
	"github.com/gohugoio/hugo/identity"

	"github.com/gohugoio/hugo/markup/converter"
	"github.com/gohugoio/hugo/markup/internal"
)

// Provider is the package entry point.
var Provider converter.ProviderProvider = provider{}

type provider struct{}

func (p provider) New(cfg converter.ProviderConfig) (converter.Provider, error) {
	return converter.NewProvider("djot", func(ctx converter.DocumentContext) (converter.Converter, error) {
		return &djotConverter{
			ctx: ctx,
			cfg: cfg,
		}, nil
	}), nil
}

type djotConverter struct {
	ctx converter.DocumentContext
	cfg converter.ProviderConfig
}

func (c *djotConverter) Convert(ctx converter.RenderContext) (converter.ResultRender, error) {
	b, err := c.getDjotContent(ctx.Src, c.ctx)
	if err != nil {
		return nil, err
	}
	return converter.Bytes(b), nil
}

func (c *djotConverter) Supports(feature identity.Identity) bool {
	return false
}

// getDjotContent calls an external helper to convert djot markup to HTML.
func (c *djotConverter) getDjotContent(src []byte, ctx converter.DocumentContext) ([]byte, error) {
	logger := c.cfg.Logger
	binaryName := getDjotBinaryName()
	if binaryName == "" {
		logger.Printf("%s not found in $PATH: leaving djot content unrendered\n", djotBinary)
		return src, nil
	}
	args := []string{} // use this for godjot
	// args := []string{"--from=djot"} // use this for pandoc
	return internal.ExternallyRenderContent(c.cfg, ctx, src, binaryName, args)
}

// TODO: jmm We may want to default to something but allow user to specify
// executable in env var that does NOT begin with HUGO_ (security reasons),
// similar to what we do with  DART_SASS_BINARY. Then uses can use djot.js,
// godjot, pandoc, the rust implementation, etc. Also see args differences above.
const djotBinary = "godjot"

func getDjotBinaryName() string {
	if hexec.InPath(djotBinary) {
		return djotBinary
	}
	return ""
}

// Supports returns whether the djot external helper is installed on this computer.
func Supports() bool {
	hasBin := getDjotBinaryName() != ""
	if htesting.SupportsAll() {
		if !hasBin {
			panic("godjot not installed")
		}
		return true
	}
	return hasBin
}
