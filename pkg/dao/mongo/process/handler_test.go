/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package process

import (
	"reflect"
	"testing"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestConvProcessFromTypesKeepsDebugCmd(t *testing.T) {
	process := &types.Process{
		TenantID:      "tenant-id",
		HostID:        1,
		BizID:         2,
		PluginName:    "plugin-name",
		PluginPkgName: "plugin-pkg-name",
		PluginGroup:   "plugin-group",
		Platform:      platfmt.Platform{},
		Generation:    3,
		BindIP:        "127.0.0.1",
		BindPort:      10000,
		Info: types.ProcessInfo{
			Pid:       4,
			Version:   "1.0.0",
			AgentID:   "agent-id",
			AutoStart: true,
			Status:    types.ProcessStatusRunning,
		},
		Identity: types.ProcessIdentity{
			Name:       "proc-name",
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

	got := convProcessFromTypes(process)
	require.Equal(t, "debug.sh", got.Controller.DebugCmd)
	require.Equal(t, "127.0.0.1", got.BindIP)
	require.Equal(t, 10000, got.BindPort)
}

func TestConvertProcessToTypesKeepsDebugCmd(t *testing.T) {
	data := &Process{
		TenantID:      "tenant-id",
		HostID:        1,
		BizID:         2,
		Name:          "plugin-name",
		PluginPkgName: "plugin-pkg-name",
		Group:         "plugin-group",
		Generation:    3,
		Platform: platform{
			OS:   "linux",
			Arch: "amd64",
		},
		BindIP:   "127.0.0.1",
		BindPort: 10000,
		Info: processInfo{
			Pid:         4,
			Version:     "1.0.0",
			AgentID:     "agent-id",
			Trusteeship: true,
			Status:      string(types.ProcessStatusRunning),
		},
		Identity: processIdentity{
			Name:       "proc-name",
			SetupPath:  "/setup",
			PidPath:    "/pid",
			ConfigPath: "/config",
			LogPath:    "/log",
			User:       "root",
		},
		Controller: processController{
			StartCmd:   "start.sh",
			StopCmd:    "stop.sh",
			RestartCmd: "restart.sh",
			ReloadCmd:  "reload.sh",
			DebugCmd:   "debug.sh",
			KillCmd:    "kill.sh",
			VersionCmd: "version.sh",
			HealthCmd:  "health.sh",
		},
		Resource: processResource{
			CPULimitPercent: 1,
			MemLimitPercent: 2,
		},
		MonitorPolicy: processMonitorPolicy{
			RestartType:    string(types.ProcessRestartTypeAuto),
			StartCheckSecs: 5,
			StopCheckSecs:  6,
			OpTimeoutSecs:  7,
		},
	}

	got := convertProcessToTypes(data)
	require.Equal(t, "debug.sh", got.Controller.DebugCmd)
	require.Equal(t, "127.0.0.1", got.BindIP)
	require.Equal(t, 10000, got.BindPort)
	require.True(t, reflect.DeepEqual(data, convProcessFromTypes(got)))
}
