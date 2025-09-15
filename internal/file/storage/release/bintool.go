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

// UpsertReleaseBinTool upserts release bintool gen2.
func (s *Storage) UpsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_bintool")
	defer metric.End(err)

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

// DeleteReleaseBinTool deletes release bintool gen2.
func (s *Storage) DeleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_bintool")
	defer metric.End(err)

	return s.daoRelease.Delete(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), fileName)
}

// GetReleaseBinTool gets release bintool gen2.
func (s *Storage) GetReleaseBinTool(ctx context.Context, gen types.Generation) (data *types.ReleaseBinTool, err error) {
	// record metric.
	metric := s.metric().Start("get_bintool")
	defer metric.End(err)

	var rls *types.Release
	if rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), ""); err != nil {
		return nil, err
	}

	return &types.ReleaseBinTool{
		Release: *rls,
	}, nil
}

// ExistReleaseBinTool checks if release bintool gen2 exists.
func (s *Storage) ExistReleaseBinTool(ctx context.Context, gen types.Generation) (result bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_bintool")
	defer metric.End(err)

	return s.daoRelease.Exist(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), "")
}
