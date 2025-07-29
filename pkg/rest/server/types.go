/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package server

import (
	"io"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// RequestBody rest request body.
type RequestBody interface {
	// Validate validate request body.
	Validate() error

	// AutoConvert auto convert some fields in request body.
	AutoConvert()
}

// Response standard response.
type Response struct {
	Code       resterrf.Code `json:"code"`
	Message    string        `json:"message,omitempty"`
	RequestID  string        `json:"request_id"`
	Data       interface{}   `json:"data,omitempty"`
	Error      *Error        `json:"error,omitempty"`
	Permission *Permission   `json:"permission,omitempty"`
}

// Error defines the error struct.
type Error struct {
	System  string   `json:"system"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
}

// Detail defines the error detail.
type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Permission defines the permission struct.
type Permission struct {
	System     string   `json:"system"`
	SystemName string   `json:"system_name"`
	Actions    []Action `json:"actions"`
}

// Action defines the action struct.
type Action struct {
	// TODO: 支持 IAM 权限
}

// FileResponse file response.
type FileResponse struct {
	// Data defines file data.
	Data io.ReadCloser

	// Size defines file size.
	Size int64

	// FilePath defines file path.
	FilePath string

	// FileName defines file name.
	FileName string

	// ContentType defines content type.
	ContentType string

	// Headers defines custom headers.
	Headers map[string]string
}
