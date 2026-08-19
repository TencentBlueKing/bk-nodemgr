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

package distinctcache

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Fake backend handler
// ---------------------------------------------------------------------------

// distinctHostCall records a single DistinctHost invocation received by the fake backend.
type distinctHostCall struct {
	tenantID string
	bkUser   string
	roles    []types.NodeRole
}

// distinctProcessCall records a single DistinctProcess invocation.
type distinctProcessCall struct {
	tenantID   string
	bkUser     string
	pluginName []string
	discovery  bool // true when only the PluginName selector was requested
}

// fakeBackendHandler implements backend.IHandler by embedding the interface
// (nil for unused methods) and overriding only DistinctHost and DistinctProcess.
type fakeBackendHandler struct {
	backend.IHandler

	mu          sync.Mutex
	hostCalls   []distinctHostCall
	procCalls   []distinctProcessCall
	agentResult *types.HostDistinctResult
	proxyResult *types.HostDistinctResult

	// per-tenant error overrides; key "" applies to any tenant without a specific entry.
	agentErrByTenant map[string]error
	proxyErrByTenant map[string]error
	// default error used when no per-tenant override is set.
	agentErr error
	proxyErr error

	// process distinct results: pluginName → result.
	procResultByPlugin map[string]*types.ProcessDistinctResult
	// plugin names returned by the discovery call (selector=PluginName only).
	procPluginNames []string
	procErr         error
	// per-tenant discovery error; overrides procErr for the discovery call.
	procDiscoveryErrByTenant map[string]error
	// per-plugin error overrides.
	procErrByPlugin map[string]error
}

func (h *fakeBackendHandler) DistinctHost(
	nCtx contextx.IContext, cond *types.HostCondition) (*types.HostDistinctResult, error) {

	roles := []types.NodeRole(nil)
	if cond != nil && cond.DynamicExactInclude != nil {
		roles = cond.DynamicExactInclude.NodeRole
	}

	h.mu.Lock()
	h.hostCalls = append(h.hostCalls, distinctHostCall{
		tenantID: nCtx.TenantID(),
		bkUser:   nCtx.BKUsername(),
		roles:    roles,
	})
	h.mu.Unlock()

	tenantID := nCtx.TenantID()
	// classify by requested roles: agent scope = [agent, blank] or [agent]; proxy scope = [proxy].
	if len(roles) == 1 && roles[0] == types.NodeRoleProxy {
		err := h.proxyErr
		if e, ok := h.proxyErrByTenant[tenantID]; ok {
			err = e
		}
		return h.proxyResult, err
	}
	err := h.agentErr
	if e, ok := h.agentErrByTenant[tenantID]; ok {
		err = e
	}
	return h.agentResult, err
}

func (h *fakeBackendHandler) DistinctProcess(
	nCtx contextx.IContext, selector types.ProcessDistinctSelector,
	cond *types.ProcessCondition) (*types.ProcessDistinctResult, error) {

	pluginName := []string(nil)
	if cond != nil && cond.ExactInclude != nil {
		pluginName = cond.ExactInclude.PluginName
	}
	isDiscovery := selector.PluginName && !selector.OSType && !selector.CPUArch &&
		!selector.Version && !selector.Status && !selector.PluginGroup && !selector.PluginPkgName

	h.mu.Lock()
	h.procCalls = append(h.procCalls, distinctProcessCall{
		tenantID:   nCtx.TenantID(),
		bkUser:     nCtx.BKUsername(),
		pluginName: pluginName,
		discovery:  isDiscovery,
	})
	h.mu.Unlock()

	tenantID := nCtx.TenantID()

	// discovery call: return only plugin names.
	if isDiscovery {
		if e, ok := h.procDiscoveryErrByTenant[tenantID]; ok {
			return nil, e
		}
		return &types.ProcessDistinctResult{PluginName: h.procPluginNames}, h.procErr
	}

	// per-plugin call: return the result for the requested plugin.
	if len(pluginName) == 1 {
		if e, ok := h.procErrByPlugin[pluginName[0]]; ok {
			return nil, e
		}
		if r, ok := h.procResultByPlugin[pluginName[0]]; ok {
			return r, nil
		}
	}
	return &types.ProcessDistinctResult{}, nil
}

func (h *fakeBackendHandler) recordedHostCalls() []distinctHostCall {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]distinctHostCall, len(h.hostCalls))
	copy(out, h.hostCalls)
	return out
}

func (h *fakeBackendHandler) hostCallCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.hostCalls)
}

func (h *fakeBackendHandler) procCallCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.procCalls)
}

// ---------------------------------------------------------------------------
// Fake tenant DAO
// ---------------------------------------------------------------------------

// fakeTenantDao implements tenant.IHandler by embedding the interface
// (nil for unused methods) and overriding only List.
type fakeTenantDao struct {
	tenant.IHandler

	tenants []*types.Tenant
	err     error
}

func (d *fakeTenantDao) List(
	nCtx contextx.IContext, page types.Page, opts ...tenant.OptFn) ([]*types.Tenant, int64, error) {

	_ = nCtx
	_ = page
	_ = opts
	if d.err != nil {
		return nil, 0, d.err
	}
	return d.tenants, int64(len(d.tenants)), nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestBaseCtx returns a tenant-less base context for the cache syncer.
// syncTenant forks it per-tenant via contextx.From(WithTenantID(...)).
func newTestBaseCtx() contextx.IContext {
	return contextx.Background()
}

func newTestCache(t *testing.T, bh backend.IHandler, td tenant.IHandler) *Cache {
	t.Helper()
	return newCache(bh, td)
}

func sampleHostResult(version string) *types.HostDistinctResult {
	return &types.HostDistinctResult{
		NodeVersion: []string{version},
		OSType:      []string{"linux"},
	}
}

func sampleProcResult(version string) *types.ProcessDistinctResult {
	return &types.ProcessDistinctResult{
		Version:    []string{version},
		PluginName: []string{"plugin-a"},
	}
}

// ===================================================================
// Host: GetHost / setHost
// ===================================================================

func TestGetHost_CacheMiss_ReturnsNil(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})

	assert.Nil(t, c.GetHost("t1", RoleTypeAgent))
	assert.Nil(t, c.GetHost("t1", RoleTypeProxy))
}

func TestGetHost_UnknownRoleType_ReturnsNil(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})
	c.setHost("t1", sampleHostResult("a"), sampleHostResult("p"))

	assert.Nil(t, c.GetHost("t1", "other"))
	assert.Nil(t, c.GetHost("t1", ""))
}

func TestGetHost_AfterSet_ReturnsAgentAndProxy(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})
	agentR := sampleHostResult("agent-v1")
	proxyR := sampleHostResult("proxy-v1")
	c.setHost("t1", agentR, proxyR)

	assert.Same(t, agentR, c.GetHost("t1", RoleTypeAgent))
	assert.Same(t, proxyR, c.GetHost("t1", RoleTypeProxy))
	// other tenant still absent
	assert.Nil(t, c.GetHost("t2", RoleTypeAgent))
}

// ===================================================================
// Host: buildRoleCondition
// ===================================================================

func TestBuildRoleCondition_Agent(t *testing.T) {
	cond := buildRoleCondition(types.NodeRoleAgent, types.NodeRoleBlank)

	require.NotNil(t, cond)
	require.NotNil(t, cond.DynamicExactInclude)
	assert.Equal(t,
		[]types.NodeRole{types.NodeRoleAgent, types.NodeRoleBlank},
		cond.DynamicExactInclude.NodeRole)
}

func TestBuildRoleCondition_Proxy(t *testing.T) {
	cond := buildRoleCondition(types.NodeRoleProxy)

	require.NotNil(t, cond)
	require.NotNil(t, cond.DynamicExactInclude)
	assert.Equal(t, []types.NodeRole{types.NodeRoleProxy}, cond.DynamicExactInclude.NodeRole)
}

// ===================================================================
// Host: sync
// ===================================================================

func TestSyncHost_PopulatesCacheForAllTenants(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	assert.Equal(t, "agent-v1", c.GetHost("tA", RoleTypeAgent).NodeVersion[0])
	assert.Equal(t, "proxy-v1", c.GetHost("tA", RoleTypeProxy).NodeVersion[0])
	assert.Equal(t, "agent-v1", c.GetHost("tB", RoleTypeAgent).NodeVersion[0])
	assert.Equal(t, "proxy-v1", c.GetHost("tB", RoleTypeProxy).NodeVersion[0])

	// 2 tenants * 2 role calls = 4 backend host calls.
	assert.Equal(t, 4, bh.hostCallCount())
}

func TestSync_TenantDaoError_NoMutation(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("v1"),
		},
		procPluginNames: []string{"plugin-a"},
	}
	td := &fakeTenantDao{err: errors.New("db down")}
	c := newTestCache(t, bh, td)

	assert.NotPanics(t, func() { c.sync(newTestBaseCtx()) })

	// no entries populated, no backend calls made.
	assert.Nil(t, c.GetHost("any", RoleTypeAgent))
	assert.Nil(t, c.GetProcess("any", "plugin-a"))
	assert.Equal(t, 0, bh.hostCallCount())
	assert.Equal(t, 0, bh.procCallCount())
}

func TestSyncHost_BackendAgentError_SkipsTenant(t *testing.T) {
	// tenant tA: agent call fails. Expect tA fully skipped (set only when both succeed).
	bh := &fakeBackendHandler{
		agentResult:       sampleHostResult("agent-v1"),
		proxyResult:       sampleHostResult("proxy-v1"),
		agentErrByTenant:  map[string]error{"tA": errors.New("agent backend failed")},
		procPluginNames:   []string{}, // no process plugins to isolate host test
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// tA must be absent (agent error short-circuits before set).
	assert.Nil(t, c.GetHost("tA", RoleTypeAgent))
	assert.Nil(t, c.GetHost("tA", RoleTypeProxy))
	// tB must be present.
	assert.NotNil(t, c.GetHost("tB", RoleTypeAgent))
	assert.NotNil(t, c.GetHost("tB", RoleTypeProxy))
}

func TestSyncHost_BackendProxyError_SkipsTenant(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult:       sampleHostResult("agent-v1"),
		proxyResult:       sampleHostResult("proxy-v1"),
		proxyErrByTenant:  map[string]error{"tA": errors.New("proxy backend failed")},
		procPluginNames:   []string{},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// tA (where proxy fails) must be absent; tB present.
	assert.Nil(t, c.GetHost("tA", RoleTypeAgent))
	assert.Nil(t, c.GetHost("tA", RoleTypeProxy))
	assert.NotNil(t, c.GetHost("tB", RoleTypeAgent))
	assert.NotNil(t, c.GetHost("tB", RoleTypeProxy))
}

func TestSyncHost_EmptyTenantList_NoCalls(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
	}
	td := &fakeTenantDao{tenants: nil}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	assert.Equal(t, 0, bh.hostCallCount())
	assert.Nil(t, c.GetHost("any", RoleTypeAgent))
}

func TestSyncHost_BackendReceivesTenantScopedContext(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// verify backend received the correct tenant-id for every host call.
	calls := bh.recordedHostCalls()
	require.Len(t, calls, 4)

	// the syncer runs as the system virtual user (no request-scoped user).
	expectedUser := access.GetVirtualUser()

	tenantSeen := map[string]bool{}
	for _, call := range calls {
		assert.True(t, call.tenantID == "tA" || call.tenantID == "tB",
			"unexpected tenant-id: %q", call.tenantID)
		tenantSeen[call.tenantID] = true

		assert.Equal(t, expectedUser, call.bkUser,
			"backend call did not carry the system virtual user as bk_username")

		if len(call.roles) == 1 && call.roles[0] == types.NodeRoleProxy {
			assert.Equal(t, []types.NodeRole{types.NodeRoleProxy}, call.roles)
		} else {
			assert.Equal(t,
				[]types.NodeRole{types.NodeRoleAgent, types.NodeRoleBlank}, call.roles)
		}
	}
	assert.True(t, tenantSeen["tA"] && tenantSeen["tB"], "both tenants should have been synced")
}

// ===================================================================
// Process: GetProcess / setProcess
// ===================================================================

func TestGetProcess_CacheMiss_ReturnsNil(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})

	assert.Nil(t, c.GetProcess("t1", "plugin-a"))
}

func TestGetProcess_AfterSet_ReturnsResult(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})
	r := sampleProcResult("v1")
	c.setProcess("t1", map[string]*types.ProcessDistinctResult{"plugin-a": r})

	assert.Same(t, r, c.GetProcess("t1", "plugin-a"))
	// other plugin absent
	assert.Nil(t, c.GetProcess("t1", "plugin-b"))
	// other tenant absent
	assert.Nil(t, c.GetProcess("t2", "plugin-a"))
}

// ===================================================================
// Process: sync
// ===================================================================

func TestSyncProcess_PopulatesCacheForAllPlugins(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{"plugin-a", "plugin-b"},
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("a-v1"),
			"plugin-b": sampleProcResult("b-v1"),
		},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	assert.Equal(t, "a-v1", c.GetProcess("tA", "plugin-a").Version[0])
	assert.Equal(t, "b-v1", c.GetProcess("tA", "plugin-b").Version[0])

	// 1 discovery call + 2 per-plugin calls = 3 process calls.
	assert.Equal(t, 3, bh.procCallCount())
}

func TestSyncProcess_EmptyPluginList_NoPerPluginCalls(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{}, // no plugins
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// only the discovery call happened, no per-plugin calls.
	assert.Equal(t, 1, bh.procCallCount())
	// cache entry exists but is empty.
	assert.Nil(t, c.GetProcess("tA", "any-plugin"))
}

func TestSyncProcess_DiscoveryError_SkipsTenant(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{"plugin-a"},
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("v1"),
		},
		procDiscoveryErrByTenant: map[string]error{
			"tA": errors.New("discovery failed"),
		},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// tA skipped (discovery error), tB present.
	assert.Nil(t, c.GetProcess("tA", "plugin-a"))
	assert.NotNil(t, c.GetProcess("tB", "plugin-a"))
}

func TestSyncProcess_PerPluginError_SkipsPluginOnly(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{"plugin-a", "plugin-b"},
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("a-v1"),
			"plugin-b": sampleProcResult("b-v1"),
		},
		procErrByPlugin: map[string]error{
			"plugin-a": errors.New("plugin-a backend failed"),
		},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// plugin-a skipped, plugin-b present.
	assert.Nil(t, c.GetProcess("tA", "plugin-a"))
	assert.NotNil(t, c.GetProcess("tA", "plugin-b"))
}

func TestSyncProcess_BackendReceivesTenantScopedContext(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{"plugin-a"},
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("v1"),
		},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	expectedUser := access.GetVirtualUser()

	bh.mu.Lock()
	defer bh.mu.Unlock()
	require.Len(t, bh.procCalls, 2) // 1 discovery + 1 per-plugin

	for _, call := range bh.procCalls {
		assert.Equal(t, "tA", call.tenantID)
		assert.Equal(t, expectedUser, call.bkUser)
	}

	// first call is discovery (no plugin_name condition, discovery selector).
	assert.True(t, bh.procCalls[0].discovery, "first call should be discovery")
	assert.Nil(t, bh.procCalls[0].pluginName)

	// second call is per-plugin (plugin_name condition, full selector).
	assert.False(t, bh.procCalls[1].discovery, "second call should be per-plugin")
	assert.Equal(t, []string{"plugin-a"}, bh.procCalls[1].pluginName)
}

// ===================================================================
// Start / run lifecycle
// ===================================================================

func TestStart_NonBlocking_RunsInitialSync(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{"plugin-a"},
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("v1"),
		},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	ctx, cancel := contextx.WithCancel(contextx.Background())
	defer cancel()
	c := newCache(bh, td)

	// Start must return immediately without blocking on the first sync.
	done := make(chan struct{})
	go func() {
		_ = c.Start(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start blocked for too long")
	}

	// the background syncer should populate both caches shortly.
	require.Eventually(t, func() bool {
		return c.GetHost("tA", RoleTypeAgent) != nil &&
			c.GetProcess("tA", "plugin-a") != nil
	}, 2*time.Second, 10*time.Millisecond, "initial sync did not populate cache in time")

	cancel()
}

func TestRun_StopsOnBaseCtxCancel(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleHostResult("agent-v1"),
		proxyResult: sampleHostResult("proxy-v1"),
		procPluginNames: []string{"plugin-a"},
		procResultByPlugin: map[string]*types.ProcessDistinctResult{
			"plugin-a": sampleProcResult("v1"),
		},
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	ctx, cancel := contextx.WithCancel(contextx.Background())
	c := newCache(bh, td)

	go c.run(ctx)

	// give the syncer a moment to finish the initial sync.
	require.Eventually(t, func() bool {
		return c.GetHost("tA", RoleTypeAgent) != nil
	}, 2*time.Second, 10*time.Millisecond)

	cancel()

	// after cancel, the goroutine should stop; wait a bit and assert the call
	// count stops growing (no more syncs fired from the ticker).
	hostCountAtCancel := bh.hostCallCount()
	procCountAtCancel := bh.procCallCount()
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, hostCountAtCancel, bh.hostCallCount(),
		"backend host call count kept growing after baseCtx was cancelled")
	assert.Equal(t, procCountAtCancel, bh.procCallCount(),
		"backend process call count kept growing after baseCtx was cancelled")
}
