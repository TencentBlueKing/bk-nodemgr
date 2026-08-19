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

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerEvent defines the backend Handler for event.
type IHandlerEvent interface {
	IHandlerTopoEvent
	IHandlerPackageEvent
	IHandlerConfigPolicyEvent
}

// IHandlerTopoEvent defines the backend Handler for topo event.
type IHandlerTopoEvent interface {
	// ListTopoEvent list topo events by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the topo-event list with page and the total count with filter and error.
	ListTopoEvent(nCtx contextx.IContext, page types.Page, condition *types.TopoEventCondition) ([]*types.TopoEvent, int64, error)

	// CountTopoEvent count topo events by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the topo-event count with filter and error.
	CountTopoEvent(nCtx contextx.IContext, condition *types.TopoEventCondition) (int64, error)

	// DistinctTopoEvent distinct topo-event by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the topo-event distinct result and error.
	DistinctTopoEvent(nCtx contextx.IContext, condition *types.TopoEventCondition) (*types.TopoEventDistinctResult, error)
}

// IHandlerPackageEvent defines the backend Handler for package event.
type IHandlerPackageEvent interface {
	// ListPackageEvent list package events by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the package-event list with page and the total count with filter and error.
	ListPackageEvent(nCtx contextx.IContext, page types.Page, condition *types.PackageEventCondition) ([]*types.PackageEvent, int64, error)

	// CountTopoEvent count package events by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the package-event count with filter and error.
	CountPackageEvent(nCtx contextx.IContext, condition *types.PackageEventCondition) (int64, error)

	// DistinctPackageEvent distinct package event by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the package-event distinct result and error.
	DistinctPackageEvent(nCtx contextx.IContext, condition *types.PackageEventCondition) (*types.PackageEventDistinctResult, error)
}

// IHandlerConfigPolicyEvent defines the backend Handler for policy event.
type IHandlerConfigPolicyEvent interface {
	// ListConfigPolicyEvent list policy events by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the config-policy-event list with page and the total count with filter and error.
	ListConfigPolicyEvent(nCtx contextx.IContext, page types.Page, condition *types.ConfigPolicyEventCondition) (
		[]*types.ConfigPolicyEvent, int64, error)

	// CountConfigPolicyEvent count policy events by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the config-policy-event count with filter and error.
	CountConfigPolicyEvent(nCtx contextx.IContext, condition *types.ConfigPolicyEventCondition) (int64, error)

	// DistinctConfigPolicyEvent distinct policy event by condition.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the config-policy-event distinct result and error.
	DistinctConfigPolicyEvent(nCtx contextx.IContext, condition *types.ConfigPolicyEventCondition) (*types.ConfigPolicyEventDistinctResult, error)
}

// ===============================================================================
// Topo Event Related Interfaces
// ===============================================================================

// ListTopoEvent list topo event within specified tenant in contextx.
func (h *Handler) ListTopoEvent(nCtx contextx.IContext, page types.Page, condition *types.TopoEventCondition) ([]*types.TopoEvent, int64, error) {
	req := &protoBackend.TopoEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listTopoEvent(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertTopoEventsToTypes()

	return events, total, nil
}

// CountTopoEvent count the number of topo events by conditions.
func (h *Handler) CountTopoEvent(nCtx contextx.IContext, condition *types.TopoEventCondition) (int64, error) {
	req := &protoBackend.TopoEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listTopoEvent(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctTopoEvent distinct the number of topo events by conditions.
func (h *Handler) DistinctTopoEvent(nCtx contextx.IContext, condition *types.TopoEventCondition) (*types.TopoEventDistinctResult, error) {
	req := &protoBackend.TopoEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctTopoEvent(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes(), nil
}

// ===============================================================================
// Package Event Related Interfaces
// ===============================================================================

// ListPackageEvent list package event within specified tenant in contextx.
func (h *Handler) ListPackageEvent(nCtx contextx.IContext, page types.Page, condition *types.PackageEventCondition) (
	[]*types.PackageEvent, int64, error) {

	req := &protoBackend.PackageEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listPackageEvent(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertPackageEventsToTypes()

	return events, total, nil
}

// CountPackageEvent count the number of package events by conditions.
func (h *Handler) CountPackageEvent(nCtx contextx.IContext, condition *types.PackageEventCondition) (int64, error) {
	req := &protoBackend.PackageEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listPackageEvent(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctPackageEvent distinct the number of package events by conditions.
func (h *Handler) DistinctPackageEvent(nCtx contextx.IContext, condition *types.PackageEventCondition) (*types.PackageEventDistinctResult, error) {
	req := &protoBackend.PackageEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctPackageEvent(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes()
}

// ===============================================================================
// Config Policy Event Related Interfaces
// ===============================================================================

// ListConfigPolicyEvent list policy event within specified tenant in contextx.
func (h *Handler) ListConfigPolicyEvent(nCtx contextx.IContext, page types.Page, condition *types.ConfigPolicyEventCondition) (
	[]*types.ConfigPolicyEvent, int64, error) {

	req := &protoBackend.ConfigPolicyEventListReq{
		Page: convertPage(page),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, err
	}

	resp, err := h.cli.listConfigPolicyEvent(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	total, events := resp.ConvertConfigPolicyEventsToTypes()

	return events, total, nil
}

// CountConfigPolicyEvent count the number of policy events by conditions.
func (h *Handler) CountConfigPolicyEvent(nCtx contextx.IContext, condition *types.ConfigPolicyEventCondition) (int64, error) {
	req := &protoBackend.ConfigPolicyEventListReq{
		OnlyCount: true,
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, err
	}

	resp, err := h.cli.listConfigPolicyEvent(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// DistinctConfigPolicyEvent distinct the number of policy events by conditions.
func (h *Handler) DistinctConfigPolicyEvent(nCtx contextx.IContext, condition *types.ConfigPolicyEventCondition) (
	*types.ConfigPolicyEventDistinctResult, error) {

	req := &protoBackend.ConfigPolicyEventDistinctReq{}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, err
	}

	resp, err := h.cli.distinctConfigPolicyEvent(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.ConvertResultToTypes()
}
