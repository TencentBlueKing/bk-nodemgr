/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides the CMDB mock API handler.
package cmdb

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/common"
	"github.com/gin-gonic/gin"
)

// SearchBusiness search businesses.
func (h *handler) SearchBusiness(gCtx *gin.Context) {
	var req cmdb.SearchBusinessReq
	if err := common.BindJSON(gCtx, &req); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to bind request")
		respondError(gCtx, CodeInvalidParameter, fmt.Sprintf("invalid request body: %v", err))

		return
	}

	// search businesses.
	businesses, err := h.store.SearchBusiness(req.Page)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to search businesses")
		respondError(gCtx, CodeServerError, fmt.Sprintf("failed to search businesses: %v", err))

		return
	}

	respondSuccess(gCtx, &cmdb.SearchBusinessResp{
		Count: h.store.GetBusinessCount(),
		Info:  businesses,
	})
}
