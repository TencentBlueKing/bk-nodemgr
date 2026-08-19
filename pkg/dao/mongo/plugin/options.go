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

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithFuzzyName filters by fuzzy name.
func WithFuzzyName(names ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyName, names...)
}

// WithGroup filters by group.
func WithGroup(groups ...string) OptFn {
	return base.WithValues(FieldKeyGroup, groups...)
}

// WithFuzzyPkgName filters by fuzzy plugin package name.
func WithFuzzyPkgName(pkgNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyPkgName, pkgNames...)
}

// WithName filters by name.
func WithName(pluginNames ...string) OptFn {
	return base.WithValues(FieldKeyName, pluginNames...)
}

// WithPkgName filters by plugin-pkg-name.
func WithPkgName(pluginPkgNames ...string) OptFn {
	return base.WithValues(FieldKeyPkgName, pluginPkgNames...)
}

// WithVisibleBizIDs filters by visible business IDs.
func WithVisibleBizIDs(bizIDs ...int64) OptFn {
	values := bson.A{
		bson.D{
			{Key: FieldKeyVisibleBizIDs, Value: []int64{}},
		},
		bson.D{
			{Key: FieldKeyVisibleBizIDs, Value: bson.D{
				{Key: "$exists", Value: false},
			}},
		},
	}

	if len(bizIDs) > 0 {
		values = append(values, bson.D{
			{Key: FieldKeyVisibleBizIDs, Value: bizIDs},
		})
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{
			Key:   "$or",
			Value: values,
		})
	}
}
