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

package types

import (
	"errors"
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
	// Release is the published release package.
	Release []Release
	// Upload is the result of uploaded plugin package.
	Upload PackageDeploymentUploadInfo
	// ImportPluginPkgOptions provides the options for importing a plugin package.
	ImportPluginPkgOptions PackageImportPluginPkgOptions
	// ExportPluginPkgOptions provides the options for exporting a plugin package.
	ExportPluginPkgOptions PackageExportPluginPkgOptions
}

// PackageDeploymentUploadInfo represents uploaded plugin package information.
type PackageDeploymentUploadInfo struct {
	UploadID  string
	Name      string
	Version   string
	Platforms []platfmt.Platform
}

// Validate checks the upload info fields are all filled.
func (upload *PackageDeploymentUploadInfo) Validate() error {
	if upload.UploadID == "" {
		return errors.New("upload id is empty")
	}
	if upload.Name == "" {
		return errors.New("plugin name is empty")
	}
	if upload.Version == "" {
		return errors.New("plugin version is empty")
	}
	if len(upload.Platforms) == 0 {
		return errors.New("plugin platforms is empty")
	}

	return nil
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
	FileName       string
	MD5            string
}

// PackageExportPluginPkgOptions represents the options of exporting a plugin package.
type PackageExportPluginPkgOptions struct {
	PluginPkgName    string
	PluginPkgVersion string
}
