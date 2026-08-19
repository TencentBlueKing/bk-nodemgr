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

// Package operation defines the workflow operation apis.
package operation

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg                *gin.RouterGroup
	nodeMgrIface      managerIface.INodeManager
	daoNodeWorkflow   nodeStg.IDaoNodeWorkflow
	daoNodeDeployment nodeStg.IDaoNodeDeployment
	storageWorkflow   workflow.IStorage
	authorizer        auth.IAuthorizer
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                rg.Group("/operation"),
		nodeMgrIface:      capability.Manager,
		daoNodeWorkflow:   capability.StorageNode,
		daoNodeDeployment: capability.StorageNode,
		storageWorkflow:   capability.StorageWorkflow,
		authorizer:        capability.Authorizer,
	}
}

// Load loads operation handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctOperation))
	h.rg.POST("/retry", restserver.Handler(h.RetryOperation))
	h.rg.POST("/terminate", restserver.Handler(h.TerminateOperation))
	h.rg.POST("/manual/info/get", restserver.Handler(h.GetManualInfo))
	h.rg.POST("/offline/info", restserver.Handler(h.GetOfflineInstallInfo))
	h.rg.POST("/offline/result", restserver.Handler(h.SubmitOfflineInstallResult))
	h.rg.POST("/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
	h.rg.POST("/instance/status_distribution/list",
		restserver.Handler(h.ListOperationInstanceStatusDistribution))
}
