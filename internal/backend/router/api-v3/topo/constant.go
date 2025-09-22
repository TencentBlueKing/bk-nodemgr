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
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// GetConstant get constant values.
func (h *handler) GetConstant(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.TopoConstantGetReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to get constant, failed to decode request body: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cloudVendors := make([]string, 0)
	osTypes := make([]string, 0)

	if req.GetCloudVendor() {
		cloudVendors = h.cmdbHandler.GetCloudVendors()
	}

	if req.GetOsType() {
		osTypes = h.cmdbHandler.GetOSTypes()
	}

	return &protoBackend.TopoConstantGetResp_Data{
		CloudVendor: cloudVendors,
		OsType:      osTypes,
	}, nil
}
