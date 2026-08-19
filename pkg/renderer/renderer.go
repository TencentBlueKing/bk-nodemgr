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
