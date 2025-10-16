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
	"text/template"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// registerDefaultFunctions register default related functions.
func registerDefaultFunctions(fnMap template.FuncMap) {
	addFunction(fnMap, customDefault, defaultFn)
}

// defaultFn returns the first value if it's non-empty; otherwise, it returns the defaultValue.
func defaultFn(defaultValue any, value ...any) any {
	if len(value) == 0 || conv.IsEmpty(value[0]) {
		return defaultValue
	}

	return value[0]
}
