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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/opluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// existReleaseOfficialPlugin checks if release plugin exists.
func (s *Storage) existReleaseOfficialPlugin(
	ctx contextx.IContext, pluginName string, version string, plats ...platform.Platform) (bool, error) {

	fileNames := make([]string, 0, len(plats))
	for _, plat := range plats {
		pluginFileName, err := opluginpkg.FormatPkgName(pluginName, types.ReleaseTypeOfficialPlugin, types.Generation2, plat, version)
		if err != nil {
			return false, err
		}

		fileNames = append(fileNames, pluginFileName)
	}

	num, err := s.daoRelease.Count(ctx, types.ReleaseTypeOfficialPlugin, release.WithFileName(fileNames...))
	if err != nil {
		return false, err
	}

	result := num > 0

	return result, nil
}

// upsertManyReleaseOfficialPlugin upsert many release.
func (s *Storage) upsertManyReleaseOfficialPlugin(ctx contextx.IUserContext, releaseOfficialPlugins []*types.ReleaseOfficialPlugin) error {
	var err error

	releases := make([]*types.Release, 0, len(releaseOfficialPlugins))
	for _, rls := range releaseOfficialPlugins {
		if rls == nil {
			continue
		}

		rls.UpdatedAt = time.Now()
		rls.Operator = ctx.BKUsername()
		rls.AdditionInfo, err = conv.StructToMap(rls.ReleaseAdditionInfoOfficialPlugin)
		if err != nil {
			return fmt.Errorf("failed to upsert many release agent: %v", err)
		}

		releases = append(releases, &rls.Release)
	}

	err = s.daoRelease.UpsertMany(ctx, types.ReleaseTypeOfficialPlugin, releases...)
	if err != nil {
		return fmt.Errorf("failed to upsert many release agent: %v", err)
	}

	return nil
}

func (s *Storage) getReleaseOfficialPlugin(ctx contextx.IContext, name string, gen types.Generation, plat platform.Platform, version string) (
	*types.ReleaseOfficialPlugin, error) {

	var (
		rls *types.Release
		err error
	)

	rls, err = s.daoRelease.Get(ctx, types.ReleaseTypeOfficialPlugin,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get release official plugin: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoOfficialPlugin)
	if err = conv.MapToStruct(rls.AdditionInfo, additionInfo); err != nil {
		return nil, fmt.Errorf("failed to get release official plugin: %v", err)
	}

	return &types.ReleaseOfficialPlugin{
		Release:                           *rls,
		ReleaseAdditionInfoOfficialPlugin: *additionInfo,
	}, nil
}
