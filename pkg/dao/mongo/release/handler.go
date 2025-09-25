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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler release Handler interface.
type IHandler interface {
	// Get gets a release.
	Get(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) (*types.Release, error)

	// Exist checks if a release exists.
	Exist(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) (bool, error)

	// List lists releases.
	List(ctx context.Context, releaseType types.ReleaseType, page types.Page, opts ...OptFn) ([]*types.Release, int64, error)

	// SetLabels sets a release's labels.
	SetLabels(ctx context.Context, releaseType types.ReleaseType, gen types.Generation, plat platform.Platform, version string, labels ...string) error

	// Count counts releases.
	Count(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) (int64, error)

	// UpsertMany upserts a release.
	UpsertMany(ctx context.Context, releaseType types.ReleaseType, releases ...*types.Release) error

	// Delete deletes a release.
	Delete(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) error

	ISwitcher
	IDistinctor
}

// ISwitcher defines the default interface.
type ISwitcher interface {
	// SetEnabled sets a release's enabled.
	SetEnabled(ctx context.Context, releaseType types.ReleaseType, enabled bool, opts ...OptFn) error

	// SetAsDefault sets a release's asDefault.
	SetAsDefault(ctx context.Context, releaseType types.ReleaseType, asDefault bool, opts ...OptFn) error

	// CancelPlatformDefault cancel a release's all version asDefault by one platform.
	CancelPlatformDefault(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) error
}

// IDistinctor defines the distinctor interface.
type IDistinctor interface {
	// DistinctOsType distincts os types.
	DistinctOsType(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) ([]string, error)

	// DistinctCPUArch distincts cpu archs.
	DistinctCPUArch(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) ([]string, error)
}

var _ IHandler = &Handler{}

// Handler implements IHandler.
type Handler struct {
	client *mongo.Database
	logger logger.ILogger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) releaseTypeDao(category types.ReleaseType) *dao {
	tableName := TableName(string(category))
	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tableName, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure release indexes: %v", errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the releaseTypeDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new trigger Handler.
func New(client *mongo.Database, logger logger.ILogger) *Handler {
	return &Handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Get gets a release.
func (h *Handler) Get(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) (*types.Release, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := h.releaseTypeDao(releaseType).Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertReleaseToTypes(data), nil
}

// Exist checks if a release exists.
func (h *Handler) Exist(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) (bool, error) {
	if ctx == nil {
		return false, errors.New("ctx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).Exist(ctx, filter)
}

// List lists releases.
func (h *Handler) List(ctx context.Context, releaseType types.ReleaseType, page types.Page, opts ...OptFn) ([]*types.Release, int64, error) {
	if ctx == nil {
		return nil, 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.releaseTypeDao(releaseType).Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	releases, err := h.releaseTypeDao(releaseType).List(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.Release, len(releases))
	for idx, release := range releases {
		data[idx] = convertReleaseToTypes(release)
	}

	return data, num, nil
}

// SetLabels sets a release's labels.
func (h *Handler) SetLabels(ctx context.Context, releaseType types.ReleaseType, gen types.Generation, plat platform.Platform, version string,
	labels ...string) error {

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	conditions := []OptFn{
		WithGeneration(gen),
		WithType(releaseType),
		WithCPUArch(string(plat.Arch)),
		WithOSType(string(plat.OS)),
		WithVersion(version),
	}
	filter := base.AliveFilter()
	for _, opt := range conditions {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(ctx, filter, FieldKeyLabels, labels)
}

// SetEnabled sets a release's enabled.
func (h *Handler) SetEnabled(ctx context.Context, releaseType types.ReleaseType, enabled bool, opts ...OptFn) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(ctx, filter, FieldKeyEnabled, enabled)
}

// SetAsDefault sets a release as default.
func (h *Handler) SetAsDefault(ctx context.Context, releaseType types.ReleaseType, asDefault bool, opts ...OptFn) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(ctx, filter, FieldKeyAsDefault, asDefault)
}

// CancelPlatformDefault cancel a release's all version asDefault by one platform.
func (h *Handler) CancelPlatformDefault(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(ctx, filter, FieldKeyAsDefault, false)
}

// Count counts releases.
func (h *Handler) Count(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) (int64, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.releaseTypeDao(releaseType).Count(ctx, filter)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// UpsertMany upsert many release.
func (h *Handler) UpsertMany(ctx context.Context, releaseType types.ReleaseType, releases ...*types.Release) error {
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

	return h.releaseTypeDao(releaseType).upsertMany(ctx, data)
}

// Delete deletes a release.
func (h *Handler) Delete(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DeleteMany(ctx, filter)
}

// DistinctOsType distincts os types.
func (h *Handler) DistinctOsType(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DistinctString(ctx, FieldKeyOSType, filter, nil)
}

// DistinctCPUArch distincts cpu archs.
func (h *Handler) DistinctCPUArch(ctx context.Context, releaseType types.ReleaseType, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DistinctString(ctx, FieldKeyCPUArch, filter, nil)
}

func convertReleaseToTypes(release *Release) *types.Release {
	return &types.Release{
		Name:       release.Name,
		Generation: types.Generation(release.Generation),
		Type:       types.ReleaseType(release.Type),
		Version:    release.Version,
		Platform: platform.Platform{
			Arch: criteria.CPUArch(release.CPUArch),
			OS:   criteria.OSType(release.OSType),
		},
		Labels:       release.Labels,
		FileName:     release.FileName,
		MD5:          release.MD5,
		Enabled:      release.Enabled,
		AsDefault:    release.AsDefault,
		UpdatedAt:    release.UpdatedAt,
		Operator:     release.Operator,
		AdditionInfo: release.AdditionInfo,
	}
}

func convertReleaseFromTypes(release *types.Release) *Release {
	return &Release{
		Name:         release.Name,
		Generation:   int64(release.Generation),
		Type:         string(release.Type),
		Version:      release.Version,
		CPUArch:      string(release.Platform.Arch),
		OSType:       string(release.Platform.OS),
		Labels:       release.Labels,
		Enabled:      release.Enabled,
		AsDefault:    release.AsDefault,
		FileName:     release.FileName,
		MD5:          release.MD5,
		UpdatedAt:    release.UpdatedAt,
		Operator:     release.Operator,
		AdditionInfo: release.AdditionInfo,
	}
}
