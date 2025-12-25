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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "release"

	metricOperateionGetReleasePluginDefaultVersion = "get_release_plugin_default_version"
)

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

// ==================== IAgent Methods ====================

// ListReleaseAgent lists agent releases by page and conditions.
func (s *Storage) ListReleaseAgent(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseAgent, int64, error) {

	var (
		results []*types.ReleaseAgent
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_release_agent")
	defer metric.End(err)

	if results, num, err = s.listReleaseAgent(nCtx, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list release agent")
		return nil, 0, fmt.Errorf("failed to list release agent: %w", err)
	}

	return results, num, nil
}

// CountReleaseAgent counts agent releases by conditions.
func (s *Storage) CountReleaseAgent(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_release_agent")
	defer metric.End(err)

	if num, err = s.countRelease(nCtx, types.ReleaseTypeAgent, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count release agent")
		return 0, fmt.Errorf("failed to count release agent: %w", err)
	}

	return num, err
}

// GetReleaseAgent gets agent release.
func (s *Storage) GetReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) (*types.ReleaseAgent, error) {
	var (
		releaseAgent *types.ReleaseAgent
		err          error
	)

	// record metric.
	metric := s.metric().Start("get_release_agent")
	defer metric.End(err)

	releaseAgent, err = s.getReleaseAgent(nCtx, key.Generation, key.Platform, key.Version)
	if err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release agent")
		return nil, fmt.Errorf("failed to get release agent: %w", err)
	}

	return releaseAgent, nil
}

// DistinctReleaseAgent gets agent releases distinct.
func (s *Storage) DistinctReleaseAgent(nCtx contextx.IContext, fields types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {
	var (
		data *types.ReleaseDistinctResult
		err  error
	)

	// record metric.
	metric := s.metric().Start("distinct_release_agent")
	defer metric.End(err)

	if data, err = s.distinctRelease(nCtx, types.ReleaseTypeAgent, fields, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to distinct release agent")
		return nil, fmt.Errorf("failed to distinct release agent: %w", err)
	}

	return data, nil
}

// DeleteReleaseAgent deletes agent release.
func (s *Storage) DeleteReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_release_agent")
	defer metric.End(err)

	if err = s.deleteRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release agent")
		return fmt.Errorf("failed to delete release agent: %w", err)
	}

	return nil
}

// SetReleaseAgentLabelsMany sets many agent release labels.
func (s *Storage) SetReleaseAgentLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	var err error

	// record metric.
	metric := s.metric().Start("set_release_agent_labels_many")
	defer metric.End(err)

	if err = s.setReleaseLabelsMany(nCtx, types.ReleaseTypeAgent, labels, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to set many release agent labels")
		return fmt.Errorf("failed to set many release agent labels: %w", err)
	}

	return nil
}

// EnableReleaseAgent enables agent release active.
func (s *Storage) EnableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("enable_release_agent")
	defer metric.End(err)

	if err = s.enableRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to enable release agent")
		return fmt.Errorf("failed to enable release agent: %w", err)
	}

	return nil
}

// DisableReleaseAgent disables agent release disactive.
func (s *Storage) DisableReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("disable_release_agent")
	defer metric.End(err)

	if err = s.disableRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to disable release agent")
		return fmt.Errorf("failed to disable release agent: %w", err)
	}

	return nil
}

// SetAsDefaultReleaseAgent sets the agent release as default.
func (s *Storage) SetAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("set_as_default_release_agent")
	defer metric.End(err)

	if err = s.setAsDefaultRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to set release agent as default")
		return fmt.Errorf("failed to set release agent as default: %w", err)
	}

	return nil
}

// CancelAsDefaultReleaseAgent cancels the agent release as default.
func (s *Storage) CancelAsDefaultReleaseAgent(nCtx contextx.IContext, key types.ReleaseAgentKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("cancel_as_default_release_agent")
	defer metric.End(err)

	if err = s.cancelAsDefaultRelease(nCtx, types.ReleaseTypeAgent, key.Generation, key.Platform, key.Version, types.ReleaseNameAgent); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to cancel release agent as default")
		return fmt.Errorf("failed to cancel release agent as default: %w", err)
	}

	return nil
}

// ==================== IProxy Methods ====================

// ListReleaseProxy lists proxy releases by page and conditions.
func (s *Storage) ListReleaseProxy(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseProxy, int64, error) {

	var (
		results []*types.ReleaseProxy
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_release_proxy")
	defer metric.End(err)

	if results, num, err = s.listReleaseProxy(nCtx, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list release proxy")
		return nil, 0, fmt.Errorf("failed to list release proxy: %w", err)
	}

	return results, num, nil
}

// CountReleaseProxy counts proxy releases by conditions.
func (s *Storage) CountReleaseProxy(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_release_proxy")
	defer metric.End(err)

	if num, err = s.countRelease(nCtx, types.ReleaseTypeProxy, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count release proxy")
		return 0, fmt.Errorf("failed to count release proxy: %w", err)
	}

	return num, err
}

// GetReleaseProxy gets proxy release.
func (s *Storage) GetReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) (*types.ReleaseProxy, error) {
	var (
		releaseProxy *types.ReleaseProxy
		err          error
	)

	// record metric.
	metric := s.metric().Start("get_release_proxy")
	defer metric.End(err)

	releaseProxy, err = s.getReleaseProxy(nCtx, key.Generation, key.Platform, key.Version)
	if err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release proxy")
		return nil, fmt.Errorf("failed to get release proxy: %w", err)
	}

	return releaseProxy, nil
}

// DistinctReleaseProxy gets proxy releases distinct.
func (s *Storage) DistinctReleaseProxy(nCtx contextx.IContext, fields types.ReleaseDistinctField,
	conditions ...*types.ReleaseCondition) (*types.ReleaseDistinctResult, error) {
	var (
		data *types.ReleaseDistinctResult
		err  error
	)

	// record metric.
	metric := s.metric().Start("distinct_release_proxy")
	defer metric.End(err)

	if data, err = s.distinctRelease(nCtx, types.ReleaseTypeProxy, fields, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to distinct release proxy")
		return nil, fmt.Errorf("failed to distinct release proxy: %w", err)
	}

	return data, nil
}

// DeleteReleaseProxy deletes proxy release.
func (s *Storage) DeleteReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_release_proxy")
	defer metric.End(err)

	if err = s.deleteRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release proxy")
		return fmt.Errorf("failed to delete release proxy: %w", err)
	}

	return nil
}

// SetReleaseProxyLabelsMany sets many proxy releases labels.
func (s *Storage) SetReleaseProxyLabelsMany(nCtx contextx.IContext, labels []string, conditions ...*types.ReleaseCondition) error {
	var err error

	// record metric.
	metric := s.metric().Start("set_release_proxy_labels_many")
	defer metric.End(err)

	if err = s.setReleaseLabelsMany(nCtx, types.ReleaseTypeProxy, labels, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to set many release proxy labels")
		return fmt.Errorf("failed to set many release proxy labels: %w", err)
	}

	return nil
}

// EnableReleaseProxy enables proxy release active.
func (s *Storage) EnableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("enable_release_proxy")
	defer metric.End(err)

	if err = s.enableRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to enable release proxy")
		return fmt.Errorf("failed to enable release proxy: %w", err)
	}

	return nil
}

// DisableReleaseProxy disables proxy release disactive.
func (s *Storage) DisableReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("disable_release_proxy")
	defer metric.End(err)

	if err = s.disableRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to disable release proxy")
		return fmt.Errorf("failed to disable release proxy: %w", err)
	}

	return nil
}

// SetAsDefaultReleaseProxy sets the proxy release as default.
func (s *Storage) SetAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("set_as_default_release_proxy")
	defer metric.End(err)

	if err = s.setAsDefaultRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to set release proxy as default")
		return fmt.Errorf("failed to set release proxy as default: %w", err)
	}

	return nil
}

// CancelAsDefaultReleaseProxy cancels the proxy release as default.
func (s *Storage) CancelAsDefaultReleaseProxy(nCtx contextx.IContext, key types.ReleaseProxyKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("cancel_as_default_release_proxy")
	defer metric.End(err)

	if err = s.cancelAsDefaultRelease(nCtx, types.ReleaseTypeProxy, key.Generation, key.Platform, key.Version, types.ReleaseNameProxy); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to cancel release proxy as default")
		return fmt.Errorf("failed to cancel release proxy as default: %w", err)
	}

	return nil
}

// ==================== IPlugin Methods ====================

// ListReleasePlugin lists plugin releases by page and conditions.
func (s *Storage) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) (
	results []*types.ReleasePlugin, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_release_plugin")
	defer metric.End(err)

	results, num, err = s.listReleasePlugin(nCtx, page, conditions...)

	return results, num, err
}

// CountReleasePlugin counts plugin release by conditions.
func (s *Storage) CountReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (num int64, err error) {
	// record metric.
	metric := s.metric().Start("count_release_plugin")
	defer metric.End(err)

	num, err = s.countRelease(nCtx, types.ReleaseTypePlugin, conditions...)

	return num, err
}

// GetReleasePlugin gets plugin release.
func (s *Storage) GetReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (
	releasePlugin *types.ReleasePlugin, err error) {

	// record metric.
	metric := s.metric().Start("get_release_release_plugin")
	defer metric.End(err)

	releasePlugin, err = s.getReleasePlugin(nCtx, key.Name, key.Generation, key.Platform, key.Version)

	return releasePlugin, err
}

// DeleteReleasePlugin deletes plugin release.
func (s *Storage) DeleteReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (err error) {

	// record metric.
	metric := s.metric().Start("delete_release_plugin")
	defer metric.End(err)

	err = s.deleteRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

	return err
}

// EnableReleasePlugin enables plugin release active.
func (s *Storage) EnableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (err error) {

	// record metric.
	metric := s.metric().Start("enable_release_plugin")
	defer metric.End(err)

	err = s.enableRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

	return err
}

// DisableReleasePlugin disables plugin release disactive.
func (s *Storage) DisableReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (err error) {

	// record metric.
	metric := s.metric().Start("disable_release_plugin")
	defer metric.End(err)

	err = s.disableRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

	return err
}

// SetAsDefaultReleasePlugin sets the plugin release as default.
func (s *Storage) SetAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (err error) {

	// record metric.
	metric := s.metric().Start("set_as_default_release_plugin")
	defer metric.End(err)

	err = s.setAsDefaultRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

	return err
}

// CancelAsDefaultReleasePlugin cancels the plugin release as default.
func (s *Storage) CancelAsDefaultReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (err error) {

	// record metric.
	metric := s.metric().Start("cancel_as_default_release_plugin")
	defer metric.End(err)

	err = s.cancelAsDefaultRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

	return err
}

// ExistReleasePlugin exist plugin release.
func (s *Storage) ExistReleasePlugin(nCtx contextx.IContext, key types.ReleasePluginKey) (
	exist bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_release_plugin")
	defer metric.End(err)

	exist, err = s.existRelease(nCtx, types.ReleaseTypePlugin, key.Generation, key.Platform, key.Version, key.Name)

	return exist, err
}

// GetReleasePluginDefaultVersion gets release plugin default version by name, generation and platform.
func (s *Storage) GetReleasePluginDefaultVersion(nCtx contextx.IContext, name string, gen types.Generation, plat platfmt.Platform) (string, error) {
	var (
		version string
		err     error
	)

	err = s.WrapFn(nCtx, metricOperateionGetReleasePluginDefaultVersion, func(nCtx contextx.IContext) error {
		// record metric.
		metric := s.metric().Start(metricOperateionGetReleasePluginDefaultVersion)
		defer metric.End(err)

		version, err = s.getReleaseDefaultVersion(nCtx, types.ReleaseTypePlugin, gen, plat, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return version, nil
}

// ==================== ICert Methods ====================

// ListReleaseCert lists cert releases by page and conditions.
func (s *Storage) ListReleaseCert(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseCert, int64, error) {
	var (
		results []*types.ReleaseCert
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_release_cert")
	defer metric.End(err)

	rls, num, err := s.listRelease(nCtx, types.ReleaseTypeCert, page, conditions...)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list release cert")
		return nil, 0, fmt.Errorf("failed to list release cert: %w", err)
	}

	results = make([]*types.ReleaseCert, len(rls))
	for idx, r := range rls {
		results[idx] = &types.ReleaseCert{
			Release: *r,
		}
	}

	return results, num, nil
}

// CountReleaseCert counts cert release by conditions.
func (s *Storage) CountReleaseCert(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_release_cert")
	defer metric.End(err)

	if num, err = s.countRelease(nCtx, types.ReleaseTypeCert, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count release cert")
		return 0, fmt.Errorf("failed to count release cert: %w", err)
	}

	return num, nil
}

// GetReleaseCert gets cert release.
func (s *Storage) GetReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) (*types.ReleaseCert, error) {
	var (
		releaseCert *types.ReleaseCert
		err         error
	)

	// record metric.
	metric := s.metric().Start("get_release_cert")
	defer metric.End(err)

	rls, err := s.getRelease(nCtx, types.ReleaseTypeCert, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionCert, types.ReleaseNameCert)
	if err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release cert")
		return nil, fmt.Errorf("failed to get release cert: %w", err)
	}

	releaseCert = &types.ReleaseCert{
		Release: *rls,
	}

	return releaseCert, nil
}

// DeleteReleaseCert deletes cert release.
func (s *Storage) DeleteReleaseCert(nCtx contextx.IContext, key types.ReleaseCertKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_release_cert")
	defer metric.End(err)

	if err = s.deleteRelease(
		nCtx, types.ReleaseTypeCert, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionCert, types.ReleaseNameCert); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release cert")
		return fmt.Errorf("failed to delete release cert: %w", err)
	}

	return nil
}

// ==================== IBinTool Methods ====================

// ListReleaseBinTool lists bintool releases by page and conditions.
func (s *Storage) ListReleaseBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleaseBinTool, int64, error) {
	var (
		results []*types.ReleaseBinTool
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_release_bintool")
	defer metric.End(err)

	rls, num, err := s.listRelease(nCtx, types.ReleaseTypeBinTool, page, conditions...)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list release bintool")
		return nil, 0, fmt.Errorf("failed to list release bintool: %w", err)
	}

	results = make([]*types.ReleaseBinTool, len(rls))
	for idx, r := range rls {
		results[idx] = &types.ReleaseBinTool{
			Release: *r,
		}
	}

	return results, num, nil
}

// CountReleaseBinTool counts bintool release by conditions.
func (s *Storage) CountReleaseBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_release_bintool")
	defer metric.End(err)

	if num, err = s.countRelease(nCtx, types.ReleaseTypeBinTool, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count release bintool")
		return 0, fmt.Errorf("failed to count release bintool: %w", err)
	}

	return num, nil
}

// GetReleaseBinTool gets bintool release.
func (s *Storage) GetReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) (*types.ReleaseBinTool, error) {
	var (
		releaseBinTool *types.ReleaseBinTool
		err            error
	)

	// record metric.
	metric := s.metric().Start("get_release_bintool")
	defer metric.End(err)

	rls, err := s.getRelease(
		nCtx, types.ReleaseTypeBinTool, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionBinTool, types.ReleaseNameBinTool)
	if err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release bintool")
		return nil, fmt.Errorf("failed to get release bintool: %w", err)
	}

	releaseBinTool = &types.ReleaseBinTool{
		Release: *rls,
	}

	return releaseBinTool, nil
}

// DeleteReleaseBinTool deletes bintool release.
func (s *Storage) DeleteReleaseBinTool(nCtx contextx.IContext, key types.ReleaseBinToolKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_release_bintool")
	defer metric.End(err)

	if err = s.deleteRelease(
		nCtx,
		types.ReleaseTypeBinTool, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionBinTool, types.ReleaseNameBinTool); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release bintool")
		return fmt.Errorf("failed to delete release bintool: %w", err)
	}

	return nil
}

// ==================== IPluginBinTool Methods ====================

// ListReleasePluginBinTool lists plugin bintool release by page and conditions.
func (s *Storage) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error) {
	var (
		results []*types.ReleasePluginBinTool
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_release_plugin_bintool")
	defer metric.End(err)

	rls, num, err := s.listRelease(nCtx, types.ReleaseTypePluginBinTool, page, conditions...)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list release plugin bintool")
		return nil, 0, fmt.Errorf("failed to list release plugin bintool: %w", err)
	}

	results = make([]*types.ReleasePluginBinTool, len(rls))
	for idx, r := range rls {
		results[idx] = &types.ReleasePluginBinTool{
			Release: *r,
		}
	}

	return results, num, nil
}

// CountReleasePluginBinTool counts plugin bintool release by conditions.
func (s *Storage) CountReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) (int64, error) {
	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_release_plugin_bintool")
	defer metric.End(err)

	if num, err = s.countRelease(nCtx, types.ReleaseTypePluginBinTool, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count release plugin bintool")
		return 0, fmt.Errorf("failed to count release plugin bintool: %w", err)
	}

	return num, nil
}

// GetReleasePluginBinTool gets plugin bintool release.
func (s *Storage) GetReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) (*types.ReleasePluginBinTool, error) {
	var (
		releasePluginBinTool *types.ReleasePluginBinTool
		err                  error
	)

	// record metric.
	metric := s.metric().Start("get_release_plugin_bintool")
	defer metric.End(err)

	rls, err := s.getRelease(
		nCtx, types.ReleaseTypePluginBinTool, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionPluginBinTool, key.Name)
	if err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to get release plugin bintool")
		return nil, fmt.Errorf("failed to get release plugin bintool: %w", err)
	}

	releasePluginBinTool = &types.ReleasePluginBinTool{
		Release: *rls,
	}

	return releasePluginBinTool, nil
}

// DeleteReleasePluginBinTool deletes plugin bintool release.
func (s *Storage) DeleteReleasePluginBinTool(nCtx contextx.IContext, key types.ReleasePluginBinToolKey) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_release_plugin_bintool")
	defer metric.End(err)

	if err = s.deleteRelease(
		nCtx, types.ReleaseTypePluginBinTool, key.Generation, platform.UnknownPlatform(), types.ReleaseVersionPluginBinTool, key.Name); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Error("failed to delete release plugin bintool")
		return fmt.Errorf("failed to delete release plugin bintool: %w", err)
	}

	return nil
}

// ==================== IPackageEvent Methods ====================

// CountPackageEvent counts package events by conditions.
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

// ListPackageEvent lists package events by page and conditions.
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

	if results, num, err = s.listPackageEvent(nCtx, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list package event")
		return nil, 0, fmt.Errorf("failed to list package event: %w", err)
	}

	return results, num, nil
}

// CreateManyPackageEvent creates multiple package events.
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

// DistinctPackageEvent distincts package event fields.
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
