/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package provider provides IAM resource callback providers for different resource types.
package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ResourceTypeNetworkArea is the IAM resource type for network area.
const ResourceTypeNetworkArea = "networkarea"

// NetworkAreaProvider implements resource.Provider interface for network area resources.
type NetworkAreaProvider struct {
	storage topo.IStorage
}

// NewNetworkAreaProvider creates a new NetworkAreaProvider.
func NewNetworkAreaProvider(storage topo.IStorage) *NetworkAreaProvider {
	return &NetworkAreaProvider{
		storage: storage,
	}
}

// ListAttr returns empty result as network area has no attributes.
func (p *NetworkAreaProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	data := ListAttrData([]ResourceAttribute{})

	return &data, nil
}

// ListAttrValue returns empty result as network area has no attribute values.
func (p *NetworkAreaProvider) ListAttrValue(_ contextx.IContext, _ *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
	data := &ListAttrValueData{
		Count:   0,
		Results: []AttributeValue{},
	}

	return data, nil
}

// ListInstance lists network area instances with pagination.
func (p *NetworkAreaProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	// Page has been converted and validated in Dispatcher
	// Query network areas
	networkAreas, total, err := p.storage.ListNetworkArea(ctx, req.Page)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list network areas")
		return nil, fmt.Errorf("failed to list network areas: %w", err)
	}

	// Convert to IAM response format
	results := make([]ResourceInstance, 0, len(networkAreas))
	for _, area := range networkAreas {
		results = append(results, ResourceInstance{
			ID:          strconv.FormatInt(area.ID, 10),
			DisplayName: area.Name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// FetchInstanceInfo fetches network area details by IDs.
func (p *NetworkAreaProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	// Convert IDs from []string to []int64
	// Skip invalid IDs with warning to allow partial success
	ids := make([]int64, 0, len(req.Filter.IDs))
	for _, idStr := range req.Filter.IDs {
		id, err := conv.ToInt64(idStr)
		if err != nil {
			// Log warning but continue processing other IDs (partial success strategy)
			logger.G.Biz(ctx).With("id", idStr).Warn("skipped invalid ID in FetchInstanceInfo")
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		data := FetchInstanceInfoData([]InstanceInfo{})
		return &data, nil
	}

	// Query network areas by IDs
	condition := &types.NetworkAreaCondition{
		ExactInclude: &types.NetworkAreaExactFields{
			NetworkAreaID: ids,
		},
	}

	networkAreas, _, err := p.storage.ListNetworkArea(ctx, types.Page{Limit: len(ids)}, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to fetch network area info")
		return nil, fmt.Errorf("failed to fetch network area info: %w", err)
	}

	// Convert to IAM response format
	results := make([]InstanceInfo, 0, len(networkAreas))
	for _, area := range networkAreas {
		results = append(results, InstanceInfo{
			ID:          strconv.FormatInt(area.ID, 10),
			DisplayName: area.Name,
			Attributes:  make(map[string]interface{}),
		})
	}

	data := FetchInstanceInfoData(results)

	return &data, nil
}

// ListInstanceByPolicy lists network area instances filtered by IAM policy expression.
func (p *NetworkAreaProvider) ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	// Load all network areas (use large limit for in-memory evaluation)
	networkAreas, _, err := p.storage.ListNetworkArea(ctx, types.Page{Offset: 0, Limit: MaxListInstanceByPolicyLimit})
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list network areas")
		return nil, fmt.Errorf("failed to list network areas: %w", err)
	}

	// Build instances with attributes for expression evaluation
	instances := make([]InstanceForEval, 0, len(networkAreas))
	for _, area := range networkAreas {
		instances = append(instances, InstanceForEval{
			Instance: ResourceInstance{
				ID:          strconv.FormatInt(area.ID, 10),
				DisplayName: area.Name,
			},
			Attributes: map[string]interface{}{},
		})
	}

	// Evaluate expression filter and apply pagination
	return evalExpressionFilter(req.Filter.Expression, ResourceTypeNetworkArea, instances, req.Page)
}

// SearchInstance searches network areas by keyword.
func (p *NetworkAreaProvider) SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	keyword := strings.TrimSpace(req.Filter.Keyword)

	// Page has been converted and validated in Dispatcher
	// Build condition with keyword filter
	var condition *types.NetworkAreaCondition
	if keyword != "" {
		condition = &types.NetworkAreaCondition{
			FuzzyInclude: &types.NetworkAreaFuzzyFields{
				NetworkAreaName: []string{regexp.QuoteMeta(keyword)},
			},
		}
	}

	// Query network areas
	networkAreas, total, err := p.storage.ListNetworkArea(ctx, req.Page, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to search network areas")
		return nil, fmt.Errorf("failed to search network areas: %w", err)
	}

	// Convert to IAM response format
	results := make([]ResourceInstance, 0, len(networkAreas))
	for _, area := range networkAreas {
		results = append(results, ResourceInstance{
			ID:          strconv.FormatInt(area.ID, 10),
			DisplayName: area.Name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// FetchInstanceList returns empty result as this is for audit center.
func (p *NetworkAreaProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}

// FetchResourceTypeSchema returns empty schema as network area has no custom schema.
func (p *NetworkAreaProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	// This method doesn't match any standard IAM callback API
	// Return empty list for now
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}
