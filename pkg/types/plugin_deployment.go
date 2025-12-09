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

	"github.com/google/uuid"
)

// PluginDeployment this is the info for plugin deployment.
type PluginDeployment struct {
	Token      string
	Info       *PluginDeploymentInfo
	PluginConf *PluginDeploymentPluginConf
}

// NewPluginDeployment new a plugin deployment.
func NewPluginDeployment(info *PluginDeploymentInfo, conf *PluginDeploymentPluginConf) *PluginDeployment {
	return &PluginDeployment{
		Token:      strings.ReplaceAll(uuid.New().String(), "-", ""),
		Info:       info,
		PluginConf: conf,
	}
}

// PluginDeploymentPluginConf defines the plugin config.
type PluginDeploymentPluginConf struct {
	// TemplateRenderer is the template renderer for plugin process.
	TemplateRenderer TemplateRendererType

	// ConfigFilesDetail is the config file detail for plugin process.
	ConfigFilesDetail []*PluginConfigDetail

	// SystemConfigContext is the system config context for plugin process.
	SystemConfigContext map[string]any

	// CustomConfigContext is the custom config context for plugin process.
	CustomConfigContext map[string]any
}

// PluginConfigDetail defines the plugin config detail.
type PluginConfigDetail struct {
	// Name is the config file name.
	Name string

	// Content is the config file content.
	Content string

	// IsMainConfig indicates whether it is the main configuration file.
	IsMainConfig bool
}

// PluginDeploymentInfo defines the plugin deployment info.
type PluginDeploymentInfo struct {
	BlockingActionName string

	// Process is the process info.
	Process Process

	// InstallerWorkDir is used to store the installation files.
	InstallerWorkDir string

	// InstallOptions is used to control the tools when install plugin.
	InstallOptions PluginDeploymentInstallOptions

	// TransferOptions is used to control the tools when transfer plugin.
	TransferOptions PluginDeploymentTransferOptions
}

// PluginDeploymentInstallOptions defines the options for plugin deployment.
type PluginDeploymentInstallOptions struct {
	Version string
}

// PluginDeploymentTransferOptions defines the options for plugin deployment.
type PluginDeploymentTransferOptions struct {
	// SelectDownloads set false by default, will download all things.
	// set true, then will only download the enabled ones following.
	SelectDownloads bool

	EnableReleasePackage bool
	EnableInstaller      bool
}
