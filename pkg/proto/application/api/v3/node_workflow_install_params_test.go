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

package v3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestNodeWorkflowInstallParamsValidate(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name         string
		workflowID   string
		operationIDs []string
		wantError    bool
	}{
		{name: "valid", workflowID: "workflow", operationIDs: []string{"operation-1", "operation-2"}},
		{name: "empty workflow", operationIDs: []string{"operation"}, wantError: true},
		{name: "blank workflow", workflowID: " \t", operationIDs: []string{"operation"}, wantError: true},
		{name: "no operations", workflowID: "workflow", wantError: true},
		{name: "empty operation", workflowID: "workflow", operationIDs: []string{""}, wantError: true},
		{name: "blank operation", workflowID: "workflow", operationIDs: []string{" \t"}, wantError: true},
		{name: "duplicate operation", workflowID: "workflow", operationIDs: []string{"operation", "operation"}, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := &NodeWorkflowOperationInstallParamsGetReq{WorkflowId: test.workflowID, OperationIds: test.operationIDs}
			req.AutoConvert()
			if test.wantError {
				require.Error(t, req.Validate())
				return
			}
			require.NoError(t, req.Validate())
		})
	}
}

func TestNodeWorkflowInstallParamsSafeResponse(t *testing.T) {
	t.Parallel()
	resp := new(NodeWorkflowOperationInstallParamsGetResp)
	resp.ConvertResultFromTypes(&types.NodeWorkflowInstallParamsResult{
		Type:                    types.NodeWorkflowType("agent_install"),
		UnavailableOperationIDs: []string{"unavailable-operation"},
		Items: []*types.NodeWorkflowInstallParams{{
			OperationID: "operation",
			Host: types.NodeWorkflowInstallHost{
				HostID: 1, LoginMode: types.LoginMode("password"), LoginUser: "root", LoginPort: 2222,
			},
			TargetVersion: []types.TargetVersion{{OsType: criteria.OSType("linux"), CPUArch: criteria.CPUArch("x86_64"), Version: "2.0.0"}},
			IsManual:      true, IsOffline: true, MissingFields: []string{"bk_host_innerip_v6"},
		}},
	})
	data := resp.GetData()
	require.Equal(t, "agent_install", data.GetType())
	require.Equal(t, []string{"unavailable-operation"}, data.GetUnavailableOperationIds())
	require.Len(t, data.GetItems(), 1)
	item := data.GetItems()[0]
	require.Equal(t, "operation", item.GetOperationId())
	require.Equal(t, int64(1), item.GetHost().GetBkHostId())
	require.True(t, item.GetIsManual())
	require.True(t, item.GetIsOffline())
	require.Equal(t, []string{"bk_host_innerip_v6"}, item.GetMissingFields())
	require.Len(t, item.GetTargetVersion(), 1)
	require.Equal(t, "linux", item.GetTargetVersion()[0].GetOsType())
	require.Equal(t, "x86_64", item.GetTargetVersion()[0].GetCpuArch())
	require.Equal(t, "2.0.0", item.GetTargetVersion()[0].GetVersion())
	encoded, err := json.Marshal(item.GetHost())
	require.NoError(t, err)
	var host map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &host))
	require.JSONEq(t, `"password"`, string(host["login_mode"]))
	require.JSONEq(t, `"root"`, string(host["login_user"]))
	require.JSONEq(t, `2222`, string(host["login_port"]))
	for _, forbidden := range []string{"login_credit_id", "LoginCreditID", "password", "private_key", "privatekey", "token", "credit_id"} {
		require.NotContains(t, host, forbidden)
	}
}
