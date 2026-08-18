/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"strings"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/google/uuid"
)

// NewPackageDeployment creates a new package deployment.
func NewPackageDeployment(info *PackageDeploymentInfo) *PackageDeployment {
	return &PackageDeployment{
		Token: strings.ReplaceAll(uuid.New().String(), "-", ""),
		Info:  info,
	}
}

// PackageDeployment represents a package deployment record.
type PackageDeployment struct {
	Token string
	Info  *PackageDeploymentInfo
}

// PackageDeploymentInfo represents the info of a package deployment.
type PackageDeploymentInfo struct {
	UploadID               string
	ImportPluginPkgOptions PackageImportPluginPkgOptions
}

// FileSourceType represents the type of file source.
type FileSourceType string

const (
	// FileSourceTypeDownload represents the file source is download.
	FileSourceTypeDownload FileSourceType = "download"
)

// PackageImportPluginPkgOptions represents the options of importing a plugin package.
type PackageImportPluginPkgOptions struct {
	FileSourceType FileSourceType
	FileSource     string
	MD5            string
	PluginPkgName  string
	PluginName     string
	Version        string
	Platforms      []platfmt.Platform
}
