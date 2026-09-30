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

// Package filecache provides a generic local file cache with MD5-based deduplication,
// atomic writes, startup restore, and background GC.
package filecache

import (
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// IFileCache is the interface for local file caching with MD5-based deduplication.
type IFileCache interface {
	// GetOrFetch returns a cached file matching filename and expectedMD5.
	// If the cache does not have a matching entry, fetchFn is called to download the file.
	// The downloaded file is written to an isolated MD5-named subdirectory so in-use references
	// to older versions remain valid. Filename is sanitized with filepath.Base before use.
	GetOrFetch(nCtx contextx.IContext, filename string, expectedMD5 string,
		fetchFn func(nCtx contextx.IContext) (io.ReadCloser, error)) (fileiface.File, string, error)

	// FileExists reports whether a file with the given filename and MD5 is present in the cache.
	FileExists(nCtx contextx.IContext, filename string, expectedMD5 string) bool

	// GetFile returns the cached file currently indexed under filename, without requiring the
	// caller to know its MD5. Only entries that already passed MD5 validation are indexed, so a
	// hit is always a verified copy. Reports false when filename is not cached.
	// Use FileExists or GetOrFetch when the expected MD5 is known.
	GetFile(nCtx contextx.IContext, filename string) (fileiface.File, string, bool)

	// Close stops the background GC goroutine and releases resources.
	Close() error
}
