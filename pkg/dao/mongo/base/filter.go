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
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// AliveFilter return a filter that only alive records.
func AliveFilter() bson.D {
	filter := bson.D{{Key: "basic.is_deleted", Value: false}}

	return filter
}

// OptFn option of find.
type OptFn func(f bson.D) bson.D

// WithInt64Values filters by int64 value.
// Deprecated: use WithValues instead.
func WithInt64Values(key string, values ...int64) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(values) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: values[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$in": values}})
	}
}

// WithoutInt64Values filters by not contains int64 value.
// Deprecated: use WithoutInt64Values instead.
func WithoutInt64Values(key string, values ...int64) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(values) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: bson.M{"$ne": values[0]}})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$nin": values}})
	}
}

// WithStringValues filters by string value.
// Deprecated: use WithValues instead.
func WithStringValues(key string, values ...string) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(values) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: values[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$in": values}})
	}
}

// WithValues filters by bool value.
func WithValues[T ~bool | ~string | ~int64](key string, values ...T) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(values) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: values[0]})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$in": values}})
	}
}

// WithAnyFieldValues filters by values matched in any field.
func WithAnyFieldValues[T ~bool | ~string | ~int64](fields []string, values ...T) OptFn {
	if len(fields) == 0 || len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(fields) == 1 {
		return WithValues(fields[0], values...)
	}

	conditions := make(bson.A, 0, len(fields))
	for _, field := range fields {
		if len(values) == 1 {
			conditions = append(conditions, bson.D{{Key: field, Value: values[0]}})
			continue
		}

		conditions = append(conditions, bson.D{{Key: field, Value: bson.M{"$in": values}}})
	}

	return func(f bson.D) bson.D {
		return appendAndCondition(f, bson.D{{Key: "$or", Value: conditions}})
	}
}

// WithoutValues filters by not contains bool value.
func WithoutValues[T ~bool | ~string | ~int64](key string, values ...T) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(values) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: bson.M{"$ne": values[0]}})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$nin": values}})
	}
}

// WithTimeRange filters by time range.
func WithTimeRange(key string, startTime, endTime time.Time) OptFn {
	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$gte": startTime, "$lte": endTime}})
	}
}

// WithoutStringValues filters by not contains string value.
// Deprecated: use WithoutStringValues instead.
func WithoutStringValues(key string, values ...string) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(values) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: bson.M{"$ne": values[0]}})
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$nin": values}})
	}
}

// WithFuzzyValues filters by fuzzy value.
func WithFuzzyValues(key string, values ...string) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$regex": "(" + strings.Join(values, "|") + ")", "$options": "i"}})
	}
}

// WithoutFuzzyValues filters by not contains fuzzy value.
func WithoutFuzzyValues(key string, values ...string) OptFn {
	if len(values) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: key, Value: bson.M{"$not": bson.M{"$regex": "(" + strings.Join(values, "|") + ")", "$options": "i"}}})
	}
}

// WithRegexMatch filters by regex match.
func WithRegexMatch(key string, patterns ...string) OptFn {
	if len(patterns) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}
	if len(patterns) == 1 {
		return func(f bson.D) bson.D {
			return append(f, bson.E{Key: key, Value: bson.M{"$regex": patterns[0]}})
		}
	}

	orConditions := make([]bson.M, len(patterns))
	for i, pattern := range patterns {
		orConditions[i] = bson.M{key: bson.M{"$regex": pattern}}
	}

	return func(f bson.D) bson.D {
		return append(f, bson.E{Key: "$or", Value: orConditions})
	}
}

// WithElemMatch filters by elem match.
func WithElemMatch(key string, optFn ...OptFn) OptFn {
	if len(optFn) == 0 {
		return func(f bson.D) bson.D {
			return f
		}
	}

	return func(f bson.D) bson.D {
		elemFilter := bson.D{}
		for _, fn := range optFn {
			elemFilter = fn(elemFilter)
		}

		return append(f, bson.E{Key: key, Value: bson.M{"$elemMatch": elemFilter}})
	}
}

func appendAndCondition(f bson.D, condition bson.D) bson.D {
	for i := range f {
		if f[i].Key != "$and" {
			continue
		}

		conditions, ok := f[i].Value.(bson.A)
		if !ok {
			break
		}

		f[i].Value = append(conditions, condition)
		return f
	}

	return append(f, bson.E{Key: "$and", Value: bson.A{condition}})
}
