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

package modules_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gohugoio/hugo/hugolib"
)

func TestModuleImportWithVersion(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
baseURL = "https://example.org"
[[module.imports]]
path    = "github.com/bep/hugo-mod-misc/dummy-content"
version = "v0.2.0"
[[module.imports]]
path    = "github.com/bep/hugo-mod-misc/dummy-content"
version = "v0.1.0"
[[module.imports.mounts]]
source = "content"
target = "content/v1"
-- layouts/all.html --
Title: {{ .Title }}|Summary: {{ .Summary }}|
Deps: {{ range hugo.Deps}}{{ printf "%s@%s" .Path .Version }}|{{ end }}$

`

	b := hugolib.Test(t, files, hugolib.TestOptWithOSFs())

	b.AssertFileContent("public/index.html", "Deps: project@|github.com/bep/hugo-mod-misc/dummy-content@v0.2.0|github.com/bep/hugo-mod-misc/dummy-content@v0.1.0|$")

	b.AssertFileContent("public/blog/music/autumn-leaves/index.html", "Autumn Leaves is a popular jazz standard") // v0.2.0
	b.AssertFileContent("public/v1/blog/music/autumn-leaves/index.html", "Lorem markdownum, placidi peremptis")   // v0.1.0
}

// Issue 14010
func TestModuleImportErrors(t *testing.T) {
	t.Parallel()

	files := `-- go.mod --
module foo

go 1.24.0
-- hugo.toml --
[[module.imports]]
PATH
VERSION
`

	// The github.com/bep/hugomodnogomod repository used in the tests below
	// does not contain a go.mod file.

	// These should not throw an error.

	f := strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "").Replace(files)
	b, err := hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = 'v1.0.0'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = 'v2.0.0'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = 'feddafb3f711ef852114e1dec6553f616ead7a37'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = 'latest'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = 'main'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = '>v1.0.0'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNil)

	// These should throw an error.

	f = strings.NewReplacer("PATH", "", "VERSION", "").Replace(files)
	b, err = hugolib.TestE(t, f)
	b.Assert(err, qt.IsNotNil)
	b.Assert(err, qt.ErrorMatches, `.*failed to load modules: module "" not found.*`)

	f = strings.NewReplacer("PATH", "path = 'foo'", "VERSION", "").Replace(files)
	b, err = hugolib.TestE(t, f)
	b.Assert(err, qt.IsNotNil)
	b.Assert(err, qt.ErrorMatches, `.*failed to load modules: module "foo" not found.*`)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = '1.0.0'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNotNil)
	b.Assert(err, qt.ErrorMatches, `.*invalid version: unknown revision 1.0.0`)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = 'v99.0.0'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNotNil)
	b.Assert(err, qt.ErrorMatches, `.*invalid version: unknown revision v99.0.0.*`)

	f = strings.NewReplacer("PATH", "path = 'github.com/bep/hugomodnogomod'", "VERSION", "version = '>1.0.0'").Replace(files)
	b, err = hugolib.TestE(t, f, hugolib.TestOptOsFs())
	b.Assert(err, qt.IsNotNil)
	b.Assert(err, qt.ErrorMatches, `.*invalid semantic version "1.0.0" in range ">1.0.0".*`)
}
