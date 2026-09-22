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

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	pkgTenant "github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	pkgTenant.SetMode(pkgTenant.ModeMultiple)
	os.Exit(m.Run())
}

type fakeTenantStorage struct {
	tenant *types.Tenant
	err    error
}

func (f *fakeTenantStorage) EnsureReservedTenant(_ contextx.IContext, tenant *types.Tenant) error {
	f.tenant = tenant
	return f.err
}

type fakeUserManagerHandler struct {
	tenants []*types.Tenant
	err     error
}

func (f *fakeUserManagerHandler) ListALLTenants(contextx.IContext) ([]*types.Tenant, error) {
	return f.tenants, f.err
}

type fakeSyncManager struct {
	managerIface.ISyncManager
	calls []string
	ctxs  []contextx.IContext
	errs  map[string]error
}

func (f *fakeSyncManager) LaunchSyncBizAndHost(ctx contextx.IContext) (string, error) {
	return f.launch(ctx, initWorkflowSyncBizAndHost)
}

func (f *fakeSyncManager) LaunchSyncNetworkArea(ctx contextx.IContext) (string, error) {
	return f.launch(ctx, initWorkflowSyncNetworkArea)
}

func (f *fakeSyncManager) LaunchEnsureDefaultPlugin(ctx contextx.IContext) (string, error) {
	return f.launch(ctx, initWorkflowEnsureDefaultPlugin)
}

func (f *fakeSyncManager) LaunchSyncSharedReleases(ctx contextx.IContext) (string, error) {
	return f.launch(ctx, initWorkflowSyncSharedReleases)
}

func (f *fakeSyncManager) launch(ctx contextx.IContext, workflow string) (string, error) {
	f.calls = append(f.calls, workflow)
	f.ctxs = append(f.ctxs, ctx)
	if f.errs != nil && f.errs[workflow] != nil {
		return "", f.errs[workflow]
	}

	return "trigger-id", nil
}

type fakeRestContext struct {
	contextx.IContext
	body string
	data restserver.IRequestData
}

func newFakeRestContext(body string) *fakeRestContext {
	return &fakeRestContext{
		IContext: contextx.New(
			context.Background(),
			contextx.WithTenantID(pkgTenant.SystemTenantID),
			contextx.WithLoginName("admin"),
			contextx.WithBKUsername("admin"),
		),
		body: body,
		data: newFakeRequestData(),
	}
}

func (f *fakeRestContext) BindJSON(body restserver.RequestBody) error {
	if err := json.Unmarshal([]byte(f.body), body); err != nil {
		return err
	}

	body.AutoConvert()
	return body.Validate()
}

func (f *fakeRestContext) GContext() *gin.Context { return nil }

func (f *fakeRestContext) ParseFileForm(restserver.RequestBody) (*multipart.FileHeader, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeRestContext) GetRequestHeader(string) string { return "" }

func (f *fakeRestContext) GetRequest() *http.Request { return nil }

func (f *fakeRestContext) GetCookie(string) (string, error) { return "", errors.New("not implemented") }

func (f *fakeRestContext) AbortWithJSONError(resterrf.Code, []error) {}

func (f *fakeRestContext) AbortWithJSONPermDenied(resterrf.Code, []error) {}

func (f *fakeRestContext) APIResponse(interface{}) {}

func (f *fakeRestContext) Data() restserver.IRequestData { return f.data }

type fakeRequestData struct {
	loginName  string
	bkUsername string
	tenantID   string
	requestID  string
}

func newFakeRequestData() *fakeRequestData {
	return &fakeRequestData{loginName: "admin", bkUsername: "admin", tenantID: pkgTenant.SystemTenantID}
}

func (f *fakeRequestData) GetLoginName() string { return f.loginName }

func (f *fakeRequestData) SetLoginName(loginName string) { f.loginName = loginName }

func (f *fakeRequestData) GetBKUsername() string { return f.bkUsername }

func (f *fakeRequestData) SetBKUsername(bkUsername string) { f.bkUsername = bkUsername }

func (f *fakeRequestData) GetTenantID() string { return f.tenantID }

func (f *fakeRequestData) SetTenantID(tenantID string) { f.tenantID = tenantID }

func (f *fakeRequestData) GetRequestID() string { return f.requestID }

func (f *fakeRequestData) SetRequestID(requestID string) { f.requestID = requestID }

func TestInitEnsuresTenantAndLaunchesInitWorkflows(t *testing.T) {
	storage := &fakeTenantStorage{}
	syncManager := &fakeSyncManager{}
	h := &handler{
		syncManager:        syncManager,
		tenantStorage:      storage,
		userManagerHandler: &fakeUserManagerHandler{tenants: []*types.Tenant{{ID: "tenant-a", Name: "Tenant A", Enabled: true}}},
	}

	result, err := h.Init(newFakeRestContext(`{"tenant_id":"tenant-a"}`))

	require.NoError(t, err)
	resp, ok := result.(*initResp)
	require.True(t, ok)
	assert.Equal(t, []string{
		initWorkflowSyncBizAndHost,
		initWorkflowSyncNetworkArea,
		initWorkflowEnsureDefaultPlugin,
		initWorkflowSyncSharedReleases,
	}, resp.TriggeredWorkflows)
	require.NotNil(t, storage.tenant)
	assert.Equal(t, "tenant-a", storage.tenant.ID)
	assert.Equal(t, "Tenant A", storage.tenant.Name)
	assert.True(t, storage.tenant.Enabled)
	assert.Equal(t, resp.TriggeredWorkflows, syncManager.calls)
	for _, syncCtx := range syncManager.ctxs {
		assert.Equal(t, "tenant-a", syncCtx.TenantID())
		assert.Equal(t, "admin", syncCtx.BKUsername())
	}
}

func TestInitRejectsSystemTenant(t *testing.T) {
	h := &handler{}

	_, err := h.Init(newFakeRestContext(`{"tenant_id":"system"}`))

	require.Error(t, err)
	code, _ := resterrf.ErrUnwrap(err)
	assert.Equal(t, resterrf.InvalidParameter, code)
}

func TestInitRejectsMissingTenant(t *testing.T) {
	h := &handler{userManagerHandler: &fakeUserManagerHandler{}}

	_, err := h.Init(newFakeRestContext(`{"tenant_id":"tenant-a"}`))

	require.Error(t, err)
	code, _ := resterrf.ErrUnwrap(err)
	assert.Equal(t, resterrf.InvalidParameter, code)
}

func TestInitRejectsDisabledTenant(t *testing.T) {
	h := &handler{userManagerHandler: &fakeUserManagerHandler{tenants: []*types.Tenant{{ID: "tenant-a"}}}}

	_, err := h.Init(newFakeRestContext(`{"tenant_id":"tenant-a"}`))

	require.Error(t, err)
	code, _ := resterrf.ErrUnwrap(err)
	assert.Equal(t, resterrf.InvalidParameter, code)
}

func TestInitContinuesLaunchingWorkflowsAndReturnsAggregateError(t *testing.T) {
	syncManager := &fakeSyncManager{errs: map[string]error{initWorkflowSyncNetworkArea: errors.New("boom")}}
	h := &handler{
		syncManager:        syncManager,
		tenantStorage:      &fakeTenantStorage{},
		userManagerHandler: &fakeUserManagerHandler{tenants: []*types.Tenant{{ID: "tenant-a", Name: "Tenant A", Enabled: true}}},
	}

	_, err := h.Init(newFakeRestContext(`{"tenant_id":"tenant-a"}`))

	require.Error(t, err)
	code, _ := resterrf.ErrUnwrap(err)
	assert.Equal(t, resterrf.BackendOperateFailed, code)
	assert.Equal(t, []string{
		initWorkflowSyncBizAndHost,
		initWorkflowSyncNetworkArea,
		initWorkflowEnsureDefaultPlugin,
		initWorkflowSyncSharedReleases,
	}, syncManager.calls)
}
