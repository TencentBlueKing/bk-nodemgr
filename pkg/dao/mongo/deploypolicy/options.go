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

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithDeployPolicyID filters by deploy-policy-id.
func WithDeployPolicyID(deployPolicyIDs ...int64) OptFn {
	return base.WithValues(FieldKeyDeployPolicyID, deployPolicyIDs...)
}

// WithoutDeployPolicyID filters by not contains deploy-policy-id.
func WithoutDeployPolicyID(deployPolicyIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyDeployPolicyID, deployPolicyIDs...)
}

// WithEnabled filters by enabled.
func WithEnabled(enabled ...bool) OptFn {
	return base.WithValues(FieldKeyEnabled, enabled...)
}

// WithoutEnabled filters by not contains enabled.
func WithoutEnabled(enabled ...bool) OptFn {
	return base.WithoutValues(FieldKeyEnabled, enabled...)
}

// WithMetaName filters by meta-name.
func WithMetaName(deployPolicyNames ...string) OptFn {
	return base.WithValues(FieldKeyMetaName, deployPolicyNames...)
}

// WithoutMetaName filters by not contains meta-name.
func WithoutMetaName(deployPolicyNames ...string) OptFn {
	return base.WithoutValues(FieldKeyMetaName, deployPolicyNames...)
}

// WithFuzzyMetaName filters by fuzzy meta-name.
func WithFuzzyMetaName(deployPolicyNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyMetaName, deployPolicyNames...)
}

// WithoutFuzzyMetaName filters by not contains fuzzy meta-name.
func WithoutFuzzyMetaName(deployPolicyNames ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyMetaName, deployPolicyNames...)
}

// WithOperator filters by operator.
func WithOperator(operators ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operators...)
}

// WithoutOperator filters by not contains operator.
func WithoutOperator(operators ...string) OptFn {
	return base.WithoutValues(FieldKeyOperator, operators...)
}

// WithFuzzyOperator filters by fuzzy operator.
func WithFuzzyOperator(operators ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyOperator, operators...)
}

// WithoutFuzzyOperator filters by not contains fuzzy operator.
func WithoutFuzzyOperator(operators ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyOperator, operators...)
}

// WithDeploySpecType filters by deploy spec type.
func WithDeploySpecType(specType ...types.DeploySpecType) OptFn {
	return base.WithElemMatch(FieldKeySpecs, base.WithValues(fieldSubKeySpecsType, specType...))
}

// WithDeploySpecPluginName filters by deploy spec plugin name.
func WithDeploySpecPluginName(pluginNames ...string) OptFn {
	if len(pluginNames) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	return base.WithElemMatch(FieldKeySpecs, base.WithValues(fieldSubKeySpecsParamSpecifyPluginPluginName, pluginNames...))
}

// WithExecutedAtTimeRange filters by executed at time range.
func WithExecutedAtTimeRange(timeRange types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyLifeCycleExecutedAt, timeRange.StartTime, timeRange.EndTime)
}

// WithDsuID filters by dsu-id.
func WithDsuID(dsuIDs ...int64) OptFn {
	return base.WithValues(FieldKeyDsuID, dsuIDs...)
}

// WithoutDsuID filters by not contains dsu-id.
func WithoutDsuID(dsuIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyDsuID, dsuIDs...)
}
