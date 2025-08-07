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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *ConfigPolicyListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *ConfigPolicyListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *ConfigPolicyListReq) ConvertConditionsToTypes() *types.ConfigPolicyCondition {
	return convertConfigPolicyConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *ConfigPolicyListReq) ConvertConditionsFromTypes(condition *types.ConfigPolicyCondition) error {
	exactCond, fuzzyCond, err := convertConfigPolicyConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

func convertConfigPolicyConditionsToTypes(
	exactCond *ConfigPolicyExactConditions, fuzzyCond *ConfigPolicyFuzzyConditions) *types.ConfigPolicyCondition {

	condition := types.ConfigPolicyCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ConfigPolicyExactFields{
			ConfigPolicyID: exactCond.GetConfigpolicyId(),
			BizID:          exactCond.GetBizId(),
			NodeRole:       types.StringListToNodeRoleList(exactCond.GetNodeRole()),
			Enabled:        exactCond.GetEnabled(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.ConfigPolicyFuzzyFields{
			ConfigPolicyName: fuzzyCond.GetConfigpolicyName(),
		}
	}

	return &condition
}

func convertConfigPolicyConditionsFromTypes(conditions *types.ConfigPolicyCondition) (
	*ConfigPolicyExactConditions, *ConfigPolicyFuzzyConditions, error) {

	if conditions == nil {
		return nil, nil, nil
	}

	var exactCond *ConfigPolicyExactConditions
	var fuzzyCond *ConfigPolicyFuzzyConditions

	if conditions.ExactInclude != nil {
		exactCond = new(ConfigPolicyExactConditions)
		exactCond.ConfigpolicyId = conditions.ExactInclude.ConfigPolicyID
		exactCond.BizId = conditions.ExactInclude.BizID
		exactCond.NodeRole = types.NodeRoleListToStringList(conditions.ExactInclude.NodeRole)
		exactCond.Enabled = conditions.ExactInclude.Enabled
	}

	if conditions.FuzzyInclude != nil {
		fuzzyCond = new(ConfigPolicyFuzzyConditions)
		fuzzyCond.ConfigpolicyName = conditions.FuzzyInclude.ConfigPolicyName
	}

	if conditions.ExactExclude != nil || conditions.FuzzyExclude != nil {
		return nil, nil, errors.New("exact-exclude, fuzzy-exclude are not supported")
	}

	return exactCond, fuzzyCond, nil
}

// ConvertConfigPoliciesFromTypes convert config policies from types.
func (x *ConfigPolicyListResp) ConvertConfigPoliciesFromTypes(total int64, configPolicies []*types.ConfigPolicy) {
	items := make([]*ConfigPolicy, len(configPolicies))
	for idx, configPolicy := range configPolicies {
		items[idx] = convertConfigPolicyFromTypes(configPolicy, nil)
	}

	x.Data = &ConfigPolicyListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertConfigPoliciesToTypes convert config policies to types.
func (x *ConfigPolicyListResp) ConvertConfigPoliciesToTypes() (int64, []*types.ConfigPolicy) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.ConfigPolicy, len(items))
	for idx, item := range items {
		result[idx] = convertConfigPolicyToTypes(item)
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *ConfigPolicyGetReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyGetReq) AutoConvert() {
}

// ConvertConfigPolicyFromTypes convert config policy from types.
func (x *ConfigPolicyGetResp) ConvertConfigPolicyFromTypes(configPolicy *types.ConfigPolicy, blocks []types.ConfigPolicyTemplateBlock) {
	x.Data = convertConfigPolicyFromTypes(configPolicy, blocks)
}

// Validate check body.
func (x *ConfigPolicyGetTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyGetTemplateReq) AutoConvert() {
}

// ConvertTemplateFromTypes convert template from types.
func (x *ConfigPolicyGetTemplateResp) ConvertTemplateFromTypes(blocks []types.ConfigPolicyTemplateBlock) {
	x.Data = &ConfigPolicyGetTemplateResp_Data{
		Templates: convertConfigPolicyConfigsFromTypes(blocks),
	}
}

// Validate check body.
func (x *ConfigPolicyListPlatformReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyListPlatformReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *ConfigPolicyListPlatformReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	releaseType, err := types.ConvertNodeRoleToReleaseType(types.NodeRole(x.GetNodeRole()))
	if err != nil {
		return &types.ReleaseCondition{}
	}

	return &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Type: []types.ReleaseType{releaseType},
		},
	}
}

// ConvertPlatformFromTypes convert platform from types.
func (x *ConfigPolicyListPlatformResp) ConvertPlatformFromTypes(result *types.ReleaseDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &ConfigPolicyListPlatformResp_Data{
		OsType:  result.OSType,
		CpuArch: result.CPUArch,
	}
}

// Validate check body.
func (x *ConfigPolicyCreateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyCreateReq) AutoConvert() {
}

// ConvertConfigPolicyToTypes convert config policy to types.
func (x *ConfigPolicyCreateReq) ConvertConfigPolicyToTypes() (*types.ConfigPolicy, []types.ConfigPolicyTemplateBlock) {
	scopes := make([]types.ConfigPolicyScope, len(x.GetScopes()))
	for idx, scope := range x.GetScopes() {
		scopes[idx] = convertConfigPolicyScopeToTypes(scope)
	}

	return &types.ConfigPolicy{
		Name:     x.GetConfigpolicyName(),
		NodeRole: types.NodeRole(x.GetNodeRole()),
		BizID:    x.GetBizId(),
		Remark:   x.GetRemark(),
		Scopes:   scopes,
		Operator: x.GetOperator(),
	}, convertConfigPolicyConfigsToTypes(x.GetConfigs())
}

// ConvertConfigPolicyID convert config policy id.
func (x *ConfigPolicyCreateResp) ConvertConfigPolicyID(configPolicyID int64) {
	data := &ConfigPolicyCreateResp_Data{ConfigpolicyId: new(int64)}
	*data.ConfigpolicyId = configPolicyID

	x.Data = data
}

// Validate check body.
func (x *ConfigPolicyUpdateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyUpdateReq) AutoConvert() {
}

// ConvertConfigPolicyToTypes convert config policy to types.
func (x *ConfigPolicyUpdateReq) ConvertConfigPolicyToTypes() (*types.ConfigPolicy, []types.ConfigPolicyTemplateBlock) {
	scopes := make([]types.ConfigPolicyScope, len(x.GetScopes()))
	for idx, scope := range x.GetScopes() {
		scopes[idx] = convertConfigPolicyScopeToTypes(scope)
	}

	return &types.ConfigPolicy{
		ID:       x.GetConfigpolicyId(),
		Name:     x.GetConfigpolicyName(),
		NodeRole: types.NodeRole(x.GetNodeRole()),
		BizID:    x.GetBizId(),
		Remark:   x.GetRemark(),
		Scopes:   scopes,
		Enabled:  x.GetEnabled(),
		Operator: x.GetOperator(),
	}, convertConfigPolicyConfigsToTypes(x.GetConfigs())
}

// ConvertConfigPolicyID convert config policy id.
func (x *ConfigPolicyUpdateResp) ConvertConfigPolicyID(configPolicyID int64) {
	data := &ConfigPolicyUpdateResp_Data{ConfigpolicyId: new(int64)}
	*data.ConfigpolicyId = configPolicyID

	x.Data = data
}

// Validate check body.
func (x *ConfigPolicyEnableReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyEnableReq) AutoConvert() {
}

// Validate check body.
func (x *ConfigPolicyDisableReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyDisableReq) AutoConvert() {
}

// Validate check body.
func (x *ConfigPolicyDeleteReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyDeleteReq) AutoConvert() {
}

func convertConfigPolicyFromTypes(configPolicy *types.ConfigPolicy, blocks []types.ConfigPolicyTemplateBlock) *ConfigPolicy {
	scopes := make([]*ConfigPolicyScope, len(configPolicy.Scopes))
	for idx, scope := range configPolicy.Scopes {
		scopes[idx] = convertConfigPolicyScopeFromTypes(scope)
	}

	item := newEmptyConfigPolicy()
	*item.TenantId = configPolicy.TenantID
	*item.ConfigpolicyId = configPolicy.ID
	*item.ConfigpolicyName = configPolicy.Name
	*item.NodeRole = string(configPolicy.NodeRole)
	item.BizId = configPolicy.BizID
	*item.Remark = configPolicy.Remark
	item.Scopes = scopes
	item.Configs = convertConfigPolicyConfigsFromTypes(blocks)
	*item.Enabled = configPolicy.Enabled
	*item.UpdatedTime = configPolicy.UpdatedAt.UnixMilli()
	*item.Operator = configPolicy.Operator
	*item.Version = int64(configPolicy.Version)

	return item
}

func convertConfigPolicyToTypes(configPolicy *ConfigPolicy) *types.ConfigPolicy {
	scopes := make([]types.ConfigPolicyScope, len(configPolicy.GetScopes()))
	for idx, scope := range configPolicy.GetScopes() {
		scopes[idx] = convertConfigPolicyScopeToTypes(scope)
	}

	return &types.ConfigPolicy{
		TenantID:  configPolicy.GetTenantId(),
		ID:        configPolicy.GetConfigpolicyId(),
		Name:      configPolicy.GetConfigpolicyName(),
		NodeRole:  types.NodeRole(configPolicy.GetNodeRole()),
		BizID:     configPolicy.GetBizId(),
		Remark:    configPolicy.GetRemark(),
		Scopes:    scopes,
		Enabled:   configPolicy.GetEnabled(),
		UpdatedAt: time.UnixMilli(configPolicy.GetUpdatedTime()),
		Operator:  configPolicy.GetOperator(),
		Version:   int(configPolicy.GetVersion()),
	}
}

func convertConfigPolicyScopeFromTypes(scope types.ConfigPolicyScope) *ConfigPolicyScope {
	item := newEmptyConfigPolicyScope()
	*item.BkNetworkareaId = int64(scope.NetworkAreaID)
	*item.BkNetworkunitId = int64(scope.NetworkUnitID)
	*item.OsType = string(scope.NodeOsType)
	*item.CpuArch = string(scope.NodeCPUArch)

	return item
}

func convertConfigPolicyScopeToTypes(scope *ConfigPolicyScope) types.ConfigPolicyScope {
	return types.ConfigPolicyScope{
		NetworkAreaID: scope.GetBkNetworkareaId(),
		NetworkUnitID: scope.GetBkNetworkunitId(),
		NodeOsType:    criteria.OSType(scope.GetOsType()),
		NodeCPUArch:   criteria.CPUArch(scope.GetCpuArch()),
	}
}

func convertConfigPolicyConfigsFromTypes(blocks []types.ConfigPolicyTemplateBlock) []*ConfigPolicyConfigBlock {
	data := make([]*ConfigPolicyConfigBlock, len(blocks))

	for idx, block := range blocks {
		items := make([]*ConfigPolicyConfigItem, len(block.Items))
		for subIdx, item := range block.Items {
			subData := newEmptyConfigPolicyConfigItem()
			*subData.Id = item.ID
			*subData.Enabled = item.Enabled
			*subData.NameEn = item.NameEN
			*subData.NameZh = item.NameZH
			*subData.RemarkEn = item.RemarkEN
			*subData.RemarkZh = item.RemarkZH
			*subData.Key = item.Key
			*subData.Type = int64(item.Type)
			*subData.ValueString = item.ValueString
			*subData.ValueInt = item.ValueInt
			*subData.ValueBool = item.ValueBool
			subData.ValueStringSelect = item.ValueStringSelect
			subData.ValueIntSelect = item.ValueIntSelect

			items[subIdx] = subData
		}

		blockData := newEmptyConfigPolicyConfigBlock()
		*blockData.Id = block.ID
		*blockData.TitleEn = block.TitleEN
		*blockData.TitleZh = block.TitleZH
		blockData.Items = items

		data[idx] = blockData
	}

	return data
}

func convertConfigPolicyConfigsToTypes(blocks []*ConfigPolicyConfigBlock) []types.ConfigPolicyTemplateBlock {
	data := make([]types.ConfigPolicyTemplateBlock, len(blocks))

	for idx, block := range blocks {
		items := make([]types.ConfigPolicyTemplateItem, len(block.Items))
		for subIdx, item := range block.Items {
			items[subIdx] = types.ConfigPolicyTemplateItem{
				ID:                item.GetId(),
				Enabled:           item.GetEnabled(),
				NameEN:            item.GetNameEn(),
				NameZH:            item.GetNameZh(),
				RemarkEN:          item.GetRemarkEn(),
				RemarkZH:          item.GetRemarkZh(),
				Key:               item.GetKey(),
				Type:              types.ConfigPolicyTemplateType(item.GetType()),
				ValueString:       item.GetValueString(),
				ValueInt:          item.GetValueInt(),
				ValueBool:         item.GetValueBool(),
				ValueStringSelect: item.GetValueStringSelect(),
				ValueIntSelect:    item.GetValueIntSelect(),
			}
		}

		data[idx] = types.ConfigPolicyTemplateBlock{
			ID:      block.GetId(),
			TitleEN: block.GetTitleEn(),
			TitleZH: block.GetTitleZh(),
			Items:   items,
		}
	}

	return data
}

func newEmptyConfigPolicy() *ConfigPolicy {
	return &ConfigPolicy{
		TenantId:         new(string),
		ConfigpolicyId:   new(int64),
		ConfigpolicyName: new(string),
		NodeRole:         new(string),
		BizId:            make([]int64, 0),
		Remark:           new(string),
		Scopes:           make([]*ConfigPolicyScope, 0),
		Configs:          make([]*ConfigPolicyConfigBlock, 0),
		Enabled:          new(bool),
		UpdatedTime:      new(int64),
		Operator:         new(string),
		Version:          new(int64),
	}
}

func newEmptyConfigPolicyScope() *ConfigPolicyScope {
	return &ConfigPolicyScope{
		BkNetworkareaId: new(int64),
		BkNetworkunitId: new(int64),
		OsType:          new(string),
		CpuArch:         new(string),
	}
}

func newEmptyConfigPolicyConfigBlock() *ConfigPolicyConfigBlock {
	return &ConfigPolicyConfigBlock{
		Id:      new(string),
		TitleEn: new(string),
		TitleZh: new(string),
		Items:   make([]*ConfigPolicyConfigItem, 0),
	}
}

func newEmptyConfigPolicyConfigItem() *ConfigPolicyConfigItem {
	return &ConfigPolicyConfigItem{
		Id:                new(string),
		Enabled:           new(bool),
		NameEn:            new(string),
		NameZh:            new(string),
		RemarkEn:          new(string),
		RemarkZh:          new(string),
		Key:               new(string),
		Type:              new(int64),
		ValueString:       new(string),
		ValueInt:          new(int64),
		ValueBool:         new(bool),
		ValueStringSelect: make([]string, 0),
		ValueIntSelect:    make([]int64, 0),
	}
}
