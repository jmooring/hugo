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

package diagrams_test

import (
	"strings"
	"testing"

	"github.com/gohugoio/hugo/htesting"
	"github.com/gohugoio/hugo/hugolib"
)

func TestD2CodeBlockRenderHook(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['home','rss','section','sitemap','taxonomy','term']
-- layouts/integration_tests/single.html --
{{- .Content | hash.XxHash -}}
-- content/integration_tests/a.md --
---
title: A
---
~~~d2svg {darkTheme="Dark Flagship Terrastruct",layoutEngine="dagre",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/b.md --
---
title: B
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="dagre",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/c.md --
---
title: C
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/d.md --
---
title: D
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/e.md --
---
title: E
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/f.md --
---
title: F
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/g.md --
---
title: G
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/h.md --
---
title: H
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/i.md --
---
title: I
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="huey",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/j.md --
---
title: J
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="huey",id="duey",title="baz"}
x -> y
~~~
-- content/integration_tests/k.md --
---
title: K
---
~~~d2svg {darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="huey",id="duey",title="louie"}
x -> y
~~~
`

	b := hugolib.Test(t, files)

	htmlFiles := []string{
		b.FileContent("public/integration_tests/a/index.html"),
		b.FileContent("public/integration_tests/b/index.html"),
		b.FileContent("public/integration_tests/c/index.html"),
		b.FileContent("public/integration_tests/d/index.html"),
		b.FileContent("public/integration_tests/e/index.html"),
		b.FileContent("public/integration_tests/f/index.html"),
		b.FileContent("public/integration_tests/g/index.html"),
		b.FileContent("public/integration_tests/h/index.html"),
		b.FileContent("public/integration_tests/i/index.html"),
		b.FileContent("public/integration_tests/j/index.html"),
		b.FileContent("public/integration_tests/k/index.html"),
	}

	// The purpose of this assertion is to verify that the rendered diagram is
	// different on each page. We use the same D2 markup for each diagram, but
	// use a different set of parameters for each diagram.

	if !allDifferent(htmlFiles) {
		b.Error("one of more of the files is not unique")
	}

	// These assertions compare a hash of the rendered page content with known
	// values. This is not a golden test. These hash values will probably
	// change over time as the upstream project updates their libraries,
	// themes, layout engines, etc. Do not in CI environments.

	if !htesting.IsCI() {
		b.AssertFileContentEquals("public/integration_tests/a/index.html", "9a2b0619970bfda8")
		b.AssertFileContentEquals("public/integration_tests/b/index.html", "2310f586a0fa6a8d")
		b.AssertFileContentEquals("public/integration_tests/c/index.html", "c419eda6832d572e")
		b.AssertFileContentEquals("public/integration_tests/d/index.html", "8b723798b9d015b8")
		b.AssertFileContentEquals("public/integration_tests/e/index.html", "d531798f422f8fd7")
		b.AssertFileContentEquals("public/integration_tests/f/index.html", "23bd83c184c964d3")
		b.AssertFileContentEquals("public/integration_tests/g/index.html", "aaaa063aef29fb32")
		b.AssertFileContentEquals("public/integration_tests/h/index.html", "4694a551a2791245")
		b.AssertFileContentEquals("public/integration_tests/i/index.html", "752ff51711ee1727")
		b.AssertFileContentEquals("public/integration_tests/j/index.html", "75dcff78b5f3b1fd")
		b.AssertFileContentEquals("public/integration_tests/k/index.html", "ae11d3ac308125d1")
	}
}

// Helper function to determine if every item in the given slice is different.
func allDifferent[T comparable](items []T) bool {
	seen := make(map[T]bool)
	for _, item := range items {
		if seen[item] {
			return false
		}
		seen[item] = true
	}
	return true
}

func TestD2CodeBlockRenderHookFromFile(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
# disableKinds = ['home','rss','section','sitemap','taxonomy','term']
-- layouts/_default/single.html --
{{- .Content }}
-- content/p1/index.md --
---
title: p1
---
~~~d2svg {file="a.d2"}
~~~
-- content/p2.md --
---
title: p2
---
~~~d2svg {file="diagrams/b.d2"}
~~~
-- content/p1/a.d2 --
a -> b
-- assets/diagrams/b.d2 --
c -> d
`
	b := hugolib.Test(t, files)

	htmlFiles := []string{
		b.FileContent("public/p1/index.html"),
		b.FileContent("public/p2/index.html"),
	}

	// These assertions verify that the length of the rendered file is longer
	// than some minimum expected length. This is not a golden test. We don't know if the content is correct, but
	// we know we have something.

	minimumExpectedFileLength := 10000
	for k, htmlFile := range htmlFiles {
		fileLength := len(htmlFile)
		if fileLength < minimumExpectedFileLength {
			b.Errorf("[%v] file length less than expected: want %v: got %v", k, minimumExpectedFileLength, fileLength)
		}
	}

	// These assertions compare a hash of the rendered page content with known
	// values. This is not a golden test. These hash values will probably
	// change over time as the upstream project updates their libraries,
	// themes, layout engines, etc. Do not in CI environments.

	if !htesting.IsCI() {
		files = strings.ReplaceAll(files, ".Content", ".Content | hash.XxHash")

		b := hugolib.Test(t, files)

		b.AssertFileContentEquals("public/p1/index.html", "5a0dd875164f4a13")
		b.AssertFileContentEquals("public/p2/index.html", "5d64b9b935bdfd16")
	}
}

func TestD2CodeBlockRenderHookFromFileError(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
# disableKinds = ['home','rss','section','sitemap','taxonomy','term']
-- layouts/_default/single.html --
{{- .Content }}
-- content/p1/index.md --
---
title: p1
---
~~~d2svg {file="a.d2"}
~~~
`
	b, err := hugolib.TestE(t, files)

	if err == nil {
		t.Error("expected error, got none")
	}

	expected := "ERROR the embedded code block render hook for D2 diagrams was unable to fine the specified file: a.d2"
	b.AssertLogContains(expected)
}

func TestD2OtherErrorMessages(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ $text := "x -> y" }}
{{ $opts := dict "layoutEngine" "foo" }}
{{ diagrams.D2 ARGS }}
`

	filesNew := strings.ReplaceAll(files, "ARGS", "")
	want := "error calling D2: requires 1 or 2 arguments"
	testErr(t, filesNew, want)

	filesNew = strings.ReplaceAll(files, "ARGS", `""`)
	want = "error calling D2: invalid markup (empty string)"
	testErr(t, filesNew, want)

	filesNew = strings.ReplaceAll(files, "ARGS", "$text $opts")
	want = "error calling D2: layout engine must be elk or dagre"
	testErr(t, filesNew, want)

	filesNew = strings.ReplaceAll(files, "ARGS", "$opts")
	want = "error calling D2: first argument must be a string"
	testErr(t, filesNew, want)
}

// Helper function for TestD2OtherErrorMessages test.
func testErr(t *testing.T, files, want string) {
	_, err := hugolib.TestE(t, files)
	if err == nil {
		t.Error("expected error, but got none")
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error message incorrect: got = %v, want = %v", err.Error(), want)
	}
}

func TestWrapped(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{- $d := diagrams.D2 "x -> y" -}}
{{- $d.Wrapped | hash.XxHash -}}
`

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", "e51f76bec42063f5")
}
