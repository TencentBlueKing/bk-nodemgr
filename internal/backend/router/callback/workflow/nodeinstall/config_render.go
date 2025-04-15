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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Template is a template.
type Template struct {
	UniqueKey string
	Content   string
}

// ConfigFieldRegex is the regex of config field.
const ConfigFieldRegex = `__BK_.*?__`

const (
	// UniqueKeyAgent is the unique key of agent.
	UniqueKeyAgent = "agent"

	// UniqueKeyData is the unique key of data.
	UniqueKeyData = "data"

	// UniqueKeyFile is the unique key of file.
	UniqueKeyFile = "file"
)

// RenderConfig render config.
func RenderConfig(template Template, nodeConf *types.NodeConf) (map[string]any, error) {
	config, err := renderPreSetting(template.Content, nodeConf)
	if err != nil {
		return nil, err
	}

	if nodeConf.CustomSetting != nil {
		// append custom setting to config.
		for key, value := range nodeConf.CustomSetting {
			if err := renderCustomSetting(template.UniqueKey, config, key, value); err != nil {
				return nil, err
			}
		}
	}

	return config, nil
}

func renderPreSetting(templateContent string, nodeConf *types.NodeConf) (map[string]any, error) {
	re := regexp.MustCompile(ConfigFieldRegex)
	result := re.FindAllStringSubmatch(templateContent, -1)

	configStr := templateContent

	for _, item := range result {
		key := item[0]

		value, ok := nodeConf.PreSetting[key]
		if !ok {
			// if not found, we ignore it.
			continue
		}

		valueOf := reflect.ValueOf(value)
		switch valueOf.Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice:
			bytes, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}

			configStr = strings.ReplaceAll(configStr, item[0], string(bytes))
		default:
			configStr = strings.ReplaceAll(configStr, item[0], fmt.Sprintf("%v", value))
		}
	}

	config := make(map[string]any)
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, err
	}

	return config, nil
}

func renderCustomSetting(uniqueKey string, config map[string]any, key string, value any) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return fmt.Errorf("key cannot be empty")
	}

	// ignore if key is not belong to this uniqueKey.
	if keys[0] != uniqueKey {
		return nil
	}

	keys = keys[1:]

	target := config

	for i, k := range keys {
		if i == len(keys)-1 {
			target[k] = value
			break
		}

		// Check the existing values
		existingVal, exists := target[k]
		if !exists {
			// If the existing value does not exist, replace it with a new map
			newMap := make(map[string]any)
			target[k] = newMap
			target = newMap
		} else if existingVal == nil {
			// The existing value is nil and replaced with a new map
			newMap := make(map[string]any)
			target[k] = newMap
			target = newMap
		} else {
			// key exists, try to convert to map
			subMap, ok := existingVal.(map[string]any)
			if !ok {
				return fmt.Errorf(
					"path conflict at %q: expected map but got %T in key path %q",
					k, existingVal, key,
				)
			}
			target = subMap
		}
	}

	return nil
}
