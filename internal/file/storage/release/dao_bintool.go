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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// upsertReleaseBinTool upserts release bintool.
func (s *Storage) upsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) (err error) {
	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypeBinTool, types.Generation2, &types.Release{
		Generation: bintool.Generation,
		Type:       types.ReleaseTypeBinTool,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   bintool.FileName,
		MD5:        bintool.MD5,
		UpdatedAt:  time.Now(),
	})
}

// deleteReleaseBinTool deletes release bintool.
func (s *Storage) deleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) (err error) {
	return s.daoRelease.Delete(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), fileName)
}

// getReleaseBinTool gets release bintool.
func (s *Storage) getReleaseBinTool(ctx context.Context, gen types.Generation) (data *types.ReleaseBinTool, err error) {
	var rls *types.Release
	if rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), ""); err != nil {
		return nil, err
	}

	return &types.ReleaseBinTool{
		Release: *rls,
	}, nil
}

// existReleaseBinTool checks if release bintool exists.
func (s *Storage) existReleaseBinTool(ctx context.Context, gen types.Generation) (result bool, err error) {
	return s.daoRelease.Exist(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), "")
}
