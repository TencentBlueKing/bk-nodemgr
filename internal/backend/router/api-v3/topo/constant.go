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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// GetConstant get constant values.
func (h *handler) GetConstant(ctx *rest.Context) (interface{}, error) {
	req := new(protoBackend.TopoConstantGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get constant, failed to decode request body. err: %v", err)
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	cloudVendors := make([]string, 0)
	osTypes := make([]string, 0)

	if req.GetCloudVendor() {
		var err error
		cloudVendors, err = h.cmdbHandler.GetCloudVendors(ctx)
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to get constant, failed to get cloud vendors. err: %v", err)
			return nil, errf.ErrWrap(errf.InvalidParameter, err)
		}
	}

	if req.GetOsType() {
		osTypes = []string{string(criteria.OSLinux), string(criteria.OSWindows), string(criteria.OSDarwin)}
	}

	return &protoBackend.TopoConstantGetResp_Data{
		CloudVendor: cloudVendors,
		OsType:      osTypes,
	}, nil
}
