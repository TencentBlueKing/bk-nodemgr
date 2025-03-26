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
	"math"
	"reflect"
	"sort"
	"testing"
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

// MapKeyToSlice map key to slice.
func TestMapKeyToSlice(t *testing.T) {
	type args struct {
		m map[string]bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal",
			args: args{
				m: map[string]bool{
					"a": true,
					"b": true,
					"c": true,
					"d": true,
				},
			},
			want: []string{"a", "b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapKeyToSlice(tt.args.m)
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i] < tt.want[j]
			})
			sort.Slice(got, func(i, j int) bool {
				return got[i] < got[j]
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapSetToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
