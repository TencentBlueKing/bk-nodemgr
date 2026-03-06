/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iamv3

import (
	"context"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestNoOpHandler_IsAllowed(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	allowed, err := h.IsAllowed(ctx, types.IAMCheckRequest{
		SystemID: "bk_nodemgr",
		Username: "admin",
		ActionID: "host_view",
	})
	if err != nil {
		t.Fatalf("IsAllowed() unexpected error: %v", err)
	}
	if !allowed {
		t.Error("IsAllowed() = false, want true (bypass)")
	}
}

func TestNoOpHandler_IsAllowedWithCache(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	allowed, err := h.IsAllowedWithCache(ctx, types.IAMCheckRequest{
		SystemID: "bk_nodemgr",
		Username: "admin",
		ActionID: "host_view",
	}, 5*time.Minute)
	if err != nil {
		t.Fatalf("IsAllowedWithCache() unexpected error: %v", err)
	}
	if !allowed {
		t.Error("IsAllowedWithCache() = false, want true (bypass)")
	}
}

func TestNoOpHandler_BatchIsAllowed(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	resourcesList := [][]types.IAMResource{
		{{SystemID: "bk_cmdb", Type: "host", ID: "host-1"}},
		{{SystemID: "bk_cmdb", Type: "host", ID: "host-2"}},
	}

	results, err := h.BatchIsAllowed(ctx, types.IAMCheckRequest{
		SystemID: "bk_nodemgr",
		Username: "admin",
		ActionID: "host_view",
	}, resourcesList)
	if err != nil {
		t.Fatalf("BatchIsAllowed() unexpected error: %v", err)
	}

	if len(results) != len(resourcesList) {
		t.Fatalf("BatchIsAllowed() returned %d results, want %d", len(results), len(resourcesList))
	}

	for key, val := range results {
		if !val {
			t.Errorf("BatchIsAllowed()[%q] = false, want true (bypass)", key)
		}
	}
}

func TestNoOpHandler_ResourceMultiActionsAllowed(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	results, err := h.ResourceMultiActionsAllowed(ctx, types.IAMMultiActionCheckRequest{
		SystemID:  "bk_nodemgr",
		Username:  "admin",
		ActionIDs: []string{"host_view", "host_edit"},
	})
	if err != nil {
		t.Fatalf("ResourceMultiActionsAllowed() unexpected error: %v", err)
	}

	for _, actionID := range []string{"host_view", "host_edit"} {
		if !results[actionID] {
			t.Errorf("ResourceMultiActionsAllowed()[%q] = false, want true (bypass)", actionID)
		}
	}
}

func TestNoOpHandler_BatchResourceMultiActionsAllowed(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	resourcesList := [][]types.IAMResource{
		{{SystemID: "bk_cmdb", Type: "host", ID: "host-1"}},
	}
	actionIDs := []string{"host_view", "host_edit"}

	results, err := h.BatchResourceMultiActionsAllowed(ctx, types.IAMMultiActionCheckRequest{
		SystemID:  "bk_nodemgr",
		Username:  "admin",
		ActionIDs: actionIDs,
	}, resourcesList)
	if err != nil {
		t.Fatalf("BatchResourceMultiActionsAllowed() unexpected error: %v", err)
	}

	if len(results) != len(resourcesList) {
		t.Fatalf("BatchResourceMultiActionsAllowed() returned %d resource groups, want %d",
			len(results), len(resourcesList))
	}

	for resKey, actionMap := range results {
		for _, actionID := range actionIDs {
			if !actionMap[actionID] {
				t.Errorf("BatchResourceMultiActionsAllowed()[%q][%q] = false, want true (bypass)",
					resKey, actionID)
			}
		}
	}
}

func TestNoOpHandler_GetToken(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	token, err := h.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken() unexpected error: %v", err)
	}
	if token != "" {
		t.Errorf("GetToken() = %q, want empty string (bypass)", token)
	}
}

func TestNoOpHandler_IsBasicAuthAllowed(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	if err := h.IsBasicAuthAllowed(ctx, "anyone", "anything"); err != nil {
		t.Errorf("IsBasicAuthAllowed() = %v, want nil (bypass)", err)
	}
}

func TestNoOpHandler_GetApplyURL(t *testing.T) {
	h := NewNoOpHandler()
	ctx := contextx.New(context.Background())

	url, err := h.GetApplyURL(ctx, types.IAMApplyRequest{
		SystemID: "bk_nodemgr",
		Actions: []types.IAMApplyAction{
			{ID: "host_view", RelatedResourceTypes: []types.IAMApplyResourceType{
				{SystemID: "bk_cmdb", Type: "host"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("GetApplyURL() unexpected error: %v", err)
	}
	if url != "" {
		t.Errorf("GetApplyURL() = %q, want empty string (bypass)", url)
	}
}
