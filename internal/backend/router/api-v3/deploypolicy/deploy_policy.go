/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package deploypolicy provides deploy policy handler.
package deploypolicy

import (
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

const (
	asyncPoolNum  = 50
	asyncPoolSize = 1000
)

type handler struct {
	rg              *gin.RouterGroup
	daoDeployPolicy deploypolicy.IDaoDeployPolicy
	goAsyncPool     goasync.IHandler
	deployPolicyMgr managerIface.IDeployPolicyManager
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	goAsyncPool, _ := goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               asyncPoolNum,
		PerPoolSize:           asyncPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})

	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:              rg.Group("/deploy_policy"),
		daoDeployPolicy: capability.StorageDeployPolicy,
		goAsyncPool:     goAsyncPool,
		deployPolicyMgr: capability.Manager,
	}
}

// Load loads deploy policy handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/create", restserver.Handler(h.Create))
	h.rg.POST("/list", restserver.Handler(h.List))
	h.rg.POST("/update", restserver.Handler(h.Update))
	h.rg.POST("/execute", restserver.Handler(h.Execute))
}
