/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package release

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IBinTool defines the bin tool interface.
type IBinTool interface {
	// GetReleaseBinTool gets release bintool.
	GetReleaseBinTool(ctx context.Context, gen types.Generation) (*types.ReleaseBinTool, error)

	// ExistReleaseBinTool checks if release bintool exists.
	ExistReleaseBinTool(ctx context.Context, gen types.Generation) (bool, error)

	// UpsertReleaseBinTool upserts release bintool.
	UpsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) error

	// DeleteReleaseBinTool deletes release bintool.
	DeleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) error
}

// UpsertReleaseBinTool upserts release bintool.
func (s *Storage) UpsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) error {
	releaseType := types.ReleaseTypeBinTool

	return s.daoRelease.UpsertMany(ctx, releaseType, &types.Release{
		Generation: bintool.Generation,
		Type:       releaseType,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   bintool.FileName,
		MD5:        bintool.MD5,
		UpdatedAt:  time.Now(),
	})
}

// DeleteReleaseBinTool deletes release bintool.
func (s *Storage) DeleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) error {
	return s.daoRelease.Delete(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), fileName)
}

// GetReleaseBinTool gets release bintool.
func (s *Storage) GetReleaseBinTool(ctx context.Context, gen types.Generation) (*types.ReleaseBinTool, error) {
	r, err := s.daoRelease.Get(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), "")
	if err != nil {
		return nil, err
	}

	return &types.ReleaseBinTool{
		Generation: r.Generation,
		FileName:   r.FileName,
		MD5:        r.MD5,
	}, nil
}

// ExistReleaseBinTool checks if release bintool exists.
func (s *Storage) ExistReleaseBinTool(ctx context.Context, gen types.Generation) (bool, error) {
	return s.daoRelease.Exist(ctx, types.ReleaseTypeBinTool, gen, platform.EmptyPlatform(), "")
}
