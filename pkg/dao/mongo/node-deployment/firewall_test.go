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

package nodedeployment

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestInstallerRuntime_FirewallBSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		firewall *types.DeploymentFirewall
	}{
		{name: "missing snapshot"},
		{name: "explicit empty disables", firewall: &types.DeploymentFirewall{}},
		{
			name: "configured remote ips and features",
			firewall: &types.DeploymentFirewall{
				RemoteIPs: []string{"192.0.2.0/24"},
				Features:  []types.FirewallFeature{types.FirewallFeatureGSE, types.FirewallFeatureNodeMgr},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := bson.Marshal(InstallerRuntime{Firewall: convertFirewallFromTypes(tt.firewall)})
			require.NoError(t, err)

			var decoded InstallerRuntime
			require.NoError(t, bson.Unmarshal(encoded, &decoded))

			got := decoded.Firewall.convertToTypes()
			if tt.firewall == nil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			if len(tt.firewall.Features) == 0 {
				assert.Empty(t, got.RemoteIPs)
				assert.Empty(t, got.Features)
				return
			}

			assert.Equal(t, tt.firewall.RemoteIPs, got.RemoteIPs)
			assert.Equal(t, tt.firewall.Features, got.Features)
		})
	}
}
