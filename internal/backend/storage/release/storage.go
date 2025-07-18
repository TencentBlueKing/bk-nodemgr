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
package release

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the release storage interface.
type IStorage interface {
	basestorage.Interface

	// ListRelease lists release by page and conditions.
	ListRelease(ctx context.Context, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.Release, int64, error)

	// CountRelease counts release by conditions.
	CountRelease(ctx context.Context, conditions ...*types.ReleaseCondition) (int64, error)

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
}

// StorageName defines the storage name.
const StorageName = "release"

// NewStorage creates a new release storage.
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
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
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

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

// ListRelease lists release by page and conditions.
func (s *Storage) ListRelease(ctx context.Context, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.Release, int64, error) {

	opts, err := convertReleaseconditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	return s.daoRelease.List(ctx, page, opts...)
}

// CountRelease counts release by conditions.
func (s *Storage) CountRelease(ctx context.Context, conditions ...*types.ReleaseCondition) (int64, error) {
	opts, err := convertReleaseconditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	return s.daoRelease.Count(ctx, opts...)
}

// SetReleaseLabels sets release labels.
func (s *Storage) SetReleaseLabels(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string,
	labels []string) error {

	return s.daoRelease.SetLabels(ctx, gen, releaseType, plat, version, labels...)
}

// EnableRelease enables release active by generation, release type, platform and version.
func (s *Storage) EnableRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	return s.daoRelease.SetEnabled(ctx, gen, releaseType, plat, version, true)
}

// DisableRelease disables release active by generation, release type, platform and version.
func (s *Storage) DisableRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	return s.daoRelease.SetEnabled(ctx, gen, releaseType, plat, version, false)
}

// SetAsDefaultRelease sets the release as default.
func (s *Storage) SetAsDefaultRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	return s.daoRelease.SetAsDefault(ctx, gen, releaseType, plat, version, true)
}

// CancelAsDefaultRelease cancels the release as default.
func (s *Storage) CancelAsDefaultRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	return s.daoRelease.SetAsDefault(ctx, gen, releaseType, plat, version, false)
}

// DeleteRelease deletes the release.
func (s *Storage) DeleteRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	return s.daoRelease.Delete(ctx, gen, releaseType, plat, version)
}

func convertReleaseconditionsToOptions(conditions ...*types.ReleaseCondition) ([]release.OptFn, error) {
	opts := make([]release.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				release.WithGeneration(condition.ExactInclude.Generation...),
				release.WithType(condition.ExactInclude.Type...),
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

// ConvertNodeRoleToReleaseType convert role to release type.
func ConvertNodeRoleToReleaseType(role types.NodeRole) (types.ReleaseType, error) {
	switch role {
	case types.NodeRoleAgent:
		return types.ReleaseTypeAgent, nil
	case types.NodeRoleProxy:
		return types.ReleaseTypeProxy, nil
	default:
		return "", fmt.Errorf("invalid node role. role(%s)", role)
	}
}
