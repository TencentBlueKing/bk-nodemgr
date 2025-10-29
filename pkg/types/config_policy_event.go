/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"fmt"
	"time"
)

// ConfigPolicyEventType defines the policy event type.
type ConfigPolicyEventType string

const (
	// ConfigPolicyEventTypeCreate defines the policy event type create.
	ConfigPolicyEventTypeCreate ConfigPolicyEventType = "create"

	// ConfigPolicyEventTypeUpdate defines the policy event type update.
	ConfigPolicyEventTypeUpdate ConfigPolicyEventType = "update"

	// ConfigPolicyEventTypeDelete defines the policy event type delete.
	ConfigPolicyEventTypeDelete ConfigPolicyEventType = "delete"

	// ConfigPolicyEventTypeEnable defines the policy event type enable.
	ConfigPolicyEventTypeEnable ConfigPolicyEventType = "enable"

	// ConfigPolicyEventTypeDisable defines the policy event type disable.
	ConfigPolicyEventTypeDisable ConfigPolicyEventType = "disable"
)

// ConfigPolicyEvent represents the policy config event.
type ConfigPolicyEvent struct {
	TenantID         string
	ConfigPolicyID   int64
	ConfigPolicyName string
	ConfigPolicyType ConfigPolicyType
	Type             ConfigPolicyEventType
	Version          int64
	OperateTime      time.Time
	Operator         string
}

// Validate validates the config policy event type.
func (eventType ConfigPolicyEventType) Validate() error {
	switch eventType {
	case ConfigPolicyEventTypeCreate, ConfigPolicyEventTypeUpdate, ConfigPolicyEventTypeDelete,
		ConfigPolicyEventTypeEnable, ConfigPolicyEventTypeDisable:
		return nil
	default:
		return fmt.Errorf("invalid policy event type, type(%s)", eventType)
	}
}

// ConfigPolicyEventTypeListToStringList converts a config policy event type list to a string list.
func ConfigPolicyEventTypeListToStringList(eventTypeList []ConfigPolicyEventType) []string {
	data := make([]string, len(eventTypeList))
	for idx, eventType := range eventTypeList {
		data[idx] = string(eventType)
	}

	return data
}

// StringListToConfigPolicyEventTypeList converts a string list to a config policy event type list.
func StringListToConfigPolicyEventTypeList(stringList []string) ([]ConfigPolicyEventType, error) {
	data := make([]ConfigPolicyEventType, len(stringList))
	for idx, eventType := range stringList {
		if err := ConfigPolicyEventType(eventType).Validate(); err != nil {
			return nil, err
		}
		data[idx] = ConfigPolicyEventType(eventType)
	}

	return data, nil
}
