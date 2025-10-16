/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package templaterender defines a lightweight template rendering engine with common functions.
package templaterender

import (
	"fmt"
	"strings"
	"text/template"
)

// Handler is the main struct for the template rendering engine.
type Handler struct {
	funcMap template.FuncMap
}

// New create a new light template handler.
func New() *Handler {
	handler := &Handler{
		funcMap: make(template.FuncMap),
	}

	registerStringsFunctions(handler.funcMap)
	registerDefaultFunctions(handler.funcMap)
	registerConversionFunctions(handler.funcMap)
	registerReflectFunctions(handler.funcMap)
	registerListFunctions(handler.funcMap)
	registerDictFunctions(handler.funcMap)

	return handler
}

// addFunction add a custom function to the funcMap.
func addFunction(funcsMap template.FuncMap, name string, function any) {
	if _, ok := funcsMap[name]; ok {
		return
	}
	funcsMap[name] = function
}

// Render render template with data.
func (h *Handler) Render(tmpl string, data map[string]any) (string, error) {
	template, err := template.New("template").Funcs(h.funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var output strings.Builder
	if err := template.Execute(&output, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}
