/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dpmgr

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

// DeployUnit define the deploy unit.
type DeployUnit struct {
	DeployPolicyID int64
	LifeCycle      types.DeployPolicyLifeCycle
	Targets        []*types.Target
	Specs          []*types.DeploySpec
}

// AnalyzeParams defines the parameters for analyze method.
type AnalyzeParams struct {
	DeployPolicyID int64
	Spec           *types.DeploySpec
	Targets        []*types.Target
}

// ChangeTask define the final change task.
type ChangeTask struct {
	DeployPolicyID int64
	Action         ChangeAction
	Spec           *types.DeploySpec
	Target         *types.Target
}

// ChangeAction defines the change action.
type ChangeAction string

const (
	// ChangeActionPluginInstall plugin install.
	ChangeActionPluginInstall ChangeAction = "plugin_install"
	// ChangeActionPluginUninstall plugin uninstall.
	ChangeActionPluginUninstall ChangeAction = "plugin_uninstall"
	// ChangeActionPluginUpgrade plugin upgrade.
	ChangeActionPluginUpgrade ChangeAction = "plugin_upgrade"

	// ChangeActionPluginPkgInstall plugin pkg install.
	ChangeActionPluginPkgInstall ChangeAction = "plugin_pkg_install"
	// ChangeActionPluginPkgUpgrade plugin pkg upgrade.
	ChangeActionPluginPkgUpgrade ChangeAction = "plugin_pkg_upgrade"
	// ChangeActionPluginPkgUninstall plugin pkg uninstall.
	ChangeActionPluginPkgUninstall ChangeAction = "plugin_pkg_uninstall"

	// ChangeActionPluginApplySubConfig plugin apply sub config.
	ChangeActionPluginApplySubConfig ChangeAction = "apply_sub_config"
	// ChangeActionPluginDeleteSubConfig plugin delete sub config.
	ChangeActionPluginDeleteSubConfig ChangeAction = "delete_sub_config"

	// ChangeActionAgentInstall agent install.
	ChangeActionAgentInstall ChangeAction = "agent_install"
	// ChangeActionAgentUninstall agent uninstall.
	ChangeActionAgentUninstall ChangeAction = "agent_uninstall"
	// ChangeActionAgentUpgrade agent upgrade.
	ChangeActionAgentUpgrade ChangeAction = "agent_upgrade"

	// ChangeActionProxyInstall proxy install.
	ChangeActionProxyInstall ChangeAction = "proxy_install"
	// ChangeActionProxyUninstall proxy uninstall.
	ChangeActionProxyUninstall ChangeAction = "proxy_uninstall"
	// ChangeActionProxyUpgrade proxy upgrade.
	ChangeActionProxyUpgrade ChangeAction = "proxy_upgrade"
)
