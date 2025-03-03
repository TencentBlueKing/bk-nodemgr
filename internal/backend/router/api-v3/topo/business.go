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
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	maxBusinessLimit = 1000
)

// ListBusiness list business with specified conditions.
func (h *handler) ListBusiness(ctx *rest.Context) (interface{}, error) {
	req := new(proto.TopoBusinessListReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.Errorf("failed to list business, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if err := validateTopoBusinessListReq(req); err != nil {
		h.logger.Errorf("failed to list business, failed to validate request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list business, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	page := types.Page{}
	if reqPage := req.GetPage(); reqPage != nil {
		page.Offset = int(reqPage.GetOffset())
		page.Limit = int(reqPage.GetLimit())
	}

	bizs, num, err := h.storage.ListBusinesses(sCtx, generatePage(req.GetPage(), maxBusinessLimit))
	if err != nil {
		h.logger.Errorf("failed to list business, err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
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

func validateTopoBusinessListReq(req *proto.TopoBusinessListReq) error {
	if err := validateTopoPage(req.GetPage()); err != nil {
		return err
	}

	return nil
}

func newEmptyBusiness() *proto.Business {
	return &proto.Business{
		TenantId:  new(string),
		BkBizId:   new(int64),
		BkBizName: new(string),
	}
}
