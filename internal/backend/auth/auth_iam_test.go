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
