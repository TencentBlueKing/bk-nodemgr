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

package v3

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoBusinessListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoBusinessListReq) AutoConvert() {
}

const (
	// business list max limit
	maxBusinessLimit = 1000
)

// PageTimeout return page timeout.
func (x *TopoBusinessListReq) PageTimeout() time.Duration {
	return backendPagingListTimeout
}

// PageLimit return page limit.
func (x *TopoBusinessListReq) PageLimit() int {
	return maxBusinessLimit
}

// ConvertPageToTypes convert page to types.
func (x *TopoBusinessListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoBusinessListReq) ConvertConditionsToTypes() *types.BusinessCondition {
	condition := &types.BusinessCondition{}

	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		condition.ExactInclude = &types.BusinessExactFields{
			BizID: exactCond.GetBkBizId(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond := x.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		condition.FuzzyInclude = &types.BusinessFuzzyFields{
			BizName: fuzzyCond.GetBkBizName(),
		}
	}

	return condition
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *TopoBusinessListReq) ConvertConditionsFromTypes(condition *types.BusinessCondition) error {
	if condition == nil {
		return nil
	}

	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &TopoBusinessListReq_ExactConditions{
			BkBizId: condition.ExactInclude.BizID,
		}
	}

	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &TopoBusinessListReq_FuzzyConditions{
			BkBizName: condition.FuzzyInclude.BizName,
		}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return fmt.Errorf("exact-exclude and fuzzy-exclude not supported")
	}

	return nil
}

// ConvertBusinessFromTypes convert business from types.
func (x *TopoBusinessListResp) ConvertBusinessFromTypes(total int64, bizs []*types.Business) {
	items := make([]*Business, len(bizs))
	for idx, biz := range bizs {
		item := newEmptyBusiness()
		*item.TenantId = biz.TenantID
		*item.BkBizId = biz.BizID
		*item.BkBizName = biz.BizName

		items[idx] = item
	}

	x.Data = &TopoBusinessListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertBusinessToTypes convert business to types.
func (x *TopoBusinessListResp) ConvertBusinessToTypes() (int64, []*types.Business) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.Business, len(items))
	for idx, item := range items {
		biz := &types.Business{
			TenantID: *item.TenantId,
			BizID:    *item.BkBizId,
			BizName:  *item.BkBizName,
		}
		result[idx] = biz
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *TopoBusinessInstTopoGetReq) Validate() error {
	if x.GetBkBizId() == 0 {
		return fmt.Errorf("bk_biz_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoBusinessInstTopoGetReq) AutoConvert() {
}

// ConvertBusinessInstTopoFromTypes converts a single business topo root node to response.
func (x *TopoBusinessInstTopoGetResp) ConvertBusinessInstTopoFromTypes(topoNode *types.TopoNodeInfo) {
	x.Data = &TopoBusinessInstTopoGetResp_Data{Items: convertBusinessInstTopoFromTypes(topoNode)}
}

func convertBusinessInstTopoFromTypes(topoNode *types.TopoNodeInfo) *TopoNodeInfo {
	if topoNode == nil {
		return nil
	}

	children := make([]*TopoNodeInfo, 0, len(topoNode.Children))
	for _, child := range topoNode.Children {
		if child == nil {
			continue
		}
		children = append(children, convertBusinessInstTopoFromTypes(child))
	}

	return &TopoNodeInfo{
		TopoInstId:   topoNode.InstID,
		TopoInstName: topoNode.InstName,
		TopoObjId:    topoNode.ObjID,
		HostCount:    topoNode.HostCount,
		Children:     children,
	}
}

// Validate check body.
func (x *TopoBusinessHostCountGetReq) Validate() error {
	if len(x.GetBkBizId()) == 0 {
		return fmt.Errorf("bk_biz_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoBusinessHostCountGetReq) AutoConvert() {
}

// ConvertHostCountFromTypes converts host count map from types to response.
func (x *TopoBusinessHostCountGetResp) ConvertHostCountFromTypes(counts map[int64]int64) {
	items := make([]*BusinessHostCount, 0, len(counts))
	for bizID, hostCount := range counts {
		items = append(items, &BusinessHostCount{BkBizId: bizID, HostCount: hostCount})
	}

	x.Data = &TopoBusinessHostCountGetResp_Data{Items: items}
}

// ConvertHostCountToTypes converts host count from response to map[bizID]hostCount.
func (x *TopoBusinessHostCountGetResp) ConvertHostCountToTypes() map[int64]int64 {
	data := x.GetData()
	if data == nil {
		return nil
	}

	result := make(map[int64]int64, len(data.GetItems()))
	for _, item := range data.GetItems() {
		result[item.GetBkBizId()] = item.GetHostCount()
	}

	return result
}

// ConvertBusinessInstTopoToTypes converts business instance topology from response to a single root types.TopoNodeInfo.
func (x *TopoBusinessInstTopoGetResp) ConvertBusinessInstTopoToTypes() *types.TopoNodeInfo {
	data := x.GetData()
	if data == nil || data.GetItems() == nil {
		return nil
	}

	return convertTopoNodeToTypes(data.GetItems())
}

func convertTopoNodeToTypes(node *TopoNodeInfo) *types.TopoNodeInfo {
	children := make([]*types.TopoNodeInfo, 0, len(node.GetChildren()))
	for _, child := range node.GetChildren() {
		if child == nil {
			continue
		}
		children = append(children, convertTopoNodeToTypes(child))
	}

	info := &types.TopoNodeInfo{
		InstID:    node.GetTopoInstId(),
		InstName:  node.GetTopoInstName(),
		ObjID:     node.GetTopoObjId(),
		HostCount: node.GetHostCount(),
		Children:  children,
	}

	return info
}

func newEmptyBusiness() *Business {
	return &Business{
		TenantId:  new(string),
		BkBizId:   new(int64),
		BkBizName: new(string),
	}
}
