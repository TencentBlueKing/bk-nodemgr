/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoBusinessListReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoBusinessListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoBusinessListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
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

// ConvertBusinessInstTopoFromTypes converts a single topo root node from types to response.
func (x *TopoBusinessInstTopoGetResp) ConvertBusinessInstTopoFromTypes(topoNode *types.TopoNodeInfo) {
	x.Data = &TopoBusinessInstTopoGetResp_Data{Items: convertTopoNodeFromTypes(topoNode)}
}

func convertTopoNodeFromTypes(node *types.TopoNodeInfo) *TopoNodeInfo {
	if node == nil {
		return nil
	}

	children := make([]*TopoNodeInfo, 0, len(node.Children))
	for _, child := range node.Children {
		if child == nil {
			continue
		}
		children = append(children, convertTopoNodeFromTypes(child))
	}

	return &TopoNodeInfo{
		TopoInstId:   node.InstID,
		TopoInstName: node.InstName,
		TopoObjId:    node.ObjID,
		HostCount:    node.HostCount,
		Children:     children,
	}
}

func newEmptyBusiness() *Business {
	return &Business{
		TenantId:  new(string),
		BkBizId:   new(int64),
		BkBizName: new(string),
	}
}
