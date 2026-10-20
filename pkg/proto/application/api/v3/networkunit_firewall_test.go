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
	"strconv"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNetworkUnitFirewallContract(t *testing.T) {
	tests := []struct {
		name    string
		runtime string
		want    *types.DeploymentFirewall
		wantErr string
	}{
		{name: "inherit", runtime: `{}`},
		{name: "explicit disable", runtime: `{"firewall":{}}`, want: &types.DeploymentFirewall{}},
		{
			name:    "normalize configured firewall",
			runtime: `{"firewall":{"remote_ips":["2001:db8::1/32","192.0.2.7/24","192.0.2.0/24"],"features":["gse","nodemgr","gse"]}}`,
			want: &types.DeploymentFirewall{
				RemoteIPs: []string{"192.0.2.0/24", "2001:db8::/32"},
				Features:  []types.FirewallFeature{types.FirewallFeatureGSE, types.FirewallFeatureNodeMgr},
			},
		},
		{
			name:    "remote ips without features",
			runtime: `{"firewall":{"remote_ips":["192.0.2.0/24"]}}`,
			want:    &types.DeploymentFirewall{RemoteIPs: []string{"192.0.2.0/24"}},
		},
		{name: "bare ip", runtime: `{"firewall":{"remote_ips":["192.0.2.1"]}}`, wantErr: "/32"},
		{name: "unknown feature", runtime: `{"firewall":{"features":["other"]}}`, wantErr: "feature"},
		{name: "feature without remote ips", runtime: `{"firewall":{"features":["gse"]}}`, wantErr: "remote ip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &TopoNetworkUnitCreateReq{}
			body := `{"bk_networkunit_name":"windows","generation":2,"custom_deploy_config":{"windows":{"installer_runtime":` + tt.runtime + `}}}`
			require.NoError(t, json.Unmarshal([]byte(body), req))
			err := req.Validate()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, "custom_deploy_config[windows].installer_runtime.firewall")
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			config := req.ConvertCustomDeployConfigToTypes()
			req.ConvertCustomDeployConfigFromTypes(config)
			encoded, err := proto.Marshal(req)
			require.NoError(t, err)
			require.NoError(t, proto.Unmarshal(encoded, req))
			encoded, err = json.Marshal(req)
			require.NoError(t, err)
			proto.Reset(req)
			require.NoError(t, json.Unmarshal(encoded, req))
			got := req.ConvertCustomDeployConfigToTypes()[criteria.OSWindows].InstallerRuntime.Firewall
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.want.RemoteIPs, got.RemoteIPs)
			assert.ElementsMatch(t, tt.want.Features, got.Features)
		})
	}
}

func TestNetworkUnitFirewallUpdateMask(t *testing.T) {
	for _, update := range []bool{false, true} {
		t.Run("update="+strconv.FormatBool(update), func(t *testing.T) {
			req := &TopoNetworkUnitUpdateReq{}
			body := `{"fields":{"custom_deploy_config":` + strconv.FormatBool(update) + `},"networkunit":{"custom_deploy_config":{"windows":{"installer_runtime":{"firewall":{"features":["gse"]}}}}}}`
			require.NoError(t, json.Unmarshal([]byte(body), req))
			err := req.Validate()
			if update {
				assert.ErrorContains(t, err, "installer_runtime.firewall")
				return
			}
			assert.NoError(t, err)
		})
	}
}
