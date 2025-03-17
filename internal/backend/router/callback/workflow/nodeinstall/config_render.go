/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstall ...
package nodeinstall

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ConfigFieldRegex is the regex of config field.
const ConfigFieldRegex = `__BK_.*?__`

// RenderConfig render config.
func RenderConfig(template string, preSetting map[string]any, customSetting map[string]any) (string, error) {
	re := regexp.MustCompile(ConfigFieldRegex)
	result := re.FindAllStringSubmatch(template, -1)

	configStr := template

	for _, item := range result {
		key := item[0]

		value, ok := preSetting[key]
		if !ok {
			// TODO: 临时方案，后续需要支持是否设置默认值
			continue
		}

		valueOf := reflect.ValueOf(value)
		switch valueOf.Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice:
			bytes, err := json.Marshal(value)
			if err != nil {
				return "", err
			}

			configStr = strings.ReplaceAll(configStr, item[0], string(bytes))
		default:
			configStr = strings.ReplaceAll(configStr, item[0], fmt.Sprintf("%v", value))
		}
	}

	if customSetting == nil {
		return configStr, nil
	}

	config := make(map[string]any)
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return "", err
	}

	// append custom setting to config.
	for key, value := range customSetting {
		// don't override existed setting
		if _, ok := config[key]; ok {
			continue
		}

		config[key] = value
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	configStr = string(jsonBytes)

	return configStr, nil
}
