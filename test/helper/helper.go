/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package helper provides utility functions for integration tests.
package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/test"
	"github.com/stretchr/testify/require"
)

const (
	// ContentTypeKey is content type header key.
	ContentTypeKey = "Content-Type"

	// DefaultTenantID is the default tenant ID.
	DefaultTenantID = tenant.SingleModeTenantID

	// DefaultHTTPTimeout is the timeout for test HTTP requests.
	DefaultHTTPTimeout = 30 * time.Second

	// Random number generation bounds.
	randomSuffixMin = 1000
	randomSuffixMax = 9999
	testIPOctetMin  = 1
	testIPOctetMax  = 254
	testPortMin     = 10000
	testPortMax     = 60000
)

// Singleton pattern for test random generator.
// nolint:gochecknoglobals
var (
	randOnce sync.Once
	randGen  *rand.Rand
)

// GetBackendBaseURL get backend base URL.
func GetBackendBaseURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.TestFlagBackendServer)
}

// GetApplicationBaseURL get application base URL.
func GetApplicationBaseURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.TestFlagApplicationServer)
}

// GetFileBaseURL get file base URL.
func GetFileBaseURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.TestFlagFileServer)
}

// SendHTTPRequest sends HTTP request.
func SendHTTPRequest(t *testing.T, method, url string, body []byte) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	require.NoError(t, err, "failed to create request")
	req.Header.Set(ContentTypeKey, restserver.MIMETypeJSON.String())
	req.Header.Set(restheader.BKTenantIDKey, DefaultTenantID)

	client := &http.Client{Timeout: DefaultHTTPTimeout}
	resp, err := client.Do(req)

	require.NoError(t, err, "failed to send request")
	require.NotNil(t, resp, "failed to get response")

	return resp
}

// ParseResponse parses response body into the provided interface.
func ParseResponse(t *testing.T, resp *http.Response, v interface{}) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	err = json.Unmarshal(body, v)
	require.NoError(t, err, "failed to unmarshal response")
}

// GenerateRandomSuffix generates a random suffix for test data.
func GenerateRandomSuffix() string {
	return fmt.Sprintf("%d", GenRandInt(randomSuffixMin, randomSuffixMax))
}

// GenerateRandomEndpoint generates a random endpoint for testing (IP:Port format).
func GenerateRandomEndpoint() string {
	ip := fmt.Sprintf("192.168.%d.%d",
		GenRandInt(testIPOctetMin, testIPOctetMax),
		GenRandInt(testIPOctetMin, testIPOctetMax))

	port := GenRandInt(testPortMin, testPortMax)

	return fmt.Sprintf("%s:%d", ip, port)
}

// GenRandInt generates a random int between minVal and maxVal.
func GenRandInt(minVal, maxVal int) int {
	randOnce.Do(func() {
		// test code, weak random is acceptable for test data.
		// nolint:gosec
		randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	})

	return minVal + randGen.Intn(maxVal-minVal+1)
}
