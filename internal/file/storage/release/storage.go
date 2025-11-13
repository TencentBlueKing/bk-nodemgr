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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
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

	daoRelease release.IHandler
}

func (s *Storage) initDao() error {
	s.daoRelease = release.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoRelease == nil {
		return errors.New("dao release is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// UpsertManyReleaseAgent upsert many release.
func (s *Storage) UpsertManyReleaseAgent(ctx contextx.IContext, releaseAgents []*types.ReleaseAgent) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("upsert_many_release_agent")
	defer metric.End(err)

	err = s.upsertManyReleaseAgent(ctx, releaseAgents)

	return err
}

// ExistReleaseAgent checks if release agent exists.
func (s *Storage) ExistReleaseAgent(ctx contextx.IContext, gen types.Generation, version string, plats ...platform.Platform) (bool, error) {
	var (
		result bool
		err    error
	)

	// record metric.
	metric := s.metric().Start("exist_release_agent")
	defer metric.End(err)

	result, err = s.existReleaseAgent(ctx, gen, version, plats...)

	return result, err
}

// GetReleaseAgent gets release by generation, type, platform and version.
func (s *Storage) GetReleaseAgent(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseAgent, error) {
	var (
		data *types.ReleaseAgent
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release_agent")
	defer metric.End(err)

	data, err = s.getReleaseAgent(ctx, gen, plat, version)

	return data, err
}

// UpsertReleaseBinTool upserts release bintool.
func (s *Storage) UpsertReleaseBinTool(ctx contextx.IContext, bintool types.ReleaseBinTool) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("upsert_release_bintool")
	defer metric.End(err)

	err = s.upsertReleaseBinTool(ctx, bintool)

	return err
}

// DeleteReleaseBinTool deletes release bintool.
func (s *Storage) DeleteReleaseBinTool(ctx contextx.IContext, gen types.Generation, fileName string) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("delete_release_bintool")
	defer metric.End(err)

	err = s.deleteReleaseBinTool(ctx, gen, fileName)

	return err
}

// GetReleaseBinTool gets release bintool.
func (s *Storage) GetReleaseBinTool(ctx contextx.IContext, gen types.Generation) (*types.ReleaseBinTool, error) {
	var (
		data *types.ReleaseBinTool
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release_bintool")
	defer metric.End(err)

	data, err = s.getReleaseBinTool(ctx, gen)

	return data, err
}

// ExistReleaseBinTool checks if release bintool exists.
func (s *Storage) ExistReleaseBinTool(ctx contextx.IContext, gen types.Generation) (bool, error) {
	var (
		result bool
		err    error
	)

	// record metric.
	metric := s.metric().Start("exist_release_bin_tool")
	defer metric.End(err)

	result, err = s.existReleaseBinTool(ctx, gen)

	return result, err
}

// GetReleaseCert gets release cert.
func (s *Storage) GetReleaseCert(ctx contextx.IContext) (*types.ReleaseCert, error) {
	var (
		data *types.ReleaseCert
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release_cert")
	defer metric.End(err)

	data, err = s.getReleaseCert(ctx)

	return data, err
}

// ExistReleaseCert checks if release cert exists.
func (s *Storage) ExistReleaseCert(ctx contextx.IContext) (bool, error) {
	var (
		result bool
		err    error
	)

	// record metric.
	metric := s.metric().Start("exist_release_cert")
	defer metric.End(err)

	result, err = s.existReleaseCert(ctx)

	return result, err
}

// UpsertReleaseCert upserts release cert.
func (s *Storage) UpsertReleaseCert(ctx contextx.IContext, cert types.ReleaseCert) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("upsert_release_cert")
	defer metric.End(err)

	err = s.upsertReleaseCert(ctx, cert)

	return err
}

// DeleteReleaseCert deletes release cert.
func (s *Storage) DeleteReleaseCert(ctx contextx.IContext, fileName string) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("delete_release_cert")
	defer metric.End(err)

	err = s.deleteReleaseCert(ctx, fileName)

	return err
}

// GetReleaseProxy gets release by generation, type, platform and version.
func (s *Storage) GetReleaseProxy(ctx contextx.IContext, gen types.Generation, plat platform.Platform, version string) (*types.ReleaseProxy, error) {
	var (
		data *types.ReleaseProxy
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release_proxy")
	defer metric.End(err)

	data, err = s.getReleaseProxy(ctx, gen, plat, version)

	return data, err
}

// UpsertManyReleaseProxy upsert many release.
func (s *Storage) UpsertManyReleaseProxy(ctx contextx.IContext, releaseProxys []*types.ReleaseProxy) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("upsert_many_release_proxy")
	defer metric.End(err)

	err = s.upsertManyReleaseProxy(ctx, releaseProxys)

	return err
}

// UpsertReleasePluginBinTool upserts release plugin bintool.
func (s *Storage) UpsertReleasePluginBinTool(ctx contextx.IContext, pluginBinTool types.ReleasePluginBinTool) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("upsert_release_plugin_bintool")
	defer metric.End(err)

	err = s.upsertReleasePluginBinTool(ctx, pluginBinTool)

	return err
}

// DeleteReleasePluginBinTool deletes release plugin bintool.
func (s *Storage) DeleteReleasePluginBinTool(ctx contextx.IContext, gen types.Generation, fileName string) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("delete_release_plugin_bintool")
	defer metric.End(err)

	err = s.deleteReleasePluginBinTool(ctx, gen, fileName)

	return err
}

// GetReleasePluginBinTool gets release plugin bintool.
func (s *Storage) GetReleasePluginBinTool(ctx contextx.IContext, gen types.Generation, name string) (*types.ReleasePluginBinTool, error) {
	var (
		data *types.ReleasePluginBinTool
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release_plugin_bintool")
	defer metric.End(err)

	data, err = s.getReleasePluginBinTool(ctx, gen, name)

	return data, err
}

// ExistReleasePluginBinTool checks if release plugin bintool exists.
func (s *Storage) ExistReleasePluginBinTool(ctx contextx.IContext, gen types.Generation) (bool, error) {
	var (
		result bool
		err    error
	)

	// record metric.
	metric := s.metric().Start("exist_release_plugin_bintool")
	defer metric.End(err)

	result, err = s.existReleasePluginBinTool(ctx, gen)

	return result, err
}

// ExistReleasePlugin checks if release plugin exists.
func (s *Storage) ExistReleasePlugin(
	ctx contextx.IContext, pluginName string, version string, plats ...platform.Platform) (bool, error) {

	var (
		result bool
		err    error
	)

	// record metric.
	metric := s.metric().Start("exist_release_plugin")
	defer metric.End(err)

	result, err = s.existReleasePlugin(ctx, pluginName, version, plats...)

	return result, err
}

// UpsertManyReleasePlugin upsert many release.
func (s *Storage) UpsertManyReleasePlugin(ctx contextx.IContext, releasePlugins []*types.ReleasePlugin) error {
	var err error

	// record metric.
	metric := s.metric().Start("upsert_many_release_plugin")
	defer metric.End(err)

	err = s.upsertManyReleasePlugin(ctx, releasePlugins)

	return err
}

// GetReleasePlugin gets release plugin.
func (s *Storage) GetReleasePlugin(ctx contextx.IContext, pluginName string, gen types.Generation, plat platform.Platform, version string) (
	*types.ReleasePlugin, error) {

	var (
		data *types.ReleasePlugin
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_release_plugin")
	defer metric.End(err)

	data, err = s.getReleasePlugin(ctx, pluginName, gen, plat, version)

	return data, err
}
