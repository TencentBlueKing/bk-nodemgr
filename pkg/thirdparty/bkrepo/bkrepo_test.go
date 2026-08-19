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

package bkrepo

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
)

const (
	testTenantModeEnv = "BKREPO_TENANT_MODE_TEST"
	testProjectID     = "blueking"
	testRepoName      = "agent"
	testUsername      = "repo-user"
	testPassword      = "repo-password"
	testContextTenant = "tenant-from-context"
)

type bkrepoTestTraceService struct{}

func (bkrepoTestTraceService) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}

func (bkrepoTestTraceService) ServiceName() string {
	return "bkrepo_test"
}

func (bkrepoTestTraceService) Shutdown(_ context.Context) error {
	return nil
}

func (bkrepoTestTraceService) TracerPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

type bkrepoRequestExpectation struct {
	method        string
	path          string
	query         map[string]string
	body          string
	responseBody  string
	wantUploadHdr bool
}

func TestClientRequestsRespectTenantMode(t *testing.T) {
	mode := os.Getenv(testTenantModeEnv)
	if mode != "" {
		runClientRequestsRespectTenantMode(t, tenant.Mode(mode))
		return
	}

	for _, mode := range []tenant.Mode{tenant.ModeSingle, tenant.ModeMultiple} {
		t.Run(string(mode), func(t *testing.T) {
			executable, err := os.Executable()
			require.NoError(t, err)

			cmd := exec.Command(executable, "-test.run=^TestClientRequestsRespectTenantMode$", "-test.v")
			cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", testTenantModeEnv, mode))
			output, err := cmd.CombinedOutput()

			require.NoError(t, err, string(output))
		})
	}
}

func runClientRequestsRespectTenantMode(t *testing.T, mode tenant.Mode) {
	t.Helper()

	tenant.SetMode(mode)

	effectiveProjectID := testProjectID
	wantTenantID := ""
	if mode == tenant.ModeMultiple {
		effectiveProjectID = tenant.SystemTenantID + "." + testProjectID
		wantTenantID = tenant.SystemTenantID
	}

	requests := []bkrepoRequestExpectation{
		{
			method:       http.MethodGet,
			path:         "/generic/" + effectiveProjectID + "/" + testRepoName + "/packages/agent.tgz",
			query:        map[string]string{"download": "true"},
			responseBody: "agent-content",
		},
		{
			method:        http.MethodPut,
			path:          "/generic/" + effectiveProjectID + "/" + testRepoName + "/packages/agent.tgz",
			body:          "upload-content",
			responseBody:  `{"code":0,"message":"OK","data":{"projectId":"` + effectiveProjectID + `","repoName":"` + testRepoName + `"}}`,
			wantUploadHdr: true,
		},
		{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/" + effectiveProjectID + "/" + testRepoName + "/packages/agent.tgz",
			responseBody: `{"code":0,"message":"OK","data":{"nodeInfo":{"projectId":"` + effectiveProjectID + `","repoName":"` + testRepoName + `"}}}`,
		},
		{
			method: http.MethodGet,
			path:   "/repository/api/node/page/" + effectiveProjectID + "/" + testRepoName + "/packages/",
			query: map[string]string{
				"pageSize":        "20",
				"pageNum":         "1",
				"includeFolder":   "true",
				"includeMetadata": "false",
				"sortProperty":    "id",
				"direction":       "DESC",
			},
			responseBody: `{"code":0,"message":"OK","data":{"pageNumber":1,"pageSize":20,"records":[]}}`,
		},
		{
			method:       http.MethodPost,
			path:         "/repository/api/node/mkdir/" + effectiveProjectID + "/" + testRepoName + "/packages/new/",
			responseBody: `{"code":0,"message":"OK","data":{}}`,
		},
	}

	server, observedRequests := newBKRepoRequestServer(requests)
	defer server.Close()

	client := newBKRepoTestClient(t, server.URL)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testContextTenant))

	downloadResp, err := client.DownloadFile(nCtx, &DownloadFileReq{Path: "packages/agent.tgz"})
	require.NoError(t, err)
	assert.Equal(t, []byte("agent-content"), downloadResp.Data)

	uploadResp, err := client.UploadFile(nCtx, &UploadFileReq{
		Path:   "packages/agent.tgz",
		Reader: bytes.NewReader([]byte("upload-content")),
		Info: UploadFileInfo{
			Sha256:     "sha256-value",
			Md5:        "md5-value",
			Overwrite:  true,
			ExpireDays: 7,
			Meta:       map[string]string{"os": "linux"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, effectiveProjectID, uploadResp.ProjectID)
	assert.Equal(t, testRepoName, uploadResp.RepoName)

	nodeResp, err := client.QueryNodeInfo(nCtx, &QueryNodeInfoReq{Path: "packages/agent.tgz"})
	require.NoError(t, err)
	assert.Equal(t, effectiveProjectID, nodeResp.NodeInfo.ProjectID)
	assert.Equal(t, testRepoName, nodeResp.NodeInfo.RepoName)

	listResp, err := client.ListNode(nCtx, &ListNodeReq{Path: "packages/", PageNum: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, 1, listResp.PageNumber)
	assert.Equal(t, 20, listResp.PageSize)

	err = client.MkDir(nCtx, &MkdirReq{Path: "packages/new/"})
	require.NoError(t, err)

	assertObservedRequests(t, requests, observedRequests(), wantTenantID)
}

func newBKRepoTestClient(t *testing.T, endpoint string) *cli {
	t.Helper()

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	require.NoError(t, err)

	client, err := newClient(&restclient.Capability{
		Name:                 "bkrepo-test",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("bkrepo-test", []string{endpoint}),
		ToleranceLatencyTime: 100 * time.Second,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             bkrepoTestTraceService{},
	}, &Config{
		RepoName:  testRepoName,
		ProjectID: testProjectID,
		Username:  testUsername,
		Password:  testPassword,
	})
	require.NoError(t, err)

	return client
}

func expectedAuthHeader() string {
	credentials := testUsername + ":" + testPassword

	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
}

func expectedQueryValues(values map[string][]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	query := make(map[string]string, len(values))
	for key, value := range values {
		if len(value) == 0 {
			continue
		}

		query[key] = value[0]
	}

	return query
}
