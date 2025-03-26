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
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func Test_getThemeID(t *testing.T) {
	tests := []struct {
		themeName string
		want      int64
		wantErr   error
	}{
		// light theme
		{"Buttered Toast", 105, nil},
		{"buttered toast", 105, nil},
		// dark theme
		{"Dark Mauve", 200, nil},
		{"dark mauve", 200, nil},
		// errors
		{"Arthur Dent", 0, fmt.Errorf("theme does not exist: Arthur Dent")},
		{"", 0, fmt.Errorf("cannot resolve an empty string to a theme ID")},
	}

	for k, tt := range tests {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			got, err := getThemeID(tt.themeName)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("got error: %v, want error: %v", err, tt.wantErr)
					return
				}
				return
			}
			if tt.wantErr != nil {
				t.Errorf("expected error, but got none")
				return
			}
			if got != tt.want {
				t.Errorf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_validateOptions(t *testing.T) {
	type testCase struct {
		name    string
		opts    *d2Options
		wantErr error
	}

	tests := []testCase{
		{
			name:    "Valid Options",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "dagre", LightTheme: "Aubergine", Padding: 20, Scale: 1.25},
			wantErr: nil,
		},
		{
			name:    "Empty Dark Theme",
			opts:    &d2Options{DarkTheme: "", LayoutEngine: "dagre", LightTheme: "Aubergine", Padding: 20, Scale: 1.25},
			wantErr: errors.New("invalid dark theme (empty string)"),
		},
		{
			name:    "Empty Layout Engine",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "", LightTheme: "Aubergine", Padding: 20, Scale: 1.25},
			wantErr: errors.New("invalid layout engine (empty string)"),
		},
		{
			name:    "Invalid Layout Engine",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "foo", LightTheme: "Aubergine", Padding: 20, Scale: 1.25},
			wantErr: errors.New("layout engine must be elk or dagre"),
		},
		{
			name:    "Empty Light Theme",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "dagre", LightTheme: "", Padding: 20, Scale: 1.25},
			wantErr: errors.New("invalid light theme (empty string)"),
		},
		{
			name:    "Padding Too Large",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "dagre", LightTheme: "Aubergine", Padding: 1001, Scale: 1.25},
			wantErr: errors.New("padding must be an integer between 0 and 1000 inclusive"),
		},
		{
			name:    "Scale Out of Range (Low)",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "dagre", LightTheme: "Aubergine", Padding: 20, Scale: 0},
			wantErr: errors.New("scale must be greater than 0 and less than or equal to 100"),
		},
		{
			name:    "Scale Out of Range (High)",
			opts:    &d2Options{DarkTheme: "Dark Mauve", LayoutEngine: "dagre", LightTheme: "Aubergine", Padding: 20, Scale: 101},
			wantErr: errors.New("scale must be greater than 0 and less than or equal to 100"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptions(tt.opts)

			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if err.Error() != tt.wantErr.Error() {
					t.Errorf("got error: %v, want error: %v", err, tt.wantErr)
				}
			} else if tt.wantErr != nil {
				t.Errorf("expected error, but got none")
			}
		})
	}
}
