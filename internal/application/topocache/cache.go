/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topocache maintains an in-memory host distinct cache on the application
// service. The cache is keyed by tenant id and role type (agent/proxy), and is
// refreshed from the backend service by a background goroutine every minute.
package topocache

import (
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// syncInterval is how often the cache is refreshed from the backend.
const syncInterval = 5 * time.Minute

// syncTenantConcurrency caps the number of tenants synced in parallel per sync
// cycle, so a large tenant fleet does not fan out into an unbounded burst of
// backend requests (2 per tenant).
const syncTenantConcurrency = 10

// RoleTypeAgent matches hosts whose node role is agent or blank.
const RoleTypeAgent = "agent"

// RoleTypeProxy matches hosts whose node role is proxy.
const RoleTypeProxy = "proxy"

// Cache holds host distinct results per tenant x role, refreshed periodically.
type Cache struct {
	backendHandler backend.IHandler
	tenantDao      tenant.IHandler

	mu    sync.RWMutex
	items map[string]*tenantEntry
}

// tenantEntry holds the distinct results for a single tenant.
type tenantEntry struct {
	agent *types.HostDistinctResult
	proxy *types.HostDistinctResult
}

// New constructs the cache. It does not start syncing; call Start to launch
// the background syncer with a cancellable context.
func New(backendHandler backend.IHandler, db *mongo.Database) *Cache {
	return newCache(backendHandler, tenant.New(db))
}

// newCache constructs the cache with an injectable tenant DAO. Used by New and tests.
func newCache(backendHandler backend.IHandler, tenantDao tenant.IHandler) *Cache {
	return &Cache{
		backendHandler: backendHandler,
		tenantDao:      tenantDao,
		items:          make(map[string]*tenantEntry),
	}
}

// Start launches the background syncer. It is non-blocking: the first sync runs
// in a goroutine, then a ticker triggers subsequent syncs every syncInterval.
// The syncer stops when nCtx is cancelled.
func (c *Cache) Start(nCtx contextx.IContext) error {
	go c.run(nCtx)
	return nil
}

func (c *Cache) run(nCtx contextx.IContext) {
	// initial sync; failures are logged and retried on the next tick.
	c.sync(nCtx)

	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nCtx.Done():
			logger.G.Sys().Info("topo distinct cache syncer stopped")
			return
		case <-ticker.C:
			c.sync(nCtx)
		}
	}
}

// sync fetches all enabled tenants and refreshes the distinct result for each.
// per-tenant failures are isolated; a failure for one tenant does not affect others.
func (c *Cache) sync(nCtx contextx.IContext) {
	tenants, _, err := c.tenantDao.List(nCtx, types.UnlimitedPage(), tenant.WithStatus(true))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("topo distinct cache: failed to list tenants")
		return
	}

	agentCond := buildRoleCondition(types.NodeRoleAgent, types.NodeRoleBlank)
	proxyCond := buildRoleCondition(types.NodeRoleProxy)

	gp := gopool.NewPool()
	gp.SetLimit(syncTenantConcurrency)
	for _, t := range tenants {
		gp.Go(func() error {
			c.syncTenant(nCtx, t.ID, agentCond, proxyCond)
			return nil
		})
	}
	// gopool converts goroutine panics into errors returned by Wait; log them
	// instead of swallowing silently so panics in a tenant sync are visible.
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("topo distinct cache: sync fan-out error")
	}
}

func (c *Cache) syncTenant(
	nCtx contextx.IContext, tenantID string, agentCond, proxyCond *types.HostCondition) {

	// The backend apigw auth rejects requests whose bk_username is empty
	// ("username is required"). The syncer has no request-scoped user, so it
	// runs as the repo-wide system virtual user "bk-nodemgr" (same convention
	// used by backend scheduled workflows via access.GetVirtualUser()).
	tCtx := contextx.From(nCtx,
		contextx.WithTenantID(tenantID),
		contextx.WithBKUsername(access.GetVirtualUser()),
	)

	agentResult, err := c.backendHandler.DistinctHost(tCtx, agentCond)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("tenant-id", tenantID, "role-type", RoleTypeAgent).
			Error("topo distinct cache: failed to sync agent distinct")

		return
	}

	proxyResult, err := c.backendHandler.DistinctHost(tCtx, proxyCond)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("tenant-id", tenantID, "role-type", RoleTypeProxy).
			Error("topo distinct cache: failed to sync proxy distinct")

		return
	}

	c.set(tenantID, agentResult, proxyResult)
}

// set updates the cache entry for the given tenant. Caller holds no lock.
func (c *Cache) set(tenantID string, agent, proxy *types.HostDistinctResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[tenantID] = &tenantEntry{agent: agent, proxy: proxy}
}

// Get returns the cached distinct result for (tenantID, roleType).
// It returns nil if the entry has not been synced yet.
func (c *Cache) Get(tenantID, roleType string) *types.HostDistinctResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[tenantID]
	if !ok {
		return nil
	}
	switch roleType {
	case RoleTypeAgent:
		return entry.agent
	case RoleTypeProxy:
		return entry.proxy
	default:
		return nil
	}
}

// buildRoleCondition builds a HostCondition that filters hosts by the given node roles.
func buildRoleCondition(roles ...types.NodeRole) *types.HostCondition {
	return &types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NodeRole: roles,
		},
	}
}
