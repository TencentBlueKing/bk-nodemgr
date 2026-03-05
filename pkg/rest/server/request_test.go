/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package server_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// fakePermProvider implements server.PermissionProvider for testing.
type fakePermProvider struct {
	system     string
	systemName string
	applyURL   string
	actions    []server.Action
}

func (f fakePermProvider) Error() string { return "permission denied" }
func (f fakePermProvider) PermissionData() server.Permission {
	return server.Permission{
		System:     f.system,
		SystemName: f.systemName,
		ApplyURL:   f.applyURL,
		Actions:    f.actions,
	}
}

// Ensure fakePermProvider satisfies both error and PermissionProvider.
var _ error = fakePermProvider{}
var _ server.PermissionProvider = fakePermProvider{}

func newTestRequest(t *testing.T) (*server.Request, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	req := server.NewRequest(ctx)
	return req, w
}

func TestAbortWithJSONPermDenied_WithPermissionProvider(t *testing.T) {
	req, w := newTestRequest(t)

	provider := fakePermProvider{
		system:     "bk_nodeman",
		systemName: "节点管理",
		applyURL:   "https://iam.example.com/apply",
		actions: []server.Action{
			{
				ID:   "agent_operate",
				Name: "Agent 操作",
				RelatedResourceTypes: []server.RelatedResourceType{
					{SystemID: "bk_cmdb", Type: "biz", TypeName: "业务"},
				},
			},
		},
	}

	req.AbortWithJSONPermDenied(resterrf.PermissionDenied, []error{provider})

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp server.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Permission == nil {
		t.Fatal("expected permission field in response, got nil")
	}
	if resp.Permission.System != "bk_nodeman" {
		t.Errorf("expected system %q, got %q", "bk_nodeman", resp.Permission.System)
	}
	if resp.Permission.SystemName != "节点管理" {
		t.Errorf("expected system_name %q, got %q", "节点管理", resp.Permission.SystemName)
	}
	if resp.Permission.ApplyURL != "https://iam.example.com/apply" {
		t.Errorf("expected apply_url %q, got %q", "https://iam.example.com/apply", resp.Permission.ApplyURL)
	}
	if len(resp.Permission.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(resp.Permission.Actions))
	}
	if resp.Permission.Actions[0].ID != "agent_operate" {
		t.Errorf("expected action ID %q, got %q", "agent_operate", resp.Permission.Actions[0].ID)
	}
}

func TestAbortWithJSONPermDenied_WithoutPermissionProvider(t *testing.T) {
	req, w := newTestRequest(t)

	plainErr := errors.New("some plain error")
	req.AbortWithJSONPermDenied(resterrf.PermissionDenied, []error{plainErr})

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp server.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Permission == nil {
		t.Fatal("expected permission field in response, got nil")
	}
	if resp.Permission.ApplyURL != "" {
		t.Errorf("expected empty apply_url, got %q", resp.Permission.ApplyURL)
	}
	if len(resp.Permission.Actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(resp.Permission.Actions))
	}
}

func TestAbortWithJSONPermDenied_TakesFirstProvider(t *testing.T) {
	req, w := newTestRequest(t)

	first := fakePermProvider{system: "bk_nodeman", systemName: "节点管理", applyURL: "https://first.example.com/apply", actions: []server.Action{{ID: "first_action"}}}
	second := fakePermProvider{system: "bk_other", systemName: "其他系统", applyURL: "https://second.example.com/apply", actions: []server.Action{{ID: "second_action"}}}

	req.AbortWithJSONPermDenied(resterrf.PermissionDenied, []error{first, second})

	var resp server.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Permission.System != "bk_nodeman" {
		t.Errorf("expected first provider's system, got %q", resp.Permission.System)
	}
	if resp.Permission.ApplyURL != "https://first.example.com/apply" {
		t.Errorf("expected first provider's apply_url, got %q", resp.Permission.ApplyURL)
	}
	if len(resp.Permission.Actions) != 1 || resp.Permission.Actions[0].ID != "first_action" {
		t.Errorf("expected first provider's actions, got %+v", resp.Permission.Actions)
	}
}

func TestAbortWithJSONPermDenied_EmptyErrs(t *testing.T) {
	req, w := newTestRequest(t)

	req.AbortWithJSONPermDenied(resterrf.PermissionDenied, []error{})

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp server.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Permission == nil {
		t.Fatal("expected permission field, got nil")
	}
}
