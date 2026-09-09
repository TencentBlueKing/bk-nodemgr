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

package deploypolicy

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler deploy policy handler interface.
type IHandler interface {
	// Create create a new deploy policy.
	Create(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error)

	// Count count deploy policy by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists deploy policy by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicy, int64, error)

	// Get gets a deploy policy by conditions.
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicy, error)

	// UpdateFields update a deploy policy by fields.
	UpdateFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error

	// Delete delete a deploy policy by conditions.
	Delete(nCtx contextx.IContext, deployPolicyID int64) error

	// Exist check a deploy policy exist by conditions.
	Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error)

	// RefreshExecuteInfo refresh deploy policies execute info.
	RefreshExecuteInfo(nCtx contextx.IContext, deployPolicy ...*types.DeployPolicy) error
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate deploy policy table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	tableName := TableName(tenantID)

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure deploy policy indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new deploy policy handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create create a new deploy policy.
func (h *Handler) Create(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return -1, err
	}

	data, err := convDeployPolicyFromTypes(deployPolicy, nCtx.TenantID())
	if err != nil {
		return -1, fmt.Errorf("failed to convert deploy policy: %w", err)
	}

	// generate deploy policy id.
	deployPolicyID, err := h.tenantDao(nCtx.TenantID()).counter.Generate(nCtx, tableNamePrefix)
	if err != nil {
		return -1, fmt.Errorf("failed to generate deploy policy id, err: %w", err)
	}
	data.DeployPolicyID = deployPolicyID
	data.Operator = nCtx.BKUsername()
	data.LifeCycle.CreatedAt = time.Now()
	data.LifeCycle.UpdatedAt = time.Now()

	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return -1, fmt.Errorf("failed to create deploy policy, err: %w", err)
	}

	return deployPolicyID, nil
}

func convDeployPolicyFromTypes(deployPolicy *types.DeployPolicy, tenantID string) (*DeployPolicy, error) {
	if deployPolicy == nil {
		return nil, errors.New("deploy policy is nil")
	}

	specs, err := convSpecsFromTypes(deployPolicy.Specs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert specs: %w", err)
	}

	scopes, err := convScopesFromTypes(deployPolicy.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert scopes: %w", err)
	}

	data := &DeployPolicy{
		TenantID:       tenantID,
		DeployPolicyID: deployPolicy.DeployPolicyID,
		DsuID:          deployPolicy.DsuID,
		Meta:           convDeployPolicyMetaFromTypes(deployPolicy.Meta),
		Specs:          specs,
		Scopes:         scopes,
		Operator:       deployPolicy.Operator,
		Enabled:        deployPolicy.Enabled,
		EnsureAbsent:   deployPolicy.EnsureAbsent,
		LifeCycle:      convDeployPolicyLifeCycleFromTypes(deployPolicy.LifeCycle),
	}

	return data, nil
}

func convDeployPolicyMetaFromTypes(deployPolicyMeta types.DeployPolicyMeta) Meta {
	return Meta{
		Name:        deployPolicyMeta.Name,
		Description: deployPolicyMeta.Description,
	}
}

func convSpecsFromTypes(specs []*types.DeploySpec) ([]*Spec, error) {
	return conv.SliceToSliceWithError[*types.DeploySpec, *Spec](specs, convSpecFromTypes)
}

func convSpecFromTypes(spec *types.DeploySpec) (*Spec, error) {
	if spec == nil {
		return nil, errors.New("deploy spec is nil")
	}

	dbSpec := &Spec{
		Type: string(spec.Type()),
	}

	switch spec.Type() {
	case types.DeploySpecTypeSpecifyAgent:
		param, err := spec.GetSpecifyAgentParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify agent param: %w", err)
		}
		dbSpec.ParamSpecifyAgent = &SpecParamSpecifyAgent{
			NodeVersion: param.NodeVersion,
		}

	case types.DeploySpecTypeSpecifyPlugin:
		param, err := spec.GetSpecifyPluginParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin param: %w", err)
		}
		dbSpec.ParamSpecifyPlugin = &SpecParamSpecifyPlugin{
			PluginName:          param.PluginName,
			Version:             param.Version,
			CustomConfigContext: param.CustomConfigContext,
		}

	case types.DeploySpecTypeSpecifyPluginPkg:
		param, err := spec.GetSpecifyPluginPkgParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin pkg param: %w", err)
		}
		dbSpec.ParamSpecifyPluginPkg = &SpecParamSpecifyPluginPkg{
			PluginPkgName:       param.PluginPkgName,
			Version:             param.Version,
			CustomConfigContext: param.CustomConfigContext,
		}

	case types.DeploySpecTypeProjectPluginPkgToHosts:
		param, err := spec.GetProjectPluginPkgToHostsParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get project plugin pkg to hosts param: %w", err)
		}
		dbSpec.ParamProjectPluginPkgToHosts = &SpecParamProjectPluginPkgToHosts{
			PluginPkgName:       param.PluginPkgName,
			Version:             param.Version,
			CustomConfigContext: param.CustomConfigContext,
			PlacementHostIDs:    param.PlacementHostIDs,
		}

	case types.DeploySpecTypeProjectPluginConfigTemplateToHosts:
		param, err := spec.GetProjectPluginConfigTemplateToHostsParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get project plugin config template to hosts param: %w", err)
		}
		dbSpec.ParamProjectPluginConfigTemplateToHosts = &SpecParamProjectPluginConfigTemplateToHosts{
			PluginName:          param.PluginName,
			ConfigFilesDetail:   convPluginConfigDetailsFromTypes(param.ConfigFilesDetail),
			CustomConfigContext: param.CustomConfigContext,
		}

	case types.DeploySpecTypeSpecifyPluginSubConfig:
		param, err := spec.GetSpecifyPluginSubConfigParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin sub config param: %w", err)
		}
		dbSpec.ParamSpecifyPluginSubConfig = &SpecParamSpecifyPluginSubConfig{
			PluginName:          param.PluginName,
			ConfigFilesDetail:   convPluginConfigDetailsFromTypes(param.ConfigFilesDetail),
			CustomConfigContext: param.CustomConfigContext,
		}

	case types.DeploySpecTypeSpecifyPluginSubConfigTemplate:
		param, err := spec.GetSpecifyPluginSubConfigTemplateParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin sub config template param: %w", err)
		}
		dbSpec.ParamSpecifyPluginSubConfigTemplate = &SpecParamSpecifyPluginSubConfigTemplate{
			PluginName:          param.PluginName,
			ConfigFilesDetail:   convPluginConfigDetailsFromTypes(param.ConfigFilesDetail),
			CustomConfigContext: param.CustomConfigContext,
		}

	case types.DeploySpecTypeSpecifyProxy:
		param, err := spec.GetSpecifyProxyParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify proxy param: %w", err)
		}
		dbSpec.ParamSpecifyProxy = &SpecParamSpecifyProxy{
			NodeVersion: param.NodeVersion,
		}

	default:
		return nil, fmt.Errorf("unknown deploy spec type: %s", spec.Type())
	}

	return dbSpec, nil
}

func convScopesFromTypes(scopes []*types.Scope) ([]*Scope, error) {
	return conv.SliceToSliceWithError[*types.Scope, *Scope](scopes, convScopeFromTypes)
}

func convScopeFromTypes(scope *types.Scope) (*Scope, error) {
	if scope == nil {
		return nil, errors.New("scope is nil")
	}

	dbScope := &Scope{
		Type: string(scope.Type()),
	}

	switch scope.Type() {
	case types.ScopeTypeServiceTemplate:
		scopeItem, err := scope.GetServiceTemplateScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get service template scope: %w", err)
		}

		dbScope.ScopeServiceTemplate = &ScopeServiceTemplate{
			Granularity:        string(scopeItem.Granularity),
			BizID:              scopeItem.BizID,
			Filter:             convTargetFilterFromTypes(scopeItem.Filter),
			ServiceTemplateIDs: scopeItem.ServiceTemplateIDs,
			ModuleIDs:          scopeItem.ModuleIDs,
		}

	case types.ScopeTypeSetTemplate:
		scopeItem, err := scope.GetSetTemplateScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get set template scope: %w", err)
		}

		dbScope.ScopeSetTemplate = &ScopeSetTemplate{
			Granularity:    string(scopeItem.Granularity),
			BizID:          scopeItem.BizID,
			Filter:         convTargetFilterFromTypes(scopeItem.Filter),
			SetTemplateIDs: scopeItem.SetTemplateIDs,
			SetIDs:         scopeItem.SetIDs,
		}

	case types.ScopeTypeInstance:
		scopeItem, err := scope.GetInstanceScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get instance scope: %w", err)
		}

		dbScope.ScopeInstance = &ScopeInstance{
			Granularity: string(scopeItem.Granularity),
			BizID:       scopeItem.BizID,
			Filter:      convTargetFilterFromTypes(scopeItem.Filter),
			InstanceIDs: scopeItem.InstanceIDs,
		}

	case types.ScopeTypeTopo:
		scopeItem, err := scope.GetTopoScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get topo scope: %w", err)
		}

		paths := make([]*ScopeTopoNode, 0, len(scopeItem.Paths))
		for _, node := range scopeItem.Paths {
			paths = append(paths, &ScopeTopoNode{
				TopoObjID:  node.TopoObjID,
				TopoInstID: node.TopoInstID,
			})
		}

		dbScope.ScopeTopo = &ScopeTopo{
			Granularity: string(scopeItem.Granularity),
			BizID:       scopeItem.BizID,
			Filter:      convTargetFilterFromTypes(scopeItem.Filter),
			Paths:       paths,
		}

	case types.ScopeTypeDynamicGroup:
		scopeItem, err := scope.GetDynamicGroupScope()
		if err != nil {
			return nil, fmt.Errorf("failed to get dynamic group scope: %w", err)
		}

		dbScope.ScopeDynamicGroup = &ScopeDynamicGroup{
			Granularity:     string(scopeItem.Granularity),
			BizID:           scopeItem.BizID,
			Filter:          convTargetFilterFromTypes(scopeItem.Filter),
			DynamicGroupIDs: scopeItem.DynamicGroupIDs,
		}

	default:
		return nil, fmt.Errorf("unknown scope type: %s", scope.Type())
	}

	return dbScope, nil
}

func convDeployPolicyLifeCycleFromTypes(deployPolicyLifeCycle types.DeployPolicyLifeCycle) LifeCycle {
	return LifeCycle{
		CreatedAt:  deployPolicyLifeCycle.CreatedAt,
		UpdatedAt:  deployPolicyLifeCycle.UpdatedAt,
		ExecutedAt: deployPolicyLifeCycle.ExecutedAt,
	}
}

func convTargetFilterFromTypes(targetFilter *types.TargetFilter) *TargetFilter {
	if targetFilter == nil {
		return nil
	}

	return &TargetFilter{}
}

// Count count deploy policy by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
}

// List list deploy policy by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicy, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	data, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	deployPolicies, err := conv.SliceToSliceWithError[*DeployPolicy, *types.DeployPolicy](data, convDeployPolicyToTypes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert deploy policies: %w", err)
	}

	return deployPolicies, num, nil
}

// Get get a deploy policy.
func (h *Handler) Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := h.tenantDao(nCtx.TenantID()).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convDeployPolicyToTypes(data)
}

func convDeployPolicyToTypes(data *DeployPolicy) (*types.DeployPolicy, error) {
	if data == nil {
		return nil, errors.New("deploy policy data is nil")
	}

	deployPolicy := &types.DeployPolicy{
		DeployPolicyID: data.DeployPolicyID,
		DsuID:          data.DsuID,
		Operator:       data.Operator,
		Enabled:        data.Enabled,
		EnsureAbsent:   data.EnsureAbsent,
		LifeCycle:      convDeployPolicyLifeCycleToTypes(data.LifeCycle),
	}

	deployPolicy.Meta = types.DeployPolicyMeta{
		Name:        data.Meta.Name,
		Description: data.Meta.Description,
	}

	if len(data.Scopes) > 0 {
		scopes, err := conv.SliceToSliceWithError[*Scope, *types.Scope](data.Scopes, convScopeToTypes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scopes: %w", err)
		}
		deployPolicy.Scopes = scopes
	}

	if len(data.Specs) > 0 {
		specs, err := conv.SliceToSliceWithError[*Spec, *types.DeploySpec](data.Specs, convSpecToTypes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert specs: %w", err)
		}
		deployPolicy.Specs = specs
	}

	return deployPolicy, nil
}

func convScopeToTypes(data *Scope) (*types.Scope, error) {
	if data == nil {
		return nil, errors.New("scope data is nil")
	}

	scopeType := types.ScopeType(data.Type)

	switch scopeType {
	case types.ScopeTypeServiceTemplate:
		if data.ScopeServiceTemplate == nil {
			return nil, fmt.Errorf("scope_service_template is nil")
		}

		return types.NewScopeWithServiceTemplate(&types.ScopeServiceTemplate{
			Granularity:        types.TargetGranularity(data.ScopeServiceTemplate.Granularity),
			BizID:              data.ScopeServiceTemplate.BizID,
			Filter:             convTargetFilterToTypes(data.ScopeServiceTemplate.Filter),
			ServiceTemplateIDs: data.ScopeServiceTemplate.ServiceTemplateIDs,
			ModuleIDs:          data.ScopeServiceTemplate.ModuleIDs,
		})

	case types.ScopeTypeSetTemplate:
		if data.ScopeSetTemplate == nil {
			return nil, fmt.Errorf("scope_set_template is nil")
		}

		return types.NewScopeWithSetTemplate(&types.ScopeSetTemplate{
			Granularity:    types.TargetGranularity(data.ScopeSetTemplate.Granularity),
			BizID:          data.ScopeSetTemplate.BizID,
			Filter:         convTargetFilterToTypes(data.ScopeSetTemplate.Filter),
			SetTemplateIDs: data.ScopeSetTemplate.SetTemplateIDs,
			SetIDs:         data.ScopeSetTemplate.SetIDs,
		})

	case types.ScopeTypeInstance:
		if data.ScopeInstance == nil {
			return nil, fmt.Errorf("scope_instance is nil")
		}

		return types.NewScopeWithInstance(&types.ScopeInstance{
			Granularity: types.TargetGranularity(data.ScopeInstance.Granularity),
			BizID:       data.ScopeInstance.BizID,
			Filter:      convTargetFilterToTypes(data.ScopeInstance.Filter),
			InstanceIDs: data.ScopeInstance.InstanceIDs,
		})

	case types.ScopeTypeTopo:
		if data.ScopeTopo == nil {
			return nil, fmt.Errorf("scope_topo is nil")
		}

		paths := make([]*types.ScopeTopoNode, 0, len(data.ScopeTopo.Paths))
		for _, node := range data.ScopeTopo.Paths {
			paths = append(paths, &types.ScopeTopoNode{
				TopoObjID:  node.TopoObjID,
				TopoInstID: node.TopoInstID,
			})
		}

		return types.NewScopeWithTopo(&types.ScopeTopo{
			Granularity: types.TargetGranularity(data.ScopeTopo.Granularity),
			BizID:       data.ScopeTopo.BizID,
			Filter:      convTargetFilterToTypes(data.ScopeTopo.Filter),
			Paths:       paths,
		})

	case types.ScopeTypeDynamicGroup:
		if data.ScopeDynamicGroup == nil {
			return nil, fmt.Errorf("scope_dynamic_group is nil")
		}

		return types.NewScopeWithDynamicGroup(&types.ScopeDynamicGroup{
			Granularity:     types.TargetGranularity(data.ScopeDynamicGroup.Granularity),
			BizID:           data.ScopeDynamicGroup.BizID,
			Filter:          convTargetFilterToTypes(data.ScopeDynamicGroup.Filter),
			DynamicGroupIDs: data.ScopeDynamicGroup.DynamicGroupIDs,
		})

	default:
		return nil, fmt.Errorf("unknown scope type: %s", data.Type)
	}
}

func convDeployPolicyLifeCycleToTypes(data LifeCycle) types.DeployPolicyLifeCycle {
	return types.DeployPolicyLifeCycle{
		CreatedAt:  data.CreatedAt,
		UpdatedAt:  data.UpdatedAt,
		ExecutedAt: data.ExecutedAt,
	}
}

func convTargetFilterToTypes(data *TargetFilter) *types.TargetFilter {
	if data == nil {
		return nil
	}

	// TODO: implement me.

	return &types.TargetFilter{}
}

func convCustomConfigContextToTypes(ctx map[string]any) map[string]any {
	if len(ctx) == 0 {
		return ctx
	}

	result := make(map[string]any, len(ctx))
	for key, value := range ctx {
		result[key] = convCustomConfigValueToTypes(value)
	}

	return result
}

func convCustomConfigValueToTypes(value any) any {
	switch typed := value.(type) {
	case primitive.M:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = convCustomConfigValueToTypes(item)
		}

		return result
	case primitive.A:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, convCustomConfigValueToTypes(item))
		}

		return result
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = convCustomConfigValueToTypes(item)
		}

		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, convCustomConfigValueToTypes(item))
		}

		return result
	default:
		return value
	}
}

func convSpecToTypes(data *Spec) (*types.DeploySpec, error) {
	if data == nil {
		return nil, errors.New("spec data is nil")
	}

	specType := types.DeploySpecType(data.Type)

	switch specType {
	case types.DeploySpecTypeSpecifyAgent:
		if data.ParamSpecifyAgent == nil {
			return nil, fmt.Errorf("param_specify_agent is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithSpecifyAgent(&types.SpecifyAgentParam{
			NodeVersion: data.ParamSpecifyAgent.NodeVersion,
		})

	case types.DeploySpecTypeSpecifyPlugin:
		if data.ParamSpecifyPlugin == nil {
			return nil, fmt.Errorf("param_specify_plugin is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithSpecifyPlugin(&types.SpecifyPluginParam{
			PluginName:          data.ParamSpecifyPlugin.PluginName,
			Version:             data.ParamSpecifyPlugin.Version,
			CustomConfigContext: convCustomConfigContextToTypes(data.ParamSpecifyPlugin.CustomConfigContext),
		})

	case types.DeploySpecTypeSpecifyPluginPkg:
		if data.ParamSpecifyPluginPkg == nil {
			return nil, fmt.Errorf("param_specify_plugin_pkg is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithSpecifyPluginPkg(&types.SpecifyPluginPkgParam{
			PluginPkgName:       data.ParamSpecifyPluginPkg.PluginPkgName,
			Version:             data.ParamSpecifyPluginPkg.Version,
			CustomConfigContext: convCustomConfigContextToTypes(data.ParamSpecifyPluginPkg.CustomConfigContext),
		})

	case types.DeploySpecTypeProjectPluginPkgToHosts:
		if data.ParamProjectPluginPkgToHosts == nil {
			return nil, fmt.Errorf("param_project_plugin_pkg_to_hosts is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithProjectPluginPkgToHosts(&types.ProjectPluginPkgToHostsParam{
			PluginPkgName:       data.ParamProjectPluginPkgToHosts.PluginPkgName,
			Version:             data.ParamProjectPluginPkgToHosts.Version,
			CustomConfigContext: convCustomConfigContextToTypes(data.ParamProjectPluginPkgToHosts.CustomConfigContext),
			PlacementHostIDs:    data.ParamProjectPluginPkgToHosts.PlacementHostIDs,
		})

	case types.DeploySpecTypeProjectPluginConfigTemplateToHosts:
		if data.ParamProjectPluginConfigTemplateToHosts == nil {
			return nil, fmt.Errorf("param_project_plugin_config_template_to_hosts is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithProjectPluginConfigTemplateToHosts(&types.ProjectPluginConfigTemplateToHostsParam{
			PluginName: data.ParamProjectPluginConfigTemplateToHosts.PluginName,
			ConfigFilesDetail: convPluginConfigDetailsToTypes(
				data.ParamProjectPluginConfigTemplateToHosts.ConfigFilesDetail,
			),
			CustomConfigContext: convCustomConfigContextToTypes(
				data.ParamProjectPluginConfigTemplateToHosts.CustomConfigContext,
			),
		})

	case types.DeploySpecTypeSpecifyPluginSubConfig:
		if data.ParamSpecifyPluginSubConfig == nil {
			return nil, fmt.Errorf("param_specify_plugin_sub_config is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithSpecifyPluginSubConfig(&types.SpecifyPluginSubConfigParam{
			PluginName:          data.ParamSpecifyPluginSubConfig.PluginName,
			ConfigFilesDetail:   convPluginConfigDetailsToTypes(data.ParamSpecifyPluginSubConfig.ConfigFilesDetail),
			CustomConfigContext: convCustomConfigContextToTypes(data.ParamSpecifyPluginSubConfig.CustomConfigContext),
		})

	case types.DeploySpecTypeSpecifyPluginSubConfigTemplate:
		if data.ParamSpecifyPluginSubConfigTemplate == nil {
			return nil, fmt.Errorf("param_specify_plugin_sub_config_template is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithSpecifyPluginSubConfigTemplate(&types.SpecifyPluginSubConfigTemplateParam{
			PluginName: data.ParamSpecifyPluginSubConfigTemplate.PluginName,
			ConfigFilesDetail: convPluginConfigDetailsToTypes(
				data.ParamSpecifyPluginSubConfigTemplate.ConfigFilesDetail,
			),
			CustomConfigContext: convCustomConfigContextToTypes(
				data.ParamSpecifyPluginSubConfigTemplate.CustomConfigContext,
			),
		})

	case types.DeploySpecTypeSpecifyProxy:
		if data.ParamSpecifyProxy == nil {
			return nil, fmt.Errorf("param_specify_proxy is required for type %s", data.Type)
		}

		return types.NewDeploySpecWithSpecifyProxy(&types.SpecifyProxyParam{
			NodeVersion: data.ParamSpecifyProxy.NodeVersion,
		})

	default:
		return nil, fmt.Errorf("unknown deploy spec type: %s", data.Type)
	}
}

func convPluginConfigDetailsFromTypes(details []*types.PluginConfigDetail) []*SpecPluginConfigDetail {
	configFilesDetail := make([]*SpecPluginConfigDetail, 0, len(details))
	for _, detail := range details {
		if detail == nil {
			continue
		}

		configFilesDetail = append(configFilesDetail, &SpecPluginConfigDetail{
			Name:         detail.Name,
			TemplateName: detail.TemplateName,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
		})
	}

	return configFilesDetail
}

func convPluginConfigDetailsToTypes(details []*SpecPluginConfigDetail) []*types.PluginConfigDetail {
	configFilesDetail := make([]*types.PluginConfigDetail, 0, len(details))
	for _, detail := range details {
		if detail == nil {
			continue
		}

		configFilesDetail = append(configFilesDetail, &types.PluginConfigDetail{
			Name:         detail.Name,
			TemplateName: detail.TemplateName,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
		})
	}

	return configFilesDetail
}

// Delete delete deploy policy.
func (h *Handler) Delete(nCtx contextx.IContext, deployPolicyID int64) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %w", err)
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithDeployPolicyID(deployPolicyID),
	}

	for _, opt := range opts {
		filter = opt(filter)
	}

	err := h.tenantDao(nCtx.TenantID()).DeleteMany(nCtx, filter)

	if err != nil {
		logger.G.Sys().With("deploy-policy-id", deployPolicyID).WithErr(err).Error("failed to delete deploy policy")

		return fmt.Errorf("failed to delete deploy policy: %v", err)
	}

	return nil
}

// Exist check deploy policy exist.
func (h *Handler) Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	exist, err := h.tenantDao(nCtx.TenantID()).Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check deploy policy exist: %v", err)
	}

	return exist, nil
}

// UpdateFields update a deploy policy by fields.
func (h *Handler) UpdateFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if len(deployPolicy) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("deploy policy list is empty"))
	}

	// Check if at least one field is set to update
	if !fields.Meta && !fields.Scopes && !fields.Specs && !fields.Enabled && !fields.EnsureAbsent {
		return base.ErrInvalidParam(fmt.Errorf("at least one field must be set to update"))
	}

	tenantID := nCtx.TenantID()

	docs := make([]*base.DocumentFieldUpdate, 0, len(deployPolicy))
	for _, policy := range deployPolicy {
		if policy == nil {
			return base.ErrInvalidItemInParamList()
		}

		updates, err := generateDeployPolicyUpdates(fields, policy)
		if err != nil {
			return fmt.Errorf("failed to generate deploy policy updates: %w", err)
		}
		if len(updates) == 0 {
			continue
		}

		updates[FieldKeyOperator] = nCtx.BKUsername()
		updates[FieldKeyLifeCycleUpdatedAt] = time.Now()

		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: func() bson.D {
				filter := base.AliveFilter()
				filter = WithDeployPolicyID(policy.DeployPolicyID)(filter)

				return filter
			}(),
			Fields: updates,
		})
	}

	if len(docs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("no valid updates to apply"))
	}

	if err := h.tenantDao(tenantID).UpdateOneFieldBulk(nCtx, docs); err != nil {
		return fmt.Errorf("failed to update deploy policy fields: %v", err)
	}

	return nil
}

// generateDeployPolicyUpdates generate deploy policy updates.
func generateDeployPolicyUpdates(fields types.DeployPolicyFields, deployPolicy *types.DeployPolicy) (map[string]any, error) {
	updates := make(map[string]any)

	if fields.Meta {
		updates[FieldKeyMeta] = convDeployPolicyMetaFromTypes(deployPolicy.Meta)
	}

	if fields.Scopes {
		scopes, err := convScopesFromTypes(deployPolicy.Scopes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scopes: %w", err)
		}
		updates[FieldKeyScopes] = scopes
	}

	if fields.Specs {
		specs, err := convSpecsFromTypes(deployPolicy.Specs)
		if err != nil {
			return nil, fmt.Errorf("failed to convert specs: %w", err)
		}
		updates[FieldKeySpecs] = specs
	}

	if fields.Enabled {
		updates[FieldKeyEnabled] = deployPolicy.Enabled
	}

	if fields.EnsureAbsent {
		updates[FieldKeyEnsureAbsent] = deployPolicy.EnsureAbsent
	}

	return updates, nil
}

// RefreshExecuteInfo refresh deploy policies execute info.
func (h *Handler) RefreshExecuteInfo(nCtx contextx.IContext, deployPolicy ...*types.DeployPolicy) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %w", err)
	}

	if len(deployPolicy) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("deploy policy list is empty"))
	}

	tenantID := nCtx.TenantID()

	docs := make([]*base.DocumentFieldUpdate, 0, len(deployPolicy))
	for _, policy := range deployPolicy {
		if policy == nil {
			return base.ErrInvalidItemInParamList()
		}

		updates := map[string]any{
			FieldKeyLifeCycleExecutedAt: policy.LifeCycle.ExecutedAt,
			FieldKeyDsuID:               policy.DsuID,
			FieldKeyLifeCycleUpdatedAt:  time.Now(),
		}

		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: func() bson.D {
				filter := base.AliveFilter()
				filter = WithDeployPolicyID(policy.DeployPolicyID)(filter)

				return filter
			}(),
			Fields: updates,
		})
	}

	if len(docs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("no valid updates to apply"))
	}

	if err := h.tenantDao(tenantID).UpdateOneFieldBulk(nCtx, docs); err != nil {
		return fmt.Errorf("failed to update deploy policies executed at and dsuid: %w", err)
	}

	return nil
}
