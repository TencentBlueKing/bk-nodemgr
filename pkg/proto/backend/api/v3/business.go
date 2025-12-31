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
	return BackendPagingListTimeout
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

func newEmptyBusiness() *Business {
	return &Business{
		TenantId:  new(string),
		BkBizId:   new(int64),
		BkBizName: new(string),
	}
}
