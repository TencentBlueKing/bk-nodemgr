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

package v3

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// mockIAMHandler is a mock implementation of iamv3.IHandler for testing.
// It mirrors the real Handler's IsBasicAuthAllowed semantics: username must be
// "bk_iam" and password must match the system token, per iam-go-sdk convention.
type mockIAMHandler struct {
	token string
}

var _ iamv3.IHandler = (*mockIAMHandler)(nil)

func (m *mockIAMHandler) IsBasicAuthAllowed(_ contextx.IContext, username, password string) error {
	if username != "bk_iam" {
		return fmt.Errorf("invalid credentials")
	}

	if password != m.token {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

// Implement other IHandler methods as no-ops for the mock.
func (m *mockIAMHandler) IsAllowed(_ contextx.IContext, _ types.IAMCheckRequest) (bool, error) {
	return false, nil
}

func (m *mockIAMHandler) IsAllowedWithCache(_ contextx.IContext, _ types.IAMCheckRequest, _ time.Duration) (bool, error) {
	return false, nil
}

func (m *mockIAMHandler) BatchIsAllowed(_ contextx.IContext, _ types.IAMCheckRequest,
	_ [][]types.IAMResource) (map[string]bool, error) {
	return nil, nil
}

func (m *mockIAMHandler) ResourceMultiActionsAllowed(_ contextx.IContext,
	_ types.IAMMultiActionCheckRequest) (map[string]bool, error) {
	return nil, nil
}

func (m *mockIAMHandler) BatchResourceMultiActionsAllowed(_ contextx.IContext,
	_ types.IAMMultiActionCheckRequest, _ [][]types.IAMResource) (map[string]map[string]bool, error) {
	return nil, nil
}

func (m *mockIAMHandler) GetToken(_ contextx.IContext) (string, error) {
	return m.token, nil
}

func (m *mockIAMHandler) GetApplyURL(_ contextx.IContext, _ types.IAMApplyRequest) (string, error) {
	return "", nil
}

func (m *mockIAMHandler) ListAuthorizedInstances(
	_ contextx.IContext,
	_ types.IAMAuthorizedInstancesRequest,
) (bool, []types.IAMResource, error) {
	return false, nil, nil
}

func setupTestRouter(mockHandler *mockIAMHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	capability := &options.Capability{
		IAMV3Handler: mockHandler,
	}

	rg := router.Group("/api/v3/iam")
	Load(rg, capability)

	return router
}

func basicAuth(username, password string) string {
	auth := username + ":" + password

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// TestRouteRegistration verifies that the IAM routes are registered correctly.
func TestRouteRegistration(t *testing.T) {
	mockHandler := &mockIAMHandler{
		token: "test_token",
	}
	router := setupTestRouter(mockHandler)

	// Check that the route exists by making a request (authentication will be checked separately)
	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", nil)
	req.Header.Set("Authorization", basicAuth("bk_iam", "test_token"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The route should exist (we should get a response, not 404)
	// Note: We may get 400 due to empty body, but not 404
	if w.Code == http.StatusNotFound {
		t.Errorf("Route /api/v3/iam/v3/resource should be registered, got 404")
	}
}

// TestBasicAuthMiddleware_NoCredentials tests that requests without credentials are rejected.
func TestBasicAuthMiddleware_NoCredentials(t *testing.T) {
	mockHandler := &mockIAMHandler{
		token: "test_token",
	}
	router := setupTestRouter(mockHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for missing credentials, got %d", w.Code)
	}

	// Check WWW-Authenticate header is set
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("Expected WWW-Authenticate header to be set")
	}
}

// TestBasicAuthMiddleware_InvalidCredentials tests that invalid credentials are rejected.
func TestBasicAuthMiddleware_InvalidCredentials(t *testing.T) {
	mockHandler := &mockIAMHandler{
		token: "test_token",
	}
	router := setupTestRouter(mockHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", nil)
	req.Header.Set("Authorization", basicAuth("wrong_user", "wrong_pass"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for invalid credentials, got %d", w.Code)
	}
}

// TestBasicAuthMiddleware_ValidCredentials tests that valid credentials are accepted.
func TestBasicAuthMiddleware_ValidCredentials(t *testing.T) {
	mockHandler := &mockIAMHandler{
		token: "test_token",
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := &handler{
		capability: &options.Capability{IAMV3Handler: mockHandler},
	}
	router.Use(h.basicAuthMiddleware())
	router.POST("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/ping", nil)
	req.Header.Set("Authorization", basicAuth("bk_iam", "test_token"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for valid credentials, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestBasicAuthMiddleware_ValidCredentials_ShouldInjectDefaultTenant(t *testing.T) {
	mockHandler := &mockIAMHandler{
		token: "test_token",
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := &handler{
		capability: &options.Capability{IAMV3Handler: mockHandler},
	}
	router.Use(restserver.MiddlewareContext())
	router.Use(h.basicAuthMiddleware())
	router.POST("/tenant", func(c *gin.Context) {
		rCtx, err := restserver.GenRestContext(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)

			return
		}

		c.JSON(http.StatusOK, gin.H{"tenant_id": rCtx.Data().GetTenantID()})
	})

	req := httptest.NewRequest(http.MethodPost, "/tenant", nil)
	req.Header.Set("Authorization", basicAuth("bk_iam", "test_token"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK for valid credentials, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		TenantID string `json:"tenant_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.TenantID != tenant.SingleModeTenantID {
		t.Fatalf("Expected injected tenant_id %q, got %q", tenant.SingleModeTenantID, resp.TenantID)
	}
}

// TestResourceCallback_UnregisteredProvider tests that unregistered resource types return 404.
func TestResourceCallback_UnregisteredProvider(t *testing.T) {
	mockHandler := &mockIAMHandler{
		token: "test_token",
	}
	router := setupTestRouter(mockHandler)

	// Send a valid callback request for an unregistered resource type
	body := `{"type": "unregistered_type", "method": "list_instance"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", strings.NewReader(body))
	req.Header.Set("Authorization", basicAuth("bk_iam", "test_token"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The dispatcher should return 200 with a response containing code 404
	if w.Code != http.StatusOK {
		t.Errorf("Expected HTTP status 200, got %d", w.Code)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != 404 {
		t.Errorf("Expected response code 404 for unregistered provider, got %d", resp.Code)
	}

	if !strings.Contains(resp.Message, "unregistered_type") {
		t.Errorf("Expected error message to contain resource type, got: %s", resp.Message)
	}
}
