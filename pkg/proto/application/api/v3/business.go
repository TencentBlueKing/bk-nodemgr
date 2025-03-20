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
	return validateTopoPage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoBusinessListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoBusinessListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoBusinessListReq) ConvertConditionsToTypes() *types.BusinessCondition {
	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		conditions := &types.BusinessCondition{Type: types.ConditionTypeExactInclude}
		conditions.Exact = &types.BusinessExactFields{
			BizID: exactCond.GetBkBizId(),
		}

		return conditions
	}

	// fuzzy conditions.
	if fuzzyCond := x.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		conditions := &types.BusinessCondition{
			Type: types.ConditionTypeFuzzyInclude,
		}
		conditions.Fuzzy = &types.BusinessFuzzyFields{
			BizName: fuzzyCond.GetBkBizName(),
		}

		return conditions
	}

	// default empty conditions.
	return &types.BusinessCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *TopoBusinessListReq) ConvertConditionsFromTypes(condition *types.BusinessCondition) error {
	if condition == nil {
		return nil
	}
	switch condition.Type {
	case types.ConditionTypeExactInclude:
		if condition.Exact != nil {
			x.ExactIncludeConditions = &TopoBusinessListReq_ExactConditions{
				BkBizId: condition.Exact.BizID,
			}
		}

	case types.ConditionTypeFuzzyInclude:
		if condition.Fuzzy != nil {
			x.FuzzyIncludeConditions = &TopoBusinessListReq_FuzzyConditions{
				BkBizName: condition.Fuzzy.BizName,
			}
		}

	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
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

func newEmptyBusiness() *Business {
	return &Business{
		TenantId:  new(string),
		BkBizId:   new(int64),
		BkBizName: new(string),
	}
}
