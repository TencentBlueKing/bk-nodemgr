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
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/globalsettings"
	pkgglobalsettings "github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "global_settings"

	metricOperationListGlobalSettings   = "list"
	metricOperationCountGlobalSettings  = "count"
	metricOperationExistGlobalSettings  = "exist"
	metricOperationGetGlobalSettings    = "get"
	metricOperationUpsertGlobalSettings = "upsert"
	metricOperationDeleteGlobalSettings = "delete"
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

	var (
		results []*types.GlobalSettings
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListGlobalSettings, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.daoGlobalSettings.List(nCtx, page, convertGlobalSettingsConditionsToOptions(condition)...)

		return err
	})

	return results, num, err
}

// CountGlobalSettings counts global settings by condition.
func (s *Storage) CountGlobalSettings(nCtx contextx.IContext, condition *types.GlobalSettingsCondition) (int64, error) {
	var num int64

	err := s.WrapFn(nCtx, metricOperationCountGlobalSettings, func(nCtx contextx.IContext) error {
		var err error
		num, err = s.daoGlobalSettings.Count(nCtx, convertGlobalSettingsConditionsToOptions(condition)...)

		return err
	})

	return num, err
}

// ExistGlobalSettings checks if global settings exist by condition.
func (s *Storage) ExistGlobalSettings(nCtx contextx.IContext, key string) (bool, error) {
	var exist bool

	err := s.WrapFn(nCtx, metricOperationExistGlobalSettings, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.daoGlobalSettings.Exist(nCtx, key)

		return err
	})

	return exist, err
}

// GetGlobalSetting gets a global settings by setting name.
func (s *Storage) GetGlobalSetting(nCtx contextx.IContext, name string) (string, error) {
	var setting *types.GlobalSettings

	err := s.WrapFn(nCtx, metricOperationGetGlobalSettings, func(nCtx contextx.IContext) error {
		var err error
		setting, err = s.daoGlobalSettings.Get(nCtx, name)

		return err
	})
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

// UpsertGlobalSettings upserts many global settings.
func (s *Storage) UpsertGlobalSettings(nCtx contextx.IContext, settings ...*types.GlobalSettings) error {
	return s.WrapFn(nCtx, metricOperationUpsertGlobalSettings, func(nCtx contextx.IContext) error {
		if err := validateNetworkUnitSegmentRuleSettings(settings); err != nil {
			return err
		}

		var err error
		err = s.daoGlobalSettings.Upsert(nCtx, settings...)

		return err
	})
}

func validateNetworkUnitSegmentRuleSettings(settings []*types.GlobalSettings) error {
	for _, setting := range settings {
		if setting == nil || setting.SettingName != pkgglobalsettings.NetworkUnitSegmentRules {
			continue
		}

		var cfg types.NetworkUnitSegmentRuleConfig
		if err := json.Unmarshal([]byte(setting.Value), &cfg); err != nil {
			return fmt.Errorf("invalid global setting %s: %w", pkgglobalsettings.NetworkUnitSegmentRules, err)
		}
		if err := validateNetworkUnitSegmentRuleConfig(cfg); err != nil {
			return fmt.Errorf("invalid global setting %s: %w", pkgglobalsettings.NetworkUnitSegmentRules, err)
		}
	}

	return nil
}

func validateNetworkUnitSegmentRuleConfig(cfg types.NetworkUnitSegmentRuleConfig) error {
	if len(cfg) == 0 {
		return errors.New("at least one networkarea rules config is required")
	}

	for areaID, areaConfig := range cfg {
		if len(areaConfig.Rules) == 0 {
			return fmt.Errorf("networkarea %s: rules is required", areaID)
		}

		for idx, rule := range areaConfig.Rules {
			if len(rule.CIDRs) == 0 {
				return fmt.Errorf("networkarea %s rule %d: cidrs is required", areaID, idx)
			}
			if rule.NetworkUnitID < 0 {
				return fmt.Errorf("networkarea %s rule %d: bk_networkunit_id is required", areaID, idx)
			}

			isFallbackRule := false
			for _, cidr := range rule.CIDRs {
				prefix, err := netip.ParsePrefix(cidr)
				if err != nil {
					return fmt.Errorf("networkarea %s rule %d: invalid cidr %s: %w", areaID, idx, cidr, err)
				}
				if !prefix.Addr().Unmap().Is4() {
					return fmt.Errorf("networkarea %s rule %d: IPv6 cidr is not supported", areaID, idx)
				}
				if prefix.Masked() == netip.MustParsePrefix("0.0.0.0/0") {
					isFallbackRule = true
				}
			}

			if isFallbackRule && idx != len(areaConfig.Rules)-1 {
				return fmt.Errorf("networkarea %s rule %d: fallback rule must be the last rule", areaID, idx)
			}
		}
	}

	return nil
}

// DeleteGlobalSettings deletes many global settings.
func (s *Storage) DeleteGlobalSettings(nCtx contextx.IContext, settingName ...string) error {
	return s.WrapFn(nCtx, metricOperationDeleteGlobalSettings, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoGlobalSettings.Delete(nCtx, settingName...)

		return err
	})
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
