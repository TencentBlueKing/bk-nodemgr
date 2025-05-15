/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rest defines rest context.
package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/gin-gonic/gin"
)

// Context rest context.
type Context struct {
	gCtx      *gin.Context
	RequestID string `json:"request_id"`
	Username  string `json:"username"`
	TenantID  string `json:"tenant_id"`
}

// BindJSON bind json
func (c *Context) BindJSON(body RequestBody) error {
	if err := c.gCtx.BindJSON(body); err != nil {
		return err
	}

	// auto convert some fields in request body.
	body.AutoConvert()

	return body.Validate()
}

// ParseFileForm bind file form.
func (c *Context) ParseFileForm(body RequestBody) (*multipart.FileHeader, error) {
	metaData := c.gCtx.PostForm("metadata")
	if err := json.Unmarshal([]byte(metaData), body); err != nil {
		return nil, fmt.Errorf("failed to parse metadata(%s), err(%v)", metaData, err)
	}

	file, err := c.gCtx.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to parse file, err(%v)", err)
	}

	return file, nil
}

// Param parse the param from url.
func (c *Context) Param(key string) string {
	return c.gCtx.Param(key)
}

// GetContext get a generic context from rest-context.
func (c *Context) GetContext() (context.Context, error) {
	ctx, err := identifier.SetRequestID(c.gCtx.Request.Context(), c.RequestID)
	if err != nil {
		return nil, err
	}

	return tenant.SetID(ctx, c.TenantID)
}

// RequestBody rest request body.
type RequestBody interface {
	// Validate validate request body.
	Validate() error

	// AutoConvert auto convert some fields in request body.
	AutoConvert()
}
