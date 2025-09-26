/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// templaterender defines a lightweight template rendering engine with common functions.
package templaterender

import (
	"strings"
	"text/template"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
)

type TemplateFunc func(*Handler)

type Handler struct {
	logger  logger.ILogger
	funcMap template.FuncMap
}

// New create a new light template handler
func New(opts ...TemplateFunc) *Handler {
	handler := &Handler{
		logger:  logger.LoggerDefault{},
		funcMap: make(template.FuncMap),
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func WithLogger(log logger.ILogger) TemplateFunc {
	return func(h *Handler) {
		h.logger = log
	}
}

func WithStringsFunctions() TemplateFunc {
	return func(h *Handler) {
		RegisterStringsFunctions(h.funcMap)
	}
}

func WithDefaultFunctions() TemplateFunc {
	return func(h *Handler) {
		RegisterDefaultFunctions(h.funcMap)
	}
}

func WithConversionFunctions() TemplateFunc {
	return func(h *Handler) {
		RegisterConversionFunctions(h.funcMap)
	}
}

func WithAllFunctions() TemplateFunc {
	return func(h *Handler) {
		RegisterStringsFunctions(h.funcMap)
		RegisterDefaultFunctions(h.funcMap)
		RegisterConversionFunctions(h.funcMap)
	}
}

// AddFunction add a custom function to the funcMap
func AddFunction(funcsMap template.FuncMap, name string, function any) {
	if _, ok := funcsMap[name]; ok {
		return
	}
	funcsMap[name] = function
}

// Render render template with data
func (h *Handler) Render(tmpl string, data any) (string, error) {
	t, err := template.New("template").Funcs(h.funcMap).Parse(tmpl)
	if err != nil {
		h.logger.Errorf("parse template failed: %v", err)
		return "", err
	}

	var output strings.Builder
	if err := t.Execute(&output, data); err != nil {
		h.logger.Errorf("execute template failed: %v", err)
		return "", err
	}

	return output.String(), nil
}
