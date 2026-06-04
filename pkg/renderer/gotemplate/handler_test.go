/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package gotemplate

import (
	"fmt"
	"testing"
)

type customStringer struct {
	value string
}

func (cs *customStringer) String() string {
	return cs.value
}

type customString string

type customByte []byte

// Test_TemplateRender_Render tests the Render method of the template rendering handler.
func Test_TemplateRender_Render(t *testing.T) {
	type args struct {
		tmpl string
		data map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic render",
			args: args{
				tmpl: "hello, {{.name}}",
				data: map[string]any{
					"name": "world",
				},
			},
			want:    "hello, world",
			wantErr: false,
		},
		{
			name: "use upper function",
			args: args{
				tmpl: "hello, {{.name | upper}}",
				data: map[string]any{
					"name": "world",
				},
			},
			want:    "hello, WORLD",
			wantErr: false,
		},
		{
			name: "use default function",
			args: args{
				tmpl: "hello, {{ .name | default \"guest\" }}",
				data: map[string]any{},
			},
			want:    "hello, guest",
			wantErr: false,
		},
		{
			name: "use join function",
			args: args{
				tmpl: "items: {{ .items | join \", \" }}",
				data: map[string]any{
					"items": []string{"apple", "banana", "cherry"},
				},
			},
			want:    "items: apple, banana, cherry",
			wantErr: false,
		},
		{
			name: "use replace function",
			args: args{
				tmpl: "path: {{ .path | replace \"/\" \"-\" }}",
				data: map[string]any{
					"path": "/usr/local/bin",
				},
			},
			want:    "path: -usr-local-bin",
			wantErr: false,
		},
		{
			name: "use first function with string slice",
			args: args{
				tmpl: "first: {{ .items | first }}",
				data: map[string]any{
					"items": []string{"apple", "banana", "cherry"},
				},
			},
			want:    "first: apple",
			wantErr: false,
		},
		{
			name: "use last function with int slice",
			args: args{
				tmpl: "last: {{ .items | last }}",
				data: map[string]any{
					"items": []int{1, 2, 3},
				},
			},
			want:    "last: 3",
			wantErr: false,
		},
		{
			name: "use first function default when slice is empty",
			args: args{
				tmpl: "first: {{ .items | first | default \"missing\" }}",
				data: map[string]any{
					"items": []string{},
				},
			},
			want:    "first: missing",
			wantErr: false,
		},
		{
			name: "use last function default when slice is empty",
			args: args{
				tmpl: "last: {{ .items | last | default \"missing\" }}",
				data: map[string]any{
					"items": []string{},
				},
			},
			want:    "last: missing",
			wantErr: false,
		},
		{
			name: "use invalid template",
			args: args{
				tmpl: "hello, {{.name | unknownFunc}}",
				data: map[string]any{
					"name": "world",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "use join int function",
			args: args{
				tmpl: "items: {{ .items | join 2 }}",
				data: map[string]any{
					"items": []string{"apple", "banana", "cherry"},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "use default function without quotes",
			args: args{
				tmpl: "hello, {{ .name | default guest }}",
				data: map[string]any{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "use replace function whithsout quotes",
			args: args{
				tmpl: "path: {{ .path | replace / - }}",
				data: map[string]any{
					"path": "/usr/local/bin",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "use replace function whith int args",
			args: args{
				tmpl: "path: {{ .path | replace 2 1 }}",
				data: map[string]any{
					"path": "usr2local2bin",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "use indent function",
			args: args{
				tmpl: "content:\n{{ .content | indent 4 }}",
				data: map[string]any{
					"content": "line1\nline2\nline3",
				},
			},
			want:    "content:\n    line1\n    line2\n    line3",
			wantErr: false,
		},
		{
			name: "use nindent function",
			args: args{
				tmpl: "content:{{ .content | nindent 4 }}",
				data: map[string]any{
					"content": "line1\nline2\nline3",
				},
			},
			want:    "content:\n    line1\n    line2\n    line3",
			wantErr: false,
		},
		{
			name: "use toString function with int",
			args: args{
				tmpl: "value: {{ .value | toString }}",
				data: map[string]any{
					"value": 123,
				},
			},
			want:    "value: 123",
			wantErr: false,
		},
		{
			name: "use toString function with error",
			args: args{
				tmpl: "error: {{ .err | toString }}",
				data: map[string]any{
					"err": func() error { return fmt.Errorf("an error occurred") }(),
				},
			},
			want:    "error: an error occurred",
			wantErr: false,
		}, {
			name: "use toString function with byte slice",
			args: args{
				tmpl: "data: {{ .data | toString }}",
				data: map[string]any{
					"data": []byte("byte slice data"),
				},
			},
			want:    "data: byte slice data",
			wantErr: false,
		},
		{
			name: "use toString function with Stringer",
			args: args{
				tmpl: "stringer: {{ .str | toString }}",
				data: map[string]any{
					"str": struct{ fmt.Stringer }{Stringer: &customStringer{"custom string"}},
				},
			},
			want:    "stringer: custom string",
			wantErr: false,
		},
		{
			name: "use toString function with custom string type",
			args: args{
				tmpl: "customString: {{ .cs | toString }}",
				data: map[string]any{
					"cs": customString("custom string type"),
				},
			},
			want:    "customString: custom string type",
			wantErr: false,
		},
		{
			name: "use toString function with custom byte slice type",
			args: args{
				tmpl: "customByte: {{ .cb | toString }}",
				data: map[string]any{
					"cb": customByte("custom byte slice"),
				},
			},
			want:    "customByte: custom byte slice",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New().Render(tt.args.tmpl, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Render() got = %v, want %v", got, tt.want)
			}

			t.Logf("Render() got = %v", got)
		})
	}
}
