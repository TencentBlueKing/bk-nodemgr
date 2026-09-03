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
	"strings"
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
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
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

// StringToPort converts a decimal string to a valid port number.
func StringToPort(port string) (int64, error) {
	if port == "" {
		return 0, errors.New("port should not be empty")
	}
	if strings.Trim(port, "0123456789") != "" {
		return 0, fmt.Errorf("port must be a decimal number, port(%s)", port)
	}

	value, err := ToInt64(port)
	if err != nil {
		return 0, fmt.Errorf("invalid port, port(%s): %w", port, err)
	}
	if value < 1 || value > 65535 {
		return 0, fmt.Errorf("port out of range, port(%d)", value)
	}

	return value, nil
}

// MapToStruct map to struct.
// Note: dst must be a pointer.
// Note: this function is based on json.Marshal and json.Unmarshal, so it will allow json tags.
func MapToStruct(src map[string]any, dst any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to convert map to struct: %v", r)
		}
	}()

	typeof := reflect.TypeOf(dst)
	if typeof.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be a pointer, pointer-kind(%v)", typeof.Kind())
	}

	if typeof.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be a struct pointer, target-kind(%v)", typeof.Elem().Kind())
	}

	data, err := json.Marshal(src)
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
// nolint: unparam
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
// nolint: varnamelen
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

// SliceIntersect returns the unique intersection of two slices in left-side order.
func SliceIntersect[itemID comparable](left, right []itemID) []itemID {
	if left == nil {
		return nil
	}

	rightSet := make(map[itemID]struct{}, len(right))
	for _, item := range right {
		rightSet[item] = struct{}{}
	}

	result := make([]itemID, 0)
	seen := make(map[itemID]struct{})
	for _, item := range left {
		if _, ok := rightSet[item]; !ok {
			continue
		}
		if _, duplicated := seen[item]; duplicated {
			continue
		}

		result = append(result, item)
		seen[item] = struct{}{}
	}

	return result
}

// MapUnion returns the union of two maps.
// It returns an error when the same key exists on both sides.
func MapUnion[K comparable, V any](left, right map[K]V) (map[K]V, error) {
	result := make(map[K]V, len(left)+len(right))
	for key, value := range left {
		result[key] = value
	}

	for key, value := range right {
		if _, ok := result[key]; ok {
			return nil, fmt.Errorf("map union conflict, key(%v)", key)
		}

		result[key] = value
	}

	return result, nil
}

// MapUnionIgnoreConflict returns the union of two maps.
// Right-side values overwrite left-side values when keys are duplicated.
func MapUnionIgnoreConflict[K comparable, V any](left, right map[K]V) map[K]V {
	result := make(map[K]V, len(left)+len(right))
	for key, value := range left {
		result[key] = value
	}
	for key, value := range right {
		result[key] = value
	}

	return result
}

// MapValueToSlice convert map value to slice.
func MapValueToSlice[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
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
// nolint: nonamedreturns,varnamelen
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

// SliceToMapIgnore converts a slice to a map using a key extraction function.
// Returns error if duplicate keys are detected.
// nolint: nonamedreturns,varnamelen
func SliceToMapIgnore[K comparable, V any](s []V, fn func(V) K) (m map[K]V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}

		if err != nil {
			m = nil
		}
	}()

	m = make(map[K]V, len(s))

	for _, v := range s {
		key := fn(v)

		m[key] = v
	}

	return m, nil
}

// IsEmpty checks if a given value is "empty".
func IsEmpty(given any) bool {
	if given == nil {
		return true
	}

	value := reflect.ValueOf(given)
	if !value.IsValid() {
		return true
	}

	switch value.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Complex64, reflect.Complex128:
		return value.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Chan, reflect.Func:
		return value.IsNil()
	case reflect.Ptr:
		if value.IsNil() {
			return true
		}

		return IsEmpty(value.Elem().Interface())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if !IsEmpty(value.Field(i).Interface()) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// StringToBool converts a string to a boolean value.
func StringToBool(val string) (bool, error) {
	str := strings.ToLower(strings.TrimSpace(val))
	switch str {
	case "true", "yes", "on", "1":
		return true, nil
	case "false", "no", "off", "0", "":
		return false, nil
	default:
		if num, err := strconv.ParseFloat(str, 64); err == nil {
			return num != 0, nil
		}

		return false, fmt.Errorf("cannot convert string(%s) to bool", str)
	}
}

// SliceToSlice converts a slice to a slice.
func SliceToSlice[T any, U any](originSlice []T, fn func(T) U) []U {
	resultSlice := make([]U, len(originSlice))
	for i := range originSlice {
		resultSlice[i] = fn(originSlice[i])
	}

	return resultSlice
}

// SliceToSliceWithError converts a slice to a slice with error handling.
func SliceToSliceWithError[T any, U any](originSlice []T, fn func(T) (U, error)) ([]U, error) {
	resultSlice := make([]U, len(originSlice))
	for i := range originSlice {
		var err error
		resultSlice[i], err = fn(originSlice[i])
		if err != nil {
			return nil, err
		}
	}

	return resultSlice, nil
}

// NumberConvertible defines a type constraint that matches all numeric types.
type NumberConvertible interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// NumberToBool converts a numeric value to a boolean value.
func NumberToBool[T NumberConvertible](val T) bool {
	return val != 0
}

// NonEmptyOr returns the given value if it is not empty, otherwise returns the default value.
func NonEmptyOr[T any](val T, defaultVal T) T {
	if IsEmpty(val) {
		return defaultVal
	}

	return val
}

// ToBool conv interface{} to bool.
func ToBool(value interface{}) (bool, error) {
	if value == nil {
		return false, errors.New("value is nil")
	}

	result, done, err := convNormalTypeToBool(value)
	if done {
		return result, err
	}

	result, err = convCustomTypeToBool(value)
	if err != nil {
		return false, err
	}

	return result, nil
}

// convNormalTypeToBool convert normal type to bool.
// nolint: gocyclo
func convNormalTypeToBool(value interface{}) (bool, bool, error) {
	// this is the most common case, but it can't handle custom types.
	switch v := value.(type) {
	case bool:
		return v, true, nil
	case string:
		result, err := StringToBool(v)
		return result, true, err
	case []byte:
		result, err := StringToBool(string(v))
		return result, true, err
	case int:
		return NumberToBool(v), true, nil
	case int64:
		return NumberToBool(v), true, nil
	case int32:
		return NumberToBool(v), true, nil
	case int16:
		return NumberToBool(v), true, nil
	case int8:
		return NumberToBool(v), true, nil
	case uint:
		return NumberToBool(v), true, nil
	case uint64:
		return NumberToBool(v), true, nil
	case uint32:
		return NumberToBool(v), true, nil
	case uint16:
		return NumberToBool(v), true, nil
	case uint8:
		return NumberToBool(v), true, nil
	case float64:
		return NumberToBool(v), true, nil
	case float32:
		return NumberToBool(v), true, nil
	case json.Number:
		result, err := StringToBool(v.String())
		return result, true, err
	default:
		return false, false, nil
	}
}

// convCustomTypeToBool convert custom type to bool.
func convCustomTypeToBool(value interface{}) (bool, error) {
	val := reflect.ValueOf(value)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false, errors.New("value is nil pointer")
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Bool:
		return val.Bool(), nil
	case reflect.String:
		return StringToBool(val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return val.Float() != 0, nil
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return StringToBool(string(val.Bytes()))
		}

		return false, fmt.Errorf("cannot convert slice type to bool, element kind(%v)", val.Type().Elem().Kind())
	default:
		return false, fmt.Errorf("cannot convert interface to bool, kind(%v)", val.Kind())
	}
}

// ToBoolDefault convert interface{} to bool with default value.
func ToBoolDefault(value interface{}, defaultValue bool) bool {
	result, err := ToBool(value)
	if err != nil {
		return defaultValue
	}

	return result
}
