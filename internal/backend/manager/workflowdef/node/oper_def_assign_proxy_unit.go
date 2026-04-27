/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameAssignProxyUnit is the name of the assign proxy unit operation definition.
	OperDefNameAssignProxyUnit = "assign_proxy_unit"
)

// NewOperAssignProxyUnit creates a new operation definition for assigning proxy network unit.
func NewOperAssignProxyUnit(param OperParamAssignProxyUnit) operation.Definition {
	return &operAssignProxyUnit{param: param}
}

type operAssignProxyUnit struct {
	param OperParamAssignProxyUnit
}

// OperParamAssignProxyUnit defines the parameters for operAssignProxyUnit.
type OperParamAssignProxyUnit struct {
	Token             string   `json:"token"`
	Operator          string   `json:"operator"`
	NetworkUnitID     int64    `json:"network_unit_id"`
	RelayCallbackPort int64    `json:"relay_callback_port"`
	RelayDownloadPort int64    `json:"relay_download_port"`
	ProxyTags         []string `json:"proxy_tags"`
}

// Name returns the name.
func (oper *operAssignProxyUnit) Name() string {
	return OperDefNameAssignProxyUnit
}

// ActionDefNames returns the action def names.
// Order matters: AssignProxyInfo → UpdateHost → InstallPreOrderedPlugins.
// UpdateHost must run before InstallPreOrderedPlugins because the plugin installation
// reads host.Dynamic.NetworkUnitID from the DB via inject_plugin_custom_deploy_config.
// This is an "eventually consistent" design: if InstallPreOrderedPlugins fails after
// UpdateHost succeeds, the host is already marked with the new unit but plugin install
// can be retried independently.
func (oper *operAssignProxyUnit) ActionDefNames() []string {
	return []string{
		ActionNameAssignProxyInfo,
		ActionNameUpdateHost,
		ActionNameInstallPreOrderedPlugins,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operAssignProxyUnit) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameAssignProxyInfo:          true,
			ActionNameInstallPreOrderedPlugins: true,
			ActionNameUpdateHost:               true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operAssignProxyUnit) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
