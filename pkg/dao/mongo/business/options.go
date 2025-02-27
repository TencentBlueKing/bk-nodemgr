/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package business

import "go.mongodb.org/mongo-driver/bson"

// OptFn option of find.
type OptFn func(f bson.D) bson.D

// WithBizID filters by biz-id.
func WithBizID(bizIDs ...int64) OptFn {
	if len(bizIDs) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	if len(bizIDs) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: "data.biz_id", Value: bizIDs[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "data.biz_id", Value: bson.M{"$in": bizIDs}})
	}
}

// WithoutBizID filters by not contains biz-id.
func WithoutBizID(bizIDs ...int64) OptFn {
	if len(bizIDs) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	if len(bizIDs) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: "data.biz_id", Value: bson.M{"$ne": bizIDs[0]}})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "data.biz_id", Value: bson.M{"$nin": bizIDs}})
	}
}

// WithBizName filters by biz-name.
func WithBizName(bizNames ...string) OptFn {
	if len(bizNames) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	if len(bizNames) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: "data.biz_name", Value: bizNames[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "data.biz_name", Value: bson.M{"$in": bizNames}})
	}
}

// WithoutBizName filters by not contains biz-name.
func WithoutBizName(bizNames ...string) OptFn {
	if len(bizNames) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	if len(bizNames) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: "data.biz_name", Value: bson.M{"$ne": bizNames[0]}})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "data.biz_name", Value: bson.M{"$nin": bizNames}})
	}
}
