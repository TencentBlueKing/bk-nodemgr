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
	defaultVersionBinTool = "default"
)

// upsertReleaseBinTool upserts release bintool.
func (s *Storage) upsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) error {
	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeBinTool, &types.Release{
		Generation: bintool.Generation,
		Type:       types.ReleaseTypeBinTool,
		Platform:   platform.EmptyPlatform(),
		Version:    defaultVersionBinTool,
		FileName:   bintool.FileName,
		MD5:        bintool.MD5,
		UpdatedAt:  time.Now(),
	})
}

// deleteReleaseBinTool deletes release bintool.
func (s *Storage) deleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) error {
	err := s.daoRelease.Delete(ctx, types.ReleaseTypeBinTool, release.WithGeneration(gen), release.WithFileName(fileName))
	if err != nil {
		return fmt.Errorf("failed to delete release bintool: %w", err)
	}

	return nil
}

// getReleaseBinTool gets release bintool.
func (s *Storage) getReleaseBinTool(ctx context.Context, gen types.Generation) (*types.ReleaseBinTool, error) {
	rls, err := s.daoRelease.Get(ctx, types.ReleaseTypeBinTool,
		release.WithGeneration(gen),
		release.WithPlatform(platform.EmptyPlatform()),
		release.WithVersion(defaultVersionBinTool),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release bintool: %w", err)
	}

	return &types.ReleaseBinTool{
		Release: *rls,
	}, nil
}

// existReleaseBinTool checks if release bintool exists.
func (s *Storage) existReleaseBinTool(ctx context.Context, gen types.Generation) (bool, error) {
	result, err := s.daoRelease.Exist(ctx, types.ReleaseTypeBinTool,
		release.WithPlatform(platform.EmptyPlatform()),
		release.WithGeneration(gen),
		release.WithVersion(defaultVersionBinTool),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check release bintool: %w", err)
	}

	return result, nil
}
