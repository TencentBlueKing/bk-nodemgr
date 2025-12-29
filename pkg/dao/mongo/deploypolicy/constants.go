/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed on the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

const (
	// FieldKeyDeployPolicyID the deploy policy id field key.
	FieldKeyDeployPolicyID = "data.deploy_policy_id"

	// FieldKeyMeta the meta field key.
	FieldKeyMeta = "data.meta"

	// FieldKeyMetaName the meta name field key.
	FieldKeyMetaName = "data.meta.name"

	// FieldKeyEnabled the enabled field key.
	FieldKeyEnabled = "data.enabled"

	// FieldKeyOperator the operator field key.
	FieldKeyOperator = "data.operator"

	// FieldKeyLifeCycleUpdateAt the life cycle update at field key.
	FieldKeyLifeCycleUpdateAt = "data.life_cycle.update_at"

	// FieldKeySpecs the specs field key.
	FieldKeySpecs = "data.specs"

	// fieldSubKeySpecsType the specs type sub key.
	fieldSubKeySpecsType = "type"

	// fieldSubKeySpecsParamSpecifyPlugin the specs param specify plugin sub key.
	fieldSubKeySpecsParamSpecifyPlugin = "param_specify_plugin"

	// fieldSubKeySpecsParamSpecifyPluginPluginName the specs param specify plugin plugin name sub key.
	fieldSubKeySpecsParamSpecifyPluginPluginName = "param_specify_plugin.plugin_name"

	// FieldSubKeySpecsParamSpecifyPluginPkg the specs param specify plugin pkg sub key.
	fieldSubKeySpecsParamSpecifyPluginPkg = "param_specify_plugin_pkg"

	// FieldSubKeySpecsParamSpecifyPluginSubConfig the specs param specify plugin sub config sub key.
	fieldSubKeySpecsParamSpecifyPluginSubConfig = "param_specify_plugin_sub_config"

	// fieldSubKeySpecsParamSpecifyPluginSubConfigPluginName the specs param specify plugin sub config plugin name sub key.
	fieldSubKeySpecsParamSpecifyPluginSubConfigPluginName = "param_specify_plugin_sub_config.plugin_name"

	// fieldSubKeySpecsParamSpecifyAgent the specs param specify agent sub key.
	fieldSubKeySpecsParamSpecifyAgent = "param_specify_agent"

	// fieldSubKeySpecsParamSpecifyProxy the specs param specify proxy sub key.
	fieldSubKeySpecsParamSpecifyProxy = "param_specify_proxy"

	// FieldKeyScopes the scopes field key.
	FieldKeyScopes = "data.scopes"
)
