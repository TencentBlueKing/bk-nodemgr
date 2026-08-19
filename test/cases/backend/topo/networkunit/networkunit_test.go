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

// Package networkunit provides test cases for network unit API.
package networkunit

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"testing"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/test"
	"github.com/TencentBlueKing/bk-nodemgr/test/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: 补全 test cases, 并使用测试提供的预设数据.

// 此部分先使用硬编码, 后续再改为使用测试提供的预设数据.
const defaultNetworkAreaID = 0

// getNetworkUnitCreateURL get create network unit URL.
func getNetworkUnitCreateURL() string {
	return helper.GetBackendBaseURL() + "/topo/networkunit/create"
}

// getNetworkUnitListURL get list network unit URL.
func getNetworkUnitListURL() string {
	return helper.GetBackendBaseURL() + "/topo/networkunit/list"
}

// TestMain parses flags before running tests.
func TestMain(m *testing.M) {
	test.InitFlags()
	flag.Parse()

	os.Exit(m.Run())
}

// newCreateNetworkUnitReq creates a default network unit creation request
func newCreateNetworkUnitReq() *protoBackend.TopoNetworkUnitCreateReq {
	return &protoBackend.TopoNetworkUnitCreateReq{
		BkNetworkareaId:   defaultNetworkAreaID,
		BkNetworkunitName: "test-networkunit-" + helper.GenerateRandomSuffix(),
		IsDirect:          true,
		DirectEndpoints: &protoBackend.Endpoints{
			Cluster: []string{helper.GenerateRandomEndpoint()},
			File:    []string{helper.GenerateRandomEndpoint()},
			Data:    []string{helper.GenerateRandomEndpoint()},
		},
	}
}

// TestNetworkUnitCreate tests network unit create.
func TestNetworkUnitCreate(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func() *protoBackend.TopoNetworkUnitCreateReq
		wantSuccess bool
		checkResp   func(t *testing.T, resp *protoBackend.TopoNetworkUnitCreateResp)
	}{
		{
			name: "CreateSuccess",
			setupReq: func() *protoBackend.TopoNetworkUnitCreateReq {
				return newCreateNetworkUnitReq()
			},
			wantSuccess: true,
			checkResp: func(t *testing.T, resp *protoBackend.TopoNetworkUnitCreateResp) {
				require.NotNil(t, resp.Data, "response data should not be nil")
				assert.NotNil(t, resp.Data.BkNetworkunitId, "network unit ID should not be nil")
			},
		},
		{
			name: "CreateWithEmptyName",
			setupReq: func() *protoBackend.TopoNetworkUnitCreateReq {
				req := newCreateNetworkUnitReq()
				req.BkNetworkunitName = ""
				return req
			},
			wantSuccess: false,
			checkResp: func(t *testing.T, resp *protoBackend.TopoNetworkUnitCreateResp) {
				assert.Equal(t, int32(errf.InvalidParameter), resp.Code, "expected code %d, got %d: %s", errf.InvalidParameter, resp.Code, resp.Message)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			reqBody, err := json.Marshal(req)
			require.NoError(t, err, "failed to marshal request body")

			url := getNetworkUnitCreateURL()
			resp := helper.SendHTTPRequest(t, http.MethodPost, url, reqBody)

			var result protoBackend.TopoNetworkUnitCreateResp
			helper.ParseResponse(t, resp, &result)

			if tt.wantSuccess {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Equal(t, int32(errf.OK), result.Code, "expected successful creation, got code=%d, message=%s", result.Code, result.Message)
			}

			tt.checkResp(t, &result)
			if result.Code != 0 {
				t.Logf("Create() error = code:%d, message:%s", result.Code, result.Message)
			}
			if result.Data != nil && result.Data.BkNetworkunitId != nil {
				t.Logf("Create() got id = %d", *result.Data.BkNetworkunitId)
			}
		})
	}
}

// TestNetworkUnitList tests network unit list query.
func TestNetworkUnitList(t *testing.T) {
	tests := []struct {
		name      string
		setupReq  func() *protoBackend.TopoNetworkUnitListReq
		wantError bool
		checkResp func(t *testing.T, resp *protoBackend.TopoNetworkUnitListResp)
	}{
		{
			name: "ListSuccess",
			setupReq: func() *protoBackend.TopoNetworkUnitListReq {
				return &protoBackend.TopoNetworkUnitListReq{
					Page: &protoBackend.Page{
						Offset: 0,
						Limit:  20,
					},
				}
			},
			wantError: false,
			checkResp: func(t *testing.T, resp *protoBackend.TopoNetworkUnitListResp) {
				assert.NotNil(t, resp.Data)
				assert.GreaterOrEqual(t, resp.Data.Total, int64(0), "unexpected total")
				t.Logf("List() total = %d, got = %d", resp.Data.Total, len(resp.Data.Items))
			},
		},
		{
			name: "ListWithInvalidNegativeOffset",
			setupReq: func() *protoBackend.TopoNetworkUnitListReq {
				return &protoBackend.TopoNetworkUnitListReq{
					Page: &protoBackend.Page{
						Offset: -1,
						Limit:  10,
					},
				}
			},
			wantError: true,
			checkResp: func(t *testing.T, resp *protoBackend.TopoNetworkUnitListResp) {
				assert.Equal(t, int32(errf.InvalidParameter), resp.Code, "expected code %d, got %d: %s", errf.InvalidParameter, resp.Code, resp.Message)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			reqBody, err := json.Marshal(req)
			require.NoError(t, err, "failed to marshal request body")

			url := getNetworkUnitListURL()
			resp := helper.SendHTTPRequest(t, http.MethodPost, url, reqBody)

			var result protoBackend.TopoNetworkUnitListResp
			helper.ParseResponse(t, resp, &result)

			if !tt.wantError {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Equal(t, int32(errf.OK), result.Code, "expected successful list, got code=%d, message=%s", result.Code, result.Message)
			}

			tt.checkResp(t, &result)
		})
	}
}
