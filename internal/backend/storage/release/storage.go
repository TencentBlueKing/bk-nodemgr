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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the release storage interface.
// nolint: interfacebloat
type IStorage interface {
	basestorage.Interface

	// GetRelease gets release by generation, release type, platform and version.
	GetRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) (*types.Release, error)

	// ListRelease lists release by page and conditions.
	ListRelease(ctx context.Context, releaseType types.ReleaseType, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.Release, int64, error)

	// DistinctRelease distincts release by conditions.
	DistinctRelease(ctx context.Context, releaseType types.ReleaseType, distinctField types.ReleaseDistinctField,
		conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error)

	// CountRelease counts release by conditions.
	CountRelease(ctx context.Context, releaseType types.ReleaseType, conditions ...*types.ReleaseCondition) (int64, error)

	// SetReleaseLabels sets release labels.
	SetReleaseLabels(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string,
		labels []string) error

	// EnableRelease enables release active by generation, release type, platform and version.
	EnableRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// DisableRelease disables release disactive by generation, release type, platform and version.
	DisableRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// SetAsDefaultRelease sets the release as default.
	SetAsDefaultRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// CancelAsDefaultRelease cancels the release as default.
	CancelAsDefaultRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// 	DeleteRelease deletes release by generation, release type, platform and version.
	DeleteRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	IAgent
	IProxy
}

// StorageName defines the storage name.
const StorageName = "release"

// NewStorage creates a new release storage.
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed: %v", err)
		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoRelease release.IHandler
}

func (s *Storage) initDao() error {
	s.daoRelease = release.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoRelease == nil {
		return errors.New("dao release is nil")
	}

	return nil
}

// GetRelease gets release by generation, release type, platform and version.
func (s *Storage) GetRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) (data *types.Release, err error) {

	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	data, err = s.daoRelease.Get(ctx, releaseType,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ListRelease lists release by page and conditions.
func (s *Storage) ListRelease(ctx context.Context, releaseType types.ReleaseType, page types.Page, conditions ...*types.ReleaseCondition) (
	results []*types.Release, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list")
	defer metric.End(err)

	var opts []release.OptFn
	if opts, err = convertReleaseConditionsToOptions(conditions...); err != nil {
		return nil, 0, err
	}

	if results, num, err = s.daoRelease.List(ctx, releaseType, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// DistinctRelease distincts release by conditions.
func (s *Storage) DistinctRelease(ctx context.Context, releaseType types.ReleaseType, distinctField types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (data *types.ReleaseDistinctResult, err error) {

	// record metric.
	metric := s.metric().Start("distinct")
	defer metric.End(err)

	opts, err := convertReleaseConditionsToOptions(conditions...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert release conditions to options: %w", err)
	}

	data = new(types.ReleaseDistinctResult)

	gp := gopool.NewPool()
	if distinctField.OSType {
		gp.Go(func() error {
			var err error
			data.OSType, err = s.daoRelease.DistinctOsType(ctx, releaseType, opts...)

			return err
		})
	}
	if distinctField.CPUArch {
		gp.Go(func() error {
			var err error
			data.CPUArch, err = s.daoRelease.DistinctCPUArch(ctx, releaseType, opts...)

			return err
		})
	}
	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

// CountRelease counts release by conditions.
func (s *Storage) CountRelease(ctx context.Context, releaseType types.ReleaseType, conditions ...*types.ReleaseCondition) (num int64, err error) {
	// record metric.
	metric := s.metric().Start("count")
	defer metric.End(err)

	var opts []release.OptFn
	if opts, err = convertReleaseConditionsToOptions(conditions...); err != nil {
		return 0, err
	}

	if num, err = s.daoRelease.Count(ctx, releaseType, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// SetReleaseLabels sets release labels.
func (s *Storage) SetReleaseLabels(
	ctx context.Context, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string, labels []string) (err error) {

	// record metric.
	metric := s.metric().Start("set_labels")
	defer metric.End(err)

	if err = s.daoRelease.SetLabels(ctx, releaseType, gen, plat, version, labels...); err != nil {
		return err
	}

	return nil
}

// EnableRelease enables release active by generation, release type, platform and version.
func (s *Storage) EnableRelease(
	ctx context.Context, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) (err error) {

	// record metric.
	metric := s.metric().Start("enable")
	defer metric.End(err)

	err = s.daoRelease.SetEnabled(ctx, releaseType, true,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return err
	}

	return nil
}

// DisableRelease disables release active by generation, release type, platform and version.
func (s *Storage) DisableRelease(
	ctx context.Context, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) (err error) {

	// record metric.
	metric := s.metric().Start("disable")
	defer metric.End(err)

	// cancel this release as default.
	err = s.daoRelease.SetAsDefault(ctx, releaseType, false,
		release.WithVersion(version),
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	)

	if err != nil {
		return fmt.Errorf("failed to cancel this release as default. platform(%s), version(%s)",
			plat.String(), version)
	}

	err = s.daoRelease.SetEnabled(ctx, releaseType, false,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return err
	}

	return nil
}

// SetAsDefaultRelease sets the release as default.
func (s *Storage) SetAsDefaultRelease(
	ctx context.Context, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) (err error) {

	// record metric.
	metric := s.metric().Start("set_as_default")
	defer metric.End(err)

	// cancel all version as-default in this platform.
	err = s.daoRelease.CancelPlatformDefault(ctx, releaseType,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
	)
	if err != nil {
		return fmt.Errorf("failed to cancel all version in this platform as default. platform(%s)", plat.String())
	}

	err = s.daoRelease.SetAsDefault(ctx, releaseType, true,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return err
	}

	return nil
}

// CancelAsDefaultRelease cancels the release as default.
func (s *Storage) CancelAsDefaultRelease(
	ctx context.Context, gen types.Generation, releaseType types.ReleaseType,
	plat platform.Platform, version string) (err error) {

	// record metric.
	metric := s.metric().Start("cancel_as_default")
	defer metric.End(err)

	err = s.daoRelease.SetAsDefault(ctx, releaseType, false,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRelease deletes the release.
func (s *Storage) DeleteRelease(ctx context.Context, gen types.Generation, releaseType types.ReleaseType, plat platform.Platform, version string,
) error {

	var err error

	// record metric.
	metric := s.metric().Start("delete")
	defer metric.End(err)
	err = s.daoRelease.Delete(ctx, releaseType,
		release.WithGeneration(gen),
		release.WithPlatform(plat),
		release.WithVersion(version),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
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
