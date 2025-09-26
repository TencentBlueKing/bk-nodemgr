/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package globalsettings provides storage for global settings.
package globalsettings

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "global_settings"
)

// NewStorage creates a new global settings storage handler.
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

// Storage provides a global settings storage handler.
type Storage struct {
	basestorage.Storage

	daoGlobalSettings globalsettings.IHandler
}

func (s *Storage) initDao() error {
	s.daoGlobalSettings = globalsettings.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoGlobalSettings == nil {
		return errors.New("dao global settings is nil")
	}

	return nil
}

// ListGlobalSettings lists global settings by page and condition.
func (s *Storage) ListGlobalSettings(nCtx contextx.IContext, page types.Page, condition *types.GlobalSettingsCondition) (
	[]*types.GlobalSettings, int64, error) {

	var results []*types.GlobalSettings
	var num int64
	var err error

	// record metric.
	metric := s.metric().Start("list")
	defer metric.End(err)

	opts := convertGlobalSettingsConditionsToOptions(condition)
	if results, num, err = s.daoGlobalSettings.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// CountGlobalSettings counts global settings by condition.
func (s *Storage) CountGlobalSettings(nCtx contextx.IContext, condition *types.GlobalSettingsCondition) (int64, error) {
	var num int64
	var err error

	// record metric.
	metric := s.metric().Start("count")
	defer metric.End(err)

	opts := convertGlobalSettingsConditionsToOptions(condition)
	if num, err = s.daoGlobalSettings.Count(nCtx, opts...); err != nil {
		return 0, nil
	}

	return num, err
}

// ExistGlobalSettings checks if global settings exist by condition.
func (s *Storage) ExistGlobalSettings(nCtx contextx.IContext, key string) (bool, error) {
	var exist bool
	var err error

	// record metric.
	metric := s.metric().Start("exist")
	defer metric.End(err)

	if exist, err = s.daoGlobalSettings.Exist(nCtx, key); err != nil {
		return false, err
	}

	return exist, err
}

// GetGlobalSetting gets a global settings by setting name.
func (s *Storage) GetGlobalSetting(nCtx contextx.IContext, name string) (string, error) {
	var setting *types.GlobalSettings
	var err error

	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	if setting, err = s.daoGlobalSettings.Get(nCtx, name); err != nil {
		return "", err
	}

	return setting.Value, nil
}

// UpsertGlobalSettings upserts many global settings.
func (s *Storage) UpsertGlobalSettings(nCtx contextx.IContext, settings ...*types.GlobalSettings) error {
	var err error

	// record metric.
	metric := s.metric().Start("upsert")
	defer metric.End(err)

	if err = s.daoGlobalSettings.Upsert(nCtx, settings...); err != nil {
		return err
	}

	return nil
}

// DeleteGlobalSettings deletes many global settings.
func (s *Storage) DeleteGlobalSettings(nCtx contextx.IContext, settingName ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete")
	defer metric.End(err)

	if err = s.daoGlobalSettings.Delete(nCtx, settingName...); err != nil {
		return err
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// convertGlobalSettingsConditionsToOptions converts global settings conditions to options.
func convertGlobalSettingsConditionsToOptions(condition *types.GlobalSettingsCondition) []globalsettings.OptFn {
	opts := make([]globalsettings.OptFn, 0)

	if condition == nil {
		return opts
	}

	if condition.ExactInclude != nil && len(condition.ExactInclude.SettingName) > 0 {
		opts = append(opts, globalsettings.WithSettingName(condition.ExactInclude.SettingName...))
	}

	if condition.ExactExclude != nil && len(condition.ExactExclude.SettingName) > 0 {
		opts = append(opts, globalsettings.WithoutSettingName(condition.ExactExclude.SettingName...))
	}

	if condition.FuzzyInclude != nil && len(condition.FuzzyInclude.SettingName) > 0 {
		opts = append(opts, globalsettings.WithFuzzySettingName(condition.FuzzyInclude.SettingName...))
	}

	if condition.FuzzyExclude != nil && len(condition.FuzzyExclude.SettingName) > 0 {
		opts = append(opts, globalsettings.WithoutFuzzySettingName(condition.FuzzyExclude.SettingName...))
	}

	return opts
}
