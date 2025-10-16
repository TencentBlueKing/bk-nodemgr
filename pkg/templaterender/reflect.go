/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package templaterender

import (
	"reflect"
	"text/template"
)

// registerReflectFunctions register the reflect functions.
func registerReflectFunctions(fnMap template.FuncMap) {
	addFunction(fnMap, customKindIs, kindIs)
	addFunction(fnMap, customKindOf, kindOf)
}

// kindIs checks if the kind of src is the same as target.
func kindIs(target string, src interface{}) bool {
	return target == kindOf(src)
}

// kindOf returns the kind of src.
func kindOf(src interface{}) string {
	return reflect.ValueOf(src).Kind().String()
}
