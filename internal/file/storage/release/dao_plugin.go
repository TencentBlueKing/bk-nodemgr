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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/pluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// existReleasePlugin checks if release plugin exists.
func (s *Storage) existReleasePlugin(
	ctx contextx.IContext, pluginName string, version string, plats ...platfmt.Platform) (bool, error) {

	fileNames := make([]string, 0, len(plats))
	for _, plat := range plats {
		pluginFileName, err := pluginpkg.FormatPkgName(pluginName, types.ReleaseTypePlugin, types.Generation2, plat, version)
		if err != nil {
			return false, err
		}

		fileNames = append(fileNames, pluginFileName)
	}

	num, err := s.daoRelease.Count(ctx, types.ReleaseTypePlugin, release.WithFileName(fileNames...))
	if err != nil {
		return false, err
	}

	result := num > 0

	return result, nil
}

// upsertManyReleasePlugin upsert many release.
func (s *Storage) upsertManyReleasePlugin(ctx contextx.IContext, releasePlugins []*types.ReleasePlugin) error {
	var err error

	releases := make([]*types.Release, 0, len(releasePlugins))
	for _, rls := range releasePlugins {
		if rls == nil {
			continue
		}

		rls.UpdatedAt = time.Now()
		rls.Operator = ctx.BKUsername()
		rls.AdditionInfo, err = conv.StructToMap(rls.ReleaseAdditionInfoPlugin)
		if err != nil {
			return fmt.Errorf("failed to upsert many release agent: %v", err)
		}

		releases = append(releases, &rls.Release)
	}

	err = s.daoRelease.UpsertMany(ctx, types.ReleaseTypePlugin, releases...)
	if err != nil {
		return fmt.Errorf("failed to upsert many release agent: %v", err)
	}

	return nil
}

func (s *Storage) getReleasePlugin(ctx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	*types.ReleasePlugin, error) {

	var (
		rls *types.Release
		err error
	)

	rls, err = s.daoRelease.Get(ctx, types.ReleaseTypePlugin,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release plugin: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoPlugin)
	if err = conv.MapToStruct(rls.AdditionInfo, additionInfo); err != nil {
		return nil, fmt.Errorf("failed to get release plugin: %v", err)
	}

	return &types.ReleasePlugin{
		Release:                   *rls,
		ReleaseAdditionInfoPlugin: *additionInfo,
	}, nil
}
