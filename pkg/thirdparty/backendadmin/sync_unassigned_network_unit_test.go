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

package backendadmin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestHandlerSyncUnassignedAgentNetworkUnitUsesTypedResult(t *testing.T) {
	h := &Handler{cli: &cli{}}
	h.cli.syncUnassignedAgentNetworkUnitFn = func(_ contextx.IContext, bizIDs []int64) (*types.NodeAgentAssignUnitResult, error) {
		assert.Equal(t, []int64{2, 3}, bizIDs)
		return &types.NodeAgentAssignUnitResult{SuccessCount: 1, FailedReasons: []string{}}, nil
	}

	result, err := h.SyncUnassignedAgentNetworkUnit(contextx.New(context.Background(), contextx.WithTenantID("t")), []int64{2, 3})

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.SuccessCount)
	assert.Equal(t, []string{}, result.FailedReasons)
}

func TestHandlerSyncUnassignedAgentNetworkUnitViaHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodPost, req.Method)
		require.Equal(t, "/admin/node/agent/sync_unassigned_network_unit", req.URL.Path)
		require.NotEmpty(t, req.Header.Get("X-Bk-Tenant-Id"))
		require.NotEmpty(t, req.Header.Get("X-Bknodemgr-Authorization"))

		var body struct {
			BKBizID []int64 `json:"bk_biz_id"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		assert.Equal(t, []int64{2, 3}, body.BKBizID)

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"code":0,"message":"OK","request_id":"rid","data":{"success_count":2,"failed_count":1,"failed_reasons":["host-id(1) has no inner ip"]}}`))
	}))
	defer server.Close()

	h := newHTTPTestHandler(t, server.URL)
	result, err := h.SyncUnassignedAgentNetworkUnit(
		contextx.New(context.Background(), contextx.WithTenantID("t"), contextx.WithLoginName("admin")),
		[]int64{2, 3},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(2), result.SuccessCount)
	assert.Equal(t, int64(1), result.FailedCount)
	assert.Equal(t, []string{"host-id(1) has no inner ip"}, result.FailedReasons)
}
