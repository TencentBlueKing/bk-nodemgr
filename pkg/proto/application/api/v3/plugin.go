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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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

func convertPluginConditionsToTypes(exactCond *PluginListExactConditions, fuzzyCond *PluginListFuzzyConditions) *types.PluginCondition {
	condition := &types.PluginCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.PluginExactFields{
			Name:          exactCond.GetName(),
			Group:         exactCond.GetGroup(),
			VisibleBizIDs: exactCond.GetVisibleBizIds(),
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
		x.ExactIncludeConditions = &PluginListExactConditions{
			Name:          condition.ExactInclude.Name,
			Group:         condition.ExactInclude.Group,
			VisibleBizIds: condition.ExactInclude.VisibleBizIDs,
		}
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &PluginListFuzzyConditions{
			Name:    condition.FuzzyInclude.Name,
			PkgName: condition.FuzzyInclude.PkgName,
		}
	}
}

// ConvertPageToTypes converts page to types.
func (x *PluginListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertPluginFromTypes converts plugin from types.
func (x *PluginListResp) ConvertPluginFromTypes(total int64, plugins []*types.Plugin) {
	x.Data = &PluginListResp_Data{
		Total: total,
		Items: conv.SliceToSlice(plugins, func(plugin *types.Plugin) *Plugin {
			item := newEmptyPlugin()
			*item.TenantId = plugin.TenantID
			*item.Name = plugin.Name
			*item.Group = plugin.Group
			*item.PkgName = plugin.PkgName
			*item.Memo = plugin.Memo

			return item
		}),
	}
}

// ConvertPluginToTypes converts plugin to types.
func (x *PluginListResp) ConvertPluginToTypes() ([]*types.Plugin, int64) {
	return conv.SliceToSlice(x.GetData().GetItems(), func(plugin *Plugin) *types.Plugin {
		return &types.Plugin{
			TenantID: plugin.GetTenantId(),
			Name:     plugin.GetName(),
			Group:    plugin.GetGroup(),
			PkgName:  plugin.GetPkgName(),
			Memo:     plugin.GetMemo(),
		}
	}), x.GetData().GetTotal()
}

func newEmptyPlugin() *Plugin {
	return &Plugin{
		TenantId:      new(string),
		Name:          new(string),
		Group:         new(string),
		PkgName:       new(string),
		Memo:          new(string),
		VisibleBizIds: make([]int64, 0),
	}
}

// Validate check body.
// nolint: protogetter
func (x *PluginUpgradeReq_UpgradeInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be negative")
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
func (x *PluginUpgradeReq_UpgradeInfo) AutoConvert() {}

// Validate check body.
// nolint: protogetter
func (x *PluginApplySubConfigReq_ApplySubConfigInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be negative")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginApplySubConfigReq_ApplySubConfigInfo) AutoConvert() {}

// Validate check body.
// nolint: protogetter
func (x *PluginUninstallReq_UninstallInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be negative")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginUninstallReq_UninstallInfo) AutoConvert() {}

// Validate check body.
// nolint: protogetter
func (x *PluginRestartReq_RestartInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be negative")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginRestartReq_RestartInfo) AutoConvert() {}

// Validate check body.
// nolint: protogetter
func (x *PluginStopReq_StopInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be negative")
	}

	if x.GetPluginName() == "" {
		return errors.New("plugin_name can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginStopReq_StopInfo) AutoConvert() {}

// Validate check body.
func (x *PluginInstallReq) Validate() error {
	plugins := x.GetPlugin()
	if len(plugins) == 0 {
		return errors.New("plugins can not be empty")
	}

	for _, plugin := range plugins {
		if plugin.GetBkHostId() <= 0 {
			return errors.New("bk_host_id can not be negative")
		}

		if plugin.GetPluginName() == "" {
			return errors.New("plugin_name can not be empty")
		}

		if plugin.GetVersion() == "" {
			return errors.New("version can not be empty")
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginInstallReq) AutoConvert() {}

// ConvertInstallParamToTypes converts install param to types.
func (x *PluginInstallReq) ConvertInstallParamToTypes() *types.PluginInstallParam {
	return &types.PluginInstallParam{
		Plugins: x.ConvertParamToTypes(),
	}

}

// ConvertParamToTypes converts param to types.
func (x *PluginInstallReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	return conv.SliceToSlice(x.GetPlugin(), func(plugin *PluginInstallReq_InstallInfo) *types.PluginDeploymentParam {
		var customSpec *types.PluginSpec
		if plugin.GetCustomSpec() != nil {
			customSpec = &types.PluginSpec{
				Resource: types.ProcessResource{
					CPULimitPercent: plugin.GetCustomSpec().GetResource().GetCpuLimitPercent(),
					MemLimitPercent: plugin.GetCustomSpec().GetResource().GetMemLimitPercent(),
				},
				MonitorPolicy: types.ProcessMonitorPolicy{
					RestartType:    types.ProcessRestartType(plugin.GetCustomSpec().GetMonitorPolicy().GetRestartType()),
					StartCheckSecs: plugin.GetCustomSpec().GetMonitorPolicy().GetStartCheckSeconds(),
					StopCheckSecs:  plugin.GetCustomSpec().GetMonitorPolicy().GetStopCheckSeconds(),
					OpTimeoutSecs:  plugin.GetCustomSpec().GetMonitorPolicy().GetOperateTimeoutSeconds(),
				},
			}
		}

		return &types.PluginDeploymentParam{
			HostID:              plugin.GetBkHostId(),
			PluginName:          plugin.GetPluginName(),
			Version:             plugin.GetVersion(),
			ConfigTemplateName:  plugin.GetConfigTemplateName(),
			CustomConfigContext: plugin.GetCustomConfigContext().AsMap(),
			CustomSpec:          customSpec,
		}
	})
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

// AutoConvert auto convert.
func (x *PluginUpgradeReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginUpgradeReq) ConvertParamFromTypes(upgradeParam ...*types.PluginDeploymentParam) error {
	var err error
	x.Plugin, err = conv.SliceToSliceWithError(upgradeParam, func(param *types.PluginDeploymentParam) (*PluginUpgradeReq_UpgradeInfo, error) {
		item := &PluginUpgradeReq_UpgradeInfo{}
		item.BkHostId = param.HostID
		item.PluginName = param.PluginName
		item.Version = param.Version
		item.ConfigTemplateName = param.ConfigTemplateName
		item.CustomConfigContext, err = structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return nil, err
		}

		return item, nil
	})

	return nil
}

// ConvertParamToTypes converts param to types.
func (x *PluginUpgradeReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	return conv.SliceToSlice(x.GetPlugin(), func(plugin *PluginUpgradeReq_UpgradeInfo) *types.PluginDeploymentParam {
		return &types.PluginDeploymentParam{
			HostID:              plugin.GetBkHostId(),
			PluginName:          plugin.GetPluginName(),
			Version:             plugin.GetVersion(),
			ConfigTemplateName:  plugin.GetConfigTemplateName(),
			CustomConfigContext: plugin.GetCustomConfigContext().AsMap(),
		}
	})
}

// Validate check body.
func (x *PluginUninstallReq) Validate() error {
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

// AutoConvert auto convert.
func (x *PluginUninstallReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginUninstallReq) ConvertParamFromTypes(uninstallParam ...*types.PluginDeploymentParam) {
	x.Plugin = conv.SliceToSlice(uninstallParam, func(param *types.PluginDeploymentParam) *PluginUninstallReq_UninstallInfo {
		item := &PluginUninstallReq_UninstallInfo{}
		item.BkHostId = param.HostID
		item.PluginName = param.PluginName

		return item
	})
}

// ConvertParamToTypes converts param to types.
func (x *PluginUninstallReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	return conv.SliceToSlice(x.GetPlugin(), func(plugin *PluginUninstallReq_UninstallInfo) *types.PluginDeploymentParam {
		return &types.PluginDeploymentParam{
			HostID:     plugin.GetBkHostId(),
			PluginName: plugin.GetPluginName(),
		}
	})
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

// AutoConvert auto convert.
func (x *PluginApplySubConfigReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginApplySubConfigReq) ConvertParamFromTypes(installParam ...*types.PluginDeploymentParam) error {
	var err error
	x.Plugin, err = conv.SliceToSliceWithError(installParam, func(param *types.PluginDeploymentParam) (*PluginApplySubConfigReq_ApplySubConfigInfo, error) {
		item := &PluginApplySubConfigReq_ApplySubConfigInfo{}
		item.BkHostId = param.HostID
		item.PluginName = param.PluginName
		item.ConfigTemplateName = param.ConfigTemplateName
		item.CustomConfigContext, err = structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return nil, err
		}

		return item, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ConvertParamToTypes converts param to types.
func (x *PluginApplySubConfigReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	return conv.SliceToSlice(x.GetPlugin(), func(plugin *PluginApplySubConfigReq_ApplySubConfigInfo) *types.PluginDeploymentParam {
		return &types.PluginDeploymentParam{
			HostID:              plugin.GetBkHostId(),
			PluginName:          plugin.GetPluginName(),
			ConfigTemplateName:  plugin.GetConfigTemplateName(),
			CustomConfigContext: plugin.GetCustomConfigContext().AsMap(),
		}
	})
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

// Validate check body.
func (x *PluginListPermittedOperationReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PluginListPermittedOperationReq) AutoConvert() {}

// ConvertPageToTypes converts page to types.
func (x *PluginListPermittedOperationReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertPluginPermittedOperationsFromTypes converts plugin permitted operations from types.
func (x *PluginListPermittedOperationResp) ConvertPluginPermittedOperationsFromTypes(plugins ...*types.Plugin) {
	x.Data = &PluginListPermittedOperationResp_Data{}
	x.Data.Operations = conv.SliceToSlice(plugins, func(plugin *types.Plugin) *PluginListPermittedOperationResp_Data_Operation {
		item := &PluginListPermittedOperationResp_Data_Operation{}
		item.Name = plugin.Name

		// default group and policy group have different permitted operations.
		if plugin.Group == types.PluginGroupDefault {
			item.Permission = conv.SliceToSlice(types.DefaultGroupPermittedOperations(), func(operation types.PermittedOperation) string {
				return operation.String()
			})
		} else {
			item.Permission = conv.SliceToSlice(types.PolicyGroupPermittedOperations(), func(operation types.PermittedOperation) string {
				return operation.String()
			})
		}

		return item
	})
}

// Validate check body.
func (x *PluginRestartReq) Validate() error {
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

// AutoConvert auto convert.
func (x *PluginRestartReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginRestartReq) ConvertParamFromTypes(restartParam ...*types.PluginDeploymentParam) {
	x.Plugin = conv.SliceToSlice(restartParam, func(param *types.PluginDeploymentParam) *PluginRestartReq_RestartInfo {
		item := &PluginRestartReq_RestartInfo{}
		item.BkHostId = param.HostID
		item.PluginName = param.PluginName

		return item
	})
}

// ConvertParamToTypes converts param to types.
func (x *PluginRestartReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	return conv.SliceToSlice(x.GetPlugin(), func(plugin *PluginRestartReq_RestartInfo) *types.PluginDeploymentParam {
		return &types.PluginDeploymentParam{
			HostID:     plugin.GetBkHostId(),
			PluginName: plugin.GetPluginName(),
		}
	})
}

// Validate check body.
func (x *PluginStopReq) Validate() error {
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

// AutoConvert auto convert.
func (x *PluginStopReq) AutoConvert() {
	plugin := x.GetPlugin()
	for idx := range plugin {
		plugin[idx].AutoConvert()
	}
}

// ConvertParamFromTypes converts param from types.
func (x *PluginStopReq) ConvertParamFromTypes(stopParam ...*types.PluginDeploymentParam) {
	x.Plugin = conv.SliceToSlice(stopParam, func(param *types.PluginDeploymentParam) *PluginStopReq_StopInfo {
		item := &PluginStopReq_StopInfo{}
		item.BkHostId = param.HostID
		item.PluginName = param.PluginName

		return item
	})
}

// ConvertParamToTypes converts param to types.
func (x *PluginStopReq) ConvertParamToTypes() []*types.PluginDeploymentParam {
	return conv.SliceToSlice(x.GetPlugin(), func(plugin *PluginStopReq_StopInfo) *types.PluginDeploymentParam {
		return &types.PluginDeploymentParam{
			HostID:     plugin.GetBkHostId(),
			PluginName: plugin.GetPluginName(),
		}
	})
}
