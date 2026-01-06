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
	"google.golang.org/protobuf/types/known/structpb"
)

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
			Name:  exactCond.GetName(),
			Group: exactCond.GetGroup(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.PluginFuzzyFields{
			Name:    fuzzyCond.GetName(),
			PkgName: fuzzyCond.GetPkgName(),
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
			Name:  condition.ExactInclude.Name,
			Group: condition.ExactInclude.Group,
		}
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &PluginListReq_FuzzyConditions{
			Name:    condition.FuzzyInclude.Name,
			PkgName: condition.FuzzyInclude.PkgName,
		}
	}
}

const (
	// plugin list max limit
	maxPluginLimit = 500
)

// PageLimit returns page limit.
func (x *PluginListReq) PageLimit() int {
	return maxPluginLimit
}

// ConvertPageToTypes converts page to types.
func (x *PluginListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertPluginFromTypes converts plugin from types.
func (x *PluginListResp) ConvertPluginFromTypes(total int64, plugins []*types.Plugin) {
	items := make([]*PluginListResp_Plugin, len(plugins))
	for idx, plugin := range plugins {
		item := newEmptyPlugin()
		*item.TenantId = plugin.TenantID
		*item.Name = plugin.Name
		*item.Group = plugin.Group
		*item.PkgName = plugin.PkgName
		*item.Memo = plugin.Memo

		items[idx] = item
	}

	x.Data = &PluginListResp_Data{
		Total: total,
		Items: items,
	}
}

func newEmptyPlugin() *PluginListResp_Plugin {
	return &PluginListResp_Plugin{
		TenantId: new(string),
		Name:     new(string),
		Group:    new(string),
		PkgName:  new(string),
		Memo:     new(string),
	}
}

// ConvertPluginToTypes converts plugin to types.
func (x *PluginListResp) ConvertPluginToTypes() ([]*types.Plugin, int64) {
	data := x.GetData()
	total := data.GetTotal()
	plugins := make([]*types.Plugin, len(data.GetItems()))
	for idx, plugin := range data.GetItems() {
		item := &types.Plugin{
			TenantID: plugin.GetTenantId(),
			Name:     plugin.GetName(),
			Group:    plugin.GetGroup(),
			PkgName:  plugin.GetPkgName(),
			Memo:     plugin.GetMemo(),
		}

		plugins[idx] = item
	}

	return plugins, total
}

// Validate check body.
func (x *PluginInstallReq) Validate() error {
	plugins := x.GetPlugin()
	if len(plugins) == 0 {
		return errors.New("plugins can not be empty")
	}

	for idx := range plugins {
		if err := plugins[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *PluginInstallReq_Plugin) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be zero")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	if x.GetVersion() == "" {
		return errors.New("version can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginInstallReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *PluginInstallReq_Plugin) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginInstallReq) ConvertParamFromTypes(installParam ...*types.PluginDeploymentParam) error {
	var err error
	plugin := make([]*PluginInstallReq_Plugin, len(installParam))
	for idx, param := range installParam {
		item := &PluginInstallReq_Plugin{}
		item.BkHostId = &param.HostID
		item.PluginName = param.PluginName
		item.Version = param.Version
		item.ConfigName = param.ConfigName
		item.CustomConfigContext, err = structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return err
		}

		plugin[idx] = item
	}

	x.Plugin = plugin

	return nil
}

// ConvertParamToTypes converts param to types.
func (x *PluginInstallReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	plugin := x.GetPlugin()
	installParam := make([]*types.PluginDeploymentParam, len(plugin))
	for idx, proc := range plugin {
		item := &types.PluginDeploymentParam{
			HostID:              proc.GetBkHostId(),
			PluginName:          proc.GetPluginName(),
			Version:             proc.GetVersion(),
			ConfigName:          proc.GetConfigName(),
			CustomConfigContext: proc.GetCustomConfigContext().AsMap(),
		}

		installParam[idx] = item
	}

	return installParam
}

// Validate check body.
func (x *PluginUpgradeReq) Validate() error {
	plugins := x.GetPlugin()
	if len(plugins) == 0 {
		return errors.New("plugins can not be empty")
	}

	for idx := range plugins {
		if err := plugins[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *PluginUpgradeReq_Plugin) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be zero")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	if x.GetVersion() == "" {
		return errors.New("version can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginUpgradeReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *PluginUpgradeReq_Plugin) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginUpgradeReq) ConvertParamFromTypes(upgradeParam ...*types.PluginDeploymentParam) error {
	var err error
	plugin := make([]*PluginUpgradeReq_Plugin, len(upgradeParam))
	for idx, param := range upgradeParam {
		item := &PluginUpgradeReq_Plugin{}
		item.BkHostId = &param.HostID
		item.PluginName = param.PluginName
		item.Version = param.Version
		item.ConfigName = param.ConfigName
		item.CustomConfigContext, err = structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return err
		}

		plugin[idx] = item
	}

	x.Plugin = plugin

	return nil
}

// ConvertParamToTypes converts param to types.
func (x *PluginUpgradeReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	plugin := x.GetPlugin()
	upgradeParam := make([]*types.PluginDeploymentParam, len(plugin))
	for idx, proc := range plugin {
		item := &types.PluginDeploymentParam{
			HostID:              proc.GetBkHostId(),
			PluginName:          proc.GetPluginName(),
			Version:             proc.GetVersion(),
			ConfigName:          proc.GetConfigName(),
			CustomConfigContext: proc.GetCustomConfigContext().AsMap(),
		}

		upgradeParam[idx] = item
	}

	return upgradeParam
}

// Validate check body.
func (x *PluginApplySubConfigReq) Validate() error {
	plugins := x.GetPlugin()
	if len(plugins) == 0 {
		return errors.New("plugins can not be empty")
	}

	for idx := range plugins {
		if err := plugins[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *PluginApplySubConfigReq_Plugin) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be zero")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	if x.GetVersion() == "" {
		return errors.New("version can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginApplySubConfigReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *PluginApplySubConfigReq_Plugin) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginApplySubConfigReq) ConvertParamFromTypes(installParam ...*types.PluginDeploymentParam) error {
	plugin := make([]*PluginApplySubConfigReq_Plugin, len(installParam))
	for idx, param := range installParam {
		item := &PluginApplySubConfigReq_Plugin{}
		item.BkHostId = &param.HostID
		item.PluginName = param.PluginName
		item.Version = param.Version
		item.ConfigName = param.ConfigName
		customContext, err := structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return err
		}
		item.CustomConfigContext = customContext

		plugin[idx] = item
	}

	x.Plugin = plugin

	return nil
}

// ConvertParamToTypes converts param to types.
func (x *PluginApplySubConfigReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	plugin := x.GetPlugin()
	installParam := make([]*types.PluginDeploymentParam, len(plugin))
	for idx, proc := range plugin {
		item := &types.PluginDeploymentParam{
			HostID:              proc.GetBkHostId(),
			PluginName:          proc.GetPluginName(),
			Version:             proc.GetVersion(),
			CustomConfigContext: proc.GetCustomConfigContext().AsMap(),
			ConfigName:          proc.GetConfigName(),
		}

		installParam[idx] = item
	}

	return installParam
}

// Validate check body.
func (x *PluginSetMemoReq) Validate() error {
	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginSetMemoReq) AutoConvert() {}
