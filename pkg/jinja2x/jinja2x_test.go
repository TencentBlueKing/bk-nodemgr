/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package jinja2x

import (
	"testing"
)

func TestRender(t *testing.T) {
	type args struct {
		tmpl    string
		context map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple variable substitution",
			args: args{
				tmpl: "Hello {{ name }}!",
				context: map[string]any{
					"name": "World",
				},
			},
			want:    "Hello World!",
			wantErr: false,
		},
		{
			name: "multiple variables",
			args: args{
				tmpl: "Server: {{ host }}:{{ port }}",
				context: map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			want:    "Server: localhost:8080",
			wantErr: false,
		},
		{
			name: "nested object access",
			args: args{
				tmpl: "IP: {{ server.ip }}, Port: {{ server.port }}",
				context: map[string]any{
					"server": map[string]any{
						"ip":   "192.168.1.100",
						"port": 80,
					},
				},
			},
			want:    "IP: 192.168.1.100, Port: 80",
			wantErr: false,
		},
		{
			name: "if condition",
			args: args{
				tmpl: "{% if debug %}Debug mode{% else %}Production mode{% endif %}",
				context: map[string]any{
					"debug": true,
				},
			},
			want:    "Debug mode",
			wantErr: false,
		},
		{
			name: "for loop",
			args: args{
				tmpl: "Items:\n{% for item in items %}- {{ item }}\n{% endfor %}",
				context: map[string]any{
					"items": []string{"apple", "banana", "orange"},
				},
			},
			want:    "Items:\n- apple\n- banana\n- orange\n",
			wantErr: false,
		},
		{
			name: "empty template",
			args: args{
				tmpl:    "",
				context: map[string]any{},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "template with missing variable",
			args: args{
				tmpl: "Hello {{ missing_var }}!",
				context: map[string]any{
					"name": "World",
				},
			},
			want:    "",
			wantErr: true, // Should error when variable is missing
		},
		{
			name: "invalid template syntax",
			args: args{
				tmpl:    "Hello {{ name",
				context: map[string]any{"name": "World"},
			},
			want:    "",
			wantErr: true, // Should error with invalid syntax
		},
		{
			name: "complex nested structure",
			args: args{
				tmpl: `Services:
{% for service in services %}
- {{ service.name }}: {{ service.host }}:{{ service.port }}
{% if service.enabled %}  [ENABLED]{% endif %}
{% endfor %}`,
				context: map[string]any{
					"services": []map[string]any{
						{
							"name":    "web",
							"host":    "10.0.1.10",
							"port":    80,
							"enabled": true,
						},
						{
							"name":    "api",
							"host":    "10.0.1.11",
							"port":    8080,
							"enabled": false,
						},
					},
				},
			},
			want: `Services:
- web: 10.0.1.10:80
  [ENABLED]
- api: 10.0.1.11:8080
`,
			wantErr: false,
		},
		{
			name: "nil context",
			args: args{
				tmpl:    "Hello {{ name }}!",
				context: nil,
			},
			want:    "",
			wantErr: true, // Should error with nil context
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.args.tmpl, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Render() got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRender_CustomLoopSyntax(t *testing.T) {
	type args struct {
		tmpl    string
		context map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "custom $for loop with YAML",
			args: args{
				tmpl: `servers:
  $for: "servers"
  $item: "server"
  $body:
    name: "{{ server.name }}"
    port: {{ server.port }}`,
				context: map[string]any{
					"servers": []map[string]any{
						{"name": "web", "port": 80},
						{"name": "api", "port": 8080},
					},
				},
			},
			want: `servers:
  - name: web
    port: 80
  - name: api
    port: 8080
`,
			wantErr: false,
		},
		{
			name: "custom $for loop with JSON",
			args: args{
				tmpl: `{
  "$for": "services",
  "$item": "service",
  "$body": {
    "name": "{{ service.name }}",
    "enabled": true
  }
}`,
				context: map[string]any{
					"services": []map[string]any{
						{"name": "auth"},
						{"name": "payment"},
					},
				},
			},
			want: `[
  {
    "name": "auth",
    "enabled": true
  },
  {
    "name": "payment",
    "enabled": true
  }
]`,
			wantErr: false,
		},
		{
			name: "nested custom loops",
			args: args{
				tmpl: `environments:
  $for: "environments"
  $item: "env"
  $body:
    name: "{{ env.name }}"
    servers:
      $for: "env.servers"
      $item: "server"
      $body:
        host: "{{ server.host }}"
        port: {{ server.port }}`,
				context: map[string]any{
					"environments": []map[string]any{
						{
							"name": "dev",
							"servers": []map[string]any{
								{"host": "dev1.example.com", "port": 8000},
								{"host": "dev2.example.com", "port": 8001},
							},
						},
						{
							"name": "prod",
							"servers": []map[string]any{
								{"host": "prod1.example.com", "port": 80},
							},
						},
					},
				},
			},
			want: `environments:
  - name: dev
    servers:
      - host: dev1.example.com
        port: 8000
      - host: dev2.example.com
        port: 8001
  - name: prod
    servers:
      - host: prod1.example.com
        port: 80
`,
			wantErr: false,
		},
		{
			name: "custom loop with empty array",
			args: args{
				tmpl: `items:
  $for: "items"
  $item: "item"
  $body:
    value: "{{ item }}"`,
				context: map[string]any{
					"items": []map[string]any{},
				},
			},
			want: `items: []
`,
			wantErr: false,
		},
		{
			name: "custom loop with missing array",
			args: args{
				tmpl: `items:
  $for: "missing_items"
  $item: "item"
  $body:
    value: "{{ item }}"`,
				context: map[string]any{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.args.tmpl, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Render() got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandler_Render(t *testing.T) {
	handler := New()

	type args struct {
		tmpl    string
		context map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "handler basic test",
			args: args{
				tmpl: "Value: {{ test }}",
				context: map[string]any{
					"test": "success",
				},
			},
			want:    "Value: success",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handler.Render(tt.args.tmpl, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Handler.Render() got = %q, want %q", got, tt.want)
			}
		})
	}
}
