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

// ResourceTypeBiz is the IAM resource type for businesses.
const ResourceTypeBiz = "biz"

// BizProvider queries locally synchronized businesses in the current tenant.
type BizProvider struct{ storage topo.IStorageBusiness }

// NewBizProvider creates a business provider.
func NewBizProvider(storage topo.IStorageBusiness) *BizProvider {
	return &BizProvider{storage: storage}
}

// ListInstance lists businesses with display-name filtering and stable ID ordering.
func (p *BizProvider) ListInstance(ctx contextx.IContext, req *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	if req.Filter.Parent != nil {
		return nil, ErrInvalidArgument
	}
	condition := &types.BusinessCondition{}
	if req.Filter.Keyword != "" {
		condition.FuzzyInclude = &types.BusinessFuzzyFields{BizName: []string{req.Filter.Keyword}}
	}
	page := req.Page
	page.Sort = "data.biz_id"
	businesses, total, err := p.storage.ListBusinesses(ctx, page, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to list businesses")
		return nil, fmt.Errorf("failed to list businesses: %w", err)
	}
	results := make([]ResourceInstance, 0, len(businesses))
	for _, business := range businesses {
		results = append(results, ResourceInstance{ID: strconv.FormatInt(business.BizID, 10), DisplayName: business.BizName})
	}

	return &ListInstanceData{Count: total, Results: results}, nil
}

// FetchInstanceInfo fetches existing businesses by canonical ID.
func (p *BizProvider) FetchInstanceInfo(ctx contextx.IContext, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
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
	condition := &types.BusinessCondition{ExactInclude: &types.BusinessExactFields{BizID: ids}}
	businesses, _, err := p.storage.ListBusinesses(ctx, types.Page{Limit: len(ids)}, condition)
	if err != nil {
		logger.G.Biz(ctx).WithErr(err).Error("failed to fetch business info")
		return nil, fmt.Errorf("failed to fetch business info: %w", err)
	}
	for _, business := range businesses {
		data = append(data, InstanceInfo{ID: strconv.FormatInt(business.BizID, 10), DisplayName: business.BizName})
	}

	return &data, nil
}
