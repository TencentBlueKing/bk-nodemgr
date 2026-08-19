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

package filecache_test

import (
	"context"
	"crypto/md5" // nolint: gosec
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
)

// testCtx returns a minimal context for testing.
func testCtx() contextx.IContext {
	return contextx.New(context.Background())
}

func md5sum(data string) string {
	h := md5.New() // nolint: gosec
	_, _ = h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

func makeFetchFn(data string) func(nCtx contextx.IContext) (io.ReadCloser, error) {
	return func(_ contextx.IContext) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(data)), nil
	}
}

func TestGetOrFetch_CacheMiss(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "hello world"
	md5val := md5sum(content)
	fetchCalled := 0
	fetchFn := func(_ contextx.IContext) (io.ReadCloser, error) {
		fetchCalled++

		return io.NopCloser(strings.NewReader(content)), nil
	}

	file, dir, err := fc.GetOrFetch(nCtx, "test.tgz", md5val, fetchFn)
	if err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	if fetchCalled != 1 {
		t.Errorf("expected fetchFn called 1 time, got %d", fetchCalled)
	}

	if file.Info().MD5 != md5val {
		t.Errorf("expected MD5 %q, got %q", md5val, file.Info().MD5)
	}

	if !strings.Contains(dir, md5val) {
		t.Errorf("expected dir to contain md5, dir=%q md5=%q", dir, md5val)
	}
}

func TestGetOrFetch_CacheHit(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "cached content"
	md5val := md5sum(content)

	var fetchCalled int64
	fetchFn := func(_ contextx.IContext) (io.ReadCloser, error) {
		atomic.AddInt64(&fetchCalled, 1)

		return io.NopCloser(strings.NewReader(content)), nil
	}

	// First call: miss
	if _, _, err = fc.GetOrFetch(nCtx, "test.tgz", md5val, fetchFn); err != nil {
		t.Fatalf("first GetOrFetch: %v", err)
	}

	// Second call: hit
	if _, _, err = fc.GetOrFetch(nCtx, "test.tgz", md5val, fetchFn); err != nil {
		t.Fatalf("second GetOrFetch: %v", err)
	}

	if atomic.LoadInt64(&fetchCalled) != 1 {
		t.Errorf("expected fetchFn called once, got %d", atomic.LoadInt64(&fetchCalled))
	}
}

func TestGetOrFetch_MD5Mismatch(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	wrongMD5 := strings.Repeat("a", 32)
	fetchFn := makeFetchFn("actual content")

	_, _, err = fc.GetOrFetch(nCtx, "test.tgz", wrongMD5, fetchFn)
	if err == nil {
		t.Fatal("expected error for MD5 mismatch, got nil")
	}

	if !strings.Contains(err.Error(), "MD5 mismatch") {
		t.Errorf("expected MD5 mismatch error, got: %v", err)
	}
}

func TestGetOrFetch_FetchFnError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	fetchFn := func(_ contextx.IContext) (io.ReadCloser, error) {
		return nil, fmt.Errorf("network error")
	}

	// Use a valid 32-char hex MD5 so the format check passes and we reach fetchFn.
	validMD5 := strings.Repeat("a", 32)
	_, _, err = fc.GetOrFetch(nCtx, "test.tgz", validMD5, fetchFn)
	if err == nil {
		t.Fatal("expected error from fetchFn, got nil")
	}

	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected network error in message, got: %v", err)
	}
}

func TestGetOrFetch_PathTraversalPrevention(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	// Path traversal filename should be sanitized to just the base name.
	traversal := "../../../etc/passwd"
	content := "content"
	md5val := md5sum(content)
	fetchFn := makeFetchFn(content)

	file, dir, err := fc.GetOrFetch(nCtx, traversal, md5val, fetchFn)
	if err != nil {
		// The call may also succeed if sanitized to "passwd" — both outcomes are acceptable.
		t.Logf("GetOrFetch returned error (acceptable): %v", err)

		return
	}

	// If success, the stored file must be inside baseDir.
	if !strings.HasPrefix(dir, baseDir) {
		t.Errorf("file dir %q escapes baseDir %q", dir, baseDir)
	}

	if file == nil {
		t.Error("expected non-nil file")
	}
}

func TestConcurrentGetOrFetch_SameFile(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "concurrent content"
	md5val := md5sum(content)

	var fetchCalled int64
	fetchFn := func(_ contextx.IContext) (io.ReadCloser, error) {
		atomic.AddInt64(&fetchCalled, 1)
		time.Sleep(10 * time.Millisecond) // simulate network latency

		return io.NopCloser(strings.NewReader(content)), nil
	}

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			_, _, errs[idx] = fc.GetOrFetch(nCtx, "concurrent.tgz", md5val, fetchFn)
		}(i)
	}

	wg.Wait()

	for i, e := range errs {
		if e != nil {
			t.Errorf("goroutine %d: %v", i, e)
		}
	}

	if atomic.LoadInt64(&fetchCalled) > 1 {
		t.Errorf("expected at most 1 fetch call, got %d", atomic.LoadInt64(&fetchCalled))
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "exists content"
	md5val := md5sum(content)

	if fc.FileExists("test.tgz", md5val) {
		t.Error("expected FileExists=false before caching")
	}

	if _, _, err = fc.GetOrFetch(nCtx, "test.tgz", md5val, makeFetchFn(content)); err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	if !fc.FileExists("test.tgz", md5val) {
		t.Error("expected FileExists=true after caching")
	}

	if fc.FileExists("test.tgz", strings.Repeat("b", 32)) {
		t.Error("expected FileExists=false for wrong MD5")
	}
}

func TestRestore(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	content := "restore content"
	md5val := md5sum(content)

	// Prime the cache.
	fc1, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, _, err = fc1.GetOrFetch(nCtx, "restore.tgz", md5val, makeFetchFn(content)); err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	_ = fc1.Close()

	// Verify the file is on disk.
	expectedPath := filepath.Join(baseDir, md5val, "restore.tgz")
	if _, statErr := os.Stat(expectedPath); statErr != nil {
		t.Fatalf("expected file on disk at %q: %v", expectedPath, statErr)
	}

	// Create a new instance: it should restore from disk.
	fc2, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New (second): %v", err)
	}

	defer func() { _ = fc2.Close() }()

	var fetchCalled int64
	fetchFn := func(_ contextx.IContext) (io.ReadCloser, error) {
		atomic.AddInt64(&fetchCalled, 1)

		return io.NopCloser(strings.NewReader(content)), nil
	}

	_, _, err = fc2.GetOrFetch(nCtx, "restore.tgz", md5val, fetchFn)
	if err != nil {
		t.Fatalf("GetOrFetch after restore: %v", err)
	}

	if atomic.LoadInt64(&fetchCalled) != 0 {
		t.Errorf("expected no fetch after restore, got %d", atomic.LoadInt64(&fetchCalled))
	}
}

func TestGC_RemovesExpiredEntries(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	// Very short expiration and GC interval for testing.
	fc, err := filecache.New(nCtx, baseDir, filecache.Options{
		ExpirationTime: 50 * time.Millisecond,
		GCInterval:     20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "gc content"
	md5val := md5sum(content)

	if _, _, err = fc.GetOrFetch(nCtx, "gc.tgz", md5val, makeFetchFn(content)); err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	// Wait for GC to run and expire the entry.
	time.Sleep(200 * time.Millisecond)

	// The MD5 directory should be cleaned up.
	expectedDir := filepath.Join(baseDir, md5val)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, statErr := os.Stat(expectedDir); os.IsNotExist(statErr) {
			return // cleaned up as expected
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Errorf("expected GC to remove dir %q, but it still exists", expectedDir)
}
