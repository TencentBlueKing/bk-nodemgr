/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package node ...
package node

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/nodeconfig"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

// Type aliases for backward compatibility within this package.
type (
	// CheckList is the pre-install check list (alias for nodeconfig.CheckList).
	CheckList = nodeconfig.CheckList
	// DiskRequire describes a disk space requirement (alias for nodeconfig.DiskRequire).
	DiskRequire = nodeconfig.DiskRequire
	// PortPolicy describes a port availability check (alias for nodeconfig.PortPolicy).
	PortPolicy = nodeconfig.PortPolicy
	// NetworkPolicy describes a network connectivity check (alias for nodeconfig.NetworkPolicy).
	NetworkPolicy = nodeconfig.NetworkPolicy
)

// GetCheckList get gse node check list.
func (h *handler) GetCheckList(gCtx *gin.Context) {
	nCtx := contextx.New(gCtx)

	req := new(protoCallback.GetCheckListReq)
	if err := gCtx.BindJSON(req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get gse node check list, failed to decode request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get gse node check list, failed to validate request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	deployInfo, err := h.GetNodeDeploymentInfo(nCtx, req.GetToken())
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get gse node check list, failed to get node deployment info")
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	nodeConf, err := h.GetNodeDeploymentNodeConf(nCtx, req.GetToken())
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get gse node check list, failed to get node deployment conf")
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := nodeconfig.BuildCheckList(deployInfo, nodeConf)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get gse node check list, failed to build check list")
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.JSON(http.StatusOK, conf)
}
