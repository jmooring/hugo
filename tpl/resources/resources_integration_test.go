// Copyright 2022s The Hugo Authors. All rights reserved.
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

package resources_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gohugoio/hugo/hugolib"
	"github.com/gohugoio/hugo/resources/resource_transformers/tocss/dartsass"
	"github.com/gohugoio/hugo/resources/resource_transformers/tocss/scss"
)

func TestCopy(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
baseURL = "http://example.com/blog"
-- assets/images/pixel.png --
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==
-- layouts/home.html --
{{/* Image resources */}}
{{ $img := resources.Get "images/pixel.png" }}
{{ $imgCopy1 := $img | resources.Copy "images/copy.png"  }}
{{ $imgCopy1 = $imgCopy1.Resize "3x4"}}
{{ $imgCopy2 := $imgCopy1 | resources.Copy "images/copy2.png" }}
{{ $imgCopy3 := $imgCopy1 | resources.Copy "images/copy3.png" }}
Image Orig:  {{ $img.RelPermalink}}|{{ $img.MediaType }}|{{ $img.Width }}|{{ $img.Height }}|
Image Copy1:  {{ $imgCopy1.RelPermalink}}|{{ $imgCopy1.MediaType }}|{{ $imgCopy1.Width }}|{{ $imgCopy1.Height }}|
Image Copy2:  {{ $imgCopy2.RelPermalink}}|{{ $imgCopy2.MediaType }}|{{ $imgCopy2.Width }}|{{ $imgCopy2.Height }}|
Image Copy3:  {{ $imgCopy3.MediaType }}|{{ $imgCopy3.Width }}|{{ $imgCopy3.Height }}|

{{/* Generic resources */}}
{{ $targetPath := "js/vars.js" }}
{{ $orig := "let foo;" | resources.FromString "js/foo.js" }}
{{ $copy1 := $orig | resources.Copy "js/copies/bar.js" }}
{{ $copy2 := $orig | resources.Copy "js/copies/baz.js" | fingerprint "md5" }}
{{ $copy3 := $copy2 | resources.Copy "js/copies/moo.js" | minify }}

Orig: {{ $orig.RelPermalink}}|{{ $orig.MediaType }}|{{ $orig.Content | safeJS }}|
Copy1: {{ $copy1.RelPermalink}}|{{ $copy1.MediaType }}|{{ $copy1.Content | safeJS }}|
Copy2: {{ $copy2.RelPermalink}}|{{ $copy2.MediaType }}|{{ $copy2.Content | safeJS }}|
Copy3: {{ $copy3.RelPermalink}}|{{ $copy3.MediaType }}|{{ $copy3.Content | safeJS }}|

	`

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", `
Image Orig:  /blog/images/pixel.png|image/png|1|1|
Image Copy1:  /blog/images/copy_hu_ef85c36d9b9eff0.png|image/png|3|4|
Image Copy2:  /blog/images/copy2.png|image/png|3|4|
Image Copy3:  image/png|3|4|
Orig: /blog/js/foo.js|text/javascript|let foo;|
Copy1: /blog/js/copies/bar.js|text/javascript|let foo;|
Copy2: /blog/js/copies/baz.a677329fc6c4ad947e0c7116d91f37a2.js|text/javascript|let foo;|
Copy3: /blog/js/copies/moo.a677329fc6c4ad947e0c7116d91f37a2.min.js|text/javascript|let foo|
		`)

	b.AssertFileExists("public/images/copy2.png", true)
	// No permalink used.
	b.AssertFileExists("public/images/copy3.png", false)
}

func TestCopyPageShouldFail(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
-- layouts/home.html --
{{/* This is currently not supported. */}}
{{ $copy := .Copy "copy.md" }}

	`

	b, err := hugolib.TestE(t, files)

	b.Assert(err, qt.IsNotNil)
}

func TestGet(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
baseURL = "http://example.com/blog"
-- assets/images/pixel.png --
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==
-- layouts/home.html --
{{ with resources.Get "images/pixel.png" }}Image OK{{ else }}Image not found{{ end }}
{{ with resources.Get "" }}Failed{{ else }}Empty string not found{{ end }}


	`

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", `
Image OK
Empty string not found

		`)
}

func TestResourcesGettersShouldNotNormalizePermalinks(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
baseURL = "http://example.com/"
-- assets/401K Prospectus.txt --
Prospectus.
-- layouts/home.html --
{{ $name := "401K Prospectus.txt" }}
Get: {{ with resources.Get $name }}{{ .RelPermalink }}|{{ .Permalink }}|{{ end }}
GetMatch: {{ with resources.GetMatch $name }}{{ .RelPermalink }}|{{ .Permalink }}|{{ end }}
Match: {{ with (index (resources.Match $name) 0) }}{{ .RelPermalink }}|{{ .Permalink }}|{{ end }}
ByType: {{ with (index (resources.ByType "text") 0) }}{{ .RelPermalink }}|{{ .Permalink }}|{{ end }}
	`

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", `
Get: /401K%20Prospectus.txt|http://example.com/401K%20Prospectus.txt|
GetMatch: /401K%20Prospectus.txt|http://example.com/401K%20Prospectus.txt|
Match: /401K%20Prospectus.txt|http://example.com/401K%20Prospectus.txt|
ByType: /401K%20Prospectus.txt|http://example.com/401K%20Prospectus.txt|

		`)
}

func TestGlobalResourcesNotPublishedRegressionIssue12190(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- assets/a.txt --
I am a.txt
-- assets/b.txt --
I am b.txt
-- layouts/home.html --
Home.
{{ with resources.ByType "text" }}
  {{ with .Get "a.txt" }}
    {{ .Publish }}
  {{ end }}
  {{ with .GetMatch "*b*" }}
    {{ .Publish }}
  {{ end }}
{{ end }}
`

	b := hugolib.Test(t, files)

	b.AssertFileExists("public/a.txt", true) // failing test
	b.AssertFileExists("public/b.txt", true) // failing test
}

func TestGlobalResourcesNotPublishedRegressionIssue12214(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- assets/files/a.txt --
I am a.txt
-- assets/files/b.txt --
I am b.txt
-- assets/files/c.txt --
I am c.txt
-- assets/files/C.txt --
I am C.txt
-- layouts/home.html --
Home.
{{ with resources.ByType "text" }}
  {{ with .Get "files/a.txt" }}
    {{ .Publish }}
	files/a.txt: {{ .Name }}
  {{ end }}
  {{ with .Get "/files/a.txt" }}
	/files/a.txt: {{ .Name }}
  {{ end }}
  {{ with .GetMatch "files/*b*" }}
    {{ .Publish }}
	files/*b*: {{ .Name }}
  {{ end }}
  {{ with .GetMatch "files/C*" }}
    {{ .Publish }}
    files/C*: {{ .Name }}
  {{ end }}
  {{ with .GetMatch "files/c*" }}
	{{ .Publish }}
	files/c*: {{ .Name }}
  {{ end }}
  {{ with .GetMatch "/files/c*" }}
    /files/c*: {{ .Name }}
  {{ end }}
  {{ with .Match "files/C*" }}
	match files/C*: {{ len . }}|
  {{ end }}
  {{ with .Match "/files/C*" }}
  match /files/C*: {{ len . }}|
{{ end }}
{{ end }}
`

	b := hugolib.Test(t, files)

	// c.txt and C.txt both normalize to /files/c.txt; lowercase wins.
	b.AssertFileContent("public/index.html", `
files/a.txt: /files/a.txt
files/C*: /files/c.txt
files/c*: /files/c.txt
files/*b*: /files/b.txt
/files/c*: /files/c.txt
/files/a.txt: /files/a.txt
match files/C*: 1|
match /files/C*: 1|
	`)

	b.AssertFileContent("public/files/a.txt", "I am a.txt")
	b.AssertFileContent("public/files/b.txt", "I am b.txt")
	b.AssertFileContent("public/files/c.txt", "I am c.txt")
}

// Issue #12961
func TestDartSassVars(t *testing.T) {
	t.Parallel()

	if !scss.Supports() || !dartsass.Supports() {
		t.Skip()
	}

	files := `
-- hugo.toml --
disableKinds = ['page','section','rss','sitemap','taxonomy','term']
-- layouts/home.html --
{{ $opts := dict "transpiler" "dartsass" "outputStyle" "compressed" "vars" (dict "color" "red") }}
{{ with resources.Get "dartsass.scss" | css.Sass $opts }}
  {{ .Content }}
{{ end }}

{{ $opts := dict "transpiler" "libsass" "outputStyle" "compressed" "vars" (dict "color" "blue") }}
{{ with resources.Get "libsass.scss" | css.Sass $opts }}
  {{ .Content }}
{{ end }}
-- assets/dartsass.scss --
@use "hugo:vars" as v;
.dartsass {
  color: v.$color;
}
-- assets/libsass.scss --
@import "hugo:vars";
.libsass {
  color: $color;
}
`

	b := hugolib.Test(t, files, hugolib.TestOptWarn())

	b.AssertFileContent(
		"public/index.html",
		".dartsass{color:red}",
		".libsass{color:blue}",
	)
	b.AssertLogContains("! WARN  Dart Sass: hugo:vars")
}

// See issue 7686.
// Assets in mixed-case directories must be findable with any glob case and
// must always publish to a lowercase path, regardless of OS case-sensitivity.
// When both assets/ab/foo.txt and assets/aB/foo.txt exist, the lowercase path wins.
func TestAssetsMixedCaseDirectory(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- assets/aB/1.txt --
I am aB/1.txt
-- assets/ab/1.txt --
I am ab/1.txt
-- assets/CD/2.txt --
I am CD/2.txt
-- assets/a/File.pdf --
I am a/File.pdf
-- assets/A/file.pdf --
I am A/file.pdf
-- layouts/home.html --
GetMatch ab: {{ with resources.GetMatch "ab/*" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
GetMatch aB: {{ with resources.GetMatch "aB/*" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
GetMatch AB: {{ with resources.GetMatch "AB/*" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
Match AB count: {{ len (resources.Match "AB/*") }}
GetMatch CD: {{ with resources.GetMatch "CD/*" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
GetMatch cd: {{ with resources.GetMatch "cd/*" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
GetMatch a/file.pdf: {{ with resources.GetMatch "a/file.pdf" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
GetMatch a/File.pdf: {{ with resources.GetMatch "a/File.pdf" }}{{ .RelPermalink }}|{{ .Content }}{{ end }}
Match a count: {{ len (resources.Match "a/*") }}
`

	b := hugolib.Test(t, files)

	b.AssertFileContent(
		"public/index.html",
		// Lowercase path wins when both ab/ and aB/ exist.
		"GetMatch ab: /ab/1.txt|I am ab/1.txt",
		"GetMatch aB: /ab/1.txt|I am ab/1.txt",
		"GetMatch AB: /ab/1.txt|I am ab/1.txt",
		// Duplicate suppressed: the lowercase directory was used as the walk root.
		"Match AB count: 1",
		// Mixed-case-only directory: found regardless of pattern case, URL is lowercase.
		"GetMatch CD: /cd/2.txt|I am CD/2.txt",
		"GetMatch cd: /cd/2.txt|I am CD/2.txt",
		// Both a/File.pdf and A/file.pdf share the same normalized path (a/file.pdf);
		// lowercase dir (a/) wins, so both patterns resolve to a/File.pdf.
		"GetMatch a/file.pdf: /a/File.pdf|I am a/File.pdf",
		"GetMatch a/File.pdf: /a/File.pdf|I am a/File.pdf",
		// A/file.pdf deduplicates against a/File.pdf (same normalized path).
		"Match a count: 1",
	)
}

// See issue 15208.
func TestPublish(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ["taxonomy", "term", "section", "RSS", "sitemap", "robotsTXT", "404"]
-- assets/js/main.js --
let foo;
-- layouts/home.html --
{{ $r := resources.Get "js/main.js" | minify | resources.Publish }}
Name: {{ $r.Name }}|
`

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", "Name: /js/main.js|")
	b.AssertFileExists("public/js/main.min.js", true)
}

// See issue 15086.
func TestPostProcessDeprecated(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ["taxonomy", "term", "section", "RSS", "sitemap", "robotsTXT", "404"]
-- assets/css/main.css --
body { color: red; }
-- layouts/home.html --
{{ $r := resources.Get "css/main.css" | minify | resources.PostProcess }}
{{ $r.RelPermalink }}
`

	b := hugolib.Test(t, files, hugolib.TestOptInfo())

	b.AssertLogContains("resources.PostProcess was deprecated in Hugo v0.164.0")
}
