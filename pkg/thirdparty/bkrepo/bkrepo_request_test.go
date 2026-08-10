/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	apigwheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/header"
)

type bkrepoObservedRequest struct {
	method   string
	path     string
	query    map[string]string
	body     string
	header   http.Header
	readErr  error
	writeErr error
}

func newBKRepoRequestServer(requests []bkrepoRequestExpectation) (*httptest.Server, func() []bkrepoObservedRequest) {
	var mu sync.Mutex
	requestIndex := 0
	observed := make([]bkrepoObservedRequest, 0, len(requests))

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		body, readErr := io.ReadAll(req.Body)

		mu.Lock()
		defer mu.Unlock()

		idx := requestIndex
		requestIndex++
		if idx >= len(requests) {
			http.Error(rw, "unexpected request", http.StatusInternalServerError)
			return
		}

		expected := requests[idx]
		rw.Header().Set("Content-Type", "application/json")
		_, writeErr := rw.Write([]byte(expected.responseBody))
		observed = append(observed, bkrepoObservedRequest{
			method:   req.Method,
			path:     req.URL.Path,
			query:    expectedQueryValues(req.URL.Query()),
			body:     string(body),
			header:   req.Header.Clone(),
			readErr:  readErr,
			writeErr: writeErr,
		})
	}))

	return server, func() []bkrepoObservedRequest {
		mu.Lock()
		defer mu.Unlock()

		return append([]bkrepoObservedRequest(nil), observed...)
	}
}

func assertObservedRequests(
	t *testing.T, requests []bkrepoRequestExpectation, observed []bkrepoObservedRequest, wantTenantID string) {
	t.Helper()

	require.Len(t, observed, len(requests))
	for i, expected := range requests {
		got := observed[i]
		require.NoError(t, got.readErr)
		require.NoError(t, got.writeErr)
		assert.Equal(t, expected.method, got.method)
		assert.Equal(t, expected.path, got.path)
		assert.Equal(t, expected.query, got.query)
		assert.Equal(t, expected.body, got.body)
		assert.Equal(t, expectedAuthHeader(), got.header.Get(HeaderKeyAuth))
		assert.NotEmpty(t, got.header.Get(apigwheader.BKGWRIDKey))

		if wantTenantID == "" {
			assert.Empty(t, got.header.Values(restheader.BKTenantIDKey))
		} else {
			assert.Equal(t, wantTenantID, got.header.Get(restheader.BKTenantIDKey))
		}

		if expected.wantUploadHdr {
			assert.Equal(t, "sha256-value", got.header.Get("X-BKREPO-SHA256"))
			assert.Equal(t, "md5-value", got.header.Get("X-BKREPO-MD5"))
			assert.Equal(t, "true", got.header.Get("X-BKREPO-OVERWRITE"))
			assert.Equal(t, "7", got.header.Get("X-BKREPO-EXPIRES"))
			assert.Equal(t, "os=linux", got.header.Get("X-BKREPO-META"))
		}
	}
}
