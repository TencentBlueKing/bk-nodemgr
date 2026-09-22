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

package v4

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v4/provider"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeIAMV4Handler struct {
	listAuthorizedResourcesCalls int
	listAuthorizedResourcesReq   types.IAMAuthorizedInstancesRequest
	listAuthorizedResourcesResp  []iamv4.AuthorizedResourceResponse

	resourcesAllowedCalls int
	resourcesAllowedReq   types.IAMCheckRequest
	allowedByResource     map[string]bool
}

func (h *fakeIAMV4Handler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	return false, nil
}

func (h *fakeIAMV4Handler) ResourcesAllowed(
	_ contextx.IContext, req types.IAMCheckRequest,
) (map[string]bool, error) {

	h.resourcesAllowedCalls++
	h.resourcesAllowedReq = req
	return h.allowedByResource, nil
}

func (h *fakeIAMV4Handler) ActionsAllowed(
	_ contextx.IContext, _ types.IAMMultiActionCheckRequest,
) (map[string]bool, error) {

	return nil, nil
}

func (h *fakeIAMV4Handler) ListAuthorizedResources(
	_ contextx.IContext, req types.IAMAuthorizedInstancesRequest,
) ([]iamv4.AuthorizedResourceResponse, error) {

	h.listAuthorizedResourcesCalls++
	h.listAuthorizedResourcesReq = req
	return h.listAuthorizedResourcesResp, nil
}

func (h *fakeIAMV4Handler) GetApplyURL(_ contextx.IContext, _ types.IAMApplyRequest) (string, error) {
	return "", nil
}

func (h *fakeIAMV4Handler) GetToken(_ contextx.IContext) (string, error) {
	return "", nil
}

func (h *fakeIAMV4Handler) IsBasicAuthAllowed(_ contextx.IContext, _, _ string) error {
	return nil
}

func (h *fakeIAMV4Handler) GrantRole(_ contextx.IContext, _ types.IAMRoleGrantRequest) error {
	return nil
}

type fakeIAMV4InstanceLister struct {
	ids          map[types.AuthResourceType][]string
	calls        int
	resourceType string
	parent       *provider.ParentFilter
}

func (l *fakeIAMV4InstanceLister) ListInstance(
	_ contextx.IContext,
	resourceType string,
	req *provider.Request[provider.ListInstanceFilter],
) (*provider.ListInstanceData, error) {

	l.calls++
	l.resourceType = resourceType
	l.parent = req.Filter.Parent

	ids := l.ids[types.AuthResourceType(resourceType)]
	results := make([]provider.ResourceInstance, 0, len(ids))
	for _, id := range ids {
		results = append(results, provider.ResourceInstance{ID: id, DisplayName: id})
	}

	return &provider.ListInstanceData{Count: int64(len(results)), Results: results}, nil
}

func newIAMV4ListAuthorizedInstancesTestContext() contextx.IContext {
	return contextx.New(context.Background(), contextx.WithBKUsername("tester"))
}

func TestIAMV4AuthorizerListAuthorizedInstances_UsesResourceChecksForLowerLevelNetworkUnit(t *testing.T) {
	handler := &fakeIAMV4Handler{allowedByResource: map[string]bool{
		"unit-1": true,
		"unit-2": false,
		"unit-3": true,
	}}
	lister := &fakeIAMV4InstanceLister{ids: map[types.AuthResourceType][]string{
		types.AuthResourceTypeNetworkUnit: {"unit-1", "unit-2", "unit-3"},
	}}
	authorizer := NewIAMV4Authorizer(types.SystemIDNodeMgr, handler, nil, lister, nil)

	scope, err := authorizer.ListAuthorizedInstances(
		newIAMV4ListAuthorizedInstancesTestContext(),
		auth.ActionNetworkUnitView,
		types.AuthResourceTypeNetworkUnit,
	)
	require.NoError(t, err)

	assert.False(t, scope.IsAny)
	assert.Equal(t, []types.AuthResource{
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkUnit, ID: "unit-1"},
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkUnit, ID: "unit-3"},
	}, scope.Resources)
	assert.Equal(t, 0, handler.listAuthorizedResourcesCalls)
	assert.Equal(t, 1, handler.resourcesAllowedCalls)
	assert.Equal(t, string(auth.ActionNetworkUnitView), handler.resourcesAllowedReq.ActionID)
	assert.Equal(t, []types.IAMResource{
		{SystemID: types.SystemIDNodeMgr, Type: string(types.AuthResourceTypeNetworkUnit), ID: "unit-1"},
		{SystemID: types.SystemIDNodeMgr, Type: string(types.AuthResourceTypeNetworkUnit), ID: "unit-2"},
		{SystemID: types.SystemIDNodeMgr, Type: string(types.AuthResourceTypeNetworkUnit), ID: "unit-3"},
	}, handler.resourcesAllowedReq.Resources)
	assert.Equal(t, 1, lister.calls)
	assert.Equal(t, string(types.AuthResourceTypeNetworkUnit), lister.resourceType)
	assert.Nil(t, lister.parent)
}

func TestIAMV4AuthorizerListAuthorizedInstances_UsesResourceChecksForLowerLevelPackage(t *testing.T) {
	handler := &fakeIAMV4Handler{allowedByResource: map[string]bool{
		"gse_agent":     true,
		"bkmonitorbeat": false,
	}}
	lister := &fakeIAMV4InstanceLister{ids: map[types.AuthResourceType][]string{
		types.AuthResourceTypePackage: {"gse_agent", "bkmonitorbeat"},
	}}
	authorizer := NewIAMV4Authorizer(types.SystemIDNodeMgr, handler, nil, lister, nil)

	scope, err := authorizer.ListAuthorizedInstances(
		newIAMV4ListAuthorizedInstancesTestContext(), auth.ActionPackageView, types.AuthResourceTypePackage,
	)
	require.NoError(t, err)

	assert.False(t, scope.IsAny)
	assert.Equal(t, []types.AuthResource{{
		SystemID: types.SystemIDNodeMgr,
		Type:     types.AuthResourceTypePackage,
		ID:       "gse_agent",
	}}, scope.Resources)
	assert.Equal(t, 0, handler.listAuthorizedResourcesCalls)
	assert.Equal(t, 1, handler.resourcesAllowedCalls)
	assert.Equal(t, string(auth.ActionPackageView), handler.resourcesAllowedReq.ActionID)
	assert.Equal(t, []types.IAMResource{
		{SystemID: types.SystemIDNodeMgr, Type: string(types.AuthResourceTypePackage), ID: "gse_agent"},
		{SystemID: types.SystemIDNodeMgr, Type: string(types.AuthResourceTypePackage), ID: "bkmonitorbeat"},
	}, handler.resourcesAllowedReq.Resources)
	assert.Equal(t, 1, lister.calls)
	assert.Equal(t, string(types.AuthResourceTypePackage), lister.resourceType)
	assert.Nil(t, lister.parent)
}

func TestIAMV4AuthorizerListAuthorizedInstances_UsesAuthorizedResourcesForTopLevelNetworkArea(t *testing.T) {
	handler := &fakeIAMV4Handler{listAuthorizedResourcesResp: []iamv4.AuthorizedResourceResponse{{
		Type: string(types.AuthResourceTypeNetworkArea),
		IDs:  []string{"area-1", "area-2"},
	}}}
	lister := &fakeIAMV4InstanceLister{}
	authorizer := NewIAMV4Authorizer(types.SystemIDNodeMgr, handler, nil, lister, nil)

	scope, err := authorizer.ListAuthorizedInstances(
		newIAMV4ListAuthorizedInstancesTestContext(), auth.ActionNetworkAreaView, types.AuthResourceTypeNetworkArea,
	)
	require.NoError(t, err)

	assert.False(t, scope.IsAny)
	assert.Equal(t, []types.AuthResource{
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkArea, ID: "area-1"},
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkArea, ID: "area-2"},
	}, scope.Resources)
	assert.Equal(t, 1, handler.listAuthorizedResourcesCalls)
	assert.Equal(t, 0, handler.resourcesAllowedCalls)
	assert.Equal(t, 0, lister.calls)
	assert.Equal(t, types.IAMAuthorizedInstancesRequest{
		SystemID:     types.SystemIDNodeMgr,
		Username:     "tester",
		ActionID:     string(auth.ActionNetworkAreaView),
		ResourceType: string(types.AuthResourceTypeNetworkArea),
	}, handler.listAuthorizedResourcesReq)
}
