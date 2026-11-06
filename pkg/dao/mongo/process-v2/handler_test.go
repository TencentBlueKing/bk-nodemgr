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

package processv2

import (
	"testing"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestConvProcessKeepsRuntimeFields(t *testing.T) {
	process := &types.Process{
		TenantID:      "tenant-id",
		HostID:        1,
		BizID:         2,
		PluginName:    "bkmonitorbeat",
		PluginPkgName: "bkmonitorbeat",
		PluginGroup:   "bkmonitorbeat",
		Platform:      platfmt.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64},
		Generation:    types.Generation2,
		Info: types.ProcessInfo{
			Pid:       4,
			Version:   "1.0.0",
			AgentID:   "agent-id",
			AutoStart: true,
			Status:    types.ProcessStatusRunning,
		},
		Identity: types.ProcessIdentity{
			Name:       "bkmonitorbeat",
			SetupPath:  "/setup",
			PidPath:    "/pid",
			ConfigPath: "/config",
			LogPath:    "/log",
			User:       "root",
		},
		Controller: types.ProcessController{
			StartCmd:   "start.sh",
			StopCmd:    "stop.sh",
			RestartCmd: "restart.sh",
			ReloadCmd:  "reload.sh",
			DebugCmd:   "debug.sh",
			KillCmd:    "kill.sh",
			VersionCmd: "version.sh",
			HealthCmd:  "health.sh",
		},
		Resource: types.ProcessResource{
			CPULimitPercent: 1,
			MemLimitPercent: 2,
		},
		MonitorPolicy: types.ProcessMonitorPolicy{
			RestartType:    types.ProcessRestartTypeAuto,
			StartCheckSecs: 5,
			StopCheckSecs:  6,
			OpTimeoutSecs:  7,
		},
	}

	got := convertProcessToTypes(convProcessFromTypes(process))

	require.Equal(t, process, got)
}

func TestConvProcessInfoFromTypesKeepsTrusteeship(t *testing.T) {
	info := convProcessInfoFromTypes(types.ProcessInfo{
		Pid:       1,
		Version:   "1.0.0",
		AgentID:   "agent-id",
		AutoStart: true,
		Status:    types.ProcessStatusUnknown,
	})

	require.True(t, info.Trusteeship)
	require.Equal(t, string(types.ProcessStatusUnknown), info.Status)
}

func TestProcessUniqueKey(t *testing.T) {
	process := &Process{HostID: 1, Name: "bkmonitorbeat"}

	require.Equal(t, "1:bkmonitorbeat", process.UniqueKey())
	require.Equal(t, []string{FieldKeyHostID, FieldKeyPluginName}, process.UniqueFields())
}
