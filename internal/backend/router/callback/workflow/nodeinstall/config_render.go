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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/orderjson"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Template is a template.
type Template struct {
	UniqueKey string
	Content   string
}

// ConfigFieldRegex is the regex of config field.
const ConfigFieldRegex = `__BK_.*?__`

// RenderConfig render config.
func RenderConfig(template Template, nodeConf *types.NodeConf) (*orderjson.OrderedData, error) {
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

func renderPreSetting(templateContent string, nodeConf *types.NodeConf) (*orderjson.OrderedData, error) {
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
			configStr = strings.ReplaceAll(configStr, item[0], escapeForJson(fmt.Sprintf("%v", value)))
		}
	}

	config := new(orderjson.OrderedData)
	if err := json.Unmarshal([]byte(configStr), config); err != nil {
		return nil, err
	}

	return config, nil
}

func renderCustomSetting(uniqueKey string, config *orderjson.OrderedData, key string, value any) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return errors.New("key cannot be empty")
	}

	// ignore if key is not belong to this uniqueKey.
	if keys[0] != uniqueKey {
		return nil
	}
	keys = keys[1:]

	target := config
	for i, k := range keys {
		if i == len(keys)-1 {
			target.Set(k, value)
			break
		}

		subTarget, err := target.Get(k)
		if err != nil {
			subTarget = new(orderjson.OrderedData)
			target.Set(k, subTarget)
		}

		orderData, ok := subTarget.(*orderjson.OrderedData)
		if !ok {
			return fmt.Errorf("key %s is not a valid order data", k)
		}

		target = orderData
	}

	return nil
}

func escapeForJson(src string) string {
	quoted := strconv.Quote(src)
	return quoted[1 : len(quoted)-1]
}
