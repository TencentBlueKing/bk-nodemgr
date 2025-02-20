/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
				"basic.is_deleted": false,
				"basic.updated_at": nowTime,
				"data":             data,
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.M{
				"basic.created_at": nowTime,
			},
		},
	}

	return update
}

// BuildUpdateField build update field param.
func BuildUpdateField(key string, value any) bson.D {
	nowTime := time.Now()
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				"basic.is_deleted":          false,
				"basic.updated_at":          nowTime,
				fmt.Sprintf("data.%s", key): value,
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
				"basic.is_deleted": false,
				"basic.updated_at": nowTime,
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
