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

package networkunit

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func TestTableName_UsesTenantSuffix(t *testing.T) {
	require.Equal(t, "networkunit_tenant-a", TableName("tenant-a"))
}

func TestCustomDeployConfig_FirewallBSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		firewall *types.DeploymentFirewall
	}{
		{name: "nil inherits"},
		{name: "explicit empty disables", firewall: &types.DeploymentFirewall{}},
		{
			name: "configured remote ips and features",
			firewall: &types.DeploymentFirewall{
				RemoteIPs: []string{"192.0.2.0/24", "2001:db8::/32"},
				Features:  []types.FirewallFeature{types.FirewallFeatureGSE, types.FirewallFeatureNodeMgr},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployConfig := map[criteria.OSType]types.CustomDeployConfig{
				criteria.OSWindows: {
					InstallerRuntime: types.InstallerRuntime{Firewall: tt.firewall},
				},
			}
			stored := NetworkUnit{CustomDeployConfig: convertDeployConfigFromTypes(deployConfig)}
			encoded, err := bson.Marshal(stored)
			require.NoError(t, err)

			runtime, err := bson.Raw(encoded).LookupErr("custom_deploy_config", "windows", "installer_runtime")
			require.NoError(t, err)
			firewall, err := runtime.Document().LookupErr("firewall")
			if tt.firewall == nil {
				assert.Error(t, err, "inherited firewall must be omitted from BSON")
			} else {
				require.NoError(t, err)
				assert.Equal(t, bsontype.EmbeddedDocument, firewall.Type)
			}

			var decoded NetworkUnit
			require.NoError(t, bson.Unmarshal(encoded, &decoded))
			restored := convertDeployConfigToTypes(decoded.CustomDeployConfig)
			config, ok := restored[criteria.OSWindows]
			require.True(t, ok)
			got := config.InstallerRuntime.Firewall
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
