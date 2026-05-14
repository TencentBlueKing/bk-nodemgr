/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package web

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	bksaasbklogin "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bksaas/bklogin"
	"github.com/gin-gonic/gin"
)

type testFrontSetting struct{}

func (testFrontSetting) BKLoginURL() string                   { return "https://bklogin.example.com" }
func (testFrontSetting) BKRequestIDHeaderKey() string         { return "X-Request-Id" }
func (testFrontSetting) BKPassAnalyticsScript() template.HTML { return "" }
func (testFrontSetting) BKIamSaaSHost() string                { return "" }
func (testFrontSetting) BKUserSaaSHost() string               { return "" }
func (testFrontSetting) BKAPIGWBaseURL() string               { return "" }
func (testFrontSetting) PasswordVaultSwitch() bool            { return false }
func (testFrontSetting) PasswordVaultName() string            { return "" }
func (testFrontSetting) BKUserWebURL() string                 { return "https://bkuser.example.com" }
func (testFrontSetting) BKDomain() string                     { return "example.com" }
func (testFrontSetting) BKDocsCenterURL() string              { return "https://docs.example.com" }
func (testFrontSetting) BKAppNavOpenSourceURL() string        { return "https://nav.example.com" }
func (testFrontSetting) WindowsWMIPortDefault() int           { return 445 }
func (testFrontSetting) UnixSSHPortDefault() int              { return 36000 }
func (testFrontSetting) EnableNotice() bool                   { return false }
func (testFrontSetting) BKIamSystemIDBKNodemgr() string       { return "" }
func (testFrontSetting) BKIamSystemIDBKCmdb() string          { return "" }

var _ frontsetting.IFrontSetting = testFrontSetting{}

type testBKLoginHandler struct {
	authType string

	webUserInfo *bksaasbklogin.WebUserInfo
	webUserErr  error

	gotToken string
}

func (h *testBKLoginHandler) GetLoginURL() string { return "https://bklogin.example.com" }
func (h *testBKLoginHandler) Verify(_ contextx.IContext, _ string) (string, string, string, error) {
	return "", "", "", nil
}
func (h *testBKLoginHandler) GetAuthIdentity() *bksaasbklogin.AuthIdentity { return nil }
func (h *testBKLoginHandler) GetAuthType() string                          { return h.authType }
func (h *testBKLoginHandler) GetWebUserInfo(_ contextx.IContext, token string) (*bksaasbklogin.WebUserInfo, error) {
	h.gotToken = token

	return h.webUserInfo, h.webUserErr
}

func TestHandlerIndexInjectLoginNameByAuthType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		authType  string
		cookieVal string
		username  string
	}{
		{
			name:      "bk_ticket",
			authType:  bksaasbklogin.CookieKeyBKTicket,
			cookieVal: "ticket-value",
			username:  "ticket_user",
		},
		{
			name:      "bk_token",
			authType:  bksaasbklogin.CookieKeyBKToken,
			cookieVal: "token-value",
			username:  "token_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, engine := gin.CreateTestContext(recorder)
			engine.SetHTMLTemplate(template.Must(template.New("index.html").Parse(`{{.LOGIN_NAME}}`)))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  tt.authType,
				Value: tt.cookieVal,
			})
			ctx.Request = req

			bkHandler := &testBKLoginHandler{
				authType:    tt.authType,
				webUserInfo: &bksaasbklogin.WebUserInfo{Username: tt.username},
			}

			h := &handler{
				frontSetting:   testFrontSetting{},
				bkloginHandler: bkHandler,
			}

			h.Index(ctx)

			if recorder.Code != http.StatusOK {
				t.Fatalf("Index() status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if body := recorder.Body.String(); body != tt.username {
				t.Fatalf("Index() body = %q, want %q", body, tt.username)
			}
			if bkHandler.gotToken != tt.cookieVal {
				t.Fatalf("Index() token = %q, want %q", bkHandler.gotToken, tt.cookieVal)
			}
		})
	}
}

func TestHandlerIndexFallbackWhenGetWebUserInfoFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)
	engine.SetHTMLTemplate(template.Must(template.New("index.html").Parse(`{{.LOGIN_NAME}}`)))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  bksaasbklogin.CookieKeyBKToken,
		Value: "token-value",
	})
	ctx.Request = req

	h := &handler{
		frontSetting: testFrontSetting{},
		bkloginHandler: &testBKLoginHandler{
			authType:   bksaasbklogin.CookieKeyBKToken,
			webUserErr: errors.New("upstream failed"),
		},
	}

	h.Index(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Index() status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "" {
		t.Fatalf("Index() body = %q, want empty", body)
	}
}
