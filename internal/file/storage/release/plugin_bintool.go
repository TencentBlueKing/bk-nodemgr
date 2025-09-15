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

// UpsertReleasePluginBinTool upserts release plugin bintool.
func (s *Storage) UpsertReleasePluginBinTool(ctx context.Context, pluginBinTool types.ReleasePluginBinTool) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_plugin_bintool")
	defer metric.End(err)

	return s.daoRelease.UpsertMany(ctx, types.ReleaseTypePluginBinTool, types.Generation2, &types.Release{
		Generation: pluginBinTool.Generation,
		Type:       types.ReleaseTypePluginBinTool,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   pluginBinTool.FileName,
		MD5:        pluginBinTool.MD5,
		UpdatedAt:  time.Now(),
	})
}

// DeleteReleasePluginBinTool deletes release plugin bintool.
func (s *Storage) DeleteReleasePluginBinTool(ctx context.Context, gen types.Generation, fileName string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_plugin_bintool")
	defer metric.End(err)

	return s.daoRelease.Delete(ctx, types.ReleaseTypePluginBinTool, gen, platform.EmptyPlatform(), fileName)
}

// GetReleasePluginBinTool gets release plugin bintool.
func (s *Storage) GetReleasePluginBinTool(ctx context.Context, gen types.Generation) (data *types.ReleasePluginBinTool, err error) {
	// record metric.
	metric := s.metric().Start("get_plugin_bintool")
	defer metric.End(err)

	var rls *types.Release
	if rls, err = s.daoRelease.Get(ctx, types.ReleaseTypePluginBinTool, gen, platform.EmptyPlatform(), ""); err != nil {
		return nil, err
	}

	return &types.ReleasePluginBinTool{
		Release: *rls,
	}, nil
}

// ExistReleasePluginBinTool checks if release plugin bintool exists.
func (s *Storage) ExistReleasePluginBinTool(ctx context.Context, gen types.Generation) (result bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_plugin_bintool")
	defer metric.End(err)

	return s.daoRelease.Exist(ctx, types.ReleaseTypePluginBinTool, gen, platform.EmptyPlatform(), "")
}
