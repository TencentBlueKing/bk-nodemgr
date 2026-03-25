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

const (
	// config policy list max limit
	maxConfigPolicyLimit = 1000
)

// PageLimit return page limit.
func (x *ConfigPolicyListReq) PageLimit() int {
	return maxConfigPolicyLimit
}

// ConvertPageToTypes convert page to types.
func (x *ConfigPolicyListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
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
	if x.GetBizId() <= 0 {
		return fmt.Errorf("biz id is required")
	}

	if err := types.ConfigPolicyType(x.GetConfigpolicyType()).Validate(); err != nil {
		return fmt.Errorf("invalid config policy type: %w", err)
	}

	if len(x.GetConfigpolicyName()) == 0 {
		return fmt.Errorf("config policy name is required")
	}

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
		Name:          x.GetConfigpolicyName(),
		Type:          types.ConfigPolicyType(x.GetConfigpolicyType()),
		BizID:         x.GetBizId(),
		Remark:        x.GetRemark(),
		Scopes:        scopes,
		TargetHostIDs: x.GetTargetHostIds(),
		Configs:       convertConfigPolicyConfigsToTypes(x.GetConfigsString(), x.GetConfigsInt(), x.GetConfigsBool()),
		Operator:      x.GetOperator(),
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
	x.TargetHostIds = configPolicy.TargetHostIDs
}

// ConvertConfigPolicyID convert config policy id.
func (x *ConfigPolicyCreateResp) ConvertConfigPolicyID(configPolicyID int64) {
	data := &ConfigPolicyCreateResp_Data{ConfigpolicyId: new(int64)}
	*data.ConfigpolicyId = configPolicyID

	x.Data = data
}

// Validate check body.
func (x *ConfigPolicyUpdateReq) Validate() error {
	if x.GetConfigpolicyId() <= 0 {
		return fmt.Errorf("config policy id is required")
	}

	if x.GetBizId() <= 0 {
		return fmt.Errorf("biz id is required")
	}

	if err := types.ConfigPolicyType(x.GetConfigpolicyType()).Validate(); err != nil {
		return fmt.Errorf("invalid config policy type: %w", err)
	}

	if len(x.GetConfigpolicyName()) == 0 {
		return fmt.Errorf("config policy name is required")
	}

	if x.GetPriority() <= 0 {
		return fmt.Errorf("config policy priority is required")
	}

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
		ID:            x.GetConfigpolicyId(),
		Name:          x.GetConfigpolicyName(),
		Type:          types.ConfigPolicyType(x.GetConfigpolicyType()),
		BizID:         x.GetBizId(),
		Remark:        x.GetRemark(),
		Scopes:        scopes,
		TargetHostIDs: x.GetTargetHostIds(),
		Configs:       configs,
		Enabled:       x.GetEnabled(),
		Priority:      x.GetPriority(),
		Operator:      x.GetOperator(),
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
	x.TargetHostIds = configPolicy.TargetHostIDs
	x.Priority = configPolicy.Priority
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
	*item.BizId = configPolicy.BizID
	*item.Remark = configPolicy.Remark
	item.Scopes = scopes
	item.TargetHostIds = configPolicy.TargetHostIDs
	item.ConfigsString, item.ConfigsInt, item.ConfigsBool = convertConfigPolicyConfigsFromTypes(configPolicy.Configs)
	*item.Enabled = configPolicy.Enabled
	*item.UpdatedTime = configPolicy.UpdatedAt.UnixMilli()
	*item.Operator = configPolicy.Operator
	*item.Version = int64(configPolicy.Version)
	item.Priority = &configPolicy.Priority

	return item
}

func convertConfigPolicyToTypes(configPolicy *ConfigPolicy) *types.ConfigPolicy {
	scopes := make([]types.ConfigPolicyScope, len(configPolicy.GetScopes()))
	for idx, scope := range configPolicy.GetScopes() {
		scopes[idx] = convertConfigPolicyScopeToTypes(scope)
	}

	return &types.ConfigPolicy{
		TenantID:      configPolicy.GetTenantId(),
		ID:            configPolicy.GetConfigpolicyId(),
		Name:          configPolicy.GetConfigpolicyName(),
		Type:          types.ConfigPolicyType(configPolicy.GetConfigpolicyType()),
		BizID:         configPolicy.GetBizId(),
		Remark:        configPolicy.GetRemark(),
		Scopes:        scopes,
		TargetHostIDs: configPolicy.GetTargetHostIds(),
		Configs: convertConfigPolicyConfigsToTypes(
			configPolicy.ConfigsString,
			configPolicy.ConfigsInt,
			configPolicy.ConfigsBool),
		Enabled:   configPolicy.GetEnabled(),
		Priority:  configPolicy.GetPriority(),
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
		BizId:            new(int64),
		Remark:           new(string),
		Scopes:           make([]*ConfigPolicyScope, 0),
		ConfigsString:    make(map[string]string),
		ConfigsInt:       make(map[string]int64),
		ConfigsBool:      make(map[string]bool),
		Enabled:          new(bool),
		UpdatedTime:      new(int64),
		Operator:         new(string),
		Version:          new(int64),
		Priority:         new(int64),
	}
}

// Validate check body.
func (x *ConfigPolicyPriorityReorderReq) Validate() error {
	if x.GetBizId() <= 0 {
		return fmt.Errorf("biz_id is required")
	}

	if err := types.ConfigPolicyType(x.GetConfigpolicyType()).Validate(); err != nil {
		return fmt.Errorf("invalid configpolicy_type: %w", err)
	}

	if ids := x.GetOrderedConfigpolicyId(); len(conv.SliceUnique(ids)) != len(ids) {
		return fmt.Errorf("duplicated ordered_configpolicy_id")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyPriorityReorderReq) AutoConvert() {}

func newEmptyConfigPolicyScope() *ConfigPolicyScope {
	return &ConfigPolicyScope{
		BkNetworkareaId: new(int64),
		BkNetworkunitId: new(int64),
		OsType:          new(string),
		CpuArch:         new(string),
	}
}

// Validate check body.
func (x *ConfigPolicyPreviewReq) Validate() error {
	if x.GetBizId() <= 0 {
		return fmt.Errorf("biz_id is required")
	}

	if err := types.ConfigPolicyType(x.GetPolicyType()).Validate(); err != nil {
		return fmt.Errorf("invalid policy_type: %w", err)
	}

	if len(x.GetHosts()) == 0 {
		return fmt.Errorf("hosts is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyPreviewReq) AutoConvert() {
	for _, host := range x.GetHosts() {
		if host.HostId == nil {
			host.HostId = new(int64)
			*host.HostId = -1
		}
		if host.NetworkUnitId == nil {
			host.NetworkUnitId = new(int64)
			*host.NetworkUnitId = types.ConfigPolicyScopeAnyID
		}
		if host.NetworkAreaId == nil {
			host.NetworkAreaId = new(int64)
			*host.NetworkAreaId = types.ConfigPolicyScopeAnyID
		}
	}
}

// ConvertFromTypes populates the preview request from types values.
func (x *ConfigPolicyPreviewReq) ConvertFromTypes(
	bizID int64, policyType types.ConfigPolicyType, hosts []types.ConfigPolicyPreviewHost) {

	x.BizId = bizID
	x.PolicyType = string(policyType)

	protoHosts := make([]*PreviewHost, len(hosts))
	for i, host := range hosts {
		protoHosts[i] = &PreviewHost{
			HostId:        &host.HostID,
			OsType:        string(host.OSType),
			CpuArch:       string(host.CPUArch),
			NetworkUnitId: &host.NetworkUnitID,
			NetworkAreaId: &host.NetworkAreaID,
		}
	}
	x.Hosts = protoHosts
}

// ConvertPreviewHostsToTypes converts proto PreviewHost slice to types.
func (x *ConfigPolicyPreviewReq) ConvertPreviewHostsToTypes() []types.ConfigPolicyPreviewHost {
	hosts := make([]types.ConfigPolicyPreviewHost, len(x.GetHosts()))
	for i, rh := range x.GetHosts() {
		hosts[i] = types.ConfigPolicyPreviewHost{
			HostID:        rh.GetHostId(),
			NetworkAreaID: rh.GetNetworkAreaId(),
			NetworkUnitID: rh.GetNetworkUnitId(),
			OSType:        criteria.OSType(rh.GetOsType()),
			CPUArch:       criteria.CPUArch(rh.GetCpuArch()),
		}
	}
	return hosts
}

// ConvertMatchResultsFromTypes converts types.ConfigPolicyPreviewResult into response data.
func (x *ConfigPolicyPreviewResp) ConvertMatchResultsFromTypes(result *types.ConfigPolicyPreviewResult) {
	if result == nil {
		return
	}

	data := &ConfigPolicyPreviewResp_Data{
		ReliableItems:   convertPreviewMatchResults(result.ReliableResults),
		UnreliableItems: convertPreviewMatchResults(result.UnreliableResults),
	}

	x.Data = data
}

// ConvertMatchResultsToTypes converts response data into types.ConfigPolicyPreviewResult.
func (x *ConfigPolicyPreviewResp) ConvertMatchResultsToTypes() *types.ConfigPolicyPreviewResult {
	data := x.GetData()
	if data == nil {
		return &types.ConfigPolicyPreviewResult{}
	}

	result := &types.ConfigPolicyPreviewResult{
		ReliableResults:   convertPreviewItemsToMatchResults(data.GetReliableItems()),
		UnreliableResults: convertPreviewItemsToMatchResults(data.GetUnreliableItems()),
	}

	return result
}

func convertPreviewMatchResults(results []types.ConfigPolicyMatchResult) []*ConfigPolicyPreviewResp_PreviewItem {
	items := make([]*ConfigPolicyPreviewResp_PreviewItem, len(results))
	for i, result := range results {
		matchedPolicies := make([]*ConfigPolicyPreviewResp_MatchedPolicy, len(result.MatchedPolicies))
		for j, matchedPolicy := range result.MatchedPolicies {
			matchedPolicies[j] = &ConfigPolicyPreviewResp_MatchedPolicy{
				ConfigpolicyId:   matchedPolicy.PolicyID,
				ConfigpolicyName: matchedPolicy.PolicyName,
				Priority:         matchedPolicy.Priority,
			}
		}

		configsString, configsInt, configsBool := convertConfigPolicyConfigsFromTypes(result.MergedConfig)
		items[i] = &ConfigPolicyPreviewResp_PreviewItem{
			HostId:              result.HostID,
			MatchedPolicies:     matchedPolicies,
			MergedConfigsString: configsString,
			MergedConfigsInt:    configsInt,
			MergedConfigsBool:   configsBool,
		}
	}

	return items
}

func convertPreviewItemsToMatchResults(items []*ConfigPolicyPreviewResp_PreviewItem) []types.ConfigPolicyMatchResult {
	results := make([]types.ConfigPolicyMatchResult, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		matchedPolicies := make([]types.ConfigPolicyMatchedPolicy, 0, len(item.GetMatchedPolicies()))
		for _, matchedPolicy := range item.GetMatchedPolicies() {
			if matchedPolicy == nil {
				continue
			}

			matchedPolicies = append(matchedPolicies, types.ConfigPolicyMatchedPolicy{
				PolicyID:   matchedPolicy.GetConfigpolicyId(),
				PolicyName: matchedPolicy.GetConfigpolicyName(),
				Priority:   matchedPolicy.GetPriority(),
			})
		}

		results = append(results, types.ConfigPolicyMatchResult{
			HostID:          item.GetHostId(),
			MatchedPolicies: matchedPolicies,
			MergedConfig:    mergeMapsFromProto(item),
		})
	}

	return results
}

func mergeMapsFromProto(item *ConfigPolicyPreviewResp_PreviewItem) map[string]any {
	mergedConfig := make(map[string]any)
	for key, value := range item.GetMergedConfigsString() {
		mergedConfig[key] = value
	}
	for key, value := range item.GetMergedConfigsInt() {
		mergedConfig[key] = value
	}
	for key, value := range item.GetMergedConfigsBool() {
		mergedConfig[key] = value
	}

	return mergedConfig
}
