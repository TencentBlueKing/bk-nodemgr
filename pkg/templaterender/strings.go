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
	"fmt"
	"strings"
	"text/template"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// registerStringsFunctions register string related template functions.
func registerStringsFunctions(fnMap template.FuncMap) {
	addFunction(fnMap, customUpper, strings.ToUpper)
	addFunction(fnMap, customLower, strings.ToLower)
	addFunction(fnMap, customReplace, replace)
	addFunction(fnMap, customJoin, join)
	addFunction(fnMap, customIndent, indent)
	addFunction(fnMap, customNindent, nindent)
	addFunction(fnMap, customToString, toString)
}

// replace is a wrapper of strings.ReplaceAll to be used in template functions.
func replace(oldStr, newStr, s string) string {
	return strings.ReplaceAll(s, oldStr, newStr)
}

// join is a wrapper of strings.Join to be used in template functions.
func join(sep string, s []string) string {
	return strings.Join(s, sep)
}

// indent indents each line in the given string with the specified number of spaces.
func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.ReplaceAll(v, "\n", "\n"+pad)
}

// nindent indents each line in the given string with the specified number of spaces and adds a newline at the beginning.
func nindent(spaces int, v string) string {
	return "\n" + indent(spaces, v)
}

// toString converts the given value to a string.
func toString(v interface{}) string {
	switch v := v.(type) {
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return conv.ToStringDefault(v, "")
	}
}
