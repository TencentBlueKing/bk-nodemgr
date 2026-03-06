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
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/test"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/cmdb"
	"gopkg.in/yaml.v3"
)

const (
	// ContentTypeKey is content type header key.
	ContentTypeKey = "Content-Type"

	// DefaultTenantID is the default tenant ID.
	DefaultTenantID = tenant.SingleModeTenantID

	// DefaultHTTPTimeout is the timeout for test HTTP requests.
	DefaultHTTPTimeout = 1 * time.Minute

	// Random number generation bounds.
	randomSuffixMin = 1000
	randomSuffixMax = 9999
	testIPOctetMin  = 1
	testIPOctetMax  = 254
	testPortMin     = 10000
	testPortMax     = 60000

	// cmdbMockDataFile is the name of the cmdb mock data file.
	cmdbMockDataFile = "cmdb.yaml"
)

// Singleton pattern for test random generator.
// nolint:gochecknoglobals
var (
	randOnce sync.Once
	randGen  *rand.Rand
)

// GetBackendBasicURL returns the backend service basic URL.
func GetBackendBasicURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.Env.Servers.BackendBasicEndpoint)
}

// GetApplicationBasicURL returns the application service basic URL.
func GetApplicationBasicURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.Env.Servers.ApplicationBasicEndpoint)
}

// GetFileBasicURL returns the file service basic URL.
func GetFileBasicURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.Env.Servers.FileBasicEndpoint)
}

// GetFileDownloadURL returns the file download service URL.
func GetFileDownloadURL() string {
	return fmt.Sprintf("http://%s/api/v3", test.Env.Servers.FileDownloadEndpoint)
}

// SendHTTPRequest sends HTTP request.
func SendHTTPRequest(t *testing.T, method, url string, body []byte) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v, method(%s), url(%s)", err, method, url)
	}
	req.Header.Set(ContentTypeKey, restserver.MIMETypeJSON.String())
	req.Header.Set(restheader.BKTenantIDKey, DefaultTenantID)

	client := &http.Client{Timeout: DefaultHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v, method(%s), url(%s)", err, method, url)
	}

	return resp
}

// SendUploadHTTPRequest sends a multipart upload request with file and metadata.
func SendUploadHTTPRequest(t *testing.T, url, filePath string, metadata []byte) *http.Response {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("metadata", string(metadata)); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	filename := filepath.Base(filePath)
	if err := writer.WriteField("filename", filename); err != nil {
		t.Fatalf("failed to write filename field: %v", err)
	}

	filePart, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	// nolint:gosec
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open file %s: %v", filePath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(filePart, f); err != nil {
		t.Fatalf("failed to copy file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatalf("failed to create upload request: %v, url(%s)", err, url)
	}
	req.Header.Set(ContentTypeKey, writer.FormDataContentType())
	req.Header.Set(restheader.BKTenantIDKey, DefaultTenantID)

	client := &http.Client{Timeout: DefaultHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send upload request: %v, url(%s)", err, url)
	}

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
	if err != nil {
		t.Fatalf("failed to read response body: %v, status(%d), status-text(%s)",
			err, resp.StatusCode, resp.Status)
	}

	if len(body) == 0 {
		t.Fatalf("response body is empty, status(%d), status-text(%s)",
			resp.StatusCode, resp.Status)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v, status(%d), status-text(%s), body(%s)",
			err, resp.StatusCode, resp.Status, string(body))
	}
}

// LoadCMDBMockData loads and parses cmdb.yaml from the test data directory.
func LoadCMDBMockData(dataDir string) (*cmdb.MockData, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data directory is empty")
	}

	path := filepath.Join(dataDir, cmdbMockDataFile)
	data, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read file from %s: %w", path, err)
	}

	var mockData cmdb.MockData
	if err := yaml.Unmarshal(data, &mockData); err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return &mockData, nil
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
