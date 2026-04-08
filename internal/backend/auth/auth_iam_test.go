/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var _ iamv3.IHandler = (*fakeIAMBatchHandler)(nil)

type fakeIAMBatchHandler struct {
	batchResults         map[string]bool
	batchResultsByAction map[string]map[string]bool
	batchErr             error
	applyURL             string
	applyErr             error
	checkResult          bool
	checkErr             error
	authorizedIsAny      bool
	authorizedResources  []types.IAMResource
	authorizedErr        error

	batchCalls              int
	isAllowedWithCacheCalls int
	applyCalls              int
	authorizedCalls         int

	lastBatchReq       types.IAMCheckRequest
	lastBatchResources [][]types.IAMResource
	lastApplyReq       types.IAMApplyRequest
	lastAuthorizedReq  types.IAMAuthorizedInstancesRequest
}

func (h *fakeIAMBatchHandler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	return false, nil
}

func (h *fakeIAMBatchHandler) IsAllowedWithCache(
	_ contextx.IContext, _ types.IAMCheckRequest, _ time.Duration,
) (bool, error) {
	h.isAllowedWithCacheCalls++
	return h.checkResult, h.checkErr
}

func (h *fakeIAMBatchHandler) BatchIsAllowed(
	_ contextx.IContext, req types.IAMCheckRequest, resourcesList [][]types.IAMResource,
) (map[string]bool, error) {
	h.batchCalls++
	h.lastBatchReq = req
	h.lastBatchResources = resourcesList
	if h.batchResultsByAction != nil {
		if results, ok := h.batchResultsByAction[req.ActionID]; ok {
			return results, h.batchErr
		}
	}
	return h.batchResults, h.batchErr
}

func (h *fakeIAMBatchHandler) ResourceMultiActionsAllowed(
	_ contextx.IContext, _ types.IAMMultiActionCheckRequest,
) (map[string]bool, error) {
	return nil, nil
}

func (h *fakeIAMBatchHandler) BatchResourceMultiActionsAllowed(
	_ contextx.IContext, _ types.IAMMultiActionCheckRequest, _ [][]types.IAMResource,
) (map[string]map[string]bool, error) {
	return nil, nil
}

func (h *fakeIAMBatchHandler) GetToken(_ contextx.IContext) (string, error) {
	return "", nil
}

func (h *fakeIAMBatchHandler) IsBasicAuthAllowed(_ contextx.IContext, _, _ string) error {
	return nil
}

func (h *fakeIAMBatchHandler) GetApplyURL(_ contextx.IContext, req types.IAMApplyRequest) (string, error) {
	h.applyCalls++
	h.lastApplyReq = req
	return h.applyURL, h.applyErr
}

func (h *fakeIAMBatchHandler) ListAuthorizedInstances(
	_ contextx.IContext, req types.IAMAuthorizedInstancesRequest,
) (bool, []types.IAMResource, error) {
	h.authorizedCalls++
	h.lastAuthorizedReq = req
	return h.authorizedIsAny, h.authorizedResources, h.authorizedErr
}

func newTestIAMContext() contextx.IContext {
	return contextx.New(
		context.Background(),
		contextx.WithTenantID("tenant-test"),
		contextx.WithBKUsername("admin"),
		contextx.WithLoginName("admin"),
	)
}

func TestIAMV3AuthorizerCheck_AllAllowed(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": true,
			"2": true,
		},
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "2"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if handler.batchCalls != 1 {
		t.Fatalf("expected BatchIsAllowed to be called once, got %d", handler.batchCalls)
	}
	if handler.isAllowedWithCacheCalls != 0 {
		t.Fatalf("expected IsAllowedWithCache not to be called, got %d", handler.isAllowedWithCacheCalls)
	}
	if handler.applyCalls != 0 {
		t.Fatalf("expected GetApplyURL not to be called, got %d", handler.applyCalls)
	}
	if handler.lastBatchReq.Username != "admin" {
		t.Fatalf("expected batch request username to be admin, got %q", handler.lastBatchReq.Username)
	}
	if len(handler.lastBatchResources) != 2 {
		t.Fatalf("expected 2 resource sets, got %d", len(handler.lastBatchResources))
	}
}

func TestIAMV3AuthorizerCheck_PartialDeniedReturnsPermissionDenied(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": true,
			"2": false,
			"3": false,
		},
		applyURL: "https://iam.example.com/apply",
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "2"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "3"},
	})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}

	var permErr PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermissionDeniedError, got %T: %v", err, err)
	}
	if permErr.ApplyURL != "https://iam.example.com/apply" {
		t.Fatalf("expected apply url to be preserved, got %q", permErr.ApplyURL)
	}
	if len(permErr.Actions) != 1 {
		t.Fatalf("expected 1 action entry, got %d", len(permErr.Actions))
	}
	if len(permErr.Actions[0].RelatedResourceTypes) != 1 {
		t.Fatalf("expected denied biz resources to be deduplicated by type, got %d", len(permErr.Actions[0].RelatedResourceTypes))
	}
	rt := permErr.Actions[0].RelatedResourceTypes[0]
	if len(rt.Instances) != 2 {
		t.Fatalf("expected 2 denied instances in permission payload, got %d", len(rt.Instances))
	}
	first := rt.Instances[0]
	if first.Type != string(types.AuthResourceTypeBiz) || first.ID != "2" {
		t.Fatalf("unexpected first denied instance node: %+v", first)
	}
	if first.TypeName != types.AuthResourceTypeDisplayName(types.AuthResourceTypeBiz) {
		t.Fatalf("expected first denied instance type name %q, got %q", types.AuthResourceTypeDisplayName(types.AuthResourceTypeBiz), first.TypeName)
	}
	if handler.applyCalls != 1 {
		t.Fatalf("expected GetApplyURL to be called once, got %d", handler.applyCalls)
	}
	if len(handler.lastApplyReq.Actions) != 1 {
		t.Fatalf("expected 1 apply action, got %d", len(handler.lastApplyReq.Actions))
	}
}

func TestIAMV3AuthorizerCheck_BatchErrorReturnsImmediately(t *testing.T) {
	sentinel := errors.New("batch iam failure")
	handler := &fakeIAMBatchHandler{batchErr: sentinel}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
	if handler.applyCalls != 0 {
		t.Fatalf("expected GetApplyURL not to be called on batch failure, got %d", handler.applyCalls)
	}
}

func TestIAMV3AuthorizerCheck_ApplyURLErrorReturnsPermissionDeniedWithoutURL(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": false,
		},
		applyErr: errors.New("apply url failure"),
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
	})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}

	var permErr PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermissionDeniedError, got %T: %v", err, err)
	}
	if permErr.ApplyURL != "" {
		t.Fatalf("expected empty apply url when GetApplyURL fails, got %q", permErr.ApplyURL)
	}
}

func TestIAMV3AuthorizerNewPermissionDeniedError_MultiActionStableOrder(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		applyURL: "https://iam.example.com/apply",
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	permErr := authorizer.newPermissionDeniedError(newTestIAMContext(), map[Action][]types.AuthResource{
		ActionProxyView: {
			{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "2"},
		},
		ActionAgentOperate: {
			{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkArea, ID: "3"},
		},
	})

	if handler.applyCalls != 1 {
		t.Fatalf("expected GetApplyURL to be called once, got %d", handler.applyCalls)
	}
	if len(handler.lastApplyReq.Actions) != 2 {
		t.Fatalf("expected 2 apply actions, got %d", len(handler.lastApplyReq.Actions))
	}
	if handler.lastApplyReq.Actions[0].ID != string(ActionAgentOperate) {
		t.Fatalf("expected first apply action %q, got %q", ActionAgentOperate, handler.lastApplyReq.Actions[0].ID)
	}
	if handler.lastApplyReq.Actions[1].ID != string(ActionProxyView) {
		t.Fatalf("expected second apply action %q, got %q", ActionProxyView, handler.lastApplyReq.Actions[1].ID)
	}
	if len(permErr.Actions) != 2 {
		t.Fatalf("expected 2 permission actions, got %d", len(permErr.Actions))
	}
	if permErr.Actions[0].ID != string(ActionAgentOperate) {
		t.Fatalf("expected first permission action %q, got %q", ActionAgentOperate, permErr.Actions[0].ID)
	}
	if permErr.Actions[1].ID != string(ActionProxyView) {
		t.Fatalf("expected second permission action %q, got %q", ActionProxyView, permErr.Actions[1].ID)
	}
	if len(permErr.Actions[0].RelatedResourceTypes) != 1 {
		t.Fatalf("expected 1 related resource type for %q, got %d", ActionAgentOperate, len(permErr.Actions[0].RelatedResourceTypes))
	}
	if permErr.Actions[0].RelatedResourceTypes[0].Type != string(types.AuthResourceTypeNetworkArea) {
		t.Fatalf("expected %q resource type for %q, got %q", types.AuthResourceTypeNetworkArea, ActionAgentOperate, permErr.Actions[0].RelatedResourceTypes[0].Type)
	}
	if len(permErr.Actions[1].RelatedResourceTypes) != 1 {
		t.Fatalf("expected 1 related resource type for %q, got %d", ActionProxyView, len(permErr.Actions[1].RelatedResourceTypes))
	}
	if permErr.Actions[1].RelatedResourceTypes[0].Type != string(types.AuthResourceTypeBiz) {
		t.Fatalf("expected %q resource type for %q, got %q", types.AuthResourceTypeBiz, ActionProxyView, permErr.Actions[1].RelatedResourceTypes[0].Type)
	}
}

func TestIAMV3AuthorizerCheckMany_AggregatesDeniedActions(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResultsByAction: map[string]map[string]bool{
			string(ActionAgentView): {
				"1": true,
				"2": false,
			},
			string(ActionProxyView): {
				"1": false,
				"2": false,
			},
		},
		applyURL: "https://iam.example.com/apply",
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}
	bizResources := []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "2"},
	}

	err := authorizer.CheckMany(newTestIAMContext(), map[Action][]types.AuthResource{
		ActionProxyView: bizResources,
		ActionAgentView: bizResources,
	})
	if err == nil {
		t.Fatal("expected aggregated permission error, got nil")
	}

	var permErr PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermissionDeniedError, got %T: %v", err, err)
	}
	if handler.batchCalls != 2 {
		t.Fatalf("expected BatchIsAllowed to be called twice, got %d", handler.batchCalls)
	}
	if handler.applyCalls != 1 {
		t.Fatalf("expected GetApplyURL to be called once, got %d", handler.applyCalls)
	}
	if len(permErr.Actions) != 2 {
		t.Fatalf("expected 2 denied actions, got %d", len(permErr.Actions))
	}
	if permErr.Actions[0].ID != string(ActionAgentView) {
		t.Fatalf("expected first denied action %q, got %q", ActionAgentView, permErr.Actions[0].ID)
	}
	if permErr.Actions[1].ID != string(ActionProxyView) {
		t.Fatalf("expected second denied action %q, got %q", ActionProxyView, permErr.Actions[1].ID)
	}
}

func TestIAMV3AuthorizerCheck_EmptyResourcesFallsBackToActionCheck(t *testing.T) {
	handler := &fakeIAMBatchHandler{checkResult: true}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if handler.batchCalls != 0 {
		t.Fatalf("expected no batch IAM calls for empty resources, got %d", handler.batchCalls)
	}
	if handler.isAllowedWithCacheCalls != 1 {
		t.Fatalf("expected action-level check to be called once, got %d", handler.isAllowedWithCacheCalls)
	}
}

func TestIAMV3AuthorizerCheck_NonEmptyResourcesUsesBatchEvaluation(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"42": true,
		},
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "42"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if handler.batchCalls != 1 {
		t.Fatalf("expected batch IAM evaluation to be called once, got %d", handler.batchCalls)
	}
	if handler.isAllowedWithCacheCalls != 0 {
		t.Fatalf("expected action-level cached check not to be called, got %d", handler.isAllowedWithCacheCalls)
	}
}

func TestIAMV3AuthorizerListAuthorizedInstances_MapsRequestAndResponse(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		authorizedResources: []types.IAMResource{
			{SystemID: types.SystemIDCMDB, Type: string(types.AuthResourceTypeBiz), ID: "1"},
			{SystemID: types.SystemIDCMDB, Type: string(types.AuthResourceTypeBiz), ID: "2"},
		},
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	scope, err := authorizer.ListAuthorizedInstances(newTestIAMContext(), ActionAgentView, types.AuthResourceTypeBiz)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if scope.IsAny {
		t.Fatal("expected IsAny=false, got true")
	}
	if len(scope.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(scope.Resources))
	}
	if scope.Resources[0].ID != "1" || scope.Resources[1].ID != "2" {
		t.Fatalf("unexpected authorized resource IDs: %+v", scope.Resources)
	}
	if scope.Resources[0].SystemID != types.SystemIDCMDB || scope.Resources[0].Type != types.AuthResourceTypeBiz {
		t.Fatalf("unexpected resource[0] systemID/type: %+v", scope.Resources[0])
	}
	if handler.authorizedCalls != 1 {
		t.Fatalf("expected ListAuthorizedInstances called once, got %d", handler.authorizedCalls)
	}
	if handler.lastAuthorizedReq.SystemID != types.SystemIDNodeMgr {
		t.Fatalf("expected systemID %q, got %q", types.SystemIDNodeMgr, handler.lastAuthorizedReq.SystemID)
	}
	if handler.lastAuthorizedReq.Username != "admin" {
		t.Fatalf("expected username admin, got %q", handler.lastAuthorizedReq.Username)
	}
	if handler.lastAuthorizedReq.ActionID != string(ActionAgentView) {
		t.Fatalf("expected actionID %q, got %q", ActionAgentView, handler.lastAuthorizedReq.ActionID)
	}
	if handler.lastAuthorizedReq.ResourceType != string(types.AuthResourceTypeBiz) {
		t.Fatalf("expected resourceType %q, got %q", types.AuthResourceTypeBiz, handler.lastAuthorizedReq.ResourceType)
	}
}

func TestIAMV3AuthorizerListAuthorizedInstances_ReturnsAnyScope(t *testing.T) {
	handler := &fakeIAMBatchHandler{authorizedIsAny: true}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	scope, err := authorizer.ListAuthorizedInstances(newTestIAMContext(), ActionAgentView, types.AuthResourceTypeBiz)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !scope.IsAny {
		t.Fatal("expected IsAny=true, got false")
	}
	if len(scope.Resources) != 0 {
		t.Fatalf("expected empty Resources for any scope, got %+v", scope.Resources)
	}
}

func TestIAMV3AuthorizerListAuthorizedInstances_PropagatesError(t *testing.T) {
	sentinel := errors.New("list authorized instances failure")
	handler := &fakeIAMBatchHandler{authorizedErr: sentinel}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	_, err := authorizer.ListAuthorizedInstances(newTestIAMContext(), ActionAgentView, types.AuthResourceTypeBiz)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}

// TestBuildIAMApplyResourceTypes_SingleResource verifies a single denied resource
// produces one RelatedResourceType with one single-node instance.
func TestBuildIAMApplyResourceTypes_SingleResource(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "42"},
	})
	if len(rts) != 1 {
		t.Fatalf("expected 1 resource type, got %d", len(rts))
	}
	rt := rts[0]
	if rt.SystemID != types.SystemIDCMDB || rt.Type != string(types.AuthResourceTypeBiz) {
		t.Fatalf("unexpected resource type %+v", rt)
	}
	if len(rt.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(rt.Instances))
	}
	if len(rt.Instances[0]) != 1 {
		t.Fatalf("expected instance path length 1, got %d", len(rt.Instances[0]))
	}
	node := rt.Instances[0][0]
	if node.Type != string(types.AuthResourceTypeBiz) || node.ID != "42" {
		t.Fatalf("unexpected instance node %+v", node)
	}
}

// TestBuildIAMApplyResourceTypes_MultiResourceSameType verifies that multiple
// denied resources of the same type are deduplicated into one RelatedResourceType
// with one instance per resource ID.
func TestBuildIAMApplyResourceTypes_MultiResourceSameType(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "2"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "3"},
	})
	if len(rts) != 1 {
		t.Fatalf("expected 1 resource type (dedup), got %d", len(rts))
	}
	if len(rts[0].Instances) != 3 {
		t.Fatalf("expected 3 instances, got %d", len(rts[0].Instances))
	}
}

// TestBuildIAMApplyResourceTypes_MultiResourceMultiType verifies that resources
// of different types produce separate RelatedResourceType entries in canonical registration order.
func TestBuildIAMApplyResourceTypes_MultiResourceMultiType(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkArea, ID: "na-1"},
	})
	if len(rts) != 2 {
		t.Fatalf("expected 2 resource types, got %d", len(rts))
	}
	if rts[0].Type != string(types.AuthResourceTypeBiz) {
		t.Fatalf("expected first type to be biz, got %q", rts[0].Type)
	}
	if rts[1].Type != string(types.AuthResourceTypeNetworkArea) {
		t.Fatalf("expected second type to be networkarea, got %q", rts[1].Type)
	}
	if len(rts[0].Instances) != 1 || len(rts[1].Instances) != 1 {
		t.Fatalf("expected 1 instance per type, got %d and %d", len(rts[0].Instances), len(rts[1].Instances))
	}
}

// TestBuildIAMApplyResourceTypes_EmptyID verifies that resources with an empty
// ID do not produce an instance entry (action-level permissions have no resource).
func TestBuildIAMApplyResourceTypes_EmptyID(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: ""},
	})
	if len(rts) != 1 {
		t.Fatalf("expected 1 resource type, got %d", len(rts))
	}
	if len(rts[0].Instances) != 0 {
		t.Fatalf("expected 0 instances for empty ID, got %d", len(rts[0].Instances))
	}
}

// TestIAMV3AuthorizerCheck_DeniedResourcesHaveInstances verifies that
// the apply request sent to IAM contains the denied resource IDs as instances.
func TestIAMV3AuthorizerCheck_DeniedResourcesHaveInstances(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": false,
			"2": false,
		},
		applyURL: "https://iam.example.com/apply",
	}
	authorizer := &iamv3Authorizer{systemID: types.SystemIDNodeMgr, handler: handler}

	err := authorizer.Check(newTestIAMContext(), ActionAgentOperate, []types.AuthResource{
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "1"},
		{SystemID: types.SystemIDCMDB, Type: types.AuthResourceTypeBiz, ID: "2"},
	})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}

	if handler.applyCalls != 1 {
		t.Fatalf("expected GetApplyURL called once, got %d", handler.applyCalls)
	}
	if len(handler.lastApplyReq.Actions) != 1 {
		t.Fatalf("expected 1 apply action, got %d", len(handler.lastApplyReq.Actions))
	}
	rts := handler.lastApplyReq.Actions[0].RelatedResourceTypes
	if len(rts) != 1 {
		t.Fatalf("expected 1 related resource type in apply req, got %d", len(rts))
	}
	if len(rts[0].Instances) != 2 {
		t.Fatalf("expected 2 instances in apply req, got %d", len(rts[0].Instances))
	}
}

// TestBuildIAMApplyResourceTypes_OrderIsCanonical verifies that resource types are emitted
// in registration-aligned canonical order regardless of input traversal order.
func TestBuildIAMApplyResourceTypes_OrderIsCanonical(t *testing.T) {
	// Supply resources in reverse canonical order: networkunit before networkarea.
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkUnit, ID: "10"},
		{SystemID: types.SystemIDNodeMgr, Type: types.AuthResourceTypeNetworkArea, ID: "5"},
	})
	if len(rts) != 2 {
		t.Fatalf("expected 2 resource types, got %d", len(rts))
	}
	// networkarea must come before networkunit per canonical registration order.
	if rts[0].Type != string(types.AuthResourceTypeNetworkArea) {
		t.Fatalf("expected first type %s, got %s", types.AuthResourceTypeNetworkArea, rts[0].Type)
	}
	if rts[1].Type != string(types.AuthResourceTypeNetworkUnit) {
		t.Fatalf("expected second type %s, got %s", types.AuthResourceTypeNetworkUnit, rts[1].Type)
	}
	// Verify instances are correctly associated after reorder.
	if len(rts[0].Instances) != 1 || rts[0].Instances[0][0].ID != "5" {
		t.Fatalf("expected networkarea instance ID=5, got %+v", rts[0].Instances)
	}
	if len(rts[1].Instances) != 1 || rts[1].Instances[0][0].ID != "10" {
		t.Fatalf("expected networkunit instance ID=10, got %+v", rts[1].Instances)
	}
}
