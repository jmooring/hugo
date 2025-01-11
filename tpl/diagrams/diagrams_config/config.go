// Copyright 2025 Hugo Authors. All rights reserved.
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

package diagrams_config

import (
	"github.com/gohugoio/hugo/common/maps"
	"github.com/gohugoio/hugo/config"

	"github.com/mitchellh/mapstructure"
)

type Config struct {
	D2 D2Config
}

type D2Config struct {
	DarkTheme    string
	LayoutEngine string
	LightTheme   string
	Minify       bool
	Padding      uint16
	Scale        float32
	Sketch       bool
}

var defaultConfig = Config{
	D2: defaultD2Config,
}

var defaultD2Config = D2Config{
	DarkTheme:    "Dark Flagship Terrastruct",
	LayoutEngine: "dagre",
	LightTheme:   "Neutral Default",
	Minify:       true,
	Padding:      0,
	Scale:        1,
	Sketch:       false,
}

func Decode(cfg config.Provider) (Config, error) {
	conf := defaultConfig

	m := cfg.GetStringMap("diagrams")
	if m == nil {
		return conf, nil
	}
	m = maps.CleanConfigStringMap(m)

	err := mapstructure.WeakDecode(m, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
