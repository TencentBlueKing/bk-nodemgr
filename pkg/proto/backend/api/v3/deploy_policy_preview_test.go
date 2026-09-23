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

package v3_test

import (
	"encoding/json"
	"testing"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestDeployPolicyPreviewRequest(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		body  string
		valid bool
	}{
		{name: "missing", body: `{}`},
		{name: "null", body: `{"deploy_policy_id":null}`},
		{name: "negative", body: `{"deploy_policy_id":-1}`},
		{name: "zero", body: `{"deploy_policy_id":0}`, valid: true},
		{name: "positive", body: `{"deploy_policy_id":42}`, valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := new(protoBackend.DeployPolicyPreviewReq)
			require.NoError(t, json.Unmarshal([]byte(tc.body), req))
			req.AutoConvert()
			if !tc.valid {
				require.Error(t, req.Validate())
				return
			}
			require.NoError(t, req.Validate())
		})
	}
}

func TestDeployPolicyPreviewResponse(t *testing.T) {
	t.Parallel()
	spec, err := types.NewDeploySpecWithSpecifyPlugin(&types.SpecifyPluginParam{
		PluginName: "example", Version: "2.0", CustomConfigContext: map[string]any{},
	})
	require.NoError(t, err)
	resp := new(protoBackend.DeployPolicyPreviewResp)
	require.NoError(t, resp.ConvertPreviewFromTypes([]*types.DeployPolicySpecPreview{{
		Spec: spec,
		Results: []*types.DeployPolicyTargetPreview{
			{Target: &types.Target{Host: types.Host{HostID: 1}}, Status: types.DeployPolicyPreviewStatusSatisfied},
			{Target: &types.Target{Host: types.Host{HostID: 2}}, Status: types.DeployPolicyPreviewStatusUnsatisfied},
			{Target: &types.Target{Host: types.Host{HostID: 3}}, Status: types.DeployPolicyPreviewStatusUnmanaged},
			{Target: &types.Target{
				Host: types.Host{HostID: 4},
				ServiceInstance: types.ServiceInstance{
					ID: 101, ModuleID: 20, Name: "not exposed", Labels: map[string]string{"role": "web"},
				},
			}, Status: types.DeployPolicyPreviewStatusSatisfied},
		},
	}}))
	body, err := json.Marshal(resp.GetData())
	require.NoError(t, err)
	require.JSONEq(t, `{"items":[{"spec":{"type":"specify_plugin","param":{
		"plugin_name":"example","version":"2.0","custom_config_context":{}}},"results":[
		{"target":{"host":{"bk_host_id":1},"service_instance":null},"status":"satisfied"},
		{"target":{"host":{"bk_host_id":2},"service_instance":null},"status":"unsatisfied"},
		{"target":{"host":{"bk_host_id":3},"service_instance":null},"status":"unmanaged"},
		{"target":{"host":{"bk_host_id":4},"service_instance":{"id":101,"bk_module_id":20}},"status":"satisfied"}
	]}]}`, string(body))

	require.NoError(t, resp.ConvertPreviewFromTypes([]*types.DeployPolicySpecPreview{{Spec: spec}}))
	body, err = json.Marshal(resp.GetData())
	require.NoError(t, err)
	require.JSONEq(t, `{"items":[{"spec":{"type":"specify_plugin","param":{
		"plugin_name":"example","version":"2.0","custom_config_context":{}}},"results":[]}]}`, string(body))

	require.NoError(t, resp.ConvertPreviewFromTypes(nil))
	body, err = json.Marshal(resp.GetData())
	require.NoError(t, err)
	require.JSONEq(t, `{"items":[]}`, string(body))
}
