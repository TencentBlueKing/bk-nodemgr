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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
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

	// UploadCategoryOriginPluginBinTool represents the origin plugin bin tool.
	UploadCategoryOriginPluginBinTool UploadCategory = "origin_plugin_bin_tool"

	// UploadCategoryOriginPluginV2 represents the origin plugin v2.
	UploadCategoryOriginPluginV2 UploadCategory = "origin_plugin_v2"

	// UploadCategoryOriginExternalPluginV2 represents the origin external plugin v2.
	UploadCategoryOriginExternalPluginV2 UploadCategory = "origin_external_plugin_v2"
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
	Platforms      []platfmt.Platform
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
	AgentPlatforms []platfmt.Platform
	ProxyPlatforms []platfmt.Platform
}

// OriginPluginBinToolPkgDetail defines the detail of plugin bin tool package.
type OriginPluginBinToolPkgDetail struct {
	fileiface.FileInfo

	UploadID string
	Existed  bool
	V2       OriginPluginBinToolPkgV2Info
	V3       OriginPluginBinToolPkgV3Info
}

// OriginPluginBinToolPkgV2Info defines the v2 info of plugin bin tool package.
type OriginPluginBinToolPkgV2Info struct {
	Platforms []platfmt.Platform
}

// OriginPluginBinToolPkgV3Info defines the v3 info of plugin bin tool package.
type OriginPluginBinToolPkgV3Info struct {
	Platforms []platfmt.Platform
}

// OriginPluginV2PkgDetail defines the detail of plugin package.
type OriginPluginV2PkgDetail struct {
	fileiface.FileInfo

	UploadID string
	Existed  bool

	PluginPkgName string
	Version       string
	Description   string
	Scenario      string
	ConfigFile    string
	ConfigFormat  string
	LaunchNode    string

	// key: platform.String()
	ConfigTemplates map[string][]PluginPkgConfigTemplate
	Controller      map[string]ProcessController

	Platforms []platfmt.Platform
}

// NewOriginPluginV2PkgDetail creates a new OriginPluginV2PkgDetail.
func NewOriginPluginV2PkgDetail() *OriginPluginV2PkgDetail {
	return &OriginPluginV2PkgDetail{
		ConfigTemplates: make(map[string][]PluginPkgConfigTemplate),
		Controller:      make(map[string]ProcessController),
	}
}

// OriginExternalPluginV2PkgDetail defines the detail of external plugin package.
type OriginExternalPluginV2PkgDetail struct {
	fileiface.FileInfo

	UploadID string
	Existed  bool

	PluginPkgName string
	Version       string
	Description   string
	Scenario      string
	ConfigFile    string
	ConfigFormat  string
	LaunchNode    string
	SubDirPaths   map[string]map[string]struct{}

	// key: platform.String()
	ConfigTemplates map[string][]PluginPkgConfigTemplate
	Controller      map[string]ProcessController

	Platforms []platfmt.Platform
}

// NewOriginExternalPluginV2PkgDetail creates a new OriginExternalPluginV2PkgDetail.
func NewOriginExternalPluginV2PkgDetail() *OriginExternalPluginV2PkgDetail {
	return &OriginExternalPluginV2PkgDetail{
		SubDirPaths:     make(map[string]map[string]struct{}),
		ConfigTemplates: make(map[string][]PluginPkgConfigTemplate),
		Controller:      make(map[string]ProcessController),
	}
}
