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
	"slices"

	"github.com/gohugoio/hugo/docshelper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"oss.terrastruct.com/d2/d2themes"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
)

// Add the list of available D2 themes to the docs/data/docs.yaml file.
func init() {
	docsProvider := func() docshelper.DocProvider {
		return docshelper.DocProvider{
			"diagrams": map[string]any{
				"d2": map[string]any{
					"themes": map[string][]string{
						"dark":  getThemeNames(d2themescatalog.DarkCatalog),
						"light": getThemeNames(d2themescatalog.LightCatalog),
					},
				},
			},
		}
	}

	docshelper.AddDocProviderFunc(docsProvider)
}

func getThemeNames(catalog []d2themes.Theme) []string {
	var themes []string
	for _, theme := range catalog {
		// workaround for https://github.com/terrastruct/d2/issues/2277
		themes = append(themes, cases.Title(language.English).String(theme.Name))
	}
	slices.Sort(themes)
	return themes
}
