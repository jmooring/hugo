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

package lang_test

import (
	"testing"

	"github.com/gohugoio/hugo/hugolib"
)

func TestLanguageName(t *testing.T) {
	t.Parallel()

	files := `
-- config.toml --
disableKinds = ['page','section','rss','sitemap','taxonomy','term']
defaultContentLanguage = 'en'
defaultContentLanguageInSubdir = true
[languages.en]
languageCode = 'en-GB'
weight = 1
[languages.de]
languageCode = 'de-DE'
weight = 2
[languages.fr]
languageCode = 'fr-FR'
weight = 2
-- content/_index.en.md --
---
title: home en
---
-- content/_index.de.md --
---
title: home de
---
-- content/_index.fr.md --
---
title: home fr
---
-- layouts/index.html --
{{ $targetLang := or .Language.LanguageCode .Language.Lang }}
{{ range .AllTranslations }}
	{{ $sourceLang := or .Language.LanguageCode .Language.Lang }}
	<li><a href="{{ .RelPermalink }}">{{ lang.LanguageName $sourceLang $targetLang }}</a></li>
{{ end }}
  `

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/en/index.html",
		`<li><a href="/en/">British English</a></li>`,
		`<li><a href="/de/">German</a></li>`,
		`<li><a href="/fr/">French</a></li>`,
	)
	b.AssertFileContent("public/de/index.html",
		`<li><a href="/en/">Britisches Englisch</a></li>`,
		`<li><a href="/de/">Deutsch</a></li>`,
		`<li><a href="/fr/">Französisch</a></li>`,
	)
	b.AssertFileContent("public/fr/index.html",
		`<li><a href="/en/">anglais britannique</a></li>`,
		`<li><a href="/de/">allemand</a></li>`,
		`<li><a href="/fr/">français</a></li>`,
	)
}
