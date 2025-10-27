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
func (x *PluginInstallReq) Validate() error {
	process := x.GetProcess()
	if len(process) == 0 {
		return errors.New("plugin can not be empty")
	}

	for idx := range process {
		if err := process[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *PluginInstallReq_Process) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be zero")
	}

	if x.GetPluginId() == "" {
		return errors.New("plugin_id can not be empty")
	}

	if x.GetVersion() == "" {
		return errors.New("version can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginInstallReq) AutoConvert() {
	process := x.GetProcess()
	for idx := range process {
		process[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *PluginInstallReq_Process) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginInstallReq) ConvertParamFromTypes(installParam ...*types.PluginInstallParam) {
	process := make([]*PluginInstallReq_Process, len(installParam))
	for idx, param := range installParam {
		item := &PluginInstallReq_Process{}
		item.BkHostId = &param.HostID
		item.PluginId = param.PluginID
		item.Version = param.Version

		process[idx] = item
	}

	x.Process = process
}

// ConvertParamToTypes converts param to types.
func (x *PluginInstallReq) ConvertParamToTypes() []*types.PluginInstallParam {
	process := x.GetProcess()
	installParam := make([]*types.PluginInstallParam, len(process))
	for idx, proc := range process {
		item := &types.PluginInstallParam{
			HostID:   proc.GetBkHostId(),
			PluginID: proc.GetPluginId(),
			Version:  proc.GetVersion(),
		}

		installParam[idx] = item
	}

	return installParam
}

// Validate check body.
func (x *PluginListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PluginListReq) AutoConvert() {}

// ConvertConditionsToTypes converts conditions to types.
func (x *PluginListReq) ConvertConditionsToTypes() *types.PluginCondition {
	return convertPluginConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

func convertPluginConditionsToTypes(
	exactCond *PluginListReq_ExactConditions, fuzzyCond *PluginListReq_FuzzyConditions) *types.PluginCondition {

	condition := &types.PluginCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.PluginExactFields{
			PluginID:    exactCond.GetPluginId(),
			PluginGroup: exactCond.GetPluginGroup(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.PluginFuzzyFields{
			PluginName:    fuzzyCond.GetPluginName(),
			PluginPkgName: fuzzyCond.GetPluginPkgName(),
		}
	}

	return condition
}

// ConvertConditionFromTypes converts condition from types.
func (x *PluginListReq) ConvertConditionFromTypes(condition *types.PluginCondition) {
	if condition == nil {
		return
	}

	// exact conditions.
	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &PluginListReq_ExactConditions{
			PluginId:    condition.ExactInclude.PluginID,
			PluginGroup: condition.ExactInclude.PluginGroup,
		}
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &PluginListReq_FuzzyConditions{
			PluginName:    condition.FuzzyInclude.PluginName,
			PluginPkgName: condition.FuzzyInclude.PluginPkgName,
		}
	}
}

// ConvertPageToTypes converts page to types.
func (x *PluginListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertPluginFromTypes converts plugin from types.
func (x *PluginListResp) ConvertPluginFromTypes(total int64, plugins []*types.Plugin) {
	items := make([]*Plugin, len(plugins))
	for idx, plugin := range plugins {
		item := newEmptyPlugin()
		*item.TenantId = plugin.TenantID
		*item.PluginId = plugin.PluginID
		*item.PluginName = plugin.PluginName
		*item.PluginGroup = plugin.PluginGroup
		*item.PluginPkgName = plugin.PluginPkgName

		items[idx] = item
	}

	x.Data = &PluginListResp_Data{
		Total: total,
		Items: items,
	}
}

func newEmptyPlugin() *Plugin {
	return &Plugin{
		TenantId:      new(string),
		PluginId:      new(string),
		PluginName:    new(string),
		PluginGroup:   new(string),
		PluginPkgName: new(string),
	}
}

// ConvertPluginToTypes converts plugin to types.
func (x *PluginListResp) ConvertPluginToTypes() ([]*types.Plugin, int64) {
	data := x.GetData()
	total := data.GetTotal()
	plugins := make([]*types.Plugin, len(data.GetItems()))
	for idx, plugin := range data.GetItems() {
		item := &types.Plugin{
			TenantID:      plugin.GetTenantId(),
			PluginID:      plugin.GetPluginId(),
			PluginName:    plugin.GetPluginName(),
			PluginGroup:   plugin.GetPluginGroup(),
			PluginPkgName: plugin.GetPluginPkgName(),
		}

		plugins[idx] = item
	}

	return plugins, total
}
