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

	if fc.FileExists(nCtx, "test.tgz", md5val) {
		t.Error("expected FileExists=false before caching")
	}

	if _, _, err = fc.GetOrFetch(nCtx, "test.tgz", md5val, makeFetchFn(content)); err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	if !fc.FileExists(nCtx, "test.tgz", md5val) {
		t.Error("expected FileExists=true after caching")
	}

	if fc.FileExists(nCtx, "test.tgz", strings.Repeat("b", 32)) {
		t.Error("expected FileExists=false for wrong MD5")
	}
}

func TestGetOrFetch_DropsStaleIndexWhenDiskFileMissing(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "stale content"
	md5val := md5sum(content)
	filename := "stale.tgz"

	var fetchCalled int64
	fetchFn := func(_ contextx.IContext) (io.ReadCloser, error) {
		atomic.AddInt64(&fetchCalled, 1)

		return io.NopCloser(strings.NewReader(content)), nil
	}

	if _, _, err = fc.GetOrFetch(nCtx, filename, md5val, fetchFn); err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	if err = os.RemoveAll(baseDir); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	if fc.FileExists(nCtx, filename, md5val) {
		t.Fatal("expected FileExists to drop stale index after backing file is removed")
	}

	if _, _, err = fc.GetOrFetch(nCtx, filename, md5val, fetchFn); err != nil {
		t.Fatalf("GetOrFetch after FileExists stale miss: %v", err)
	}

	if err = os.RemoveAll(baseDir); err != nil {
		t.Fatalf("RemoveAll after refetch: %v", err)
	}

	if _, _, ok := fc.GetFile(nCtx, filename); ok {
		t.Fatal("expected GetFile to drop stale index after backing file is removed")
	}

	if _, _, err = fc.GetOrFetch(nCtx, filename, md5val, fetchFn); err != nil {
		t.Fatalf("GetOrFetch after GetFile stale miss: %v", err)
	}

	if err = os.RemoveAll(baseDir); err != nil {
		t.Fatalf("RemoveAll after second refetch: %v", err)
	}

	if _, _, err = fc.GetOrFetch(nCtx, filename, md5val, fetchFn); err != nil {
		t.Fatalf("GetOrFetch after stale lookup: %v", err)
	}

	if atomic.LoadInt64(&fetchCalled) != 4 {
		t.Fatalf("expected stale index to force each refetch, got %d calls", atomic.LoadInt64(&fetchCalled))
	}

	if !fc.FileExists(nCtx, filename, md5val) {
		t.Fatal("expected FileExists=true after final refetch")
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
	fc2, err := filecache.New(nCtx, baseDir, filecache.Options{RestoreOnStart: true})
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

// sizedContent returns a payload of roughly the requested KB, unique per tag so each
// payload lands in its own MD5 directory.
func sizedContent(tag string, kb int) string {
	return tag + strings.Repeat("x", kb*1024)
}

// waitForDirGone polls until dir disappears. Cache directories are removed asynchronously,
// so callers cannot assert on their absence immediately.
func waitForDirGone(t *testing.T, dir string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Errorf("expected dir %q to be removed, but it still exists", dir)
}

func TestEvictBySize_RemovesLeastRecentlyUsed(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	// 1MB cap with two 600KB payloads: storing the second one must push the first out.
	fc, err := filecache.New(nCtx, baseDir, filecache.Options{MaxSizeMB: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	oldContent, newContent := sizedContent("old", 600), sizedContent("new", 600)
	oldMD5, newMD5 := md5sum(oldContent), md5sum(newContent)

	if _, _, err = fc.GetOrFetch(nCtx, "old.tgz", oldMD5, makeFetchFn(oldContent)); err != nil {
		t.Fatalf("GetOrFetch(old): %v", err)
	}

	// Keep the access timestamps distinguishable so LRU order is deterministic.
	time.Sleep(10 * time.Millisecond)

	if _, _, err = fc.GetOrFetch(nCtx, "new.tgz", newMD5, makeFetchFn(newContent)); err != nil {
		t.Fatalf("GetOrFetch(new): %v", err)
	}

	if fc.FileExists(nCtx, "old.tgz", oldMD5) {
		t.Error("expected the least recently used entry to be evicted, but it is still cached")
	}

	if !fc.FileExists(nCtx, "new.tgz", newMD5) {
		t.Error("expected the most recently stored entry to survive eviction")
	}

	waitForDirGone(t, filepath.Join(baseDir, oldMD5))
}

func TestEvictBySize_UnlimitedByDefault(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	// MaxSizeMB left at 0: this is what internal/file and backend rely on, so nothing may
	// be evicted regardless of total size.
	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	first, second := sizedContent("first", 600), sizedContent("second", 600)
	firstMD5, secondMD5 := md5sum(first), md5sum(second)

	if _, _, err = fc.GetOrFetch(nCtx, "first.tgz", firstMD5, makeFetchFn(first)); err != nil {
		t.Fatalf("GetOrFetch(first): %v", err)
	}

	if _, _, err = fc.GetOrFetch(nCtx, "second.tgz", secondMD5, makeFetchFn(second)); err != nil {
		t.Fatalf("GetOrFetch(second): %v", err)
	}

	if !fc.FileExists(nCtx, "first.tgz", firstMD5) || !fc.FileExists(nCtx, "second.tgz", secondMD5) {
		t.Error("expected no eviction when MaxSizeMB is 0")
	}
}

func TestGetFile(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	content := "get file content"
	md5val := md5sum(content)

	if _, _, ok := fc.GetFile(nCtx, "absent.tgz"); ok {
		t.Error("expected GetFile to report a miss for an uncached filename")
	}

	_, wantDir, err := fc.GetOrFetch(nCtx, "present.tgz", md5val, makeFetchFn(content))
	if err != nil {
		t.Fatalf("GetOrFetch: %v", err)
	}

	file, gotDir, ok := fc.GetFile(nCtx, "present.tgz")
	if !ok {
		t.Fatal("expected GetFile to report a hit for a cached filename")
	}

	if gotDir != wantDir {
		t.Errorf("GetFile dir = %q, want %q", gotDir, wantDir)
	}

	if file.Info().MD5 != md5val {
		t.Errorf("GetFile MD5 = %q, want %q", file.Info().MD5, md5val)
	}
}

func TestGetFile_RefreshesEvictionOrder(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	// 2MB cap with three 800KB payloads: the third store evicts exactly one entry.
	// Touching the first entry via GetFile in between must make the second one the victim.
	fc, err := filecache.New(nCtx, baseDir, filecache.Options{MaxSizeMB: 2})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	names := []string{"a.tgz", "b.tgz", "c.tgz"}
	md5s := make([]string, len(names))

	for i, name := range names {
		content := sizedContent(name, 800)
		md5s[i] = md5sum(content)

		if i == 2 {
			// Refresh a.tgz so b.tgz becomes the least recently used entry.
			if _, _, ok := fc.GetFile(nCtx, names[0]); !ok {
				t.Fatalf("expected %s to still be cached before the evicting store", names[0])
			}

			time.Sleep(10 * time.Millisecond)
		}

		if _, _, err = fc.GetOrFetch(nCtx, name, md5s[i], makeFetchFn(content)); err != nil {
			t.Fatalf("GetOrFetch(%s): %v", name, err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !fc.FileExists(nCtx, names[0], md5s[0]) {
		t.Errorf("expected %s to survive: GetFile should have refreshed its access time", names[0])
	}

	if fc.FileExists(nCtx, names[1], md5s[1]) {
		t.Errorf("expected %s to be evicted as the least recently used entry", names[1])
	}

	if !fc.FileExists(nCtx, names[2], md5s[2]) {
		t.Errorf("expected %s to survive as the newest entry", names[2])
	}
}

func TestRestore_SkipsNonMD5Dirs(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	content := "restore filter content"
	md5val := md5sum(content)

	// A valid content-addressed entry.
	validDir := filepath.Join(baseDir, md5val)
	if err := os.MkdirAll(validDir, 0750); err != nil {
		t.Fatalf("MkdirAll(valid): %v", err)
	}

	if err := os.WriteFile(filepath.Join(validDir, "keep.tgz"), []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(valid): %v", err)
	}

	// A directory from an older layout, whose name is not an MD5. Indexing it would create
	// an entry that can never be hit yet still counts as active, so it must be skipped and
	// then reclaimed as an orphan.
	staleDir := filepath.Join(baseDir, "20250101120000_1")
	if err := os.MkdirAll(staleDir, 0750); err != nil {
		t.Fatalf("MkdirAll(stale): %v", err)
	}

	if err := os.WriteFile(filepath.Join(staleDir, "stale.tgz"), []byte("stale"), 0600); err != nil {
		t.Fatalf("WriteFile(stale): %v", err)
	}

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{
		RestoreOnStart: true,
		GCInterval:     50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	if !fc.FileExists(nCtx, "keep.tgz", md5val) {
		t.Error("expected the MD5-named dir to be restored into the index")
	}

	if _, _, ok := fc.GetFile(nCtx, "stale.tgz"); ok {
		t.Error("expected the non-MD5 dir to be skipped during restore")
	}

	waitForDirGone(t, staleDir)

	if !fc.FileExists(nCtx, "keep.tgz", md5val) {
		t.Error("expected the valid entry to survive the orphan scan")
	}
}

func TestConcurrentGetOrFetch_WithEviction(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	nCtx := testCtx()

	fc, err := filecache.New(nCtx, baseDir, filecache.Options{MaxSizeMB: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	defer func() { _ = fc.Close() }()

	const goroutines = 12

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()

			name := fmt.Sprintf("pkg-%d.tgz", idx)
			content := sizedContent(name, 200)

			// Eviction may remove the entry right after it is stored, which is expected under
			// a tight cap; only a hard error is a failure here.
			if _, _, fetchErr := fc.GetOrFetch(nCtx, name, md5sum(content), makeFetchFn(content)); fetchErr != nil {
				t.Errorf("GetOrFetch(%s): %v", name, fetchErr)
			}

			fc.GetFile(nCtx, name)
		}(i)
	}

	wg.Wait()
}
