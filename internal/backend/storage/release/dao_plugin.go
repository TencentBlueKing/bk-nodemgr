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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getEnabledReleasePlugin gets release by generation, release type, platform and version.
func (s *Storage) getEnabledReleasePlugin(
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleasePlugin, error) {

	rls, err := s.daoRelease.Get(nCtx, types.ReleaseTypePlugin,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
		release.WithEnabled(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled release plugin: %w", err)
	}

	additionInfo := new(types.ReleaseAdditionInfoPlugin)
	err = conv.MapToStruct(rls.AdditionInfo, additionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert addition info to struct: %w", err)
	}

	releasePlugin := &types.ReleasePlugin{
		Release:                   *rls,
		ReleaseAdditionInfoPlugin: *additionInfo,
	}

	return releasePlugin, nil
}

func (s *Storage) listReleasePlugin(
	nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error) {

	rls, total, err := s.listRelease(nCtx, types.ReleaseTypePlugin, page, conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list release plugin: %w", err)
	}

	releasePlugins := make([]*types.ReleasePlugin, len(rls))
	for idx, r := range rls {
		additionInfo := new(types.ReleaseAdditionInfoPlugin)
		err = conv.MapToStruct(r.AdditionInfo, additionInfo)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert addition info to struct: %w", err)
		}

		releasePlugins[idx] = &types.ReleasePlugin{
			Release:                   *r,
			ReleaseAdditionInfoPlugin: *additionInfo,
		}
	}

	return releasePlugins, total, nil
}

func (s *Storage) disableReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) error {
	// cancel this release as default.
	if err := s.daoRelease.SetAsDefault(nCtx, types.ReleaseTypePlugin, false,
		release.WithName(name),
		release.WithVersion(version),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	); err != nil {
		return fmt.Errorf("failed to cancel this platform(%s) and version(%s) release as default: %w",
			plat.String(), version, err)
	}

	if err := s.daoRelease.SetEnabled(nCtx, types.ReleaseTypePlugin, false,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to disable platform(%s) and version(%s) release: %w",
			plat.String(), version, err)
	}

	return nil
}

func (s *Storage) cancelAsDefaultReleasePlugin(
	nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) error {

	err := s.daoRelease.SetAsDefault(nCtx, types.ReleaseTypePlugin, false,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)

	if err != nil {
		return fmt.Errorf("failed to cancel release plugin as default: %w", err)
	}

	return nil
}

func (s *Storage) setAsDefaultReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) error {
	// cancel all version as-default in this platform.
	if err := s.daoRelease.CancelPlatformDefault(nCtx, types.ReleaseTypePlugin,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	); err != nil {
		return fmt.Errorf("failed to cancel all version in this platform(%s) as default: %w", plat.String(), err)
	}

	if err := s.daoRelease.SetAsDefault(nCtx, types.ReleaseTypePlugin, true,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to set platform(%s) and version(%s) release as default: %w",
			plat.String(), version, err)
	}

	return nil
}

func (s *Storage) enableReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) error {
	err := s.daoRelease.SetEnabled(nCtx, types.ReleaseTypePlugin, true,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return fmt.Errorf("failed to enable release plugin: %w", err)
	}

	return nil
}

func (s *Storage) deleteReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) error {
	err := s.daoRelease.Delete(nCtx, types.ReleaseTypePlugin,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)

	if err != nil {
		return fmt.Errorf("failed to delete release plugin: %w", err)
	}

	return nil
}

func (s *Storage) countReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return 0, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	num, err := s.daoRelease.Count(nCtx, types.ReleaseTypePlugin, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count release plugin: %w", err)
	}

	return num, nil
}

func (s *Storage) existReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (bool, error) {
	exist, err := s.daoRelease.Exist(nCtx, types.ReleaseTypePlugin,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return false, fmt.Errorf("failed to exist release plugin: %w", err)
	}

	return exist, nil
}
