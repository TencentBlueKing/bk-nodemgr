/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"go.mongodb.org/mongo-driver/bson"
)

// OptFn option of find.
type OptFn func(f bson.D) bson.D

// WithTriggerID filter by trigger id.
func WithTriggerID(triggerID ...string) OptFn {
	if len(triggerID) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: "data.trigger_id", Value: triggerID[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "data.trigger_id", Value: bson.M{"$in": triggerID}})
	}
}

// WithOperationID ...
func WithOperationID(ids ...string) OptFn {
	if len(ids) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: "data.operation_id", Value: ids[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "data.operation_id", Value: bson.M{"$in": ids}})
	}
}

// WithEmptyOperation filter by operation_instance.
func WithEmptyOperation() OptFn {
	return func(f bson.D) bson.D {
		return append(f, bson.E{
			Key:   "data.oper_inst_empty",
			Value: true,
		})
	}
}
