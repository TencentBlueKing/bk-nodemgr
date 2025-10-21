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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoPackageEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "release"

// NewStorage creates a new release storage.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoRelease      release.IHandler
	daoPackageEvent daoPackageEvent.IHandler
}

func (s *Storage) initDao() error {
	s.daoRelease = release.New(s.Database)
	s.daoPackageEvent = daoPackageEvent.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoRelease == nil {
		return errors.New("dao release is nil")
	}

	if s.daoPackageEvent == nil {
		return errors.New("dao package event is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// GetRelease gets release by generation, release type, platform and version.
func (s *Storage) GetRelease(nCtx contextx.IContext,
	gen types.Generation, releaseType types.ReleaseType, name string, plat platfmt.Platform, version string) (
	*types.Release, error) {

	var (
		data *types.Release
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release")
	defer metric.End(err)

	data, err = s.getRelease(nCtx, releaseType, gen, name, plat, version)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get release")
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	return data, nil
}

// ListRelease lists release by page and conditions.
func (s *Storage) ListRelease(nCtx contextx.IContext, releaseType types.ReleaseType, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.Release, int64, error) {

	var (
		results []*types.Release
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_release")
	defer metric.End(err)

	if results, num, err = s.listRelease(nCtx, releaseType, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list release")
		return nil, 0, fmt.Errorf("failed to list release: %w", err)
	}

	return results, num, nil
}

// DistinctRelease distincts release by conditions.
func (s *Storage) DistinctRelease(nCtx contextx.IContext, releaseType types.ReleaseType,
	distinctField types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) (
	*types.ReleaseDistinctResult, error) {

	var (
		data *types.ReleaseDistinctResult
		err  error
	)

	// record metric.
	metric := s.metric().Start("distinct_release")
	defer metric.End(err)

	if data, err = s.distinctRelease(nCtx, releaseType, distinctField, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to distinct release")
		return nil, fmt.Errorf("failed to distinct release: %w", err)
	}

	return data, nil
}

// CountRelease counts release by conditions.
func (s *Storage) CountRelease(
	nCtx contextx.IContext, releaseType types.ReleaseType, conditions ...*types.ReleaseCondition) (int64, error) {

	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_release")
	defer metric.End(err)

	if num, err = s.countRelease(nCtx, releaseType, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count release")
		return 0, fmt.Errorf("failed to count release: %w", err)
	}

	return num, nil
}

// SetReleaseLabels sets release labels.
func (s *Storage) SetReleaseLabels(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platfmt.Platform, version string, labels []string) error {

	var err error

	// record metric.
	metric := s.metric().Start("set_release_labels")
	defer metric.End(err)

	if err = s.setReleaseLabels(nCtx, gen, releaseType, plat, version, labels); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to set release labels")
		return fmt.Errorf("failed to set release labels: %w", err)
	}

	return nil
}

// SetReleaseLabelsMany sets many release labels.
func (s *Storage) SetReleaseLabelsMany(nCtx contextx.IContext, releaseType types.ReleaseType, gens []types.Generation,
	plats []platfmt.Platform, versions []string, labels []string) error {

	var err error

	// record metric.
	metric := s.metric().Start("set_many_release_labels")
	defer metric.End(err)

	if err = s.setReleaseLabelsMany(nCtx, releaseType, gens, plats, versions, labels); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to set many release labels")
		return fmt.Errorf("failed to set many release labels: %w", err)
	}

	return nil
}

// EnableRelease enables release active by generation, release type, platform and version.
func (s *Storage) EnableRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platfmt.Platform, version string) error {

	var err error

	// record metric.
	metric := s.metric().Start("enable_release")
	defer metric.End(err)

	if err = s.enableRelease(nCtx, gen, releaseType, plat, version); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to enable release")
		return fmt.Errorf("failed to enable release: %w", err)
	}

	return nil
}

// DisableRelease disables release active by generation, release type, platform and version.
func (s *Storage) DisableRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platfmt.Platform, version string) error {

	var err error

	// record metric.
	metric := s.metric().Start("disable_release")
	defer metric.End(err)

	if err = s.disableRelease(nCtx, gen, releaseType, plat, version); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to disable release")
		return fmt.Errorf("failed to disable release: %w", err)
	}

	return nil
}

// SetAsDefaultRelease sets the release as default.
func (s *Storage) SetAsDefaultRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platfmt.Platform, version string) error {

	var err error

	// record metric.
	metric := s.metric().Start("set_as_default_release")
	defer metric.End(err)

	if err = s.setAsDefaultRelease(nCtx, gen, releaseType, plat, version); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to set release as default")
		return fmt.Errorf("failed to set release as default: %w", err)
	}

	return nil
}

// CancelAsDefaultRelease cancels the release as default.
func (s *Storage) CancelAsDefaultRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platfmt.Platform, version string) error {

	var err error

	// record metric.
	metric := s.metric().Start("cancel_as_default_release")
	defer metric.End(err)

	if err = s.cancelAsDefaultRelease(nCtx, gen, releaseType, plat, version); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to cancel release as default")
		return fmt.Errorf("failed to cancel release as default: %w", err)
	}

	return nil
}

// DeleteRelease deletes the release.
func (s *Storage) DeleteRelease(nCtx contextx.IContext, gen types.Generation, releaseType types.ReleaseType,
	plat platfmt.Platform, version string) error {

	var err error

	// record metric.
	metric := s.metric().Start("delete_release")
	defer metric.End(err)

	if err = s.deleteRelease(nCtx, gen, releaseType, plat, version); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to delete release")
		return fmt.Errorf("failed to delete release: %w", err)
	}

	return nil
}

// GetReleaseAgent gets release by generation, release type, platform and version.
func (s *Storage) GetReleaseAgent(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseAgent, error) {

	var (
		releaseAgent *types.ReleaseAgent
		err          error
	)

	// record metric.
	metric := s.metric().Start("get_release_agent")
	defer metric.End(err)

	releaseAgent, err = s.getReleaseAgent(nCtx, gen, plat, version)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get release agent")
		return nil, fmt.Errorf("failed to get release agent: %w", err)
	}

	return releaseAgent, nil
}

// GetReleaseProxy gets release by generation, release type, platform and version.
func (s *Storage) GetReleaseProxy(
	nCtx contextx.IContext, gen types.Generation, plat platfmt.Platform, version string) (*types.ReleaseProxy, error) {

	var (
		releaseProxy *types.ReleaseProxy
		err          error
	)

	// record metric.
	metric := s.metric().Start("get_release_proxy")
	defer metric.End(err)

	releaseProxy, err = s.getReleaseProxy(nCtx, gen, plat, version)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get release proxy")
		return nil, fmt.Errorf("failed to get release proxy: %w", err)
	}

	return releaseProxy, nil
}

// GetReleasePlugin gets release by generation, release type, platform and version.
func (s *Storage) GetReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	releasePlugin *types.ReleasePlugin, err error) {

	// record metric.
	metric := s.metric().Start("get_release_plugin")
	defer metric.End(err)

	releasePlugin, err = s.getReleasePlugin(nCtx, name, gen, plat, version)

	return releasePlugin, err
}

// ListReleasePlugin list release plugin.
func (s *Storage) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	results []*types.ReleasePlugin, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_release_plugin")
	defer metric.End(err)

	results, num, err = s.listReleasePlugin(nCtx, page, conditions...)

	return results, num, err
}

// DeleteReleasePlugin delete release plugin.
func (s *Storage) DeleteReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	err error) {

	// record metric.
	metric := s.metric().Start("delete_release_plugin")
	defer metric.End(err)

	err = s.deleteReleasePlugin(nCtx, name, gen, plat, version)

	return err
}

// EnableReleasePlugin enable release plugin.
func (s *Storage) EnableReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	err error) {

	// record metric.
	metric := s.metric().Start("enable_release_plugin")
	defer metric.End(err)

	err = s.enableReleasePlugin(nCtx, name, gen, plat, version)

	return err
}

// DisableReleasePlugin disable release plugin.
func (s *Storage) DisableReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	err error) {

	// record metric.
	metric := s.metric().Start("disable_release_plugin")
	defer metric.End(err)

	err = s.disableReleasePlugin(nCtx, name, gen, plat, version)

	return err
}

// SetAsDefaultReleasePlugin set as default release plugin.
func (s *Storage) SetAsDefaultReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	err error) {

	// record metric.
	metric := s.metric().Start("set_as_default_release_plugin")
	defer metric.End(err)

	err = s.setAsDefaultReleasePlugin(nCtx, name, gen, plat, version)

	return err
}

// CancelAsDefaultReleasePlugin cancel as default release plugin.
func (s *Storage) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform,
	version string) (err error) {

	// record metric.
	metric := s.metric().Start("cancel_as_default_release_plugin")
	defer metric.End(err)

	err = s.cancelAsDefaultReleasePlugin(nCtx, name, gen, plat, version)

	return err
}

// CountReleasePlugin count release plugin.
func (s *Storage) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (num int64, err error) {
	// record metric.
	metric := s.metric().Start("count_release_plugin")
	defer metric.End(err)

	num, err = s.countReleasePlugin(nCtx, conditions...)

	return num, err
}

// ExistReleasePlugin check release plugin exist.
func (s *Storage) ExistReleasePlugin(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform, version string) (
	exist bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_release_plugin")
	defer metric.End(err)

	exist, err = s.existReleasePlugin(nCtx, name, gen, plat, version)

	return exist, err
}

// CountPackageEvent counts package events.
func (s *Storage) CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_package_event")
	defer metric.End(err)

	if num, err = s.countPakcageEvent(nCtx, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get count package event")
		return 0, fmt.Errorf("failed to count package event: %w", err)
	}

	return num, nil
}

// ListPackageEvent lists package events.
func (s *Storage) ListPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) (
	[]*types.PackageEvent, int64, error) {

	var (
		results []*types.PackageEvent
		num     int64
		err     error
	)
	// record metric.
	metric := s.metric().Start("list_package_event")
	defer metric.End(err)

	if results, num, err = s.listPakcageEvent(nCtx, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list package event")
		return nil, 0, fmt.Errorf("failed to list package event: %w", err)
	}

	return results, num, nil
}

// CreateManyPackageEvent creates package events.
func (s *Storage) CreateManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	var err error

	// record metric.
	metric := s.metric().Start("create_many_package_event")
	defer metric.End(err)

	if err = s.createManyPackageEvent(nCtx, events...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to create many package event")
		return fmt.Errorf("failed to create many package event: %w", err)
	}

	return nil
}

// DistinctPackageEvent distincts release by conditions.
func (s *Storage) DistinctPackageEvent(
	nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
	*types.PackageEventDistinctResult, error) {

	var (
		data *types.PackageEventDistinctResult
		err  error
	)

	// record metric.
	metric := s.metric().Start("distinct_package_event")
	defer metric.End(err)

	if data, err = s.distinctPackageEvent(nCtx, request, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to distinct package event")
		return nil, fmt.Errorf("failed to distinct package event: %w", err)
	}

	return data, nil
}
