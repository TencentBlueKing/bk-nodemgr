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
	// ActionAgentView ...
	ActionAgentView Action = "agent_view"
	// ActionAgentOperate ...
	ActionAgentOperate Action = "agent_operate"
	// ActionAgentHistoryView ...
	ActionAgentHistoryView Action = "agent_history_view"

	// ActionProxyView ...
	ActionProxyView Action = "proxy_view"
	// ActionProxyOperate ...
	ActionProxyOperate Action = "proxy_operate"
	// ActionProxyHistoryView ...
	ActionProxyHistoryView Action = "proxy_history_view"

	// ActionPluginView ...
	ActionPluginView Action = "plugin_view"
	// ActionPluginOperate ...
	ActionPluginOperate Action = "plugin_operate"
	// ActionPluginHistoryView ...
	ActionPluginHistoryView Action = "plugin_history_view"

	// ActionConfigPolicyView ...
	ActionConfigPolicyView Action = "config_policy_view"
	// ActionConfigPolicyManage ...
	ActionConfigPolicyManage Action = "config_policy_manage"

	// ActionNetworkAreaView ...
	ActionNetworkAreaView Action = "networkarea_view"
	// ActionNetworkAreaManage ...
	ActionNetworkAreaManage Action = "networkarea_manage"

	// ActionNetworkUnitView ...
	ActionNetworkUnitView Action = "networkunit_view"
	// ActionNetworkUnitManage ...
	ActionNetworkUnitManage Action = "networkunit_manage"

	// ActionPackageView ...
	ActionPackageView Action = "package_view"
	// ActionPackageManage ...
	ActionPackageManage Action = "package_manage"
	// ActionPackageUpload ...
	ActionPackageUpload Action = "package_upload"
)

// ActionDisplayName returns the human-readable display name for the given action.
// Returns an empty string for unknown actions.
func ActionDisplayName(a Action) string {
	switch a {
	case ActionAgentView:
		return "Agent 查看"
	case ActionAgentOperate:
		return "Agent 操作"
	case ActionAgentHistoryView:
		return "Agent 历史查看"
	case ActionProxyView:
		return "Proxy 查看"
	case ActionProxyOperate:
		return "Proxy 操作"
	case ActionProxyHistoryView:
		return "Proxy 历史查看"
	case ActionPluginView:
		return "插件 查看"
	case ActionPluginOperate:
		return "插件 操作"
	case ActionPluginHistoryView:
		return "插件 历史查看"
	case ActionConfigPolicyView:
		return "配置策略 查看"
	case ActionConfigPolicyManage:
		return "配置策略 管理"
	case ActionNetworkAreaView:
		return "云区域 查看"
	case ActionNetworkAreaManage:
		return "云区域 管理"
	case ActionNetworkUnitView:
		return "网络单元 查看"
	case ActionNetworkUnitManage:
		return "网络单元 管理"
	case ActionPackageView:
		return "安装包 查看"
	case ActionPackageManage:
		return "安装包 管理"
	case ActionPackageUpload:
		return "安装包 上传"
	default:
		return ""
	}
}
