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
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

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

// ensureReleaseToLocal ensures the release file is available locally.
// It delegates to IFileCache.GetOrFetch: on a cache hit the cached file is returned immediately;
// on a miss the upstream FileGroup is used as the fetchFn to download and cache the file.
func (m *Manager) ensureReleaseToLocal(nCtx contextx.IContext, release types.Release) (fileiface.File, string, error) {
	var ufg fileiface.FileGroup
	switch release.Type {
	case types.ReleaseTypeAgent:
		ufg = m.upstreamReleaseAgent
	case types.ReleaseTypeProxy:
		ufg = m.upstreamReleaseProxy
	case types.ReleaseTypePlugin:
		ufg = m.upstreamReleasePlugin
	case types.ReleaseTypeCert:
		ufg = m.upstreamReleaseCert
	case types.ReleaseTypeBinTool:
		ufg = m.upstreamReleaseBinTool
	case types.ReleaseTypePluginBinTool:
		ufg = m.upstreamReleasePluginBinTool
	default:
		return nil, "", fmt.Errorf("not support ensuring file to local with release type, type(%s)", release.Type)
	}

	if ufg == nil {
		return nil, "", fmt.Errorf("upstream file group is nil with release type, type(%s)", release.Type)
	}

	filename := release.FileName
	expectedMD5 := release.MD5

	fetchFn := func(ctx contextx.IContext) (io.ReadCloser, error) {
		upstreamFile, err := ufg.GetFile(ctx, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to get upstream file, filename(%s): %w", filename, err)
		}

		content, err := upstreamFile.Content(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get upstream file content, filename(%s): %w", filename, err)
		}

		return content, nil
	}

	return m.fileCache.GetOrFetch(nCtx, filename, expectedMD5, fetchFn)
}
