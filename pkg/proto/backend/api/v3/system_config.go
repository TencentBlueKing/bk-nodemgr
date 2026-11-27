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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate validates the parameter-free request.
func (x *GetSystemConfigReq) Validate() error { return nil }

// AutoConvert has no fields to normalize.
func (x *GetSystemConfigReq) AutoConvert() {
	// Intentionally empty: the request has no fields.
}

// Validate validates the parameter-free user-state request.
func (x *OpenSystemConfigGetReq) Validate() error { return nil }

// AutoConvert has no fields to normalize.
func (x *OpenSystemConfigGetReq) AutoConvert() {
	// Intentionally empty: the request has no fields.
}

// ConvertFromTypes converts the business view without accessing startup configuration.
func (x *SystemConfig) ConvertFromTypes(value types.SystemConfig) {
	x.TenantMode = value.TenantMode
	x.System = &SystemConfig_System{Env: value.System.Env, Edition: value.System.Edition}
	x.Iam = &SystemConfig_IAM{V3Enabled: value.IAM.V3Enabled, V4Enabled: value.IAM.V4Enabled}
	x.NodeEventDataId = &SystemConfig_NodeEventDataID{
		AgentBaseAlarmEventDataId: value.NodeEventDataID.AgentBaseAlarmEventDataID,
		TaskProcEventDataId:       value.NodeEventDataID.TaskProcEventDataID,
	}
	x.DefaultDirectNetworkUnit = &SystemConfig_DefaultDirectNetworkUnit{
		Enabled: value.DefaultDirectNetworkUnit.Enabled,
		Name:    value.DefaultDirectNetworkUnit.Name,
	}
	x.GseDeployConfigs = make([]*SystemConfig_GSEDeployConfig, 0, len(value.GSEDeployConfigs))
	for _, conf := range value.GSEDeployConfigs {
		x.GseDeployConfigs = append(x.GseDeployConfigs, &SystemConfig_GSEDeployConfig{
			Generation:     conf.Generation,
			OsType:         conf.OSType,
			BaseWorkDir:    conf.BaseWorkDir,
			BaseDeployDir:  conf.BaseDeployDir,
			ExtraConfigDir: conf.ExtraConfigDir,
			HostIdPath:     conf.HostIDPath,
			V2RunDir:       conf.V2RunDir,
			V2DataDir:      conf.V2DataDir,
			ZoneId:         conf.ZoneID,
			CityId:         conf.CityID,
			HttpProxyMode:  string(conf.HTTPProxyMode),
			FirewallFeatures: conv.SliceToSlice(conf.FirewallFeatures, func(feature types.FirewallFeature) string {
				return string(feature)
			}),
		})
	}
}

// ConvertToTypes converts the Backend wire view to the business-facing configuration.
func (x *SystemConfig) ConvertToTypes() types.SystemConfig {
	return types.SystemConfig{
		TenantMode: x.GetTenantMode(),
		System: types.SystemConfigSystem{
			Env: x.GetSystem().GetEnv(), Edition: x.GetSystem().GetEdition(),
		},
		IAM: types.SystemConfigIAM{
			V3Enabled: x.GetIam().GetV3Enabled(), V4Enabled: x.GetIam().GetV4Enabled(),
		},
		NodeEventDataID: types.SystemConfigNodeEventDataID{
			AgentBaseAlarmEventDataID: x.GetNodeEventDataId().GetAgentBaseAlarmEventDataId(),
			TaskProcEventDataID:       x.GetNodeEventDataId().GetTaskProcEventDataId(),
		},
		DefaultDirectNetworkUnit: types.SystemConfigDefaultDirectNetworkUnit{
			Enabled: x.GetDefaultDirectNetworkUnit().GetEnabled(),
			Name:    x.GetDefaultDirectNetworkUnit().GetName(),
		},
		GSEDeployConfigs: conv.SliceToSlice(x.GetGseDeployConfigs(), func(conf *SystemConfig_GSEDeployConfig) types.SystemConfigGSEDeploy {
			return types.SystemConfigGSEDeploy{
				Generation: conf.GetGeneration(), OSType: conf.GetOsType(),
				BaseWorkDir: conf.GetBaseWorkDir(), BaseDeployDir: conf.GetBaseDeployDir(),
				ExtraConfigDir: conf.GetExtraConfigDir(), HostIDPath: conf.GetHostIdPath(),
				V2RunDir: conf.GetV2RunDir(), V2DataDir: conf.GetV2DataDir(),
				ZoneID: conf.GetZoneId(), CityID: conf.GetCityId(),
				HTTPProxyMode: types.SystemConfigHTTPProxyMode(conf.GetHttpProxyMode()),
				FirewallFeatures: conv.SliceToSlice(conf.GetFirewallFeatures(), func(feature string) types.FirewallFeature {
					return types.FirewallFeature(feature)
				}),
			}
		}),
	}
}

// ConvertSystemConfigFromTypes sets the application-state response view.
func (x *GetSystemConfigResp) ConvertSystemConfigFromTypes(value types.SystemConfig) {
	x.Data = new(SystemConfig)
	x.Data.ConvertFromTypes(value)
}

// ConvertSystemConfigFromTypes sets the user-state response view.
func (x *OpenSystemConfigGetResp) ConvertSystemConfigFromTypes(value types.SystemConfig) {
	x.Data = new(SystemConfig)
	x.Data.ConvertFromTypes(value)
}
