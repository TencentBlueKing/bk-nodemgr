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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// upsertReleasePluginBinToolV2 upserts release plugin bintool v2.
func (s *Storage) upsertReleasePluginBinToolV2(ctx contextx.IContext, pluginBinTool types.ReleasePluginBinToolV2) error {
	rls := &pluginBinTool.Release
	rls.Operator = ctx.BKUsername()
	rls.Name = releaseNamePluginBinTool
	rls.Version = releaseVersionPluginBinTool
	rls.UpdatedAt = time.Now()

	err := s.daoRelease.UpsertMany(ctx, types.ReleaseTypePluginBinToolV2, rls)
	if err != nil {
		return fmt.Errorf("failed to upsert release plugin bintool v2: %w", err)
	}

	return nil
}

// deleteReleasePluginBinToolV2 deletes release plugin bintool v2.
func (s *Storage) deleteReleasePluginBinToolV2(ctx contextx.IContext, gen types.Generation, fileName string) error {
	err := s.daoRelease.Delete(ctx, types.ReleaseTypePluginBinToolV2,
		release.WithGeneration(gen),
		release.WithPlatform(platfmt.EmptyPlatform()),
		release.WithFileName(fileName),
	)
	if err != nil {
		return fmt.Errorf("failed to delete release plugin bintool v2: %w", err)
	}

	return nil
}

// getReleasePluginBinToolV2 gets release plugin bintool v2.
func (s *Storage) getReleasePluginBinToolV2(ctx contextx.IContext, gen types.Generation) (*types.ReleasePluginBinToolV2, error) {
	var (
		rls  *types.Release
		data *types.ReleasePluginBinToolV2
		err  error
	)

	rls, err = s.daoRelease.Get(ctx, types.ReleaseTypePluginBinToolV2,
		release.WithGeneration(gen),
		release.WithPlatform(platfmt.EmptyPlatform()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin bintool v2: %w", err)
	}

	data = &types.ReleasePluginBinToolV2{
		Release: *rls,
	}

	return data, nil
}

// existReleasePluginBinToolV2 checks if release plugin bintool v2 exists.
func (s *Storage) existReleasePluginBinToolV2(ctx contextx.IContext, gen types.Generation) (bool, error) {
	var (
		result bool
		err    error
	)

	result, err = s.daoRelease.Exist(ctx, types.ReleaseTypePluginBinToolV2,
		release.WithGeneration(gen),
		release.WithPlatform(platfmt.EmptyPlatform()),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check release plugin bintool v2: %w", err)
	}

	return result, nil
}
