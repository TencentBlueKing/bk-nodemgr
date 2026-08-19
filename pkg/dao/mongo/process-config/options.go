/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package processconfig

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithName filters by name.
func WithName(names ...string) OptFn {
	return base.WithValues(FieldKeyName, names...)
}

// WithoutName filters by no contains name.
func WithoutName(names ...string) OptFn {
	return base.WithoutValues(FieldKeyName, names...)
}

// WithTemplateName filters by template name.
func WithTemplateName(templateNames ...string) OptFn {
	return base.WithValues(FieldKeyTemplateName, templateNames...)
}

// WithoutTemplateName filters by no contains template name.
func WithoutTemplateName(templateNames ...string) OptFn {
	return base.WithoutValues(FieldKeyTemplateName, templateNames...)
}

// WithProcessName filters by process name.
func WithProcessName(processNames ...string) OptFn {
	return base.WithValues(FieldKeyProcessName, processNames...)
}

// WithoutProcessName filters by no contains process name.
func WithoutProcessName(processNames ...string) OptFn {
	return base.WithoutValues(FieldKeyProcessName, processNames...)
}

// WithHostID filters by host id.
func WithHostID(hostIDs ...int64) OptFn {
	return base.WithValues(FieldKeyHostID, hostIDs...)
}

// WithoutHostID filters by no contains host id.
func WithoutHostID(hostIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyHostID, hostIDs...)
}

// WithSet filters by set.
func WithSet(sets ...string) OptFn {
	return base.WithValues(FieldKeySet, sets...)
}

// WithoutSet filters by no contains set.
func WithoutSet(sets ...string) OptFn {
	return base.WithoutValues(FieldKeySet, sets...)
}

// WithIsMainConfig filters by is main config.
func WithIsMainConfig(isMainConfigs ...bool) OptFn {
	return base.WithValues(FieldKeyIsMainConfig, isMainConfigs...)
}

// WithProcessUniqueKeys filters by process unique keys as paired host-id + process-name conditions.
func WithProcessUniqueKeys(keys ...*types.ProcessUniqueKey) OptFn {
	return func(f bson.D) bson.D {
		if len(keys) == 0 {
			return f
		}

		conditions := make(bson.A, 0, len(keys))
		for _, key := range keys {
			if key == nil {
				continue
			}

			conditions = append(conditions, bson.D{
				{Key: FieldKeyHostID, Value: key.HostID},
				{Key: FieldKeyProcessName, Value: key.Name},
			})
		}

		if len(conditions) == 0 {
			return f
		}

		return append(f, bson.E{Key: "$or", Value: conditions})
	}
}
