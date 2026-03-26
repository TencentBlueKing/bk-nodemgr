/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package auth

// Action represents an IAM action identifier.
// Action IDs are sourced from support-files/bkiamv3/templates/0003_bk_nodemgr_actions.json.tpl.
type Action string

const (
	// ActionAgentView represents the action to view Agent.
	ActionAgentView Action = "agent_view"
	// ActionAgentOperate represents the action to operate Agent.
	ActionAgentOperate Action = "agent_operate"
	// ActionAgentHistoryView represents the action to view Agent history.
	ActionAgentHistoryView Action = "agent_history_view"

	// ActionProxyView represents the action to view Proxy.
	ActionProxyView Action = "proxy_view"
	// ActionProxyOperate represents the action to operate Proxy.
	ActionProxyOperate Action = "proxy_operate"
	// ActionProxyHistoryView represents the action to view Proxy history.
	ActionProxyHistoryView Action = "proxy_history_view"

	// ActionPluginView represents the action to view Plugin.
	ActionPluginView Action = "plugin_view"
	// ActionPluginOperate represents the action to operate Plugin.
	ActionPluginOperate Action = "plugin_operate"
	// ActionPluginHistoryView represents the action to view Plugin history.
	ActionPluginHistoryView Action = "plugin_history_view"

	// ActionConfigPolicyView represents the action to view config policy.
	ActionConfigPolicyView Action = "config_policy_view"
	// ActionConfigPolicyManage represents the action to manage config policy.
	ActionConfigPolicyManage Action = "config_policy_manage"
	// ActionConfigPolicyHistoryView represents the action to view config policy history.
	ActionConfigPolicyHistoryView Action = "config_policy_history_view"

	// ActionDeployPolicyView represents the action to view deploy policy.
	ActionDeployPolicyView Action = "deploy_policy_view"
	// ActionDeployPolicyManage represents the action to manage deploy policy.
	ActionDeployPolicyManage Action = "deploy_policy_manage"
	// ActionDeployPolicyHistoryView represents the action to view deploy policy history.
	ActionDeployPolicyHistoryView Action = "deploy_policy_history_view"

	// ActionNetworkAreaView represents the action to view network area.
	ActionNetworkAreaView Action = "networkarea_view"
	// ActionNetworkAreaCreate represents the action to create network area.
	ActionNetworkAreaCreate Action = "networkarea_create"
	// ActionNetworkAreaEdit represents the action to edit network area.
	ActionNetworkAreaEdit Action = "networkarea_edit"
	// ActionNetworkAreaDelete represents the action to delete network area.
	ActionNetworkAreaDelete Action = "networkarea_delete"
	// ActionNetworkAreaHistoryView represents the action to view network area history.
	ActionNetworkAreaHistoryView Action = "networkarea_history_view"

	// ActionNetworkUnitView represents the action to view network unit.
	ActionNetworkUnitView Action = "networkunit_view"
	// ActionNetworkUnitCreate represents the action to create network unit.
	ActionNetworkUnitCreate Action = "networkunit_create"
	// ActionNetworkUnitEdit represents the action to edit network unit.
	ActionNetworkUnitEdit Action = "networkunit_edit"
	// ActionNetworkUnitDelete represents the action to delete network unit.
	ActionNetworkUnitDelete Action = "networkunit_delete"
	// ActionNetworkUnitUseForAgent represents the action to use network unit for Agent deployment.
	ActionNetworkUnitUseForAgent Action = "networkunit_use_for_agent"
	// ActionNetworkUnitUseForProxy represents the action to use network unit for Proxy deployment.
	ActionNetworkUnitUseForProxy Action = "networkunit_use_for_proxy"
	// ActionNetworkUnitHistoryView represents the action to view network unit history.
	ActionNetworkUnitHistoryView Action = "networkunit_history_view"

	// ActionPackageTypeUpload represents the action to upload package.
	ActionPackageTypeUpload Action = "package_type_upload"
	// ActionPackageView represents the action to view package.
	ActionPackageView Action = "package_view"
	// ActionPackageManage represents the action to manage package.
	ActionPackageManage Action = "package_manage"
	// ActionPackageHistoryView represents the action to view package history.
	ActionPackageHistoryView Action = "package_history_view"
)

// ActionDisplayName returns the human-readable display name for the given action.
// Returns an empty string for unknown actions.
func ActionDisplayName(a Action) string {
	switch a {
	case ActionAgentView:
		return "查看 Agent"
	case ActionAgentOperate:
		return "操作 Agent"
	case ActionAgentHistoryView:
		return "查看 Agent 历史"
	case ActionProxyView:
		return "查看 Proxy"
	case ActionProxyOperate:
		return "操作 Proxy"
	case ActionProxyHistoryView:
		return "查看 Proxy 历史"
	case ActionPluginView:
		return "查看插件"
	case ActionPluginOperate:
		return "操作插件"
	case ActionPluginHistoryView:
		return "查看插件历史"
	case ActionConfigPolicyView:
		return "查看配置策略"
	case ActionConfigPolicyManage:
		return "管理配置策略"
	case ActionConfigPolicyHistoryView:
		return "查看配置策略历史"
	case ActionDeployPolicyView:
		return "查看部署策略"
	case ActionDeployPolicyManage:
		return "管理部署策略"
	case ActionDeployPolicyHistoryView:
		return "查看部署策略历史"
	case ActionNetworkAreaView:
		return "查看管控区域"
	case ActionNetworkAreaCreate:
		return "创建管控区域"
	case ActionNetworkAreaEdit:
		return "编辑管控区域"
	case ActionNetworkAreaDelete:
		return "删除管控区域"
	case ActionNetworkAreaHistoryView:
		return "查看管控区域历史"
	case ActionNetworkUnitView:
		return "查看管控单元"
	case ActionNetworkUnitCreate:
		return "创建管控单元"
	case ActionNetworkUnitEdit:
		return "编辑管控单元"
	case ActionNetworkUnitDelete:
		return "删除管控单元"
	case ActionNetworkUnitUseForAgent:
		return "使用管控单元部署 Agent"
	case ActionNetworkUnitUseForProxy:
		return "使用管控单元部署 Proxy"
	case ActionNetworkUnitHistoryView:
		return "查看管控单元历史"
	case ActionPackageTypeUpload:
		return "上传资源包"
	case ActionPackageView:
		return "查看资源包"
	case ActionPackageManage:
		return "管理资源包"
	case ActionPackageHistoryView:
		return "查看资源包历史"
	default:
		return ""
	}
}
