/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *ListGlobalSettingsReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert auto convert.
func (x *ListGlobalSettingsReq) AutoConvert() {
}

const (
	// global settings list max limit
	maxGlobalSettingsLimit = 500
)

// PageLimit return page limit.
func (x *ListGlobalSettingsReq) PageLimit() int {
	return maxGlobalSettingsLimit
}

// ConvertPageToTypes convert page to types.
func (x *ListGlobalSettingsReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *ListGlobalSettingsReq) ConvertConditionsToTypes() *types.GlobalSettingsCondition {
	condition := new(types.GlobalSettingsCondition)
	if x.GetExactIncludeConditions() != nil {
		condition.ExactInclude = &types.GlobalSettingsExactFields{
			SettingName: x.GetExactIncludeConditions().GetSettingName(),
		}
	}
	if x.GetFuzzyIncludeConditions() != nil {
		condition.FuzzyInclude = &types.GlobalSettingsFuzzyFields{
			SettingName: x.GetFuzzyIncludeConditions().GetSettingName(),
		}
	}
	if x.GetExactExcludeConditions() != nil {
		condition.ExactExclude = &types.GlobalSettingsExactFields{
			SettingName: x.GetExactExcludeConditions().GetSettingName(),
		}
	}
	if x.GetFuzzyExcludeConditions() != nil {
		condition.FuzzyExclude = &types.GlobalSettingsFuzzyFields{
			SettingName: x.GetFuzzyExcludeConditions().GetSettingName(),
		}
	}

	return condition
}

// ConvertConditionsFromTypes convert types to proto.
func (x *ListGlobalSettingsReq) ConvertConditionsFromTypes(condition *types.GlobalSettingsCondition) error {
	if condition == nil {
		return errors.New("condition is nil")
	}

	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &ListGlobalSettingsReq_ExactConditions{
			SettingName: condition.ExactInclude.SettingName,
		}
	}

	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &ListGlobalSettingsReq_FuzzyConditions{
			SettingName: condition.FuzzyInclude.SettingName,
		}
	}

	if condition.ExactExclude != nil {
		x.ExactExcludeConditions = &ListGlobalSettingsReq_ExactConditions{
			SettingName: condition.ExactExclude.SettingName,
		}
	}

	if condition.FuzzyExclude != nil {
		x.FuzzyExcludeConditions = &ListGlobalSettingsReq_FuzzyConditions{
			SettingName: condition.FuzzyExclude.SettingName,
		}
	}

	return nil
}

// ConvertGlobalSettingsFromTypes converts global settings from types to proto.
func (x *ListGlobalSettingsResp) ConvertGlobalSettingsFromTypes(num int64, settings []*types.GlobalSettings) {
	if x.Data == nil {
		x.Data = &ListGlobalSettingsResp_Data{}
	}

	x.Data.Total = num
	if settings == nil {
		x.Data.Items = nil
		return
	}

	x.Data.Items = make([]*GlobalSetting, 0, len(settings))
	for _, setting := range settings {
		x.Data.Items = append(x.Data.Items, &GlobalSetting{
			SettingName: setting.SettingName,
			Value:       setting.Value,
		})
	}
}

// Validate check body.
func (x *GetGlobalSettingReq) Validate() error {
	if x.GetSettingName() == "" {
		return errors.New("setting name cannot be empty")
	}
	return nil
}

// AutoConvert auto convert.
func (x *GetGlobalSettingReq) AutoConvert() {
}

// Validate check body.
func (x *UpsertGlobalSettingsReq) Validate() error {
	if len(x.GetSettings()) == 0 {
		return errors.New("settings cannot be empty")
	}
	for _, setting := range x.GetSettings() {
		if setting.GetSettingName() == "" {
			return errors.New("setting name cannot be empty")
		}
	}
	return nil
}

// AutoConvert auto convert.
func (x *UpsertGlobalSettingsReq) AutoConvert() {
}

// ConvertGlobalSettingsToTypes converts global settings to types.
func (x *UpsertGlobalSettingsReq) ConvertGlobalSettingsToTypes() []*types.GlobalSettings {
	settings := make([]*types.GlobalSettings, 0, len(x.GetSettings()))
	for _, setting := range x.GetSettings() {
		settings = append(settings, &types.GlobalSettings{
			SettingName: setting.GetSettingName(),
			Value:       setting.GetValue(),
		})
	}
	return settings
}

// Validate check body.
func (x *DeleteGlobalSettingsReq) Validate() error {
	if len(x.GetSettingName()) == 0 {
		return errors.New("setting name cannot be empty")
	}
	return nil
}

// AutoConvert auto convert.
func (x *DeleteGlobalSettingsReq) AutoConvert() {
}
