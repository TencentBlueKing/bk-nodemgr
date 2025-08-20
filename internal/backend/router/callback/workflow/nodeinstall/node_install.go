/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstall ...
package nodeinstall

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg      *gin.RouterGroup
	logger  logger.Logger
	crypter crypter.Crypter
	nodedeployment.IStorageNodeDeployment
	operinstdata.IDomainNodeInstall
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                     rg.Group("/node_install"),
		logger:                 capability.Logger,
		crypter:                capability.Crypter,
		IStorageNodeDeployment: capability.StorageNodeDeployment,
		IDomainNodeInstall:     capability.StorageOperInst,
	}
}

// Load ter register the api v3 router.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// TODO: 设置权限封禁
	h.rg.POST("/report_log", h.ReportLog)
	h.rg.POST("/report_status", h.ReportStatus)
	h.rg.POST("/report_data", h.ReportData)
	h.rg.POST("/get_check_list", h.GetCheckList)
	h.rg.POST("/get_agent_config", h.GetAgentConfig)
	h.rg.POST("/get_data_proxy_config", h.GetDataProxyConfig)
	h.rg.POST("/get_file_proxy_config", h.GetFileProxyConfig)
	h.rg.POST("/relay/report_file_state", h.RelayReportFileState)
	h.rg.POST("/relay/report_storage_result", h.RelayReportStorageResult)
}
