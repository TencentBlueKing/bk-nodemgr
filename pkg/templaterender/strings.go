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
	"strings"
	"text/template"
)

// registerStringsFunctions register string related template functions.
func registerStringsFunctions(fnMap template.FuncMap) {
	addFunction(fnMap, "upper", strings.ToUpper)
	addFunction(fnMap, "lower", strings.ToLower)
	addFunction(fnMap, "replace", replace)
	addFunction(fnMap, "join", join)
}

// replace is a wrapper of strings.ReplaceAll to be used in template functions.
func replace(oldStr, newStr, s string) string {
	return strings.ReplaceAll(s, oldStr, newStr)
}

// join is a wrapper of strings.Join to be used in template functions.
func join(sep string, s []string) string {
	return strings.Join(s, sep)
}
