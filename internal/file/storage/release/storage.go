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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the interface of release storage.
// nolint: interfacebloat
type IStorage interface {
	basestorage.Interface

	// GetRelease gets release by generation, type, platform and version.
	GetRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) (
		*types.Release, error)

	// GetReleaseCert gets release cert.
	GetReleaseCert(ctx context.Context) (*types.ReleaseCert, error)

	// GetReleaseBinTool gets release bintool.
	GetReleaseBinTool(ctx context.Context, gen types.Generation) (*types.ReleaseBinTool, error)

	// ExistRelease checks if release exists.
	ExistRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) (bool, error)

	// ExistReleaseAgent checks if release agent exists.
	ExistReleaseAgent(ctx context.Context, version string, plats ...platform.Platform) (bool, error)

	// ExistReleaseCert checks if release cert exists.
	ExistReleaseCert(ctx context.Context) (bool, error)

	// ExistReleaseBinTool checks if release bintool exists.
	ExistReleaseBinTool(ctx context.Context, gen types.Generation) (bool, error)

	// ListRelease lists release by page and conditions.
	ListRelease(ctx context.Context, page types.Page, conditions ...*types.ReleaseCondition) (
		[]*types.Release, int64, error)

	// UpsertManyRelease upserts many release.
	UpsertManyRelease(ctx context.Context, releases []*types.Release) error

	// UpsertReleaseCert upserts release cert.
	UpsertReleaseCert(ctx context.Context, cert types.ReleaseCert) error

	// UpsertReleaseBinTool upserts release bintool.
	UpsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) error

	// DeleteRelease deletes release.
	DeleteRelease(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error

	// DeleteReleaseCert deletes release cert.
	DeleteReleaseCert(ctx context.Context, fileName string) error

	// DeleteReleaseBinTool deletes release bintool.
	DeleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) error
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

// GetRelease gets release by generation, type, platform and version.
func (s *Storage) GetRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) (
	*types.Release, error) {

	return s.daoRelease.Get(ctx, gen, releaseType, plat, version)
}

// GetReleaseCert gets release cert.
func (s *Storage) GetReleaseCert(ctx context.Context) (*types.ReleaseCert, error) {
	r, err := s.daoRelease.Get(ctx, types.GenerationAll, types.ReleaseTypeCert, platform.EmptyPlatform(), "")
	if err != nil {
		return nil, err
	}

	return &types.ReleaseCert{
		FileName: r.FileName,
		MD5:      r.MD5,
	}, nil
}

// GetReleaseBinTool gets release bintool.
func (s *Storage) GetReleaseBinTool(ctx context.Context, gen types.Generation) (*types.ReleaseBinTool, error) {
	r, err := s.daoRelease.Get(ctx, gen, types.ReleaseTypeBinTool, platform.EmptyPlatform(), "")
	if err != nil {
		return nil, err
	}

	return &types.ReleaseBinTool{
		Generation: r.Generation,
		FileName:   r.FileName,
		MD5:        r.MD5,
	}, nil
}

// ExistRelease checks if release exists.
func (s *Storage) ExistRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) (bool, error) {

	return s.daoRelease.Exist(ctx, gen, releaseType, plat, version)
}

// ExistReleaseAgent checks if release agent exists.
func (s *Storage) ExistReleaseAgent(ctx context.Context, version string, plats ...platform.Platform) (bool, error) {
	count, err := s.daoRelease.Count(ctx,
		release.WithGeneration(types.Generation2),
		release.WithType(types.ReleaseTypeAgent),
		release.WithVersion(version),
		release.WithPlatform(plats...))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistReleaseCert checks if release cert exists.
func (s *Storage) ExistReleaseCert(ctx context.Context) (bool, error) {
	return s.daoRelease.Exist(ctx, types.GenerationAll, types.ReleaseTypeCert, platform.EmptyPlatform(), "")
}

// ExistReleaseBinTool checks if release bintool exists.
func (s *Storage) ExistReleaseBinTool(ctx context.Context, gen types.Generation) (bool, error) {
	return s.daoRelease.Exist(ctx, gen, types.ReleaseTypeBinTool, platform.EmptyPlatform(), "")
}

// ListRelease lists release.
func (s *Storage) ListRelease(ctx context.Context, page types.Page, conditions ...*types.ReleaseCondition) (
	[]*types.Release, int64, error) {

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
				release.WithPlatform(condition.ExactInclude.Platform...))
		}

		if condition.FuzzyInclude != nil {
			return nil, 0, errors.New("fuzzy include is not supported")
		}

		if condition.ExactExclude != nil {
			return nil, 0, errors.New("exact exclude is not supported")
		}

		if condition.FuzzyExclude != nil {
			return nil, 0, errors.New("fuzzy exclude is not supported")
		}
	}

	return s.daoRelease.List(ctx, page, opts...)
}

// UpsertManyRelease upserts many release.
func (s *Storage) UpsertManyRelease(ctx context.Context, releases []*types.Release) error {
	for _, rls := range releases {
		if rls != nil {
			rls.UpdatedAt = time.Now()
		}
	}

	return s.daoRelease.UpsertMany(ctx, releases...)
}

// UpsertReleaseCert upserts release cert.
func (s *Storage) UpsertReleaseCert(ctx context.Context, cert types.ReleaseCert) error {
	return s.daoRelease.UpsertMany(ctx, &types.Release{
		Generation: types.GenerationAll,
		Type:       types.ReleaseTypeCert,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   cert.FileName,
		MD5:        cert.MD5,
		UpdatedAt:  time.Now(),
	})
}

// UpsertReleaseBinTool upserts release bintool.
func (s *Storage) UpsertReleaseBinTool(ctx context.Context, bintool types.ReleaseBinTool) error {
	return s.daoRelease.UpsertMany(ctx, &types.Release{
		Generation: bintool.Generation,
		Type:       types.ReleaseTypeBinTool,
		Platform:   platform.EmptyPlatform(),
		Version:    "",
		FileName:   bintool.FileName,
		MD5:        bintool.MD5,
		UpdatedAt:  time.Now(),
	})
}

// DeleteRelease deletes release.
func (s *Storage) DeleteRelease(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	return s.daoRelease.Delete(ctx, gen, releaseType, plat, version)
}

// DeleteReleaseCert deletes release cert.
func (s *Storage) DeleteReleaseCert(ctx context.Context, fileName string) error {
	return s.daoRelease.Delete(ctx, types.GenerationAll, types.ReleaseTypeCert, platform.EmptyPlatform(), fileName)
}

// DeleteReleaseBinTool deletes release bintool.
func (s *Storage) DeleteReleaseBinTool(ctx context.Context, gen types.Generation, fileName string) error {
	return s.daoRelease.Delete(ctx, gen, types.ReleaseTypeBinTool, platform.EmptyPlatform(), fileName)
}
