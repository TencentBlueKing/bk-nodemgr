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

package bklogin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
	"github.com/stretchr/testify/require"
)

func TestHandlerVerify(t *testing.T) {
	mode, ok := configuredTenantMode(t)
	if !ok {
		for _, mode := range []tenant.Mode{tenant.ModeSingle, tenant.ModeMultiple} {
			t.Run(string(mode), func(t *testing.T) {
				runTenantModeTest(t, "TestHandlerVerify", mode)
			})
		}
		return
	}

	nCtx := contextx.New(context.Background())

	switch mode {
	case tenant.ModeSingle:
		t.Run("single_bk_token_uses_legacy_api", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, req *http.Request) {
				require.Equal(t, "/accounts/get_user/", req.URL.Path)
				require.Equal(t, "token-value", req.URL.Query().Get(CookieKeyBKToken))

				_, err := rw.Write([]byte(`{"result":true,"code":"00","message":"ok","data":{"username":"token_user"}}`))
				require.NoError(t, err)
			})

			tenantID, bkUsername, loginName, err := h.Verify(nCtx, "token-value")
			require.NoError(t, err)
			require.Equal(t, tenant.SingleModeTenantID, tenantID)
			require.Equal(t, "token_user", bkUsername)
			require.Equal(t, "token_user", loginName)
		})
	case tenant.ModeMultiple:
		t.Run("multiple_bk_token_uses_multiple_tenant_mode_api", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, req *http.Request) {
				require.Equal(t, "/login/api/v3/open/bk-tokens/userinfo/", req.URL.Path)
				require.Equal(t, "token-value", req.URL.Query().Get(CookieKeyBKToken))
				require.JSONEq(t, `{
						"bk_app_code": "bk-nodemgr",
						"bk_app_secret": "app-secret",
						"bk_username": "admin"
					}`, req.Header.Get(apigwheader.BKGWAuthKey))

				_, err := rw.Write([]byte(`{
					"data": {
						"bk_username": "nteuuhzxlh0jcanw",
						"tenant_id": "system",
						"login_name": "admin",
						"display_name": "admin",
						"language": "zh-cn",
						"time_zone": "Asia/Shanghai"
					}
				}`))
				require.NoError(t, err)
			})

			tenantID, bkUsername, loginName, err := h.Verify(nCtx, "token-value")
			require.NoError(t, err)
			require.Equal(t, tenant.SystemTenantID, tenantID)
			require.Equal(t, "nteuuhzxlh0jcanw", bkUsername)
			require.Equal(t, "admin", loginName)
		})

		t.Run("multiple_bk_token_maps_multiple_tenant_mode_error", func(t *testing.T) {
			h := newTestHandler(t, CookieKeyBKToken, func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusBadRequest)
				_, err := rw.Write([]byte(`{"error":{"code":"VALIDATION_ERROR","message":"登录态已过期"}}`))
				require.NoError(t, err)
			})

			_, _, _, err := h.Verify(nCtx, "token-value")
			require.Error(t, err)
			require.Contains(t, err.Error(), "VALIDATION_ERROR")
			require.Contains(t, err.Error(), "登录态已过期")
		})
	default:
		require.Failf(t, "unsupported tenant mode", "mode=%s", mode)
	}
}

func TestHandlerVerifyUsesClientEndpointAndKeepsLoginURL(t *testing.T) {
	backendCalled := false
	backendSrv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		backendCalled = true
		require.Equal(t, "/user/get_info/", req.URL.Path)

		_, err := rw.Write([]byte(`{"ret":0,"msg":"ok","data":{"username":"ticket_user"}}`))
		require.NoError(t, err)
	}))
	t.Cleanup(backendSrv.Close)

	h := newTestHandlerWithEndpoint(t, CookieKeyBKTicket, backendSrv.URL)
	require.Equal(t, "https://bklogin.example.com/login", h.GetLoginURL())

	_, _, _, err := h.Verify(contextx.New(context.Background()), "ticket-value")
	require.NoError(t, err)
	require.True(t, backendCalled)
}
