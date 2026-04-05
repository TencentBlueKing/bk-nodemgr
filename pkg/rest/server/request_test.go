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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
)

// fakePermProvider implements resterrf.PermissionError for testing.
type fakePermProvider struct {
	system     string
	systemName string
	applyURL   string
	actions    []resterrf.Action
}

func (f fakePermProvider) Error() string { return "permission denied" }
func (f fakePermProvider) PermissionData() resterrf.Permission {
	return resterrf.Permission{
		System:     f.system,
		SystemName: f.systemName,
		ApplyURL:   f.applyURL,
		Actions:    f.actions,
	}
}

// Ensure fakePermProvider satisfies both error and PermissionError.
var _ error = fakePermProvider{}
var _ resterrf.PermissionError = fakePermProvider{}

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
		actions: []resterrf.Action{
			{
				ID:   "agent_operate",
				Name: "Agent 操作",
				RelatedResourceTypes: []resterrf.RelatedResourceType{
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
	if resp.Permission.Actions != nil {
		t.Errorf("expected nil actions, got %+v", resp.Permission.Actions)
	}
}

func TestAbortWithJSONPermDenied_TakesFirstProvider(t *testing.T) {
	req, w := newTestRequest(t)

	first := fakePermProvider{system: "bk_nodeman", systemName: "节点管理", applyURL: "https://first.example.com/apply", actions: []resterrf.Action{{ID: "first_action"}}}
	second := fakePermProvider{system: "bk_other", systemName: "其他系统", applyURL: "https://second.example.com/apply", actions: []resterrf.Action{{ID: "second_action"}}}

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

func TestAbortWithJSONPermDenied_WrappedErrorChain(t *testing.T) {
	req, w := newTestRequest(t)

	inner := fakePermProvider{
		system:     "bk_nodeman",
		systemName: "节点管理",
		applyURL:   "https://iam.example.com/apply",
		actions:    []resterrf.Action{{ID: "wrapped_action", Name: "Wrapped Action"}},
	}
	wrapped := fmt.Errorf("outer context: %w", inner)

	req.AbortWithJSONPermDenied(resterrf.PermissionDenied, []error{wrapped})

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
	if resp.Permission.ApplyURL != "https://iam.example.com/apply" {
		t.Errorf("expected apply_url %q, got %q", "https://iam.example.com/apply", resp.Permission.ApplyURL)
	}
	if len(resp.Permission.Actions) != 1 || resp.Permission.Actions[0].ID != "wrapped_action" {
		t.Errorf("expected wrapped provider's actions, got %+v", resp.Permission.Actions)
	}
}

func TestAbortWithJSONPermDenied_EmptySystemRetainsDefaults(t *testing.T) {
	req, w := newTestRequest(t)

	// Provider returns empty system/systemName — defaults should be preserved.
	provider := fakePermProvider{
		system:     "",
		systemName: "",
		applyURL:   "https://iam.example.com/apply",
		actions:    []resterrf.Action{{ID: "some_action"}},
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
	if resp.Permission.System != system.Code {
		t.Errorf("expected default system %q, got %q", system.Code, resp.Permission.System)
	}
	if resp.Permission.SystemName != system.Name {
		t.Errorf("expected default system_name %q, got %q", system.Name, resp.Permission.SystemName)
	}
	if resp.Permission.ApplyURL != "https://iam.example.com/apply" {
		t.Errorf("expected apply_url %q, got %q", "https://iam.example.com/apply", resp.Permission.ApplyURL)
	}
	if len(resp.Permission.Actions) != 1 || resp.Permission.Actions[0].ID != "some_action" {
		t.Errorf("expected provider's actions, got %+v", resp.Permission.Actions)
	}
}

func TestHandler_ThirdpartyWrappedPermissionErrorRoutesToPermDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("rest_request", server.NewRequest(c))
		c.Next()
	})
	engine.GET("/permission", server.Handler(func(rCtx server.IContext) (interface{}, error) {
		provider := fakePermProvider{
			system:     "bk_nodeman",
			systemName: "节点管理",
			applyURL:   "https://iam.example.com/apply",
			actions:    []resterrf.Action{{ID: "agent_operate"}},
		}

		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, provider)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/permission", nil)
	engine.ServeHTTP(w, req)

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
	if resp.Permission.ApplyURL != "https://iam.example.com/apply" {
		t.Errorf("expected apply_url %q, got %q", "https://iam.example.com/apply", resp.Permission.ApplyURL)
	}
}

type countingAuthIdentity struct {
	verifyCalls int
	err         error
}

func (i *countingAuthIdentity) Verify(_ server.IRequest) error {
	i.verifyCalls++

	return i.err
}

func TestMiddlewareAuth_WithSkipPathPrefixes_ShouldSkipForIAMPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	identity := &countingAuthIdentity{err: errors.New("verify failed")}
	engine.Use(server.MiddlewareContext())
	engine.Use(server.MiddlewareAuth(identity, server.WithSkipPathPrefixes("/api/v3/iam")))
	engine.GET("/api/v3/iam/v3/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v3/iam/v3/resource", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 for skipped path, got %d", recorder.Code)
	}
	if identity.verifyCalls != 0 {
		t.Fatalf("expected verify not called for skipped path, got %d", identity.verifyCalls)
	}
}

func TestMiddlewareAuth_WithSkipPathPrefixes_ShouldVerifyForNonSkippedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	identity := &countingAuthIdentity{err: errors.New("verify failed")}
	engine.Use(server.MiddlewareContext())
	engine.Use(server.MiddlewareAuth(identity, server.WithSkipPathPrefixes("/api/v3/iam")))
	engine.GET("/api/v3/node/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v3/node/list", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for non-skipped path, got %d", recorder.Code)
	}
	if identity.verifyCalls != 1 {
		t.Fatalf("expected verify called once for non-skipped path, got %d", identity.verifyCalls)
	}
}

func TestMiddlewareAuth_WithSkipPathPrefixes_ShouldNotSkipPrefixCollisionPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	identity := &countingAuthIdentity{err: errors.New("verify failed")}
	engine.Use(server.MiddlewareContext())
	engine.Use(server.MiddlewareAuth(identity, server.WithSkipPathPrefixes("/api/v3/iam")))
	engine.GET("/api/v3/iamx/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v3/iamx/resource", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for prefix-collision path, got %d", recorder.Code)
	}
	if identity.verifyCalls != 1 {
		t.Fatalf("expected verify called once for prefix-collision path, got %d", identity.verifyCalls)
	}
}

func TestMiddlewareAuth_WithoutOptions_ShouldRemainBackwardCompatible(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	identity := &countingAuthIdentity{err: errors.New("verify failed")}
	engine.Use(server.MiddlewareContext())
	engine.Use(server.MiddlewareAuth(identity))
	engine.GET("/api/v3/iam/v3/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v3/iam/v3/resource", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for middleware without options, got %d", recorder.Code)
	}
	if identity.verifyCalls != 1 {
		t.Fatalf("expected verify called once for middleware without options, got %d", identity.verifyCalls)
	}
}
