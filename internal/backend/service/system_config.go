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

package service

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/systemconfig"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (svc *Service) initialSystemConfig() {
	deployConfigs := make([]types.SystemConfigGSEDeploy, 0, len(svc.conf.GSEDeployConfs))
	for _, conf := range svc.conf.GSEDeployConfs {
		var proxyMode types.SystemConfigHTTPProxyMode
		switch {
		case !conf.HTTPProxy.Override:
			proxyMode = types.SystemConfigHTTPProxyModeInherit
		case conf.HTTPProxy.Address == "":
			proxyMode = types.SystemConfigHTTPProxyModeDirect
		default:
			proxyMode = types.SystemConfigHTTPProxyModeSpecified
		}

		deployConfigs = append(deployConfigs, types.SystemConfigGSEDeploy{
			Generation:       conf.Generation,
			OSType:           conf.OsType,
			BaseWorkDir:      conf.BaseWorkDir,
			BaseDeployDir:    conf.BaseDeployDir,
			ExtraConfigDir:   conf.Custom.ExtraConfigDir,
			HostIDPath:       conf.PluginCustom.HostIDPath,
			V2RunDir:         conf.PluginCustom.V2RunDir,
			V2DataDir:        conf.PluginCustom.V2DataDir,
			ZoneID:           conf.Custom.ZoneID,
			CityID:           conf.Custom.CityID,
			HTTPProxyMode:    proxyMode,
			FirewallFeatures: conf.Firewall.Features,
		})
	}

	svc.Cap.SystemConfig = systemconfig.New(types.SystemConfig{
		TenantMode: string(svc.conf.TenantMode),
		System: types.SystemConfigSystem{
			Env:     svc.conf.System.Env,
			Edition: svc.conf.System.Edition,
		},
		IAM: types.SystemConfigIAM{
			V3Enabled: svc.conf.IAMV3.Enable,
			V4Enabled: svc.conf.IAMV4.Enable,
		},
		NodeEventDataID: types.SystemConfigNodeEventDataID{
			AgentBaseAlarmEventDataID: svc.conf.NodeEventDataID.Default.AgentBaseAlarmEventDataID,
			TaskProcEventDataID:       svc.conf.NodeEventDataID.Default.TaskProcEventDataID,
		},
		DefaultDirectNetworkUnit: types.SystemConfigDefaultDirectNetworkUnit{
			Enabled: svc.conf.NetworkUnit.DefaultDirectUnit.Enabled,
			Name:    svc.conf.NetworkUnit.DefaultDirectUnit.Name,
		},
		GSEDeployConfigs: deployConfigs,
	})
}
