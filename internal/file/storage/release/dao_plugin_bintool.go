/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release storage interface.
// nolint: nonamedreturns
package release

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	defaultVersionPluginBinTool = "default"
)

// upsertReleasePluginBinTool upserts release plugin bintool.
func (s *Storage) upsertReleasePluginBinTool(ctx context.Context, pluginBinTool types.ReleasePluginBinTool) error {
	err := s.daoRelease.UpsertMany(ctx, types.ReleaseTypePluginBinTool, &types.Release{
		Generation: pluginBinTool.Generation,
		Type:       types.ReleaseTypePluginBinTool,
		Platform:   platform.EmptyPlatform(),
		Version:    defaultVersionPluginBinTool,
		FileName:   pluginBinTool.FileName,
		MD5:        pluginBinTool.MD5,
		UpdatedAt:  time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to upsert release plugin bintool: %w", err)
	}

	return nil
}

// deleteReleasePluginBinTool deletes release plugin bintool.
func (s *Storage) deleteReleasePluginBinTool(ctx context.Context, gen types.Generation, fileName string) error {
	err := s.daoRelease.Delete(ctx, types.ReleaseTypePluginBinTool,
		release.WithGeneration(gen),
		release.WithPlatform(platform.EmptyPlatform()),
		release.WithVersion(defaultVersionPluginBinTool),
		release.WithFileName(fileName),
	)
	if err != nil {
		return fmt.Errorf("failed to delete release plugin bintool: %w", err)
	}

	return nil
}

// getReleasePluginBinTool gets release plugin bintool.
func (s *Storage) getReleasePluginBinTool(ctx context.Context, gen types.Generation) (*types.ReleasePluginBinTool, error) {
	var (
		rls  *types.Release
		data *types.ReleasePluginBinTool
		err  error
	)

	rls, err = s.daoRelease.Get(ctx, types.ReleaseTypePluginBinTool,
		release.WithVersion(defaultVersionPluginBinTool),
		release.WithGeneration(gen),
		release.WithPlatform(platform.EmptyPlatform()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin bintool: %w", err)
	}

	data = &types.ReleasePluginBinTool{
		Release: *rls,
	}

	return data, nil
}

// existReleasePluginBinTool checks if release plugin bintool exists.
func (s *Storage) existReleasePluginBinTool(ctx context.Context, gen types.Generation) (bool, error) {
	var (
		result bool
		err    error
	)

	result, err = s.daoRelease.Exist(ctx, types.ReleaseTypePluginBinTool,
		release.WithVersion(defaultVersionPluginBinTool),
		release.WithGeneration(gen),
		release.WithPlatform(platform.EmptyPlatform()),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check release plugin bintool: %w", err)
	}

	return result, nil
}
