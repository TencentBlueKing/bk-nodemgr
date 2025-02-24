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

// Business represents a cmdb business under a tenant.
type Business struct {
	// belongs to.
	TenantID string

	// biz-id is the unique identifier for a business.
	BizID   int64
	BizName string
}

// Host represents a cmdb host.
type Host struct {
	// belongs to
	TenantID      string
	NetworkAreaID int64
	UnitID        int64
	BizID         int64

	// host-id is the unique identifier for a host.
	HostID int64

	// host information.
	InnerIP string
	Mac     string
	OSType  string
}

// NetworkArea represents a cmdb network-area. In which IPs will not be duplicated.
type NetworkArea struct {
	// belongs to
	TenantID string

	// area-id is the unique identifier for a network-area.
	ID int64

	// area-name the name of a network-area.
	Name string
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
	AccessPoints []*AccessPoint

	// links link to upstreams.
	Links map[LinkChannel]*Link
}

// LinkChannel represents a link channel in gse topology.
type LinkChannel string

const (
	// LinkChannelCluster represents a cluster network channel.
	// It is the basic message channel in gse.
	LinkChannelCluster LinkChannel = "cluster"

	// LinkChannelFile represents a file network channel.
	// It is the file transferring channel in gse.
	LinkChannelFile LinkChannel = "file"

	// LinkChannelData represents a data network channel.
	// It is the data transferring channel in gse.
	LinkChannelData LinkChannel = "data"
)

// AccessPoint represents an access point for connecting network units.
type AccessPoint struct {
	// belongs to
	TenantID      string
	NetworkAreaID int64
	UnitID        int64

	// access-point-id is the unique identifier for an access point.
	// access-point-id should be globally unique among all tenants.
	ID int64

	// access-point-name is the name of an access point.
	Name string

	// access configs.
	// each channel should have a endpoint list.
	AccessEndpoints map[LinkChannel][]string
}

// Link represents a link from one unit to one access point.
type Link struct {
	// channel of this link.
	Channel LinkChannel

	TargetTenantID      string
	TargetAccessPointID int64
}
