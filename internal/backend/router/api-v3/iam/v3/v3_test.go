/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/gin-gonic/gin"
)

// mockIAMHandler is a mock implementation of iamv3.IHandler for testing.
type mockIAMHandler struct {
	allowedUsername string
	allowedPassword string
}

func (m *mockIAMHandler) IsBasicAuthAllowed(_ contextx.IContext, username, password string) error {
	if username == m.allowedUsername && password == m.allowedPassword {
		return nil
	}

	return fmt.Errorf("invalid credentials")
}

// Implement other IHandler methods as no-ops for the mock.
func (m *mockIAMHandler) IsAllowed(_ contextx.IContext, _ iamv3.Request) (bool, error) {
	return false, nil
}

func (m *mockIAMHandler) IsAllowedWithCache(_ contextx.IContext, _ iamv3.Request, _ time.Duration) (bool, error) {
	return false, nil
}

func (m *mockIAMHandler) BatchIsAllowed(_ contextx.IContext, _ iamv3.Request,
	_ []iamv3.Resources) (map[string]bool, error) {
	return nil, nil
}

func (m *mockIAMHandler) ResourceMultiActionsAllowed(_ contextx.IContext,
	_ iamv3.MultiActionRequest) (map[string]bool, error) {
	return nil, nil
}

func (m *mockIAMHandler) BatchResourceMultiActionsAllowed(_ contextx.IContext,
	_ iamv3.MultiActionRequest, _ []iamv3.Resources) (map[string]map[string]bool, error) {
	return nil, nil
}

func (m *mockIAMHandler) GetToken(_ contextx.IContext) (string, error) {
	return m.allowedPassword, nil
}

func (m *mockIAMHandler) GetApplyURL(_ contextx.IContext, _ iamv3.Application) (string, error) {
	return "", nil
}

func (m *mockIAMHandler) GenPermissionApplyData(_ iamv3.ApplicationActionListForApply) (map[string]interface{}, error) {
	return nil, nil
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
		allowedUsername: "test_system",
		allowedPassword: "test_token",
	}
	router := setupTestRouter(mockHandler)

	// Check that the route exists by making a request (authentication will be checked separately)
	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", nil)
	req.Header.Set("Authorization", basicAuth("test_system", "test_token"))
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
		allowedUsername: "test_system",
		allowedPassword: "test_token",
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
		allowedUsername: "test_system",
		allowedPassword: "test_token",
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
		allowedUsername: "test_system",
		allowedPassword: "test_token",
	}
	router := setupTestRouter(mockHandler)

	// Make a request with valid credentials but empty body
	// The dispatcher will return 400 for bad request, but not 401
	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", nil)
	req.Header.Set("Authorization", basicAuth("test_system", "test_token"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not be 401 (auth succeeded, but request body is invalid)
	if w.Code == http.StatusUnauthorized {
		t.Errorf("Expected authentication to succeed with valid credentials, got 401")
	}
}

// TestResourceCallback_UnregisteredProvider tests that unregistered resource types return 404.
func TestResourceCallback_UnregisteredProvider(t *testing.T) {
	mockHandler := &mockIAMHandler{
		allowedUsername: "test_system",
		allowedPassword: "test_token",
	}
	router := setupTestRouter(mockHandler)

	// Send a valid callback request for an unregistered resource type
	body := `{"type": "unregistered_type", "method": "list_instance"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v3/iam/v3/resource", strings.NewReader(body))
	req.Header.Set("Authorization", basicAuth("test_system", "test_token"))
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
