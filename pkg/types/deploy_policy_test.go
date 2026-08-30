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

import "testing"

func TestDeploySpecTypeIsConflict(t *testing.T) {
	tests := []struct {
		name     string
		specType DeploySpecType
		other    DeploySpecType
		want     bool
	}{
		{
			name:     "same type",
			specType: DeploySpecTypeSpecifyPluginSubConfigTemplate,
			other:    DeploySpecTypeSpecifyPluginSubConfigTemplate,
			want:     true,
		},
		{
			name:     "plugin conflicts with plugin package",
			specType: DeploySpecTypeSpecifyPlugin,
			other:    DeploySpecTypeSpecifyPluginPkg,
			want:     true,
		},
		{
			name:     "sub config conflicts with sub config template",
			specType: DeploySpecTypeSpecifyPluginSubConfig,
			other:    DeploySpecTypeSpecifyPluginSubConfigTemplate,
			want:     true,
		},
		{
			name:     "sub config template conflicts with sub config",
			specType: DeploySpecTypeSpecifyPluginSubConfigTemplate,
			other:    DeploySpecTypeSpecifyPluginSubConfig,
			want:     true,
		},
		{
			name:     "sub config template does not conflict with plugin package",
			specType: DeploySpecTypeSpecifyPluginSubConfigTemplate,
			other:    DeploySpecTypeSpecifyPluginPkg,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.specType.IsConflict(tt.other); got != tt.want {
				t.Fatalf("IsConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}
