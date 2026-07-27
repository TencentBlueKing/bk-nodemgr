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
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestConvSpecFromTypes_ToProto_AndBack(t *testing.T) {
	tests := []struct {
		name string
		spec *types.DeploySpec
	}{
		{
			name: "SpecifyAgent",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithSpecifyAgent(&types.SpecifyAgentParam{
					NodeVersion: "1.0.0",
				})
				return spec
			}(),
		},
		{
			name: "SpecifyProxy with param",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithSpecifyProxy(&types.SpecifyProxyParam{
					NodeVersion: "2.0.0",
				})
				return spec
			}(),
		},
		{
			name: "SpecifyPlugin",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithSpecifyPlugin(&types.SpecifyPluginParam{
					PluginName:          "test-plugin",
					Version:             "1.0.0",
					CustomConfigContext: map[string]any{"key": "value"},
				})
				return spec
			}(),
		},
		{
			name: "SpecifyPluginPkg",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithSpecifyPluginPkg(&types.SpecifyPluginPkgParam{
					PluginPkgName:       "test-plugin-pkg",
					Version:             "1.0.0",
					CustomConfigContext: map[string]any{"key": "value"},
				})
				return spec
			}(),
		},
		{
			name: "ProjectPluginPkgToHosts",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithProjectPluginPkgToHosts(&types.ProjectPluginPkgToHostsParam{
					PluginPkgName:       "test-plugin-pkg",
					Version:             "1.0.0",
					CustomConfigContext: map[string]any{"key": "value"},
					PlacementHostIDs:    []int64{1001, 1002},
				})
				return spec
			}(),
		},
		{
			name: "SpecifyPluginSubConfig",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithSpecifyPluginSubConfig(&types.SpecifyPluginSubConfigParam{
					PluginName: "test-plugin",
					ConfigFilesDetail: []*types.PluginConfigDetail{
						{
							Name:         "config.conf",
							Content:      "content",
							IsMainConfig: true,
						},
					},
					CustomConfigContext: map[string]any{"key": "value"},
				})
				return spec
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert types -> proto
			protoSpec, err := convSpecFromTypes(tt.spec)
			if err != nil {
				t.Fatalf("convSpecFromTypes() error = %v", err)
			}

			// Convert proto -> types
			backSpec, err := convSpecToTypes(protoSpec)
			if err != nil {
				t.Fatalf("convSpecToTypes() error = %v", err)
			}

			// Verify type
			if backSpec.Type() != tt.spec.Type() {
				t.Errorf("Type() = %v, want %v", backSpec.Type(), tt.spec.Type())
			}

			// Verify param based on type
			switch tt.spec.Type() {
			case types.DeploySpecTypeSpecifyAgent:
				originalParam, _ := tt.spec.GetSpecifyAgentParam()
				backParam, _ := backSpec.GetSpecifyAgentParam()
				if originalParam.NodeVersion != backParam.NodeVersion {
					t.Errorf("NodeVersion = %v, want %v", backParam.NodeVersion, originalParam.NodeVersion)
				}

			case types.DeploySpecTypeSpecifyProxy:
				originalParam, err1 := tt.spec.GetSpecifyProxyParam()
				backParam, err2 := backSpec.GetSpecifyProxyParam()
				if err1 != nil {
					t.Fatalf("GetSpecifyProxyParam() error = %v", err1)
				}
				if err2 != nil {
					t.Fatalf("GetSpecifyProxyParam() error = %v", err2)
				}
				if originalParam.NodeVersion != backParam.NodeVersion {
					t.Errorf("NodeVersion = %v, want %v", backParam.NodeVersion, originalParam.NodeVersion)
				}

			case types.DeploySpecTypeSpecifyPlugin:
				originalParam, _ := tt.spec.GetSpecifyPluginParam()
				backParam, _ := backSpec.GetSpecifyPluginParam()
				if originalParam.PluginName != backParam.PluginName {
					t.Errorf("PluginName = %v, want %v", backParam.PluginName, originalParam.PluginName)
				}
				if originalParam.Version != backParam.Version {
					t.Errorf("Version = %v, want %v", backParam.Version, originalParam.Version)
				}

			case types.DeploySpecTypeSpecifyPluginPkg:
				originalParam, _ := tt.spec.GetSpecifyPluginPkgParam()
				backParam, _ := backSpec.GetSpecifyPluginPkgParam()
				if originalParam.PluginPkgName != backParam.PluginPkgName {
					t.Errorf("PluginPkgName = %v, want %v", backParam.PluginPkgName, originalParam.PluginPkgName)
				}
				if originalParam.Version != backParam.Version {
					t.Errorf("Version = %v, want %v", backParam.Version, originalParam.Version)
				}

			case types.DeploySpecTypeProjectPluginPkgToHosts:
				originalParam, _ := tt.spec.GetProjectPluginPkgToHostsParam()
				backParam, _ := backSpec.GetProjectPluginPkgToHostsParam()
				if originalParam.PluginPkgName != backParam.PluginPkgName {
					t.Errorf("PluginPkgName = %v, want %v", backParam.PluginPkgName, originalParam.PluginPkgName)
				}
				if originalParam.Version != backParam.Version {
					t.Errorf("Version = %v, want %v", backParam.Version, originalParam.Version)
				}
				if len(originalParam.PlacementHostIDs) != len(backParam.PlacementHostIDs) {
					t.Errorf("PlacementHostIDs length = %v, want %v", len(backParam.PlacementHostIDs), len(originalParam.PlacementHostIDs))
				}

			case types.DeploySpecTypeSpecifyPluginSubConfig:
				originalParam, _ := tt.spec.GetSpecifyPluginSubConfigParam()
				backParam, _ := backSpec.GetSpecifyPluginSubConfigParam()
				if originalParam.PluginName != backParam.PluginName {
					t.Errorf("PluginName = %v, want %v", backParam.PluginName, originalParam.PluginName)
				}
				if len(originalParam.ConfigFilesDetail) != len(backParam.ConfigFilesDetail) {
					t.Errorf("ConfigFilesDetail length = %v, want %v", len(backParam.ConfigFilesDetail), len(originalParam.ConfigFilesDetail))
				}
			}
		})
	}
}
