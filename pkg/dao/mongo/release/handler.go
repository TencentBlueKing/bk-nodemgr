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
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler release Handler interface.
type IHandler interface {
	// Get gets a release.
	Get(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) (*types.Release, error)

	// Exist checks if a release exists.
	Exist(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) (bool, error)

	// List lists releases.
	List(nCtx contextx.IContext, releaseType types.ReleaseType, page types.Page, opts ...OptFn) ([]*types.Release, int64, error)

	// SetLabels sets a release's labels.
	SetLabels(nCtx contextx.IContext, releaseType types.ReleaseType, labels []string, opts ...OptFn) error

	// Count counts releases.
	Count(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) (int64, error)

	// UpsertMany upserts a release.
	UpsertMany(nCtx contextx.IContext, releaseType types.ReleaseType, releases ...*types.Release) error

	// Delete deletes a release.
	Delete(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) error

	ISwitcher
	IDistinctor
}

// ISwitcher defines the default interface.
type ISwitcher interface {
	// SetEnabled sets a release's enabled.
	SetEnabled(nCtx contextx.IContext, releaseType types.ReleaseType, enabled bool, opts ...OptFn) error

	// SetAsDefault sets a release's asDefault.
	SetAsDefault(nCtx contextx.IContext, releaseType types.ReleaseType, asDefault bool, opts ...OptFn) error

	// SetVisible sets whether a release is visible to the frontend.
	SetVisible(nCtx contextx.IContext, releaseType types.ReleaseType, isVisible bool, opts ...OptFn) error

	// CancelPlatformDefault cancel a release's all version asDefault by one platform.
	CancelPlatformDefault(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) error
}

// IDistinctor defines the distinctor interface.
type IDistinctor interface {
	// DistinctOsType distincts os types.
	DistinctOsType(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error)

	// DistinctCPUArch distincts cpu archs.
	DistinctCPUArch(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error)

	// DistinctName distincts release names.
	DistinctName(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error)

	// DistinctVersion distincts release versions.
	DistinctVersion(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error)
}

var _ IHandler = &Handler{}

// Handler implements IHandler.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) releaseTypeDao(category types.ReleaseType) *dao {
	tableName := TableName(string(category))
	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tableName, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure release indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the releaseTypeDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new trigger Handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Get gets a release.
func (h *Handler) Get(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) (*types.Release, error) {
	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := h.releaseTypeDao(releaseType).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertReleaseToTypes(data), nil
}

// Exist checks if a release exists.
func (h *Handler) Exist(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) (bool, error) {
	if nCtx == nil {
		return false, errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).Exist(nCtx, filter)
}

// List lists releases.
func (h *Handler) List(nCtx contextx.IContext, releaseType types.ReleaseType, page types.Page, opts ...OptFn) ([]*types.Release, int64, error) {
	if nCtx == nil {
		return nil, 0, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.releaseTypeDao(releaseType).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	releases, err := h.releaseTypeDao(releaseType).List(nCtx, filter, findOpt)
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
func (h *Handler) SetLabels(nCtx contextx.IContext, releaseType types.ReleaseType, labels []string, opts ...OptFn) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(nCtx, filter, FieldKeyLabels, labels)
}

// SetEnabled sets a release's enabled.
func (h *Handler) SetEnabled(nCtx contextx.IContext, releaseType types.ReleaseType, enabled bool, opts ...OptFn) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(nCtx, filter, FieldKeyEnabled, enabled)
}

// SetAsDefault sets a release as default.
func (h *Handler) SetAsDefault(nCtx contextx.IContext, releaseType types.ReleaseType, asDefault bool, opts ...OptFn) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(nCtx, filter, FieldKeyAsDefault, asDefault)
}

// SetVisible sets whether a release is visible to the frontend.
func (h *Handler) SetVisible(nCtx contextx.IContext, releaseType types.ReleaseType, isVisible bool, opts ...OptFn) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(nCtx, filter, FieldKeyIsVisible, isVisible)
}

// CancelPlatformDefault cancel a release's all version asDefault by one platform.
func (h *Handler) CancelPlatformDefault(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).UpdateField(nCtx, filter, FieldKeyAsDefault, false)
}

// Count counts releases.
func (h *Handler) Count(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.releaseTypeDao(releaseType).Count(nCtx, filter)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// UpsertMany upsert many release.
func (h *Handler) UpsertMany(nCtx contextx.IContext, releaseType types.ReleaseType, releases ...*types.Release) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if len(releases) == 0 {
		return errors.New("release is nil")
	}

	data := make([]*Release, len(releases))
	for idx, release := range releases {
		data[idx] = convertReleaseFromTypes(release)
	}

	return h.releaseTypeDao(releaseType).upsertMany(nCtx, data)
}

// Delete deletes a release.
func (h *Handler) Delete(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	opts = append(opts, WithType(releaseType))
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DeleteMany(nCtx, filter)
}

// DistinctOsType distincts os types.
func (h *Handler) DistinctOsType(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DistinctString(nCtx, FieldKeyOSType, filter, nil)
}

// DistinctCPUArch distincts cpu archs.
func (h *Handler) DistinctCPUArch(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DistinctString(nCtx, FieldKeyCPUArch, filter, nil)
}

// DistinctName distincts release names.
func (h *Handler) DistinctName(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error) {
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DistinctString(nCtx, FieldKeyName, filter, nil)
}

// DistinctVersion distincts release versions.
func (h *Handler) DistinctVersion(nCtx contextx.IContext, releaseType types.ReleaseType, opts ...OptFn) ([]string, error) {
	if nCtx == nil {
		return nil, errors.New("nCtx is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.releaseTypeDao(releaseType).DistinctString(nCtx, FieldKeyVersion, filter, nil)
}

func convertReleaseToTypes(release *Release) *types.Release {
	return &types.Release{
		Name:       release.Name,
		Generation: types.Generation(release.Generation),
		Type:       types.ReleaseType(release.Type),
		Version:    release.Version,
		Platform: platfmt.Platform{
			Arch: criteria.CPUArch(release.CPUArch),
			OS:   criteria.OSType(release.OSType),
		},
		Labels:       release.Labels,
		FileName:     release.FileName,
		MD5:          release.MD5,
		Enabled:      release.Enabled,
		IsVisible:    release.IsVisible,
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
		IsVisible:    release.IsVisible,
		AsDefault:    release.AsDefault,
		FileName:     release.FileName,
		MD5:          release.MD5,
		UpdatedAt:    release.UpdatedAt,
		Operator:     release.Operator,
		AdditionInfo: release.AdditionInfo,
	}
}
