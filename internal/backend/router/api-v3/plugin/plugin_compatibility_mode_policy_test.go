/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/compatibility"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestApplyPluginCompatibilityModePolicy(t *testing.T) {
	tests := []struct {
		name     string
		policy   compatibility.Policy
		tenantID string
		param    *types.PluginDeploymentParam
		want     bool
	}{
		{
			name:     "default policy enables bkmonitorbeat in system biz 2",
			policy:   compatibility.DefaultPolicy(),
			tenantID: "system",
			param: &types.PluginDeploymentParam{
				BizID:      2,
				PluginName: "bkmonitorbeat",
			},
			want: true,
		},
		{
			name: "disabled biz overrides bkmonitorbeat allowlist",
			policy: compatibility.Policy{
				EnabledPlugins: []string{"bkmonitorbeat"},
				DisabledBiz: []compatibility.DisabledBiz{
					{TenantID: "system", BKBizID: 2},
				},
			},
			tenantID: "system",
			param: &types.PluginDeploymentParam{
				BizID:      2,
				PluginName: "bkmonitorbeat",
			},
			want: false,
		},
		{
			name: "disabled biz overrides non listed plugin",
			policy: compatibility.Policy{
				EnabledPlugins: []string{"bkmonitorbeat"},
				DisabledBiz: []compatibility.DisabledBiz{
					{TenantID: "system", BKBizID: 2},
				},
			},
			tenantID: "system",
			param: &types.PluginDeploymentParam{
				BizID:      2,
				PluginName: "gse_agent",
			},
			want: false,
		},
		{
			name:     "default policy keeps non listed plugin disabled outside disabled biz",
			policy:   compatibility.DefaultPolicy(),
			tenantID: "system",
			param: &types.PluginDeploymentParam{
				BizID:      3,
				PluginName: "gse_agent",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyPluginCompatibilityModePolicy(tt.policy, tt.tenantID, tt.param)

			if tt.param.EnableCompatibilityMode != tt.want {
				t.Fatalf("EnableCompatibilityMode = %t, want %t", tt.param.EnableCompatibilityMode, tt.want)
			}
		})
	}
}
