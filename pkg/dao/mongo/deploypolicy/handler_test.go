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

package deploypolicy

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvSpecFromTypes_ToDAO_AndBack(t *testing.T) {
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
			name: "SpecifyProxy",
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
			name: "ProjectPluginConfigTemplateToHosts",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithProjectPluginConfigTemplateToHosts(
					&types.ProjectPluginConfigTemplateToHostsParam{
						PluginName: "test-plugin",
						ConfigFilesDetail: []*types.PluginConfigDetail{
							{
								TemplateName: "template.conf",
								IsMainConfig: false,
							},
						},
						CustomConfigContext: map[string]any{"key": "value"},
					},
				)
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
							TemplateName: "template.conf",
							Content:      "content",
							IsMainConfig: true,
						},
					},
					CustomConfigContext: map[string]any{"key": "value"},
				})
				return spec
			}(),
		},
		{
			name: "SpecifyPluginSubConfigTemplate",
			spec: func() *types.DeploySpec {
				spec, _ := types.NewDeploySpecWithSpecifyPluginSubConfigTemplate(&types.SpecifyPluginSubConfigTemplateParam{
					PluginName: "test-plugin",
					ConfigFilesDetail: []*types.PluginConfigDetail{
						{
							Name:         "config_deploy_1.conf",
							TemplateName: "template.conf",
							Content:      "content",
							IsMainConfig: false,
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
			// Convert types -> dao
			daoSpec, err := convSpecFromTypes(tt.spec)
			if err != nil {
				t.Fatalf("convSpecFromTypes() error = %v", err)
			}

			// Convert dao -> types
			backSpec, err := convSpecToTypes(daoSpec)
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

			case types.DeploySpecTypeProjectPluginConfigTemplateToHosts:
				originalParam, _ := tt.spec.GetProjectPluginConfigTemplateToHostsParam()
				backParam, _ := backSpec.GetProjectPluginConfigTemplateToHostsParam()
				if originalParam.PluginName != backParam.PluginName {
					t.Errorf("PluginName = %v, want %v", backParam.PluginName, originalParam.PluginName)
				}
				if !reflect.DeepEqual(originalParam.ConfigFilesDetail, backParam.ConfigFilesDetail) {
					t.Errorf("ConfigFilesDetail = %+v, want %+v", backParam.ConfigFilesDetail, originalParam.ConfigFilesDetail)
				}

			case types.DeploySpecTypeSpecifyPluginSubConfig:
				originalParam, _ := tt.spec.GetSpecifyPluginSubConfigParam()
				backParam, _ := backSpec.GetSpecifyPluginSubConfigParam()
				if originalParam.PluginName != backParam.PluginName {
					t.Errorf("PluginName = %v, want %v", backParam.PluginName, originalParam.PluginName)
				}
				if !reflect.DeepEqual(originalParam.ConfigFilesDetail, backParam.ConfigFilesDetail) {
					t.Errorf("ConfigFilesDetail = %+v, want %+v", backParam.ConfigFilesDetail, originalParam.ConfigFilesDetail)
				}

			case types.DeploySpecTypeSpecifyPluginSubConfigTemplate:
				originalParam, _ := tt.spec.GetSpecifyPluginSubConfigTemplateParam()
				backParam, _ := backSpec.GetSpecifyPluginSubConfigTemplateParam()
				if originalParam.PluginName != backParam.PluginName {
					t.Errorf("PluginName = %v, want %v", backParam.PluginName, originalParam.PluginName)
				}
				if !reflect.DeepEqual(originalParam.ConfigFilesDetail, backParam.ConfigFilesDetail) {
					t.Errorf("ConfigFilesDetail = %+v, want %+v", backParam.ConfigFilesDetail, originalParam.ConfigFilesDetail)
				}
			}
		})
	}
}

func TestConvSpecToTypesCustomConfigContextWithBSONContainers(t *testing.T) {
	daoSpec := &Spec{
		Type: string(types.DeploySpecTypeSpecifyPluginSubConfigTemplate),
		ParamSpecifyPluginSubConfigTemplate: &SpecParamSpecifyPluginSubConfigTemplate{
			PluginName: "test-plugin",
			CustomConfigContext: map[string]any{
				"items": primitive.A{
					"item-a",
					primitive.M{
						"enabled": true,
						"ports":   primitive.A{int32(80), int64(443)},
					},
				},
			},
		},
	}

	spec, err := convSpecToTypes(daoSpec)
	if err != nil {
		t.Fatalf("convSpecToTypes() error = %v", err)
	}

	param, err := spec.GetSpecifyPluginSubConfigTemplateParam()
	if err != nil {
		t.Fatalf("GetSpecifyPluginSubConfigTemplateParam() error = %v", err)
	}

	want := map[string]any{
		"items": []any{
			"item-a",
			map[string]any{
				"enabled": true,
				"ports":   []any{int32(80), int64(443)},
			},
		},
	}
	if !reflect.DeepEqual(param.CustomConfigContext, want) {
		t.Fatalf("CustomConfigContext = %#v, want %#v", param.CustomConfigContext, want)
	}

	if _, err := structpb.NewStruct(param.CustomConfigContext); err != nil {
		t.Fatalf("structpb.NewStruct() error = %v", err)
	}
}
