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
	"google.golang.org/protobuf/encoding/protojson"
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

const (
	// deploy policy list max limit
	maxDeployPolicyLimit = 1000
)

// PageLimit return page limit.
func (x *DeployPolicyListReq) PageLimit() int {
	return maxDeployPolicyLimit
}

// ConvertPageToTypes convert page to types.
func (x *DeployPolicyListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *DeployPolicyListReq) ConvertConditionsToTypes() (*types.DeployPolicyCondition, error) {
	exactIncludeCond := convertDeployPolicyExactConditionsToTypes(x.GetExactIncludeConditions())
	fuzzyIncludeCond := convertDeployPolicyFuzzyConditionsToTypes(x.GetFuzzyIncludeConditions())
	exactExcludeCond := convertDeployPolicyExactConditionsToTypes(x.GetExactExcludeConditions())
	fuzzyExcludeCond := convertDeployPolicyFuzzyConditionsToTypes(x.GetFuzzyExcludeConditions())
	executedTimeRangeCond, err := convertTimeRangeToTypes(x.GetExecutedTimeRange())
	if err != nil {
		return nil, fmt.Errorf("failed to convert executed time range: %w", err)
	}

	return &types.DeployPolicyCondition{
		ExecutedTimeRange: executedTimeRangeCond,
		ExactInclude:      exactIncludeCond,
		FuzzyInclude:      fuzzyIncludeCond,
		ExactExclude:      exactExcludeCond,
		FuzzyExclude:      fuzzyExcludeCond,
	}, nil
}

// convertDeployPolicyExactConditionsToTypes convert deploy policy exact conditions to types.
func convertDeployPolicyExactConditionsToTypes(exactCond *DeployPolicyExactConditions) *types.DeployPolicyExactFields {
	if exactCond == nil {
		return nil
	}

	return &types.DeployPolicyExactFields{
		DeployPolicyID:   exactCond.GetDeployPolicyId(),
		DsuID:            exactCond.GetDsuId(),
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
		DsuId:          deployPolicy.DsuID,
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

	specType := spec.Type()
	result := &DeploySpec{
		Type: string(specType),
	}

	// Convert param to protobuf message based on type, then to structpb.Struct
	var paramProto interface{}

	switch specType {
	case types.DeploySpecTypeSpecifyAgent:
		param, err := spec.GetSpecifyAgentParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify agent param: %w", err)
		}
		paramProto = &SpecifyAgentParam{
			NodeVersion: param.NodeVersion,
		}

	case types.DeploySpecTypeSpecifyProxy:
		param, err := spec.GetSpecifyProxyParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify proxy param: %w", err)
		}
		paramProto = &SpecifyProxyParam{
			NodeVersion: param.NodeVersion,
		}

	case types.DeploySpecTypeSpecifyPlugin:
		param, err := spec.GetSpecifyPluginParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin param: %w", err)
		}
		customConfigContext, err := structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to convert custom config context: %w", err)
		}
		paramProto = &SpecifyPluginParam{
			PluginName:          param.PluginName,
			Version:             param.Version,
			CustomConfigContext: customConfigContext,
		}

	case types.DeploySpecTypeSpecifyPluginPkg:
		param, err := spec.GetSpecifyPluginPkgParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin pkg param: %w", err)
		}
		customConfigContext, err := structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to convert custom config context: %w", err)
		}
		paramProto = &SpecifyPluginPkgParam{
			PluginPkgName:       param.PluginPkgName,
			Version:             param.Version,
			CustomConfigContext: customConfigContext,
		}

	case types.DeploySpecTypeSpecifyPluginSubConfig:
		param, err := spec.GetSpecifyPluginSubConfigParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin sub config param: %w", err)
		}
		configFilesDetail := make([]*PluginConfigDetail, 0, len(param.ConfigFilesDetail))
		for _, detail := range param.ConfigFilesDetail {
			configFilesDetail = append(configFilesDetail, &PluginConfigDetail{
				Name:         detail.Name,
				Content:      detail.Content,
				IsMainConfig: detail.IsMainConfig,
			})
		}
		customConfigContext, err := structpb.NewStruct(param.CustomConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to convert custom config context: %w", err)
		}
		paramProto = &SpecifyPluginSubConfigParam{
			PluginName:          param.PluginName,
			ConfigFilesDetail:   configFilesDetail,
			CustomConfigContext: customConfigContext,
		}

	default:
		return nil, fmt.Errorf("unknown deploy spec type: %s", specType)
	}

	paramMap, err := conv.StructToMap(paramProto)
	if err != nil {
		return nil, fmt.Errorf("failed to convert param to map: %w", err)
	}
	result.Param, err = structpb.NewStruct(paramMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert param to struct: %w", err)
	}

	return result, nil
}

func convScopesFromTypes(scopes []*types.Scope) ([]*Scope, error) {
	return conv.SliceToSliceWithError[*types.Scope, *Scope](scopes, convScopeFromTypes)
}

func convScopeFromTypes(scope *types.Scope) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("scope is nil")
	}

	scopeType := scope.Type()
	result := &Scope{
		Type: string(scopeType),
	}

	// Convert param to protobuf message based on type, then to structpb.Struct
	var scopeProto interface{}

	switch scopeType {
	case types.ScopeTypeServiceTemplate:
		item, err := scope.GetServiceTemplateScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get service template scope: %w", err)
		}

		scopeProto = &ScopeServiceTemplate{
			Granularity:        string(item.Granularity),
			BkBizId:            item.BizID,
			Filter:             convTargetFilterFromTypes(item.Filter),
			ServiceTemplateIds: item.ServiceTemplateIDs,
			ModuleIds:          item.ModuleIDs,
		}

	case types.ScopeTypeSetTemplate:
		item, err := scope.GetSetTemplateScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get set template scope: %w", err)
		}

		scopeProto = &ScopeSetTemplate{
			Granularity:    string(item.Granularity),
			BkBizId:        item.BizID,
			Filter:         convTargetFilterFromTypes(item.Filter),
			SetTemplateIds: item.SetTemplateIDs,
			SetIds:         item.SetIDs,
		}

	case types.ScopeTypeInstance:
		item, err := scope.GetInstanceScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get instance scope: %w", err)
		}

		scopeProto = &ScopeInstance{
			Granularity: string(item.Granularity),
			BkBizId:     item.BizID,
			Filter:      convTargetFilterFromTypes(item.Filter),
			InstanceIds: item.InstanceIDs,
		}

	case types.ScopeTypeTopo:
		item, err := scope.GetTopoScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get topo scope: %w", err)
		}

		paths := make([]*ScopeTopo_ScopeTopoNode, 0, len(item.Paths))
		for _, p := range item.Paths {
			paths = append(paths, &ScopeTopo_ScopeTopoNode{
				TopoObjId:  p.TopoObjID,
				TopoInstId: p.TopoInstID,
			})
		}

		scopeProto = &ScopeTopo{
			Granularity: string(item.Granularity),
			BkBizId:     item.BizID,
			Filter:      convTargetFilterFromTypes(item.Filter),
			Paths:       paths,
		}

	case types.ScopeTypeDynamicGroup:
		item, err := scope.GetDynamicGroupScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get dynamic group scope: %w", err)
		}

		scopeProto = &ScopeDynamicGroup{
			Granularity:     string(item.Granularity),
			BkBizId:         item.BizID,
			Filter:          convTargetFilterFromTypes(item.Filter),
			DynamicGroupIds: item.DynamicGroupIDs,
		}

	default:
		return nil, fmt.Errorf("unknown scope type: %s", scopeType)
	}

	scopeMap, err := conv.StructToMap(scopeProto)
	if err != nil {
		return nil, fmt.Errorf("failed to convert param to map: %w", err)
	}
	result.Scope, err = structpb.NewStruct(scopeMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert param to struct: %w", err)
	}

	return result, nil
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
		DsuID:          deployPolicy.GetDsuId(),
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
		return nil, fmt.Errorf("invalid spec type(%s): %w", scope.GetType(), err)
	}

	// Get scope as structpb.Struct
	scopeStruct := scope.GetScope()
	if scopeStruct == nil {
		return nil, fmt.Errorf("scope is required for type %s", scopeType)
	}

	// Convert structpb.Struct to JSON, then unmarshal to protobuf message
	scopeJSON, err := scopeStruct.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scope struct to JSON: %w", err)
	}

	// Unmarshal to corresponding protobuf message based on type
	switch scopeType {
	case types.ScopeTypeServiceTemplate:
		var scopeProto ScopeServiceTemplate
		if err := protojson.Unmarshal(scopeJSON, &scopeProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scope for type %s: %w", scopeType, err)
		}
		filter := convTargetFilterToTypes(scopeProto.Filter)
		return types.NewScopeWithServiceTemplate(&types.ScopeServiceTemplate{
			Granularity:        types.TargetGranularity(scopeProto.Granularity),
			BizID:              scopeProto.BkBizId,
			Filter:             filter,
			ServiceTemplateIDs: scopeProto.ServiceTemplateIds,
			ModuleIDs:          scopeProto.ModuleIds,
		})

	case types.ScopeTypeSetTemplate:
		var scopeProto ScopeSetTemplate
		if err := protojson.Unmarshal(scopeJSON, &scopeProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scope for type %s: %w", scopeType, err)
		}
		filter := convTargetFilterToTypes(scopeProto.Filter)
		return types.NewScopeWithSetTemplate(&types.ScopeSetTemplate{
			Granularity:    types.TargetGranularity(scopeProto.Granularity),
			BizID:          scopeProto.BkBizId,
			Filter:         filter,
			SetTemplateIDs: scopeProto.SetTemplateIds,
			SetIDs:         scopeProto.SetIds,
		})

	case types.ScopeTypeInstance:
		var scopeProto ScopeInstance
		if err := protojson.Unmarshal(scopeJSON, &scopeProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scope for type %s: %w", scopeType, err)
		}
		filter := convTargetFilterToTypes(scopeProto.Filter)
		return types.NewScopeWithInstance(&types.ScopeInstance{
			Granularity: types.TargetGranularity(scopeProto.Granularity),
			BizID:       scopeProto.BkBizId,
			Filter:      filter,
			InstanceIDs: scopeProto.InstanceIds,
		})

	case types.ScopeTypeTopo:
		var scopeProto ScopeTopo
		if err := protojson.Unmarshal(scopeJSON, &scopeProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scope for type %s: %w", scopeType, err)
		}
		filter := convTargetFilterToTypes(scopeProto.Filter)
		paths := make([]*types.ScopeTopoNode, 0, len(scopeProto.Paths))
		for _, p := range scopeProto.Paths {
			paths = append(paths, &types.ScopeTopoNode{
				TopoObjID:  p.TopoObjId,
				TopoInstID: p.TopoInstId,
			})
		}
		return types.NewScopeWithTopo(&types.ScopeTopo{
			Granularity: types.TargetGranularity(scopeProto.Granularity),
			BizID:       scopeProto.BkBizId,
			Filter:      filter,
			Paths:       paths,
		})

	case types.ScopeTypeDynamicGroup:
		var scopeProto ScopeDynamicGroup
		if err := protojson.Unmarshal(scopeJSON, &scopeProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scope for type %s: %w", scopeType, err)
		}
		filter := convTargetFilterToTypes(scopeProto.Filter)
		return types.NewScopeWithDynamicGroup(&types.ScopeDynamicGroup{
			Granularity:     types.TargetGranularity(scopeProto.Granularity),
			BizID:           scopeProto.BkBizId,
			Filter:          filter,
			DynamicGroupIDs: scopeProto.DynamicGroupIds,
		})

	default:
		return nil, fmt.Errorf("unknown scope type: %s", scopeType)
	}
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

	// Get param as structpb.Struct
	paramStruct := spec.GetParam()
	if paramStruct == nil {
		return nil, fmt.Errorf("param is required for type %s", specType)
	}

	// Convert structpb.Struct to JSON, then unmarshal to protobuf message
	paramJSON, err := paramStruct.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal param struct to JSON: %w", err)
	}

	// Unmarshal to corresponding protobuf message based on type
	switch specType {
	case types.DeploySpecTypeSpecifyAgent:
		var paramProto SpecifyAgentParam
		if err := protojson.Unmarshal(paramJSON, &paramProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal param for type %s: %w", specType, err)
		}
		return types.NewDeploySpecWithSpecifyAgent(&types.SpecifyAgentParam{
			NodeVersion: paramProto.NodeVersion,
		})

	case types.DeploySpecTypeSpecifyProxy:
		var paramProto SpecifyProxyParam
		if err := protojson.Unmarshal(paramJSON, &paramProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal param for type %s: %w", specType, err)
		}
		return types.NewDeploySpecWithSpecifyProxy(&types.SpecifyProxyParam{
			NodeVersion: paramProto.NodeVersion,
		})

	case types.DeploySpecTypeSpecifyPlugin:
		var paramProto SpecifyPluginParam
		if err := protojson.Unmarshal(paramJSON, &paramProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal param for type %s: %w", specType, err)
		}
		customConfigContext := make(map[string]any)
		if paramProto.CustomConfigContext != nil {
			customConfigContext = paramProto.CustomConfigContext.AsMap()
		}
		return types.NewDeploySpecWithSpecifyPlugin(&types.SpecifyPluginParam{
			PluginName:          paramProto.PluginName,
			Version:             paramProto.Version,
			CustomConfigContext: customConfigContext,
		})

	case types.DeploySpecTypeSpecifyPluginPkg:
		var paramProto SpecifyPluginPkgParam
		if err := protojson.Unmarshal(paramJSON, &paramProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal param for type %s: %w", specType, err)
		}
		customConfigContext := make(map[string]any)
		if paramProto.CustomConfigContext != nil {
			customConfigContext = paramProto.CustomConfigContext.AsMap()
		}
		return types.NewDeploySpecWithSpecifyPluginPkg(&types.SpecifyPluginPkgParam{
			PluginPkgName:       paramProto.PluginPkgName,
			Version:             paramProto.Version,
			CustomConfigContext: customConfigContext,
		})

	case types.DeploySpecTypeSpecifyPluginSubConfig:
		var paramProto SpecifyPluginSubConfigParam
		if err := protojson.Unmarshal(paramJSON, &paramProto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal param for type %s: %w", specType, err)
		}
		configFilesDetail := make([]*types.PluginConfigDetail, 0, len(paramProto.ConfigFilesDetail))
		for _, detailProto := range paramProto.ConfigFilesDetail {
			configFilesDetail = append(configFilesDetail, &types.PluginConfigDetail{
				Name:         detailProto.Name,
				Content:      detailProto.Content,
				IsMainConfig: detailProto.IsMainConfig,
			})
		}
		customConfigContext := make(map[string]any)
		if paramProto.CustomConfigContext != nil {
			customConfigContext = paramProto.CustomConfigContext.AsMap()
		}
		return types.NewDeploySpecWithSpecifyPluginSubConfig(&types.SpecifyPluginSubConfigParam{
			PluginName:          paramProto.PluginName,
			ConfigFilesDetail:   configFilesDetail,
			CustomConfigContext: customConfigContext,
		})

	default:
		return nil, fmt.Errorf("unknown deploy spec type: %s", specType)
	}
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
	if x.GetDeployPolicyId() < 0 {
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
