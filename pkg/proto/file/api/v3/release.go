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
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertReleaseConditionsToTypes(items []*ReleaseConditions) []*types.ReleaseCondition {
	result := make([]*types.ReleaseCondition, len(items))
	for index, item := range items {
		condition := &types.ReleaseCondition{}
		if exact := item.GetExactInclude(); exact != nil {
			platforms := make([]platform.Platform, 0, len(exact.GetPlatform()))
			for _, platform := range exact.GetPlatform() {
				platforms = append(platforms, ConvertPlatformToTypes(platform))
			}
			generations := make([]types.Generation, len(exact.GetGeneration()))
			for generationIndex, generation := range exact.GetGeneration() {
				generations[generationIndex] = types.Generation(generation)
			}
			condition.ExactInclude = &types.ReleaseExactFields{
				Name:       exact.GetName(),
				FileName:   exact.GetFileName(),
				Generation: generations,
				Platform:   platforms,
				Version:    exact.GetVersion(),
				AsDefault:  exact.GetAsDefault(),
				Enabled:    exact.GetEnabled(),
				IsHidden:   exact.GetIsHidden(),
				IsShared:   exact.GetIsShared(),
				IsSynced:   exact.GetIsSynced(),
			}
		}
		result[index] = condition
	}
	return result
}

func convertReleaseConditionsFromTypes(conditions ...*types.ReleaseCondition) ([]*ReleaseConditions, error) {
	result := make([]*ReleaseConditions, len(conditions))
	for index, condition := range conditions {
		if condition == nil {
			result[index] = &ReleaseConditions{}
			continue
		}
		if condition.FuzzyInclude != nil || condition.ExactExclude != nil || condition.FuzzyExclude != nil {
			return nil, fmt.Errorf("fuzzy-include, exact-exclude, fuzzy-exclude are not supported")
		}
		item := &ReleaseConditions{}
		if exact := condition.ExactInclude; exact != nil {
			platforms := make([]*Platform, 0, len(exact.Platform))
			for _, platform := range exact.Platform {
				platforms = append(platforms, ConvertPlatformFromTypes(platform))
			}
			generations := make([]int64, len(exact.Generation))
			for generationIndex, generation := range exact.Generation {
				generations[generationIndex] = int64(generation)
			}
			item.ExactInclude = &ReleaseExactConditions{
				Name:       exact.Name,
				FileName:   exact.FileName,
				Generation: generations,
				Platform:   platforms,
				Version:    exact.Version,
				AsDefault:  exact.AsDefault,
				Enabled:    exact.Enabled,
				IsHidden:   exact.IsHidden,
				IsShared:   exact.IsShared,
				IsSynced:   exact.IsSynced,
			}
		}
		result[index] = item
	}
	return result, nil
}

func convertReleaseFromTypes(release *types.Release) (*ReleaseInfo, error) {
	if release == nil {
		return nil, nil
	}
	var additionInfo *structpb.Struct
	if release.AdditionInfo != nil {
		jsonBytes, err := json.Marshal(release.AdditionInfo)
		if err != nil {
			return nil, fmt.Errorf("marshal release addition info: %w", err)
		}

		additionInfo = new(structpb.Struct)
		if err := additionInfo.UnmarshalJSON(jsonBytes); err != nil {
			return nil, fmt.Errorf("unmarshal release addition info: %w", err)
		}
	}

	return &ReleaseInfo{
		Name:         release.Name,
		Generation:   int64(release.Generation),
		ReleaseType:  string(release.Type),
		Platform:     ConvertPlatformFromTypes(release.Platform),
		Version:      release.Version,
		Labels:       release.Labels,
		FileName:     release.FileName,
		Md5:          release.MD5,
		Enabled:      release.Enabled,
		IsHidden:     release.IsHidden,
		IsShared:     release.IsShared,
		IsSynced:     release.IsSynced,
		AsDefault:    release.AsDefault,
		UpdatedAt:    timestamppb.New(release.UpdatedAt),
		Operator:     release.Operator,
		AdditionInfo: additionInfo,
	}, nil
}

func convertReleaseToTypes(release *ReleaseInfo) (*types.Release, error) {
	if release == nil {
		return nil, nil
	}
	if timestamp := release.GetUpdatedAt(); timestamp != nil {
		if err := timestamp.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid updated_at: %w", err)
		}
	}
	additionInfo := map[string]any(nil)
	if release.GetAdditionInfo() != nil {
		additionInfo = release.GetAdditionInfo().AsMap()
	}
	return &types.Release{
		Name:         release.GetName(),
		Generation:   types.Generation(release.GetGeneration()),
		Type:         types.ReleaseType(release.GetReleaseType()),
		Platform:     ConvertPlatformToTypes(release.GetPlatform()),
		Version:      release.GetVersion(),
		Labels:       release.GetLabels(),
		FileName:     release.GetFileName(),
		MD5:          release.GetMd5(),
		Enabled:      release.GetEnabled(),
		IsHidden:     release.GetIsHidden(),
		IsShared:     release.GetIsShared(),
		IsSynced:     release.GetIsSynced(),
		AsDefault:    release.GetAsDefault(),
		UpdatedAt:    release.GetUpdatedAt().AsTime(),
		Operator:     release.GetOperator(),
		AdditionInfo: additionInfo,
	}, nil
}

// -----------------------------------------------------------------------------
// ReleaseAgent Related Interface
// -----------------------------------------------------------------------------

// Validate validates the request body.
func (x *ReleaseAgentEnableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseAgentEnableReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to an agent release key.
func (x *ReleaseAgentEnableReq) GetIdentifier() types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts an agent release key to the request.
func (x *ReleaseAgentEnableReq) ConvertFromTypes(key types.ReleaseAgentKey) {
	x.Generation = int64(key.Generation)
	x.Platform = ConvertPlatformFromTypes(key.Platform)
	x.Version = key.Version
}

// Validate validates the request body.
func (x *ReleaseAgentDisableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseAgentDisableReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to an agent release key.
func (x *ReleaseAgentDisableReq) GetIdentifier() types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts an agent release key to the request.
func (x *ReleaseAgentDisableReq) ConvertFromTypes(key types.ReleaseAgentKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleaseAgentSetAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseAgentSetAsDefaultReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to an agent release key.
func (x *ReleaseAgentSetAsDefaultReq) GetIdentifier() types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts an agent release key to the request.
func (x *ReleaseAgentSetAsDefaultReq) ConvertFromTypes(key types.ReleaseAgentKey) {
	x.Generation = int64(key.Generation)
	x.Platform = ConvertPlatformFromTypes(key.Platform)
	x.Version = key.Version
}

// Validate validates the request body.
func (x *ReleaseAgentCancelAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseAgentCancelAsDefaultReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to an agent release key.
func (x *ReleaseAgentCancelAsDefaultReq) GetIdentifier() types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts an agent release key to the request.
func (x *ReleaseAgentCancelAsDefaultReq) ConvertFromTypes(key types.ReleaseAgentKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleaseAgentDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseAgentDeleteReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleaseAgentDeleteReq) GetIdentifier() types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleaseAgentDeleteReq) ConvertFromTypes(key types.ReleaseAgentKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleaseAgentSetLabelsManyReq) Validate() error {
	if len(x.GetConditions()) != 1 || x.GetConditions()[0] == nil {
		return fmt.Errorf("release labels require exactly one condition")
	}
	exact := x.GetConditions()[0].GetExactInclude()
	if len(exact.GetGeneration()) != 1 {
		return fmt.Errorf("release labels require exactly one generation")
	}
	if len(exact.GetIsShared()) != 0 || len(exact.GetIsSynced()) != 0 {
		return fmt.Errorf("release labels do not support shared or synced conditions")
	}
	return types.Generation(exact.GetGeneration()[0]).Validate()
}

// AutoConvert auto converts the request.
func (x *ReleaseAgentSetLabelsManyReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// ConvertToTypes converts the request to domain labels and conditions.
func (x *ReleaseAgentSetLabelsManyReq) ConvertToTypes() ([]string, []*types.ReleaseCondition) {
	return x.GetLabels(), convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts labels and the supported exact condition to the request.
func (x *ReleaseAgentSetLabelsManyReq) ConvertFromTypes(labels []string, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	if err != nil {
		return err
	}
	x.Conditions = converted
	x.Labels = labels
	return x.Validate()
}

func convertReleaseAgentToTypes(item *ReleaseInfo) (*types.ReleaseAgent, error) {
	if item == nil {
		return nil, nil
	}
	release, err := convertReleaseToTypes(item)
	if err != nil {
		return nil, err
	}
	result := &types.ReleaseAgent{Release: *release}
	if release.AdditionInfo == nil {
		return result, nil
	}
	if err := conv.MapToStruct(release.AdditionInfo, &result.ReleaseAdditionInfoAgent); err != nil {
		return nil, fmt.Errorf("convert agent addition info: %w", err)
	}
	return result, nil
}

// Validate validates the agent list request.
func (x *ReleaseAgentListReq) Validate() error { return x.GetPage().Validate() }

// AutoConvert applies request defaults.
func (x *ReleaseAgentListReq) AutoConvert() {}

// ConvertPageToTypes converts the request page.
func (x *ReleaseAgentListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseAgentListReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseAgentListReq) ConvertFromTypes(page types.Page, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	if err != nil {
		return err
	}
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.Conditions = converted
	return nil
}

// Validate validates the agent count request.
func (x *ReleaseAgentCountReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleaseAgentCountReq) AutoConvert() {}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseAgentCountReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseAgentCountReq) ConvertFromTypes(conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Conditions = converted
	return err
}

// Validate validates the agent get request.
func (x *ReleaseAgentGetReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert applies request defaults.
func (x *ReleaseAgentGetReq) AutoConvert() {}

// GetIdentifier converts the request to an agent release key.
func (x *ReleaseAgentGetReq) GetIdentifier() types.ReleaseAgentKey {
	return types.ReleaseAgentKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts an agent release key to the request.
func (x *ReleaseAgentGetReq) ConvertFromTypes(key types.ReleaseAgentKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the agent distinct request.
func (x *ReleaseAgentDistinctReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleaseAgentDistinctReq) AutoConvert() {}

// ConvertDistinctFieldToTypes converts selected distinct fields.
func (x *ReleaseAgentDistinctReq) ConvertDistinctFieldToTypes() types.ReleaseDistinctField {
	fields := x.GetFields()
	return types.ReleaseDistinctField{
		OSType: fields.GetOsType(), CPUArch: fields.GetCpuArch(), Name: fields.GetName(), Version: fields.GetVersion(),
	}
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseAgentDistinctReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseAgentDistinctReq) ConvertFromTypes(fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Fields = &ReleaseDistinctFields{
		OsType: fields.OSType, CpuArch: fields.CPUArch, Name: fields.Name, Version: fields.Version,
	}
	x.Conditions = converted
	return err
}

// ConvertFromTypes converts agent releases to the response.
func (x *ReleaseAgentListResp) ConvertFromTypes(total int64, items []*types.ReleaseAgent) error {
	data := &ReleaseAgentListResp_Data{Total: total, Items: make([]*ReleaseInfo, len(items))}
	for index, item := range items {
		if item == nil {
			continue
		}
		converted, err := convertReleaseFromTypes(&item.Release)
		if err != nil {
			return err
		}
		data.Items[index] = converted
	}
	x.Data = data
	return nil
}

// ConvertToTypes converts the response data to agent releases.
func (x *ReleaseAgentListResp_Data) ConvertToTypes() (int64, []*types.ReleaseAgent, error) {
	if x == nil {
		return 0, nil, nil
	}
	items := make([]*types.ReleaseAgent, len(x.GetItems()))
	for index, item := range x.GetItems() {
		converted, err := convertReleaseAgentToTypes(item)
		if err != nil {
			return 0, nil, err
		}
		items[index] = converted
	}
	return x.GetTotal(), items, nil
}

// ConvertFromTypes converts an agent count to the response.
func (x *ReleaseAgentCountResp) ConvertFromTypes(count int64) {
	x.Data = &ReleaseAgentCountResp_Data{Count: count}
}

// ConvertToTypes converts the response data to an agent count.
func (x *ReleaseAgentCountResp_Data) ConvertToTypes() int64 {
	if x == nil {
		return 0
	}
	return x.GetCount()
}

// ConvertFromTypes converts an agent release to the response.
func (x *ReleaseAgentGetResp) ConvertFromTypes(item *types.ReleaseAgent) error {
	if item == nil {
		x.Data = &ReleaseAgentGetResp_Data{}
		return nil
	}
	converted, err := convertReleaseFromTypes(&item.Release)
	if err != nil {
		return err
	}
	x.Data = &ReleaseAgentGetResp_Data{Item: converted}
	return nil
}

// ConvertToTypes converts the response data to an agent release.
func (x *ReleaseAgentGetResp_Data) ConvertToTypes() (*types.ReleaseAgent, error) {
	if x == nil {
		return nil, nil
	}
	return convertReleaseAgentToTypes(x.GetItem())
}

// ConvertFromTypes converts a distinct result to the response.
func (x *ReleaseAgentDistinctResp) ConvertFromTypes(result *types.ReleaseDistinctResult) {
	data := &ReleaseDistinctResult{}
	if result != nil {
		data.OsType = result.OSType
		data.CpuArch = result.CPUArch
		data.Name = result.Name
		data.Version = result.Version
	}
	x.Data = &ReleaseAgentDistinctResp_Data{Result: data}
}

// ConvertToTypes converts the response data to a distinct result.
func (x *ReleaseAgentDistinctResp_Data) ConvertToTypes() *types.ReleaseDistinctResult {
	if x == nil {
		return nil
	}
	result := x.GetResult()
	if result == nil {
		return nil
	}
	return &types.ReleaseDistinctResult{
		OSType: result.GetOsType(), CPUArch: result.GetCpuArch(), Name: result.GetName(), Version: result.GetVersion(),
	}
}

// -----------------------------------------------------------------------------
// ReleaseProxy Related Interface
// -----------------------------------------------------------------------------

// Validate validates the request body.
func (x *ReleaseProxyEnableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseProxyEnableReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a proxy release key.
func (x *ReleaseProxyEnableReq) GetIdentifier() types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a proxy release key to the request.
func (x *ReleaseProxyEnableReq) ConvertFromTypes(key types.ReleaseProxyKey) {
	x.Generation = int64(key.Generation)
	x.Platform = ConvertPlatformFromTypes(key.Platform)
	x.Version = key.Version
}

// Validate validates the request body.
func (x *ReleaseProxyDisableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseProxyDisableReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a proxy release key.
func (x *ReleaseProxyDisableReq) GetIdentifier() types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a proxy release key to the request.
func (x *ReleaseProxyDisableReq) ConvertFromTypes(key types.ReleaseProxyKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleaseProxySetAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseProxySetAsDefaultReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a proxy release key.
func (x *ReleaseProxySetAsDefaultReq) GetIdentifier() types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a proxy release key to the request.
func (x *ReleaseProxySetAsDefaultReq) ConvertFromTypes(key types.ReleaseProxyKey) {
	x.Generation = int64(key.Generation)
	x.Platform = ConvertPlatformFromTypes(key.Platform)
	x.Version = key.Version
}

// Validate validates the request body.
func (x *ReleaseProxyCancelAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseProxyCancelAsDefaultReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a proxy release key.
func (x *ReleaseProxyCancelAsDefaultReq) GetIdentifier() types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a proxy release key to the request.
func (x *ReleaseProxyCancelAsDefaultReq) ConvertFromTypes(key types.ReleaseProxyKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleaseProxyDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleaseProxyDeleteReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleaseProxyDeleteReq) GetIdentifier() types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleaseProxyDeleteReq) ConvertFromTypes(key types.ReleaseProxyKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleaseProxySetLabelsManyReq) Validate() error {
	if len(x.GetConditions()) != 1 || x.GetConditions()[0] == nil {
		return fmt.Errorf("release labels require exactly one condition")
	}
	exact := x.GetConditions()[0].GetExactInclude()
	if len(exact.GetGeneration()) != 1 {
		return fmt.Errorf("release labels require exactly one generation")
	}
	if len(exact.GetIsShared()) != 0 || len(exact.GetIsSynced()) != 0 {
		return fmt.Errorf("release labels do not support shared or synced conditions")
	}
	return types.Generation(exact.GetGeneration()[0]).Validate()
}

// AutoConvert auto converts the request.
func (x *ReleaseProxySetLabelsManyReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// ConvertToTypes converts the request to domain labels and conditions.
func (x *ReleaseProxySetLabelsManyReq) ConvertToTypes() ([]string, []*types.ReleaseCondition) {
	return x.GetLabels(), convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts labels and the supported exact condition to the request.
func (x *ReleaseProxySetLabelsManyReq) ConvertFromTypes(labels []string, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	if err != nil {
		return err
	}
	x.Conditions = converted
	x.Labels = labels
	return x.Validate()
}

func convertReleaseProxyToTypes(item *ReleaseInfo) (*types.ReleaseProxy, error) {
	if item == nil {
		return nil, nil
	}
	release, err := convertReleaseToTypes(item)
	if err != nil {
		return nil, err
	}
	result := &types.ReleaseProxy{Release: *release}
	if release.AdditionInfo == nil {
		return result, nil
	}
	if err := conv.MapToStruct(release.AdditionInfo, &result.ReleaseAdditionInfoProxy); err != nil {
		return nil, fmt.Errorf("convert proxy addition info: %w", err)
	}
	return result, nil
}

// Validate validates the proxy list request.
func (x *ReleaseProxyListReq) Validate() error { return x.GetPage().Validate() }

// AutoConvert applies request defaults.
func (x *ReleaseProxyListReq) AutoConvert() {}

// ConvertPageToTypes converts the request page.
func (x *ReleaseProxyListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseProxyListReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseProxyListReq) ConvertFromTypes(page types.Page, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.Conditions = converted
	return err
}

// Validate validates the proxy count request.
func (x *ReleaseProxyCountReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleaseProxyCountReq) AutoConvert() {}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseProxyCountReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseProxyCountReq) ConvertFromTypes(conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Conditions = converted
	return err
}

// Validate validates the proxy get request.
func (x *ReleaseProxyGetReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert applies request defaults.
func (x *ReleaseProxyGetReq) AutoConvert() {}

// GetIdentifier converts the request to a proxy release key.
func (x *ReleaseProxyGetReq) GetIdentifier() types.ReleaseProxyKey {
	return types.ReleaseProxyKey{
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a proxy release key to the request.
func (x *ReleaseProxyGetReq) ConvertFromTypes(key types.ReleaseProxyKey) {
	x.Generation, x.Platform, x.Version = int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the proxy distinct request.
func (x *ReleaseProxyDistinctReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleaseProxyDistinctReq) AutoConvert() {}

// ConvertDistinctFieldToTypes converts selected distinct fields.
func (x *ReleaseProxyDistinctReq) ConvertDistinctFieldToTypes() types.ReleaseDistinctField {
	fields := x.GetFields()
	return types.ReleaseDistinctField{
		OSType: fields.GetOsType(), CPUArch: fields.GetCpuArch(), Name: fields.GetName(), Version: fields.GetVersion(),
	}
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseProxyDistinctReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseProxyDistinctReq) ConvertFromTypes(fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Fields = &ReleaseDistinctFields{
		OsType: fields.OSType, CpuArch: fields.CPUArch, Name: fields.Name, Version: fields.Version,
	}
	x.Conditions = converted
	return err
}

// ConvertFromTypes converts proxy releases to the response.
func (x *ReleaseProxyListResp) ConvertFromTypes(total int64, items []*types.ReleaseProxy) error {
	data := &ReleaseProxyListResp_Data{Total: total, Items: make([]*ReleaseInfo, len(items))}
	for index, item := range items {
		if item == nil {
			continue
		}
		converted, err := convertReleaseFromTypes(&item.Release)
		if err != nil {
			return err
		}
		data.Items[index] = converted
	}
	x.Data = data
	return nil
}

// ConvertToTypes converts the response data to proxy releases.
func (x *ReleaseProxyListResp_Data) ConvertToTypes() (int64, []*types.ReleaseProxy, error) {
	if x == nil {
		return 0, nil, nil
	}
	items := make([]*types.ReleaseProxy, len(x.GetItems()))
	for index, item := range x.GetItems() {
		converted, err := convertReleaseProxyToTypes(item)
		if err != nil {
			return 0, nil, err
		}
		items[index] = converted
	}
	return x.GetTotal(), items, nil
}

// ConvertFromTypes converts a proxy count to the response.
func (x *ReleaseProxyCountResp) ConvertFromTypes(count int64) {
	x.Data = &ReleaseProxyCountResp_Data{Count: count}
}

// ConvertToTypes converts the response data to a proxy count.
func (x *ReleaseProxyCountResp_Data) ConvertToTypes() int64 {
	if x == nil {
		return 0
	}
	return x.GetCount()
}

// ConvertFromTypes converts a proxy release to the response.
func (x *ReleaseProxyGetResp) ConvertFromTypes(item *types.ReleaseProxy) error {
	if item == nil {
		x.Data = &ReleaseProxyGetResp_Data{}
		return nil
	}
	converted, err := convertReleaseFromTypes(&item.Release)
	if err != nil {
		return err
	}
	x.Data = &ReleaseProxyGetResp_Data{Item: converted}
	return nil
}

// ConvertToTypes converts the response data to a proxy release.
func (x *ReleaseProxyGetResp_Data) ConvertToTypes() (*types.ReleaseProxy, error) {
	if x == nil {
		return nil, nil
	}
	return convertReleaseProxyToTypes(x.GetItem())
}

// ConvertFromTypes converts a distinct result to the response.
func (x *ReleaseProxyDistinctResp) ConvertFromTypes(result *types.ReleaseDistinctResult) {
	data := &ReleaseDistinctResult{}
	if result != nil {
		data.OsType = result.OSType
		data.CpuArch = result.CPUArch
		data.Name = result.Name
		data.Version = result.Version
	}
	x.Data = &ReleaseProxyDistinctResp_Data{Result: data}
}

// ConvertToTypes converts the response data to a distinct result.
func (x *ReleaseProxyDistinctResp_Data) ConvertToTypes() *types.ReleaseDistinctResult {
	if x == nil {
		return nil
	}
	result := x.GetResult()
	if result == nil {
		return nil
	}
	return &types.ReleaseDistinctResult{
		OSType: result.GetOsType(), CPUArch: result.GetCpuArch(), Name: result.GetName(), Version: result.GetVersion(),
	}
}

// -----------------------------------------------------------------------------
// ReleasePlugin Related Interface
// -----------------------------------------------------------------------------

// Validate validates the request body.
func (x *ReleasePluginEnableReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginEnableReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginEnableReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginEnableReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name = key.Name
	x.Generation = int64(key.Generation)
	x.Platform = ConvertPlatformFromTypes(key.Platform)
	x.Version = key.Version
}

// Validate validates the request body.
func (x *ReleasePluginDisableReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginDisableReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginDisableReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginDisableReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name, x.Generation, x.Platform, x.Version = key.Name, int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleasePluginSetAsDefaultReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginSetAsDefaultReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginSetAsDefaultReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginSetAsDefaultReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name = key.Name
	x.Generation = int64(key.Generation)
	x.Platform = ConvertPlatformFromTypes(key.Platform)
	x.Version = key.Version
}

// Validate validates the request body.
func (x *ReleasePluginCancelAsDefaultReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginCancelAsDefaultReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginCancelAsDefaultReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginCancelAsDefaultReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name, x.Generation, x.Platform, x.Version = key.Name, int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleasePluginDeleteReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginDeleteReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleasePluginDeleteReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleasePluginDeleteReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name, x.Generation, x.Platform, x.Version = key.Name, int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleasePluginSetHiddenReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginSetHiddenReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleasePluginSetHiddenReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleasePluginSetHiddenReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name, x.Generation, x.Platform, x.Version = key.Name, int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the request body.
func (x *ReleasePluginCancelHiddenReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto converts the request.
func (x *ReleasePluginCancelHiddenReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginCancelHiddenReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Name:       x.GetName(),
		Generation: types.Generation(x.GetGeneration()),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginCancelHiddenReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Name, x.Generation, x.Platform, x.Version = key.Name, int64(key.Generation), ConvertPlatformFromTypes(key.Platform), key.Version
}

func convertReleasePluginToTypes(item *ReleaseInfo) (*types.ReleasePlugin, error) {
	if item == nil {
		return nil, nil
	}
	release, err := convertReleaseToTypes(item)
	if err != nil {
		return nil, err
	}
	result := &types.ReleasePlugin{Release: *release}
	if release.AdditionInfo == nil {
		return result, nil
	}
	if err := conv.MapToStruct(release.AdditionInfo, &result.ReleaseAdditionInfoPlugin); err != nil {
		return nil, fmt.Errorf("convert plugin addition info: %w", err)
	}
	return result, nil
}

// Validate validates the plugin list request.
func (x *ReleasePluginListReq) Validate() error { return x.GetPage().Validate() }

// AutoConvert applies request defaults.
func (x *ReleasePluginListReq) AutoConvert() {}

// ConvertPageToTypes converts the request page.
func (x *ReleasePluginListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleasePluginListReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleasePluginListReq) ConvertFromTypes(page types.Page, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.Conditions = converted
	return err
}

// Validate validates the plugin count request.
func (x *ReleasePluginCountReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleasePluginCountReq) AutoConvert() {}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleasePluginCountReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleasePluginCountReq) ConvertFromTypes(conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Conditions = converted
	return err
}

// Validate validates the plugin get request.
func (x *ReleasePluginGetReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert applies request defaults.
func (x *ReleasePluginGetReq) AutoConvert() {}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginGetReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Generation: types.Generation(x.GetGeneration()),
		Name:       x.GetName(),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginGetReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Generation, x.Name, x.Platform, x.Version = int64(key.Generation), key.Name, ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the plugin distinct request.
func (x *ReleasePluginDistinctReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleasePluginDistinctReq) AutoConvert() {}

// ConvertDistinctFieldToTypes converts selected distinct fields.
func (x *ReleasePluginDistinctReq) ConvertDistinctFieldToTypes() types.ReleaseDistinctField {
	fields := x.GetFields()
	return types.ReleaseDistinctField{
		OSType: fields.GetOsType(), CPUArch: fields.GetCpuArch(), Name: fields.GetName(), Version: fields.GetVersion(),
	}
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleasePluginDistinctReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleasePluginDistinctReq) ConvertFromTypes(fields types.ReleaseDistinctField, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Fields = &ReleaseDistinctFields{
		OsType: fields.OSType, CpuArch: fields.CPUArch, Name: fields.Name, Version: fields.Version,
	}
	x.Conditions = converted
	return err
}

// Validate validates the plugin existence request.
func (x *ReleasePluginExistReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert applies request defaults.
func (x *ReleasePluginExistReq) AutoConvert() {}

// GetIdentifier converts the request to a plugin release key.
func (x *ReleasePluginExistReq) GetIdentifier() types.ReleasePluginKey {
	return types.ReleasePluginKey{
		Generation: types.Generation(x.GetGeneration()),
		Name:       x.GetName(),
		Platform:   ConvertPlatformToTypes(x.GetPlatform()),
		Version:    x.GetVersion(),
	}
}

// ConvertConditionsToTypes converts the existence key to an exact condition.
func (x *ReleasePluginExistReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return []*types.ReleaseCondition{
		{
			ExactInclude: &types.ReleaseExactFields{
				Name:       []string{x.GetName()},
				Generation: []types.Generation{types.Generation(x.GetGeneration())},
				Platform:   []platform.Platform{ConvertPlatformToTypes(x.GetPlatform())},
				Version:    []string{x.GetVersion()},
			},
		},
	}
}

// ConvertFromTypes converts a plugin release key to the request.
func (x *ReleasePluginExistReq) ConvertFromTypes(key types.ReleasePluginKey) {
	x.Generation, x.Name, x.Platform, x.Version = int64(key.Generation), key.Name, ConvertPlatformFromTypes(key.Platform), key.Version
}

// Validate validates the default plugin version request.
func (x *ReleasePluginGetDefaultVersionReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, platform(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert applies request defaults.
func (x *ReleasePluginGetDefaultVersionReq) AutoConvert() {}

// GetIdentifier returns the default-version query identifier.
func (x *ReleasePluginGetDefaultVersionReq) GetIdentifier() (string, types.Generation, platform.Platform) {
	return x.GetName(), types.Generation(x.GetGeneration()), ConvertPlatformToTypes(x.GetPlatform())
}

// ConvertFromTypes converts the default-version query identifier to the request.
func (x *ReleasePluginGetDefaultVersionReq) ConvertFromTypes(name string, generation types.Generation, platform platform.Platform) {
	x.Name, x.Generation, x.Platform = name, int64(generation), ConvertPlatformFromTypes(platform)
}

// Validate validates the plugin distinct-name request.
func (x *ReleasePluginDistinctNameReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleasePluginDistinctNameReq) AutoConvert() {}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleasePluginDistinctNameReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleasePluginDistinctNameReq) ConvertFromTypes(conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Conditions = converted
	return err
}

// ConvertFromTypes converts plugin releases to the response.
func (x *ReleasePluginListResp) ConvertFromTypes(total int64, items []*types.ReleasePlugin) error {
	data := &ReleasePluginListResp_Data{Total: total, Items: make([]*ReleaseInfo, len(items))}
	for index, item := range items {
		if item == nil {
			continue
		}
		converted, err := convertReleaseFromTypes(&item.Release)
		if err != nil {
			return err
		}
		data.Items[index] = converted
	}
	x.Data = data
	return nil
}

// ConvertToTypes converts the response data to plugin releases.
func (x *ReleasePluginListResp_Data) ConvertToTypes() (int64, []*types.ReleasePlugin, error) {
	if x == nil {
		return 0, nil, nil
	}
	items := make([]*types.ReleasePlugin, len(x.GetItems()))
	for index, item := range x.GetItems() {
		converted, err := convertReleasePluginToTypes(item)
		if err != nil {
			return 0, nil, err
		}
		items[index] = converted
	}
	return x.GetTotal(), items, nil
}

// ConvertFromTypes converts a plugin count to the response.
func (x *ReleasePluginCountResp) ConvertFromTypes(count int64) {
	x.Data = &ReleasePluginCountResp_Data{Count: count}
}

// ConvertToTypes converts the response data to a plugin count.
func (x *ReleasePluginCountResp_Data) ConvertToTypes() int64 {
	if x == nil {
		return 0
	}
	return x.GetCount()
}

// ConvertFromTypes converts a plugin release to the response.
func (x *ReleasePluginGetResp) ConvertFromTypes(item *types.ReleasePlugin) error {
	if item == nil {
		x.Data = &ReleasePluginGetResp_Data{}
		return nil
	}
	converted, err := convertReleaseFromTypes(&item.Release)
	if err != nil {
		return err
	}
	x.Data = &ReleasePluginGetResp_Data{Item: converted}
	return nil
}

// ConvertToTypes converts the response data to a plugin release.
func (x *ReleasePluginGetResp_Data) ConvertToTypes() (*types.ReleasePlugin, error) {
	if x == nil {
		return nil, nil
	}
	return convertReleasePluginToTypes(x.GetItem())
}

// ConvertFromTypes converts a distinct result to the response.
func (x *ReleasePluginDistinctResp) ConvertFromTypes(result *types.ReleaseDistinctResult) {
	data := &ReleaseDistinctResult{}
	if result != nil {
		data.OsType = result.OSType
		data.CpuArch = result.CPUArch
		data.Name = result.Name
		data.Version = result.Version
	}
	x.Data = &ReleasePluginDistinctResp_Data{Result: data}
}

// ConvertToTypes converts the response data to a distinct result.
func (x *ReleasePluginDistinctResp_Data) ConvertToTypes() *types.ReleaseDistinctResult {
	if x == nil {
		return nil
	}
	result := x.GetResult()
	if result == nil {
		return nil
	}
	return &types.ReleaseDistinctResult{
		OSType: result.GetOsType(), CPUArch: result.GetCpuArch(), Name: result.GetName(), Version: result.GetVersion(),
	}
}

// ConvertFromTypes converts a plugin existence result to the response.
func (x *ReleasePluginExistResp) ConvertFromTypes(exists bool) {
	x.Data = &ReleasePluginExistResp_Data{Exists: exists}
}

// ConvertToTypes converts the response data to a plugin existence result.
func (x *ReleasePluginExistResp_Data) ConvertToTypes() bool { return x != nil && x.GetExists() }

// ConvertFromTypes converts a default plugin version to the response.
func (x *ReleasePluginGetDefaultVersionResp) ConvertFromTypes(version string) {
	x.Data = &ReleasePluginGetDefaultVersionResp_Data{Version: version}
}

// ConvertToTypes converts the response data to a default plugin version.
func (x *ReleasePluginGetDefaultVersionResp_Data) ConvertToTypes() string {
	if x == nil {
		return ""
	}
	return x.GetVersion()
}

// ConvertFromTypes converts plugin names to the response.
func (x *ReleasePluginDistinctNameResp) ConvertFromTypes(names []string) {
	x.Data = &ReleasePluginDistinctNameResp_Data{Names: names}
}

// ConvertToTypes converts the response data to plugin names.
func (x *ReleasePluginDistinctNameResp_Data) ConvertToTypes() []string {
	if x == nil {
		return nil
	}
	return x.GetNames()
}

// -----------------------------------------------------------------------------
// ReleaseCert Related Interface
// -----------------------------------------------------------------------------

// Validate validates the request body.
func (x *ReleaseCertDeleteReq) Validate() error {
	return types.Generation(x.GetGeneration()).Validate()
}

// AutoConvert auto converts the request.
func (x *ReleaseCertDeleteReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleaseCertDeleteReq) GetIdentifier() types.ReleaseCertKey {
	return types.ReleaseCertKey{
		Generation: types.Generation(x.GetGeneration()),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleaseCertDeleteReq) ConvertFromTypes(key types.ReleaseCertKey) {
	x.Generation = int64(key.Generation)
}

// Validate validates the cert list request.
func (x *ReleaseCertListReq) Validate() error { return x.GetPage().Validate() }

// AutoConvert applies request defaults.
func (x *ReleaseCertListReq) AutoConvert() {}

// ConvertPageToTypes converts the request page.
func (x *ReleaseCertListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseCertListReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseCertListReq) ConvertFromTypes(page types.Page, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.Conditions = converted
	return err
}

// ConvertFromTypes converts cert releases to the response.
func (x *ReleaseCertListResp) ConvertFromTypes(total int64, items []*types.ReleaseCert) error {
	data := &ReleaseCertListResp_Data{Total: total, Items: make([]*ReleaseInfo, len(items))}
	for index, item := range items {
		release, err := convertReleaseFromTypes(&item.Release)
		if err != nil {
			return err
		}
		data.Items[index] = release
	}
	x.Data = data
	return nil
}

// ConvertToTypes converts the response data to cert releases.
func (x *ReleaseCertListResp_Data) ConvertToTypes() (int64, []*types.ReleaseCert, error) {
	if x == nil {
		return 0, nil, nil
	}
	items := make([]*types.ReleaseCert, len(x.GetItems()))
	for index, item := range x.GetItems() {
		release, err := convertReleaseToTypes(item)
		if err != nil {
			return 0, nil, err
		}
		items[index] = &types.ReleaseCert{Release: *release}
	}
	return x.GetTotal(), items, nil
}

// -----------------------------------------------------------------------------
// ReleaseBinTool Related Interface
// -----------------------------------------------------------------------------

// Validate validates the request body.
func (x *ReleaseBinToolDeleteReq) Validate() error {
	return types.Generation(x.GetGeneration()).Validate()
}

// AutoConvert auto converts the request.
func (x *ReleaseBinToolDeleteReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleaseBinToolDeleteReq) GetIdentifier() types.ReleaseBinToolKey {
	return types.ReleaseBinToolKey{
		Generation: types.Generation(x.GetGeneration()),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleaseBinToolDeleteReq) ConvertFromTypes(key types.ReleaseBinToolKey) {
	x.Generation = int64(key.Generation)
}

// Validate validates the bin tool list request.
func (x *ReleaseBinToolListReq) Validate() error { return x.GetPage().Validate() }

// AutoConvert applies request defaults.
func (x *ReleaseBinToolListReq) AutoConvert() {}

// ConvertPageToTypes converts the request page.
func (x *ReleaseBinToolListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleaseBinToolListReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleaseBinToolListReq) ConvertFromTypes(page types.Page, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.Conditions = converted
	return err
}

// ConvertFromTypes converts bin tool releases to the response.
func (x *ReleaseBinToolListResp) ConvertFromTypes(total int64, items []*types.ReleaseBinTool) error {
	data := &ReleaseBinToolListResp_Data{Total: total, Items: make([]*ReleaseInfo, len(items))}
	for index, item := range items {
		release, err := convertReleaseFromTypes(&item.Release)
		if err != nil {
			return err
		}
		data.Items[index] = release
	}
	x.Data = data
	return nil
}

// ConvertToTypes converts the response data to bin tool releases.
func (x *ReleaseBinToolListResp_Data) ConvertToTypes() (int64, []*types.ReleaseBinTool, error) {
	if x == nil {
		return 0, nil, nil
	}
	items := make([]*types.ReleaseBinTool, len(x.GetItems()))
	for index, item := range x.GetItems() {
		release, err := convertReleaseToTypes(item)
		if err != nil {
			return 0, nil, err
		}
		items[index] = &types.ReleaseBinTool{Release: *release}
	}
	return x.GetTotal(), items, nil
}

// -----------------------------------------------------------------------------
// ReleasePluginBinTool Related Interface
// -----------------------------------------------------------------------------

// Validate validates the request body.
func (x *ReleasePluginBinToolDeleteReq) Validate() error {
	return types.Generation(x.GetGeneration()).Validate()
}

// AutoConvert auto converts the request.
func (x *ReleasePluginBinToolDeleteReq) AutoConvert() {
	// Intentionally empty: this request has no semantic defaults.
}

// GetIdentifier converts the request to a release key.
func (x *ReleasePluginBinToolDeleteReq) GetIdentifier() types.ReleasePluginBinToolKey {
	return types.ReleasePluginBinToolKey{
		Generation: types.Generation(x.GetGeneration()),
		Name:       x.GetName(),
	}
}

// ConvertFromTypes converts a release key to the request.
func (x *ReleasePluginBinToolDeleteReq) ConvertFromTypes(key types.ReleasePluginBinToolKey) {
	x.Generation, x.Name = int64(key.Generation), key.Name
}

// Validate validates the plugin bin tool list request.
func (x *ReleasePluginBinToolListReq) Validate() error { return x.GetPage().Validate() }

// AutoConvert applies request defaults.
func (x *ReleasePluginBinToolListReq) AutoConvert() {}

// ConvertPageToTypes converts the request page.
func (x *ReleasePluginBinToolListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleasePluginBinToolListReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleasePluginBinToolListReq) ConvertFromTypes(page types.Page, conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.Conditions = converted
	return err
}

// Validate validates the plugin bin tool distinct-name request.
func (x *ReleasePluginBinToolDistinctNameReq) Validate() error { return nil }

// AutoConvert applies request defaults.
func (x *ReleasePluginBinToolDistinctNameReq) AutoConvert() {}

// ConvertConditionsToTypes converts request conditions in order.
func (x *ReleasePluginBinToolDistinctNameReq) ConvertConditionsToTypes() []*types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetConditions())
}

// ConvertFromTypes converts query arguments to the request.
func (x *ReleasePluginBinToolDistinctNameReq) ConvertFromTypes(conditions ...*types.ReleaseCondition) error {
	converted, err := convertReleaseConditionsFromTypes(conditions...)
	x.Conditions = converted
	return err
}

// ConvertFromTypes converts plugin bin tool releases to the response.
func (x *ReleasePluginBinToolListResp) ConvertFromTypes(total int64, items []*types.ReleasePluginBinTool) error {
	data := &ReleasePluginBinToolListResp_Data{Total: total, Items: make([]*ReleaseInfo, len(items))}
	for index, item := range items {
		release, err := convertReleaseFromTypes(&item.Release)
		if err != nil {
			return err
		}
		data.Items[index] = release
	}
	x.Data = data
	return nil
}

// ConvertToTypes converts the response data to plugin bin tool releases.
func (x *ReleasePluginBinToolListResp_Data) ConvertToTypes() (int64, []*types.ReleasePluginBinTool, error) {
	if x == nil {
		return 0, nil, nil
	}
	items := make([]*types.ReleasePluginBinTool, len(x.GetItems()))
	for index, item := range x.GetItems() {
		release, err := convertReleaseToTypes(item)
		if err != nil {
			return 0, nil, err
		}
		items[index] = &types.ReleasePluginBinTool{Release: *release}
	}
	return x.GetTotal(), items, nil
}

// ConvertFromTypes converts plugin bin tool names to the response.
func (x *ReleasePluginBinToolDistinctNameResp) ConvertFromTypes(names []string) {
	x.Data = &ReleasePluginBinToolDistinctNameResp_Data{Names: names}
}

// ConvertToTypes converts the response data to plugin bin tool names.
func (x *ReleasePluginBinToolDistinctNameResp_Data) ConvertToTypes() []string {
	if x == nil {
		return nil
	}
	return x.GetNames()
}
