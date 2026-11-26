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

// Package jinja2x provides jinja2x render.
package jinja2x

import (
	"context"
	"time"
)

const renderTimeout = time.Minute

// Handler the handler for jinja2x.
type Handler struct {
}

// Render implement IHandler.
func (h Handler) Render(tmpl string, content map[string]any) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), renderTimeout)
	defer cancel()

	client, err := defaultServer.getClient(ctx)
	if err != nil {
		return "", err
	}

	return client.render(ctx, tmpl, content)
}

// New create a new Handler.
func New() *Handler {
	return &Handler{}
}

// nolint: gochecknoglobals
var handler = New()

// Render render the template with context.
func Render(tmpl string, context map[string]any) (string, error) {
	return handler.Render(tmpl, context)
}
