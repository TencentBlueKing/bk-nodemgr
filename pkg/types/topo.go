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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
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

// DefaultNetworkAreaID is the default network area.
const DefaultNetworkAreaID = 0

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

	// generation refers to the node generation of this network unit.
	Generation Generation

	// custom deploy config for this network unit.
	CustomDeployConfig map[criteria.OSType]CustomDeployConfig
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

// TopoNodeInfo represents a business instance topology node.
type TopoNodeInfo struct {
	InstID    int64
	InstName  string
	ObjID     string
	HostCount int64
	Children  []*TopoNodeInfo
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

// TopoConstant defines the topo constants.
type TopoConstant struct {
	CloudVendor []string
	OSType      []string
}

// InstallerRuntime defines the installer runtime.
type InstallerRuntime struct {
	BaseWorkDir string
}

// NodeRuntime defines the node runtime.
type NodeRuntime struct {
	BaseDeployDir string
	DataIPC       string
	PluginIPC     string
	LogDir        string
	ZoneID        string
	CityID        string
}

// PluginRuntime defines the plugin runtime.
type PluginRuntime struct {
	BaseDeployDir string
	LogDir        string
}

// CustomDeployConfig defines the custom deploy config for a network unit.
type CustomDeployConfig struct {
	InstallerRuntime InstallerRuntime
	NodeRuntime      NodeRuntime
	PluginRuntime    PluginRuntime
}

// NetworkUnitUpdateFields defines the fields to be updated for a network unit.
type NetworkUnitUpdateFields struct {
	Name               bool
	AccessPoints       bool
	Links              bool
	DirectEndpoints    bool
	CustomDeployConfig bool
}

// NetworkUnitSegmentRecommendationItem represents an item of network unit segment recommendation.
type NetworkUnitSegmentRecommendationItem struct {
	NetworkAreaID int64
	IP            string
}

// NetworkUnitSegmentRecommendationResult represents the result of network unit segment recommendation.
type NetworkUnitSegmentRecommendationResult struct {
	NetworkAreaID int64
	IP            string
	NetworkUnitID int64
	Message       string
}

// NetworkUnitSegmentRuleConfig represents the configuration for network unit segment rules.
// The key is the network area ID in string format.
type NetworkUnitSegmentRuleConfig map[string]NetworkUnitSegmentRuleAreaConfig

// NetworkUnitSegmentRuleAreaConfig represents the rules for a specific network area.
type NetworkUnitSegmentRuleAreaConfig struct {
	Rules []NetworkUnitSegmentRule `json:"rules"`
}

// NetworkUnitSegmentRule represents a single rule mapping CIDRs to a network unit.
type NetworkUnitSegmentRule struct {
	CIDRs         []string `json:"cidrs"`
	NetworkUnitID int64    `json:"bk_networkunit_id"`
}

// NewNetworkUnitUpdateFields creates a new NetworkUnitUpdateFields with all fields set to true.
func NewNetworkUnitUpdateFields() NetworkUnitUpdateFields {
	return NetworkUnitUpdateFields{
		Name:               true,
		AccessPoints:       true,
		Links:              true,
		DirectEndpoints:    true,
		CustomDeployConfig: true,
	}
}
