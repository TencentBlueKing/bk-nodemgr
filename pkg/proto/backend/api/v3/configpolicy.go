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
	"fmt"
	"reflect"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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
func (x *ConfigPolicyListReq) ConvertConditionsToTypes() (*types.ConfigPolicyCondition, error) {
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
	exactCond *ConfigPolicyExactConditions, fuzzyCond *ConfigPolicyFuzzyConditions) (*types.ConfigPolicyCondition, error) {

	condition := types.ConfigPolicyCondition{}

	configPolicyTypeList, err := types.StringListToConfigPolicyTypeList(exactCond.GetConfigpolicyType())
	if err != nil {
		return nil, fmt.Errorf("failed to convert config policy type list: %w", err)
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ConfigPolicyExactFields{
			ConfigPolicyID: exactCond.GetConfigpolicyId(),
			BizID:          exactCond.GetBizId(),
			Type:           configPolicyTypeList,
			Enabled:        exactCond.GetEnabled(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.ConfigPolicyFuzzyFields{
			ConfigPolicyName: fuzzyCond.GetConfigpolicyName(),
			Operator:         fuzzyCond.GetOperator(),
		}
	}

	return &condition, nil
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
		exactCond.ConfigpolicyType = types.ConfigPolicyTypeListToStringList(conditions.ExactInclude.Type)
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
		items[idx] = convertConfigPolicyFromTypes(configPolicy)
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
func (x *ConfigPolicyGetResp) ConvertConfigPolicyFromTypes(configPolicy *types.ConfigPolicy) {
	x.Data = convertConfigPolicyFromTypes(configPolicy)
}

// ConvertConfigPolicyToTypes convert config policy to types.
func (x *ConfigPolicyGetResp) ConvertConfigPolicyToTypes() *types.ConfigPolicy {
	return convertConfigPolicyToTypes(x.Data)
}

// Validate check body.
func (x *ConfigPolicyCreateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyCreateReq) AutoConvert() {
}

// ConvertConfigPolicyToTypes convert config policy to types.
func (x *ConfigPolicyCreateReq) ConvertConfigPolicyToTypes() *types.ConfigPolicy {
	scopes := make([]types.ConfigPolicyScope, len(x.GetScopes()))
	for idx, scope := range x.GetScopes() {
		scopes[idx] = convertConfigPolicyScopeToTypes(scope)
	}

	return &types.ConfigPolicy{
		Name:     x.GetConfigpolicyName(),
		Type:     types.ConfigPolicyType(x.GetConfigpolicyType()),
		BizID:    x.GetBizId(),
		Remark:   x.GetRemark(),
		Scopes:   scopes,
		Configs:  convertConfigPolicyConfigsToTypes(x.GetConfigsString(), x.GetConfigsInt(), x.GetConfigsBool()),
		Operator: x.GetOperator(),
	}
}

// ConvertConfigPolicyFromTypes convert config policy from types.
func (x *ConfigPolicyCreateReq) ConvertConfigPolicyFromTypes(configPolicy *types.ConfigPolicy) {
	scopes := make([]*ConfigPolicyScope, len(configPolicy.Scopes))
	for idx, scope := range configPolicy.Scopes {
		scopes[idx] = convertConfigPolicyScopeFromTypes(scope)
	}

	x.ConfigpolicyName = configPolicy.Name
	x.ConfigpolicyType = string(configPolicy.Type)
	x.BizId = configPolicy.BizID
	x.Remark = configPolicy.Remark
	x.Scopes = scopes
	x.ConfigsString, x.ConfigsInt, x.ConfigsBool = convertConfigPolicyConfigsFromTypes(configPolicy.Configs)
	x.Operator = configPolicy.Operator
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
func (x *ConfigPolicyUpdateReq) ConvertConfigPolicyToTypes() *types.ConfigPolicy {
	scopes := make([]types.ConfigPolicyScope, len(x.GetScopes()))
	for idx, scope := range x.GetScopes() {
		scopes[idx] = types.ConfigPolicyScope{
			NetworkAreaID: scope.GetBkNetworkareaId(),
			NetworkUnitID: scope.GetBkNetworkunitId(),
			NodeOsType:    criteria.OSType(scope.GetOsType()),
			NodeCPUArch:   criteria.CPUArch(scope.GetCpuArch()),
		}
	}
	configs := make(map[string]any)
	for key, value := range x.GetConfigsString() {
		configs[key] = value
	}
	for key, value := range x.GetConfigsInt() {
		configs[key] = value
	}
	for key, value := range x.GetConfigsBool() {
		configs[key] = value
	}

	return &types.ConfigPolicy{
		ID:       x.GetConfigpolicyId(),
		Name:     x.GetConfigpolicyName(),
		Type:     types.ConfigPolicyType(x.GetConfigpolicyType()),
		BizID:    x.GetBizId(),
		Remark:   x.GetRemark(),
		Scopes:   scopes,
		Configs:  configs,
		Enabled:  x.GetEnabled(),
		Operator: x.GetOperator(),
	}
}

// ConvertConfigPolicyFromTypes convert config policy from types.
func (x *ConfigPolicyUpdateReq) ConvertConfigPolicyFromTypes(configPolicy *types.ConfigPolicy) {
	scopes := make([]*ConfigPolicyScope, len(configPolicy.Scopes))
	for idx, scope := range configPolicy.Scopes {
		scopes[idx] = convertConfigPolicyScopeFromTypes(scope)
	}

	x.ConfigpolicyId = configPolicy.ID
	x.ConfigpolicyName = configPolicy.Name
	x.ConfigpolicyType = string(configPolicy.Type)
	x.BizId = configPolicy.BizID
	x.Remark = configPolicy.Remark
	x.Scopes = scopes
	x.ConfigsString, x.ConfigsInt, x.ConfigsBool = convertConfigPolicyConfigsFromTypes(configPolicy.Configs)
	x.Enabled = configPolicy.Enabled
	x.Operator = configPolicy.Operator
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

func convertConfigPolicyFromTypes(configPolicy *types.ConfigPolicy) *ConfigPolicy {
	scopes := make([]*ConfigPolicyScope, len(configPolicy.Scopes))
	for idx, scope := range configPolicy.Scopes {
		scopes[idx] = convertConfigPolicyScopeFromTypes(scope)
	}

	item := newEmptyConfigPolicy()
	*item.TenantId = configPolicy.TenantID
	*item.ConfigpolicyId = configPolicy.ID
	*item.ConfigpolicyName = configPolicy.Name
	*item.ConfigpolicyType = string(configPolicy.Type)
	item.BizId = configPolicy.BizID
	*item.Remark = configPolicy.Remark
	item.Scopes = scopes
	item.ConfigsString, item.ConfigsInt, item.ConfigsBool = convertConfigPolicyConfigsFromTypes(configPolicy.Configs)
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
		TenantID: configPolicy.GetTenantId(),
		ID:       configPolicy.GetConfigpolicyId(),
		Name:     configPolicy.GetConfigpolicyName(),
		Type:     types.ConfigPolicyType(configPolicy.GetConfigpolicyType()),
		BizID:    configPolicy.GetBizId(),
		Remark:   configPolicy.GetRemark(),
		Scopes:   scopes,
		Configs: convertConfigPolicyConfigsToTypes(
			configPolicy.ConfigsString,
			configPolicy.ConfigsInt,
			configPolicy.ConfigsBool),
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

func convertConfigPolicyConfigsFromTypes(
	configs map[string]any) (map[string]string, map[string]int64, map[string]bool) {

	configsString := make(map[string]string)
	configsInt := make(map[string]int64)
	configsBool := make(map[string]bool)
	for key, value := range configs {
		switch reflect.TypeOf(value).Kind() {
		case reflect.String:
			configsString[key], _ = value.(string)

		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
			configsInt[key], _ = conv.ToInt64(value)

		case reflect.Bool:
			configsBool[key], _ = value.(bool)

		default:
			continue
		}
	}

	return configsString, configsInt, configsBool
}

func convertConfigPolicyConfigsToTypes(
	configsString map[string]string, configsInt map[string]int64, configsBool map[string]bool) map[string]any {

	configs := make(map[string]any)
	for key, value := range configsString {
		configs[key] = value
	}
	for key, value := range configsInt {
		configs[key] = value
	}
	for key, value := range configsBool {
		configs[key] = value
	}

	return configs
}

func newEmptyConfigPolicy() *ConfigPolicy {
	return &ConfigPolicy{
		TenantId:         new(string),
		ConfigpolicyId:   new(int64),
		ConfigpolicyName: new(string),
		ConfigpolicyType: new(string),
		BizId:            make([]int64, 0),
		Remark:           new(string),
		Scopes:           make([]*ConfigPolicyScope, 0),
		ConfigsString:    make(map[string]string),
		ConfigsInt:       make(map[string]int64),
		ConfigsBool:      make(map[string]bool),
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
