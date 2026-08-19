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

// Package cmdb provides precheck tests that verify CMDB sync data is ready before running other test cases.
package cmdb

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/test"
	"github.com/TencentBlueKing/bk-nodemgr/test/helper"
	cmdbMock "github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/cmdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPageOffset = 0
	defaultPageLimit  = 100
)

var presetData *cmdbMock.MockData

// getBusinessListURL returns the business list API URL.
func getBusinessListURL() string {
	return helper.GetBackendBaseURL() + "/topo/business/list"
}

// getNetworkAreaListURL returns the network area list API URL.
func getNetworkAreaListURL() string {
	return helper.GetBackendBaseURL() + "/topo/networkarea/list"
}

// getHostListURL returns the host list API URL.
func getHostListURL() string {
	return helper.GetBackendBaseURL() + "/topo/host/list"
}

// TestMain parses flags and validates required flags before running tests.
func TestMain(m *testing.M) {
	test.InitFlags()
	flag.Parse()

	// load cmdb mock data.
	mockData, err := helper.LoadCMDBMockData(test.TestFlagDataDir)
	if err != nil {
		fmt.Printf("failed to load cmdb mock data: %v\n", err)
		os.Exit(1)
	}

	presetData = mockData

	os.Exit(m.Run())
}

func newDefaultPage() *protoBackend.Page {
	return &protoBackend.Page{
		Offset: defaultPageOffset,
		Limit:  defaultPageLimit,
	}
}

// newBusinessListReq creates a default business list request.
func newBusinessListReq() *protoBackend.TopoBusinessListReq {
	return &protoBackend.TopoBusinessListReq{
		Page: newDefaultPage(),
	}
}

// newNetworkAreaListReq creates a default network area list request.
func newNetworkAreaListReq() *protoBackend.TopoNetworkAreaListReq {
	return &protoBackend.TopoNetworkAreaListReq{
		Page: newDefaultPage(),
	}
}

// newHostListReq creates a default host list request.
func newHostListReq() *protoBackend.TopoHostListReq {
	return &protoBackend.TopoHostListReq{
		Page: newDefaultPage(),
	}
}

// TestSyncCMDBBusiness verifies that CMDB businesses have been synced to backend.
func TestSyncCMDBBusiness(t *testing.T) {
	require.NotEmpty(t, presetData.Businesses)

	tests := []struct {
		name      string
		checkResp func(t *testing.T, resp *protoBackend.TopoBusinessListResp)
	}{
		{
			name: "BusinessListSynced",
			checkResp: func(t *testing.T, resp *protoBackend.TopoBusinessListResp) {
				assert.Equal(t, int32(errf.OK), resp.Code)
				require.NotNil(t, resp.Data)
				assert.Equal(t, int64(len(presetData.Businesses)), resp.Data.Total)
				assertBusinessSyncResult(t, presetData.Businesses, resp.Data.Items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(newBusinessListReq())
			require.NoError(t, err)

			resp := helper.SendHTTPRequest(t, http.MethodPost, getBusinessListURL(), reqBody)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var result protoBackend.TopoBusinessListResp
			helper.ParseResponse(t, resp, &result)
			tt.checkResp(t, &result)
		})
	}
}

// TestSyncCMDBNetworkArea verifies that CMDB cloud areas have been synced to backend as network areas.
func TestSyncCMDBNetworkArea(t *testing.T) {
	require.NotEmpty(t, presetData.Areas)

	tests := []struct {
		name      string
		checkResp func(t *testing.T, resp *protoBackend.TopoNetworkAreaListResp)
	}{
		{
			name: "NetworkAreaListSynced",
			checkResp: func(t *testing.T, resp *protoBackend.TopoNetworkAreaListResp) {
				assert.Equal(t, int32(errf.OK), resp.Code)
				require.NotNil(t, resp.Data)
				require.Equal(t, int64(len(presetData.Areas)), resp.Data.Total)
				assertNetworkAreaSyncResult(t, presetData.Areas, resp.Data.Items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(newNetworkAreaListReq())
			require.NoError(t, err)

			resp := helper.SendHTTPRequest(t, http.MethodPost, getNetworkAreaListURL(), reqBody)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var result protoBackend.TopoNetworkAreaListResp
			helper.ParseResponse(t, resp, &result)
			tt.checkResp(t, &result)
		})
	}
}

// TestSyncCMDBHost verifies that CMDB hosts have been synced to backend.
func TestSyncCMDBHost(t *testing.T) {
	require.NotEmpty(t, presetData.Hosts)

	tests := []struct {
		name      string
		checkResp func(t *testing.T, resp *protoBackend.TopoHostListResp)
	}{
		{
			name: "HostListSynced",
			checkResp: func(t *testing.T, resp *protoBackend.TopoHostListResp) {
				assert.Equal(t, int32(errf.OK), resp.Code)
				require.NotNil(t, resp.Data)
				require.Equal(t, int64(len(presetData.Hosts)), resp.Data.Total)
				assertHostSyncResult(t, presetData.Hosts, resp.Data.Items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(newHostListReq())
			require.NoError(t, err)

			resp := helper.SendHTTPRequest(t, http.MethodPost, getHostListURL(), reqBody)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var result protoBackend.TopoHostListResp
			helper.ParseResponse(t, resp, &result)
			tt.checkResp(t, &result)
		})
	}
}

func assertBusinessSyncResult(t *testing.T, expected []cmdbMock.BusinessConfig, resultItems []*protoBackend.Business) {
	t.Helper()

	byID, err := conv.SliceToMap(resultItems, func(b *protoBackend.Business) int64 {
		return b.GetBkBizId()
	})
	require.NoError(t, err)
	require.Equal(t, len(expected), len(byID))

	for _, exp := range expected {
		actual, ok := byID[exp.BKBizID]
		require.True(t, ok)

		assert.Equal(t, exp.BKBizName, actual.GetBkBizName())
	}
}

func assertNetworkAreaSyncResult(t *testing.T, expected []cmdbMock.CloudAreaConfig, resultItems []*protoBackend.NetworkArea) {
	t.Helper()

	byID, err := conv.SliceToMap(resultItems, func(a *protoBackend.NetworkArea) int64 {
		return a.GetBkNetworkareaId()
	})
	require.NoError(t, err)
	require.Equal(t, len(expected), len(byID))

	for _, exp := range expected {
		actual, ok := byID[exp.BKCloudID]
		require.True(t, ok)

		assert.Equal(t, exp.BKCloudName, actual.GetBkNetworkareaName())
	}
}

func assertHostSyncResult(t *testing.T, expected []cmdbMock.HostConfig, resultItems []*protoBackend.Host) {
	t.Helper()

	byID, err := conv.SliceToMap(resultItems, func(h *protoBackend.Host) int64 {
		return h.GetBkHostId()
	})
	require.NoError(t, err)
	require.Equal(t, len(expected), len(byID))

	for _, exp := range expected {
		actual, ok := byID[exp.BKHostID]
		require.True(t, ok)

		info := actual.GetInfo()

		assert.Equal(t, exp.BKBizID, info.GetBkBizId())
		assert.Equal(t, exp.BKCloudID, info.GetBkNetworkareaId())
		assert.True(t, len(info.GetBkHostInneripList()) > 0 || len(info.GetBkHostInneripV6List()) > 0)
	}
}
