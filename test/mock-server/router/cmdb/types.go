/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides mock CMDB server types.
package cmdb

import "fmt"

// CMDB-standard error codes. Reference: CMDB API error codes.
const (
	// CodeOK is the success code.
	CodeOK = 0

	// CodeInvalidParameter is the invalid parameter code.
	CodeInvalidParameter = 1199000

	// CodeServerError is the server internal error code.
	CodeServerError = 1199001
)

// ResourceType represents the type of event resource.
type ResourceType string

const (
	// ResourceTypeHost represents the host resource type.
	ResourceTypeHost ResourceType = "host"

	// ResourceTypeHostRelation represents the host relation resource type.
	ResourceTypeHostRelation ResourceType = "host_relation"
)

// EventType represents the event type of watch event.
type EventType string

const (
	// EventTypeCreate represents the create event type.
	EventTypeCreate EventType = "create"

	// EventTypeUpdate represents the update event type.
	EventTypeUpdate EventType = "update"

	// EventTypeDelete represents the delete event type.
	EventTypeDelete EventType = "delete"
)

// CMDB watch event JSON field names.
const (
	// WatchFieldCursor is the field name for cursor in watch event.
	WatchFieldCursor = "bk_cursor"

	// WatchFieldResource is the field name for resource type in watch event.
	WatchFieldResource = "bk_resource"

	// WatchFieldEventType is the field name for event type in watch event.
	WatchFieldEventType = "bk_event_type"

	// WatchFieldDetail is the field name for detail in watch event.
	WatchFieldDetail = "bk_detail"
)

// Config holds the configuration for CMDB router.
// Currently empty, reserved for future configuration options.
type Config struct {
}

// Validate validates the Config.
func (cfg *Config) Validate() error {
	return nil
}

// MockData holds the preset mock data for CMDB.
type MockData struct {
	// Businesses is the list of mock business data.
	Businesses []BusinessConfig `yaml:"businesses"`
	// Areas is the list of mock cloud areas.
	Areas []CloudAreaConfig `yaml:"areas"`
	// Hosts is the list of mock hosts.
	Hosts []HostConfig `yaml:"hosts"`
}

// BusinessConfig holds the configuration for a single business.
type BusinessConfig struct {
	// BKBizID is the business ID.
	BKBizID int64 `yaml:"bk_biz_id"`
	// BKBizName is the business name.
	BKBizName string `yaml:"bk_biz_name"`
}

// CloudAreaConfig holds the configuration for a single cloud area.
type CloudAreaConfig struct {
	// BKCloudID is the cloud area ID.
	BKCloudID int64 `yaml:"bk_cloud_id"`
	// BKCloudName is the cloud area name.
	BKCloudName string `yaml:"bk_cloud_name"`
}

// HostConfig holds the configuration for a single host.
type HostConfig struct {
	// BKHostID is the host ID.
	BKHostID int64 `yaml:"bk_host_id"`
	// BKBizID is the business ID.
	BKBizID int64 `yaml:"bk_biz_id"`
	// BKCloudID is the cloud area ID.
	BKCloudID int64 `yaml:"bk_cloud_id"`
	// BKInnerIP is the IPv4 inner IP.
	BKInnerIP string `yaml:"bk_inner_ip"`
	// BKInnerIPv6 is the IPv6 inner IP.
	BKInnerIPv6 string `yaml:"bk_inner_ipv6"`
	// BKOSType is the OS type.
	BKOSType string `yaml:"bk_os_type"`
	// BKCpuArch is the CPU architecture.
	BKCpuArch string `yaml:"bk_cpu_arch"`
}

// Validate validates the BusinessConfig.
func (bc *BusinessConfig) Validate() error {
	if bc.BKBizID < 0 {
		return fmt.Errorf("bk_biz_id must be >= 0, bk_biz_id(%d)", bc.BKBizID)
	}
	if bc.BKBizName == "" {
		return fmt.Errorf("bk_biz_name cannot be empty")
	}

	return nil
}

// Validate validates the CloudAreaConfig.
func (cac *CloudAreaConfig) Validate() error {
	if cac.BKCloudID < 0 {
		return fmt.Errorf("bk_cloud_id must be >= 0, bk_cloud_id(%d)", cac.BKCloudID)
	}
	if cac.BKCloudName == "" {
		return fmt.Errorf("bk_cloud_name cannot be empty")
	}

	return nil
}

// Validate validates the HostConfig.
func (hc *HostConfig) Validate() error {
	if hc.BKHostID <= 0 {
		return fmt.Errorf("bk_host_id must be > 0, bk_host_id(%d)", hc.BKHostID)
	}
	if hc.BKBizID < 0 {
		return fmt.Errorf("bk_biz_id must be >= 0, bk_biz_id(%d)", hc.BKBizID)
	}
	if hc.BKCloudID < 0 {
		return fmt.Errorf("bk_cloud_id must be >= 0, bk_cloud_id(%d)", hc.BKCloudID)
	}
	if hc.BKInnerIP == "" && hc.BKInnerIPv6 == "" {
		return fmt.Errorf("bk_inner_ip and bk_inner_ipv6 cannot both be empty")
	}

	return nil
}

// Validate validates the MockData.
func (cfg *MockData) Validate() error {
	for _, biz := range cfg.Businesses {
		if err := biz.Validate(); err != nil {
			return err
		}
	}

	for _, area := range cfg.Areas {
		if err := area.Validate(); err != nil {
			return err
		}
	}

	for _, host := range cfg.Hosts {
		if err := host.Validate(); err != nil {
			return err
		}
	}

	return nil
}
