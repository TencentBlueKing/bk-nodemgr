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

package renderer

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/gotemplate"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/jinja2x"
)

// TestRenderNestedMapGoTemplate checks that gotemplate strings nested in maps and lists are rendered.
func TestRenderNestedMapGoTemplate(t *testing.T) {
	nested := map[string]any{
		"listen": "{{ .inner_ip }}:{{ .listen_port }}",
		"plain":  "no placeholder",
		"count":  5,
		"items":  []any{"{{ .inner_ip }}-1", 2, true},
		"nested": map[string]any{"name": "{{ .plugin_name }}"},
	}
	context := map[string]any{
		"inner_ip":    "127.0.0.1",
		"listen_port": 8000,
		"plugin_name": "bkmonitorbeat",
	}

	rendered, err := RenderNestedMap(gotemplate.New(), nested, context)
	if err != nil {
		t.Fatalf("RenderNestedMap() error = %v", err)
	}

	want := map[string]any{
		"listen": "127.0.0.1:8000",
		"plain":  "no placeholder",
		"count":  5,
		"items":  []any{"127.0.0.1-1", 2, true},
		"nested": map[string]any{"name": "bkmonitorbeat"},
	}
	if !reflect.DeepEqual(rendered, want) {
		t.Errorf("RenderNestedMap() = %v, want %v", rendered, want)
	}

	// the input map must stay untouched, callers persist it as user input.
	if nested["listen"] != "{{ .inner_ip }}:{{ .listen_port }}" {
		t.Errorf("RenderNestedMap() modified the input map: %v", nested)
	}
}

// TestRenderNestedMapJinja2 checks that jinja2 strings nested in maps and lists are rendered.
func TestRenderNestedMapJinja2(t *testing.T) {
	// the jinja2 executable is only embedded for linux.
	if runtime.GOOS != "linux" {
		t.Skipf("jinja2x is not supported on %s", runtime.GOOS)
	}

	nested := map[string]any{
		"listen": "{{ inner_ip }}:{{ listen_port }}",
		"plain":  "no placeholder",
		"items":  []any{"{{ inner_ip }}-1", 2},
	}
	context := map[string]any{
		"inner_ip":    "127.0.0.1",
		"listen_port": 8000,
	}

	rendered, err := RenderNestedMap(jinja2x.New(), nested, context)
	if err != nil {
		t.Fatalf("RenderNestedMap() error = %v", err)
	}

	want := map[string]any{
		"listen": "127.0.0.1:8000",
		"plain":  "no placeholder",
		"items":  []any{"127.0.0.1-1", 2},
	}
	if !reflect.DeepEqual(rendered, want) {
		t.Errorf("RenderNestedMap() = %v, want %v", rendered, want)
	}
}

// TestRenderNestedMapRenderError checks that a render failure is returned instead of being swallowed.
func TestRenderNestedMapRenderError(t *testing.T) {
	nested := map[string]any{
		"broken": "{{ .inner_ip | undefined_function }}",
	}

	rendered, err := RenderNestedMap(gotemplate.New(), nested, map[string]any{"inner_ip": "127.0.0.1"})
	if err == nil {
		t.Fatalf("RenderNestedMap() error = nil, want error, rendered = %v", rendered)
	}
}

// TestRenderNestedMapDeepNesting checks that values nested in slices of maps and slices of slices are rendered.
func TestRenderNestedMapDeepNesting(t *testing.T) {
	nested := map[string]any{
		"listeners": []any{
			map[string]any{"ip": "{{ .inner_ip }}", "port": 8000},
			[]any{"{{ .inner_ip }}-nested"},
		},
	}

	rendered, err := RenderNestedMap(gotemplate.New(), nested, map[string]any{"inner_ip": "127.0.0.1"})
	if err != nil {
		t.Fatalf("RenderNestedMap() error = %v", err)
	}

	want := map[string]any{
		"listeners": []any{
			map[string]any{"ip": "127.0.0.1", "port": 8000},
			[]any{"127.0.0.1-nested"},
		},
	}
	if !reflect.DeepEqual(rendered, want) {
		t.Errorf("RenderNestedMap() = %v, want %v", rendered, want)
	}
}

// TestRenderNestedMapSinglePass checks that every value is rendered exactly once, a rendered value is never re-rendered.
func TestRenderNestedMapSinglePass(t *testing.T) {
	context := map[string]any{
		"a": "{{ .b }}",
		"b": "{{ .c }}",
		"c": "done",
	}

	rendered, err := RenderNestedMap(gotemplate.New(), context, context)
	if err != nil {
		t.Fatalf("RenderNestedMap() error = %v", err)
	}

	want := map[string]any{
		"a": "{{ .c }}", // "b" is rendered from the unrendered context, so "a" stops after one pass.
		"b": "done",
		"c": "done",
	}
	if !reflect.DeepEqual(rendered, want) {
		t.Errorf("RenderNestedMap() = %v, want %v", rendered, want)
	}
}
