/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecommendNetworkUnitByNetworkSegmentUsesNetworkAreaViewOnly(t *testing.T) {
	h, mockAuth, mockStorage := setupHandlerWithMocks(t)

	req := &protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
		},
	}

	mockAuth.listScope = auth.AuthorizedScope{Resources: buildNetworkAreaResources([]int64{1001})}
	mockStorage.results = []*types.NetworkUnitSegmentRecommendationResult{
		{NetworkAreaID: 1001, IP: "10.0.0.1", NetworkUnitID: 200101, Message: "matched"},
	}

	rCtx := newMockRestContext(t, req)
	data, err := h.RecommendNetworkUnitByNetworkSegment(rCtx)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, auth.ActionNetworkAreaView, mockAuth.lastListAction)
	assert.Equal(t, types.AuthResourceTypeNetworkArea, mockAuth.lastListResourceType)
	assert.Len(t, mockAuth.checkCalls, 0)
	assert.Len(t, mockStorage.receivedItems, 1)
	assert.Equal(t, int64(1001), mockStorage.receivedItems[0].NetworkAreaID)
	assert.Equal(t, "10.0.0.1", mockStorage.receivedItems[0].IP)
}

func TestRecommendNetworkUnitByNetworkSegmentReturnsPermissionDeniedWhenNoAuthorizedAreas(t *testing.T) {
	h, mockAuth, _ := setupHandlerWithMocks(t)

	req := &protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
		},
	}

	mockAuth.checkErr = errors.New("permission denied")

	rCtx := newMockRestContext(t, req)
	_, err := h.RecommendNetworkUnitByNetworkSegment(rCtx)

	assert.Error(t, err)
	assert.Equal(t, auth.ActionNetworkAreaView, mockAuth.lastListAction)
	assert.Equal(t, types.AuthResourceTypeNetworkArea, mockAuth.lastListResourceType)
	assert.Len(t, mockAuth.checkCalls, 1)
	assert.Nil(t, mockAuth.checkCalls[0].resources)
}

func TestRecommendNetworkUnitByNetworkSegmentReturnsBatchResults(t *testing.T) {
	h, mockAuth, mockStorage := setupHandlerWithMocks(t)

	req := &protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq{
		Items: []*protoBackend.TopoRecommendNetworkUnitByNetworkSegmentReq_Item{
			{BkNetworkareaId: 1001, Ip: "10.0.0.1"},
			{BkNetworkareaId: 1001, Ip: "bad-ip"},
		},
	}

	mockAuth.listScope = auth.AuthorizedScope{Resources: buildNetworkAreaResources([]int64{1001})}
	mockStorage.results = []*types.NetworkUnitSegmentRecommendationResult{
		{NetworkAreaID: 1001, IP: "10.0.0.1", NetworkUnitID: 200101, Message: "matched"},
		{NetworkAreaID: 1001, IP: "bad-ip", NetworkUnitID: -1, Message: "invalid ip"},
	}

	data, err := h.RecommendNetworkUnitByNetworkSegment(newMockRestContext(t, req))

	assert.NoError(t, err)
	respData, ok := data.(*protoBackend.TopoRecommendNetworkUnitByNetworkSegmentResp_Data)
	assert.True(t, ok)
	assert.Len(t, respData.Items, 2)
	assert.Equal(t, int64(200101), respData.Items[0].GetBkNetworkunitId())
	assert.Equal(t, "matched", respData.Items[0].GetMessage())
	assert.Equal(t, int64(-1), respData.Items[1].GetBkNetworkunitId())
	assert.Equal(t, "invalid ip", respData.Items[1].GetMessage())
	assert.Len(t, mockStorage.receivedItems, 2)
	assert.Equal(t, "10.0.0.1", mockStorage.receivedItems[0].IP)
	assert.Equal(t, "bad-ip", mockStorage.receivedItems[1].IP)
}

type mockAuthorizer struct {
	auth.IAuthorizer
	listScope            auth.AuthorizedScope
	listErr              error
	checkErr             error
	lastListAction       auth.Action
	lastListResourceType types.AuthResourceType
	checkCalls           []authorizerCheckCall
}

func (m *mockAuthorizer) Check(ctx contextx.IContext, action auth.Action, resources []types.AuthResource) error {
	m.checkCalls = append(m.checkCalls, authorizerCheckCall{ctx: ctx, action: action, resources: resources})

	return m.checkErr
}

func (m *mockAuthorizer) ListAuthorizedInstances(
	ctx contextx.IContext, action auth.Action, resourceType types.AuthResourceType,
) (auth.AuthorizedScope, error) {
	m.lastListAction = action
	m.lastListResourceType = resourceType
	_ = ctx

	return m.listScope, m.listErr
}

type mockStorage struct {
	topoStg.IStorage
	results       []*types.NetworkUnitSegmentRecommendationResult
	err           error
	receivedItems []*types.NetworkUnitSegmentRecommendationItem
}

func (m *mockStorage) RecommendNetworkUnitByNetworkSegment(
	nCtx contextx.IContext, items ...*types.NetworkUnitSegmentRecommendationItem,
) ([]*types.NetworkUnitSegmentRecommendationResult, error) {
	_ = nCtx
	m.receivedItems = append([]*types.NetworkUnitSegmentRecommendationItem(nil), items...)

	return m.results, m.err
}

func setupHandlerWithMocks(t *testing.T) (*handler, *mockAuthorizer, *mockStorage) {
	t.Helper()
	mockAuth := &mockAuthorizer{}
	mockStorage := &mockStorage{}

	return &handler{
		authorizer: mockAuth,
		storage:    mockStorage,
	}, mockAuth, mockStorage
}

func newMockRestContext(t *testing.T, body interface{}) restserver.IContext {
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

	request := restserver.NewRequest(gCtx)

	return &restserver.Context{
		Context: contextx.New(
			context.Background(),
			contextx.WithTenantID("tenant-test"),
			contextx.WithBKUsername("admin"),
			contextx.WithLoginName("admin"),
			contextx.WithMessageID("request-id"),
		),
		IRequest: request,
	}
}

type authorizerCheckCall struct {
	ctx       contextx.IContext
	action    auth.Action
	resources []types.AuthResource
}

func (m *mockAuthorizer) CheckMany(
	ctx contextx.IContext, actionResources map[auth.Action][]types.AuthResource,
) error {
	return fmt.Errorf("unexpected CheckMany call: %v %v", ctx, actionResources)
}

func TestMockStorageRecommendNetworkUnitByNetworkSegmentReceivesVariadicItems(t *testing.T) {
	t.Helper()

	m := &mockStorage{}
	expected := []*types.NetworkUnitSegmentRecommendationItem{{NetworkAreaID: 1, IP: "1.1.1.1"}}
	_, _ = m.RecommendNetworkUnitByNetworkSegment(contextx.New(context.Background()), expected...)

	assert.True(t, reflect.DeepEqual(expected, m.receivedItems))
}
