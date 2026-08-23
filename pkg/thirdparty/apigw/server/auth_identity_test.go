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

package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

func TestBKGWJWTAuthIdentityAppStateVerifyResolvesBKUsername(t *testing.T) {
	access.SetVirtualUserResolver(testVirtualUserResolver{})
	t.Cleanup(func() { access.SetVirtualUserResolver(access.NewIdentityVirtualUserResolver()) })

	privateKey, publicKeyPEM := newTestRSAKey(t)
	jwtToken := newTestJWTToken(t, privateKey, &bkAppStateClaims{
		App: app{
			AppCode:  "bk-nodemgr",
			Verified: true,
		},
		User: user{
			UserName: "bk-nodemgr",
		},
	})
	req := newTestRequest(t, "tenant-a", jwtToken)

	err := NewBKGWJWTAuthIdentity(publicKeyPEM).Verify(req)
	require.NoError(t, err)
	assert.Equal(t, "tenant-a", req.Data().GetTenantID())
	assert.Equal(t, "bk-nodemgr", req.Data().GetLoginName())
	assert.Equal(t, "bk-nodemgr@tenant-a", req.Data().GetBKUsername())
}

func TestBKGWJWTAuthIdentityUserStateVerifyResolvesBKUsername(t *testing.T) {
	access.SetVirtualUserResolver(testVirtualUserResolver{})
	t.Cleanup(func() { access.SetVirtualUserResolver(access.NewIdentityVirtualUserResolver()) })

	privateKey, publicKeyPEM := newTestRSAKey(t)
	jwtToken := newTestJWTToken(t, privateKey, &bkUserStateClaims{
		App: app{
			AppCode:  "bk-nodemgr",
			Verified: true,
		},
		User: user{
			UserName: "bk-nodemgr",
			Verified: true,
		},
	})
	req := newTestRequest(t, "tenant-a", jwtToken)
	restCtx := &restserver.Context{
		Context:  contextx.New(req.GetRequest().Context()),
		IRequest: req,
	}

	err := NewBKGWJWTAuthIdentityUserState(publicKeyPEM).Verify(restCtx)
	require.NoError(t, err)
	assert.Equal(t, "tenant-a", req.Data().GetTenantID())
	assert.Equal(t, "bk-nodemgr", req.Data().GetLoginName())
	assert.Equal(t, "bk-nodemgr@tenant-a", req.Data().GetBKUsername())
}

type testVirtualUserResolver struct{}

func (testVirtualUserResolver) GetBKUsernameByLoginName(nCtx contextx.IContext, loginName string) (string, error) {
	return loginName + "@" + nCtx.TenantID(), nil
}

func newTestRequest(t *testing.T, tenantID string, jwtToken string) *restserver.Request {
	t.Helper()

	gin.SetMode(gin.TestMode)
	gCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	gCtx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	gCtx.Request.Header.Set(apigwheader.BKGWTenantIDKey, tenantID)
	gCtx.Request.Header.Set(apigwheader.BKGWJWTTokenKey, jwtToken)

	return restserver.NewRequest(gCtx)
}

func newTestRSAKey(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	return privateKey, publicKeyPEM
}

func newTestJWTToken(t *testing.T, privateKey *rsa.PrivateKey, claims jwt.Claims) string {
	t.Helper()

	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	require.NoError(t, err)

	return jwtToken
}
