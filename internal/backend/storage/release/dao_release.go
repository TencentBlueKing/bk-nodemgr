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

package release

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// getRelease gets release by generation, release type, platform and version.
func (s *Storage) getRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version, name string) (
	*types.Release, error) {

	return s.daoRelease.Get(nCtx, releaseType,
		release.WithName(name),
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
	if distinctField.Name {
		gp.Go(func() error {
			var err error
			data.Name, err = s.daoRelease.DistinctName(nCtx, releaseType, opts...)

			return err
		})
	}
	if distinctField.Version {
		gp.Go(func() error {
			var err error
			data.Version, err = s.daoRelease.DistinctVersion(nCtx, releaseType, opts...)

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

// setReleaseLabelsMany sets release labels.
func (s *Storage) setReleaseLabelsMany(
	nCtx contextx.IContext, releaseType types.ReleaseType, labels []string, conditions ...*types.ReleaseCondition) error {

	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	return s.daoRelease.SetLabels(nCtx, releaseType, labels, opts...)
}

// enableRelease enables release active.
func (s *Storage) enableRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version, name string) error {

	opts := []release.OptFn{
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	}

	if err := s.daoRelease.SetEnabled(nCtx, releaseType, true, opts...); err != nil {
		return fmt.Errorf("failed to enable release: %w", err)
	}

	return nil
}

// disableRelease disables release active by generation, release type, platform and version.
func (s *Storage) disableRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version, name string) error {

	opts := []release.OptFn{
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	}

	// cancel this release as default.
	if err := s.daoRelease.SetAsDefault(nCtx, releaseType, false, opts...); err != nil {
		return fmt.Errorf("failed to cancel release as default: %w", err)
	}

	if err := s.daoRelease.SetEnabled(nCtx, releaseType, false, opts...); err != nil {
		return fmt.Errorf("failed to disable release: %w", err)
	}

	return nil
}

// setAsDefaultRelease sets the release as default.
func (s *Storage) setAsDefaultRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version, name string) error {

	// cancel all version as-default in this platform.
	if err := s.daoRelease.CancelPlatformDefault(nCtx, releaseType,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	); err != nil {
		return fmt.Errorf("failed to cancel all version in platform(%s) as default: %w", plat.String(), err)
	}

	if err := s.daoRelease.SetAsDefault(nCtx, releaseType, true,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to set release as default: %w", err)
	}

	return nil
}

// cancelAsDefaultRelease cancels the release as default.
func (s *Storage) cancelAsDefaultRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version, name string) error {

	if err := s.daoRelease.SetAsDefault(nCtx, releaseType, false,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to cancel release as default: %w", err)
	}

	return nil
}

func (s *Storage) setReleaseHidden(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform,
	version, name string, isHidden bool) error {

	if err := s.daoRelease.SetHidden(nCtx, releaseType, isHidden,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to set release hidden: %w", err)
	}

	return nil
}

// deleteRelease deletes the release.
func (s *Storage) deleteRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version, name string) error {

	if err := s.daoRelease.Delete(nCtx, releaseType,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	); err != nil {
		return fmt.Errorf("failed to delete release: %w", err)
	}

	return nil
}

func (s *Storage) existRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, version string, name string) (bool, error) {

	exist, err := s.daoRelease.Exist(nCtx, releaseType,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check release exist: %w", err)
	}

	return exist, nil
}

func (s *Storage) getReleaseDefaultVersion(
	nCtx contextx.IContext, releaseType types.ReleaseType, gen types.Generation, plat platfmt.Platform, name string) (string, error) {

	rls, err := s.daoRelease.Get(nCtx, releaseType,
		release.WithName(name),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithAsDefault(true),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get release default: %w", err)
	}

	return rls.Version, nil
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
