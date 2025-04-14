/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package types define all common types used in nodeman runtime.
// Everything from API or Database should be converted into types in this package before using.
package types

import (
	"time"
)

// Business represents a business under a tenant.
type Business struct {
	// belongs to.
	TenantID string

	// biz-id is the unique identifier for a business.
	BizID   int64
	BizName string
}

// NetworkArea represents a network-area. In which IPs will not be duplicated.
type NetworkArea struct {
	// belongs to
	TenantID string

	// area-id is the unique identifier for a network-area.
	ID int64

	// area-name the name of a network-area.
	Name string

	// cloud vendor.
	CloudVendorCCID string
	CloudVendor     string
}

// NetworkUnit represents a basic unit for proxy management.
type NetworkUnit struct {
	// belongs to
	TenantID      string
	NetworkAreaID int64

	// unit-id is the unique identifier for a network-unit.
	// unit-id should be globally unique among all tenants.
	ID int64

	// unit-name is the name of a network-unit.
	Name string

	// access points for this network-unit.
	AccessPoints []int64

	// links link to upstreams.
	Links Links

	// direct unit links to gse server directly.
	IsDirect        bool
	DirectEndpoints *Endpoints
}

// TopoNameMapping represents id to name mapping.
type TopoNameMapping struct {
	NetworkArea map[int64]string
	NetworkUnit map[int64]string
	AccessPoint map[int64]string
	OsType      map[string]string
}

// GetNetworkAreaName gets the name of a network area.
func (tnm *TopoNameMapping) GetNetworkAreaName(id int64) string {
	if tnm.NetworkArea == nil {
		return ""
	}

	if name, ok := tnm.NetworkArea[id]; ok {
		return name
	}

	return ""
}

// GetNetworkUnitName gets the name of a network unit.
func (tnm *TopoNameMapping) GetNetworkUnitName(id int64) string {
	if tnm.NetworkUnit == nil {
		return ""
	}

	if name, ok := tnm.NetworkUnit[id]; ok {
		return name
	}

	return ""
}

// GetAccessPointName gets the name of a access point.
func (tnm *TopoNameMapping) GetAccessPointName(id int64) string {
	if tnm.AccessPoint == nil {
		return ""
	}

	if name, ok := tnm.AccessPoint[id]; ok {
		return name
	}

	return ""
}

// Links represents links.
type Links struct {
	Cluster *Link
	File    *Link
	Data    *Link
}

// Link represents link.
type Link struct {
	NetworkAreaID int64
	NetworkUnitID int64
	AccessPointID int64
}

// AccessPoint represents an access point for connecting network units.
type AccessPoint struct {
	// belongs to
	TenantID      string
	NetworkAreaID int64

	// access-point-id is the unique identifier for an access point.
	// access-point-id should be globally unique among all tenants.
	ID int64

	// access-point-name is the name of an access point.
	Name string

	// access configs.
	Endpoints Endpoints
}

// AccessPointList represents a list of access points.
type AccessPointList []*AccessPoint

// Found finds an access point by id.
func (apList AccessPointList) Found(id int64) (*AccessPoint, bool) {
	for _, accessPoint := range apList {
		if accessPoint.ID == id {
			return accessPoint, true
		}
	}

	return nil, false
}

// Endpoints represents endpoints of access point.
type Endpoints struct {
	Cluster []string
	File    []string
	Data    []string
}

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

// TopoConstant defines the topo constants.
type TopoConstant struct {
	CloudVendor []string
	OSType      []string
}
