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

package types

import (
	"reflect"
	"testing"
)

func TestDeploymentFirewallNormalize(t *testing.T) {
	tests := []struct {
		name     string
		firewall DeploymentFirewall
		want     DeploymentFirewall
		wantErr  bool
	}{
		{
			name:     "empty firewall is valid",
			firewall: DeploymentFirewall{},
			want:     DeploymentFirewall{RemoteIPs: []string{}, Features: []FirewallFeature{}},
		},
		{
			name:     "remote ips without feature is valid",
			firewall: DeploymentFirewall{RemoteIPs: []string{"10.0.0.0/24"}},
			want:     DeploymentFirewall{RemoteIPs: []string{"10.0.0.0/24"}, Features: []FirewallFeature{}},
		},
		{
			name: "remote ips are masked deduplicated and sorted, features keep first-seen order",
			firewall: DeploymentFirewall{
				RemoteIPs: []string{"fd00::/8", "10.0.1.7/24", "10.0.0.0/24", "10.0.0.0/8"},
				Features:  []FirewallFeature{FirewallFeatureNodeMgr, FirewallFeatureGSE, FirewallFeatureNodeMgr},
			},
			want: DeploymentFirewall{
				RemoteIPs: []string{"10.0.0.0/8", "10.0.0.0/24", "10.0.1.0/24", "fd00::/8"},
				Features:  []FirewallFeature{FirewallFeatureNodeMgr, FirewallFeatureGSE},
			},
		},
		{
			name: "features without remote ips is invalid",
			firewall: DeploymentFirewall{
				Features: []FirewallFeature{FirewallFeatureGSE},
			},
			wantErr: true,
		},
		{
			name: "unknown feature is invalid",
			firewall: DeploymentFirewall{
				RemoteIPs: []string{"10.0.0.0/24"},
				Features:  []FirewallFeature{"ssh"},
			},
			wantErr: true,
		},
		{
			name:     "bare ipv4 is invalid",
			firewall: DeploymentFirewall{RemoteIPs: []string{"10.0.0.1"}},
			wantErr:  true,
		},
		{
			name:     "bare ipv6 is invalid",
			firewall: DeploymentFirewall{RemoteIPs: []string{"fd00::1"}},
			wantErr:  true,
		},
		{
			name:     "ip range is invalid",
			firewall: DeploymentFirewall{RemoteIPs: []string{"10.0.0.1-10.0.0.9"}},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.firewall.Normalize()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
