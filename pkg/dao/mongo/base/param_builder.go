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

// Package base ...
package base

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// BuildUpsertParam build update param.
func BuildUpsertParam(data any) bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				FieldKeyIsDeleted: false,
				FieldKeyUpdatedAt: nowTime,
				"data":            data,
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.M{
				FieldKeyCreatedAt: nowTime,
			},
		},
	}

	return update
}

// BuildUpdateField build update field param.
// Deprecated: use Orm.UpdateField instead.
func BuildUpdateField(key string, value any) bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				FieldKeyIsDeleted:           false,
				FieldKeyUpdatedAt:           nowTime,
				fmt.Sprintf("data.%s", key): value,
			},
		},
	}

	return update
}

// BuildDeleteParam build delete param.
func BuildDeleteParam() bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				FieldKeyIsDeleted:  true,
				"basic.deleted_at": nowTime,
			},
		},
	}

	return update
}

// BuildPushField build push field param.
// please don't use slice as value.
func BuildPushField(key string, value any) bson.D {
	nowTime := time.Now()
	push := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				FieldKeyIsDeleted: false,
				FieldKeyUpdatedAt: nowTime,
			},
		}, {
			Key: "$push",
			Value: bson.M{
				fmt.Sprintf("data.%s", key): value,
			},
		},
	}

	return push
}

// BuildPullField build pull field param.
// please don't use slice as value.
func BuildPullField(key string, value any) bson.D {
	nowTime := time.Now()
	pull := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				FieldKeyIsDeleted: false,
				FieldKeyUpdatedAt: nowTime,
			},
		}, {
			Key: "$pull",
			Value: bson.M{
				fmt.Sprintf("data.%s", key): bson.M{
					"$in": value,
				},
			},
		},
	}

	return pull
}
