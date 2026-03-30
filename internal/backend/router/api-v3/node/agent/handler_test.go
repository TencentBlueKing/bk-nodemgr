/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var _ auth.IAuthorizer = (*fakeAuthorizer)(nil)
var _ auth.IAuthorizer = (*fakeAuthorizerSeq)(nil)
var _ topoStg.IStorageHost = (*fakeStorageHost)(nil)
var _ topoStg.IStorageNetworkUnit = (*fakeStorageNetworkUnit)(nil)

type fakeAuthorizer struct {
	batchCheckErr        error
	batchCheckCalls      int
	lastBatchCheckAction auth.Action
	lastBatchResources   []auth.Resource
}

func (f *fakeAuthorizer) Check(_ contextx.IContext, action auth.Action, resources []auth.Resource) error {
	f.batchCheckCalls++
	f.lastBatchCheckAction = action
	f.lastBatchResources = append([]auth.Resource(nil), resources...)

	return f.batchCheckErr
}

func (f *fakeAuthorizer) CheckMany(_ contextx.IContext, _ map[auth.Action][]auth.Resource) error {
	return nil
}

type fakeAuthorizerSeq struct {
	errs                  []error
	batchCheckCalls       int
	lastBatchCheckAction  auth.Action
	lastBatchResources    []auth.Resource
	batchCheckActions     []auth.Action
	batchCheckResourceses [][]auth.Resource
}

func (f *fakeAuthorizerSeq) Check(_ contextx.IContext, action auth.Action, resources []auth.Resource) error {
	f.batchCheckCalls++
	f.lastBatchCheckAction = action
	f.lastBatchResources = append([]auth.Resource(nil), resources...)
	f.batchCheckActions = append(f.batchCheckActions, action)
	f.batchCheckResourceses = append(f.batchCheckResourceses, append([]auth.Resource(nil), resources...))

	idx := f.batchCheckCalls - 1
	if idx < len(f.errs) {
		return f.errs[idx]
	}

	return nil
}

func (f *fakeAuthorizerSeq) CheckMany(_ contextx.IContext, _ map[auth.Action][]auth.Resource) error {
	return nil
}

type fakeStorageHost struct {
	listHosts []*types.Host
	listErr   error
	listCalls int
}

func (f *fakeStorageHost) UpsertManyHost(contextx.IContext, ...*types.Host) error {
	return nil
}

func (f *fakeStorageHost) UpsertManyHostStatic(contextx.IContext, ...*types.Host) error {
	return nil
}

func (f *fakeStorageHost) UpdateManyHostDynamic(contextx.IContext, ...*types.Host) error {
	return nil
}

func (f *fakeStorageHost) ListHostWithFields(contextx.IContext, types.Page, *types.HostFieldSelection, ...*types.HostCondition) (
	[]*types.Host, int64, error,
) {
	return nil, 0, nil
}

func (f *fakeStorageHost) ListHost(contextx.IContext, types.Page, ...*types.HostCondition) ([]*types.Host, int64, error) {
	f.listCalls++

	return f.listHosts, int64(len(f.listHosts)), f.listErr
}

func (f *fakeStorageHost) ListHostOrderByUpdateTime(contextx.IContext, types.Page, ...*types.HostCondition) (
	[]*types.Host, int64, error,
) {
	return nil, 0, nil
}

func (f *fakeStorageHost) DeleteManyHost(contextx.IContext, ...int64) error {
	return nil
}

func (f *fakeStorageHost) CountHost(contextx.IContext, ...*types.HostCondition) (int64, error) {
	return 0, nil
}

func (f *fakeStorageHost) GetHostByID(contextx.IContext, int64) (*types.Host, error) {
	return nil, nil
}

func (f *fakeStorageHost) FindHostWithDynamic(contextx.IContext, types.Page, ...*types.HostCondition) ([]*types.Host, error) {
	return nil, nil
}

func (f *fakeStorageHost) UpdateHostDynamicFields(contextx.IContext, types.HostDynamicFields, ...*types.Host) error {
	return nil
}

func (f *fakeStorageHost) TouchHostOperationTime(contextx.IContext, ...int64) error {
	return nil
}

func (f *fakeStorageHost) DistinctHost(contextx.IContext, types.HostDistinctRequest, ...*types.HostCondition) (
	*types.HostDistinctResult, error,
) {
	return nil, nil
}

func (f *fakeStorageHost) GetHostDistributionByNodeRole(contextx.IContext, ...*types.HostCondition) (map[string]int64, error) {
	return nil, nil
}

func (f *fakeStorageHost) GetHostDistributionByNetworkAreaID(contextx.IContext, ...*types.HostCondition) (
	map[int64]int64, error,
) {
	return nil, nil
}

func (f *fakeStorageHost) GetRelayInfosInNetworkUnit(contextx.IContext, int64) ([]*types.RelayInfo, error) {
	return nil, nil
}

type fakeStorageNetworkUnit struct {
	listUnits []*types.NetworkUnit
	listErr   error
	listCalls int
}

func (f *fakeStorageNetworkUnit) ListNetworkUnit(
	contextx.IContext, types.Page, ...*types.NetworkUnitCondition,
) ([]*types.NetworkUnit, int64, error) {
	f.listCalls++

	return f.listUnits, int64(len(f.listUnits)), f.listErr
}

func (f *fakeStorageNetworkUnit) GetNetworkUnit(contextx.IContext, int64) (*types.NetworkUnit, error) {
	return nil, nil
}

func (f *fakeStorageNetworkUnit) CreateNetworkUnit(
	contextx.IContext, *types.NetworkUnit, ...*types.AccessPoint,
) (int64, *topoStg.AccessPointResult, error) {
	return 0, nil, nil
}

func (f *fakeStorageNetworkUnit) UpdateNetworkUnit(
	contextx.IContext, *types.NetworkUnit, ...*types.AccessPoint,
) (*topoStg.AccessPointResult, error) {
	return nil, nil
}

func (f *fakeStorageNetworkUnit) DeleteManyNetworkUnit(contextx.IContext, ...int64) error {
	return nil
}

func TestAgentInstallCheck_PermissionDenied(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authorizer := &fakeAuthorizer{batchCheckErr: errors.New("install-check denied")}
	h := &handler{authorizer: authorizer}
	rCtx := newTestRestContext(t, `{"host":[{"bk_biz_id":2,"bk_host_innerip_list":["127.0.0.1"]}]}`)

	_, err := h.AgentInstallCheck(rCtx)
	assertPermissionDenied(t, err)

	if authorizer.batchCheckCalls != 1 {
		t.Fatalf("expected BatchCheck to be called once, got %d", authorizer.batchCheckCalls)
	}
	if authorizer.lastBatchCheckAction != auth.ActionAgentOperate {
		t.Fatalf("expected action %q, got %q", auth.ActionAgentOperate, authorizer.lastBatchCheckAction)
	}
	assertBizResources(t, authorizer.lastBatchResources, "2")
	if !errors.Is(err, authorizer.batchCheckErr) {
		t.Fatalf("expected wrapped auth error, got %v", err)
	}
}

func TestAgentReconfig_PermissionDenied(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authorizer := &fakeAuthorizer{batchCheckErr: errors.New("reconfig denied")}
	storageHost := &fakeStorageHost{listHosts: []*types.Host{newTestHost(11, 3, 1001)}}
	h := &handler{authorizer: authorizer, storageHost: storageHost}
	rCtx := newTestRestContext(t, `{"host":[{"bk_host_id":11}]}`)

	_, err := h.AgentReconfig(rCtx)
	assertPermissionDenied(t, err)

	if storageHost.listCalls != 1 {
		t.Fatalf("expected ListHost to be called once, got %d", storageHost.listCalls)
	}
	if authorizer.batchCheckCalls != 1 {
		t.Fatalf("expected BatchCheck to be called once, got %d", authorizer.batchCheckCalls)
	}
	if authorizer.lastBatchCheckAction != auth.ActionAgentOperate {
		t.Fatalf("expected action %q, got %q", auth.ActionAgentOperate, authorizer.lastBatchCheckAction)
	}
	assertBizResources(t, authorizer.lastBatchResources, "3")
	if !errors.Is(err, authorizer.batchCheckErr) {
		t.Fatalf("expected wrapped auth error, got %v", err)
	}
}

func TestAgentRestart_PermissionDenied(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authorizer := &fakeAuthorizer{batchCheckErr: errors.New("restart denied")}
	storageHost := &fakeStorageHost{listHosts: []*types.Host{newTestHost(21, 5, 2001)}}
	h := &handler{authorizer: authorizer, storageHost: storageHost}
	rCtx := newTestRestContext(t, `{"host":[{"bk_host_id":21}]}`)

	_, err := h.AgentRestart(rCtx)
	assertPermissionDenied(t, err)

	if storageHost.listCalls != 1 {
		t.Fatalf("expected ListHost to be called once, got %d", storageHost.listCalls)
	}
	if authorizer.batchCheckCalls != 1 {
		t.Fatalf("expected BatchCheck to be called once, got %d", authorizer.batchCheckCalls)
	}
	if authorizer.lastBatchCheckAction != auth.ActionAgentOperate {
		t.Fatalf("expected action %q, got %q", auth.ActionAgentOperate, authorizer.lastBatchCheckAction)
	}
	assertBizResources(t, authorizer.lastBatchResources, "5")
	if !errors.Is(err, authorizer.batchCheckErr) {
		t.Fatalf("expected wrapped auth error, got %v", err)
	}
}

func newTestRestContext(t *testing.T, body string) restserver.IContext {
	t.Helper()

	recorder := httptest.NewRecorder()
	gCtx, _ := gin.CreateTestContext(recorder)
	gCtx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	gCtx.Request.Header.Set("Content-Type", "application/json")

	return &restserver.Context{
		Context: contextx.New(
			context.Background(),
			contextx.WithTenantID("tenant-test"),
			contextx.WithBKUsername("admin"),
			contextx.WithLoginName("admin"),
		),
		IRequest: restserver.NewRequest(gCtx),
	}
}

func newTestHost(hostID, bizID, networkUnitID int64) *types.Host {
	return &types.Host{
		HostID: hostID,
		Static: &types.HostStatic{BizID: bizID},
		Dynamic: &types.HostDynamic{
			NetworkUnitID: networkUnitID,
		},
	}
}

func assertPermissionDenied(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected permission denied error, got nil")
	}

	code, unwrapErrs := resterrf.ErrUnwrap(err)
	if code != resterrf.PermissionDenied {
		t.Fatalf("expected code %d, got %d (err=%v)", resterrf.PermissionDenied, code, err)
	}
	if len(unwrapErrs) == 0 {
		t.Fatalf("expected wrapped detail errors, got none")
	}
}

func assertBizResources(t *testing.T, resources []auth.Resource, wantIDs ...string) {
	t.Helper()

	if len(resources) != len(wantIDs) {
		t.Fatalf("expected %d resources, got %d", len(wantIDs), len(resources))
	}

	for idx, wantID := range wantIDs {
		resource := resources[idx]
		if resource.SystemID != auth.SystemIDCMDB {
			t.Fatalf("expected system id %q, got %q", auth.SystemIDCMDB, resource.SystemID)
		}
		if resource.Type != auth.ResourceTypeBiz {
			t.Fatalf("expected resource type %q, got %q", auth.ResourceTypeBiz, resource.Type)
		}
		if resource.ID != wantID {
			t.Fatalf("expected resource id %q, got %q", wantID, resource.ID)
		}
	}
}

func TestAgentAssignUnit_NetworkUnitPermissionDenied(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authorizer := &fakeAuthorizer{batchCheckErr: errors.New("networkunit denied")}
	storageNetworkUnit := &fakeStorageNetworkUnit{listUnits: []*types.NetworkUnit{{ID: 10, NetworkAreaID: 0}}}
	storageHost := &fakeStorageHost{listHosts: []*types.Host{newTestHost(100, 3, -1)}}
	h := &handler{authorizer: authorizer, storageNetworkUnit: storageNetworkUnit, storageHost: storageHost}
	rCtx := newTestRestContext(t, `{"bk_networkunit_id":10,"bk_host_id":[100]}`)

	_, err := h.AgentAssignUnit(rCtx)
	assertPermissionDenied(t, err)

	if storageNetworkUnit.listCalls != 1 {
		t.Fatalf("expected ListNetworkUnit to be called once, got %d", storageNetworkUnit.listCalls)
	}
	if storageHost.listCalls != 1 {
		t.Fatalf("expected ListHost to be called once, got %d", storageHost.listCalls)
	}
	if authorizer.batchCheckCalls != 1 {
		t.Fatalf("expected BatchCheck to be called once, got %d", authorizer.batchCheckCalls)
	}
	if authorizer.lastBatchCheckAction != auth.ActionNetworkUnitUseForAgent {
		t.Fatalf("expected action %q, got %q", auth.ActionNetworkUnitUseForAgent, authorizer.lastBatchCheckAction)
	}
	if len(authorizer.lastBatchResources) != 1 {
		t.Fatalf("expected 1 network unit resource, got %d", len(authorizer.lastBatchResources))
	}
	resource := authorizer.lastBatchResources[0]
	if resource.SystemID != auth.SystemIDNodeMgr {
		t.Fatalf("expected system id %q, got %q", auth.SystemIDNodeMgr, resource.SystemID)
	}
	if resource.Type != auth.ResourceTypeNetworkUnit {
		t.Fatalf("expected resource type %q, got %q", auth.ResourceTypeNetworkUnit, resource.Type)
	}
	if resource.ID != "10" {
		t.Fatalf("expected resource id %q, got %q", "10", resource.ID)
	}
	if !errors.Is(err, authorizer.batchCheckErr) {
		t.Fatalf("expected wrapped auth error, got %v", err)
	}
}

func TestAgentAssignUnit_AgentOperatePermissionDenied(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authorizer := &fakeAuthorizerSeq{errs: []error{nil, errors.New("agent operate denied")}}
	storageNetworkUnit := &fakeStorageNetworkUnit{listUnits: []*types.NetworkUnit{{ID: 10, NetworkAreaID: 0}}}
	storageHost := &fakeStorageHost{listHosts: []*types.Host{newTestHost(100, 3, -1)}}
	h := &handler{authorizer: authorizer, storageNetworkUnit: storageNetworkUnit, storageHost: storageHost}
	rCtx := newTestRestContext(t, `{"bk_networkunit_id":10,"bk_host_id":[100]}`)

	_, err := h.AgentAssignUnit(rCtx)
	assertPermissionDenied(t, err)

	if authorizer.batchCheckCalls != 2 {
		t.Fatalf("expected BatchCheck to be called twice, got %d", authorizer.batchCheckCalls)
	}
	if len(authorizer.batchCheckActions) != 2 {
		t.Fatalf("expected 2 recorded actions, got %d", len(authorizer.batchCheckActions))
	}
	if authorizer.batchCheckActions[0] != auth.ActionNetworkUnitUseForAgent {
		t.Fatalf("expected first action %q, got %q", auth.ActionNetworkUnitUseForAgent, authorizer.batchCheckActions[0])
	}
	if authorizer.batchCheckActions[1] != auth.ActionAgentOperate {
		t.Fatalf("expected second action %q, got %q", auth.ActionAgentOperate, authorizer.batchCheckActions[1])
	}
	if len(authorizer.batchCheckResourceses) != 2 {
		t.Fatalf("expected 2 recorded resource batches, got %d", len(authorizer.batchCheckResourceses))
	}
	firstResources := authorizer.batchCheckResourceses[0]
	if len(firstResources) != 1 || firstResources[0].SystemID != auth.SystemIDNodeMgr ||
		firstResources[0].Type != auth.ResourceTypeNetworkUnit || firstResources[0].ID != "10" {
		t.Fatalf("unexpected first auth resources: %+v", firstResources)
	}
	assertBizResources(t, authorizer.batchCheckResourceses[1], "3")
	if !errors.Is(err, authorizer.errs[1]) {
		t.Fatalf("expected wrapped auth error, got %v", err)
	}
}

func TestAgentAssignUnit_BothPermissionsGranted(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authorizer := &fakeAuthorizerSeq{errs: []error{nil, nil}}
	storageNetworkUnit := &fakeStorageNetworkUnit{listUnits: []*types.NetworkUnit{{ID: 10, NetworkAreaID: 0}}}
	storageHost := &fakeStorageHost{listHosts: []*types.Host{newTestHost(100, 3, -1)}}
	h := &handler{authorizer: authorizer, storageNetworkUnit: storageNetworkUnit, storageHost: storageHost}
	rCtx := newTestRestContext(t, `{"bk_networkunit_id":10,"bk_host_id":[100]}`)

	_, err := h.AgentAssignUnit(rCtx)
	if err != nil {
		code, _ := resterrf.ErrUnwrap(err)
		if code == resterrf.PermissionDenied {
			t.Fatalf("expected non-permission-denied result, got %v", err)
		}
		t.Fatalf("expected nil error after auth passes, got %v", err)
	}

	if authorizer.batchCheckCalls != 2 {
		t.Fatalf("expected BatchCheck to be called twice, got %d", authorizer.batchCheckCalls)
	}
	if len(authorizer.batchCheckActions) != 2 {
		t.Fatalf("expected 2 recorded actions, got %d", len(authorizer.batchCheckActions))
	}
	if authorizer.batchCheckActions[0] != auth.ActionNetworkUnitUseForAgent {
		t.Fatalf("expected first action %q, got %q", auth.ActionNetworkUnitUseForAgent, authorizer.batchCheckActions[0])
	}
	if authorizer.batchCheckActions[1] != auth.ActionAgentOperate {
		t.Fatalf("expected second action %q, got %q", auth.ActionAgentOperate, authorizer.batchCheckActions[1])
	}
	assertBizResources(t, authorizer.batchCheckResourceses[1], "3")
}
