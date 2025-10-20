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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

type localFile struct {
	file fileiface.File
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

// EnsureNodeToLocal ensure the node to local.
func (m *Manager) EnsureNodeToLocal(nCtx contextx.IContext, rt types.ReleaseType, gen types.Generation, plat platfmt.Platform, version string) (
	fileiface.File, string, error) {

	if gen != types.Generation2 {
		return nil, "", fmt.Errorf("not support generation: %d", gen)
	}

	var release types.Release
	switch rt {
	case types.ReleaseTypeAgent:
		releaseAgent, err := m.storageRelease.GetReleaseAgent(nCtx, gen, plat, version)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get release: %w", err)
		}

		release = releaseAgent.Release
	case types.ReleaseTypeProxy:
		releaseProxy, err := m.storageRelease.GetReleaseProxy(nCtx, gen, plat, version)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get release: %w", err)
		}

		release = releaseProxy.Release
	default:
		return nil, "", fmt.Errorf("not support release type: %s", rt)
	}

	return m.ensureReleaseToLocal(nCtx, release)
}

// EnsurePluginToLocal ensure the plugin to local.
func (m *Manager) EnsurePluginToLocal(
	nCtx contextx.IContext,
	name string,
	gen types.Generation,
	plat platfmt.Platform,
	version string,
) (fileiface.File, string, error) {

	if gen != types.Generation2 {
		return nil, "", fmt.Errorf("not support generation: %d", gen)
	}

	releasePlugin, err := m.storageRelease.GetReleasePlugin(nCtx, name, gen, plat, version)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get release: %w", err)
	}

	return m.ensureReleaseToLocal(nCtx, releasePlugin.Release)
}

// ensureReleaseToLocal ensure the release to local.
func (m *Manager) ensureReleaseToLocal(nCtx contextx.IContext, release types.Release) (fileiface.File, string, error) {
	cache, ok := m.localFilePool.get(release.FileName)
	if ok {
		info := cache.file.Info()

		// hit cache. return local file.
		if info.MD5 == release.MD5 {
			return cache.file, cache.path, nil
		}
	}

	var ufg fileiface.FileGroup
	switch release.Type {
	case types.ReleaseTypeAgent:
		ufg = m.upstreamReleaseAgent
	case types.ReleaseTypeProxy:
		ufg = m.upstreamReleaseProxy
	case types.ReleaseTypePlugin:
		ufg = m.upstreamReleasePlugin

	default:
		return nil, "", fmt.Errorf("not support ensuring file to local with release type, type(%s)", release.Type)
	}

	if ufg == nil {
		return nil, "", fmt.Errorf("upstream file group is nil with release type, type(%s)", release.Type)
	}

	// get upstream file.
	upstreamFile, err := ufg.GetFile(nCtx, release.FileName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get upstream file, filename(%s): %w", release.FileName, err)
	}

	// get upstream content.
	content, err := upstreamFile.Content(nCtx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get upstream file content, filename(%s): %w", release.FileName, err)
	}

	// create new local dir.
	cacheDir := filepath.Join(local.GetLocalFileGroupAbsDirPath(m.cacheFileGroup), uuid.New().String())
	if err = os.MkdirAll(cacheDir, 0700); err != nil { // nolint: mnd,gosec
		return nil, "", fmt.Errorf("failed to create temp cache dir, dirpath(%s): %w", cacheDir, err)
	}

	lfg, err := local.NewLocalDir(cacheDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create local file group, dirpath(%s): %w", cacheDir, err)
	}

	// save file to loca.
	if err = lfg.Store(nCtx, fileiface.FileInfo{Name: release.FileName}, content, true); err != nil {
		return nil, "", fmt.Errorf("failed to store file, filename(%s): %w", release.FileName, err)
	}

	file, err := lfg.GetFile(nCtx, release.FileName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get local file, filename(%s): %w", release.FileName, err)
	}

	m.localFilePool.filesMutex.Lock()
	m.localFilePool.files[release.FileName] = &localFile{
		file: file,
		path: cacheDir,
	}
	m.localFilePool.filesMutex.Unlock()

	logger.G.Biz(nCtx).With("filename", file.Info().Name, "cache-dir", cacheDir).Info("ensured release to local cache")

	return file, cacheDir, nil
}
