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

// ConvertFieldsToTypes convert fields from proto to types.
func (x *TopoConstantGetReq) ConvertFieldsToTypes() types.TopoConstantFields {
	return types.TopoConstantFields{
		CloudVendor: x.GetCloudVendor(),
		OSType:      x.GetOsType(),
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
func (x *NodeConstantDeployGetResp) ConvertConstantFromTypes(deployConfig *types.CustomDeployConfig) {
	x.Data = &NodeConstantDeployGetResp_Data{
		DefaultDeployConfig: &CustomDeployConfig{
			InstallerRuntime: &InstallerRuntime{
				BaseWorkDir: &deployConfig.InstallerRuntime.BaseWorkDir,
			},
			NodeRuntime: &NodeRuntime{
				BaseDeployDir: &deployConfig.NodeRuntime.BaseDeployDir,
				DataIpc:       &deployConfig.NodeRuntime.DataIPC,
				PluginIpc:     &deployConfig.NodeRuntime.PluginIPC,
				LogDir:        &deployConfig.NodeRuntime.LogDir,
			},
			PluginRuntime: &PluginRuntime{
				BaseDeployDir: &deployConfig.PluginRuntime.BaseDeployDir,
				LogDir:        &deployConfig.PluginRuntime.LogDir,
			},
		},
	}
}
