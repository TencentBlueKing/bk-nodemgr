/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release data models.
package release

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler release handler interface.
type IHandler interface {
	// Get gets a release.
	Get(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) (
		*types.Release, error)

	// Exist checks if a release exists.
	Exist(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) (
		bool, error)

	// List lists releases.
	List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Release, int64, error)

	// Count counts releases.
	Count(ctx context.Context, opts ...OptFn) (int64, error)

	// UpsertMany upserts a release.
	UpsertMany(ctx context.Context, releases ...*types.Release) error

	// Delete deletes a release.
	Delete(ctx context.Context,
		gen types.Generation,
		releaseType types.ReleaseType,
		plat platform.Platform,
		version string) error
}

type handler struct {
	logger logger.Logger
	dao    *dao
}

// New create a new trigger handler.
func New(client *mongo.Database, logger logger.Logger) IHandler {
	return &handler{
		logger: logger,
		dao:    newDao(client, logger),
	}
}

// Get gets a release.
func (h *handler) Get(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) (
	*types.Release, error) {

	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	conditions := []OptFn{
		WithGeneration(gen),
		WithType(releaseType),
		WithCPUArch(plat.Arch),
		WithOSType(plat.OS),
		WithVersion(version),
	}
	filter := base.AliveFilter()
	for _, opt := range conditions {
		filter = opt(filter)
	}

	data, err := h.dao.Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertReleaseToTypes(data), nil
}

// Exist checks if a release exists.
func (h *handler) Exist(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) (
	bool, error) {

	if ctx == nil {
		return false, errors.New("ctx is nil")
	}

	conditions := []OptFn{
		WithGeneration(gen),
		WithType(releaseType),
		WithCPUArch(plat.Arch),
		WithOSType(plat.OS),
		WithVersion(version),
	}
	filter := base.AliveFilter()
	for _, opt := range conditions {
		filter = opt(filter)
	}

	return h.dao.Exist(ctx, filter)
}

// List lists releases.
func (h *handler) List(ctx context.Context, page types.Page, opts ...OptFn) ([]*types.Release, int64, error) {
	if ctx == nil {
		return nil, 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	releases, err := h.dao.List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.Release, len(releases))
	for idx, release := range releases {
		data[idx] = convertReleaseToTypes(release)
	}

	return data, num, nil
}

// Count counts releases.
func (h *handler) Count(ctx context.Context, opts ...OptFn) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(ctx, filter)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// Upsert upserts a release.
func (h *handler) UpsertMany(ctx context.Context, releases ...*types.Release) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if releases == nil || len(releases) == 0 {
		return errors.New("release is nil")
	}

	data := make([]*Release, len(releases))
	for idx, release := range releases {
		data[idx] = convertReleaseFromTypes(release)
	}

	return h.dao.upsertMany(ctx, data)
}

// Delete deletes a release.
func (h *handler) Delete(ctx context.Context,
	gen types.Generation,
	releaseType types.ReleaseType,
	plat platform.Platform,
	version string) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	conditions := []OptFn{
		WithGeneration(gen),
		WithType(releaseType),
		WithCPUArch(plat.Arch),
		WithOSType(plat.OS),
		WithVersion(version),
	}
	filter := base.AliveFilter()
	for _, opt := range conditions {
		filter = opt(filter)
	}

	return h.dao.DeleteMany(ctx, filter)
}

func convertReleaseToTypes(release *Release) *types.Release {
	return &types.Release{
		Generation: types.Generation(release.Generation),
		Type:       types.ReleaseType(release.Type),
		Version:    release.Version,
		Platform: platform.Platform{
			Arch: release.CPUArch,
			OS:   release.OSType,
		},
		Labels:      release.Labels,
		ChangeLogEN: release.ChangeLogEN,
		ChangeLogZH: release.ChangeLogZH,
		FileName:    release.FileName,
		UpstreamDir: release.UpstreamDir,
		LocalDir:    release.LocalDir,
		MD5:         release.MD5,
		UpdatedAt:   release.UpdatedAt,
	}
}

func convertReleaseFromTypes(release *types.Release) *Release {
	return &Release{
		Generation:  int64(release.Generation),
		Type:        string(release.Type),
		Version:     release.Version,
		CPUArch:     release.Platform.Arch,
		OSType:      release.Platform.OS,
		Labels:      release.Labels,
		ChangeLogEN: release.ChangeLogEN,
		ChangeLogZH: release.ChangeLogZH,
		FileName:    release.FileName,
		UpstreamDir: release.UpstreamDir,
		LocalDir:    release.LocalDir,
		MD5:         release.MD5,
		UpdatedAt:   release.UpdatedAt,
	}
}
