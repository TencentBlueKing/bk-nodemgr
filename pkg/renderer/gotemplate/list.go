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

package gotemplate

import (
	"reflect"
	"text/template"
)

// registerListFunctions register the list functions.
func registerListFunctions(fnMap template.FuncMap) {
	addFunction(fnMap, customList, list)
	addFunction(fnMap, customFirst, first)
	addFunction(fnMap, customLast, last)
}

// list creates a list from the given items.
func list(items ...any) []any {
	return items
}

// first returns the first element of the list, or nil if the list is nil, empty, or not a slice/array.
func first(list any) any {
	if list == nil {
		return nil
	}

	listType := reflect.TypeOf(list).Kind()
	switch listType {
	case reflect.Slice, reflect.Array:
		listValue := reflect.ValueOf(list)

		listLen := listValue.Len()
		if listLen == 0 {
			return nil
		}

		return listValue.Index(0).Interface()
	default:
		return nil
	}
}

// last returns the last element of the list, or nil if the list is nil, empty, or not a slice/array.
func last(list any) any {
	if list == nil {
		return nil
	}

	listType := reflect.TypeOf(list).Kind()
	switch listType {
	case reflect.Slice, reflect.Array:
		listValue := reflect.ValueOf(list)

		listLen := listValue.Len()
		if listLen == 0 {
			return nil
		}

		return listValue.Index(listLen - 1).Interface()
	default:
		return nil
	}
}
