/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package v3

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate checks pagination values.
func (x *ListGlobalSettingsReq) Validate() error {
	return validatePage(x.GetPage())
}

// AutoConvert intentionally leaves the request unchanged.
func (x *ListGlobalSettingsReq) AutoConvert() {}

// ConvertPageToTypes converts pagination within the Backend query limit.
func (x *ListGlobalSettingsReq) ConvertPageToTypes() (types.Page, error) {
	page, err := convPageToTypes(x.GetPage())
	if err != nil {
		return page, fmt.Errorf("invalid global settings page: %w", err)
	}
	// The Backend global settings API accepts at most 500 items per page.
	const maxGlobalSettingsLimit = 500
	if page.Limit <= 0 || page.Limit > maxGlobalSettingsLimit {
		return page, fmt.Errorf("page.limit must be in (0, %d]", maxGlobalSettingsLimit)
	}

	return page, nil
}

// ConvertConditionsToTypes converts variable-name filters to the domain contract.
func (x *ListGlobalSettingsReq) ConvertConditionsToTypes() *types.GlobalSettingsCondition {
	condition := new(types.GlobalSettingsCondition)
	if exact := x.GetExactIncludeConditions(); exact != nil {
		condition.ExactInclude = &types.GlobalSettingsExactFields{SettingName: exact.GetSettingName()}
	}
	if fuzzy := x.GetFuzzyIncludeConditions(); fuzzy != nil {
		condition.FuzzyInclude = &types.GlobalSettingsFuzzyFields{SettingName: fuzzy.GetSettingName()}
	}
	if exact := x.GetExactExcludeConditions(); exact != nil {
		condition.ExactExclude = &types.GlobalSettingsExactFields{SettingName: exact.GetSettingName()}
	}
	if fuzzy := x.GetFuzzyExcludeConditions(); fuzzy != nil {
		condition.FuzzyExclude = &types.GlobalSettingsFuzzyFields{SettingName: fuzzy.GetSettingName()}
	}

	return condition
}

// ConvertGlobalSettingsFromTypes converts stored values without interpreting them.
func (x *ListGlobalSettingsResp) ConvertGlobalSettingsFromTypes(total int64, settings []*types.GlobalSettings) {
	x.Data = &ListGlobalSettingsResp_Data{
		Total: total,
		Items: conv.SliceToSlice(settings, func(setting *types.GlobalSettings) *GlobalSetting {
			return &GlobalSetting{SettingName: setting.SettingName, Value: setting.Value}
		}),
	}
}
