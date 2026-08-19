//go:build !windows

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package systeminfo

import "testing"

func TestParseOSReleasePrettyName(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "quoted pretty name",
			content: `NAME="BlueKing Linux"
PRETTY_NAME="BlueKing Linux 3.0"`,
			want: "BlueKing Linux 3.0",
		},
		{
			name:    "unquoted pretty name",
			content: `PRETTY_NAME=BlueKing Linux`,
			want:    "BlueKing Linux",
		},
		{
			name:    "missing pretty name",
			content: `NAME="BlueKing Linux"`,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOSReleasePrettyName(tt.content)
			if got != tt.want {
				t.Fatalf("parseOSReleasePrettyName(%q) = %q, want %q", tt.content, got, tt.want)
			}
		})
	}
}

func TestUnquoteOSReleaseValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "double quoted value",
			value: `"BlueKing Linux 3.0"`,
			want:  "BlueKing Linux 3.0",
		},
		{
			name:  "quoted value trims inner spaces",
			value: `"  BlueKing Linux 3.0  "`,
			want:  "BlueKing Linux 3.0",
		},
		{
			name:  "unquoted value is preserved after outer trim",
			value: " BlueKing Linux 3.0 ",
			want:  "BlueKing Linux 3.0",
		},
		{
			name:  "empty value",
			value: " ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unquoteOSReleaseValue(tt.value)
			if got != tt.want {
				t.Fatalf("unquoteOSReleaseValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
