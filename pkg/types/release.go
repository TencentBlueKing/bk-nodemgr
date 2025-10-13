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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// ReleaseType defines the type of release.
type ReleaseType string

const (
	// ReleaseTypeOriginAgent defines the release of origin gse agent package.
	ReleaseTypeOriginAgent ReleaseType = "origin_agent"

	// ReleaseTypeOriginServer defines the release of origin gse server package.
	ReleaseTypeOriginServer ReleaseType = "origin_server"

	// ReleaseTypeOriginOfficialPlugin defines the release of nodemgr origin official plugin package.
	ReleaseTypeOriginOfficialPlugin ReleaseType = "origin_official_plugin"

	// ReleaseTypeOriginExternalPlugin defines the release of nodemgr origin external plugin package.
	ReleaseTypeOriginExternalPlugin ReleaseType = "origin_external_plugin"

	// ReleaseTypeAgent defines the release of nodemgr agent package transformed from origin gse agent package.
	ReleaseTypeAgent ReleaseType = "agent"

	// ReleaseTypeProxy defines the release of nodemgr proxy package transformed from origin gse agent+server package.
	ReleaseTypeProxy ReleaseType = "proxy"

	// ReleaseTypeCert defines the release of nodemgr cert package.
	ReleaseTypeCert ReleaseType = "cert"

	// ReleaseTypeBinTool defines the release of nodemgr bin tool package.
	ReleaseTypeBinTool ReleaseType = "bintool"

	// ReleaseTypePluginBinTool defines the release of nodemgr plugin bin tool package.
	ReleaseTypePluginBinTool ReleaseType = "plugin_bintool"

	// ReleaseTypeOfficialPlugin defines the release of nodemgr official plugin package.
	ReleaseTypeOfficialPlugin ReleaseType = "official_plugin"

	// ReleaseTypeExternalPlugin defines the release of nodemgr external plugin package.
	ReleaseTypeExternalPlugin ReleaseType = "external_plugin"
)

// Validate validates the release type.
func (rt ReleaseType) Validate() error {
	switch rt {
	case ReleaseTypeOriginAgent,
		ReleaseTypeOriginServer,
		ReleaseTypeAgent,
		ReleaseTypeProxy,
		ReleaseTypeCert,
		ReleaseTypeBinTool,
		ReleaseTypeOriginOfficialPlugin,
		ReleaseTypeOriginExternalPlugin,
		ReleaseTypePluginBinTool,
		ReleaseTypeOfficialPlugin,
		ReleaseTypeExternalPlugin:
		return nil
	default:
		return fmt.Errorf("invalid release type, type(%s)", rt)
	}
}

// ReleaseTypeListToStringList converts a release type list to string list.
func ReleaseTypeListToStringList(releaseTypeList []ReleaseType) []string {
	data := make([]string, len(releaseTypeList))
	for idx, releaseType := range releaseTypeList {
		data[idx] = string(releaseType)
	}

	return data
}

// ConvertReleaseTypeToNodeRole convert release type to node role.
func ConvertReleaseTypeToNodeRole(releaseType ReleaseType) (NodeRole, error) {
	switch releaseType {
	case ReleaseTypeAgent:
		return NodeRoleAgent, nil
	case ReleaseTypeProxy:
		return NodeRoleProxy, nil
	default:
		return "", fmt.Errorf("failed to convert release type to node role, releaseType(%s)", releaseType)
	}
}

// ConvertNodeRoleToReleaseType convert role to release type.
func ConvertNodeRoleToReleaseType(role NodeRole) (ReleaseType, error) {
	switch role {
	case NodeRoleAgent:
		return ReleaseTypeAgent, nil
	case NodeRoleProxy:
		return ReleaseTypeProxy, nil
	default:
		return "", fmt.Errorf("invalid node role. role(%s)", role)
	}
}

// Release defines the release package information.
type Release struct {
	Name         string
	Generation   Generation
	Type         ReleaseType
	Version      string
	Platform     platform.Platform
	Labels       []string
	FileName     string
	MD5          string
	Enabled      bool
	AsDefault    bool
	UpdatedAt    time.Time
	Operator     string
	AdditionInfo map[string]any
}

// ReleaseAgent defines the agent, it is kind of Release.
type ReleaseAgent struct {
	Release
	ReleaseAdditionInfoAgent
}

// ReleaseAdditionInfoAgent defines the addition info of release agent.
type ReleaseAdditionInfoAgent struct {
	ConfigTemplate map[string]string
	ConfigEnviron  map[string]any
	ChangeLogEN    string
	ChangeLogZH    string
}

// ReleaseProxy defines the proxy, it is kind of Release.
type ReleaseProxy struct {
	Release
	ReleaseAdditionInfoProxy
}

// ReleaseAdditionInfoProxy defines the addition info of release proxy.
type ReleaseAdditionInfoProxy struct {
	ConfigTemplate map[string]string
	ConfigEnviron  map[string]any
	ChangeLogEN    string
	ChangeLogZH    string
}

// ReleaseCert defines the cert, it is kind of Release.
type ReleaseCert struct {
	Release
}

// ReleaseBinTool defines the bin tool, it is kind of Release.
type ReleaseBinTool struct {
	Release
}

// ReleasePluginBinTool defines the plugin bin tool, it is kind of Release.
type ReleasePluginBinTool struct {
	Release
}

// ReleaseOfficialPlugin defines the official plugin, it is kind of Release.
type ReleaseOfficialPlugin struct {
	Release
	ReleaseAdditionInfoOfficialPlugin
}

// ReleaseAdditionInfoOfficialPlugin defines the addition info of release official plugin.
type ReleaseAdditionInfoOfficialPlugin struct {
	ConfigTemplates  []PluginPkgConfigTemplate
	PluginController PluginController
}

// ReleaseExternalPlugin defines the external plugin, it is kind of Release.
type ReleaseExternalPlugin struct {
	Release
	ReleaseAdditionInfoExternalPlugin
}

// ReleaseAdditionInfoExternalPlugin defines the addition info of release external plugin.
type ReleaseAdditionInfoExternalPlugin struct {
	ConfigTemplates  []PluginPkgConfigTemplate
	PluginController PluginController
}

// PluginPkgConfigTemplate defines the template of plugin package.
type PluginPkgConfigTemplate struct {
	PluginVersion string
	Name          string
	Version       string
	FilePath      string
	Format        string
	IsMainConfig  bool
	SourcePath    string
	SourceContent string
	Variables     *PluginPkgConfigTemplateProperty
}

// PluginPkgConfigTemplateProperty defines the template of plugin package.
type PluginPkgConfigTemplateProperty struct {
	Title      string                                      `yaml:"title,omitempty"`
	Type       string                                      `yaml:"type,omitempty"`
	Required   bool                                        `yaml:"required,omitempty"`
	Default    any                                         `yaml:"default,omitempty"`
	Items      *PluginPkgConfigTemplateProperty            `yaml:"items,omitempty"`
	Properties map[string]*PluginPkgConfigTemplateProperty `yaml:"properties,omitempty"`
}

// PluginController defines the control of plugin.
type PluginController struct {
	StartCmd   string
	StopCmd    string
	RestartCmd string
	ReloadCmd  string
	KillCmd    string
	VersionCmd string
	HealthCmd  string
}
