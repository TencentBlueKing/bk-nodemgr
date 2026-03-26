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
	batchResults map[string]bool
	batchErr     error
	applyURL     string
	applyErr     error

	batchCalls              int
	isAllowedWithCacheCalls int
	applyCalls              int

	lastBatchReq       types.IAMCheckRequest
	lastBatchResources [][]types.IAMResource
	lastApplyReq       types.IAMApplyRequest
}

func (h *fakeIAMBatchHandler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	return false, nil
}

func (h *fakeIAMBatchHandler) IsAllowedWithCache(
	_ contextx.IContext, _ types.IAMCheckRequest, _ time.Duration,
) (bool, error) {
	h.isAllowedWithCacheCalls++
	return false, nil
}

func (h *fakeIAMBatchHandler) BatchIsAllowed(
	_ contextx.IContext, req types.IAMCheckRequest, resourcesList [][]types.IAMResource,
) (map[string]bool, error) {
	h.batchCalls++
	h.lastBatchReq = req
	h.lastBatchResources = resourcesList
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

func newTestIAMContext() contextx.IContext {
	return contextx.New(
		context.Background(),
		contextx.WithTenantID("tenant-test"),
		contextx.WithBKUsername("admin"),
		contextx.WithLoginName("admin"),
	)
}

func TestIAMV3AuthorizerBatchCheck_AllAllowed(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": true,
			"2": true,
		},
	}
	authorizer := &iamv3Authorizer{systemID: SystemIDNodeMgr, handler: handler}

	err := authorizer.BatchCheck(newTestIAMContext(), ActionAgentOperate, []Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "2"},
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

func TestIAMV3AuthorizerBatchCheck_PartialDeniedReturnsPermissionDenied(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": true,
			"2": false,
			"3": false,
		},
		applyURL: "https://iam.example.com/apply",
	}
	authorizer := &iamv3Authorizer{systemID: SystemIDNodeMgr, handler: handler}

	err := authorizer.BatchCheck(newTestIAMContext(), ActionAgentOperate, []Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "2"},
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "3"},
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
	if handler.applyCalls != 1 {
		t.Fatalf("expected GetApplyURL to be called once, got %d", handler.applyCalls)
	}
	if len(handler.lastApplyReq.Actions) != 1 {
		t.Fatalf("expected 1 apply action, got %d", len(handler.lastApplyReq.Actions))
	}
}

func TestIAMV3AuthorizerBatchCheck_BatchErrorReturnsImmediately(t *testing.T) {
	sentinel := errors.New("batch iam failure")
	handler := &fakeIAMBatchHandler{batchErr: sentinel}
	authorizer := &iamv3Authorizer{systemID: SystemIDNodeMgr, handler: handler}

	err := authorizer.BatchCheck(newTestIAMContext(), ActionAgentOperate, []Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
	if handler.applyCalls != 0 {
		t.Fatalf("expected GetApplyURL not to be called on batch failure, got %d", handler.applyCalls)
	}
}

func TestIAMV3AuthorizerBatchCheck_ApplyURLErrorReturnsPermissionDeniedWithoutURL(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": false,
		},
		applyErr: errors.New("apply url failure"),
	}
	authorizer := &iamv3Authorizer{systemID: SystemIDNodeMgr, handler: handler}

	err := authorizer.BatchCheck(newTestIAMContext(), ActionAgentOperate, []Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
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

func TestIAMV3AuthorizerBatchCheck_EmptyResourcesReturnsNil(t *testing.T) {
	handler := &fakeIAMBatchHandler{}
	authorizer := &iamv3Authorizer{systemID: SystemIDNodeMgr, handler: handler}

	err := authorizer.BatchCheck(nil, ActionAgentOperate, nil)
	if err != nil {
		t.Fatalf("expected nil for empty resources, got: %v", err)
	}
	if handler.batchCalls != 0 {
		t.Fatalf("expected no batch IAM calls for empty resources, got %d", handler.batchCalls)
	}
}

// TestBuildIAMApplyResourceTypes_SingleResource verifies a single denied resource
// produces one RelatedResourceType with one single-node instance.
func TestBuildIAMApplyResourceTypes_SingleResource(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "42"},
	})
	if len(rts) != 1 {
		t.Fatalf("expected 1 resource type, got %d", len(rts))
	}
	rt := rts[0]
	if rt.SystemID != SystemIDCMDB || rt.Type != string(ResourceTypeBiz) {
		t.Fatalf("unexpected resource type %+v", rt)
	}
	if len(rt.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(rt.Instances))
	}
	if len(rt.Instances[0]) != 1 {
		t.Fatalf("expected instance path length 1, got %d", len(rt.Instances[0]))
	}
	node := rt.Instances[0][0]
	if node.Type != string(ResourceTypeBiz) || node.ID != "42" {
		t.Fatalf("unexpected instance node %+v", node)
	}
}

// TestBuildIAMApplyResourceTypes_MultiResourceSameType verifies that multiple
// denied resources of the same type are deduplicated into one RelatedResourceType
// with one instance per resource ID.
func TestBuildIAMApplyResourceTypes_MultiResourceSameType(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "2"},
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "3"},
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
	rts := buildIAMApplyResourceTypes([]Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
		{SystemID: SystemIDNodeMgr, Type: ResourceTypeNetworkArea, ID: "na-1"},
	})
	if len(rts) != 2 {
		t.Fatalf("expected 2 resource types, got %d", len(rts))
	}
	if rts[0].Type != string(ResourceTypeBiz) {
		t.Fatalf("expected first type to be biz, got %q", rts[0].Type)
	}
	if rts[1].Type != string(ResourceTypeNetworkArea) {
		t.Fatalf("expected second type to be networkarea, got %q", rts[1].Type)
	}
	if len(rts[0].Instances) != 1 || len(rts[1].Instances) != 1 {
		t.Fatalf("expected 1 instance per type, got %d and %d", len(rts[0].Instances), len(rts[1].Instances))
	}
}

// TestBuildIAMApplyResourceTypes_EmptyID verifies that resources with an empty
// ID do not produce an instance entry (action-level permissions have no resource).
func TestBuildIAMApplyResourceTypes_EmptyID(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: ""},
	})
	if len(rts) != 1 {
		t.Fatalf("expected 1 resource type, got %d", len(rts))
	}
	if len(rts[0].Instances) != 0 {
		t.Fatalf("expected 0 instances for empty ID, got %d", len(rts[0].Instances))
	}
}

// TestIAMV3AuthorizerBatchCheck_DeniedResourcesHaveInstances verifies that
// the apply request sent to IAM contains the denied resource IDs as instances.
func TestIAMV3AuthorizerBatchCheck_DeniedResourcesHaveInstances(t *testing.T) {
	handler := &fakeIAMBatchHandler{
		batchResults: map[string]bool{
			"1": false,
			"2": false,
		},
		applyURL: "https://iam.example.com/apply",
	}
	authorizer := &iamv3Authorizer{systemID: SystemIDNodeMgr, handler: handler}

	err := authorizer.BatchCheck(newTestIAMContext(), ActionAgentOperate, []Resource{
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "1"},
		{SystemID: SystemIDCMDB, Type: ResourceTypeBiz, ID: "2"},
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
	rts := buildIAMApplyResourceTypes([]Resource{
		{SystemID: SystemIDNodeMgr, Type: ResourceTypeNetworkUnit, ID: "10"},
		{SystemID: SystemIDNodeMgr, Type: ResourceTypeNetworkArea, ID: "5"},
	})
	if len(rts) != 2 {
		t.Fatalf("expected 2 resource types, got %d", len(rts))
	}
	// networkarea must come before networkunit per canonical registration order.
	if rts[0].Type != string(ResourceTypeNetworkArea) {
		t.Fatalf("expected first type %s, got %s", ResourceTypeNetworkArea, rts[0].Type)
	}
	if rts[1].Type != string(ResourceTypeNetworkUnit) {
		t.Fatalf("expected second type %s, got %s", ResourceTypeNetworkUnit, rts[1].Type)
	}
	// Verify instances are correctly associated after reorder.
	if len(rts[0].Instances) != 1 || rts[0].Instances[0][0].ID != "5" {
		t.Fatalf("expected networkarea instance ID=5, got %+v", rts[0].Instances)
	}
	if len(rts[1].Instances) != 1 || rts[1].Instances[0][0].ID != "10" {
		t.Fatalf("expected networkunit instance ID=10, got %+v", rts[1].Instances)
	}
}
