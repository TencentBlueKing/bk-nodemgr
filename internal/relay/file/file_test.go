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

package file

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPkgName = "gse_agent-2-v2.1.6-windows_amd64.tgz"

func md5Hex(data []byte) string {
	sum := md5.Sum(data)

	return hex.EncodeToString(sum[:])
}

// newTestFileManager returns a manager over a temporary cache dir plus a staging dir holding
// testPkgName, mirroring how GSE drops a package before the relay promotes it.
func newTestFileManager(t *testing.T, content []byte) (IFileManager, string) {
	t.Helper()

	stagingDir := stageFile(t, testPkgName, content)

	return newManager(t, Options{}), stagingDir
}

// newManager builds a manager over a fresh cache dir with the given options.
func newManager(t *testing.T, opts Options) IFileManager {
	t.Helper()

	fm, err := NewFileManager(contextx.New(context.Background()), t.TempDir(), opts)
	require.NoError(t, err)

	t.Cleanup(func() { _ = fm.Close() })

	return fm
}

// stageFile writes content into a fresh staging dir, standing in for a completed GSE transfer.
func stageFile(t *testing.T, name string, content []byte) string {
	t.Helper()

	stagingDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, name), content, 0600))

	return stagingDir
}

func readAll(t *testing.T, file fileiface.File) []byte {
	t.Helper()

	nCtx := contextx.New(context.Background())

	reader, err := file.Content(nCtx)
	require.NoError(t, err)

	defer func() { _ = reader.Close() }()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)

	return got
}

func TestStoreFile_PromotesStagedPackage(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("relay package content")
	wantMD5 := md5Hex(content)

	fm, stagingDir := newTestFileManager(t, content)

	info, err := fm.StoreFile(nCtx, stagingDir, testPkgName, wantMD5)
	require.NoError(t, err)
	assert.Equal(t, testPkgName, info.Name)
	assert.Equal(t, wantMD5, info.MD5)
	assert.Equal(t, int64(len(content)), info.Size)

	assert.True(t, fm.FileExists(nCtx, testPkgName, wantMD5), "stored package should be reported as present")

	file, err := fm.GetFile(nCtx, testPkgName)
	require.NoError(t, err)

	reader, err := file.Content(nCtx)
	require.NoError(t, err)

	defer func() { _ = reader.Close() }()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, got, "served content must match what was staged")
}

// TestStoreFile_RejectsMD5Mismatch guards the outage this validation was added for: a staging
// file that is still being written hashes differently, and caching it would make the relay
// serve a truncated package to every node that follows.
func TestStoreFile_RejectsMD5Mismatch(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("truncated package")

	fm, stagingDir := newTestFileManager(t, content)

	otherMD5 := md5Hex([]byte("the complete package"))

	_, err := fm.StoreFile(nCtx, stagingDir, testPkgName, otherMD5)
	require.Error(t, err, "content that does not match the expected md5 must be rejected")

	assert.False(t, fm.FileExists(nCtx, testPkgName, otherMD5), "rejected package must not be indexed")
	assert.False(t, fm.FileExists(nCtx, testPkgName, md5Hex(content)), "rejected package must not be indexed")

	_, err = fm.GetFile(nCtx, testPkgName)
	assert.Error(t, err, "rejected package must not be servable")
}

// TestStoreFile_ReusesCachedContent covers the reuse path: once a package is cached, a later
// installation must be served from the cache even though its own staging copy is gone.
func TestStoreFile_ReusesCachedContent(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("shared package content")
	wantMD5 := md5Hex(content)

	fm, stagingDir := newTestFileManager(t, content)

	_, err := fm.StoreFile(nCtx, stagingDir, testPkgName, wantMD5)
	require.NoError(t, err)

	// Drop the staged copy: a cache hit must not touch the staging directory at all.
	require.NoError(t, os.Remove(filepath.Join(stagingDir, testPkgName)))

	info, err := fm.StoreFile(nCtx, stagingDir, testPkgName, wantMD5)
	require.NoError(t, err, "a cached package must be reused without reading the staging dir")
	assert.Equal(t, wantMD5, info.MD5)
}

// TestStoreFile_ConcurrentInstancesSamePackage reproduces the batch-install scenario that broke
// production: many installations promote the same package at the same time, each from its own
// staging directory. Every one of them must succeed and end up serving identical content.
func TestStoreFile_ConcurrentInstancesSamePackage(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("concurrently promoted package")
	wantMD5 := md5Hex(content)

	fm := newManager(t, Options{})

	const instances = 16

	stagingDirs := make([]string, instances)
	for i := range stagingDirs {
		stagingDirs[i] = stageFile(t, testPkgName, content)
	}

	var wg sync.WaitGroup

	wg.Add(instances)

	for i := range instances {
		go func(idx int) {
			defer wg.Done()

			info, err := fm.StoreFile(nCtx, stagingDirs[idx], testPkgName, wantMD5)
			if err != nil {
				t.Errorf("instance %d failed to store: %v", idx, err)

				return
			}

			assert.Equal(t, wantMD5, info.MD5, "instance %d got unexpected md5", idx)
		}(i)
	}

	wg.Wait()

	file, err := fm.GetFile(nCtx, testPkgName)
	require.NoError(t, err)
	assert.Equal(t, content, readAll(t, file), "the served package must match what every instance staged")
}

// TestStoreFile_ConcurrentReadWhileStaging covers the interleaving that produced the corrupt
// "gzip: invalid header" download: one installation reads the package while another is still
// writing its own staged copy. Isolated staging dirs plus the md5 check must keep whatever the
// cache serves intact.
func TestStoreFile_ConcurrentReadWhileStaging(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte(strings.Repeat("payload", 4096))
	wantMD5 := md5Hex(content)

	fm := newManager(t, Options{})

	firstStaging := stageFile(t, testPkgName, content)
	secondStaging := t.TempDir()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		if _, err := fm.StoreFile(nCtx, firstStaging, testPkgName, wantMD5); err != nil {
			t.Errorf("store from a complete staging dir failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()

		// This installation is still receiving its copy, so its staged file is truncated.
		staged := filepath.Join(secondStaging, testPkgName)
		if err := os.WriteFile(staged, content[:len(content)/2], 0600); err != nil {
			t.Errorf("failed to stage a partial copy: %v", err)

			return
		}

		// A truncated copy must be refused rather than published; whether this call wins the
		// race or finds the cache already warm, it must never surface bad content.
		if _, err := fm.StoreFile(nCtx, secondStaging, testPkgName, wantMD5); err == nil {
			file, getErr := fm.GetFile(nCtx, testPkgName)
			if getErr == nil {
				assert.Equal(t, content, readAll(t, file), "a truncated copy must never be served")
			}
		}
	}()

	wg.Wait()

	file, err := fm.GetFile(nCtx, testPkgName)
	require.NoError(t, err)
	assert.Equal(t, content, readAll(t, file), "the cache must only ever serve the complete package")
}

// TestStoreFile_RejectsEscapingFilename covers the layered confinement of a package name that
// tries to climb out of the staging directory: the cache reduces it to a base name and os.Root
// keeps the open call inside the staging dir. Neither layer may end up caching a foreign file.
func TestStoreFile_RejectsEscapingFilename(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	secret := []byte("must not be readable")
	parent := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(parent, "secret.tgz"), secret, 0600))

	stagingDir := filepath.Join(parent, "staging")
	require.NoError(t, os.MkdirAll(stagingDir, 0750))

	fm := newManager(t, Options{})

	_, err := fm.StoreFile(nCtx, stagingDir, "../secret.tgz", md5Hex(secret))
	require.Error(t, err, "a filename escaping the staging dir must be refused")

	_, err = fm.GetFile(nCtx, "secret.tgz")
	assert.Error(t, err, "nothing from outside the staging dir may be cached")
}

func TestStoreFile_MissingStagedFile(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	fm := newManager(t, Options{})

	_, err := fm.StoreFile(nCtx, t.TempDir(), testPkgName, md5Hex([]byte("anything")))
	require.Error(t, err, "a package that was never staged must not be reported as stored")

	_, err = fm.GetFile(nCtx, testPkgName)
	assert.Error(t, err)
}

func TestStoreFile_RejectsMalformedMD5(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("package with a bad md5 argument")

	fm := newManager(t, Options{})
	stagingDir := stageFile(t, testPkgName, content)

	// An empty or non-hex md5 cannot verify anything, and it must not be used as a cache
	// directory name either.
	for _, badMD5 := range []string{"", "not-a-md5", "../../escape", strings.Repeat("z", 32)} {
		_, err := fm.StoreFile(nCtx, stagingDir, testPkgName, badMD5)
		assert.Error(t, err, "md5(%q) must be refused", badMD5)
	}

	_, err := fm.GetFile(nCtx, testPkgName)
	assert.Error(t, err, "nothing may be cached when the md5 argument is unusable")
}

// TestStoreFile_DistinctPackagesCoexist guards the reuse model: the cache is keyed by content,
// so different packages, and different versions of one package, must not evict each other.
func TestStoreFile_DistinctPackagesCoexist(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	fm := newManager(t, Options{})

	packages := map[string][]byte{
		"gse_agent-2-v2.1.6-windows_amd64.tgz": []byte("agent package"),
		"installer_windows_amd64.exe":          []byte("installer package"),
	}

	for name, content := range packages {
		stagingDir := stageFile(t, name, content)

		_, err := fm.StoreFile(nCtx, stagingDir, name, md5Hex(content))
		require.NoError(t, err)
	}

	for name, content := range packages {
		assert.True(t, fm.FileExists(nCtx, name, md5Hex(content)), "package(%s) should still be cached", name)

		file, err := fm.GetFile(nCtx, name)
		require.NoError(t, err)
		assert.Equal(t, content, readAll(t, file), "package(%s) served wrong content", name)
	}
}

// TestStoreFile_RepublishedPackageReplacesIndex covers a rebuilt release keeping its filename:
// the cache must start serving the new content while the old md5 stops matching.
func TestStoreFile_RepublishedPackageReplacesIndex(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	oldContent := []byte("release build 1")
	newContent := []byte("release build 2")

	fm := newManager(t, Options{})

	_, err := fm.StoreFile(nCtx, stageFile(t, testPkgName, oldContent), testPkgName, md5Hex(oldContent))
	require.NoError(t, err)

	_, err = fm.StoreFile(nCtx, stageFile(t, testPkgName, newContent), testPkgName, md5Hex(newContent))
	require.NoError(t, err)

	assert.True(t, fm.FileExists(nCtx, testPkgName, md5Hex(newContent)), "the republished content must be cached")
	assert.False(t, fm.FileExists(nCtx, testPkgName, md5Hex(oldContent)), "the superseded md5 must stop matching")

	file, err := fm.GetFile(nCtx, testPkgName)
	require.NoError(t, err)
	assert.Equal(t, newContent, readAll(t, file))
}

// TestNewFileManager_AppliesSizeCap proves the relay's cache options really reach the underlying
// cache: without the plumbing a misconfigured relay would silently keep an unbounded cache.
func TestNewFileManager_AppliesSizeCap(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())

	fm := newManager(t, Options{MaxSizeMB: 1})

	oldName, newName := "old-pkg.tgz", "new-pkg.tgz"
	oldContent := []byte(strings.Repeat("o", 600*1024))
	newContent := []byte(strings.Repeat("n", 600*1024))

	_, err := fm.StoreFile(nCtx, stageFile(t, oldName, oldContent), oldName, md5Hex(oldContent))
	require.NoError(t, err)

	// Keep the access timestamps apart so the eviction order is deterministic.
	time.Sleep(10 * time.Millisecond)

	_, err = fm.StoreFile(nCtx, stageFile(t, newName, newContent), newName, md5Hex(newContent))
	require.NoError(t, err)

	assert.False(t, fm.FileExists(nCtx, oldName, md5Hex(oldContent)),
		"the least recently used package should have been evicted by the size cap")
	assert.True(t, fm.FileExists(nCtx, newName, md5Hex(newContent)),
		"the newest package must survive")
}

// TestNewFileManager_RestoresCacheAcrossRestart covers the reuse guarantee a relay restart must
// keep: packages already on disk stay usable instead of being transferred all over again.
func TestNewFileManager_RestoresCacheAcrossRestart(t *testing.T) {
	t.Parallel()

	nCtx := contextx.New(context.Background())
	content := []byte("package surviving a restart")
	wantMD5 := md5Hex(content)

	cacheDir := t.TempDir()

	first, err := NewFileManager(nCtx, cacheDir, Options{RestoreOnStart: true})
	require.NoError(t, err)

	_, err = first.StoreFile(nCtx, stageFile(t, testPkgName, content), testPkgName, wantMD5)
	require.NoError(t, err)
	require.NoError(t, first.Close())

	// A new manager over the same directory stands in for the restarted relay.
	second, err := NewFileManager(nCtx, cacheDir, Options{RestoreOnStart: true})
	require.NoError(t, err)

	t.Cleanup(func() { _ = second.Close() })

	assert.True(t, second.FileExists(nCtx, testPkgName, wantMD5),
		"a restarted relay must still report the cached package as present")

	file, err := second.GetFile(nCtx, testPkgName)
	require.NoError(t, err)
	assert.Equal(t, content, readAll(t, file))
}
