/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

const (
	// PluginGroupDefault default plugin group name.
	PluginGroupDefault = "default"
)

// Plugin define the all info of plugin.
type Plugin struct {
	TenantID string

	Name    string
	PkgName string
	Group   string
	Memo    string
}

// PermittedOperation defines the plugin permitted operation type.
type PermittedOperation string

const (
	// PermittedOperationInstall install operation.
	PermittedOperationInstall PermittedOperation = "install"

	// PermittedOperationUpgrade upgrade operation.
	PermittedOperationUpgrade PermittedOperation = "upgrade"

	// PermittedOperationReconfig reconfig operation.
	PermittedOperationReconfig PermittedOperation = "reconfig"

	// PermittedOperationRestart restart operation.
	PermittedOperationRestart PermittedOperation = "restart"

	// PermittedOperationStop stop operation.
	PermittedOperationStop PermittedOperation = "stop"

	// PermittedOperationUninstall uninstall operation.
	PermittedOperationUninstall PermittedOperation = "uninstall"
)

func (p PermittedOperation) String() string {
	return string(p)
}

// DefaultGroupPermittedOperations returns the default plugin group permitted operations.
func DefaultGroupPermittedOperations() []PermittedOperation {
	return []PermittedOperation{
		PermittedOperationInstall,
		PermittedOperationUpgrade,
		PermittedOperationReconfig,
		PermittedOperationRestart,
		PermittedOperationStop,
		PermittedOperationUninstall,
	}
}

// PolicyGroupPermittedOperations returns the policy plugin group permitted operations.
func PolicyGroupPermittedOperations() []PermittedOperation {
	return []PermittedOperation{
		PermittedOperationRestart,
	}
}
