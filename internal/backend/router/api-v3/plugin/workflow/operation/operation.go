/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation defines the workflow operation apis.
package operation

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	pluginMgrIface managerIface.IPluginManager

	daoPluginWorkflow   pluginStg.IDaoPluginWorkflow
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost

	storageWorkflow workflow.IStorage

	authorizer auth.IAuthorizer
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:                  rg.Group("/operation"),
		pluginMgrIface:      capability.Manager,
		daoPluginWorkflow:   capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		storageWorkflow:     capability.StorageWorkflow,
		authorizer:          capability.Authorizer,
	}
}

// Load loads workflow handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListOperation))
	h.rg.POST("/distinct", restserver.Handler(h.DistinctOperation))
	h.rg.POST("/retry", restserver.Handler(h.RetryOperation))
	h.rg.POST("/terminate", restserver.Handler(h.TerminateOperation))
	h.rg.POST("/instance/list", restserver.Handler(h.ListOperationInstance))
	h.rg.POST("/instance/status_distribution/list", restserver.Handler(h.ListOperationInstanceStatusDistribution))
	h.rg.POST("/instance/log/get", restserver.Handler(h.GetOperationInstanceLog))
}
