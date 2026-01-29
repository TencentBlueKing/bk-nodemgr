/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package notice defines notice api v3 router.
package notice

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg         *gin.RouterGroup
	capability *options.Capability
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:         rg.Group("/notice"),
		capability: capability,
	}
}

// GetCurrentAnnouncements is a placeholder handler for getting current announcements.
// Full implementation will be added in Phase 3.
func (h *handler) GetCurrentAnnouncements(_ restserver.IContext) (interface{}, error) {
	// Placeholder: return empty data structure
	// Full implementation will be added in Phase 3
	return map[string]interface{}{
		"items": []interface{}{},
	}, nil
}

// Load loads notice handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// Register routes
	h.rg.GET("/announcements/current", restserver.Handler(h.GetCurrentAnnouncements))
}
