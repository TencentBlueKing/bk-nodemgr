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

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// ReleaseType defines the type of release.
type ReleaseType string

const (
	// ReleaseTypeOriginAgent defines the release of origin gse agent package.
	ReleaseTypeOriginAgent ReleaseType = "origin_agent"

	// ReleaseTypeOriginServer defines the release of origin gse server package.
	ReleaseTypeOriginServer ReleaseType = "origin_server"

	// ReleaseTypeOriginPluginV2 defines the release of nodemgr origin plugin v2 package.
	ReleaseTypeOriginPluginV2 ReleaseType = "origin_plugin_v2"

	// ReleaseTypeOriginExternalPluginV2 defines the release of nodemgr origin external plugin v2 package.
	ReleaseTypeOriginExternalPluginV2 ReleaseType = "origin_external_plugin_v2"

	// ReleaseTypeOriginPluginV3 defines the release of nodemgr origin plugin v3 package.
	ReleaseTypeOriginPluginV3 ReleaseType = "origin_plugin_v3"

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

	// ReleaseTypePlugin defines the release of nodemgr plugin package.
	ReleaseTypePlugin ReleaseType = "plugin"
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
		ReleaseTypeOriginPluginV2,
		ReleaseTypeOriginExternalPluginV2,
		ReleaseTypeOriginPluginV3,
		ReleaseTypePluginBinTool,
		ReleaseTypePlugin:
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

// StringListToReleaseTypeList converts a string list to a release type list.
func StringListToReleaseTypeList(stringList []string) []ReleaseType {
	data := make([]ReleaseType, len(stringList))
	for idx, eventType := range stringList {
		data[idx] = ReleaseType(eventType)
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

// ConvertConfigPolicyTypeToReleaseType convert config policy type to release type.
func ConvertConfigPolicyTypeToReleaseType(cpType ConfigPolicyType) (ReleaseType, error) {
	switch cpType {
	case ConfigPolicyTypeAgent:
		return ReleaseTypeAgent, nil
	case ConfigPolicyTypeProxy:
		return ReleaseTypeProxy, nil
	default:
		return "", fmt.Errorf("invalid config policy type. type(%s)", cpType)
	}
}

// Release defines the release package information.
type Release struct {
	Name         string
	Generation   Generation
	Type         ReleaseType
	Version      string
	Platform     platfmt.Platform
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

// ReleasePlugin defines the plugin, it is kind of Release.
type ReleasePlugin struct {
	Release
	ReleaseAdditionInfoPlugin
}

// ReleaseAdditionInfoPlugin defines the addition info of release plugin.
type ReleaseAdditionInfoPlugin struct {
	TemplateRendererType TemplateRendererType
	ConfigTemplates      []PluginPkgConfigTemplate
	PluginController     ProcessController
	Description          string
	DescriptionEn        string
	Scenario             string
	ScenarioEn           string
}

// TemplateRendererType defines the type of template renderer.
type TemplateRendererType string

const (
	// TemplateRendererTypeJinja2 defines the jinja2 template renderer.
	// In v2 plugin, we only support jinja2 template renderer.
	TemplateRendererTypeJinja2 TemplateRendererType = "jinja2"

	// TemplateRendererTypeGoTemplate defines the go-template template renderer.
	TemplateRendererTypeGoTemplate TemplateRendererType = "go-template"
)

// Validate validates the template renderer type.
func (t TemplateRendererType) Validate() error {
	switch t {
	case TemplateRendererTypeJinja2, TemplateRendererTypeGoTemplate:
		return nil
	default:
		return fmt.Errorf("invalid template renderer type, type(%s)", t)
	}
}

// PluginPkgConfigTemplate defines the template of plugin package.
// @Name: template name.
// @FilePath: the path where the template is located in the package.
// @SourcePath: the source path of the template.
// @IsMainConfig: whether it is the main configuration file.
// @SourceContent: the content of the template.
// @Variables: the variables used in the template.
type PluginPkgConfigTemplate struct {
	Name          string
	FilePath      string
	SourcePath    string
	IsMainConfig  bool
	SourceContent string
	Variables     map[string]*PluginPkgConfigTemplateProperty
}

// PluginPkgConfigTemplateProperty defines the template of plugin package.
// @Title: property title.
// @Type: property type.
// @Required: whether the property is required.
// @Default: default value of the property.
// @Description: property description in Chinese.
// @DescriptionEn: property description in English.
// @Properties: nested properties.
type PluginPkgConfigTemplateProperty struct {
	Title         string
	Type          string
	Required      bool
	Default       any
	Description   string
	DescriptionEn string
	Properties    map[string]*PluginPkgConfigTemplateProperty
}

// PackageEventType represents the event type of watch event.
type PackageEventType string

const (
	// PackageEventTypePublish represents the publish event type.
	PackageEventTypePublish PackageEventType = "publish"

	// PackageEventTypeDelete represents the delete event type.
	PackageEventTypeDelete PackageEventType = "delete"

	// PackageEventTypeEnable represents the enable release event type.
	PackageEventTypeEnable PackageEventType = "enable"

	// PackageEventTypeDisable represents the disable release event type.
	PackageEventTypeDisable PackageEventType = "disable"

	// PackageEventTypeSetAsDefault represents the set as default event type.
	PackageEventTypeSetAsDefault PackageEventType = "set_as_default"

	// PackageEventTypeCancelAsDefault represents the cancel as default event type.
	PackageEventTypeCancelAsDefault PackageEventType = "cancel_as_default"

	// PackageEventTypeUpload represents the upload event type.
	PackageEventTypeUpload PackageEventType = "upload"
)

// Validate validates the eventType type.
func (eventType PackageEventType) Validate() error {
	switch eventType {
	case PackageEventTypePublish, PackageEventTypeDelete, PackageEventTypeEnable,
		PackageEventTypeDisable, PackageEventTypeSetAsDefault, PackageEventTypeCancelAsDefault:
		return nil
	default:
		return fmt.Errorf("invalid eventType type, type(%s)", eventType)
	}
}

// PackageEventTypeListToStringList converts a packageevent type list to a string list.
func PackageEventTypeListToStringList(eventTypeList []PackageEventType) []string {
	data := make([]string, len(eventTypeList))
	for idx, eventType := range eventTypeList {
		data[idx] = string(eventType)
	}

	return data
}

// StringListToPackageEventTypeList converts a string list to a packageevent type list.
func StringListToPackageEventTypeList(stringList []string) []PackageEventType {
	data := make([]PackageEventType, len(stringList))
	for idx, eventType := range stringList {
		data[idx] = PackageEventType(eventType)
	}

	return data
}

// PackageEvent represents the event of package.
type PackageEvent struct {
	Name        string
	EventType   PackageEventType
	Generation  Generation
	ReleaseType ReleaseType
	OSType      criteria.OSType
	CPUArch     criteria.CPUArch
	Version     string
	OperateTime time.Time
	Operator    string
}
