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

// Package renderer provides template rendering capabilities.
package renderer

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/gotemplate"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer/jinja2x"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler defines the interface for jinja2 template rendering handlers.
type IHandler interface {
	Render(tmpl string, context map[string]any) (string, error)
}

// NewRenderer create a new template renderer based on the specified type.
func NewRenderer(renderType types.TemplateRendererType) (IHandler, error) {
	switch renderType {
	case types.TemplateRendererTypeJinja2:
		return jinja2x.New(), nil
	case types.TemplateRendererTypeGoTemplate:
		return gotemplate.New(), nil
	default:
		return nil, fmt.Errorf("unknown template renderer: %s", renderType)
	}
}

// RenderNestedMap renders every template string nested in nested with context and keeps other values as is.
func RenderNestedMap(handler IHandler, nested, context map[string]any) (map[string]any, error) {
	rendered := make(map[string]any, len(nested))
	for key, value := range nested {
		renderedValue, err := renderNestedValue(handler, value, context, key)
		if err != nil {
			return nil, err
		}

		rendered[key] = renderedValue
	}

	return rendered, nil
}

// renderNestedValue renders the template strings inside a single value, the path locates the value in error messages.
func renderNestedValue(handler IHandler, value any, context map[string]any, path string) (any, error) {
	switch typedValue := value.(type) {
	case string:
		rendered, err := handler.Render(typedValue, context)
		if err != nil {
			return nil, fmt.Errorf("failed to render %s: %w", path, err)
		}

		return rendered, nil
	case map[string]any:
		rendered := make(map[string]any, len(typedValue))
		for key, item := range typedValue {
			renderedItem, err := renderNestedValue(handler, item, context, path+"."+key)
			if err != nil {
				return nil, err
			}

			rendered[key] = renderedItem
		}

		return rendered, nil
	case []any:
		rendered := make([]any, len(typedValue))
		for index, item := range typedValue {
			renderedItem, err := renderNestedValue(handler, item, context, fmt.Sprintf("%s[%d]", path, index))
			if err != nil {
				return nil, err
			}

			rendered[index] = renderedItem
		}

		return rendered, nil
	default:
		return value, nil
	}
}
