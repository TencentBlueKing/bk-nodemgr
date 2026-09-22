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

// Package tenant defines the admin tenant APIs.
package tenant

import (
	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type tenantStorage interface {
	EnsureReservedTenant(nCtx contextx.IContext, tenant *types.Tenant) error
}

type userManagerHandler interface {
	ListALLTenants(nCtx contextx.IContext) ([]*types.Tenant, error)
}

type handler struct {
	rg                 *gin.RouterGroup
	syncManager        managerIface.ISyncManager
	tenantStorage      tenantStorage
	userManagerHandler userManagerHandler
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:                 rg.Group("/tenant"),
		syncManager:        capability.Manager,
		tenantStorage:      capability.StorageTenant,
		userManagerHandler: capability.UserManagerHandler,
	}
}

// Load loads admin tenant handlers.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/init", restserver.Handler(h.Init))
}
