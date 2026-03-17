/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package filecache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

const (
	defaultGCInterval      = 1 * time.Hour
	defaultExpirationTime  = 72 * time.Hour
	dirDot                 = "."
)

// cachedEntry represents a single cached file entry.
type cachedEntry struct {
	file       fileiface.File
	dirPath    string // absolute path: {cacheBaseDir}/{md5}/
	md5        string
	lastAccess time.Time
	mu         sync.Mutex
}

func (e *cachedEntry) updateLastAccess() {
	e.mu.Lock()
	e.lastAccess = time.Now()
	e.mu.Unlock()
}

func (e *cachedEntry) getLastAccess() time.Time {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.lastAccess
}

// fileCache is the concrete IFileCache implementation.
type fileCache struct {
	baseDir        string
	expirationTime time.Duration

	// indexMu protects the index map (held briefly for reads/writes).
	indexMu sync.RWMutex
	index   map[string]*cachedEntry // filename -> entry

	// dlMu provides per-filename download serialization.
	dlMu sync.Map // filename -> *sync.Mutex

	// pendingDirs tracks directories that are being written by download() but
	// not yet committed to the index. GC skips these to avoid TOCTOU races.
	pendingDirs sync.Map // dirPath -> struct{}

	stopCh chan struct{}
	doneCh chan struct{}
}

// Options configures a new FileCache.
type Options struct {
	// ExpirationTime is how long an entry can be unused before GC removes it.
	// Defaults to 72 hours.
	ExpirationTime time.Duration
	// GCInterval controls how often the GC goroutine runs. Defaults to 1 hour.
	GCInterval time.Duration
	// RestoreOnStart controls whether existing MD5 subdirectories are scanned and
	// loaded into the in-memory index at startup. Defaults to false: the cache
	// starts empty and warms up on first access. Enable when cache entries should
	// survive service restarts (e.g. large files that are expensive to re-download).
	RestoreOnStart bool
}

// New creates a new IFileCache rooted at baseDir.
// The directory is created if it does not exist.
// Existing entries are restored from disk only when opts.RestoreOnStart is true.
func New(nCtx contextx.IContext, baseDir string, opts Options) (IFileCache, error) {
	if err := os.MkdirAll(baseDir, 0750); err != nil { // nolint: mnd
		return nil, fmt.Errorf("failed to create cache base dir. path(%s): %w", baseDir, err)
	}

	expiration := opts.ExpirationTime
	if expiration <= 0 {
		expiration = defaultExpirationTime
	}

	gcInterval := opts.GCInterval
	if gcInterval <= 0 {
		gcInterval = defaultGCInterval
	}

	fc := &fileCache{
		baseDir:        baseDir,
		expirationTime: expiration,
		index:          make(map[string]*cachedEntry),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}

	if opts.RestoreOnStart {
		fc.restore(nCtx)
	}

	go fc.runGCLoop(gcInterval)

	return fc, nil
}

// GetOrFetch implements IFileCache.
func (fc *fileCache) GetOrFetch(
	nCtx contextx.IContext,
	filename string,
	expectedMD5 string,
	fetchFn func(nCtx contextx.IContext) (io.ReadCloser, error),
) (fileiface.File, string, error) {
	// sanitize filename to prevent path traversal.
	filename = filepath.Base(filename)
	if filename == "" || filename == dirDot {
		return nil, "", fmt.Errorf("invalid filename after sanitization: %q", filename)
	}

	// fast path: check index under read lock.
	if file, dirPath, ok := fc.lookupIndex(filename, expectedMD5); ok {
		logger.G.Sys().With("filename", filename, "md5", expectedMD5).Info("file cache hit")

		return file, dirPath, nil
	}

	// slow path: serialize downloads per filename.
	mu := fc.getOrCreateDLMutex(filename)
	mu.Lock()
	defer mu.Unlock()

	// double-check after acquiring the per-filename lock.
	if file, dirPath, ok := fc.lookupIndex(filename, expectedMD5); ok {
		logger.G.Sys().With("filename", filename, "md5", expectedMD5).Info("file cache hit (double-check)")

		return file, dirPath, nil
	}

	logger.G.Sys().With("filename", filename, "md5", expectedMD5).Info("file cache miss, downloading")

	file, dirPath, err := fc.download(nCtx, filename, expectedMD5, fetchFn)
	if err != nil {
		return nil, "", err
	}

	return file, dirPath, nil
}

// FileExists implements IFileCache.
func (fc *fileCache) FileExists(filename string, expectedMD5 string) bool {
	filename = filepath.Base(filename)
	_, _, ok := fc.lookupIndex(filename, expectedMD5)

	return ok
}

// Close implements IFileCache.
func (fc *fileCache) Close() error {
	close(fc.stopCh)
	<-fc.doneCh

	return nil
}

// lookupIndex checks the index under a read lock.
// Returns (file, dirPath, true) on a cache hit.
func (fc *fileCache) lookupIndex(filename, expectedMD5 string) (fileiface.File, string, bool) {
	fc.indexMu.RLock()
	entry, ok := fc.index[filename]
	fc.indexMu.RUnlock()

	if !ok || entry.md5 != expectedMD5 {
		return nil, "", false
	}

	entry.updateLastAccess()

	return entry.file, entry.dirPath, true
}

// getOrCreateDLMutex returns the per-filename download mutex, creating it if absent.
func (fc *fileCache) getOrCreateDLMutex(filename string) *sync.Mutex {
	mu := &sync.Mutex{}
	actual, _ := fc.dlMu.LoadOrStore(filename, mu)

	return actual.(*sync.Mutex) // nolint: forcetypeassert
}

// isMD5Hex reports whether s is a valid lowercase or uppercase 32-char MD5 hex string.
func isMD5Hex(s string) bool {
	if len(s) != 32 { // nolint: mnd
		return false
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// download fetches the file via fetchFn, stores it under {baseDir}/{md5}/, validates MD5,
// and updates the index.
func (fc *fileCache) download(
	nCtx contextx.IContext,
	filename string,
	expectedMD5 string,
	fetchFn func(nCtx contextx.IContext) (io.ReadCloser, error),
) (fileiface.File, string, error) {
	// Validate MD5 format before using it as a directory name to prevent path traversal.
	if !isMD5Hex(expectedMD5) {
		return nil, "", fmt.Errorf("invalid expectedMD5 format %q: must be 32 hex chars", expectedMD5)
	}

	// Compute dirPath early so we can register it as pending before touching the disk.
	dirPath := filepath.Join(fc.baseDir, expectedMD5)

	// Mark this directory as in-progress so GC will not treat it as orphaned
	// during the window between os.MkdirAll and the index update.
	fc.pendingDirs.Store(dirPath, struct{}{})
	defer fc.pendingDirs.Delete(dirPath)

	content, err := fetchFn(nCtx)
	if err != nil {
		// Log with filename context; propagate the original error unchanged so
		// callers can inspect it directly (avoiding a double-wrapped message).
		logger.G.Sys().WithErr(err).With("filename", filename).Error("file cache fetch failed")

		return nil, "", err
	}

	if err = os.MkdirAll(dirPath, 0750); err != nil { // nolint: mnd
		_ = content.Close()

		return nil, "", fmt.Errorf("failed to create cache subdir %q: %w", dirPath, err)
	}

	ld, err := local.NewLocalDir(dirPath)
	if err != nil {
		_ = content.Close()

		return nil, "", fmt.Errorf("failed to open local dir %q: %w", dirPath, err)
	}

	if err = ld.Store(nCtx, fileiface.FileInfo{Name: filename}, content, true); err != nil {
		_ = safeRemoveAll(dirPath, fc.baseDir)

		return nil, "", fmt.Errorf("failed to store file %q in cache: %w", filename, err)
	}

	file, err := ld.GetFile(nCtx, filename)
	if err != nil {
		_ = safeRemoveAll(dirPath, fc.baseDir)

		return nil, "", fmt.Errorf("failed to get stored file %q: %w", filename, err)
	}

	// Cross-validate MD5.
	actualMD5 := file.Info().MD5
	if actualMD5 != expectedMD5 {
		logger.G.Sys().
			With("filename", filename, "expected-md5", expectedMD5, "actual-md5", actualMD5).
			Error("file cache MD5 mismatch after download, discarding")

		_ = safeRemoveAll(dirPath, fc.baseDir)

		return nil, "", fmt.Errorf("MD5 mismatch for %q: expected=%s actual=%s", filename, expectedMD5, actualMD5)
	}

	// Update index under write lock.
	fc.indexMu.Lock()
	fc.index[filename] = &cachedEntry{
		file:       file,
		dirPath:    dirPath,
		md5:        expectedMD5,
		lastAccess: time.Now(),
	}
	fc.indexMu.Unlock()

	logger.G.Sys().With("filename", filename, "md5", expectedMD5, "dir", dirPath).Info("file downloaded and cached")

	return file, dirPath, nil
}

// restore scans the base directory to rebuild the in-memory index from MD5-named subdirectories.
// No MD5 re-computation is needed: the directory name IS the MD5.
func (fc *fileCache) restore(nCtx contextx.IContext) {
	entries, err := os.ReadDir(fc.baseDir)
	if err != nil {
		logger.G.Sys().WithErr(err).With("basedir", fc.baseDir).Error("failed to read cache base dir during restore")

		return
	}

	restored := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		md5Val := entry.Name()
		subDir := filepath.Join(fc.baseDir, md5Val)

		ld, ldErr := local.NewLocalDir(subDir)
		if ldErr != nil {
			continue
		}

		files, filesErr := ld.AllFiles(nCtx)
		if filesErr != nil || len(files) == 0 {
			continue
		}

		// Pick the file with the most recent mtime if there are multiple.
		var newest fileiface.File
		for _, f := range files {
			if newest == nil || f.Info().ModTime.After(newest.Info().ModTime) {
				newest = f
			}
		}

		filename := newest.Info().Name

		// Keep the entry with the newest mtime if filename already registered.
		fc.indexMu.Lock()
		existing, ok := fc.index[filename]
		if !ok || newest.Info().ModTime.After(existing.file.Info().ModTime) {
			fc.index[filename] = &cachedEntry{
				file:       newest,
				dirPath:    subDir,
				md5:        md5Val,
				lastAccess: time.Now(),
			}
			restored++
		}
		fc.indexMu.Unlock()
	}

	logger.G.Sys().With("basedir", fc.baseDir, "restored", restored).Info("file cache restore complete")
}

// runGCLoop runs the GC periodically until stopCh is closed.
func (fc *fileCache) runGCLoop(interval time.Duration) {
	defer close(fc.doneCh)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-fc.stopCh:
			return
		case <-ticker.C:
			fc.runGC()
		}
	}
}

// runGC removes expired index entries and orphaned MD5 directories.
func (fc *fileCache) runGC() {
	cutoff := time.Now().Add(-fc.expirationTime)

	// Remove expired entries and build a fresh activeDirs snapshot in a single critical section.
	// Building the snapshot here (after deletions) ensures it reflects the current index state,
	// which is used below to detect orphaned directories without a stale-read race.
	fc.indexMu.Lock()

	var toDelete []string

	for filename, entry := range fc.index {
		if entry.getLastAccess().Before(cutoff) {
			toDelete = append(toDelete, filename)
		}
	}

	for _, filename := range toDelete {
		entry := fc.index[filename]
		delete(fc.index, filename)

		// Clean up the per-filename download mutex to avoid unbounded growth.
		fc.dlMu.Delete(filename)

		logger.G.Sys().With("filename", filename, "dir", entry.dirPath).Info("file cache GC: removing expired entry")

		go func(dir string) {
			_ = safeRemoveAll(dir, fc.baseDir)
		}(entry.dirPath)
	}

	// Snapshot active dirs AFTER deletions so the orphan scan below sees the true live set.
	activeDirs := make(map[string]struct{}, len(fc.index))
	for _, entry := range fc.index {
		activeDirs[entry.dirPath] = struct{}{}
	}

	fc.indexMu.Unlock()

	// Scan for orphaned MD5 directories not referenced by the index.
	// We intentionally skip directories registered in pendingDirs: those belong to
	// downloads that have created the directory but not yet committed to the index.
	// Deleting them would corrupt in-flight writes (TOCTOU race).
	entries, err := os.ReadDir(fc.baseDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(fc.baseDir, entry.Name())

		if _, active := activeDirs[dirPath]; active {
			continue
		}

		if _, pending := fc.pendingDirs.Load(dirPath); pending {
			continue
		}

		logger.G.Sys().With("dir", dirPath).Info("file cache GC: removing orphaned MD5 dir")

		go func(dir string) {
			_ = safeRemoveAll(dir, fc.baseDir)
		}(dirPath)
	}
}

// safeRemoveAll removes absPath only if it is a sub-path of baseDir.
// Cheap well-known checks are applied first; filepath.Rel is only called afterwards.
func safeRemoveAll(absPath, baseDir string) error {
	if absPath == "" || absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/") {

		return fmt.Errorf("refused to remove protected path %q", absPath)
	}

	if !isSubPath(absPath, baseDir) {
		logger.G.Sys().With("path", absPath, "basedir", baseDir).Error("refused to remove path outside basedir")

		return fmt.Errorf("path %q is not under basedir %q", absPath, baseDir)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove %q: %w", absPath, err)
	}

	return nil
}

// isSubPath reports whether target is a descendant of base.
func isSubPath(target, base string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, dirDot) && rel != dirDot
}
