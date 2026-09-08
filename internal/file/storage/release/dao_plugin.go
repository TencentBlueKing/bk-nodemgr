/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package release provides the release storage interface.
// nolint: nonamedreturns
package release

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/pluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) listReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.ReleasePlugin, int64, error) {

	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	releases, total, err := s.daoRelease.List(
		nCtx,
		types.ReleaseTypePlugin,
		page,
		opts...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list release plugin: %w", err)
	}

	result := make([]*types.ReleasePlugin, len(releases))
	for i, item := range releases {
		additionInfo := new(types.ReleaseAdditionInfoPlugin)
		if err := conv.MapToStruct(item.AdditionInfo, additionInfo); err != nil {
			return nil, 0, fmt.Errorf("failed to list release plugin, failed to convert addition info to struct: %w", err)
		}

		result[i] = &types.ReleasePlugin{
			Release:                   *item,
			ReleaseAdditionInfoPlugin: *additionInfo,
		}
	}

	return result, total, nil
}

// existReleasePlugin checks if release plugin exists.
func (s *Storage) existReleasePlugin(nCtx contextx.IContext, pluginName string, version string, plats ...platfmt.Platform) (bool, error) {
	fileNames := make([]string, 0, len(plats))
	for _, plat := range plats {
		pluginFileName, err := pluginpkg.FormatPkgFileName(pluginName, types.ReleaseTypePlugin, types.Generation2, plat, version)
		if err != nil {
			return false, fmt.Errorf("failed to check exist release plugin: %w", err)
		}

		fileNames = append(fileNames, pluginFileName)
	}

	exist, err := s.daoRelease.Exist(nCtx, types.ReleaseTypePlugin, release.WithFileName(fileNames...))
	if err != nil {
		return false, fmt.Errorf("failed to check exist release plugin: %w", err)
	}

	return exist, nil
}

// upsertManyReleasePlugin upsert many release.
func (s *Storage) upsertManyReleasePlugin(nCtx contextx.IContext, releasePlugins []*types.ReleasePlugin) error {
	releases := make([]*types.Release, 0, len(releasePlugins))
	for _, rls := range releasePlugins {
		if rls == nil {
			continue
		}

		additionInfo, err := conv.StructToMap(rls.ReleaseAdditionInfoPlugin)
		if err != nil {
			return fmt.Errorf("failed to upsert many release plugin, failed to convert addition info to map: %w", err)
		}

		rls.UpdatedAt = time.Now()
		rls.Operator = nCtx.BKUsername()
		rls.AdditionInfo = additionInfo

		releases = append(releases, &rls.Release)
	}

	if err := s.daoRelease.UpsertMany(nCtx, types.ReleaseTypePlugin, releases...); err != nil {
		return fmt.Errorf("failed to upsert many release plugin: %w", err)
	}

	return nil
}

func (s *Storage) getReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	*types.ReleasePlugin, error) {

	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypePlugin,
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
		return nil, fmt.Errorf("failed to get release plugin, failed to convert addition info to struct: %w", err)
	}

	return &types.ReleasePlugin{
		Release:                   *rls,
		ReleaseAdditionInfoPlugin: *additionInfo,
	}, nil
}

func convertReleaseConditionsToOptions(conditions ...*types.ReleaseCondition) ([]release.OptFn, error) {
	opts := make([]release.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				release.WithName(condition.ExactInclude.Name...),
				release.WithFileName(condition.ExactInclude.FileName...),
				release.WithGeneration(condition.ExactInclude.Generation...),
				release.WithVersion(condition.ExactInclude.Version...),
				release.WithPlatform(condition.ExactInclude.Platform...),
				release.WithEnabled(condition.ExactInclude.Enabled...),
				release.WithIsHidden(condition.ExactInclude.IsHidden...),
				release.WithIsShared(condition.ExactInclude.IsShared...),
				release.WithIsSynced(condition.ExactInclude.IsSynced...),
				release.WithAsDefault(condition.ExactInclude.AsDefault...))
		}

		if condition.FuzzyInclude != nil {
			return nil, errors.New("fuzzy include is not supported")
		}

		if condition.ExactExclude != nil {
			return nil, errors.New("exact exclude is not supported")
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
