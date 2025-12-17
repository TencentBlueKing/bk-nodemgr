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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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

// PluginDeploymentParam defines the parameters for plugin deployment.
type PluginDeploymentParam struct {
	HostID              int64
	PluginName          string
	Version             string
	ConfigName          []string
	CustomConfigContext map[string]any
}

// Validate validates the plugin deployment param.
func (p *PluginDeploymentParam) Validate() error {
	if p.HostID <= 0 {
		return fmt.Errorf("invalid HostID: %d", p.HostID)
	}

	if p.PluginName == "" {
		return fmt.Errorf("empty PluginName: %s", p.PluginName)
	}

	if p.Version == "" {
		return fmt.Errorf("empty Version: %s", p.Version)
	}

	return nil
}

// NewPluginDeploymentsByParams create base plugin deployments by params.
func NewPluginDeploymentsByParams(tenantID string, params ...*PluginDeploymentParam) ([]*PluginDeployment, []int64, error) {
	pluginDeployments := make([]*PluginDeployment, 0, len(params))
	for _, param := range params {
		if err := param.Validate(); err != nil {
			return nil, nil, err
		}

		conf := &PluginDeploymentPluginConf{
			ConfigFilesDetail:   make([]*PluginConfigDetail, 0, len(param.ConfigName)),
			CustomConfigContext: param.CustomConfigContext,
		}
		for _, item := range param.ConfigName {
			conf.ConfigFilesDetail = append(conf.ConfigFilesDetail, &PluginConfigDetail{Name: item})
		}

		deploymentInfo := &PluginDeploymentInfo{
			Process: Process{
				TenantID:   tenantID,
				HostID:     param.HostID,
				PluginName: param.PluginName,
			},
			InstallOptions: PluginDeploymentInstallOptions{
				Version: param.Version,
			},
			TransferOptions: PluginDeploymentTransferOptions{
				SelectDownloads:      true,
				EnableReleasePackage: false,
				EnableInstaller:      true,
			},
		}

		pluginDeployments = append(pluginDeployments, NewPluginDeployment(deploymentInfo, conf))
	}

	hostIDMap := make(map[int64]struct{})
	for _, host := range params {
		hostIDMap[host.HostID] = struct{}{}
	}
	hostIDs := conv.MapKeyToSlice(hostIDMap)

	return pluginDeployments, hostIDs, nil
}
