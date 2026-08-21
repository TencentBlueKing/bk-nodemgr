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

// Package distinctcache maintains in-memory distinct caches on the application
// service. Two caches are maintained and refreshed by a single background
// goroutine:
//   - Host distinct: keyed by tenant id x role type (agent/proxy).
//   - Process distinct: keyed by tenant id x plugin name.
//
// Both caches are global per tenant (no bk_biz_id filtering) and refreshed
// from the backend service every syncInterval.
package distinctcache

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
// backend requests.
const syncTenantConcurrency = 10

// RoleTypeAgent matches hosts whose node role is agent or blank.
const RoleTypeAgent = "agent"

// RoleTypeProxy matches hosts whose node role is proxy.
const RoleTypeProxy = "proxy"

// Cache holds host and process distinct results per tenant, refreshed
// periodically by a single background goroutine.
type Cache struct {
	backendHandler backend.IHandler
	tenantDao      tenant.IHandler

	hostMu    sync.RWMutex
	hostItems map[string]*hostTenantEntry

	processMu    sync.RWMutex
	processItems map[string]map[string]*types.ProcessDistinctResult
}

// hostTenantEntry holds the host distinct results for a single tenant.
type hostTenantEntry struct {
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
		hostItems:      make(map[string]*hostTenantEntry),
		processItems:   make(map[string]map[string]*types.ProcessDistinctResult),
	}
}

// Start launches the background syncer. It is non-blocking: the first sync runs
// in a goroutine, then a ticker triggers subsequent syncs every syncInterval.
// The syncer stops when nCtx is cancelled.
func (cache *Cache) Start(nCtx contextx.IContext) error {
	go cache.run(nCtx)
	return nil
}

func (cache *Cache) run(nCtx contextx.IContext) {
	// initial sync; failures are logged and retried on the next tick.
	cache.sync(nCtx)

	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nCtx.Done():
			logger.G.Sys().Info("distinct cache syncer stopped")
			return
		case <-ticker.C:
			cache.sync(nCtx)
		}
	}
}

// sync fetches all enabled tenants and refreshes both host and process distinct
// caches. Per-tenant failures are isolated; a failure for one tenant does not
// affect others.
func (cache *Cache) sync(nCtx contextx.IContext) {
	tenants, _, err := cache.tenantDao.List(nCtx, types.UnlimitedPage(), tenant.WithStatus(true))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("distinct cache: failed to list tenants")
		return
	}

	cache.syncHost(nCtx, tenants)
	cache.syncProcess(nCtx, tenants)
}

// ---------------------------------------------------------------------------
// Host distinct cache
// ---------------------------------------------------------------------------

// syncHost fans out per-tenant host distinct syncs across all tenants.
func (cache *Cache) syncHost(nCtx contextx.IContext, tenants []*types.Tenant) {
	agentCond := buildRoleCondition(types.NodeRoleAgent, types.NodeRoleBlank)
	proxyCond := buildRoleCondition(types.NodeRoleProxy)

	gp := gopool.NewPool()
	gp.SetLimit(syncTenantConcurrency)
	for _, t := range tenants {
		gp.Go(func() error {
			cache.syncHostTenant(nCtx, t.ID, agentCond, proxyCond)
			return nil
		})
	}
	// gopool converts goroutine panics into errors returned by Wait; log them
	// instead of swallowing silently so panics in a tenant sync are visible.
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("distinct cache: host sync fan-out error")
	}
}

func (cache *Cache) syncHostTenant(
	nCtx contextx.IContext, tenantID string, agentCond, proxyCond *types.HostCondition) {

	tCtx, err := buildTenantCtx(nCtx, tenantID)
	if err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Error("failed to build tenant context")

		return
	}

	agentResult, err := cache.backendHandler.DistinctHost(tCtx, agentCond)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("tenant-id", tenantID, "role-type", RoleTypeAgent).
			Error("distinct cache: failed to sync host agent distinct")

		return
	}

	proxyResult, err := cache.backendHandler.DistinctHost(tCtx, proxyCond)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("tenant-id", tenantID, "role-type", RoleTypeProxy).
			Error("distinct cache: failed to sync host proxy distinct")

		return
	}

	cache.setHost(tenantID, agentResult, proxyResult)
}

// setHost updates the host cache entry for the given tenant. Caller holds no lock.
func (cache *Cache) setHost(tenantID string, agent, proxy *types.HostDistinctResult) {
	cache.hostMu.Lock()
	defer cache.hostMu.Unlock()
	cache.hostItems[tenantID] = &hostTenantEntry{agent: agent, proxy: proxy}
}

// GetHost returns the cached host distinct result for (tenantID, roleType).
// It returns nil if the entry has not been synced yet.
func (cache *Cache) GetHost(tenantID, roleType string) *types.HostDistinctResult {
	cache.hostMu.RLock()
	defer cache.hostMu.RUnlock()

	entry, ok := cache.hostItems[tenantID]
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

// ---------------------------------------------------------------------------
// Process distinct cache
// ---------------------------------------------------------------------------

// syncProcess fans out per-tenant process distinct syncs across all tenants.
// Each tenant discovers plugin names that have running processes, then fetches
// the full distinct result for each plugin.
func (cache *Cache) syncProcess(nCtx contextx.IContext, tenants []*types.Tenant) {
	gp := gopool.NewPool()
	gp.SetLimit(syncTenantConcurrency)
	for _, t := range tenants {
		gp.Go(func() error {
			cache.syncProcessTenant(nCtx, t.ID)
			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("distinct cache: process sync fan-out error")
	}
}

func (cache *Cache) syncProcessTenant(nCtx contextx.IContext, tenantID string) {
	tCtx, err := buildTenantCtx(nCtx, tenantID)
	if err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Error("failed to build tenant context")

		return
	}

	// Step 1: discover plugin names that have running processes.
	nameSelector := types.ProcessDistinctSelector{PluginName: true}
	nameResult, err := cache.backendHandler.DistinctProcess(tCtx, nameSelector, nil)
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("tenant-id", tenantID).
			Error("distinct cache: failed to discover process plugin names")

		return
	}

	if len(nameResult.PluginName) == 0 {
		cache.setProcess(tenantID, map[string]*types.ProcessDistinctResult{})

		return
	}

	// Step 2: for each plugin name, fetch the full distinct result.
	results := make(map[string]*types.ProcessDistinctResult, len(nameResult.PluginName))
	for _, pluginName := range nameResult.PluginName {
		cond := &types.ProcessCondition{
			ExactInclude: &types.ProcessExactFields{
				PluginName: []string{pluginName},
			},
		}
		result, err := cache.backendHandler.DistinctProcess(
			tCtx, types.NewProcessDistinctSelectorAllSet(), cond)
		if err != nil {
			logger.G.Sys().
				WithErr(err).
				With("tenant-id", tenantID, "plugin-name", pluginName).
				Error("distinct cache: failed to sync process distinct")

			continue
		}
		results[pluginName] = result
	}

	cache.setProcess(tenantID, results)
}

// setProcess updates the process cache entry for the given tenant. Caller holds no lock.
func (cache *Cache) setProcess(tenantID string, results map[string]*types.ProcessDistinctResult) {
	cache.processMu.Lock()
	defer cache.processMu.Unlock()
	cache.processItems[tenantID] = results
}

// GetProcess returns the cached process distinct result for (tenantID, pluginName).
// It returns nil if the entry has not been synced yet.
func (cache *Cache) GetProcess(tenantID, pluginName string) *types.ProcessDistinctResult {
	cache.processMu.RLock()
	defer cache.processMu.RUnlock()

	plugins, ok := cache.processItems[tenantID]
	if !ok {
		return nil
	}

	return plugins[pluginName]
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// buildTenantCtx forks the base context with the tenant id and the system
// virtual user. The backend apigw auth rejects requests whose bk_username is
// empty ("username is required"); the syncer has no request-scoped user, so it
// runs as "bk-nodemgr" (same convention used by backend scheduled workflows).
func buildTenantCtx(nCtx contextx.IContext, tenantID string) (contextx.IContext, error) {
	bkUsername, err := access.GetVirtualUserBKUsername(contextx.From(nCtx, contextx.WithTenantID(tenantID)))
	if err != nil {
		return nil, err
	}

	return contextx.From(nCtx,
		contextx.WithTenantID(tenantID),
		contextx.WithBKUsername(bkUsername),
	), nil
}
