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

package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testStorePkgName = "installer_windows_amd64.exe"

// stubClientMessager records what the handler reports back instead of pushing it through GSE.
type stubClientMessager struct {
	mu       sync.Mutex
	requests map[string][][]byte
}

func newStubClientMessager() *stubClientMessager {
	return &stubClientMessager{requests: make(map[string][][]byte)}
}

func (s *stubClientMessager) ClientPushReq(_ contextx.IContext, callbackURL string, body []byte) <-chan error {
	s.mu.Lock()
	s.requests[callbackURL] = append(s.requests[callbackURL], body)
	s.mu.Unlock()

	ch := make(chan error, 1)
	ch <- nil
	close(ch)

	return ch
}

// bodies returns the payloads reported to callbackURL.
func (s *stubClientMessager) bodies(callbackURL string) [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.requests[callbackURL]
}

func (s *stubClientMessager) RequestCallback(
	_ contextx.IContext, _, _, _ string, _ []byte) ([]byte, int, error) {

	return nil, http.StatusOK, nil
}

func (s *stubClientMessager) Start(_ contextx.IContext) error { return nil }

func (s *stubClientMessager) Stop(_ contextx.IContext) error { return nil }

func (s *stubClientMessager) EventDispatcher() manager.EventDispatcher {
	return manager.NewDefaultEventDispatcher()
}

func testMD5(data []byte) string {
	sum := md5.Sum(data)

	return hex.EncodeToString(sum[:])
}

// newStoreTestHandler builds a handler over temporary staging and cache directories, plus the
// staging dir of one operation instance holding the given packages.
func newStoreTestHandler(t *testing.T, staged map[string][]byte) (*handler, string) {
	t.Helper()

	h, stagingDir, _ := newStoreTestHandlerWithClient(t, staged)

	return h, stagingDir
}

func newStoreTestHandlerWithClient(
	t *testing.T, staged map[string][]byte) (*handler, string, *stubClientMessager) {

	t.Helper()

	nCtx := contextx.New(context.Background())

	fm, err := file.NewFileManager(nCtx, t.TempDir(), file.Options{})
	require.NoError(t, err)

	t.Cleanup(func() { _ = fm.Close() })

	storageFS := newWorkspaceFS(filepath.Join(t.TempDir(), storageTmpDirName))

	stagingDir, err := storageFS.ensureInstanceDir(realOperInstID)
	require.NoError(t, err)

	for name, content := range staged {
		require.NoError(t, os.WriteFile(filepath.Join(stagingDir, name), content, 0600))
	}

	client := newStubClientMessager()

	return &handler{fileManager: fm, storageFS: storageFS, client: client}, stagingDir, client
}

func TestStoreTransferredPkgs_StoresVerifiedPackages(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	agentPkg := []byte("agent package payload")
	installerPkg := []byte("installer payload")

	h, stagingDir := newStoreTestHandler(t, map[string][]byte{
		"gse_agent-2-v2.1.6-windows_amd64.tgz": agentPkg,
		testStorePkgName:                       installerPkg,
	})

	errMsg := h.storeTransferredPkgs(nCtx, stagingDir, []protoRelay.FileInfo{
		{FileName: "gse_agent-2-v2.1.6-windows_amd64.tgz", FileMD5: testMD5(agentPkg)},
		{FileName: testStorePkgName, FileMD5: testMD5(installerPkg)},
	})

	assert.Empty(t, errMsg, "verified packages must be stored without reporting an error")
	assert.True(t, h.fileManager.FileExists(nCtx, "gse_agent-2-v2.1.6-windows_amd64.tgz", testMD5(agentPkg)))
	assert.True(t, h.fileManager.FileExists(nCtx, testStorePkgName, testMD5(installerPkg)))
}

// TestStoreTransferredPkgs_FailsClosedWithoutFileList pins the fail-closed contract: an older
// backend that sends no file list carries no md5, and storing unverified packages is exactly how
// a truncated transfer used to reach the download server.
func TestStoreTransferredPkgs_FailsClosedWithoutFileList(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("unverifiable package")

	h, stagingDir := newStoreTestHandler(t, map[string][]byte{testStorePkgName: content})

	errMsg := h.storeTransferredPkgs(nCtx, stagingDir, nil)

	assert.NotEmpty(t, errMsg, "an empty file list must be reported as an error")
	assert.False(t, h.fileManager.FileExists(nCtx, testStorePkgName, testMD5(content)),
		"nothing may be cached when no md5 was provided")
}

func TestStoreTransferredPkgs_FailsClosedOnBadInput(t *testing.T) {
	t.Parallel()

	content := []byte("package content")

	tests := []struct {
		name  string
		files []protoRelay.FileInfo
	}{
		{
			name:  "empty md5",
			files: []protoRelay.FileInfo{{FileName: testStorePkgName, FileMD5: ""}},
		},
		{
			name:  "md5 does not match the staged content",
			files: []protoRelay.FileInfo{{FileName: testStorePkgName, FileMD5: testMD5([]byte("other"))}},
		},
		{
			name:  "filename escapes the staging dir",
			files: []protoRelay.FileInfo{{FileName: "../" + testStorePkgName, FileMD5: testMD5(content)}},
		},
		{
			name:  "filename is a nested path",
			files: []protoRelay.FileInfo{{FileName: "sub/" + testStorePkgName, FileMD5: testMD5(content)}},
		},
		{
			name:  "package was never staged",
			files: []protoRelay.FileInfo{{FileName: "absent.tgz", FileMD5: testMD5(content)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nCtx := contextx.New(context.Background())
			h, stagingDir := newStoreTestHandler(t, map[string][]byte{testStorePkgName: content})

			errMsg := h.storeTransferredPkgs(nCtx, stagingDir, tt.files)

			assert.NotEmpty(t, errMsg, "the failure must be reported back to the backend")
			assert.False(t, h.fileManager.FileExists(nCtx, testStorePkgName, testMD5(content)),
				"a rejected transfer must leave the cache untouched")
		})
	}
}

// TestStoreTransferredPkgs_StopsAtFirstFailure documents that a batch is all-or-nothing from the
// backend's point of view: it retries the whole action, so continuing past a bad package would
// only cache a partial set.
func TestStoreTransferredPkgs_StopsAtFirstFailure(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	goodPkg := []byte("good package")
	h, stagingDir := newStoreTestHandler(t, map[string][]byte{
		testStorePkgName: goodPkg,
	})

	errMsg := h.storeTransferredPkgs(nCtx, stagingDir, []protoRelay.FileInfo{
		{FileName: "absent.tgz", FileMD5: testMD5([]byte("absent"))},
		{FileName: testStorePkgName, FileMD5: testMD5(goodPkg)},
	})

	assert.NotEmpty(t, errMsg)
	assert.False(t, h.fileManager.FileExists(nCtx, testStorePkgName, testMD5(goodPkg)),
		"packages after the first failure must not be stored")
}

// TestCheckPkgStats_StagingDirCreatedOnlyWhenNeeded pins when the staging directory appears.
// The backend only sends the store event, which is what removes the directory, if at least one
// package has to be transferred. Creating it on a pure cache hit would therefore leave an empty
// directory behind for every host in a batch install until the orphan gc caught up.
func TestCheckPkgStats_StagingDirCreatedOnlyWhenNeeded(t *testing.T) {
	nCtx := contextx.New(context.Background())
	content := []byte("already cached package")

	h, stagingDir, client := newStoreTestHandlerWithClient(t, map[string][]byte{testStorePkgName: content})

	// Prime the cache so the package counts as complete, then drop the staging dir that the
	// helper created so we can observe whether CheckPkgStats recreates it.
	_, err := h.fileManager.StoreFile(nCtx, stagingDir, testStorePkgName, testMD5(content))
	require.NoError(t, err)
	require.NoError(t, h.storageFS.removeInstanceDir(realOperInstID))

	cached := protoRelay.FileInfo{FileName: testStorePkgName, FileMD5: testMD5(content)}
	missing := protoRelay.FileInfo{FileName: "not-cached.tgz", FileMD5: testMD5([]byte("missing"))}

	h.CheckPkgStats(nCtx, mustMarshal(t, protoRelay.CheckPkgStateReq{
		ActionName: "ensure_pkg_to_relay",
		OperInstID: realOperInstID,
		FileList:   []protoRelay.FileInfo{cached},
	}))

	_, err = os.Stat(stagingDir)
	assert.True(t, os.IsNotExist(err), "no staging dir should be created when every package is cached")

	// The path is still reported: the backend rejects an empty storage dir outright, and it
	// only needs a usable path once it actually transfers something.
	reports := client.bodies(reportRelayFileStateURL)
	require.Len(t, reports, 1)

	var cacheHit reportRelayFileState
	require.NoError(t, json.Unmarshal(reports[0], &cacheHit))
	assert.Equal(t, stagingDir, cacheHit.StorageTmpDir)
	require.Len(t, cacheHit.FileState, 1)
	assert.Equal(t, string(relayconstant.RelayReportPkgComplete), cacheHit.FileState[0].FileStatus)

	h.CheckPkgStats(nCtx, mustMarshal(t, protoRelay.CheckPkgStateReq{
		ActionName: "ensure_pkg_to_relay",
		OperInstID: realOperInstID,
		FileList:   []protoRelay.FileInfo{cached, missing},
	}))

	info, err := os.Stat(stagingDir)
	require.NoError(t, err, "a staging dir must exist once a package has to be transferred")
	assert.True(t, info.IsDir())

	reports = client.bodies(reportRelayFileStateURL)
	require.Len(t, reports, 2)

	var withMissing reportRelayFileState
	require.NoError(t, json.Unmarshal(reports[1], &withMissing))
	require.Len(t, withMissing.FileState, 2)
	assert.Equal(t, string(relayconstant.RelayReportPkgInComplete), withMissing.FileState[1].FileStatus,
		"the uncached package must be reported as incomplete so the backend transfers it")
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}

func TestCollectOrphanStagingDirs(t *testing.T) {

	nCtx := contextx.New(context.Background())

	rootDir := t.TempDir()
	h := &handler{storageFS: newWorkspaceFS(rootDir)}

	// An installation that died before StoragePkg could clean up.
	abandoned := filepath.Join(rootDir, "oper-inst_abandoned")
	require.NoError(t, os.MkdirAll(abandoned, storageTmpDirMode))
	require.NoError(t, os.WriteFile(filepath.Join(abandoned, testStorePkgName), []byte("leaked"), 0600))

	staleTime := time.Now().Add(-stagingOrphanAge - time.Hour)
	require.NoError(t, os.Chtimes(abandoned, staleTime, staleTime))

	// An installation that is still running.
	active := filepath.Join(rootDir, "oper-inst_active")
	require.NoError(t, os.MkdirAll(active, storageTmpDirMode))

	h.collectOrphanStagingDirs(nCtx)

	_, err := os.Stat(abandoned)
	assert.True(t, os.IsNotExist(err), "an abandoned staging dir must be reclaimed")

	_, err = os.Stat(active)
	assert.NoError(t, err, "a staging dir still in use must be left alone")
}

func TestCollectOrphanStagingDirs_MissingRoot(t *testing.T) {
	t.Parallel()

	// The GC also runs at startup, before any installation created the staging root.
	h := &handler{storageFS: newWorkspaceFS(filepath.Join(t.TempDir(), "never-created"))}

	h.collectOrphanStagingDirs(contextx.New(context.Background()))
}
