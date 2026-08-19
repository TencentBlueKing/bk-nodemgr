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

// Package topo provides topology Storage interface.
package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the Storage interface.
type IStorage interface {
	basestorage.Interface

	IStorageNetworkArea
	IStorageNetworkUnit
	IStorageAccessPoint
	IStorageBusiness
	IStorageHost
	IStorageTopoEvent
	IStorageDomainGse

	IDomainNodeInstall
}

// IDomainNodeInstall defines the Storage interface for domain node install.
type IDomainNodeInstall interface {
	// ExistDedicatedInstallerProxyHost exists dedicated installer proxy host by network unit id.
	ExistDedicatedInstallerProxyHost(nCtx contextx.IContext, networkUnitIDs []int64) (map[int64]bool, error)

	// GetNetworkUnitByIDs list network unit by unit ids.
	GetNetworkUnitByIDs(nCtx contextx.IContext, networkUnitIDs []int64) ([]*types.NetworkUnit, error)
}

// IStorageTopoEvent this interface defines the operations which is only for topo event.
type IStorageTopoEvent interface {
	// CountTopoEvent counts topo events by conditions.
	CountTopoEvent(nCtx contextx.IContext, conditions ...*types.TopoEventCondition) (int64, error)

	// ListTopoEvent lists topo events by page and conditions.
	ListTopoEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.TopoEventCondition) (
		[]*types.TopoEvent, int64, error)

	// CreateManyTopoEvent creates multiple topo events.
	CreateManyTopoEvent(nCtx contextx.IContext, events ...*types.TopoEvent) error

	// DistinctTopoEvent distincts topoevent fields.
	DistinctTopoEvent(
		nCtx contextx.IContext, request types.TopoEventDistinctRequest, conditions ...*types.TopoEventCondition) (
		*types.TopoEventDistinctResult, error)
}

// IStorageNetworkUnit this interface defines the operations which is only for network unit.
type IStorageNetworkUnit interface {
	// ListNetworkUnit lists networkunit by page and conditions.
	ListNetworkUnit(nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkUnitCondition) (
		[]*types.NetworkUnit, int64, error)

	// GetNetworkUnit gets networkunit by id.
	GetNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, error)

	// CreateNetworkUnit creates networkunit.
	CreateNetworkUnit(nCtx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		int64, *AccessPointResult, error)

	// UpdateNetworkUnit updates networkunit.
	UpdateNetworkUnit(
		nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		*AccessPointResult, error)

	// DeleteManyNetworkUnit deletes networkunits.
	DeleteManyNetworkUnit(nCtx contextx.IContext, networkUnitIDs ...int64) error

	// GetNetworkUnitIDsByAccessPoints returns NetworkUnit IDs that contain the given AccessPoint IDs.
	GetNetworkUnitIDsByAccessPoints(nCtx contextx.IContext, accessPointIDs []int64) ([]int64, error)

	// GetNetworkUnitDistributionByNetworkAreaID get networkunit distribution by network area id.
	GetNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, conditions ...*types.NetworkUnitCondition) (
		map[int64]int64, error)

	// RecommendNetworkUnitByNetworkSegment recommends network units based on network segment rules.
	RecommendNetworkUnitByNetworkSegment(nCtx contextx.IContext, items ...*types.NetworkUnitSegmentRecommendationItem) (
		[]*types.NetworkUnitSegmentRecommendationResult, error)
}

// IStorageBusiness this interface defines the operations which is only for business.
type IStorageBusiness interface {
	// UpsertManyBusiness updates or inserts a business.
	UpsertManyBusiness(nCtx contextx.IContext, biz ...*types.Business) error

	// ListBusinesses lists businesses by page and conditions.
	ListBusinesses(nCtx contextx.IContext, page types.Page, conditions ...*types.BusinessCondition) (
		[]*types.Business, int64, error)
}

// IStorageNetworkArea this interface defines the operations which is only for network area.
type IStorageNetworkArea interface {
	// ListNetworkArea lists networkarea by page and conditions.
	ListNetworkArea(nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkAreaCondition) (
		[]*types.NetworkArea, int64, error)

	// GetNetworkArea gets networkarea by id.
	GetNetworkArea(nCtx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error)

	// UpsertManyNetworkArea updates or inserts networkarea.
	UpsertManyNetworkArea(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) error

	// UpdateManyNetworkArea updates networkarea.
	UpdateManyNetworkArea(nCtx contextx.IContext, networkArea ...*types.NetworkArea) error

	// DeleteManyNetworkArea deletes networkarea.
	DeleteManyNetworkArea(nCtx contextx.IContext, networkAreaIDs ...int64) error
}

// IStorageAccessPoint this interface defines the operations which is only for accesspoint.
type IStorageAccessPoint interface {
	// CountAccessPoint counts accesspoint by conditions.
	CountAccessPoint(nCtx contextx.IContext, conditions ...*types.AccessPointCondition) (int64, error)

	// ListAccessPoint lists accesspoint by page and conditions.
	ListAccessPoint(nCtx contextx.IContext, page types.Page, conditions ...*types.AccessPointCondition) (
		[]*types.AccessPoint, int64, error)
}

// IStorageHost this interface defines the operations which is only for host.
// nolint: interfacebloat
type IStorageHost interface {
	// UpsertManyHost updates or inserts host.
	UpsertManyHost(nCtx contextx.IContext, host ...*types.Host) error

	// CreateManyHost create host.
	CreateManyHost(nCtx contextx.IContext, hosts ...*types.Host) error

	// UpsertManyHostStatic updates or inserts host statics.
	// Deprecated: UpsertManyHostStatic is deprecated, please use UpdateHostStaticFields instead.
	UpsertManyHostStatic(nCtx contextx.IContext, host ...*types.Host) error

	// UpdateManyHostDynamic updates host dynamic.
	UpdateManyHostDynamic(nCtx contextx.IContext, host ...*types.Host) error

	// ListHostWithFields lists hosts by fields and conditions.
	ListHostWithFields(nCtx contextx.IContext, page types.Page, selection *types.HostFieldSelection, conditions ...*types.HostCondition) (
		[]*types.Host, int64, error)

	// ScanAllHostWithFields scans all hosts by fields and conditions.
	ScanAllHostWithFields(nCtx contextx.IContext, selection *types.HostFieldSelection, conditions ...*types.HostCondition) (
		[]*types.Host, error)

	// ListHost lists hosts by page and conditions.
	ListHost(nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) ([]*types.Host, int64, error)

	// ListHostOrderByUpdateTime lists hosts by page and conditions, sorted by operation time.
	ListHostOrderByUpdateTime(nCtx contextx.IContext, page types.Page,
		conditions ...*types.HostCondition) ([]*types.Host, int64, error)

	// DeleteManyHost deletes hosts.
	DeleteManyHost(nCtx contextx.IContext, hostIDs ...int64) error

	// CountHost counts hosts by conditions.
	CountHost(nCtx contextx.IContext, conditions ...*types.HostCondition) (int64, error)

	// CountHostGroupByNetworkUnitID counts hosts by networkunit-id.
	CountHostGroupByNetworkUnitID(nCtx contextx.IContext, conditions ...*types.HostCondition) (map[int64]int64, error)

	// CountHostGroupByBizID counts hosts by biz-id.
	CountHostGroupByBizID(nCtx contextx.IContext, conditions ...*types.HostCondition) (map[int64]int64, error)

	// CountHostGroupBySetID counts hosts by set-id.
	CountHostGroupBySetID(nCtx contextx.IContext, conditions ...*types.HostCondition) (map[int64]int64, error)

	// CountHostGroupByModuleID counts hosts by module-id.
	CountHostGroupByModuleID(nCtx contextx.IContext, conditions ...*types.HostCondition) (map[int64]int64, error)

	// GetHostByID gets host by id.
	GetHostByID(nCtx contextx.IContext, hostID int64) (*types.Host, error)

	// FindHostWithDynamic finds hosts with dynamic fields.
	FindHostWithDynamic(nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) ([]*types.Host, error)

	// UpdateHostStaticFields updates the static fields of a host.
	UpdateHostStaticFields(nCtx contextx.IContext, fields types.HostStaticFields, hosts ...*types.Host) error

	// UpsertHostTopo upserts host topology relation create or update events.
	UpsertHostTopo(nCtx contextx.IContext, hostRels ...*types.HostTopoRelation) error

	// PopHostTopo pops host topology relation delete events.
	PopHostTopo(nCtx contextx.IContext, hostRels ...*types.HostTopoRelation) error

	// UpdateHostDynamicFields updates the dynamic fields of a host.
	UpdateHostDynamicFields(nCtx contextx.IContext, fields types.HostDynamicFields, hosts ...*types.Host) error

	// TouchHostOperationTime marks the given hosts as recently operated by a user
	// or API action, updating the business operation time used for list ordering.
	TouchHostOperationTime(nCtx contextx.IContext, hostIDs ...int64) error

	// DistinctHost distincts host fields.
	DistinctHost(nCtx contextx.IContext, request types.HostDistinctRequest, conditions ...*types.HostCondition) (
		*types.HostDistinctResult, error)

	// GetHostDistributionByNodeRole get host distribution by node role.
	GetHostDistributionByNodeRole(nCtx contextx.IContext, conditions ...*types.HostCondition) (
		map[string]int64, error)

	// GetHostDistributionByNetworkAreaID get host distribution by network area ID.
	GetHostDistributionByNetworkAreaID(nCtx contextx.IContext, conditions ...*types.HostCondition) (
		map[int64]int64, error)

	// GetHostDistributionByNodeVersion get host distribution by node version.
	GetHostDistributionByNodeVersion(nCtx contextx.IContext, conditions ...*types.HostCondition) (
		map[string]int64, error)

	// GetRelayInfosInNetworkUnit gets available Relay Infos in the specified network unit.
	// Returns RelayInfo list with DedicatedInstaller tag and Running status.
	GetRelayInfosInNetworkUnit(nCtx contextx.IContext, networkUnitID int64) ([]*types.RelayInfo, error)

	// GetHostTopoRelationMapping gets host topo relation mapping.
	// Returns a map with host ID as key and HostTopoRelation as value.
	GetHostTopoRelationMapping(nCtx contextx.IContext, hostIDs ...int64) (map[int64]types.HostTopoRelation, error)

	// ExistHost checks if the host exists by conditions.
	ExistHost(nCtx contextx.IContext, conditions ...*types.HostCondition) (bool, error)
}

// IStorageDomainGse this interface defines the operations which is only for domain gse.
type IStorageDomainGse interface {
	// GetV4AgentAccessEndpoints get agent v4 access endpoints.
	GetV4AgentAccessEndpoints(nCtx contextx.IContext, networkUnitID int64) (
		clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error)

	// GetV6AgentAccessEndpoints get agent v6 access endpoints.
	GetV6AgentAccessEndpoints(nCtx contextx.IContext, networkUnitID int64) (
		clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error)

	// GetProxyUpstreamAccessEndpoints get proxy upstream access endpoints.
	GetProxyUpstreamAccessEndpoints(nCtx contextx.IContext, networkUnitID int64) (
		clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error)

	// NeedStaticAccess check host is need static access or not.
	NeedStaticAccess(nCtx contextx.IContext, networkUnitID int64) (bool, error)

	// GetNetworkUnitCustomDeployConfig gets network unit custom deploy config by network unit id and os type.
	GetNetworkUnitCustomDeployConfig(nCtx contextx.IContext, networkUnitID int64, osType criteria.OSType) (*types.CustomDeployConfig, error)
}
