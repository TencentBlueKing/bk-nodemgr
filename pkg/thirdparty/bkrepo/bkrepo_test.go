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
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
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
		{
			method: http.MethodPost,
			path:   "/repository/api/node/copy",
			body: `{"srcProjectId":"` + effectiveProjectID + `","srcRepoName":"` + testRepoName +
				`","srcFullPath":"/system/release/plugin/source.tgz","destProjectId":"` + effectiveProjectID +
				`","destRepoName":"` + testRepoName +
				`","destFullPath":"/tenant-a/release/plugin","overwrite":false}`,
			responseBody: `{"code":0,"message":"OK","data":{}}`,
		},
		{
			method: http.MethodDelete,
			path:   "/repository/api/node/delete/" + effectiveProjectID + "/" + testRepoName + "/packages/obsolete.tgz",
			responseBody: `{"code":0,"message":"OK","data":{` +
				`"deletedNumber":1,"deletedSize":128,"deletedTime":"2026-09-09T00:00:00Z"}}`,
		},
	}

	server, observedRequests := newBKRepoRequestServer(requests)
	defer server.Close()

	client := newBKRepoTestClient(t, server.URL)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testContextTenant))

	downloadResp, err := client.DownloadFile(nCtx, &DownloadFileReq{Path: "packages/agent.tgz"})
	require.NoError(t, err)
	downloadData, readErr := io.ReadAll(downloadResp.Data)
	closeErr := downloadResp.Data.Close()
	require.NoError(t, readErr)
	require.NoError(t, closeErr)
	assert.Equal(t, []byte("agent-content"), downloadData)

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

	handler := &Handler{cli: client}
	sourceGroup := &FileGroup{
		info: NodeInfo{
			ProjectID: effectiveProjectID,
			RepoName:  testRepoName,
			FullPath:  "/system/release/plugin",
		},
		handler: handler,
	}
	destinationGroup := &FileGroup{
		info: NodeInfo{
			ProjectID: effectiveProjectID,
			RepoName:  testRepoName,
			FullPath:  "/tenant-a/release/plugin",
		},
		handler: handler,
	}
	err = sourceGroup.Copy(nCtx, "source.tgz", destinationGroup, ".", false)
	require.NoError(t, err)

	deleteNodeResp, err := client.DeleteNode(nCtx, &DeleteNodeReq{Path: "packages/obsolete.tgz"})
	require.NoError(t, err)
	assert.Equal(t, 1, deleteNodeResp.DeletedNumber)
	assert.Equal(t, int64(128), deleteNodeResp.DeletedSize)
	assert.Equal(t, "2026-09-09T00:00:00Z", deleteNodeResp.DeletedTime)

	assertObservedRequests(t, requests, observedRequests(), wantTenantID)
}

func TestCopyAndDeleteNodeRejectBusinessFailure(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() {
		tenant.SetMode(originalMode)
	})

	requests := []bkrepoRequestExpectation{
		{
			method: http.MethodPost,
			path:   "/repository/api/node/copy",
			body: `{"srcProjectId":"blueking","srcRepoName":"agent",` +
				`"srcFullPath":"packages/source.tgz","destProjectId":"blueking",` +
				`"destRepoName":"agent","destFullPath":"packages/destination.tgz","overwrite":false}`,
			responseBody: `{"code":1,"message":"copy rejected"}`,
		},
		{
			method:       http.MethodDelete,
			path:         "/repository/api/node/delete/blueking/agent/packages/obsolete.tgz",
			responseBody: `{"code":2,"message":"delete rejected"}`,
		},
	}

	server, observedRequests := newBKRepoRequestServer(requests)
	defer server.Close()

	client := newBKRepoTestClient(t, server.URL)
	nCtx := contextx.New(context.Background())

	_, err := client.CopyNode(nCtx, &CopyNodeReq{
		SrcProjectID:  testProjectID,
		SrcRepoName:   testRepoName,
		SrcFullPath:   "packages/source.tgz",
		DestProjectID: testProjectID,
		DestRepoName:  testRepoName,
		DestFullPath:  "packages/destination.tgz",
	})
	require.ErrorContains(t, err, "copy rejected")

	_, err = client.DeleteNode(nCtx, &DeleteNodeReq{Path: "packages/obsolete.tgz"})
	require.ErrorContains(t, err, "delete rejected")

	assertObservedRequests(t, requests, observedRequests(), "")
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

func TestFileGroupDirectoryPaths(t *testing.T) {
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}}
	for _, name := range []string{"", "/absolute", "../escape", "nested/../../escape", `nested\file`, `C:\file`, "name\x00part"} {
		t.Run(name, func(t *testing.T) {
			_, err := group.IsDir(contextx.Background(), name)
			require.Error(t, err)
			_, err = group.GetSubGroup(contextx.Background(), name)
			require.Error(t, err)
			_, err = group.EnsureSubGroup(contextx.Background(), name)
			require.Error(t, err)
		})
	}
}

func TestFileGroupCopyRejectsInvalidNativePathsAndNilDestination(t *testing.T) {
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{}}
	var nilDestination *FileGroup
	err := group.Copy(contextx.Background(), "file", nilDestination, ".", false)
	require.Error(t, err)

	for _, name := range []string{"/absolute", "../escape"} {
		err = group.Copy(contextx.Background(), name, group, ".", false)
		require.Error(t, err)
	}
}

func TestFileGroupCopyUsesCopyNode(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	requests := []bkrepoRequestExpectation{{
		method:       http.MethodPost,
		path:         "/repository/api/node/copy",
		body:         `{"srcProjectId":"blueking","srcRepoName":"agent","srcFullPath":"/repo/source.txt","destProjectId":"blueking","destRepoName":"agent","destFullPath":"/repo/target","overwrite":false}`,
		responseBody: `{"code":0,"data":{}}`,
	}}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	handler := &Handler{cli: newBKRepoTestClient(t, server.URL)}
	source := &FileGroup{info: NodeInfo{ProjectID: testProjectID, RepoName: testRepoName, FullPath: "/repo"}, handler: handler}
	destination := &FileGroup{info: NodeInfo{ProjectID: testProjectID, RepoName: testRepoName, FullPath: "/repo"}, handler: handler}

	require.NoError(t, source.Copy(contextx.Background(), "source.txt", destination, "target", false))
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupNativeCopyPreservesPathGrammar(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	for _, tc := range []struct {
		name string
		src  string
		body string
	}{
		{
			name: "backslash",
			src:  `nested\file`,
			body: `{"srcProjectId":"blueking","srcRepoName":"agent","srcFullPath":"/repo/nested\\file",` +
				`"destProjectId":"blueking","destRepoName":"agent","destFullPath":"/repo/target","overwrite":false}`,
		},
		{
			name: "colon",
			src:  "C:source",
			body: `{"srcProjectId":"blueking","srcRepoName":"agent","srcFullPath":"/repo/C:source",` +
				`"destProjectId":"blueking","destRepoName":"agent","destFullPath":"/repo/target","overwrite":false}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := []bkrepoRequestExpectation{{
				method: http.MethodPost, path: "/repository/api/node/copy", body: tc.body,
				responseBody: `{"code":0,"data":{}}`,
			}}
			server, observed := newBKRepoRequestServer(requests)
			defer server.Close()
			handler := &Handler{cli: newBKRepoTestClient(t, server.URL)}
			group := &FileGroup{info: NodeInfo{ProjectID: testProjectID, RepoName: testRepoName, FullPath: "/repo"}, handler: handler}

			require.NoError(t, group.Copy(contextx.Background(), tc.src, group, "target", false))
			assertObservedRequests(t, requests, observed(), "")
		})
	}
}

func TestFileGroupIsDirMissing(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })

	requests := []bkrepoRequestExpectation{{
		method:       http.MethodGet,
		path:         "/repository/api/node/detail/blueking/agent/repo/missing",
		responseBody: `{"code":251010,"message":"missing"}`,
	}}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	isDir, err := group.IsDir(contextx.Background(), "missing")
	require.False(t, isDir)
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorIs(t, err, errNodeNotFound)
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupGetSubGroupMissing(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	requests := []bkrepoRequestExpectation{{
		method:       http.MethodGet,
		path:         "/repository/api/node/detail/blueking/agent/repo/missing",
		responseBody: `{"code":251010,"message":"missing"}`,
	}}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	_, err := group.GetSubGroup(contextx.Background(), "missing")
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorIs(t, err, errNodeNotFound)
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupEnsureSubGroupCreatesParents(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })

	detail := func(name string) bkrepoRequestExpectation {
		return bkrepoRequestExpectation{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/blueking/agent/repo/" + name,
			responseBody: `{"code":0,"data":{"nodeInfo":{"name":"` + path.Base(name) + `","fullPath":"/repo/` + name + `","folder":true}}}`,
		}
	}
	missing := func(name string) bkrepoRequestExpectation {
		request := detail(name)
		request.responseBody = `{"code":251010,"message":"missing"}`
		return request
	}
	mkdir := func(name string) bkrepoRequestExpectation {
		return bkrepoRequestExpectation{
			method:       http.MethodPost,
			path:         "/repository/api/node/mkdir/blueking/agent/repo/" + name,
			responseBody: `{"code":0,"data":{}}`,
		}
	}
	requests := []bkrepoRequestExpectation{
		missing("a/b"),
		{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/blueking/agent/repo",
			responseBody: `{"code":0,"data":{"nodeInfo":{"name":"repo","fullPath":"/repo","folder":true}}}`,
		},
		missing("a"), mkdir("a"), detail("a"),
		missing("a/b"), mkdir("a/b"), detail("a/b"),
	}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	subGroup, err := group.EnsureSubGroup(contextx.Background(), "a/b")
	require.NoError(t, err)
	require.Equal(t, "b", subGroup.Name())
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupEnsureSubGroupUsesExistingDeepDirectory(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	requests := []bkrepoRequestExpectation{{
		method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo/a/b",
		responseBody: `{"code":0,"data":{"nodeInfo":{"name":"b","fullPath":"/repo/a/b","folder":true}}}`,
	}}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	result, err := group.EnsureSubGroup(contextx.Background(), "a/b")
	require.NoError(t, err)
	require.Equal(t, "b", result.Name())
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupEnsureSubGroupPropagatesLookupError(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	requests := []bkrepoRequestExpectation{{
		method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo/a/b",
		responseBody: `{"code":777,"message":"lookup rejected"}`,
	}}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	result, err := group.EnsureSubGroup(contextx.Background(), "a/b")
	require.Nil(t, result)
	require.ErrorContains(t, err, "lookup rejected")
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupRootDirectoryUsesLiveState(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	requests := []bkrepoRequestExpectation{
		{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/blueking/agent/repo",
			responseBody: `{"code":251010,"message":"missing"}`,
		},
		{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/blueking/agent/repo",
			responseBody: `{"code":251010,"message":"missing"}`,
		},
		{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/blueking/agent/repo",
			responseBody: `{"code":251010,"message":"missing"}`,
		},
		{
			method:       http.MethodPost,
			path:         "/repository/api/node/mkdir/blueking/agent/repo",
			responseBody: `{"code":0,"data":{}}`,
		},
		{
			method:       http.MethodGet,
			path:         "/repository/api/node/detail/blueking/agent/repo",
			responseBody: `{"code":0,"data":{"nodeInfo":{"name":"repo","fullPath":"/repo","folder":true}}}`,
		},
	}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	_, err := group.GetSubGroup(contextx.Background(), ".")
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorIs(t, err, errNodeNotFound)
	created, err := group.EnsureSubGroup(contextx.Background(), ".")
	require.NoError(t, err)
	require.Equal(t, "repo", created.Name())
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupRootDirectoryRejectsFile(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	response := `{"code":0,"data":{"nodeInfo":{"name":"repo","fullPath":"/repo","folder":false}}}`
	requests := []bkrepoRequestExpectation{
		{method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo", responseBody: response},
		{method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo", responseBody: response},
	}
	server, observed := newBKRepoRequestServer(requests)
	defer server.Close()
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

	_, err := group.GetSubGroup(contextx.Background(), ".")
	require.Error(t, err)
	_, err = group.EnsureSubGroup(contextx.Background(), ".")
	require.Error(t, err)
	assertObservedRequests(t, requests, observed(), "")
}

func TestFileGroupEnsureSubGroupAcceptsMkdirRaceOnlyForExistingDirectory(t *testing.T) {
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })
	for _, tc := range []struct {
		name         string
		secondDetail string
		wantError    bool
	}{
		{name: "created concurrently", secondDetail: `{"code":0,"data":{"nodeInfo":{"name":"a","fullPath":"/repo/a","folder":true}}}`},
		{name: "still missing", secondDetail: `{"code":251010,"message":"missing"}`, wantError: true},
		{name: "replaced by file", secondDetail: `{"code":0,"data":{"nodeInfo":{"name":"a","fullPath":"/repo/a","folder":false}}}`, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := []bkrepoRequestExpectation{
				{method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo/a", responseBody: `{"code":251010,"message":"missing"}`},
				{method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo", responseBody: `{"code":0,"data":{"nodeInfo":{"name":"repo","fullPath":"/repo","folder":true}}}`},
				{method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo/a", responseBody: `{"code":251010,"message":"missing"}`},
				{method: http.MethodPost, path: "/repository/api/node/mkdir/blueking/agent/repo/a", responseBody: `{"code":777,"message":"mkdir rejected"}`},
				{method: http.MethodGet, path: "/repository/api/node/detail/blueking/agent/repo/a", responseBody: tc.secondDetail},
			}
			server, observed := newBKRepoRequestServer(requests)
			defer server.Close()
			group := &FileGroup{info: NodeInfo{FullPath: "/repo"}, handler: &Handler{cli: newBKRepoTestClient(t, server.URL)}}

			created, err := group.EnsureSubGroup(contextx.Background(), "a")
			if tc.wantError {
				require.ErrorContains(t, err, "mkdir rejected")
				require.Nil(t, created)
			} else {
				require.NoError(t, err)
				require.Equal(t, "a", created.Name())
			}
			assertObservedRequests(t, requests, observed(), "")
		})
	}
}

func TestFileGroupDirectoryMethodsRejectNilContext(t *testing.T) {
	group := &FileGroup{info: NodeInfo{FullPath: "/repo"}}
	_, err := group.IsDir(nil, ".")
	require.ErrorIs(t, err, errInvalidContext)
	_, err = group.GetSubGroup(nil, ".")
	require.ErrorIs(t, err, errInvalidContext)
	_, err = group.EnsureSubGroup(nil, ".")
	require.ErrorIs(t, err, errInvalidContext)
}

type transferRepoNode struct {
	directory bool
	content   string
}

type transferRepo struct {
	mu    sync.Mutex
	nodes map[string]transferRepoNode
}

func newTransferRepo(t *testing.T) (*FileGroup, *transferRepo) {
	t.Helper()
	originalMode := tenant.GetMode()
	tenant.SetMode(tenant.ModeSingle)
	t.Cleanup(func() { tenant.SetMode(originalMode) })

	repo := &transferRepo{nodes: map[string]transferRepoNode{"/repo": {directory: true}}}
	server := httptest.NewServer(http.HandlerFunc(repo.serveHTTP))
	t.Cleanup(server.Close)
	group := &FileGroup{
		info:    NodeInfo{ProjectID: testProjectID, RepoName: testRepoName, FullPath: "/repo", Folder: true},
		handler: &Handler{cli: newBKRepoTestClient(t, server.URL)},
	}
	return group, repo
}

func (repo *transferRepo) addDirectory(name string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.nodes[path.Clean(name)] = transferRepoNode{directory: true}
}

func (repo *transferRepo) addFile(name, content string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.nodes[path.Clean(name)] = transferRepoNode{content: content}
}

func (repo *transferRepo) node(name string) (transferRepoNode, bool) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	node, ok := repo.nodes[path.Clean(name)]
	return node, ok
}

func (repo *transferRepo) serveHTTP(rw http.ResponseWriter, req *http.Request) {
	const detailPrefix = "/repository/api/node/detail/blueking/agent/"
	const pagePrefix = "/repository/api/node/page/blueking/agent/"
	const mkdirPrefix = "/repository/api/node/mkdir/blueking/agent/"
	const contentPrefix = "/generic/blueking/agent/"

	switch {
	case req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, detailPrefix):
		repo.serveDetail(rw, "/"+strings.TrimPrefix(req.URL.Path, detailPrefix))
	case req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, pagePrefix):
		repo.servePage(rw, "/"+strings.TrimPrefix(req.URL.Path, pagePrefix))
	case req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, mkdirPrefix):
		repo.serveMkdir(rw, "/"+strings.TrimPrefix(req.URL.Path, mkdirPrefix))
	case req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, contentPrefix):
		repo.serveDownload(rw, "/"+strings.TrimPrefix(req.URL.Path, contentPrefix))
	case req.Method == http.MethodPut && strings.HasPrefix(req.URL.Path, contentPrefix):
		repo.serveUpload(rw, req, "/"+strings.TrimPrefix(req.URL.Path, contentPrefix))
	default:
		http.Error(rw, "unexpected BKRepo request", http.StatusNotFound)
	}
}

func writeTransferRepoResponse(rw http.ResponseWriter, code int, data any) {
	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(map[string]any{"code": code, "data": data})
}

func (repo *transferRepo) serveDetail(rw http.ResponseWriter, name string) {
	node, ok := repo.node(name)
	if !ok {
		writeTransferRepoResponse(rw, errNumNodeNotFound, nil)
		return
	}
	fullPath := path.Clean(name)
	writeTransferRepoResponse(rw, CodeOK, map[string]any{"nodeInfo": NodeInfo{
		Name: path.Base(fullPath), FullPath: fullPath, Folder: node.directory,
		ProjectID: testProjectID, RepoName: testRepoName,
	}})
}

func (repo *transferRepo) servePage(rw http.ResponseWriter, name string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	parent := path.Clean(name)
	records := make([]NodeRecord, 0)
	for fullPath, node := range repo.nodes {
		if fullPath == parent || path.Dir(fullPath) != parent {
			continue
		}
		records = append(records, NodeRecord{Name: path.Base(fullPath), FullPath: fullPath, Folder: node.directory})
	}
	slices.SortFunc(records, func(a, b NodeRecord) int { return cmp.Compare(a.Name, b.Name) })
	writeTransferRepoResponse(rw, CodeOK, ListNodeResp{Records: records})
}

func (repo *transferRepo) serveMkdir(rw http.ResponseWriter, name string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	fullPath := path.Clean(name)
	parent, ok := repo.nodes[path.Dir(fullPath)]
	if !ok || !parent.directory {
		writeTransferRepoResponse(rw, 1, nil)
		return
	}
	repo.nodes[fullPath] = transferRepoNode{directory: true}
	writeTransferRepoResponse(rw, CodeOK, map[string]any{})
}

func (repo *transferRepo) serveDownload(rw http.ResponseWriter, name string) {
	node, ok := repo.node(name)
	if !ok || node.directory {
		http.Error(rw, "file not found", http.StatusNotFound)
		return
	}
	_, _ = io.WriteString(rw, node.content)
}

func (repo *transferRepo) serveUpload(rw http.ResponseWriter, req *http.Request, name string) {
	content, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()
	fullPath := path.Clean(name)
	parent, ok := repo.nodes[path.Dir(fullPath)]
	if !ok || !parent.directory {
		writeTransferRepoResponse(rw, 1, nil)
		return
	}
	if _, exists := repo.nodes[fullPath]; exists && req.Header.Get("X-BKREPO-OVERWRITE") != "true" {
		writeTransferRepoResponse(rw, 1, nil)
		return
	}
	repo.nodes[fullPath] = transferRepoNode{content: string(content)}
	writeTransferRepoResponse(rw, CodeOK, map[string]any{})
}

func TestCrossBackendCopyFilesAndOverwrite(t *testing.T) {
	repoGroup, repo := newTransferRepo(t)
	repo.addFile("/repo/remote.txt", "remote-v1")
	localPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local-v1"), 0600))
	localGroup, err := local.NewLocalDir(localPath)
	require.NoError(t, err)
	ctx := contextx.Background()

	require.NoError(t, localGroup.Copy(ctx, "local.txt", repoGroup, ".", false))
	got, ok := repo.node("/repo/local.txt")
	require.True(t, ok)
	require.Equal(t, "local-v1", got.content)
	require.NoError(t, repoGroup.Copy(ctx, "remote.txt", localGroup, ".", false))
	localData, err := os.ReadFile(filepath.Join(localPath, "remote.txt"))
	require.NoError(t, err)
	require.Equal(t, "remote-v1", string(localData))

	require.NoError(t, os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local-v2"), 0600))
	require.Error(t, localGroup.Copy(ctx, "local.txt", repoGroup, ".", false))
	got, _ = repo.node("/repo/local.txt")
	require.Equal(t, "local-v1", got.content)
	require.NoError(t, localGroup.Copy(ctx, "local.txt", repoGroup, ".", true))
	got, _ = repo.node("/repo/local.txt")
	require.Equal(t, "local-v2", got.content)

	repo.addFile("/repo/remote.txt", "remote-v2")
	require.Error(t, repoGroup.Copy(ctx, "remote.txt", localGroup, ".", false))
	localData, err = os.ReadFile(filepath.Join(localPath, "remote.txt"))
	require.NoError(t, err)
	require.Equal(t, "remote-v1", string(localData))
	require.NoError(t, repoGroup.Copy(ctx, "remote.txt", localGroup, ".", true))
	localData, err = os.ReadFile(filepath.Join(localPath, "remote.txt"))
	require.NoError(t, err)
	require.Equal(t, "remote-v2", string(localData))
}

func TestCrossBackendCopyRejectsInvalidPathsWithoutWriting(t *testing.T) {
	repoGroup, repo := newTransferRepo(t)
	repo.addFile("/repo/remote.txt", "remote")
	repo.addFile("/repo/nested\\file", "backslash")
	repo.addFile("/outside.txt", "outside")
	testRoot := t.TempDir()
	localPath := filepath.Join(testRoot, "local")
	require.NoError(t, os.Mkdir(localPath, 0700))
	outsidePath := filepath.Join(testRoot, "outside.txt")
	require.NoError(t, os.WriteFile(outsidePath, []byte("outside"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local"), 0600))
	require.NoError(t, os.Mkdir(filepath.Join(localPath, "nested"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "nested\\file"), []byte("backslash"), 0600))
	localGroup, err := local.NewLocalDir(localPath)
	require.NoError(t, err)
	ctx := contextx.Background()
	invalid := []struct {
		name string
		path string
	}{
		{name: "empty", path: ""},
		{name: "absolute", path: outsidePath},
		{name: "backslash", path: `nested\file`},
		{name: "NUL", path: "name\x00part"},
		{name: "escape", path: "../outside.txt"},
		{name: "nested escape", path: "nested/../../outside.txt"},
	}
	repo.mu.Lock()
	beforeRepo := maps.Clone(repo.nodes)
	repo.mu.Unlock()
	beforeEntries, err := os.ReadDir(localPath)
	require.NoError(t, err)
	beforeNames := make([]string, 0, len(beforeEntries))
	for _, entry := range beforeEntries {
		beforeNames = append(beforeNames, entry.Name())
	}
	beforeNestedEntries, err := os.ReadDir(filepath.Join(localPath, "nested"))
	require.NoError(t, err)

	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			require.Error(t, repoGroup.Copy(ctx, tc.path, localGroup, "copied.txt", true))
			require.Error(t, repoGroup.Copy(ctx, "remote.txt", localGroup, tc.path, true))
			require.Error(t, localGroup.Copy(ctx, tc.path, repoGroup, "new.txt", true))
			require.Error(t, localGroup.Copy(ctx, "local.txt", repoGroup, tc.path, true))
		})
	}
	for _, source := range []string{".", "nested/.."} {
		t.Run("root source/"+source, func(t *testing.T) {
			require.Error(t, repoGroup.Copy(ctx, source, localGroup, "copied.txt", true))
			require.Error(t, localGroup.Copy(ctx, source, repoGroup, "new.txt", true))
		})
	}
	require.Error(t, repoGroup.Copy(ctx, "missing.txt", localGroup, "copied.txt", true))
	require.Error(t, localGroup.Copy(ctx, "missing.txt", repoGroup, "new.txt", true))
	repo.mu.Lock()
	afterRepo := maps.Clone(repo.nodes)
	repo.mu.Unlock()
	require.Equal(t, beforeRepo, afterRepo)
	afterEntries, err := os.ReadDir(localPath)
	require.NoError(t, err)
	afterNames := make([]string, 0, len(afterEntries))
	for _, entry := range afterEntries {
		afterNames = append(afterNames, entry.Name())
	}
	require.Equal(t, beforeNames, afterNames)
	afterNestedEntries, err := os.ReadDir(filepath.Join(localPath, "nested"))
	require.NoError(t, err)
	require.Equal(t, beforeNestedEntries, afterNestedEntries)
	for filePath, want := range map[string]string{
		filepath.Join(localPath, "local.txt"):    "local",
		filepath.Join(localPath, "nested\\file"): "backslash",
		outsidePath:                              "outside",
	} {
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, want, string(data))
	}
}

func TestCrossBackendCopyRejectsLocalSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links require platform-specific privileges on Windows")
	}
	repoGroup, repo := newTransferRepo(t)
	repo.addFile("/repo/source.txt", "replacement")
	repo.addDirectory("/repo/tree")
	repo.addFile("/repo/tree/item.txt", "replacement")
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	require.NoError(t, os.WriteFile(outside, []byte("original"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "source.txt"), []byte("source"), 0600))
	require.NoError(t, os.Symlink("source.txt", filepath.Join(root, "alias.txt")))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "source-tree"), 0700))
	require.NoError(t, os.Symlink("../source.txt", filepath.Join(root, "source-tree", "alias.txt")))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "target", "tree"), 0700))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "target", "source.txt")))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "target", "tree", "item.txt")))
	localGroup, err := local.NewLocalDir(root)
	require.NoError(t, err)
	ctx := contextx.Background()

	require.ErrorContains(t, localGroup.Copy(ctx, "alias.txt", repoGroup, "copied.txt", false), "unsupported file type")
	require.ErrorContains(t, localGroup.Copy(ctx, "source-tree", repoGroup, ".", false), "unsupported file type")
	require.ErrorContains(t, repoGroup.Copy(ctx, "source.txt", localGroup, "target", true), "unsupported file type")
	require.ErrorContains(t, repoGroup.Copy(ctx, "tree", localGroup, "target", true), "unsupported file type")
	content, err := os.ReadFile(outside)
	require.NoError(t, err)
	require.Equal(t, "original", string(content))
}

func TestCrossBackendCopyRejectsTypedNilLocalDestination(t *testing.T) {
	repoGroup, repo := newTransferRepo(t)
	repo.addFile("/repo/source.txt", "source")
	var destination *local.LocalDir

	err := repoGroup.Copy(contextx.Background(), "source.txt", destination, "target.txt", false)
	require.ErrorContains(t, err, "file group cannot be nil")
}

func TestCrossBackendCopyNormalizesRelativePaths(t *testing.T) {
	repoGroup, repo := newTransferRepo(t)
	repo.addFile("/repo/remote.txt", "remote")
	repo.addDirectory("/repo/dir")
	localPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "foo.txt"), []byte("local"), 0600))
	require.NoError(t, os.Mkdir(filepath.Join(localPath, "target"), 0700))
	localGroup, err := local.NewLocalDir(localPath)
	require.NoError(t, err)
	ctx := contextx.Background()

	require.NoError(t, localGroup.Copy(ctx, "./foo.txt", repoGroup, "dir/.", false))
	got, exists := repo.node("/repo/dir/foo.txt")
	require.True(t, exists)
	require.Equal(t, "local", got.content)
	require.NoError(t, repoGroup.Copy(ctx, "nested/../remote.txt", localGroup, "target/.", false))
	localData, err := os.ReadFile(filepath.Join(localPath, "target", "remote.txt"))
	require.NoError(t, err)
	require.Equal(t, "remote", string(localData))
}

func TestCrossBackendCopyDirectoriesMergeAndKeepEmpty(t *testing.T) {
	repoGroup, repo := newTransferRepo(t)
	repo.addDirectory("/repo/target")
	repo.addDirectory("/repo/target/tree")
	repo.addDirectory("/repo/target/tree/nested")
	repo.addFile("/repo/target/tree/nested/item.txt", "old")
	repo.addFile("/repo/target/tree/keep.txt", "keep")
	repo.addDirectory("/repo/remote")
	repo.addDirectory("/repo/remote/nested")
	repo.addDirectory("/repo/remote/empty")
	repo.addFile("/repo/remote/nested/item.txt", "remote")

	localPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(localPath, "tree", "nested"), 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(localPath, "tree", "empty"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "tree", "nested", "item.txt"), []byte("local"), 0600))
	require.NoError(t, os.MkdirAll(filepath.Join(localPath, "target", "remote", "nested"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "target", "remote", "nested", "item.txt"), []byte("old"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "target", "remote", "keep.txt"), []byte("keep"), 0600))
	localGroup, err := local.NewLocalDir(localPath)
	require.NoError(t, err)
	ctx := contextx.Background()

	require.NoError(t, localGroup.Copy(ctx, "tree", repoGroup, "target", true))
	got, ok := repo.node("/repo/target/tree/nested/item.txt")
	require.True(t, ok)
	require.Equal(t, "local", got.content)
	got, ok = repo.node("/repo/target/tree/keep.txt")
	require.True(t, ok)
	require.Equal(t, "keep", got.content)
	got, ok = repo.node("/repo/target/tree/empty")
	require.True(t, ok)
	require.True(t, got.directory)

	require.NoError(t, repoGroup.Copy(ctx, "remote", localGroup, "target", true))
	localData, err := os.ReadFile(filepath.Join(localPath, "target", "remote", "nested", "item.txt"))
	require.NoError(t, err)
	require.Equal(t, "remote", string(localData))
	localData, err = os.ReadFile(filepath.Join(localPath, "target", "remote", "keep.txt"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(localData))
	info, err := os.Stat(filepath.Join(localPath, "target", "remote", "empty"))
	require.NoError(t, err)
	require.True(t, info.IsDir())
}
