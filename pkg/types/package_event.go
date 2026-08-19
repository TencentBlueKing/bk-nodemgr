/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package types

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// PackageEventType represents the event type of watch event.
type PackageEventType string

const (
	// PackageEventTypePublish represents the publish event type.
	PackageEventTypePublish PackageEventType = "publish"

	// PackageEventTypeDelete represents the delete event type.
	PackageEventTypeDelete PackageEventType = "delete"

	// PackageEventTypeEnable represents the enable release event type.
	PackageEventTypeEnable PackageEventType = "enable"

	// PackageEventTypeDisable represents the disable release event type.
	PackageEventTypeDisable PackageEventType = "disable"

	// PackageEventTypeSetAsDefault represents the set as default event type.
	PackageEventTypeSetAsDefault PackageEventType = "set_as_default"

	// PackageEventTypeCancelAsDefault represents the cancel as default event type.
	PackageEventTypeCancelAsDefault PackageEventType = "cancel_as_default"

	// PackageEventTypeUpload represents the upload event type.
	PackageEventTypeUpload PackageEventType = "upload"
)

// Validate validates the eventType type.
func (eventType PackageEventType) Validate() error {
	switch eventType {
	case PackageEventTypePublish, PackageEventTypeDelete, PackageEventTypeEnable,
		PackageEventTypeDisable, PackageEventTypeSetAsDefault, PackageEventTypeCancelAsDefault,
		PackageEventTypeUpload:
		return nil
	default:
		return fmt.Errorf("invalid eventType type, type(%s)", eventType)
	}
}

// PackageEventTypeListToStringList converts a packageevent type list to a string list.
func PackageEventTypeListToStringList(eventTypeList []PackageEventType) []string {
	data := make([]string, len(eventTypeList))
	for idx, eventType := range eventTypeList {
		data[idx] = string(eventType)
	}

	return data
}

// StringListToPackageEventTypeList converts a string list to a packageevent type list.
func StringListToPackageEventTypeList(stringList []string) []PackageEventType {
	data := make([]PackageEventType, len(stringList))
	for idx, eventType := range stringList {
		data[idx] = PackageEventType(eventType)
	}

	return data
}

// PackageEvent represents the event of package.
type PackageEvent struct {
	TenantID    string
	Name        string
	EventType   PackageEventType
	Generation  Generation
	ReleaseType ReleaseType
	OSType      criteria.OSType
	CPUArch     criteria.CPUArch
	Version     string
	OperateTime time.Time
	Operator    string
}
