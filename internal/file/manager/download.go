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

// Package manager provides the file manager.
package manager

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/downloader"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// remoteFileDownloadTimeout bounds the whole remote package download.
const remoteFileDownloadTimeout = 25 * time.Minute

// DownloadRemoteFile downloads, verifies, and caches a remote file.
// It delegates to IFileCache.GetOrFetch with the downloader as the fetchFn: on
// a cache hit the cached file is returned immediately; on a miss the downloader
// fetches the content, the cache validates the MD5 and persists it locally,
// then a fileiface.File backed by the local file is returned.
//
// filename is the caller-requested basename used as the cache key and the
// download target name. When it cannot be used safely, the URL basename is
// preferred, and a generated default name is used as the last resort.
func (m *Manager) DownloadRemoteFile(nCtx contextx.IContext, filename, downloadURL, expectedMD5 string) (fileiface.File, error) {
	if m.downloader == nil {
		return nil, errors.New("downloader is nil")
	}
	if m.fileCache == nil {
		return nil, errors.New("file cache is nil")
	}

	// Prefer the caller-requested basename; fall back to the URL basename, then
	// to a generated default when neither yields a safe basename.
	remoteFilename := sanitizeRemoteFilename(filename)
	if remoteFilename == "" {
		remoteFilename = resolveRemoteFilename(downloadURL)
	}

	fetchFn := func(fetchCtx contextx.IContext) (io.ReadCloser, error) {
		downloadCtx, cancel := contextx.WithTimeout(fetchCtx, remoteFileDownloadTimeout)
		defer cancel()

		file, err := m.downloader.Download(downloadCtx, downloadURL, downloader.DownloadOptions{
			Filename: remoteFilename,
			Checksum: downloader.Checksum{
				Algorithm: downloader.ChecksumAlgorithmMD5,
				Value:     expectedMD5,
			},
		})
		if err != nil {
			return nil, err
		}

		// The reader is owned by the cache: LocalDir.Store consumes it fully and
		// closes it exactly once, which also removes the downloader's temporary
		// file. Create it under fetchCtx (not the downloadCtx canceled above) so
		// the context stays valid until the cache finishes copying.
		return file.Content(fetchCtx)
	}

	file, _, err := m.fileCache.GetOrFetch(nCtx, remoteFilename, expectedMD5, fetchFn)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote file into cache: %w", err)
	}

	return file, nil
}

func resolveRemoteFilename(downloadURL string) string {
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		return defaultRemoteFilename()
	}
	if parsedURL.Path == "" || strings.HasSuffix(parsedURL.Path, "/") {
		return defaultRemoteFilename()
	}
	if filename := sanitizeRemoteFilename(path.Base(parsedURL.Path)); filename != "" {
		return filename
	}

	return defaultRemoteFilename()
}

func defaultRemoteFilename() string {
	return fmt.Sprintf("file-remote-download-%d", time.Now().Unix())
}

func sanitizeRemoteFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.NewReplacer("\r", "", "\n", "", "\\", "/").Replace(filename)
	filename = filepath.Base(filename)
	if filename == "." || filename == ".." || filename == string(filepath.Separator) {
		return ""
	}

	return filename
}
