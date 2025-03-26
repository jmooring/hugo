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

	"github.com/gohugoio/hugo/hugolib"
)

func TestD2CodeBlockRenderHook(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['home','rss','section','sitemap','taxonomy','term']
-- layouts/integration_tests/single.html --
Hash: {{ .Content | hash.XxHash }}
Content: {{ .Content }}
-- content/integration_tests/a.md --
---
title: A
---
~~~d2svg {center=true,darkTheme="Dark Flagship Terrastruct",layoutEngine="dagre",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/b.md --
---
title: B
---
~~~d2svg {center=false,darkTheme="Dark Flagship Terrastruct",layoutEngine="dagre",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/c.md --
---
title: C
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="dagre",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/d.md --
---
title: D
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Aubergine",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/e.md --
---
title: E
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=true,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/f.md --
---
title: F
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=10,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/g.md --
---
title: G
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=1.5,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/h.md --
---
title: H
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=true,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/i.md --
---
title: I
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="foo",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/j.md --
---
title: J
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="huey",id="bar",title="baz"}
x -> y
~~~
-- content/integration_tests/k.md --
---
title: K
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="huey",id="duey",title="baz"}
x -> y
~~~
-- content/integration_tests/l.md --
---
title: L
---
~~~d2svg {center=false,darkTheme="Dark Mauve",layoutEngine="elk",lightTheme="Terminal",minify=false,padding=20,scale=0.75,sketch=false,class="huey",id="duey",title="louie"}
x -> y
~~~
`

	b := hugolib.Test(t, files)

	// These assertions compare a hash of the rendered page content with known
	// values. This is not a golden test. These hash values will probably
	// change over time as the upstream project updates their libraries,
	// themes, layout engines, etc.

	b.AssertFileContent("public/integration_tests/a/index.html", "cd483532ec666058")
	b.AssertFileContent("public/integration_tests/b/index.html", "4057f10a658ff976")
	b.AssertFileContent("public/integration_tests/c/index.html", "9ae93a5b67a8c3c9")
	b.AssertFileContent("public/integration_tests/d/index.html", "82fd52d74872d284")
	b.AssertFileContent("public/integration_tests/e/index.html", "e81101dac8d62b21")
	b.AssertFileContent("public/integration_tests/f/index.html", "71e271a31c1b25a1")
	b.AssertFileContent("public/integration_tests/g/index.html", "400c7820d4e180ac")
	b.AssertFileContent("public/integration_tests/h/index.html", "36c98cf90dfd5935")
	b.AssertFileContent("public/integration_tests/i/index.html", "31caa6579dbab424")
	b.AssertFileContent("public/integration_tests/j/index.html", "cb463e60247af9ff")
	b.AssertFileContent("public/integration_tests/k/index.html", "49320a2a56c5da6b")
	b.AssertFileContent("public/integration_tests/l/index.html", "3a4a3e759cd8f874")
}

func TestD2CodeBlockRenderHookFromFile(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
# disableKinds = ['home','rss','section','sitemap','taxonomy','term']
-- layouts/_default/single.html --
Hash: {{ .Content | hash.XxHash }}
Content: {{ .Content }}
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

	// These assertions compare a hash of the rendered page content with known
	// values. This is not a golden test. These hash values will probably
	// change over time as the upstream project updates their libraries,
	// themes, layout engines, etc.

	b.AssertFileContent("public/p1/index.html", "dc8ef1272745e7a8")
	b.AssertFileContent("public/p2/index.html", "d37cf450e2a50a48")
}

func TestD2CodeBlockRenderHookFromFileError(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
# disableKinds = ['home','rss','section','sitemap','taxonomy','term']
-- layouts/_default/single.html --
{{ .Content }}
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
Hash: {{ $d.Wrapped | hash.XxHash }}
Content: {{ $d.Wrapped }}
`

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", "ea6cd6ed61d7d949")
}
