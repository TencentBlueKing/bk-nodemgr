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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToInt64 ...
func TestToInt64(t *testing.T) {
	type MyInt64 int64
	type MyFloat64 float64

	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "zero",
			args: args{
				value: 0,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "conv int to int64",
			args: args{
				value: int(1),
			},
			want:    int64(1),
			wantErr: false,
		},
		{
			name: "conv string to int64",
			args: args{
				value: "1",
			},
			want:    int64(1),
			wantErr: false,
		},
		// base type test.
		{
			name: "conv int64 to int64",
			args: args{
				value: int64(9223372036854775807), // max int64
			},
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name: "conv int32 to int64",
			args: args{
				value: int32(2147483647), // max int32
			},
			want:    2147483647,
			wantErr: false,
		},
		{
			name: "conv int16 to int64",
			args: args{
				value: int16(32767), // max int16
			},
			want:    32767,
			wantErr: false,
		},
		{
			name: "conv int8 to int64",
			args: args{
				value: int8(127), // max int8
			},
			want:    127,
			wantErr: false,
		},
		// unsigned number test
		{
			name: "conv uint to int64",
			args: args{
				value: uint(1),
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "conv uint64 to int64",
			args: args{
				value: uint64(922337203685477580), // don't exceed int64 max positive value.
			},
			want:    922337203685477580,
			wantErr: false,
		},
		// float number test
		{
			name: "conv float64 to int64",
			args: args{
				value: float64(1.9),
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "conv float32 to int64",
			args: args{
				value: float32(1.9),
			},
			want:    1,
			wantErr: false,
		},
		// string test
		{
			name: "conv string max int64 to int64",
			args: args{
				value: "9223372036854775807", // max int64
			},
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name: "conv invalid string to int64",
			args: args{
				value: "invalid",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv empty string to int64",
			args: args{
				value: "",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv MyInt64 to int64",
			args: args{
				value: MyInt64(123),
			},
			want:    123,
			wantErr: false,
		},
		{
			name: "conv MyInt64 to int64",
			args: args{
				value: MyInt64(123),
			},
			want:    123,
			wantErr: false,
		},
		{
			name: "conv MyFloat64 to int64",
			args: args{
				value: MyFloat64(123.0),
			},
			want:    123,
			wantErr: false,
		},
		// pointer test
		{
			name: "conv *int64 to int64",
			args: args{
				value: func() interface{} {
					v := int64(789)
					return &v
				}(),
			},
			want:    789,
			wantErr: false,
		},
		{
			name: "conv **int64 to int64",
			args: args{
				value: func() interface{} {
					v := int64(789)
					vv := &v

					return &vv
				}(),
			},
			want:    789,
			wantErr: false,
		},
		// nil pointer test
		{
			name: "conv nil to int64",
			args: args{
				value: nil,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv nil pointer to int64",
			args: args{
				value: (*int64)(nil),
			},
			want:    0,
			wantErr: true,
		},
		// json.Number test
		{
			name: "conv json.Number to int64",
			args: args{
				value: json.Number("123"),
			},
			want:    123,
			wantErr: false,
		},
		// invalid type test
		{
			name: "conv bool to int64",
			args: args{
				value: true,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv struct to int64",
			args: args{
				value: struct{}{},
			},
			want:    0,
			wantErr: true,
		},
		// boundary value test
		{
			name: "conv overflow string to int64",
			args: args{
				value: "9223372036854775808", // max int64 + 1
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv underflow string to int64",
			args: args{
				value: "-9223372036854775809", // min int64 - 1
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv overflow uint64 to int64",
			args: args{
				value: uint64(1 << 63), // don't exceed int64 max positive value.
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv overflow float64 to int64",
			args: args{
				value: float64(math.MaxInt64 + 1), // don't exceed int64 max positive value.
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "conv underflow float64 to int64",
			args: args{
				value: float64(math.MinInt64 - 1), // don't less than int64 min positive value.
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt64(tt.args.value)
			if err != nil {
				t.Logf("ToInt64() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToInt64() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToInt64Default ...
func TestToInt64Default(t *testing.T) {
	type args struct {
		value      interface{}
		defaultVal int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "conv empty value",
			args: args{
				value:      "",
				defaultVal: 1,
			},
			want: 1,
		},
		{
			name: "conv int64",
			args: args{
				value:      int64(123),
				defaultVal: 0,
			},
			want: 123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToInt64Default(tt.args.value, tt.args.defaultVal); got != tt.want {
				t.Errorf("ToInt64Default() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMapToStruct ...
func TestMapToStruct(t *testing.T) {
	type args struct {
		m   map[string]any
		dst any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				m: map[string]any{
					"name": "test",
				},
				dst: &struct {
					Name string `json:"name"`
				}{},
			},
			wantErr: false,
		},
		{
			name: "multiple struct",
			args: args{
				m: map[string]any{
					"name": "test",
					"sub": map[string]any{
						"name": "sub",
					},
				},
				dst: &struct {
					Name string `json:"name"`
					Sub  struct {
						Name string `json:"name"`
					} `json:"sub"`
				}{},
			},
			wantErr: false,
		},
		{
			name: "not a pointer",
			args: args{
				m: map[string]any{
					"name": "test",
					"sub": map[string]any{
						"name": "sub",
					},
				},
				dst: struct {
					Name string `json:"name"`
					Sub  struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					} `json:"sub"`
				}{},
			},
			wantErr: true,
		},
		{
			name: "not a struct",
			args: args{
				m: map[string]any{
					"name": "test",
					"sub": map[string]any{
						"name": "sub",
					},
				},
				dst: &[]string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MapToStruct(tt.args.m, tt.args.dst)
			if err != nil {
				t.Logf("err: %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("MapToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("dst: %+v", tt.args.dst)
		})
	}
}

// TestToStringDefault
func TestToStringDefault(t *testing.T) {
	type args struct {
		value      interface{}
		defaultVal string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "val is nil, should return default value",
			args: args{
				value:      nil,
				defaultVal: "default",
			},
			want: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToStringDefault(tt.args.value, tt.args.defaultVal); got != tt.want {
				t.Errorf("ToStringDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToString
func TestToString(t *testing.T) {
	type args struct {
		value interface{}
	}

	var nilPtr *string
	str := "test"
	strPtr := &str

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// normal type test.
		{
			name:    "nil",
			args:    args{value: nil},
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			args:    args{value: ""},
			want:    "",
			wantErr: false,
		},
		{
			name:    "normal string",
			args:    args{value: "hello"},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "int",
			args:    args{value: 123},
			want:    "123",
			wantErr: false,
		},
		{
			name:    "int64",
			args:    args{value: int64(9223372036854775807)},
			want:    "9223372036854775807",
			wantErr: false,
		},
		{
			name:    "uint",
			args:    args{value: uint(123)},
			want:    "123",
			wantErr: false,
		},
		{
			name:    "float64",
			args:    args{value: 123.456},
			want:    "123.456",
			wantErr: false,
		},
		{
			name:    "float32",
			args:    args{value: float32(123.456)},
			want:    "123.456",
			wantErr: false,
		},
		{
			name:    "bool true",
			args:    args{value: true},
			want:    "true",
			wantErr: false,
		},
		{
			name:    "bool false",
			args:    args{value: false},
			want:    "false",
			wantErr: false,
		},
		{
			name:    "nil pointer",
			args:    args{value: nilPtr},
			want:    "",
			wantErr: true,
		},
		{
			name:    "string pointer",
			args:    args{value: strPtr},
			want:    "test",
			wantErr: false,
		},
		{
			name:    "[]byte",
			args:    args{value: []byte("hello")},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "empty []byte",
			args:    args{value: []byte{}},
			want:    "",
			wantErr: false,
		},
		{
			name:    "json.Number",
			args:    args{value: json.Number("123")},
			want:    "123",
			wantErr: false,
		},
		{
			name:    "negative number",
			args:    args{value: -123},
			want:    "-123",
			wantErr: false,
		},
		{
			name:    "zero number",
			args:    args{value: 0},
			want:    "0",
			wantErr: false,
		},
		{
			name:    "unsupported slice type tests",
			args:    args{value: []int{1, 2, 3}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "complex structural body testing",
			args:    args{value: struct{ Name string }{"test"}},
			want:    "",
			wantErr: true,
		},

		// special floating point test
		{
			name:    "precision floating point testing",
			args:    args{value: 123.000},
			want:    "123",
			wantErr: false,
		},
		{
			name:    "scientific notation floating point test",
			args:    args{value: 1.23e2},
			want:    "123",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToString(tt.args.value)
			if err != nil {
				t.Logf("ToString() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ToString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// StructToMap struct to map
func TestStructToMap(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				obj: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{"test", 123},
			},
			want: map[string]interface{}{
				"name": "test",
				"age":  123,
			},
			wantErr: false,
		},
		{
			name: "ignore some field",
			args: args{
				obj: struct {
					Name string `json:"name"`
					Age  int    `json:"-"`
				}{"test", 123},
			},
			want: map[string]interface{}{
				"name": "test",
			},
			wantErr: false,
		},
		{
			name: "struct pointer",
			args: args{
				obj: &struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{
					Name: "test",
					Age:  123,
				},
			},
			want: map[string]interface{}{
				"name": "test",
				"age":  123,
			},
			wantErr: false,
		},
		{
			name: "nil struct pointer",
			args: args{
				obj: (*struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				})(nil),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToMap(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Logf("StructToMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			keys := make([]string, 0)

			for k, _ := range got {
				keys = append(keys, k)
			}

			for k, _ := range tt.want {
				keys = append(keys, k)
			}

			SliceUnique(keys)

			for _, key := range keys {
				t.Logf("got[%v] = %v,\t want[%v] = %v\n", key, got[key], key, tt.want[key])
			}
		})
	}
}

// StructToMapIgnoreError struct to map
func TestStructToMapIgnoreError(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "normal",
			args: args{
				obj: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{"test", 123},
			},
			want: map[string]interface{}{
				"name": "test",
				"age":  123,
			},
		},
		{
			name: "occur error, but ignore",
			args: args{
				obj: []string{"a", "b", "c"},
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StructToMapIgnoreError(tt.args.obj)

			if len(got) == 0 {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("StructToMapIgnoreError() = %#v, want %#v", got, tt.want)
				}
			}

			keys := make([]string, 0)

			for k, _ := range got {
				keys = append(keys, k)
			}

			for k, _ := range tt.want {
				keys = append(keys, k)
			}

			SliceUnique(keys)

			for _, key := range keys {
				t.Logf("got[%v] = %v,\t want[%v] = %v\n", key, got[key], key, tt.want[key])
			}
		})
	}
}

// SliceUnique slice unique.
func TestSliceUnique(t *testing.T) {
	type args[T comparable] struct {
		source []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[string]{
		{
			name: "normal",
			args: args[string]{
				source: []string{"a", "b", "a", "c"},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "nil",
			args: args[string]{
				source: nil,
			},
			want: nil,
		},
		{
			name: "empty",
			args: args[string]{
				source: []string{},
			},
			want: []string{},
		},
		{
			name: "single element",
			args: args[string]{
				source: []string{"a"},
			},
			want: []string{"a"},
		},
		{
			name: "all unique",
			args: args[string]{
				source: []string{"a", "b", "c", "d"},
			},
			want: []string{"a", "b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceUnique(tt.args.source); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMapMapValueToSlice map to slice.
func TestMapMapValueToSlice(t *testing.T) {
	t.Run("string_key", func(t *testing.T) {
		tests := []struct {
			name string
			m    map[string]int
			want []int
		}{
			{
				name: "normal",
				m: map[string]int{
					"a": 1,
					"b": 2,
					"c": 3,
					"d": 4,
				},
				want: []int{1, 2, 3, 4},
			},
			{
				name: "empty_map",
				m:    map[string]int{},
				want: []int{},
			},
			{
				name: "single_element",
				m: map[string]int{
					"a": 1,
				},
				want: []int{1},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := MapValueToSlice(tt.m)

				sort.Slice(got, func(i, j int) bool {
					return got[i] < got[j]
				})
				sort.Slice(tt.want, func(i, j int) bool {
					return tt.want[i] < tt.want[j]
				})

				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("int_key", func(t *testing.T) {
		tests := []struct {
			name string
			m    map[int]string
			want []string
		}{
			{
				name: "normal",
				m: map[int]string{
					1: "a",
					2: "b",
					3: "c",
				},
				want: []string{"a", "b", "c"},
			},
			{
				name: "empty_map",
				m:    map[int]string{},
				want: []string{},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := MapValueToSlice(tt.m)

				sort.Strings(got)
				sort.Strings(tt.want)

				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("int64_key", func(t *testing.T) {
		tests := []struct {
			name string
			m    map[int64]float64
			want []float64
		}{
			{
				name: "normal",
				m: map[int64]float64{
					1: 1.1,
					2: 2.2,
					3: 3.3,
				},
				want: []float64{1.1, 2.2, 3.3},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := MapValueToSlice(tt.m)

				sort.Float64s(got)
				sort.Float64s(tt.want)

				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("custom_comparable_key", func(t *testing.T) {
		type customKey struct {
			ID   int
			Name string
		}

		type customValue struct {
			Value string
		}

		tests := []struct {
			name string
			m    map[customKey]customValue
			want []customValue
		}{
			{
				name: "normal",
				m: map[customKey]customValue{
					{ID: 1, Name: "a"}: {Value: "value1"},
					{ID: 2, Name: "b"}: {Value: "value2"},
					{ID: 3, Name: "c"}: {Value: "value3"},
				},
				want: []customValue{
					{Value: "value1"},
					{Value: "value2"},
					{Value: "value3"},
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := MapValueToSlice(tt.m)

				// Sort by Value for comparison
				sort.Slice(got, func(i, j int) bool {
					return got[i].Value < got[j].Value
				})
				sort.Slice(tt.want, func(i, j int) bool {
					return tt.want[i].Value < tt.want[j].Value
				})

				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("pointer_value", func(t *testing.T) {
		tests := []struct {
			name string
			m    map[string]*int
			want []*int
		}{
			{
				name: "normal",
				m: map[string]*int{
					"a": intPtr(1),
					"b": intPtr(2),
					"c": intPtr(3),
				},
				want: []*int{intPtr(1), intPtr(2), intPtr(3)},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := MapValueToSlice(tt.m)

				// Sort by dereferenced value for comparison
				sort.Slice(got, func(i, j int) bool {
					return *got[i] < *got[j]
				})
				sort.Slice(tt.want, func(i, j int) bool {
					return *tt.want[i] < *tt.want[j]
				})

				assert.Equal(t, len(tt.want), len(got), "slice length mismatch")
				for i := range got {
					assert.Equal(t, *tt.want[i], *got[i], "element at index %d", i)
				}
			})
		}
	})
}

// intPtr returns a pointer to the given int value.
func intPtr(v int) *int {
	return &v
}

// MapKeyToSlice map key to slice.
func TestMapKeyToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "Empty map with int keys",
			input:    map[int]string{},
			expected: []int{},
		},
		{
			name:     "Single element map with int keys",
			input:    map[int]string{1: "one"},
			expected: []int{1},
		},
		{
			name:     "Multiple elements map with int keys",
			input:    map[int]string{3: "three", 1: "one", 2: "two"},
			expected: []int{1, 2, 3},
		},
		{
			name:     "Empty map with string keys",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name:     "Single element map with string keys",
			input:    map[string]int{"one": 1},
			expected: []string{"one"},
		},
		{
			name:     "Multiple elements map with string keys",
			input:    map[string]int{"c": 3, "a": 1, "b": 2},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch input := tt.input.(type) {
			case map[int]string:
				result := MapKeyToSlice(input)
				expected := tt.expected.([]int)
				if len(result) != len(expected) {
					t.Errorf("Expected slice length %d, got %d", len(expected), len(result))
					return
				}
				for i, v := range expected {
					if result[i] != v {
						t.Errorf("Expected sorted keys %v, got %v", expected, result)
						break
					}
				}
			case map[string]int:
				result := MapKeyToSlice(input)
				expected := tt.expected.([]string)
				if len(result) != len(expected) {
					t.Errorf("Expected slice length %d, got %d", len(expected), len(result))
					return
				}
				for i, v := range expected {
					if result[i] != v {
						t.Errorf("Expected sorted keys %v, got %v", expected, result)
						break
					}
				}
			default:
				t.Errorf("Unsupported input type: %T", input)
			}
		})
	}
}

// TestSliceToMap tests SliceToMap.
func TestSliceToMap(t *testing.T) {
	type args[V any, K comparable] struct {
		s  []V
		fn func(V) K
	}
	type testCase[V any, K comparable] struct {
		name    string
		args    args[V, K]
		want    map[K]V
		wantErr bool
	}

	// 创建测试函数
	strKeyFn := func(v string) string { return v }
	uppercaseKeyFn := func(v string) string { return strings.ToUpper(v) }
	firstCharKeyFn := func(v string) string { return string(v[0]) }
	emptyKeyFn := func(v string) string { return "" }
	panicFn := func(v string) string {
		if len(v) == 0 {
			panic("empty string")
		}
		return v
	}

	tests := []testCase[string, string]{
		{
			name:    "nil slice",
			args:    args[string, string]{s: nil, fn: strKeyFn},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "empty slice",
			args:    args[string, string]{s: []string{}, fn: strKeyFn},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "nil function",
			args:    args[string, string]{s: []string{"a", "b"}, fn: nil},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "basic case - identity key function",
			args:    args[string, string]{s: []string{"a", "b", "c"}, fn: strKeyFn},
			want:    map[string]string{"a": "a", "b": "b", "c": "c"},
			wantErr: false,
		},
		{
			name:    "transformation - uppercase keys",
			args:    args[string, string]{s: []string{"a", "b", "c"}, fn: uppercaseKeyFn},
			want:    map[string]string{"A": "a", "B": "b", "C": "c"},
			wantErr: false,
		},
		{
			name:    "duplicate keys - first char as key",
			args:    args[string, string]{s: []string{"apple", "apricot", "banana"}, fn: firstCharKeyFn},
			want:    nil,
			wantErr: true, // 'apple' and 'apricot' would both produce key 'a'
		},
		{
			name:    "all empty keys",
			args:    args[string, string]{s: []string{"a", "b", "c"}, fn: emptyKeyFn},
			want:    nil,
			wantErr: true, // All keys would be empty, causing duplicates
		},
		{
			name:    "function panics with empty string",
			args:    args[string, string]{s: []string{"a", "", "c"}, fn: panicFn},
			want:    nil,
			wantErr: true, // Should detect the panic
		},
		{
			name:    "single element",
			args:    args[string, string]{s: []string{"solo"}, fn: strKeyFn},
			want:    map[string]string{"solo": "solo"},
			wantErr: false,
		},
		{
			name: "large slice (1000 elements)",
			args: args[string, string]{
				s: func() []string {
					result := make([]string, 1000)
					for i := range result {
						result[i] = fmt.Sprintf("str%d", i)
					}
					return result
				}(),
				fn: strKeyFn,
			},
			want: func() map[string]string {
				result := make(map[string]string)
				for i := 0; i < 1000; i++ {
					key := fmt.Sprintf("str%d", i)
					result[key] = key
				}
				return result
			}(),
			wantErr: false,
		},
		{
			name: "special characters in strings",
			args: args[string, string]{
				s:  []string{"hello", "世界", "!@#$%^&*()"},
				fn: strKeyFn,
			},
			want: map[string]string{
				"hello":      "hello",
				"世界":         "世界",
				"!@#$%^&*()": "!@#$%^&*()",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]string
			var err error

			got, err = SliceToMap(tt.args.s, tt.args.fn)

			if (err != nil) != tt.wantErr {
				t.Errorf("SliceToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("SliceToMap() got = %v, want %v", got, tt.want)
			}

		})
	}
}

// TestEmpty tests the Empty function.
func TestEmpty(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want bool
	}{
		{
			name: "empty string",
			s:    "",
			want: true,
		},
		{
			name: "string with spaces",
			s:    "   ",
			want: false,
		},
		{
			name: "non-empty string",
			s:    "hello",
			want: false,
		},
		{
			name: "empty slice",
			s:    []string{},
			want: true,
		},
		{
			name: "nil slice",
			s:    []string(nil),
			want: true,
		},
		{
			name: "non-empty slice",
			s:    []string{"apple", "banana"},
			want: false,
		},
		{
			name: "empty map",
			s:    map[string]int{},
			want: true,
		},
		{
			name: "nil map",
			s:    map[string]int(nil),
			want: true,
		},
		{
			name: "non-empty map",
			s:    map[string]int{"a": 1},
			want: false,
		},
		{
			name: "nil value",
			s:    nil,
			want: true,
		},
		{
			name: "integer zero",
			s:    0,
			want: true,
		},
		{
			name: "float zero",
			s:    0.0,
			want: true,
		},
		{
			name: "nil channel",
			s:    (chan int)(nil),
			want: true},
		{
			name: "non-nil channel",
			s:    make(chan int),
			want: false},
		{
			name: "empty struct",
			s:    struct{}{},
			want: true},
		{
			name: "nil function",
			s:    (func())(nil),
			want: true},
		{
			name: "pointer to zero",
			s: func() interface{} {
				x := 0
				return &x
			}(),
			want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.s); got != tt.want {
				t.Errorf("Empty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStringToBool tests the StringToBool function.
func TestStringToBool(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "string true",
			args:    args{value: "true"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "string false",
			args:    args{value: "false"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "string 1",
			args:    args{value: "1"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "string 0",
			args:    args{value: "0"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "string yes",
			args:    args{value: "yes"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "string no",
			args:    args{value: "no"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "invalid string",
			args:    args{value: "notabool"},
			want:    false,
			wantErr: true,
		},
		{
			name:    "empty string",
			args:    args{value: ""},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToBool(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToBool() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNumberToBool tests the NumberToBool function.
func TestNumberToBool(t *testing.T) {
	type intCase struct {
		val  int
		want bool
	}
	type uintCase struct {
		val  uint
		want bool
	}
	type floatCase struct {
		val  float64
		want bool
	}

	intCases := []intCase{
		{val: 0, want: false},
		{val: 1, want: true},
		{val: -1, want: true},
	}
	uintCases := []uintCase{
		{val: 0, want: false},
		{val: 1, want: true},
	}
	floatCases := []floatCase{
		{val: 0.0, want: false},
		{val: 1.5, want: true},
		{val: -2.3, want: true},
	}

	for _, tc := range intCases {
		t.Run(fmt.Sprintf("int_%v", tc.val), func(t *testing.T) {
			got := NumberToBool(tc.val)
			if got != tc.want {
				t.Errorf("NumberToBool(int: %v) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
	for _, tc := range uintCases {
		t.Run(fmt.Sprintf("uint_%v", tc.val), func(t *testing.T) {
			got := NumberToBool(tc.val)
			if got != tc.want {
				t.Errorf("NumberToBool(uint: %v) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
	for _, tc := range floatCases {
		t.Run(fmt.Sprintf("float64_%v", tc.val), func(t *testing.T) {
			got := NumberToBool(tc.val)
			if got != tc.want {
				t.Errorf("NumberToBool(float64: %v) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
}

// Test types for SliceToSlice tests
type Person struct {
	Name string
	Age  int
}
type PersonSummary struct {
	NameLen int
	IsAdult bool
}

// Types for nested struct test
type Address struct {
	Street string
	City   string
}
type PersonWithAddress struct {
	Name    string
	Age     int
	Address Address
}
type PersonWithAddressSummary struct {
	FullName  string
	IsAdult   bool
	City      string
	StreetLen int
}

// ValidationError for custom error type testing
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %s: %s", ve.Field, ve.Message)
}

// TestSliceToSlice tests the SliceToSlice function.
func TestSliceToSlice(t *testing.T) {
	type args[T any, U any] struct {
		originSlice []T
		fn          func(T) U
	}
	type testCase[T any, U any] struct {
		name string
		args args[T, U]
		want []U
	}

	// Test case 1: String to string transformation
	stringTests := []testCase[string, string]{
		{
			name: "nil slice",
			args: args[string, string]{
				originSlice: nil,
				fn:          func(s string) string { return strings.ToUpper(s) },
			},
			want: []string{},
		},
		{
			name: "empty slice",
			args: args[string, string]{
				originSlice: []string{},
				fn:          func(s string) string { return strings.ToUpper(s) },
			},
			want: []string{},
		},
		{
			name: "normal transformation",
			args: args[string, string]{
				originSlice: []string{"hello", "world", "test"},
				fn:          func(s string) string { return strings.ToUpper(s) },
			},
			want: []string{"HELLO", "WORLD", "TEST"},
		},
		{
			name: "identity transformation",
			args: args[string, string]{
				originSlice: []string{"a", "b", "c"},
				fn:          func(s string) string { return s },
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "prefix addition",
			args: args[string, string]{
				originSlice: []string{"apple", "banana"},
				fn:          func(s string) string { return "fruit_" + s },
			},
			want: []string{"fruit_apple", "fruit_banana"},
		},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceToSlice(tt.args.originSlice, tt.args.fn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 2: Integer to string transformation
	intToStringTests := []testCase[int, string]{
		{
			name: "int to string",
			args: args[int, string]{
				originSlice: []int{1, 2, 3, 4, 5},
				fn:          func(i int) string { return fmt.Sprintf("num_%d", i) },
			},
			want: []string{"num_1", "num_2", "num_3", "num_4", "num_5"},
		},
		{
			name: "negative int to string",
			args: args[int, string]{
				originSlice: []int{-1, 0, 1},
				fn:          func(i int) string { return strconv.Itoa(i) },
			},
			want: []string{"-1", "0", "1"},
		},
		{
			name: "int to binary string",
			args: args[int, string]{
				originSlice: []int{2, 4, 8},
				fn:          func(i int) string { return strconv.FormatInt(int64(i), 2) },
			},
			want: []string{"10", "100", "1000"},
		},
	}

	for _, tt := range intToStringTests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceToSlice(tt.args.originSlice, tt.args.fn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 3: String to integer transformation
	stringToIntTests := []testCase[string, int]{
		{
			name: "string length to int",
			args: args[string, int]{
				originSlice: []string{"a", "ab", "abc", "abcd"},
				fn:          func(s string) int { return len(s) },
			},
			want: []int{1, 2, 3, 4},
		},
		{
			name: "string to parsed int",
			args: args[string, int]{
				originSlice: []string{"10", "20", "30"},
				fn: func(s string) int {
					if val, err := strconv.Atoi(s); err == nil {
						return val
					}
					return 0
				},
			},
			want: []int{10, 20, 30},
		},
	}

	for _, tt := range stringToIntTests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceToSlice(tt.args.originSlice, tt.args.fn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 4: Struct transformation

	structTests := []testCase[Person, PersonSummary]{
		{
			name: "person to summary",
			args: args[Person, PersonSummary]{
				originSlice: []Person{
					{Name: "Alice", Age: 25},
					{Name: "Bob", Age: 17},
					{Name: "Charlie", Age: 30},
				},
				fn: func(p Person) PersonSummary {
					return PersonSummary{
						NameLen: len(p.Name),
						IsAdult: p.Age >= 18,
					}
				},
			},
			want: []PersonSummary{
				{NameLen: 5, IsAdult: true},
				{NameLen: 3, IsAdult: false},
				{NameLen: 7, IsAdult: true},
			},
		},
	}

	for _, tt := range structTests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceToSlice(tt.args.originSlice, tt.args.fn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 5: Pointer handling
	pointerTests := []testCase[*string, string]{
		{
			name: "pointer slice to value slice",
			args: args[*string, string]{
				originSlice: func() []*string {
					a, b, c := "a", "b", "c"
					return []*string{&a, &b, &c}
				}(),
				fn: func(p *string) string {
					if p != nil {
						return *p
					}
					return ""
				},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "pointer slice with nil values",
			args: args[*string, string]{
				originSlice: []*string{nil, nil, nil},
				fn: func(p *string) string {
					if p != nil {
						return *p
					}
					return "nil"
				},
			},
			want: []string{"nil", "nil", "nil"},
		},
	}

	for _, tt := range pointerTests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceToSlice(tt.args.originSlice, tt.args.fn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 6: Performance test with large slice
	t.Run("large slice performance", func(t *testing.T) {
		// Create a large slice (10000 elements)
		largeSlice := make([]int, 10000)
		for i := range largeSlice {
			largeSlice[i] = i
		}

		// Transform to strings
		got := SliceToSlice(largeSlice, func(i int) string {
			return strconv.Itoa(i)
		})

		// Verify length
		if len(got) != 10000 {
			t.Errorf("Expected length 10000, got %d", len(got))
		}

		// Verify first and last elements
		if got[0] != "0" {
			t.Errorf("Expected first element '0', got '%s'", got[0])
		}
		if got[9999] != "9999" {
			t.Errorf("Expected last element '9999', got '%s'", got[9999])
		}
	})

	// Test case 7: Complex transformation
	t.Run("complex transformation", func(t *testing.T) {
		type Input struct {
			ID   int
			Name string
			Tags []string
		}
		type Output struct {
			IDStr    string
			Initial  string
			TagCount int
		}

		inputs := []Input{
			{ID: 1, Name: "Alice", Tags: []string{"dev", "go"}},
			{ID: 2, Name: "Bob", Tags: []string{"ops"}},
			{ID: 3, Name: "Charlie", Tags: []string{"dev", "python", "ml"}},
		}

		want := []Output{
			{IDStr: "ID_1", Initial: "A", TagCount: 2},
			{IDStr: "ID_2", Initial: "B", TagCount: 1},
			{IDStr: "ID_3", Initial: "C", TagCount: 3},
		}

		got := SliceToSlice(inputs, func(in Input) Output {
			initial := ""
			if len(in.Name) > 0 {
				initial = string(in.Name[0])
			}
			return Output{
				IDStr:    fmt.Sprintf("ID_%d", in.ID),
				Initial:  initial,
				TagCount: len(in.Tags),
			}
		})

		if !reflect.DeepEqual(got, want) {
			t.Errorf("SliceToSlice() = %v, want %v", got, want)
		}
	})

	// Test case 8: Interface type slice transformation
	t.Run("interface type slice transformation", func(t *testing.T) {
		// Create interface slice with different types
		var interfaceSlice []interface{} = []interface{}{
			"hello",
			42,
			3.14,
			true,
		}

		want := []string{"string", "int", "float", "bool"}
		got := SliceToSlice(interfaceSlice, func(item interface{}) string {
			switch item.(type) {
			case string:
				return "string"
			case int:
				return "int"
			case float64:
				return "float"
			case bool:
				return "bool"
			default:
				return "unknown"
			}
		})

		if !reflect.DeepEqual(got, want) {
			t.Errorf("SliceToSlice() = %v, want %v", got, want)
		}
	})

	// Test case 9: Function type slice transformation
	t.Run("function type slice transformation", func(t *testing.T) {
		// Create slice of functions
		functionSlice := []func() int{
			func() int { return 1 },
			func() int { return 2 },
			func() int { return 3 },
		}

		want := []int{1, 2, 3}
		got := SliceToSlice(functionSlice, func(f func() int) int {
			return f()
		})

		if !reflect.DeepEqual(got, want) {
			t.Errorf("SliceToSlice() = %v, want %v", got, want)
		}
	})

	// Test case 10: Channel type slice transformation
	t.Run("channel type slice transformation", func(t *testing.T) {
		// Create slice of channels
		channelSlice := []chan int{
			make(chan int),
			make(chan int),
			make(chan int),
		}

		// Close channels to avoid blocking
		for _, ch := range channelSlice {
			close(ch)
		}

		want := []bool{true, true, true} // All channels are closed
		got := SliceToSlice(channelSlice, func(ch chan int) bool {
			// Check if channel is closed
			select {
			case <-ch:
				return true
			default:
				return false
			}
		})

		if !reflect.DeepEqual(got, want) {
			t.Errorf("SliceToSlice() = %v, want %v", got, want)
		}
	})

	// Test case 11: Nested struct transformation
	nestedStructTests := []testCase[PersonWithAddress, PersonWithAddressSummary]{
		{
			name: "nested struct transformation",
			args: args[PersonWithAddress, PersonWithAddressSummary]{
				originSlice: []PersonWithAddress{
					{
						Name: "Alice Johnson",
						Age:  25,
						Address: Address{
							Street: "123 Main St",
							City:   "New York",
						},
					},
					{
						Name: "Bob Smith",
						Age:  17,
						Address: Address{
							Street: "456 Oak Ave",
							City:   "Los Angeles",
						},
					},
				},
				fn: func(p PersonWithAddress) PersonWithAddressSummary {
					return PersonWithAddressSummary{
						FullName:  p.Name,
						IsAdult:   p.Age >= 18,
						City:      p.Address.City,
						StreetLen: len(p.Address.Street),
					}
				},
			},
			want: []PersonWithAddressSummary{
				{
					FullName:  "Alice Johnson",
					IsAdult:   true,
					City:      "New York",
					StreetLen: 11,
				},
				{
					FullName:  "Bob Smith",
					IsAdult:   false,
					City:      "Los Angeles",
					StreetLen: 11,
				},
			},
		},
	}

	for _, tt := range nestedStructTests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceToSlice(tt.args.originSlice, tt.args.fn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 12: Panic recovery in transformation function
	t.Run("panic recovery transformation", func(t *testing.T) {
		// Test that panics in transformation function are not caught by SliceToSlice
		// (this is expected behavior - panics should propagate)
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, but function did not panic")
			}
		}()

		input := []string{"a", "b", "c"}
		_ = SliceToSlice(input, func(s string) string {
			if s == "b" {
				panic("intentional panic for testing")
			}
			return s + "_transformed"
		})
	})

	// Test case 13: Nil transformation function
	t.Run("nil transformation function", func(t *testing.T) {
		// Note: This will cause a compile-time error, so we can't test it directly
		// The function signature requires a non-nil function
		input := []string{"a", "b", "c"}

		// Create a function that returns nil for specific input to test behavior
		got := SliceToSlice(input, func(s string) *string {
			if s == "b" {
				return nil
			}
			result := s + "_transformed"
			return &result
		})

		// Check that we get the expected results with nil pointer in middle
		if len(got) != 3 {
			t.Errorf("Expected length 3, got %d", len(got))
		}
		if got[0] == nil || *got[0] != "a_transformed" {
			t.Errorf("Expected 'a_transformed', got %v", got[0])
		}
		if got[1] != nil {
			t.Errorf("Expected nil for 'b', got %v", got[1])
		}
		if got[2] == nil || *got[2] != "c_transformed" {
			t.Errorf("Expected 'c_transformed', got %v", got[2])
		}
	})
}

// TestSliceToSliceWithError tests the SliceToSliceWithError function using table-driven testing.
func TestSliceToSliceWithError(t *testing.T) {
	// Define test case structures for different transformation types
	type stringTestCase struct {
		name        string
		originSlice []string
		fn          func(string) (string, error)
		want        []string
		wantErr     bool
		errMsg      string
	}

	type intToStringTestCase struct {
		name        string
		originSlice []int
		fn          func(int) (string, error)
		want        []string
		wantErr     bool
		errMsg      string
	}

	type personTestCase struct {
		name        string
		originSlice []Person
		fn          func(Person) (PersonSummary, error)
		want        []PersonSummary
		wantErr     bool
		errMsg      string
	}

	// Test case 1: String to string transformations with error handling
	stringTests := []stringTestCase{
		{
			name:        "nil slice",
			originSlice: nil,
			fn: func(s string) (string, error) {
				return strings.ToUpper(s), nil
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:        "empty slice",
			originSlice: []string{},
			fn: func(s string) (string, error) {
				return strings.ToUpper(s), nil
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:        "normal transformation",
			originSlice: []string{"hello", "world", "test"},
			fn: func(s string) (string, error) {
				return strings.ToUpper(s), nil
			},
			want:    []string{"HELLO", "WORLD", "TEST"},
			wantErr: false,
		},
		{
			name:        "first element error",
			originSlice: []string{"error", "hello", "world"},
			fn: func(s string) (string, error) {
				if s == "error" {
					return "", fmt.Errorf("transformation error for: %s", s)
				}
				return strings.ToUpper(s), nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "transformation error for: error",
		},
		{
			name:        "middle element error",
			originSlice: []string{"hello", "error", "world"},
			fn: func(s string) (string, error) {
				if s == "error" {
					return "", fmt.Errorf("transformation error for: %s", s)
				}
				return strings.ToUpper(s), nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "transformation error for: error",
		},
		{
			name:        "last element error",
			originSlice: []string{"hello", "world", "error"},
			fn: func(s string) (string, error) {
				if s == "error" {
					return "", fmt.Errorf("transformation error for: %s", s)
				}
				return strings.ToUpper(s), nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "transformation error for: error",
		},
		{
			name:        "all elements error",
			originSlice: []string{"error", "error", "error"},
			fn: func(s string) (string, error) {
				return "", fmt.Errorf("transformation error for: %s", s)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "transformation error for: error",
		},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SliceToSliceWithError(tt.originSlice, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceToSliceWithError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("SliceToSliceWithError() error message = %v, expected to contain %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSliceWithError() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 2: Integer to string transformations with error handling
	intToStringTests := []intToStringTestCase{
		{
			name:        "int to string with error",
			originSlice: []int{1, 2, 3, 4, 5},
			fn: func(i int) (string, error) {
				if i == 3 {
					return "", fmt.Errorf("invalid number: %d", i)
				}
				return fmt.Sprintf("num_%d", i), nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "invalid number: 3",
		},
		{
			name:        "negative int to string",
			originSlice: []int{-1, 0, 1},
			fn: func(i int) (string, error) {
				return strconv.Itoa(i), nil
			},
			want:    []string{"-1", "0", "1"},
			wantErr: false,
		},
		{
			name:        "zero division error",
			originSlice: []int{1, 2, 0, 4},
			fn: func(i int) (string, error) {
				if i == 0 {
					return "", fmt.Errorf("division by zero")
				}
				return fmt.Sprintf("result_%d", 10/i), nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "division by zero",
		},
	}

	for _, tt := range intToStringTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SliceToSliceWithError(tt.originSlice, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceToSliceWithError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("SliceToSliceWithError() error message = %v, expected to contain %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSliceWithError() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 3: Struct transformations with error handling
	structTests := []personTestCase{
		{
			name: "person to summary with age validation error",
			originSlice: []Person{
				{Name: "Alice", Age: 25},
				{Name: "Bob", Age: -5}, // Invalid age
				{Name: "Charlie", Age: 30},
			},
			fn: func(p Person) (PersonSummary, error) {
				if p.Age < 0 {
					return PersonSummary{}, fmt.Errorf("invalid age: %d for person: %s", p.Age, p.Name)
				}
				return PersonSummary{
					NameLen: len(p.Name),
					IsAdult: p.Age >= 18,
				}, nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "invalid age: -5 for person: Bob",
		},
		{
			name: "person to summary with name validation error",
			originSlice: []Person{
				{Name: "", Age: 25}, // Empty name
				{Name: "Bob", Age: 17},
			},
			fn: func(p Person) (PersonSummary, error) {
				if p.Name == "" {
					return PersonSummary{}, fmt.Errorf("empty name not allowed")
				}
				return PersonSummary{
					NameLen: len(p.Name),
					IsAdult: p.Age >= 18,
				}, nil
			},
			want:    nil,
			wantErr: true,
			errMsg:  "empty name not allowed",
		},
		{
			name: "successful person transformation",
			originSlice: []Person{
				{Name: "Alice", Age: 25},
				{Name: "Bob", Age: 17},
				{Name: "Charlie", Age: 30},
			},
			fn: func(p Person) (PersonSummary, error) {
				return PersonSummary{
					NameLen: len(p.Name),
					IsAdult: p.Age >= 18,
				}, nil
			},
			want: []PersonSummary{
				{NameLen: 5, IsAdult: true},
				{NameLen: 3, IsAdult: false},
				{NameLen: 7, IsAdult: true},
			},
			wantErr: false,
		},
	}

	for _, tt := range structTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SliceToSliceWithError(tt.originSlice, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceToSliceWithError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("SliceToSliceWithError() error message = %v, expected to contain %v", err.Error(), tt.errMsg)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSliceWithError() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case 4: Large slice with error handling
	t.Run("large slice with error handling", func(t *testing.T) {
		// Create a large slice (1000 elements)
		largeSlice := make([]int, 1000)
		for i := range largeSlice {
			largeSlice[i] = i
		}

		// Transform with error at position 500
		got, err := SliceToSliceWithError(largeSlice, func(i int) (string, error) {
			if i == 500 {
				return "", fmt.Errorf("error at position %d", i)
			}
			return strconv.Itoa(i), nil
		})

		// Should return error
		if err == nil {
			t.Errorf("Expected error for large slice with error, got nil")
		}
		if !strings.Contains(err.Error(), "error at position 500") {
			t.Errorf("Expected error message to contain 'error at position 500', got %v", err.Error())
		}
		if got != nil {
			t.Errorf("Expected nil result on error, got %v", got)
		}
	})

	// Test case 5: Custom error types
	t.Run("custom error types", func(t *testing.T) {

		got, err := SliceToSliceWithError([]int{1, 2, 3}, func(i int) (string, error) {
			if i == 2 {
				return "", ValidationError{Field: "number", Value: i, Message: "invalid number"}
			}
			return fmt.Sprintf("num_%d", i), nil
		})

		if err == nil {
			t.Errorf("Expected ValidationError, got nil")
		}
		if got != nil {
			t.Errorf("Expected nil result on error, got %v", got)
		}

		// Check error type
		var validationErr ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected ValidationError type, got %T", err)
		}
	})
}
