/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology Storage interface.
package topo

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the Storage interface.
type IStorage interface {
	base.Interface

	IDaoNetworkArea
	IDaoNetworkUnit
	IDaoAccessPoint
	IDaoBusiness
	IDaoHost
	IDaoTopoEvent
	IDomainGse
}

// IDaoTopoEvent this interface defines the operations which is only for topo event.
type IDaoTopoEvent interface {
	// CountTopoEvent counts topo events by conditions.
	CountTopoEvent(ctx context.Context, conditions ...types.TopoEventCondition) (int64, error)

	// ListTopoEvent lists topo events by page and conditions.
	ListTopoEvent(ctx context.Context, page types.Page, conditions ...types.TopoEventCondition) (
		[]*types.TopoEvent, int64, error)

	// CreateManyTopoEvent creates multiple topo events.
	CreateManyTopoEvent(ctx context.Context, events ...*types.TopoEvent) error
}

// IDaoNetworkUnit this interface defines the operations which is only for network unit.
type IDaoNetworkUnit interface {
	// ListNetworkUnit lists networkunit by page and conditions.
	ListNetworkUnit(ctx context.Context, page types.Page, conditions ...types.NetworkUnitCondition) (
		[]*types.NetworkUnit, int64, error)

	// GetNetworkUnit gets networkunit by id.
	GetNetworkUnit(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, error)

	// CreateNetworkUnit creates networkunit.
	CreateNetworkUnit(ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		int64, *AccessPointResult, error)

	// UpdateNetworkUnit updates networkunit.
	UpdateNetworkUnit(ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		*AccessPointResult, error)

	// DeleteManyNetworkUnit deletes networkunits.
	DeleteManyNetworkUnit(ctx context.Context, networkUnitIDs ...int64) error
}

// IDaoBusiness this interface defines the operations which is only for business.
type IDaoBusiness interface {
	// UpsertManyBusiness updates or inserts a business.
	UpsertManyBusiness(ctx context.Context, biz ...*types.Business) error

	// ListBusinesses lists businesses by page and conditions.
	ListBusinesses(ctx context.Context, page types.Page, conditions ...types.BusinessCondition) (
		[]*types.Business, int64, error)
}

// IDaoNetworkArea this interface defines the operations which is only for network area.
type IDaoNetworkArea interface {
	// ListNetworkArea lists networkarea by page and conditions.
	ListNetworkArea(ctx context.Context, page types.Page, conditions ...types.NetworkAreaCondition) (
		[]*types.NetworkArea, int64, error)

	// GetNetworkArea gets networkarea by id.
	GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error)

	// UpsertManyNetworkArea updates or inserts networkarea.
	UpsertManyNetworkArea(ctx context.Context, networkAreas ...*types.NetworkArea) error

	// UpdateManyNetworkArea updates networkarea.
	UpdateManyNetworkArea(ctx context.Context, networkArea ...*types.NetworkArea) error

	// DeleteManyNetworkArea deletes networkarea.
	DeleteManyNetworkArea(ctx context.Context, networkAreaIDs ...int64) error
}

// IDaoAccessPoint this interface defines the operations which is only for accesspoint.
type IDaoAccessPoint interface {
	// CountAccessPoint counts accesspoint by conditions.
	CountAccessPoint(ctx context.Context, conditions ...types.AccessPointCondition) (int64, error)

	// ListAccessPoint lists accesspoint by page and conditions.
	ListAccessPoint(ctx context.Context, page types.Page, conditions ...types.AccessPointCondition) (
		[]*types.AccessPoint, int64, error)
}

// IDaoHost this interface defines the operations which is only for host.
type IDaoHost interface {
	// UpsertManyHost updates or inserts host.
	UpsertManyHost(ctx context.Context, host ...*types.Host) error

	// UpsertManyHostStatic updates or inserts host statics.
	UpsertManyHostStatic(ctx context.Context, host ...*types.Host) error

	// UpdateManyHostDynamic updates host dynamic.
	UpdateManyHostDynamic(ctx context.Context, host ...*types.Host) error

	// ListHost lists hosts by page and conditions.
	ListHost(ctx context.Context, page types.Page, conditions ...types.HostCondition) ([]*types.Host, int64, error)

	// CountHost counts hosts by conditions.
	CountHost(ctx context.Context, conditions ...types.HostCondition) (int64, error)

	// GetHostByID gets host by id.
	GetHostByID(ctx context.Context, hostID int64) (*types.Host, error)
}

// IDomainGse this interface defines the operations which is only for domain gse.
type IDomainGse interface {
	// GetAgentAccessEndpoints get agent access endpoints.
	GetAgentAccessEndpoints(ctx context.Context, unitID int64) (clusterEndpoints []string, dataEndpoints []string,
		fileEndpoints []string, err error)
}
