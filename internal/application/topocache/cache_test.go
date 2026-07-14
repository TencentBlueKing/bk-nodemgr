/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topocache

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

// distinctCall records a single DistinctHost invocation received by the fake backend.
type distinctCall struct {
	tenantID  string
	bkUser    string
	roles     []types.NodeRole
}

// fakeBackendHandler implements backend.IHandler by embedding the interface
// (nil for unused methods) and overriding only DistinctHost.
type fakeBackendHandler struct {
	backend.IHandler

	mu          sync.Mutex
	calls       []distinctCall
	agentResult *types.HostDistinctResult
	proxyResult *types.HostDistinctResult

	// per-tenant error overrides; key "" applies to any tenant without a specific entry.
	agentErrByTenant map[string]error
	proxyErrByTenant map[string]error
	// default error used when no per-tenant override is set.
	agentErr error
	proxyErr error
}

func (h *fakeBackendHandler) DistinctHost(
	nCtx contextx.IContext, cond *types.HostCondition) (*types.HostDistinctResult, error) {

	roles := []types.NodeRole(nil)
	if cond != nil && cond.DynamicExactInclude != nil {
		roles = cond.DynamicExactInclude.NodeRole
	}

	h.mu.Lock()
	h.calls = append(h.calls, distinctCall{
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

func (h *fakeBackendHandler) recordedCalls() []distinctCall {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]distinctCall, len(h.calls))
	copy(out, h.calls)
	return out
}

func (h *fakeBackendHandler) callCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.calls)
}

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

// newTestBaseCtx returns a tenant-less base context for the cache syncer.
// syncTenant forks it per-tenant via contextx.From(WithTenantID(...)).
func newTestBaseCtx() contextx.IContext {
	return contextx.Background()
}

func newTestCache(t *testing.T, bh backend.IHandler, td tenant.IHandler) *Cache {
	t.Helper()
	return newCache(bh, td)
}

func sampleResult(version string) *types.HostDistinctResult {
	return &types.HostDistinctResult{
		NodeVersion: []string{version},
		OSType:      []string{"linux"},
	}
}

// ---------------------------------------------------------------------------
// Get / set
// ---------------------------------------------------------------------------

func TestGet_CacheMiss_ReturnsNil(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})

	assert.Nil(t, c.Get("t1", RoleTypeAgent))
	assert.Nil(t, c.Get("t1", RoleTypeProxy))
}

func TestGet_UnknownRoleType_ReturnsNil(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})
	c.set("t1", sampleResult("a"), sampleResult("p"))

	assert.Nil(t, c.Get("t1", "other"))
	assert.Nil(t, c.Get("t1", ""))
}

func TestGet_AfterSet_ReturnsAgentAndProxy(t *testing.T) {
	c := newTestCache(t, &fakeBackendHandler{}, &fakeTenantDao{})
	agentR := sampleResult("agent-v1")
	proxyR := sampleResult("proxy-v1")
	c.set("t1", agentR, proxyR)

	assert.Same(t, agentR, c.Get("t1", RoleTypeAgent))
	assert.Same(t, proxyR, c.Get("t1", RoleTypeProxy))
	// other tenant still absent
	assert.Nil(t, c.Get("t2", RoleTypeAgent))
}

// ---------------------------------------------------------------------------
// buildRoleCondition
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// sync
// ---------------------------------------------------------------------------

func TestSync_PopulatesCacheForAllTenants(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleResult("agent-v1"),
		proxyResult: sampleResult("proxy-v1"),
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	assert.Equal(t, "agent-v1", c.Get("tA", RoleTypeAgent).NodeVersion[0])
	assert.Equal(t, "proxy-v1", c.Get("tA", RoleTypeProxy).NodeVersion[0])
	assert.Equal(t, "agent-v1", c.Get("tB", RoleTypeAgent).NodeVersion[0])
	assert.Equal(t, "proxy-v1", c.Get("tB", RoleTypeProxy).NodeVersion[0])

	// 2 tenants * 2 role calls = 4 backend calls.
	assert.Equal(t, 4, bh.callCount())
}

func TestSync_TenantDaoError_NoMutation(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleResult("agent-v1"),
		proxyResult: sampleResult("proxy-v1"),
	}
	td := &fakeTenantDao{err: errors.New("db down")}
	c := newTestCache(t, bh, td)

	assert.NotPanics(t, func() { c.sync(newTestBaseCtx()) })

	// no entries populated, no backend calls made.
	assert.Nil(t, c.Get("any", RoleTypeAgent))
	assert.Equal(t, 0, bh.callCount())
}

func TestSync_BackendAgentError_SkipsTenant(t *testing.T) {
	// tenant tA: agent call fails. Expect tA fully skipped (set only when both succeed).
	bh := &fakeBackendHandler{
		agentResult:      sampleResult("agent-v1"),
		proxyResult:      sampleResult("proxy-v1"),
		agentErrByTenant: map[string]error{"tA": errors.New("agent backend failed")},
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
	assert.Nil(t, c.Get("tA", RoleTypeAgent))
	assert.Nil(t, c.Get("tA", RoleTypeProxy))
	// tB must be present.
	assert.NotNil(t, c.Get("tB", RoleTypeAgent))
	assert.NotNil(t, c.Get("tB", RoleTypeProxy))
}

func TestSync_BackendProxyError_SkipsTenant(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult:      sampleResult("agent-v1"),
		proxyResult:      sampleResult("proxy-v1"),
		proxyErrByTenant: map[string]error{"tA": errors.New("proxy backend failed")},
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
	assert.Nil(t, c.Get("tA", RoleTypeAgent))
	assert.Nil(t, c.Get("tA", RoleTypeProxy))
	assert.NotNil(t, c.Get("tB", RoleTypeAgent))
	assert.NotNil(t, c.Get("tB", RoleTypeProxy))
}

func TestSync_EmptyTenantList_NoCalls(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleResult("agent-v1"),
		proxyResult: sampleResult("proxy-v1"),
	}
	td := &fakeTenantDao{tenants: nil}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	assert.Equal(t, 0, bh.callCount())
	assert.Nil(t, c.Get("any", RoleTypeAgent))
}

func TestSync_BackendReceivesTenantScopedContext(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleResult("agent-v1"),
		proxyResult: sampleResult("proxy-v1"),
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{
			{ID: "tA", Name: "A", Enabled: true},
			{ID: "tB", Name: "B", Enabled: true},
		},
	}
	c := newTestCache(t, bh, td)

	c.sync(newTestBaseCtx())

	// verify backend received the correct tenant-id for every call.
	calls := bh.recordedCalls()
	require.Len(t, calls, 4)

	// the syncer runs as the system virtual user (no request-scoped user).
	expectedUser := access.GetVirtualUser()

	tenantSeen := map[string]bool{}
	for _, call := range calls {
		// each call must carry a non-empty tenant-id that matches a known tenant.
		assert.True(t, call.tenantID == "tA" || call.tenantID == "tB",
			"unexpected tenant-id: %q", call.tenantID)
		tenantSeen[call.tenantID] = true

		// backend apigw auth rejects empty username; verify the virtual user was propagated.
		assert.Equal(t, expectedUser, call.bkUser,
			"backend call did not carry the system virtual user as bk_username")

		// verify the role-scope classification.
		if len(call.roles) == 1 && call.roles[0] == types.NodeRoleProxy {
			assert.Equal(t, []types.NodeRole{types.NodeRoleProxy}, call.roles)
		} else {
			assert.Equal(t,
				[]types.NodeRole{types.NodeRoleAgent, types.NodeRoleBlank}, call.roles)
		}
	}
	assert.True(t, tenantSeen["tA"] && tenantSeen["tB"], "both tenants should have been synced")
}

// ---------------------------------------------------------------------------
// Start / run lifecycle
// ---------------------------------------------------------------------------

func TestStart_NonBlocking_RunsInitialSync(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleResult("agent-v1"),
		proxyResult: sampleResult("proxy-v1"),
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

	// the background syncer should populate the cache shortly.
	require.Eventually(t, func() bool {
		return c.Get("tA", RoleTypeAgent) != nil
	}, 2*time.Second, 10*time.Millisecond, "initial sync did not populate cache in time")

	cancel()
}

func TestRun_StopsOnBaseCtxCancel(t *testing.T) {
	bh := &fakeBackendHandler{
		agentResult: sampleResult("agent-v1"),
		proxyResult: sampleResult("proxy-v1"),
	}
	td := &fakeTenantDao{
		tenants: []*types.Tenant{{ID: "tA", Name: "A", Enabled: true}},
	}
	ctx, cancel := contextx.WithCancel(contextx.Background())
	c := newCache(bh, td)

	go c.run(ctx)

	// give the syncer a moment to finish the initial sync.
	require.Eventually(t, func() bool {
		return c.Get("tA", RoleTypeAgent) != nil
	}, 2*time.Second, 10*time.Millisecond)

	cancel()

	// after cancel, the goroutine should stop; wait a bit and assert the call
	// count stops growing (no more syncs fired from the ticker).
	countAtCancel := bh.callCount()
	time.Sleep(150 * time.Millisecond)
	countAfterWait := bh.callCount()
	assert.Equal(t, countAtCancel, countAfterWait,
		"backend call count kept growing after baseCtx was cancelled")
}
