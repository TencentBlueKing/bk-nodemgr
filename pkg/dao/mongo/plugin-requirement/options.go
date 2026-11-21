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

package pluginrequirement

import (
	"slices"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides requirement metadata filtering options.
type OptFn = base.OptFn

// WithID filters by stable file record IDs.
func WithID(id ...string) OptFn {
	return base.WithValues(FieldKeyID, id...)
}

// WithName filters by exact display filenames.
func WithName(name ...string) OptFn {
	return base.WithValues(FieldKeyName, name...)
}

// WithOS filters stored OS values, including legacy unrestricted values for all.
func WithOS(os ...criteria.OSType) OptFn {
	return withScopeValues(FieldKeyOS, os...)
}

// WithArch filters stored architecture values, including legacy unrestricted values for all.
func WithArch(arch ...criteria.CPUArch) OptFn {
	return withScopeValues(FieldKeyArch, arch...)
}

// WithVersion filters applicable package versions, including legacy unrestricted values for all.
func WithVersion(version ...string) OptFn {
	return withScopeValues(FieldKeyVersion, version...)
}

func withScopeValues[T ~string](key string, values ...T) OptFn {
	if !slices.Contains(values, T(types.PluginRequirementScopeAll)) {
		return base.WithValues(key, values...)
	}

	// Empty and absent dimensions were unrestricted before scope metadata was introduced.
	storedValues := make(bson.A, 0, len(values))
	for _, value := range values {
		storedValues = append(storedValues, value)
	}
	// A null query value also matches an absent field in MongoDB.
	storedValues = append(storedValues, "", nil)

	return func(filter bson.D) bson.D {
		return append(filter, bson.E{Key: key, Value: bson.D{{Key: "$in", Value: storedValues}}})
	}
}
