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
}
