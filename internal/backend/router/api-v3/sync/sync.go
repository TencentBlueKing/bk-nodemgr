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

// Package sync is use to sync data.
package sync

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg      *gin.RouterGroup
	manager manager.IManager
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:      rg.Group("/sync"),
		manager: capability.Manager,
	}
}

// Load load sync handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/cmdb/host", restserver.Handler(h.SyncCmdbHost))
	h.rg.POST("/cmdb/networkarea", restserver.Handler(h.SyncCmdbNetworkArea))

	h.rg.POST("/gse/agent/state", restserver.Handler(h.SyncAgentState))
	h.rg.POST("/gse/agent/state/all", restserver.Handler(h.SyncAllAgentState))
	h.rg.POST("/gse/agent/info", restserver.Handler(h.SyncAgentInfo))
	h.rg.POST("/gse/agent/correct_agent_id", restserver.Handler(h.SyncCorrectAgentID))
	h.rg.POST("/gse/agent/info/alive", restserver.Handler(h.SyncAliveHostAgentInfo))
	h.rg.POST("/gse/plugin/process/info", restserver.Handler(h.SyncAlivePluginProcessInfo))
	h.rg.POST("/gse/plugin/process/info/all", restserver.Handler(h.SyncAllAlivePluginProcessInfo))
	h.rg.POST("/gse/plugin/process/correct_unknown_status", restserver.Handler(h.CorrectUnknownProcessStatus))
	h.rg.POST("/gse/plugin/process/correct_unknown_status/all", restserver.Handler(h.CorrectAllUnknownProcessStatus))

	h.rg.POST("/plugin/ensure_default", restserver.Handler(h.EnsureDefaultPlugin))

	h.rg.POST("/release/shared", restserver.Handler(h.SyncSharedReleases))

	h.rg.POST("/process/clean_orphan", restserver.Handler(h.CleanOrphanProcess))
}
