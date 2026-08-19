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

package backend

import (
	"bytes"
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeCMDRegistersAgentSubcommand(t *testing.T) {
	cmd := NewNodeCMD(
		func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler { return &backendadmin.Handler{} },
		func() (string, string) { return "default", "admin" },
	)

	agentCMD := findSubcommand(cmd, "agent")
	require.NotNil(t, agentCMD)
	assert.NotNil(t, findSubcommand(agentCMD, "sync-unassigned-network-unit"))
}

func TestSyncUnassignedNetworkUnitCommandPrintsJSON(t *testing.T) {
	handler := &fakeBackendAdminHandler{syncResult: &types.NodeAgentAssignUnitResult{
		SuccessCount:  2,
		FailedReasons: []string{},
	}}

	cmd := NewSyncUnassignedNetworkUnitCMD(
		func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler { return handler },
		func() (string, string) { return "t", "admin" },
	)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--bk-biz-id", "2,3"})
	cmd.SetContext(contextx.New(context.Background(), contextx.WithTenantID("t"), contextx.WithLoginName("admin")))

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, []int64{2, 3}, handler.syncedBKBizIDs)
	assert.Contains(t, out.String(), `"success_count": 2`)
	assert.Contains(t, out.String(), `"failed_count": 0`)
}

func TestSyncUnassignedNetworkUnitCommandAllPassesNilBizIDs(t *testing.T) {
	handler := &fakeBackendAdminHandler{syncResult: &types.NodeAgentAssignUnitResult{FailedReasons: []string{}}}

	cmd := NewSyncUnassignedNetworkUnitCMD(
		func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler { return handler },
		func() (string, string) { return "t", "admin" },
	)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Nil(t, handler.syncedBKBizIDs)
}

func TestSyncUnassignedNetworkUnitCommandRejectsMissingScope(t *testing.T) {
	cmd := NewSyncUnassignedNetworkUnitCMD(
		func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler { return &fakeBackendAdminHandler{} },
		func() (string, string) { return "t", "admin" },
	)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of the flags")
}

func TestBackendCMDRejectsMissingSyncScopeBeforeHandlerCreation(t *testing.T) {
	cmd := NewBackendCMD(func(string) (backendadmin.IHandler, error) {
		return nil, assert.AnError
	})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"node", "agent", "sync-unassigned-network-unit"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of --bk-biz-id or --all is required")
}

func TestBackendCMDRejectsSyncPositionalArgsBeforeHandlerCall(t *testing.T) {
	handler := &fakeBackendAdminHandler{syncResult: &types.NodeAgentAssignUnitResult{}}
	factoryCalls := 0
	cmd := NewBackendCMD(func(string) (backendadmin.IHandler, error) {
		factoryCalls++
		return handler, nil
	})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"node", "agent", "sync-unassigned-network-unit", "unexpected", "--all"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
	assert.Zero(t, factoryCalls)
	assert.Zero(t, handler.syncCalls)
}

func TestSyncUnassignedNetworkUnitCommandReturnsErrorOnPartialFailure(t *testing.T) {
	handler := &fakeBackendAdminHandler{syncResult: &types.NodeAgentAssignUnitResult{
		FailedCount:   1,
		FailedReasons: []string{"host-id(1) has no inner ip"},
	}}

	cmd := NewSyncUnassignedNetworkUnitCMD(
		func() backendadmin.ISyncUnassignedAgentNetworkUnitHandler { return handler },
		func() (string, string) { return "t", "admin" },
	)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"--bk-biz-id", "2"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, out.String(), `"failed_count": 1`)
	assert.Contains(t, err.Error(), "completed with 1 failed hosts")
}
