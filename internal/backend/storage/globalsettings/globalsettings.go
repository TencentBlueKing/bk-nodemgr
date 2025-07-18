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
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "global_settings"
)

// NewStorage creates a new global settings storage handler.
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

// Storage provides a global settings storage handler.
type Storage struct {
	basestorage.Storage

	daoGlobalSettings globalsettings.IHandler
}

func (s *Storage) initDao() error {
	s.daoGlobalSettings = globalsettings.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoGlobalSettings == nil {
		return errors.New("dao global settings is nil")
	}

	return nil
}

// ListGlobalSettings lists global settings by page and condition.
func (s *Storage) ListGlobalSettings(
	ctx context.Context, page types.Page, condition *types.GlobalSettingsCondition) (
	[]*types.GlobalSettings, int64, error) {

	opts := convertGlobalSettingsConditionsToOptions(condition)

	return s.daoGlobalSettings.List(ctx, page, opts...)
}

// CountGlobalSettings counts global settings by condition.
func (s *Storage) CountGlobalSettings(
	ctx context.Context, condition *types.GlobalSettingsCondition) (int64, error) {

	opts := convertGlobalSettingsConditionsToOptions(condition)

	return s.daoGlobalSettings.Count(ctx, opts...)
}

// GetGlobalSetting gets a global settings by setting name.
func (s *Storage) GetGlobalSetting(ctx context.Context, name string) (string, error) {
	setting, err := s.daoGlobalSettings.Get(ctx, name)
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

// UpsertManyGlobalSettings upserts many global settings.
func (s *Storage) UpsertManyGlobalSettings(ctx context.Context, settings ...*types.GlobalSettings) error {
	if len(settings) == 0 {
		return nil
	}

	return s.daoGlobalSettings.UpsertMany(ctx, settings...)
}

// DeleteManyGlobalSettings deletes many global settings.
func (s *Storage) DeleteManyGlobalSettings(ctx context.Context, settingName ...string) error {
	if len(settingName) == 0 {
		return errors.New("setting name cannot be empty")
	}

	return s.daoGlobalSettings.DeleteMany(ctx, settingName...)
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
