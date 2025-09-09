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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/google/uuid"
)

// PluginDeployment this is the info for plugin deployment.
type PluginDeployment struct {
	Token string
	Info  *PluginDeploymentInfo
}

// NewPluginDeployment new a plugin deployment.
func NewPluginDeployment(info *PluginDeploymentInfo) *PluginDeployment {
	return &PluginDeployment{
		Token: strings.ReplaceAll(uuid.New().String(), "-", ""),
		Info:  info,
	}
}

// PluginDeploymentInfo defines the plugin deployment info.
type PluginDeploymentInfo struct {
	BlockingActionName string

	Plugin Plugin

	// InstallerWorkDir is used to store the installation files.
	InstallerWorkDir string

	// InstallOptions is used to control the tools when install plugin.
	InstallOptions PluginDeploymentInstallOptions

	// TransferOptions is used to control the tools when transfer plugin.
	TransferOptions PluginDeploymentTransferOptions

	// TargetVersion is used to control the target version for plugin.
	TargetVersion []TargetPluginVersion
}

// PluginDeploymentInstallOptions defines the options for plugin deployment.
type PluginDeploymentInstallOptions struct {
}

// PluginDeploymentTransferOptions defines the options for plugin deployment.
type PluginDeploymentTransferOptions struct {
	// SelectDownloads set false by default, will download all things.
	// set true, then will only download the enabled ones following.
	SelectDownloads bool

	EnableReleasePackage bool
	EnableInstaller      bool
}

// TargetPluginVersion defines the target version for plugin.
type TargetPluginVersion struct {
	Platform platform.Platform
	Version  string
}
