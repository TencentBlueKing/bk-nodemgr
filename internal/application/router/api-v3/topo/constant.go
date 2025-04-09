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
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// GetConstant get constant values.
func (h *handler) GetConstant(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to get constant, failed to get request context. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoApplication.TopoConstantGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to get constant, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	result, err := h.backendHandler.GetConstant(sCtx, req.ConvertFieldsToTypes())
	if err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to get constant, failed to get constant values. err: %v", err)
		return nil, errf.ErrWrap(errf.ThirdpartyRequestFailed, err)
	}

	return &protoApplication.TopoConstantGetResp_Data{
		CloudVendor: result.CloudVendor,
		OsType:      result.OSType,
	}, nil
}
