/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package config provides the config policy router.
package config

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

const (
	initVersion = 1

	minConfigPolicyPriority int64 = 1

	asyncPoolNum  = 10
	asyncPoolSize = 100
)

type handler struct {
	rg                  *gin.RouterGroup
	storageConfigPolicy configpolicy.IStorage
	authorizer          auth.IAuthorizer
	goAsyncPool         goasync.IHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	goAsyncPool, _ := goasync.NewHandler(goasync.HandlerOption{
		PoolNum:               asyncPoolNum,
		PerPoolSize:           asyncPoolSize,
		LoadBalancingStrategy: goasync.LoadBalancingStrategyLeastFirst,
	})

	return &handler{
		rg:                  rg.Group("/config"),
		storageConfigPolicy: capability.StorageConfigPolicy,
		authorizer:          capability.Authorizer,
		goAsyncPool:         goAsyncPool,
	}
}

// Load loads node handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/list", restserver.Handler(h.ListConfigPolicy))
	h.rg.POST("/get", restserver.Handler(h.GetConfigPolicy))
	h.rg.POST("/create", restserver.Handler(h.CreateConfigPolicy))
	h.rg.POST("/update", restserver.Handler(h.UpdateConfigPolicy))
	h.rg.POST("/enable", restserver.Handler(h.EnableConfigPolicy))
	h.rg.POST("/disable", restserver.Handler(h.DisableConfigPolicy))
	h.rg.POST("/delete", restserver.Handler(h.DeleteConfigPolicy))
	h.rg.POST("/reorder_priorities", restserver.Handler(h.ReorderPrioritiesConfigPolicy))
	h.rg.POST("/preview", restserver.Handler(h.PreviewConfigPolicy))

	// package event apis.
	h.rg.POST("/event/list", restserver.Handler(h.ListConfigPolicyEvent))
	h.rg.POST("/event/distinct", restserver.Handler(h.DistinctConfigPolicyEvent))
}
