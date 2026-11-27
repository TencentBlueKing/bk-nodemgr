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

// ConvertSystemConfigFromTypes sets the read-only configuration response.
func (x *GetSystemConfigResp) ConvertSystemConfigFromTypes(value types.SystemConfig) {
	x.Data = &SystemConfig{
		TenantMode: value.TenantMode,
		System:     &SystemConfig_System{Env: value.System.Env, Edition: value.System.Edition},
		Iam:        &SystemConfig_IAM{V3Enabled: value.IAM.V3Enabled, V4Enabled: value.IAM.V4Enabled},
		NodeEventDataId: &SystemConfig_NodeEventDataID{
			AgentBaseAlarmEventDataId: value.NodeEventDataID.AgentBaseAlarmEventDataID,
			TaskProcEventDataId:       value.NodeEventDataID.TaskProcEventDataID,
		},
		DefaultDirectNetworkUnit: &SystemConfig_DefaultDirectNetworkUnit{
			Enabled: value.DefaultDirectNetworkUnit.Enabled, Name: value.DefaultDirectNetworkUnit.Name,
		},
		GseDeployConfigs: conv.SliceToSlice(value.GSEDeployConfigs, func(conf types.SystemConfigGSEDeploy) *SystemConfig_GSEDeployConfig {
			return &SystemConfig_GSEDeployConfig{
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
			}
		}),
	}
}
