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

// IPluginBinTool defines the bin tool interface.
type IPluginBinTool interface {
	// GetReleasePluginBinTool gets release plugin bintool gen2.
	GetReleasePluginBinTool(ctx context.Context, gen types.Generation) (*types.ReleasePluginBinTool, error)

	// ExistReleasePluginBinTool checks if release plugin bintool gen2 exists.
	ExistReleasePluginBinTool(ctx context.Context, gen types.Generation) (bool, error)

	// UpsertReleasePluginBinTool upserts release plugin bintool gen2.
	UpsertReleasePluginBinTool(ctx context.Context, pluginBinTool types.ReleasePluginBinTool) error

	// DeleteReleasePluginBinTool deletes release plugin bintool gen2.
	DeleteReleasePluginBinTool(ctx context.Context, gen types.Generation, fileName string) error
}

// UpsertReleasePluginBinTool upserts release plugin bintool gen2.
func (s *Storage) UpsertReleasePluginBinTool(ctx context.Context, pluginBinTool types.ReleasePluginBinTool) error {
	releaseType := types.ReleaseTypePluginBinTool

	return s.daoRelease.UpsertMany(ctx, releaseType, types.Generation2, &types.Release{
		Generation: pluginBinTool.Generation,
		Type:       releaseType,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   pluginBinTool.FileName,
		MD5:        pluginBinTool.MD5,
		UpdatedAt:  time.Now(),
	})
}

// DeleteReleasePluginBinTool deletes release plugin bintool gen2.
func (s *Storage) DeleteReleasePluginBinTool(ctx context.Context, gen types.Generation, fileName string) error {
	return s.daoRelease.Delete(ctx, types.ReleaseTypePluginBinTool, gen, platform.EmptyPlatform(), fileName)
}

// GetReleasePluginBinTool gets release plugin bintool gen2.
func (s *Storage) GetReleasePluginBinTool(ctx context.Context, gen types.Generation) (*types.ReleasePluginBinTool, error) {
	r, err := s.daoRelease.Get(ctx, types.ReleaseTypePluginBinTool, gen, platform.EmptyPlatform(), "")
	if err != nil {
		return nil, err
	}

	return &types.ReleasePluginBinTool{
		Release: *r,
	}, nil
}

// ExistReleasePluginBinTool checks if release plugin bintool gen2 exists.
func (s *Storage) ExistReleasePluginBinTool(ctx context.Context, gen types.Generation) (bool, error) {
	return s.daoRelease.Exist(ctx, types.ReleaseTypePluginBinTool, gen, platform.EmptyPlatform(), "")
}
