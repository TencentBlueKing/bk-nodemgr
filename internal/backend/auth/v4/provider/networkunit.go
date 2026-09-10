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

package provider

import (
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ResourceTypeNetworkUnit is the IAM resource type for network units.
const ResourceTypeNetworkUnit = "networkunit"

// NetworkUnitProvider queries network units in the current tenant.
type NetworkUnitProvider struct{ storage topo.IStorage }

// NewNetworkUnitProvider creates a network unit provider.
func NewNetworkUnitProvider(storage topo.IStorage) *NetworkUnitProvider {
	return &NetworkUnitProvider{storage: storage}
}

// ListInstance intersects optional parent and display-name filters.
func (p *NetworkUnitProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	condition := &types.NetworkUnitCondition{}
	if parent := req.Filter.Parent; parent != nil {
		areaID, exists, err := p.findParentNetworkArea(ctx, parent)
		if err != nil {
			return nil, err
		}
		if !exists {
			return newEmptyListInstanceData(), nil
		}
		condition.ExactInclude = &types.NetworkUnitExactFields{NetworkAreaID: []int64{areaID}}
	}
	if req.Filter.Keyword != "" {
		condition.FuzzyInclude = &types.NetworkUnitFuzzyFields{NetworkUnitName: []string{req.Filter.Keyword}}
	}
	units, total, err := p.storage.ListNetworkUnit(ctx, req.Page, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list network units")
		return nil, fmt.Errorf("failed to list network units: %w", err)
	}
	results := make([]ResourceInstance, 0, len(units))
	for _, unit := range units {
		results = append(results, ResourceInstance{ID: strconv.FormatInt(unit.ID, 10), DisplayName: unit.Name})
	}

	return &ListInstanceData{Count: total, Results: results}, nil
}

func (p *NetworkUnitProvider) findParentNetworkArea(ctx contextx.IContext, parent *ParentFilter) (int64, bool, error) {
	if parent.Type != ResourceTypeNetworkArea {
		return 0, false, ErrInvalidArgument
	}
	areaID, err := conv.ToInt64(parent.ID)
	if err != nil {
		return 0, false, fmt.Errorf("%w: invalid parent ID", ErrInvalidArgument)
	}
	if areaID < 0 || strconv.FormatInt(areaID, 10) != parent.ID {
		return 0, false, ErrInvalidArgument
	}
	areas, _, err := p.storage.ListNetworkArea(ctx, types.SingleItemPage(), &types.NetworkAreaCondition{
		ExactInclude: &types.NetworkAreaExactFields{NetworkAreaID: []int64{areaID}},
	})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to fetch parent network area")
		return 0, false, fmt.Errorf("failed to fetch parent network area: %w", err)
	}

	return areaID, len(areas) > 0, nil
}

// FetchInstanceInfo fetches existing network units and their ancestor paths.
func (p *NetworkUnitProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	ids := make([]int64, 0, len(req.Filter.IDs))
	for _, raw := range req.Filter.IDs {
		id, err := conv.ToInt64(raw)
		if err != nil {
			continue
		}
		if strconv.FormatInt(id, 10) != raw {
			continue
		}
		ids = append(ids, id)
	}
	data := FetchInstanceInfoData{}
	if len(ids) == 0 {
		return &data, nil
	}
	condition := &types.NetworkUnitCondition{ExactInclude: &types.NetworkUnitExactFields{NetworkUnitID: ids}}
	units, _, err := p.storage.ListNetworkUnit(ctx, types.Page{Limit: len(ids)}, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to fetch network unit info")
		return nil, fmt.Errorf("failed to fetch network unit info: %w", err)
	}
	for _, unit := range units {
		id := strconv.FormatInt(unit.ID, 10)
		data = append(data, InstanceInfo{ID: id, DisplayName: unit.Name,
			Attributes: buildInstanceAttributes(id, ResourceTypeNetworkArea, strconv.FormatInt(unit.NetworkAreaID, 10))})
	}

	return &data, nil
}
