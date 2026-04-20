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

// ResourceTypeNetworkUnit is the IAM resource type for network unit.
const ResourceTypeNetworkUnit = "networkunit"

// NetworkUnitProvider implements resource.Provider interface for network unit resources.
type NetworkUnitProvider struct {
	storage topo.IStorage
}

// NewNetworkUnitProvider creates a new NetworkUnitProvider.
func NewNetworkUnitProvider(storage topo.IStorage) *NetworkUnitProvider {
	return &NetworkUnitProvider{
		storage: storage,
	}
}

// ListAttr returns empty result as network unit has no attributes.
func (p *NetworkUnitProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	data := ListAttrData([]ResourceAttribute{})

	return &data, nil
}

// ListAttrValue returns empty result as network unit has no attribute values.
func (p *NetworkUnitProvider) ListAttrValue(_ contextx.IContext, _ *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
	data := &ListAttrValueData{
		Count:   0,
		Results: []AttributeValue{},
	}

	return data, nil
}

// ListInstance lists network unit instances with pagination.
// NOTE: NetworkUnit resource has a required parent (networkarea) in IAM permission model.
// When parent=nil, return empty result because IAM will never send list_instance requests
// without specifying the parent network area.
func (p *NetworkUnitProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	// Check if parent is specified
	if req.Filter.Parent == nil {
		data := &ListInstanceData{
			Count:   0,
			Results: []ResourceInstance{},
		}

		return data, nil
	}

	// Parse parent network area ID
	areaID, err := conv.ToInt64(req.Filter.Parent.ID)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to parse parent network area ID")
		return nil, fmt.Errorf("failed to parse parent network area ID: %w", err)
	}

	condition := &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkAreaID: []int64{areaID},
		},
	}

	networkUnits, total, err := p.storage.ListNetworkUnit(ctx, req.Page, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list network units")
		return nil, fmt.Errorf("failed to list network units: %w", err)
	}

	// Convert to IAM response format
	results := make([]ResourceInstance, 0, len(networkUnits))
	for _, unit := range networkUnits {
		results = append(results, ResourceInstance{
			ID:          strconv.FormatInt(unit.ID, 10),
			DisplayName: unit.Name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// FetchInstanceInfo fetches network unit details by IDs.
func (p *NetworkUnitProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
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

	// Query network units by IDs
	condition := &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkUnitID: ids,
		},
	}

	networkUnits, _, err := p.storage.ListNetworkUnit(ctx, types.Page{Limit: len(ids)}, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to fetch network unit info")
		return nil, fmt.Errorf("failed to fetch network unit info: %w", err)
	}

	// Convert to IAM response format
	results := make([]InstanceInfo, 0, len(networkUnits))
	for _, unit := range networkUnits {
		results = append(results, InstanceInfo{
			ID:          strconv.FormatInt(unit.ID, 10),
			DisplayName: unit.Name,
			Attributes: map[string]interface{}{
				AttrIAMPath: BuildIAMPath(ResourceTypeNetworkArea, strconv.FormatInt(unit.NetworkAreaID, 10)),
			},
		})
	}

	data := FetchInstanceInfoData(results)

	return &data, nil
}

// ListInstanceByPolicy lists network unit instances filtered by IAM policy expression.
func (p *NetworkUnitProvider) ListInstanceByPolicy(ctx contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	// Load all network units (use large limit for in-memory evaluation)
	networkUnits, _, err := p.storage.ListNetworkUnit(ctx, types.Page{Offset: 0, Limit: MaxListInstanceByPolicyLimit}, nil)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list network units")
		return nil, fmt.Errorf("failed to list network units: %w", err)
	}

	// Build instances with attributes for expression evaluation
	instances := make([]InstanceForEval, 0, len(networkUnits))
	for _, unit := range networkUnits {
		instances = append(instances, InstanceForEval{
			Instance: ResourceInstance{
				ID:          strconv.FormatInt(unit.ID, 10),
				DisplayName: unit.Name,
			},
			Attributes: map[string]interface{}{},
		})
	}

	// Evaluate expression filter and apply pagination
	return evalExpressionFilter(req.Filter.Expression, ResourceTypeNetworkUnit, instances, req.Page)
}

// SearchInstance searches network units by keyword and parent filter.
func (p *NetworkUnitProvider) SearchInstance(ctx contextx.IContext, req *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	// Check if parent is specified
	if req.Filter.Parent == nil {
		data := &ListInstanceData{
			Count:   0,
			Results: []ResourceInstance{},
		}

		return data, nil
	}

	// Parse parent network area ID
	areaID, err := conv.ToInt64(req.Filter.Parent.ID)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to parse parent network area ID")
		return nil, fmt.Errorf("failed to parse parent network area ID: %w", err)
	}

	keyword := strings.TrimSpace(req.Filter.Keyword)

	// Build condition with parent filter and optional keyword filter
	var condition *types.NetworkUnitCondition
	if keyword != "" {
		condition = &types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkAreaID: []int64{areaID},
			},
			FuzzyInclude: &types.NetworkUnitFuzzyFields{
				NetworkUnitName: []string{regexp.QuoteMeta(keyword)},
			},
		}
	} else {
		condition = &types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{
				NetworkAreaID: []int64{areaID},
			},
		}
	}

	// Query network units
	networkUnits, total, err := p.storage.ListNetworkUnit(ctx, req.Page, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to search network units")
		return nil, fmt.Errorf("failed to search network units: %w", err)
	}

	// Convert to IAM response format
	results := make([]ResourceInstance, 0, len(networkUnits))
	for _, unit := range networkUnits {
		results = append(results, ResourceInstance{
			ID:          strconv.FormatInt(unit.ID, 10),
			DisplayName: unit.Name,
		})
	}

	data := &ListInstanceData{
		Count:   total,
		Results: results,
	}

	return data, nil
}

// FetchInstanceList returns empty result as this is for audit center.
func (p *NetworkUnitProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}

// FetchResourceTypeSchema returns empty schema as network unit has no custom schema.
func (p *NetworkUnitProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	// This method doesn't match any standard IAM callback API
	// Return empty list for now
	data := &ListInstanceData{
		Count:   0,
		Results: []ResourceInstance{},
	}

	return data, nil
}
