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

package installer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginInstallCPULimitScript(t *testing.T) {
	base := PluginInstallParams{
		PluginCommonParams: PluginCommonParams{
			InstallWorkDir:    "/opt",
			InstallerFileName: "installer",
			BaseDeployDir:     "/data",
			BaseWorkDir:       "/var/lib",
			DeployEnv:         "test",
		},
		PluginGroup:   "group",
		PluginName:    "demo",
		PluginVersion: "1.0.0",
		PluginPkgName: "demo",
		DeployToken:   "token",
		OperInstID:    "operation",
		SkipCallback:  true,
		SkipDownload:  true,
	}

	tests := []struct {
		name       string
		limit      float64
		contains   string
		notContain string
	}{
		{name: "unset", limit: 0, notContain: "--cpu-limit"},
		{name: "one", limit: 1, notContain: "--cpu-limit"},
		{name: "throttled", limit: 0.5, contains: "--cpu-limit 0.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := base
			params.InstallCPULimit = tt.limit
			_, script, err := params.ToUnixScript()
			require.NoError(t, err)
			if tt.contains != "" {
				require.Contains(t, script, tt.contains)
			}
			if tt.notContain != "" {
				require.NotContains(t, script, tt.notContain)
			}
		})
	}
}

func TestPluginUpgradeCPULimitScript(t *testing.T) {
	params := PluginUpgradeParams{
		PluginCommonParams: PluginCommonParams{
			InstallWorkDir:    "/opt",
			InstallerFileName: "installer",
			BaseDeployDir:     "/data",
			BaseWorkDir:       "/var/lib",
			DeployEnv:         "test",
		},
		PluginGroup:     "group",
		PluginName:      "demo",
		PluginVersion:   "1.0.0",
		PluginPkgName:   "demo",
		InstallCPULimit: 0.5,
		DeployToken:     "token",
		OperInstID:      "operation",
		SkipCallback:    true,
		SkipDownload:    true,
	}

	_, unixScript, err := params.ToUnixScript()
	require.NoError(t, err)
	require.Contains(t, unixScript, "--cpu-limit 0.5")

	_, windowsScript, err := params.ToWindowsScript()
	require.NoError(t, err)
	require.Contains(t, windowsScript, "--cpu-limit 0.5")
}

func TestValidateInstallCPULimit(t *testing.T) {
	for _, limit := range []float64{-0.1, 1.1, math.NaN(), math.Inf(1)} {
		require.Error(t, validateInstallCPULimit(limit))
	}
	for _, limit := range []float64{0, 0.5, 1} {
		require.NoError(t, validateInstallCPULimit(limit))
	}
}
