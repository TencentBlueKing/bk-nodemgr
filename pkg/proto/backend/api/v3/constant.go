/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoConstantGetReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoConstantGetReq) AutoConvert() {
}

// ConvertFieldsFromTypes convert fields from types to proto.
func (x *TopoConstantGetReq) ConvertFieldsFromTypes(fields types.TopoConstantFields) error {
	x.CloudVendor = fields.CloudVendor
	x.OsType = fields.OSType

	return nil
}

// ConvertConstantToTypes convert constant to types.
func (x *TopoConstantGetResp) ConvertConstantToTypes() *types.TopoConstant {
	if x.GetData() == nil {
		return &types.TopoConstant{}
	}

	return &types.TopoConstant{
		CloudVendor: x.GetData().GetCloudVendor(),
		OSType:      x.GetData().GetOsType(),
	}
}

// Validate check body.
func (x *NodeConstantDeployGetReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return fmt.Errorf("invalid generation, generation(%d): %w", x.GetGeneration(), err)
	}

	if err := criteria.OSType(x.GetOsType()).Validate(); err != nil {
		return fmt.Errorf("invalid os_type, os_type(%s): %w", x.GetOsType(), err)
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeConstantDeployGetReq) AutoConvert() {
}

// ConvertConstantFromTypes converts NodeDeployConf and PluginDeployConf into the response data.
// Node role defaults to agent; plugin group defaults to "default"; plugin name defaults to "bk-nodemgr-relay".
func (x *NodeConstantDeployGetResp) ConvertConstantFromTypes(
	osType criteria.OSType,
	nodeConf deployconstant.NodeDeployConf,
	pluginConf deployconstant.PluginDeployConf) {

	if x.Data == nil {
		x.Data = &NodeConstantDeployGetResp_Data{}
	}

	x.Data.DefaultDeployConfig = &CustomDeployConfig{
		InstallerRuntime: &InstallerRuntime{
			BaseWorkDir: &nodeConf.BaseWorkDir,
		},
		NodeRuntime: &NodeRuntime{
			BaseDeployDir: &nodeConf.BaseDeployDir,
			LogDir:        &nodeConf.LogDir,
		},
		PluginRuntime: &PluginRuntime{
			BaseDeployDir: &pluginConf.BaseDeployDir,
			LogDir:        &pluginConf.LogDir,
		},
	}

	if osType == criteria.OSWindows {
		dataIPC := deployconstant.GetWindowsDefaultDataIPCPort()
		pluginIPC := deployconstant.GetWindowsDefaultPluginIPCPort()
		x.Data.DefaultDeployConfig.NodeRuntime.DataIpc = &dataIPC
		x.Data.DefaultDeployConfig.NodeRuntime.PluginIpc = &pluginIPC
	}
}

// ConvertConstantToTypes converts the response data into a types.DeployConfig.
func (x *NodeConstantDeployGetResp) ConvertConstantToTypes() *types.CustomDeployConfig {
	config := x.GetData().GetDefaultDeployConfig()
	if config == nil {
		return nil
	}

	return &types.CustomDeployConfig{
		InstallerRuntime: types.InstallerRuntime{
			BaseWorkDir: config.GetInstallerRuntime().GetBaseWorkDir(),
		},
		NodeRuntime: types.NodeRuntime{
			BaseDeployDir: config.GetNodeRuntime().GetBaseDeployDir(),
			DataIPC:       config.GetNodeRuntime().GetDataIpc(),
			PluginIPC:     config.GetNodeRuntime().GetPluginIpc(),
			LogDir:        config.GetNodeRuntime().GetLogDir(),
		},
		PluginRuntime: types.PluginRuntime{
			BaseDeployDir: config.GetPluginRuntime().GetBaseDeployDir(),
			LogDir:        config.GetPluginRuntime().GetLogDir(),
		},
	}
}
