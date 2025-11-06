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
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// registerConversionFunctions register conversion related functions.
func registerConversionFunctions(fnMap template.FuncMap) {
	addFunction(fnMap, customInt, toInt)
	addFunction(fnMap, customToString, toString)
	addFunction(fnMap, customToJSON, toJSON)
}

func toInt(value any) (int, error) {
	res, err := conv.ToInt64(value)
	if err != nil {
		return 0, err
	}

	return int(res), nil
}

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

func toJSON(v interface{}) (string, error) {
	output, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
