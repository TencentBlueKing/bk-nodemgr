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

	// FieldKeyLifeCycleCreatedAt the life cycle created at field key.
	FieldKeyLifeCycleCreatedAt = "data.life_cycle.created_at"

	// FieldKeyLifeCycleUpdatedAt the life cycle updated at field key.
	FieldKeyLifeCycleUpdatedAt = "data.life_cycle.updated_at"

	// FieldKeyLifeCycleExecutedAt the life cycle executed at field key.
	FieldKeyLifeCycleExecutedAt = "data.life_cycle.executed_at"

	// FieldKeySpecs the specs field key.
	FieldKeySpecs = "data.specs"

	// fieldSubKeySpecsType the specs type sub key.
	fieldSubKeySpecsType = "type"

	// fieldSubKeySpecsParamSpecifyPluginPluginName the specs param specify plugin name sub key.
	fieldSubKeySpecsParamSpecifyPluginPluginName = "param_specify_plugin.plugin_name"

	// FieldKeyScopes the scopes field key.
	FieldKeyScopes = "data.scopes"

	// FieldKeyDsuID the dsu id field key.
	FieldKeyDsuID = "data.dsu_id"
)
