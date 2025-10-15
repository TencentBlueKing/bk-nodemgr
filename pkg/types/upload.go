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
	"time"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// UploadCategory defines the category of upload.
type UploadCategory string

const (
	// UploadCategoryOriginAgent represents the origin agent.
	UploadCategoryOriginAgent UploadCategory = "origin_agent"

	// UploadCategoryOriginServer represents the origin server.
	UploadCategoryOriginServer UploadCategory = "origin_server"

	// UploadCategoryOriginCert represents the origin cert.
	UploadCategoryOriginCert UploadCategory = "origin_cert"

	// UploadCategoryOriginBinTool represents the origin bin tool.
	UploadCategoryOriginBinTool UploadCategory = "origin_bin_tool"

	// UploadCategoryOriginPluginBinTool represents the origin bin tool.
	UploadCategoryOriginPluginBinTool UploadCategory = "origin_plugin_bin_tool"

	// UploadCategoryOriginOfficialPlugin represents the origin official plugin.
	UploadCategoryOriginOfficialPlugin UploadCategory = "origin_official_plugin"

	// UploadCategoryOriginExternalPlugin represents the origin external plugin.
	UploadCategoryOriginExternalPlugin UploadCategory = "origin_external_plugin"
)

// Upload defines the upload struct.
type Upload struct {
	UploadID  string
	Category  UploadCategory
	SavedName string
	Operator  string
	CreatedAt time.Time
}

// OriginPkgDetail defines the detail of origin package.
type OriginPkgDetail struct {
	fileiface.FileInfo

	UploadID       string
	Existed        bool
	Version        string
	Platforms      []platform.Platform
	ChangeLogEN    string
	ChangeLogZH    string
	ConfigTemplate map[string]string
	ConfigEnviron  map[string]any
}

// NewOriginPkgDetail creates a new OriginPkgDetail.
func NewOriginPkgDetail() *OriginPkgDetail {
	return &OriginPkgDetail{
		ConfigTemplate: make(map[string]string),
		ConfigEnviron:  make(map[string]any),
	}
}

// OriginCertPkgDetail defines the detail of cert package.
type OriginCertPkgDetail struct {
	fileiface.FileInfo

	UploadID  string
	Existed   bool
	CertFiles []string
}

// OriginBinToolPkgDetail defines the detail of bin tool package.
type OriginBinToolPkgDetail struct {
	fileiface.FileInfo

	UploadID       string
	Existed        bool
	AgentPlatforms []platform.Platform
	ProxyPlatforms []platform.Platform
}

// OriginPluginBinToolPkgDetail defines the detail of plugin bin tool package.
type OriginPluginBinToolPkgDetail struct {
	fileiface.FileInfo

	UploadID  string
	Existed   bool
	Platforms []platform.Platform
}

// OriginOfficialPluginPkgDetail defines the detail of official plugin package.
type OriginOfficialPluginPkgDetail struct {
	fileiface.FileInfo

	UploadID string
	Existed  bool

	Name            string
	Version         string
	Description     string
	Scenario        string
	ConfigFile      string
	ConfigFormat    string
	LaunchNode      string
	ConfigTemplates []PluginPkgConfigTemplate

	Controller ProcessController

	Platforms []platform.Platform
}

// OriginExternalPluginPkgDetail defines the detail of external plugin package.
type OriginExternalPluginPkgDetail struct {
	fileiface.FileInfo

	UploadID string
	Existed  bool

	Name            string
	Version         string
	Description     string
	Scenario        string
	ConfigFile      string
	ConfigFormat    string
	LaunchMode      string
	SubDirPaths     map[string]map[string]struct{}
	ConfigTemplates []PluginPkgConfigTemplate

	Controller ProcessController

	Platforms []platform.Platform
}
