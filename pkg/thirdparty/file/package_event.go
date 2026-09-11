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

package file

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoFile "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/file/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPackageEventHandler defines package event query operations.
type IPackageEventHandler interface {
	// ListPackageEvent lists package events by page and conditions.
	ListPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) ([]*types.PackageEvent, int64, error)

	// CountPackageEvent counts package events by conditions.
	CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error)

	// DistinctPackageEvent distincts selected package event fields by conditions.
	DistinctPackageEvent(nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
		*types.PackageEventDistinctResult, error)
}

// ListPackageEvent lists package events with a page and condition.
func (h *handler) ListPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) (
	[]*types.PackageEvent, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}
	req := new(protoFile.PackageEventListReq)
	if err := req.ConvertFromTypes(page, conditions...); err != nil {
		return nil, 0, err
	}
	resp, err := h.cli.listPackageEvent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, 0, err
	}
	total, events := resp.ConvertPackageEventsToTypes()

	return events, total, nil
}

// CountPackageEvent counts package events with a condition.
func (h *handler) CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}
	req := new(protoFile.PackageEventListReq)
	req.ConvertCountFromTypes(conditions...)
	resp, err := h.cli.listPackageEvent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return 0, err
	}

	return resp.GetTotal(), nil
}

// DistinctPackageEvent returns distinct package event values with a condition.
func (h *handler) DistinctPackageEvent(
	nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
	*types.PackageEventDistinctResult, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}
	req := new(protoFile.PackageEventDistinctReq)
	req.ConvertFromTypes(request, conditions...)
	resp, err := h.cli.distinctPackageEvent(nCtx, nCtx.TenantID(), req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes()
}
