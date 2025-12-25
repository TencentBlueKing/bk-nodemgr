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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"google.golang.org/protobuf/types/known/structpb"
)

// Validate check body.
func (x *DeployPolicyListReq) Validate() error {
	if err := validatePage(x.GetPage()); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *DeployPolicyListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *DeployPolicyListReq) ConvertPageToTypes(maxLimit int) (types.Page, error) {
	return convPageToTypes(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *DeployPolicyListReq) ConvertConditionsToTypes() *types.DeployPolicyCondition {
	exactIncludeCond := convertDeployPolicyExactConditionsToTypes(x.GetExactIncludeConditions())
	fuzzyIncludeCond := convertDeployPolicyFuzzyConditionsToTypes(x.GetFuzzyIncludeConditions())

	return &types.DeployPolicyCondition{
		ExactInclude: exactIncludeCond,
		FuzzyInclude: fuzzyIncludeCond,
	}
}

// convertDeployPolicyExactConditionsToTypes convert deploy policy exact conditions to types.
func convertDeployPolicyExactConditionsToTypes(exactCond *DeployPolicyExactConditions) *types.DeployPolicyExactFields {
	if exactCond == nil {
		return nil
	}

	return &types.DeployPolicyExactFields{
		DeployPolicyID:   exactCond.GetDeployPolicyId(),
		Enabled:          exactCond.GetEnabled(),
		DeployPolicyName: exactCond.GetDeployPolicyName(),
		Operator:         exactCond.GetOperator(),
	}
}

// convertDeployPolicyFuzzyConditionsToTypes convert deploy policy fuzzy conditions to types.
func convertDeployPolicyFuzzyConditionsToTypes(fuzzyCond *DeployPolicyFuzzyConditions) *types.DeployPolicyFuzzyFields {
	if fuzzyCond == nil {
		return nil
	}

	return &types.DeployPolicyFuzzyFields{
		DeployPolicyName: fuzzyCond.GetDeployPolicyName(),
		Operator:         fuzzyCond.GetOperator(),
	}
}

// ConvertDeployPoliciesFromTypes convert deploy policies from types.
func (x *DeployPolicyListResp) ConvertDeployPoliciesFromTypes(total int64, deployPolicies []*types.DeployPolicy) error {
	items, err := convDeployPoliciesFromTypes(deployPolicies)
	if err != nil {
		return err
	}

	x.Data = &DeployPolicyListResp_Data{
		Total: total,
		Items: items,
	}

	return nil
}

func convDeployPoliciesFromTypes(deployPolicies []*types.DeployPolicy) ([]*DeployPolicy, error) {
	return conv.SliceToSliceWithError[*types.DeployPolicy, *DeployPolicy](deployPolicies, convDeployPolicyFromTypes)
}

func convDeployPolicyFromTypes(deployPolicy *types.DeployPolicy) (*DeployPolicy, error) {
	if deployPolicy == nil {
		return nil, fmt.Errorf("deploy policy is nil")
	}

	meta := convDeployPolicyMetaFromTypes(deployPolicy.Meta)

	specs, err := convSpecsFromTypes(deployPolicy.Specs)
	if err != nil {
		return nil, err
	}

	scopes, err := convScopesFromTypes(deployPolicy.Scopes)
	if err != nil {
		return nil, err
	}

	return &DeployPolicy{
		DeployPolicyId: deployPolicy.DeployPolicyID,
		Meta:           &meta,
		Specs:          specs,
		Scopes:         scopes,
		Operator:       deployPolicy.Operator,
		Enabled:        deployPolicy.Enabled,
	}, nil
}

func convDeployPolicyMetaFromTypes(deployPolicyMeta types.DeployPolicyMeta) DeployPolicyMeta {
	return DeployPolicyMeta{
		Name:        deployPolicyMeta.Name,
		Description: deployPolicyMeta.Description,
	}
}

func convSpecsFromTypes(specs []*types.DeploySpec) ([]*DeploySpec, error) {
	return conv.SliceToSliceWithError[*types.DeploySpec, *DeploySpec](specs, convSpecFromTypes)
}

func convSpecFromTypes(spec *types.DeploySpec) (*DeploySpec, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is nil")
	}

	param, err := structpb.NewStruct(spec.Param)
	if err != nil {
		return nil, err
	}

	return &DeploySpec{
		Type:  string(spec.Type),
		Param: param,
	}, nil
}

func convScopesFromTypes(scopes []*types.Scope) ([]*Scope, error) {
	return conv.SliceToSliceWithError[*types.Scope, *Scope](scopes, convScopeFromTypes)
}

func convScopeFromTypes(scope *types.Scope) (*Scope, error) {
	items, err := conv.SliceToSliceWithError[map[string]any, *structpb.Struct](scope.Items, structpb.NewStruct)
	if err != nil {
		return nil, err
	}

	return &Scope{
		BkBizId:     scope.BizID,
		Type:        string(scope.Type),
		Granularity: string(scope.Granularity),
		Filter:      convTargetFilterFromTypes(scope.Filter),
		Items:       items,
	}, nil
}

func convTargetFilterFromTypes(targetFilter *types.TargetFilter) *TargetFilter {
	return &TargetFilter{}
}

// Validate check body.
func (x *DeployPolicyCreateReq) Validate() error {
	if x.GetName() == "" {
		return fmt.Errorf("name is required")
	}

	// we allow to create a deploy_policy with empty scopes and specs.

	return nil
}

// AutoConvert auto convert.
func (x *DeployPolicyCreateReq) AutoConvert() {
}

// ConvertDeployPolicyToTypes convert deploy policy to types.
func (x *DeployPolicyCreateReq) ConvertDeployPolicyToTypes() (*types.DeployPolicy, error) {
	meta := types.DeployPolicyMeta{
		Name:        x.GetName(),
		Description: x.GetDescription(),
	}

	scopes, err := convScopesToTypes(x.GetScopes())
	if err != nil {
		return nil, err
	}

	specs, err := convSpecsToTypes(x.GetSpecs())
	if err != nil {
		return nil, err
	}

	deployPolicy := &types.DeployPolicy{
		// this is a new deploy policy, so we don't set DeployPolicyID.
		DeployPolicyID: -1,
		Meta:           meta,
		Scopes:         scopes,
		Specs:          specs,
		Operator:       "",
		Enabled:        x.GetEnabled(),
	}

	return deployPolicy, nil
}

func convDeployPolicyToTypes(deployPolicy *DeployPolicy) (*types.DeployPolicy, error) {
	if deployPolicy == nil {
		return nil, fmt.Errorf("deploy policy is nil")
	}

	scopes, err := convScopesToTypes(deployPolicy.GetScopes())
	if err != nil {
		return nil, err
	}

	specs, err := convSpecsToTypes(deployPolicy.GetSpecs())
	if err != nil {
		return nil, err
	}

	return &types.DeployPolicy{
		DeployPolicyID: deployPolicy.GetDeployPolicyId(),
		Meta:           convDeployPolicyMetaToTypes(deployPolicy.GetMeta()),
		Scopes:         scopes,
		Specs:          specs,
		Operator:       deployPolicy.GetOperator(),
		Enabled:        deployPolicy.GetEnabled(),
	}, nil
}

func convSpecsToTypes(specs []*DeploySpec) ([]*types.DeploySpec, error) {
	return conv.SliceToSliceWithError[*DeploySpec, *types.DeploySpec](specs, convSpecToTypes)
}

func convScopesToTypes(scopes []*Scope) ([]*types.Scope, error) {
	return conv.SliceToSliceWithError[*Scope, *types.Scope](scopes, convScopeToTypes)
}

func convDeployPolicyMetaToTypes(deployPolicyMeta *DeployPolicyMeta) types.DeployPolicyMeta {
	if deployPolicyMeta == nil {
		return types.DeployPolicyMeta{}
	}

	return types.DeployPolicyMeta{
		Name:        deployPolicyMeta.Name,
		Description: deployPolicyMeta.Description,
	}
}

func convScopeToTypes(scope *Scope) (*types.Scope, error) {
	scopeType := types.ScopeType(scope.GetType())
	if err := scopeType.Validate(); err != nil {
		return nil, err
	}

	targetGranularity := types.TargetGranularity(scope.GetGranularity())
	if err := targetGranularity.Validate(); err != nil {
		return nil, err
	}

	result := &types.Scope{
		BizID:       scope.GetBkBizId(),
		Type:        scopeType,
		Granularity: targetGranularity,
		Filter:      convTargetFilterToTypes(scope.GetFilter()),
		Items: conv.SliceToSlice[*structpb.Struct, map[string]any](scope.GetItems(), func(s *structpb.Struct) map[string]any {
			return s.AsMap()
		}),
	}

	if err := result.Validate(); err != nil {
		return nil, err
	}

	return result, nil
}

func convTargetFilterToTypes(_ *TargetFilter) *types.TargetFilter {
	// TODO: implement me.
	return &types.TargetFilter{}
}

func convSpecToTypes(spec *DeploySpec) (*types.DeploySpec, error) {
	specType := types.DeploySpecType(spec.GetType())
	if err := specType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid spec type(%s): %w", spec.GetType(), err)
	}

	return &types.DeploySpec{
		Type:  specType,
		Param: spec.Param.AsMap(),
	}, nil
}

// ConvertDeployPolicyID convert deploy policy id.
func (x *DeployPolicyCreateResp) ConvertDeployPolicyID(deployPolicyID int64) {
	data := &DeployPolicyCreateResp_Data{DeployPolicyId: &deployPolicyID}
	x.Data = data
}

// Validate check body.
func (x *DeployPolicyUpdateReq) Validate() error {
	if len(x.GetDeployPolicies()) == 0 {
		return fmt.Errorf("deploy policies is required")
	}

	if x.GetFields() == nil {
		return fmt.Errorf("fields is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *DeployPolicyUpdateReq) AutoConvert() {
}

// ConvertFieldsToTypes convert fields to types.
func (x *DeployPolicyUpdateReq) ConvertFieldsToTypes() types.DeployPolicyFields {
	if x.GetFields() == nil {
		return types.DeployPolicyFields{}
	}

	return types.DeployPolicyFields{
		Meta:    x.GetFields().GetMeta(),
		Scopes:  x.GetFields().GetScopes(),
		Specs:   x.GetFields().GetSpecs(),
		Enabled: x.GetFields().GetEnabled(),
	}
}

// ConvertDeployPoliciesToTypes convert deploy policies to types.
func (x *DeployPolicyUpdateReq) ConvertDeployPoliciesToTypes() ([]*types.DeployPolicy, error) {
	return conv.SliceToSliceWithError[*DeployPolicy, *types.DeployPolicy](x.GetDeployPolicies(), convDeployPolicyToTypes)
}

// Validate check body.
func (x *DeployPolicyExecuteReq) Validate() error {
	if x.GetDeployPolicyId() <= 0 {
		return fmt.Errorf("deploy policy id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *DeployPolicyExecuteReq) AutoConvert() {
	if x.DeployPolicyId == nil {
		x.DeployPolicyId = new(int64)
		*x.DeployPolicyId = -1
	}
}
