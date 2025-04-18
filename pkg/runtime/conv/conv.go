/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package conv ...
package conv

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
)

// ToInt64 conv interface{} to int64.
func ToInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, errors.New("value is nil")
	}

	i, done, err := convNormalTypeToInt64(value)
	if done {
		return i, err
	}

	number, err := convCustomTypeToInt64(value)
	if err != nil {
		return 0, err
	}

	return number, nil
}

// convCustomTypeToInt64 convert custom type to int64.
func convCustomTypeToInt64(value interface{}) (int64, error) {
	val := reflect.ValueOf(value)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, errors.New("value is nil pointer")
		}
		val = val.Elem()
	}

	// get the underlying value.
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number := val.Uint()
		if number > math.MaxInt64 {
			return 0, errors.New("value is overflow")
		}

		return int64(number), nil
	case reflect.Float32, reflect.Float64:
		number := val.Float()
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return 0, errors.New("value is not a number")
		}

		err := floatToInt64BoundaryCheck(number)
		if err != nil {
			return 0, err
		}

		return int64(number), nil
	default:
		return 0, fmt.Errorf("cannot convert interface to int64, kind(%v)", val.Kind())
	}
}

// convNormalTypeToInt64 convert base type to int64.
func convNormalTypeToInt64(value interface{}) (int64, bool, error) {
	// this is the most common case, but it can't handle custom types.
	switch v := value.(type) {
	case int64:
		return v, true, nil
	case int:
		return int64(v), true, nil
	case int32:
		return int64(v), true, nil
	case int16:
		return int64(v), true, nil
	case int8:
		return int64(v), true, nil
	case uint:
		if v > math.MaxInt {
			return 0, true, errors.New("value is overflow")
		}

		return int64(v), true, nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, true, errors.New("value is overflow")
		}

		return int64(v), true, nil
	case uint32:
		return int64(v), true, nil
	case uint16:
		return int64(v), true, nil
	case uint8:
		return int64(v), true, nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, true, errors.New("value is not a number")
		}

		err := floatToInt64BoundaryCheck(v)
		if err != nil {
			return 0, true, err
		}

		return int64(v), true, nil
	case float32:
		err := floatToInt64BoundaryCheck(float64(v))
		if err != nil {
			return 0, true, err
		}

		return int64(v), true, nil
	case string:
		if len(v) == 0 {
			return 0, true, errors.New("empty string")
		}

		number, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, true, err
		}

		return number, true, nil
	case json.Number:
		number, err := v.Int64()
		if err != nil {
			return 0, true, err
		}

		return number, true, nil
	}

	return 0, false, nil
}

func floatToInt64BoundaryCheck(v float64) error {
	// notice: float can't be lossless converted to int64. so we shrank accuracy.
	if v >= math.MaxInt64 {
		return errors.New("value is overflow")
	}

	if v <= math.MinInt64 {
		return errors.New("value is underflow")
	}
	return nil
}

// ToInt64Default conv interface{} to int64, return defaultVal if error.
func ToInt64Default(value interface{}, defaultVal int64) int64 {
	val, err := ToInt64(value)
	if err != nil {
		return defaultVal
	}

	return val
}

// MapToStruct map to struct.
// Note: dst must be a pointer.
// Note: this function is based on json.Marshal and json.Unmarshal, so it will allow json tags.
func MapToStruct(m map[string]any, dst any) error {
	typeof := reflect.TypeOf(dst)
	if typeof.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be a pointer, pointer-kind(%v)", typeof.Kind())
	}

	if typeof.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be a struct pointer, target-kind(%v)", typeof.Elem().Kind())
	}

	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("failed to unmarshal into dst: %w", err)
	}
	return nil
}

// ToString ...
func ToString(value interface{}) (string, error) {
	if value == nil {
		return "", errors.New("value is nil")
	}

	str, done, err := convNormalTypeToString(value)
	if done {
		return str, err
	}

	str, err = convCustomTypeToString(value)
	if err != nil {
		return "", err
	}

	return str, nil
}

// convNormalTypeToString convert normal type to string.
func convNormalTypeToString(value interface{}) (string, bool, error) {
	// this is the most common case, but it can't handle custom types.
	switch v := value.(type) {
	case string:
		return v, true, nil
	case bool:
		return strconv.FormatBool(v), true, nil
	case int:
		return strconv.FormatInt(int64(v), 10), true, nil
	case int64:
		return strconv.FormatInt(v, 10), true, nil
	case int32:
		return strconv.FormatInt(int64(v), 10), true, nil
	case int16:
		return strconv.FormatInt(int64(v), 10), true, nil
	case int8:
		return strconv.FormatInt(int64(v), 10), true, nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), true, nil
	case uint64:
		return strconv.FormatUint(v, 10), true, nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true, nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true, nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true, nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true, nil
	case []byte:
		return string(v), true, nil
	case json.Number:
		return v.String(), true, nil
	default:
		return "", false, nil
	}
}

// convNormalTypeToString convert normal type to string.
func convCustomTypeToString(value interface{}) (string, error) {
	val := reflect.ValueOf(value)

	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "", errors.New("value is nil pointer")
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return val.String(), nil

	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10), nil

	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64), nil

	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return string(val.Bytes()), nil
		}

		return "", fmt.Errorf("cannot convert slice type to string, element kind(%v)", val.Type().Elem().Kind())
	default:
		return "", fmt.Errorf("cannot convert interface to string, kind(%v)", val.Kind())
	}
}

// ToStringDefault convert value to string with default value.
func ToStringDefault(value interface{}, defaultVal string) string {
	str, err := ToString(value)
	if err != nil {
		return defaultVal
	}

	return str
}

// StructToMap convert struct to map.
// Note: this function is based on json.Marshal and json.Unmarshal, so it will allow json tags.
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("obj is not struct")
	}

	result := make(map[string]any)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// StructToMapIgnoreError convert struct to map, ignore error.
// this function should only be used in certain situations where an error is not possible
// note: please make sure the obj is a struct or a pointer to a struct
func StructToMapIgnoreError(obj interface{}) map[string]interface{} {
	result, err := StructToMap(obj)
	if err != nil {
		return make(map[string]interface{})
	}

	return result
}

// SliceUnique this is a function used to deduplicate slice.
func SliceUnique[T comparable](source []T) []T {
	if source == nil {
		return nil
	}

	target := make([]T, 0)

	uniqueMap := make(map[T]struct{})
	for _, one := range source {
		if _, exists := uniqueMap[one]; !exists {
			target = append(target, one)
			uniqueMap[one] = struct{}{}
		}
	}

	return target
}

// MapToSlice convert map to slice.
func MapToSlice[T any](m map[string]T) []T {
	values := make([]T, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}

	return values
}

// MapKeyToSlice converts a map[Key]Value to a sorted slice of keys ([]Key).
func MapKeyToSlice[K cmp.Ordered, V any](source map[K]V) []K {
	target := make([]K, 0, len(source))
	for key := range source {
		target = append(target, key)
	}

	slices.Sort(target)

	return target
}

// SliceToMap converts a slice to a map using a key extraction function.
// Returns error if duplicate keys are detected.
func SliceToMap[K comparable, V any](s []V, fn func(V) K) (m map[K]V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}

		if err != nil {
			m = nil
		}
	}()

	m = make(map[K]V, len(s))

	for i, v := range s {
		key := fn(v)

		// Use a single mapping to check for duplicates and store values
		if _, exists := m[key]; exists {
			return nil, fmt.Errorf("duplicate key %v found at index %d", key, i)
		}

		m[key] = v
	}

	return m, nil
}
