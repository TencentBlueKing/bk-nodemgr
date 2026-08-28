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
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	// deploy policy list max limit.
	maxDeployPolicyLimit = 1000
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

// PageLimit return page limit.
func (x *DeployPolicyListReq) PageLimit() int {
	return maxDeployPolicyLimit
}

// ConvertPageToTypes convert page to types.
func (x *DeployPolicyListReq) ConvertPageToTypes() (types.Page, error) {
	page, err := convPageToTypes(x.GetPage())
	if err != nil {
		return page, err
	}

	if page.Limit <= 0 || page.Limit > x.PageLimit() {
		return page, fmt.Errorf("page.limit must be in (0, %d]", x.PageLimit())
	}

	return page, nil
}

// ConvertConditionsToTypes convert conditions to types.
func (x *DeployPolicyListReq) ConvertConditionsToTypes() (*types.DeployPolicyCondition, error) {
	executedTimeRangeCond, err := convertTimeRangeToTypes(x.GetExecutedTimeRange())
	if err != nil {
		return nil, fmt.Errorf("failed to convert executed time range: %w", err)
	}

	return &types.DeployPolicyCondition{
		ExecutedTimeRange: executedTimeRangeCond,
		ExactInclude:      convertDeployPolicyExactConditionsToTypes(x.GetExactIncludeConditions()),
		FuzzyInclude:      convertDeployPolicyFuzzyConditionsToTypes(x.GetFuzzyIncludeConditions()),
		ExactExclude:      convertDeployPolicyExactConditionsToTypes(x.GetExactExcludeConditions()),
		FuzzyExclude:      convertDeployPolicyFuzzyConditionsToTypes(x.GetFuzzyExcludeConditions()),
	}, nil
}

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

	specs, err := convDeploySpecsFromTypes(deployPolicy.Specs)
	if err != nil {
		return nil, err
	}

	scopes, err := convDeployPolicyScopesFromTypes(deployPolicy.Scopes)
	if err != nil {
		return nil, err
	}

	return &DeployPolicy{
		DeployPolicyId: deployPolicy.DeployPolicyID,
		DsuId:          deployPolicy.DsuID,
		Meta:           convDeployPolicyMetaFromTypes(deployPolicy.Meta),
		Specs:          specs,
		Scopes:         scopes,
		Operator:       deployPolicy.Operator,
		Enabled:        deployPolicy.Enabled,
	}, nil
}

func convDeployPolicyMetaFromTypes(deployPolicyMeta types.DeployPolicyMeta) *DeployPolicyMeta {
	return &DeployPolicyMeta{
		Name:        deployPolicyMeta.Name,
		Description: deployPolicyMeta.Description,
	}
}

func convDeploySpecsFromTypes(specs []*types.DeploySpec) ([]*DeploySpec, error) {
	return conv.SliceToSliceWithError[*types.DeploySpec, *DeploySpec](specs, convDeploySpecFromTypes)
}

func convDeploySpecFromTypes(spec *types.DeploySpec) (*DeploySpec, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is nil")
	}

	specType := spec.Type()
	result := &DeploySpec{Type: string(specType)}
	var paramProto interface{}

	switch specType {
	case types.DeploySpecTypeSpecifyAgent:
		param, err := spec.GetSpecifyAgentParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify agent param: %w", err)
		}
		paramProto = &SpecifyAgentParam{NodeVersion: param.NodeVersion}

	case types.DeploySpecTypeSpecifyProxy:
		param, err := spec.GetSpecifyProxyParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify proxy param: %w", err)
		}
		paramProto = &SpecifyProxyParam{NodeVersion: param.NodeVersion}

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

func convDeployPolicyScopesFromTypes(scopes []*types.Scope) ([]*Scope, error) {
	return conv.SliceToSliceWithError[*types.Scope, *Scope](scopes, convDeployPolicyScopeFromTypes)
}

func convDeployPolicyScopeFromTypes(scope *types.Scope) (*Scope, error) {
	if scope == nil {
		return nil, fmt.Errorf("scope is nil")
	}

	scopeType := scope.Type()
	result := &Scope{Type: string(scopeType)}
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
			Filter:             &TargetFilter{},
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
			Filter:         &TargetFilter{},
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
			Filter:      &TargetFilter{},
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
			Filter:      &TargetFilter{},
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
			Filter:          &TargetFilter{},
			DynamicGroupIds: item.DynamicGroupIDs,
		}

	default:
		return nil, fmt.Errorf("unknown scope type: %s", scopeType)
	}

	scopeMap, err := conv.StructToMap(scopeProto)
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to map: %w", err)
	}
	result.Scope, err = structpb.NewStruct(scopeMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to struct: %w", err)
	}

	return result, nil
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

// ConvertTriggerID convert trigger id.
func (x *DeployPolicyExecuteResp) ConvertTriggerID(triggerID string) {
	x.Data = &DeployPolicyExecuteResp_Data{TriggerId: triggerID}
}
