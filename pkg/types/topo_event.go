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

// Package types define all common types used in nodeman runtime.
// Everything from API or Database should be converted into types in this package before using.
package types

import "time"

// TopoEventType represents the type of topology event.
type TopoEventType string

// TopoEventTypeListToStringList converts a topoevent type list to a string list.
func TopoEventTypeListToStringList(eventTypeList []TopoEventType) []string {
	data := make([]string, len(eventTypeList))
	for idx, eventType := range eventTypeList {
		data[idx] = string(eventType)
	}

	return data
}

// StringListToTopoEventTypeList converts a string list to a topoevent type list.
func StringListToTopoEventTypeList(stringList []string) []TopoEventType {
	data := make([]TopoEventType, len(stringList))
	for idx, eventType := range stringList {
		data[idx] = TopoEventType(eventType)
	}

	return data
}

const (
	// TopoEventNetworkAreaCreate represents the event of creating a network area.
	TopoEventNetworkAreaCreate TopoEventType = "networkarea-create"
	// TopoEventNetworkAreaUpdate represents the event of updating a network area.
	TopoEventNetworkAreaUpdate TopoEventType = "networkarea-update"
	// TopoEventNetworkAreaDelete represents the event of deleting a network area.
	TopoEventNetworkAreaDelete TopoEventType = "networkarea-delete"
	// TopoEventNetworkUnitCreate represents the event of creating a network unit.
	TopoEventNetworkUnitCreate TopoEventType = "networkunit-create"
	// TopoEventNetworkUnitUpdate represents the event of updating a network unit.
	TopoEventNetworkUnitUpdate TopoEventType = "networkunit-update"
	// TopoEventNetworkUnitDelete represents the event of deleting a network unit.
	TopoEventNetworkUnitDelete TopoEventType = "networkunit-delete"
	// TopoEventAccessPointCreate represents the event of creating an access point.
	TopoEventAccessPointCreate TopoEventType = "accesspoint-create"
	// TopoEventAccessPointUpdate represents the event of updating an access point.
	TopoEventAccessPointUpdate TopoEventType = "accesspoint-update"
	// TopoEventAccessPointDelete represents the event of deleting an access point.
	TopoEventAccessPointDelete TopoEventType = "accesspoint-delete"
)

// TopoEvent represents an event record of topology.
type TopoEvent struct {
	TenantID        string
	Type            TopoEventType
	NetworkAreaID   int64
	NetworkAreaName string
	NetworkUnitID   int64
	NetworkUnitName string
	AccessPointID   int64
	AccessPointName string
	OperateTime     time.Time
	Operator        string
}
