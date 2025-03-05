/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology storage interface.
package topo

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Storage defines the storage interface.
type Storage interface {
	base.Interface

	// UpsertBusiness updates or inserts a business.
	UpsertBusiness(ctx context.Context, biz ...*types.Business) error

	// ListBusinesses lists businesses by page and conditions.
	ListBusinesses(ctx context.Context, page types.Page, conditions ...BusinessCondition) ([]*types.Business, int64, error)

	// UpsertHosts updates or inserts host.
	UpsertHosts(ctx context.Context, host ...*types.Host) error

	// UpsertHostStatics updates or inserts host statics.
	UpsertHostStatics(ctx context.Context, host ...*types.Host) error

	// ListHosts lists hosts by page and conditions.
	ListHosts(ctx context.Context, page types.Page, conditions ...HostCondition) ([]*types.Host, int64, error)

	// ListNetworkArea lists networkarea by page and conditions.
	ListNetworkArea(ctx context.Context, page types.Page, conditions ...NetworkAreaCondition) (
		[]*types.NetworkArea, int64, error)

	// GetNetworkArea gets networkarea by id.
	GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error)

	// UpsertManyNetworkArea updates or inserts networkarea.
	UpsertManyNetworkArea(ctx context.Context, networkAreas ...*types.NetworkArea) error

	// UpdateManyNetworkArea updates networkarea.
	UpdateManyNetworkArea(ctx context.Context, networkArea ...*types.NetworkArea) error

	// DeleteManyNetworkArea deletes networkarea.
	DeleteManyNetworkArea(ctx context.Context, networkAreaIDs ...int64) error

	// ListNetworkUnit lists networkunit by page and conditions.
	ListNetworkUnit(ctx context.Context, page types.Page, conditions ...NetworkUnitCondition) (
		[]*types.NetworkUnit, int64, error)

	// GetNetworkUnit gets networkunit by id.
	GetNetworkUnit(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, error)

	// CreateNetworkUnit creates networkunit.
	CreateNetworkUnit(ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
		int64, error)

	// UpdateNetworkUnit updates networkunit.
	UpdateNetworkUnit(ctx context.Context, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) error

	// DeleteManyNetworkUnit deletes networkunits.
	DeleteManyNetworkUnit(ctx context.Context, networkUnitIDs ...int64) error

	// ListAccessPoint lists accesspoint by page and conditions.
	ListAccessPoint(ctx context.Context, page types.Page, conditions ...AccessPointCondition) (
		[]*types.AccessPoint, int64, error)
}
