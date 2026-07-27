/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicy

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithConfigPolicyID filters by config policy id.
func WithConfigPolicyID(configPolicyIDs ...int64) OptFn {
	return base.WithValues(FieldKeyConfigPolicyID, configPolicyIDs...)
}

// WithBizID filters by biz id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizID, bizIDs...)
}

// WithConfigPolicyType filters by config policy type.
func WithConfigPolicyType(configPolicyTypes ...types.ConfigPolicyType) OptFn {
	return base.WithStringValues(FieldKeyConfigPolicyType, types.ConfigPolicyTypeListToStringList(configPolicyTypes)...)
}

// WithEnabled filters by enabled.
func WithEnabled(enabled ...bool) OptFn {
	return base.WithValues(FieldKeyEnabled, enabled...)
}

// WithFuzzyConfigPolicyName filters by config policy name.
func WithFuzzyConfigPolicyName(names ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyConfigPolicyName, names...)
}

// WithFuzzyOperator filters by config policy operator.
func WithFuzzyOperator(operators ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyOperator, operators...)
}

// WithTargetPluginName filters by target plugin name.
func WithTargetPluginName(names ...string) OptFn {
	return base.WithStringValues(FieldKeyTargetPluginName, names...)
}

// WithEnabledScope filters by enabled scope OR target host ID.
func WithEnabledScope(
	bizID, networkAreaID, networkUnitID int64,
	osType criteria.OSType, cpuArch criteria.CPUArch,
	hostID int64,
) OptFn {

	scopeMatch := bson.M{
		FieldKeyScopes: bson.M{"$elemMatch": bson.M{
			"networkarea_id": bson.M{"$in": []int64{-1, networkAreaID}},
			"networkunit_id": bson.M{"$in": []int64{-1, networkUnitID}},
			"node_os_type":   bson.M{"$in": []string{"", string(osType)}},
			"node_cpu_arch":  bson.M{"$in": []string{"", string(cpuArch)}},
		}},
	}

	targetHostMatch := bson.M{
		FieldKeyTargetHostIDs: hostID,
	}

	opts := bson.D{
		bson.E{Key: FieldKeyBizID, Value: bizID},
		bson.E{Key: "$or", Value: bson.A{scopeMatch, targetHostMatch}},
		bson.E{Key: FieldKeyEnabled, Value: true},
	}

	return func(f bson.D) bson.D {
		return append(f, opts...)
	}
}
