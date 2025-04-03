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
	"errors"
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

// Addressing represents an addressing type.
type Addressing string

const (
	// AddressingDynamic means the addressing type is dynamic.
	AddressingDynamic Addressing = "dynamic"

	// AddressingStatic means the addressing type is static.
	AddressingStatic Addressing = "static"
)

// HostStatic represents a static host under a host.
// static means it is synced from CMDB.
// or sometimes it will be insert first into database in case of syncing latency.
type HostStatic struct {
	// belongs to
	BizID         int64
	NetworkAreaID int64

	// host information.
	HostName   string
	DeptName   string
	InnerIP    string
	InnerIPV6  string
	OuterIP    string
	OuterIPV6  string
	Mac        string
	OSTypeCCID string
	OSType     string
	Arch       string
	Addressing Addressing

	// synced types, do not use this for processing.
	// just use it for comparing and checking.
	SyncedAgentID string
}

// NodeRole represents a node role.
type NodeRole string

// Validate validates the node role.
func (nodeRole NodeRole) Validate() error {
	switch nodeRole {
	case NodeRoleBlank, NodeRoleAgent, NodeRoleProxy:
		return nil
	default:
		return errors.New("invalid node role")
	}
}

// NodeRoleListToStringList converts a node role list to a string list.
func NodeRoleListToStringList(nodeRoleList []NodeRole) []string {
	data := make([]string, len(nodeRoleList))
	for idx, nodeRole := range nodeRoleList {
		data[idx] = string(nodeRole)
	}

	return data
}

// StringListToNodeRoleList converts a string list to a node role list.
func StringListToNodeRoleList(stringList []string) []NodeRole {
	data := make([]NodeRole, len(stringList))
	for idx, nodeRole := range stringList {
		data[idx] = NodeRole(nodeRole)
	}

	return data
}

const (
	// NodeRoleBlank means this node is blank. nothing installed.
	NodeRoleBlank NodeRole = "blank"

	// NodeRoleAgent means this node is an agent.
	NodeRoleAgent NodeRole = "agent"

	// NodeRoleProxy means this node is a proxy.
	NodeRoleProxy NodeRole = "proxy"
)

// NodeStatus represents a node status when node role is not blank.
type NodeStatus string

// Validate validates the node status.
func (nodeStatus NodeStatus) Validate() error {
	switch nodeStatus {
	case NodeStatusUnknown, NodeStatusInit, NodeStatusRunning,
		NodeStatusDamaged, NodeStatusBusy, NodeStatusUpgrading, NodeStatusOffline:
		return nil
	default:
		return errors.New("invalid node status")
	}
}

// NodeStatusListToStringList converts a node status list to a string list.
func NodeStatusListToStringList(nodeStatusList []NodeStatus) []string {
	data := make([]string, len(nodeStatusList))
	for idx, nodeStatus := range nodeStatusList {
		data[idx] = string(nodeStatus)
	}

	return data
}

// StringListToNodeStatusList converts a string list to a node status list.
func StringListToNodeStatusList(stringList []string) []NodeStatus {
	data := make([]NodeStatus, len(stringList))
	for idx, nodeStatus := range stringList {
		data[idx] = NodeStatus(nodeStatus)
	}

	return data
}

const (
	// NodeStatusUnknown means this node status is unknown.
	NodeStatusUnknown NodeStatus = "unknown"

	// NodeStatusInit means this node status is init.
	NodeStatusInit NodeStatus = "init"

	// NodeStatusRunning means this node status is running.
	NodeStatusRunning NodeStatus = "running"

	// NodeStatusDamaged means this node status is damaged.
	NodeStatusDamaged NodeStatus = "damaged"

	// NodeStatusBusy means this node status is busy.
	NodeStatusBusy NodeStatus = "busy"

	// NodeStatusUpgrading means this node status is upgrading.
	NodeStatusUpgrading NodeStatus = "upgrading"

	// NodeStatusOffline means this node status is offline.
	NodeStatusOffline NodeStatus = "offline"
)

// NodeGeneration represents a node generation.
type NodeGeneration int64

const (
	// NodeGeneration1 means this node is the first generation.
	NodeGeneration1 NodeGeneration = 1

	// NodeGeneration2 means this node is the second generation.
	NodeGeneration2 NodeGeneration = 2
)

// Validate validates the node generation.
func (nodeGeneration NodeGeneration) Validate() error {
	switch nodeGeneration {
	case NodeGeneration1, NodeGeneration2:
		return nil
	default:
		return errors.New("invalid node generation")
	}
}

// HostDynamic represents a dynamic host under a host.
// dynamic means it is set by user.
type HostDynamic struct {
	NodeRole       NodeRole
	NodeStatus     NodeStatus
	NodeVersion    string
	NodeGeneration NodeGeneration
	AgentID        string
	NetworkUnitID  int64

	// ProxyAccessDisabled This means that there will be no new proxy access connection establishment for this node.
	// ! This setting does not affect the established connections.
	ProxyAccessDisabled bool

	// ProxyTags represents the tags of this proxy.
	// ! Please make sure to use these tags on a whitelist basis.
	ProxyTags        []ProxyTag
	ProxyClusterPort int64
	ProxyDataPort    int64
	ProxyFilePort    int64
}

// ProxyTag represents a proxy tag.
type ProxyTag string

const (
	// ProxyTagDedicatedInstaller means this node is a dedicated installer.
	ProxyTagDedicatedInstaller = "dedicated_installer"
)

// Validate validates the proxy tag.
func (tag ProxyTag) Validate() error {
	switch tag {
	case ProxyTagDedicatedInstaller:
		return nil
	default:
		return errors.New("invalid proxy tag")
	}
}

// NewBlankNodeDynamic returns a blank node dynamic.
func NewBlankNodeDynamic() *HostDynamic {
	return &HostDynamic{
		NodeRole:       NodeRoleBlank,
		NodeStatus:     NodeStatusUnknown,
		NodeVersion:    "",
		NodeGeneration: 0,
		AgentID:        "",
		NetworkUnitID:  -1,
	}
}

// Host represents a host.
type Host struct {
	HostID   int64
	TenantID string

	Static  *HostStatic
	Dynamic *HostDynamic
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
