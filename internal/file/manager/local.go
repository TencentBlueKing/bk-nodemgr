/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

type localFile struct {
	file iface.File
	path string
}

type localFilePool struct {
	filesMutex sync.RWMutex
	files      map[string]*localFile
}

func (lfp *localFilePool) get(filename string) (*localFile, bool) {
	lfp.filesMutex.RLock()
	defer lfp.filesMutex.RUnlock()

	data, ok := lfp.files[filename]

	return data, ok
}

// EnsureFileToLocal ensure the file to local.
func (m *Manager) EnsureFileToLocal(ctx context.Context,
	gen types.Generation,
	rt types.ReleaseType,
	plat platform.Platform,
	version string) (iface.File, string, error) {

	if gen != types.Generation2 {
		return nil, "", fmt.Errorf("not support generation: %d", gen)
	}

	if rt != types.ReleaseTypeAgent && rt != types.ReleaseTypeProxy {
		return nil, "", fmt.Errorf("not support release type: %s", rt)
	}

	release, err := m.storageRelease.GetRelease(ctx, gen, rt, plat, version)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get release: %w", err)
	}

	return m.EnsureReleaseToLocal(ctx, release)
}

// EnsureReleaseToLocal ensure the release to local.
func (m *Manager) EnsureReleaseToLocal(ctx context.Context, release *types.Release) (iface.File, string, error) {
	cache, ok := m.localFilePool.get(release.FileName)
	if ok {
		info := cache.file.Info()

		// hit cache. return local file.
		if info.MD5 == release.MD5 {
			return cache.file, "", nil
		}
	}

	var ufg iface.FileGroup
	switch release.Type {
	case types.ReleaseTypeAgent:
		ufg = m.upstreamReleaseAgent

	case types.ReleaseTypeProxy:
		ufg = m.upstreamReleaseProxy

	default:
		return nil, "", fmt.Errorf("not support ensuring file to local with release type: %s", release.Type)
	}

	if ufg == nil {
		return nil, "", fmt.Errorf("upstream file group is nil with release type: %s", release.Type)
	}

	// get upstream file.
	upstreamFile, err := ufg.GetFile(ctx, release.FileName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get upstream file: %s, err: %w", release.FileName, err)
	}

	// get upstream content.
	content, err := upstreamFile.Content(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get upstream file content: %s, err: %w", release.FileName, err)
	}

	// create new local dir.
	tempDir := filepath.Join(m.localDir, uuid.New().String())
	if err = os.MkdirAll(tempDir, 0755); err != nil { // nolint: mnd,gosec
		return nil, "", fmt.Errorf("failed to create temp dir: %s, err: %w", tempDir, err)
	}

	lfg, err := local.NewLocalDir(tempDir, m.logger)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create local file group: %s, err: %w", tempDir, err)
	}

	// save file to loca.
	if err = lfg.Store(ctx, iface.FileInfo{Name: release.FileName}, content, true); err != nil {
		return nil, "", fmt.Errorf("failed to store file: %s, err: %w", release.FileName, err)
	}

	file, _ := lfg.GetFile(ctx, release.FileName)

	m.localFilePool.filesMutex.Lock()
	m.localFilePool.files[release.FileName] = &localFile{
		file: file,
		path: tempDir,
	}
	m.localFilePool.filesMutex.Unlock()

	return file, tempDir, nil
}
