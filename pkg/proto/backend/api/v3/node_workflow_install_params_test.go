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

func TestNodeWorkflowInstallParamsRoundTrip(t *testing.T) {
	t.Parallel()
	result := &types.NodeWorkflowInstallParamsResult{
		Type:                    types.NodeWorkflowTypeInstallProxy,
		UnavailableOperationIDs: []string{"unavailable-operation"},
		Items: []*types.NodeWorkflowInstallParams{{
			OperationID: "operation",
			Host: types.NodeWorkflowInstallHost{
				HostID: 1, BizID: 2, NetworkAreaID: 3, NetworkUnitID: 4,
				Addressing:  types.Addressing("static"),
				InnerIPList: []string{"10.0.0.1"}, InnerIPV6List: []string{"::1"},
				OSType: "linux", CPUArch: "x86_64",
				LoginIP: "10.0.0.2", LoginPort: 2222, LoginUser: "root",
				LoginMode: types.LoginMode("password"), InstallMethod: types.NodeInstallMethodSSH,
				IsSudo: true, ReRegister: true, InstallPreOrderedPlugins: true,
				RenewGSETask: true, RenewGSEProc: true,
				ExportIP: "10.0.0.3", ExportIPV6: "::3", AdvertiseIP: "10.0.0.4", AdvertiseIPV6: "::4",
				ProxyTags: []types.ProxyTag{"test"}, ProxyInstallOriginUnitID: 5,
				RelayDownloadPort: 17980, RelayCallbackPort: 17981,
			},
			TargetVersion: []types.TargetVersion{{OsType: criteria.OSType("linux"), CPUArch: criteria.CPUArch("x86_64"), Version: "2.0.0"}},
			IsManual:      true, IsOffline: true, MissingFields: []string{"bk_host_innerip_v6"},
		}},
	}
	resp := new(NodeWorkflowOperationInstallParamsGetResp)
	resp.ConvertResultFromTypes(result)
	require.Equal(t, result, resp.ConvertResultToTypes())

	encoded, err := json.Marshal(resp.GetData())
	require.NoError(t, err)
	var data map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &data))
	var items []struct {
		Host map[string]json.RawMessage `json:"host"`
	}
	require.NoError(t, json.Unmarshal(data["items"], &items))
	require.JSONEq(t, `"password"`, string(items[0].Host["login_mode"]))
	require.JSONEq(t, `"root"`, string(items[0].Host["login_user"]))
	require.JSONEq(t, `2222`, string(items[0].Host["login_port"]))
	for _, forbidden := range []string{"login_credit_id", "LoginCreditID", "password", "private_key", "privatekey", "token", "credit_id"} {
		require.NotContains(t, items[0].Host, forbidden)
	}

	wireHost := new(NodeWorkflowInstallHost)
	require.NoError(t, json.Unmarshal([]byte(`{
		"login_mode":"password","login_user":"root","login_port":2222,
		"login_credit_id":"secret-credit","password":"secret-password",
		"private_key":"secret-private-key","token":"secret-token"
	}`), wireHost))
	projectedHost := new(NodeWorkflowInstallHost)
	projectedHost.ConvertFromTypes(wireHost.ConvertToTypes())
	encoded, err = json.Marshal(projectedHost)
	require.NoError(t, err)
	for _, secret := range []string{"secret-credit", "secret-password", "secret-private-key", "secret-token"} {
		require.NotContains(t, string(encoded), secret)
	}
	require.Equal(t, "password", projectedHost.GetLoginMode())
	require.Equal(t, "root", projectedHost.GetLoginUser())
	require.Equal(t, int64(2222), projectedHost.GetLoginPort())
}
