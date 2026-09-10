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

// ResourceTypeNetworkArea is the IAM resource type for network areas.
const ResourceTypeNetworkArea = "networkarea"

// NetworkAreaProvider queries network areas in the current tenant.
type NetworkAreaProvider struct{ storage topo.IStorage }

// NewNetworkAreaProvider creates a network area provider.
func NewNetworkAreaProvider(storage topo.IStorage) *NetworkAreaProvider {
	return &NetworkAreaProvider{storage: storage}
}

// ListInstance lists network areas with display-name filtering and stable DAO ordering.
func (p *NetworkAreaProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	if req.Filter.Parent != nil {
		return nil, ErrInvalidArgument
	}
	condition := &types.NetworkAreaCondition{}
	if req.Filter.Keyword != "" {
		condition.FuzzyInclude = &types.NetworkAreaFuzzyFields{NetworkAreaName: []string{req.Filter.Keyword}}
	}
	areas, total, err := p.storage.ListNetworkArea(ctx, req.Page, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list network areas")
		return nil, fmt.Errorf("failed to list network areas: %w", err)
	}
	results := make([]ResourceInstance, 0, len(areas))
	for _, area := range areas {
		results = append(results, ResourceInstance{ID: strconv.FormatInt(area.ID, 10), DisplayName: area.Name})
	}

	return &ListInstanceData{Count: total, Results: results}, nil
}

// FetchInstanceInfo fetches existing network areas by canonical ID.
func (p *NetworkAreaProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
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
	condition := &types.NetworkAreaCondition{ExactInclude: &types.NetworkAreaExactFields{NetworkAreaID: ids}}
	areas, _, err := p.storage.ListNetworkArea(ctx, types.Page{Limit: len(ids)}, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to fetch network area info")
		return nil, fmt.Errorf("failed to fetch network area info: %w", err)
	}
	for _, area := range areas {
		data = append(data, InstanceInfo{ID: strconv.FormatInt(area.ID, 10), DisplayName: area.Name})
	}

	return &data, nil
}
