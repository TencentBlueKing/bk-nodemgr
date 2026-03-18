/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeconfig provides shared config rendering and pre-check logic for node deployment.
package nodeconfig

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

// Template is a config template with a unique key and raw content.
type Template struct {
	UniqueKey string
	Content   string
}

// ConfigFieldRegex is the regex of config field.
const ConfigFieldRegex = `__BK_.*?__`

// RenderConfig renders a config template with pre-settings and custom settings from NodeConf.
func RenderConfig(template Template, nodeConf *types.NodeConf) (*orderjson.OrderedData, error) {
	config, err := renderPreSetting(template.Content, nodeConf)
	if err != nil {
		return nil, err
	}

	if nodeConf.CustomSetting != nil {
		for key, value := range nodeConf.CustomSetting {
			if err := renderCustomSetting(template.UniqueKey, config, key, value); err != nil {
				return nil, err
			}
		}
	}

	return config, nil
}

// RenderNodeConfig is a convenience wrapper that looks up the config template by key
// from nodeConf.ConfigTemplate and renders it. Returns an error if the template is missing.
func RenderNodeConfig(configKey string, nodeConf *types.NodeConf) (*orderjson.OrderedData, error) {
	templateContent, ok := nodeConf.ConfigTemplate[configKey]
	if !ok {
		return nil, fmt.Errorf("config template not found: %s", configKey)
	}

	return RenderConfig(Template{UniqueKey: configKey, Content: templateContent}, nodeConf)
}

func renderPreSetting(templateContent string, nodeConf *types.NodeConf) (*orderjson.OrderedData, error) {
	re := regexp.MustCompile(ConfigFieldRegex)
	result := re.FindAllStringSubmatch(templateContent, -1)

	configStr := templateContent

	for _, item := range result {
		key := item[0]

		value, ok := nodeConf.PreSetting[key]
		if !ok {
			continue
		}

		valueOf := reflect.ValueOf(value)
		switch valueOf.Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice:
			bytes, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("json marshal failed. content(%v): %w", value, err)
			}

			configStr = strings.ReplaceAll(configStr, item[0], string(bytes))
		default:
			configStr = strings.ReplaceAll(configStr, item[0], escapeForJSON(fmt.Sprintf("%v", value)))
		}
	}

	config := new(orderjson.OrderedData)
	if err := json.Unmarshal([]byte(configStr), config); err != nil {
		return nil, fmt.Errorf("json unmarshal failed. content(%s): %w", configStr, err)
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

	if keys[0] != uniqueKey {
		return nil
	}
	keys = keys[1:]

	target := config
	for i, k := range keys {
		if i == len(keys)-1 {
			if err := target.Set(k, value); err != nil {
				return err
			}

			break
		}

		subTarget, err := target.Get(k)
		if err != nil {
			subTarget = new(orderjson.OrderedData)
			if err := target.Set(k, subTarget); err != nil {
				return err
			}
		}

		orderData, ok := subTarget.(*orderjson.OrderedData)
		if !ok {
			return fmt.Errorf("key %s is not a valid order data", k)
		}

		target = orderData
	}

	return nil
}

func escapeForJSON(src string) string {
	quoted := strconv.Quote(src)
	return quoted[1 : len(quoted)-1]
}
