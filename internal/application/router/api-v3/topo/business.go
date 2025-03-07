/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxBusinessLimit = 1000
)

// ListBusiness list business with specified conditions.
func (h *handler) ListBusiness(ctx *rest.Context) (interface{}, error) {
	req := new(topoBusinessListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list business, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list business, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	bizs, num, err := h.backendHandler.ListBusiness(
		sCtx,
		generatePage(req.GetPage(), maxBusinessLimit),
		generateBusinessConditions(req))
	if err != nil {
		h.logger.Errorf("failed to list business, err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	items := make([]*proto.Business, len(bizs))
	for idx, biz := range bizs {
		item := newEmptyBusiness()
		*item.TenantId = biz.TenantID
		*item.BkBizId = biz.BizID
		*item.BkBizName = biz.BizName

		items[idx] = item
	}

	resp := &proto.TopoBusinessListResp_Data{
		Total: num,
		Items: items,
	}

	return resp, nil
}

func generateBusinessConditions(req *topoBusinessListReq) *types.BusinessCondition {
	// exact conditions.
	if exactCond := req.GetExactIncludeConditions(); exactCond != nil {
		conditions := &types.BusinessCondition{Type: types.ConditionTypeExactInclude}
		conditions.Exact = &types.BusinessExactFields{
			BizID: exactCond.GetBkBizId(),
		}

		return conditions
	}

	// fuzzy conditions.
	if fuzzyCond := req.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		conditions := &types.BusinessCondition{
			Type: types.ConditionTypeFuzzyInclude,
		}
		conditions.Fuzzy = &types.BusinessFuzzyFields{
			BizName: fuzzyCond.GetBkBizName(),
		}

		return conditions
	}

	// default empty conditions.
	return nil
}

func newEmptyBusiness() *proto.Business {
	return &proto.Business{
		TenantId:  new(string),
		BkBizId:   new(int64),
		BkBizName: new(string),
	}
}

type topoBusinessListReq struct {
	proto.TopoBusinessListReq
}

// Validate check body.
func (req *topoBusinessListReq) Validate() error {
	return validateTopoPage(req.GetPage())
}
