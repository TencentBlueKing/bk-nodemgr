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

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func pluginProgramName(pkgName string, osType criteria.OSType) string {
	if osType == criteria.OSWindows {
		return pkgName + ".exe"
	}
	return pkgName
}

// buildPluginProcessSpec assembles shared defaults; callers own release selection and policy overrides.
func buildPluginProcessSpec(process types.Process, host *types.Host, pluginPkg *types.ReleasePlugin,
	runtime types.PluginDeploymentBaseRuntime,
) types.ProcessSpec {
	osType := process.Platform.OS
	var mainConfigPath string
	for _, configTemplate := range pluginPkg.ConfigTemplates {
		if !configTemplate.IsMainConfig {
			continue
		}
		mainConfigPath = tool.JoinPath(osType, runtime.PluginHomeDir, configTemplate.FilePath, configTemplate.Name)
		break
	}

	user := host.Dynamic.LoginUser
	if host.Dynamic.NodeOsType == criteria.OSWindows {
		user = gse.WindowsOperateUser
	}

	// nolint: mnd
	return types.ProcessSpec{
		PluginName: process.PluginName,
		AgentID:    host.Dynamic.AgentID,
		Identity: types.ProcessIdentity{
			Name:       pluginProgramName(process.PluginPkgName, osType),
			SetupPath:  runtime.PluginHomeDir,
			PidPath:    tool.JoinPath(osType, runtime.RunDir, process.PluginPkgName+".pid"),
			ConfigPath: mainConfigPath,
			LogPath:    runtime.LogDir,
			User:       user,
		},
		Controller: pluginPkg.PluginController,
		Resource: types.ProcessResource{
			CPULimitPercent: 10,
			MemLimitPercent: 10,
		},
		MonitorPolicy: types.ProcessMonitorPolicy{
			RestartType:    types.ProcessRestartTypeAuto,
			StartCheckSecs: 5,
			StopCheckSecs:  5,
			OpTimeoutSecs:  5,
		},
	}
}
