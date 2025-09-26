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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// RegisterStringsFunctions register string related template functions
func RegisterStringsFunctions(fnMap template.FuncMap) {
	AddFunction(fnMap, "upper", strings.ToUpper)
	AddFunction(fnMap, "lower", strings.ToLower)
	AddFunction(fnMap, "replace", replace)
	AddFunction(fnMap, "join", join)
}

// replace is a wrapper of strings.ReplaceAll to be used in template functions
func replace(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

// join is a wrapper of strings.Join to be used in template functions
func join(sep string, s any) string {
	return strings.Join(conv.StrSlice(s), sep)
}
