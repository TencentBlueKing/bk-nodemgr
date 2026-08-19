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

package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// GetConstant get constant values.
func (h *handler) GetConstant(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoConstantGetReq)

	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get constant, failed to decode request body")

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
