//go:build integration

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
	"context"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/require"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

func testProcess(tenantID string) *types.Process {
	return &types.Process{
		TenantID:      tenantID,
		HostID:        1,
		BizID:         2,
		PluginName:    "plugin-name",
		PluginPkgName: "plugin-pkg-name",
		PluginGroup:   "plugin-group",
		Platform:      platfmt.Platform{OS: "linux", Arch: "amd64"},
		Generation:    3,
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
}

func TestHandler_ProcessIntegration(t *testing.T) {
	const tenantID = "test-tenant-id"

	nCtx := contextx.New(context.Background(), contextx.WithTenantID(tenantID))
	h := testClient(t)

	created := testProcess(tenantID)
	require.NoError(t, h.Create(nCtx, created))

	t.Run("count list get exist", func(t *testing.T) {
		count, err := h.Count(nCtx, WithHostID(created.HostID), WithPluginName(created.PluginName))
		require.NoError(t, err)
		require.Equal(t, int64(1), count)

		items, total, err := h.List(nCtx, types.UnlimitedPage(), WithHostID(created.HostID), WithPluginName(created.PluginName))
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, items, 1)
		require.Equal(t, created.PluginName, items[0].PluginName)
		require.Equal(t, created.PluginPkgName, items[0].PluginPkgName)
		require.Equal(t, created.Controller.DebugCmd, items[0].Controller.DebugCmd)
		require.True(t, reflect.DeepEqual(created, items[0]))

		got, err := h.Get(nCtx, WithHostID(created.HostID), WithPluginName(created.PluginName))
		require.NoError(t, err)
		require.True(t, reflect.DeepEqual(created, got))

		exist, err := h.Exist(nCtx, created.HostID, created.PluginName)
		require.NoError(t, err)
		require.True(t, exist)
	})

	require.NoError(t, h.Delete(nCtx, created.HostID, created.PluginName))

	t.Run("deleted record is gone", func(t *testing.T) {
		count, err := h.Count(nCtx, WithHostID(created.HostID), WithPluginName(created.PluginName))
		require.NoError(t, err)
		require.Zero(t, count)

		items, total, err := h.List(nCtx, types.UnlimitedPage(), WithHostID(created.HostID), WithPluginName(created.PluginName))
		require.NoError(t, err)
		require.Zero(t, total)
		require.Empty(t, items)

		got, err := h.Get(nCtx, WithHostID(created.HostID), WithPluginName(created.PluginName))
		require.Error(t, err)
		require.Nil(t, got)

		exist, err := h.Exist(nCtx, created.HostID, created.PluginName)
		require.NoError(t, err)
		require.False(t, exist)
	})
}
