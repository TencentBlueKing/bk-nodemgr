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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getRelease gets release by generation, release type, platform and version.
func (s *Storage) getRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) (*types.Release, error) {

	return s.daoRelease.Get(nCtx, releaseType,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
}

// listRelease lists release by page and conditions.
func (s *Storage) listRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.Release, int64, error) {

	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	return s.daoRelease.List(nCtx, releaseType, page, opts...)
}

// distinctRelease distincts release by conditions.
func (s *Storage) distinctRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, distinctField types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {

	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	data := new(types.ReleaseDistinctResult)

	gp := gopool.NewPool()
	if distinctField.OSType {
		gp.Go(func() error {
			var err error
			data.OSType, err = s.daoRelease.DistinctOsType(nCtx, releaseType, opts...)

			return err
		})
	}
	if distinctField.CPUArch {
		gp.Go(func() error {
			var err error
			data.CPUArch, err = s.daoRelease.DistinctCPUArch(nCtx, releaseType, opts...)

			return err
		})
	}
	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

// countRelease counts release by conditions.
func (s *Storage) countRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, conditions ...*types.ReleaseCondition) (
	int64, error) {

	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return 0, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	return s.daoRelease.Count(nCtx, releaseType, opts...)
}

// setReleaseLabels sets release labels.
func (s *Storage) setReleaseLabels(
	nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string, labels []string) error {

	return s.daoRelease.SetLabels(nCtx, releaseType, gen, plat, version, labels...)
}

// setManyReleaseLabels sets release labels.
func (s *Storage) setManyReleaseLabels(
	nCtx contextx.IContext, releaseType types.ReleaseType, gens []types.Generation,
	plats []platform.Platform, versions []string, labels []string) error {

	return s.daoRelease.SetManyLabels(nCtx, releaseType, gens, plats, versions, labels...)
}

// enableRelease enables release active by generation, release type, platform and version.
func (s *Storage) enableRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) error {

	return s.daoRelease.SetEnabled(nCtx, releaseType, true,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
}

// disableRelease disables release active by generation, release type, platform and version.
func (s *Storage) disableRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) error {

	// cancel this release as default.
	if err := s.daoRelease.SetAsDefault(nCtx, releaseType, false,
		release.WithVersion(version),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	); err != nil {
		return fmt.Errorf("failed to cancel this platform(%s) and version(%s) release as default: %w",
			plat.String(), version, err)
	}

	if err := s.daoRelease.SetEnabled(nCtx, releaseType, false,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to disable platform(%s) and version(%s) release: %w",
			plat.String(), version, err)
	}

	return nil
}

// setAsDefaultRelease sets the release as default.
func (s *Storage) setAsDefaultRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) error {

	// cancel all version as-default in this platform.
	if err := s.daoRelease.CancelPlatformDefault(nCtx, releaseType,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	); err != nil {
		return fmt.Errorf("failed to cancel all version in this platform(%s) as default: %w", plat.String(), err)
	}

	if err := s.daoRelease.SetAsDefault(nCtx, releaseType, true,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to set platform(%s) and version(%s) release as default: %w",
			plat.String(), version, err)
	}

	return nil
}

// cancelAsDefaultRelease cancels the release as default.
func (s *Storage) cancelAsDefaultRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) error {

	return s.daoRelease.SetAsDefault(nCtx, releaseType, false,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
}

// deleteRelease deletes the release.
func (s *Storage) deleteRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) error {

	return s.daoRelease.Delete(nCtx, releaseType,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
}

func convertReleaseConditionsToOptions(conditions ...*types.ReleaseCondition) ([]release.OptFn, error) {
	opts := make([]release.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				release.WithFileName(condition.ExactInclude.FileName...),
				release.WithGeneration(condition.ExactInclude.Generation...),
				release.WithVersion(condition.ExactInclude.Version...),
				release.WithPlatform(condition.ExactInclude.Platform...),
				release.WithEnabled(condition.ExactInclude.Enabled...),
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
