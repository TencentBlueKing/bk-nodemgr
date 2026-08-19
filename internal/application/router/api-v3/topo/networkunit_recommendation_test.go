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

package topo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	backendtp "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendNetworkUnitByNetworkSegmentProxySuccess(t *testing.T) {
	fakeBackend := &fakeBackendHandler{
		results: []*types.NetworkUnitSegmentRecommendationResult{
			{NetworkAreaID: 1001, IP: "10.0.0.1", NetworkUnitID: 200101, Message: "matched"},
			{NetworkAreaID: 1001, IP: "bad-ip", NetworkUnitID: -1, Message: "invalid ip"},
		},
	}
	h := &handler{backendHandler: fakeBackend}
	req := &protoApplication.TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*protoApplication.TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
			{BkNetworkareaId: 1001, Ip: "bad-ip"},
		},
	}

	data, err := h.RecommendNetworkUnitByNetworkSegment(newMockApplicationRestContext(t, req))

	require.NoError(t, err)
	respData, ok := data.(*protoApplication.TopoRecommendNetworkUnitByNetworkSegmentResp_Data)
	require.True(t, ok)
	require.Len(t, respData.Items, 2)
	assert.Len(t, fakeBackend.receivedItems, 2)
	assert.Equal(t, int64(1001), fakeBackend.receivedItems[0].NetworkAreaID)
	assert.Equal(t, "10.0.0.1", fakeBackend.receivedItems[0].IP)
	assert.Equal(t, int64(200101), respData.Items[0].GetBkNetworkunitId())
	assert.Equal(t, "matched", respData.Items[0].GetMessage())
	assert.Equal(t, int64(-1), respData.Items[1].GetBkNetworkunitId())
	assert.Equal(t, "invalid ip", respData.Items[1].GetMessage())
}

func TestRecommendNetworkUnitByNetworkSegmentProxyFailureWrapsThirdpartyError(t *testing.T) {
	h := &handler{backendHandler: &fakeBackendHandler{err: errors.New("backend failed")}}
	req := &protoApplication.TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*protoApplication.TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
		},
	}

	_, err := h.RecommendNetworkUnitByNetworkSegment(newMockApplicationRestContext(t, req))

	require.Error(t, err)
	code, _ := resterrf.ErrUnwrap(err)
	assert.Equal(t, resterrf.ThirdpartyRequestFailed, code)
}

func TestLoadRegistersRecommendNetworkUnitByNetworkSegmentRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/api/v3")
	Load(rg, &options.Capability{BackendHandler: &fakeBackendHandler{}})

	req := httptest.NewRequest(http.MethodPost, "/api/v3/topo/networkunit/recommend_by_network_segment", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

type fakeBackendHandler struct {
	backendtp.IHandler
	results       []*types.NetworkUnitSegmentRecommendationResult
	err           error
	receivedItems []*types.NetworkUnitSegmentRecommendationItem
}

func (h *fakeBackendHandler) RecommendNetworkUnitByNetworkSegment(
	nCtx contextx.IContext, items ...*types.NetworkUnitSegmentRecommendationItem,
) ([]*types.NetworkUnitSegmentRecommendationResult, error) {
	_ = nCtx
	h.receivedItems = append([]*types.NetworkUnitSegmentRecommendationItem(nil), items...)

	return h.results, h.err
}

func newMockApplicationRestContext(t *testing.T, body interface{}) restserver.IContext {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	gCtx, _ := gin.CreateTestContext(recorder)
	gCtx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	gCtx.Request.Header.Set("Content-Type", "application/json")

	return &restserver.Context{
		Context: contextx.New(
			context.Background(),
			contextx.WithTenantID("tenant-test"),
			contextx.WithBKUsername("admin"),
			contextx.WithLoginName("admin"),
			contextx.WithMessageID("request-id"),
		),
		IRequest: restserver.NewRequest(gCtx),
	}
}
