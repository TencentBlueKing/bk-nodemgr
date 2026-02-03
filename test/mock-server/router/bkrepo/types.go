/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package bkrepo provides mock BKRepo server types.
package bkrepo

import "fmt"

// BKRepo-standard error codes. Reference: BKRepo API error codes.
const (
	// CodeOK is the success code.
	CodeOK = 0

	// CodeInvalidParameter is the invalid parameter code.
	CodeInvalidParameter = 251000

	// CodeServerError is the server internal error code.
	CodeServerError = 251001

	// CodeNodeNotFound is the node not found error code.
	CodeNodeNotFound = 251010
)

const (
	// HeaderOverwrite is the header name for overwrite flag.
	HeaderOverwrite = "X-BKREPO-OVERWRITE"
)

// Config holds the configuration for BKRepo mock data.
type Config struct {
	// BaseDir is the root directory for storing files.
	BaseDir string `yaml:"baseDir"`
}

// Validate validates the Config.
func (cfg *Config) Validate() error {
	if cfg.BaseDir == "" {
		return fmt.Errorf("baseDir cannot be empty")
	}

	return nil
}

// PathParams holds the path parameters for BKRepo API.
type PathParams struct {
	Project string `uri:"project" binding:"required"`
	Repo    string `uri:"repo" binding:"required"`
	Path    string `uri:"path" binding:"required"`
}

// ListNodeParams holds the parameters for ListNode API.
type ListNodeParams struct {
	PathParams
	PageNum  int `form:"pageNum" binding:"required"`
	PageSize int `form:"pageSize" binding:"required"`
}
