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

// Package file defines the file manager interface.
package file

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// Options configures the relay file manager cache.
type Options struct {
	// ExpirationTime is how long an entry can be unused before it is reclaimed.
	ExpirationTime time.Duration

	// GCInterval controls how often expired entries are reclaimed.
	GCInterval time.Duration

	// MaxSizeMB caps the total size of cached packages. 0 means unlimited.
	MaxSizeMB int64

	// RestoreOnStart rebuilds the index from disk at startup so a relay restart does not
	// force every package to be transferred again.
	RestoreOnStart bool
}

// fileManagerImpl adapts the generic filecache to the relay's package staging workflow:
// packages arrive in a staging directory pushed by GSE, and the relay promotes them into
// the content-addressed cache that its download server reads from.
type fileManagerImpl struct {
	cache filecache.IFileCache
}

// NewFileManager creates a new file manager backed by a content-addressed local cache.
func NewFileManager(nCtx contextx.IContext, baseDir string, opts Options) (IFileManager, error) {
	cache, err := filecache.New(nCtx, baseDir, filecache.Options{
		ExpirationTime: opts.ExpirationTime,
		GCInterval:     opts.GCInterval,
		MaxSizeMB:      opts.MaxSizeMB,
		RestoreOnStart: opts.RestoreOnStart,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create relay file cache. basedir(%s): %w", baseDir, err)
	}

	return &fileManagerImpl{cache: cache}, nil
}

// StoreFile promotes filename from the staging directory srcPath into the cache.
//
// expectedMD5 is mandatory: the staging file is written by an external GSE transfer, so the
// only way to know it is complete is to hash what was read. A package whose content does not
// match is discarded instead of indexed, which keeps a torn transfer from ever being served.
// When the cache already holds this exact content the staging file is not read at all.
func (fm *fileManagerImpl) StoreFile(
	nCtx contextx.IContext, srcPath, filename, expectedMD5 string) (*fileiface.FileInfo, error) {

	file, _, err := fm.cache.GetOrFetch(nCtx, filename, expectedMD5,
		func(_ contextx.IContext) (io.ReadCloser, error) {
			// Resolve the name through os.Root so the kernel confines it to the staging
			// directory: the package name arrives from a server push and must not be able
			// to reach a file outside it.
			root, rootErr := os.OpenRoot(srcPath)
			if rootErr != nil {
				return nil, fmt.Errorf("failed to open staging dir. path(%s): %w", srcPath, rootErr)
			}

			defer func() { _ = root.Close() }()

			staged, openErr := root.Open(filename)
			if openErr != nil {
				return nil, fmt.Errorf("failed to open staged package. name(%s): %w", filename, openErr)
			}

			return staged, nil
		})
	if err != nil {
		return nil, err
	}

	info := file.Info()

	return &info, nil
}

// GetFile returns the cached file indexed under filename.
// Only MD5-validated content is indexed, so a hit is always a complete package.
func (fm *fileManagerImpl) GetFile(nCtx contextx.IContext, filename string) (fileiface.File, error) {
	file, _, ok := fm.cache.GetFile(nCtx, filename)
	if !ok {
		return nil, fmt.Errorf("file not found. filename(%s)", filename)
	}

	return file, nil
}

// FileExists reports whether the cache already holds this exact content.
func (fm *fileManagerImpl) FileExists(nCtx contextx.IContext, filename, md5 string) bool {
	return fm.cache.FileExists(nCtx, filename, md5)
}

// Close stops the background reclaim goroutine.
func (fm *fileManagerImpl) Close() error {
	return fm.cache.Close()
}
