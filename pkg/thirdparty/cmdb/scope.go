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

package cmdb

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IScope define the scope interface.
type IScope interface {
	// GetTargetByScopeSetTemplate get target by scope set template.
	GetTargetByScopeSetTemplate(nCtx contextx.IContext, scope *types.ScopeSetTemplate) ([]*types.Target, error)

	// GetTargetByScopeServiceTemplate get target by scope service template.
	GetTargetByScopeServiceTemplate(nCtx contextx.IContext, scope *types.ScopeServiceTemplate) ([]*types.Target, error)

	// GetTargetByScopeInstance get target by scope instance.
	GetTargetByScopeInstance(nCtx contextx.IContext, scope *types.ScopeInstance) ([]*types.Target, error)

	// GetTargetByScopeTopo get target by scope topo.
	GetTargetByScopeTopo(nCtx contextx.IContext, scope *types.ScopeTopo) ([]*types.Target, error)

	// GetTargetByScopeDynamicGroup get target by scope dynamic group.
	GetTargetByScopeDynamicGroup(nCtx contextx.IContext, scope *types.ScopeDynamicGroup) ([]*types.Target, error)
}

var _ IScope = &Handler{}

const (
	ccQueryTimeout = 30 * time.Minute
)

// GetTargetByScopeSetTemplate get target by scope set template.
func (h *Handler) GetTargetByScopeSetTemplate(nCtx contextx.IContext, scope *types.ScopeSetTemplate) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target, nCtx is nil")
	}

	// TODO: 考虑此处的权限代持问题
	nCtx = contextx.From(nCtx, contextx.WithBKUsername(h.cli.config.VirtualUser))

	if scope == nil {
		return nil, fmt.Errorf("failed to get target, scope is nil")
	}

	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
	}

	var (
		targets []*types.Target
		err     error
	)
	switch scope.Granularity {
	case types.TargetGranularityHost:
		targets, err = h.getHostTargetByScopeSetTemplate(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
		}
	case types.TargetGranularityServiceInstance:
		targets, err = h.getServiceTargetByScopeSetTemplate(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
		}
	default:
		return nil, fmt.Errorf("failed to get target by scope set template, granularity(%s), not supported", scope.Granularity)
	}

	return targets, nil
}

func (h *Handler) getHostTargetByScopeSetTemplate(nCtx contextx.IContext, scope *types.ScopeSetTemplate) ([]*types.Target, error) {
	hosts, err := h.FindHostBySetTemplate(nCtx, scope.BizID, types.UnlimitedPage(), scope.SetTemplateIDs, scope.SetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
	}

	targets := convHostToTarget(hosts)

	return targets, nil
}

func (h *Handler) getServiceTargetByScopeSetTemplate(nCtx contextx.IContext, scope *types.ScopeSetTemplate) ([]*types.Target, error) {
	executor := pageexecutor.NewPageExecutor[*ServiceInstanceDetailInfo](CCPageSizeLimit, ccQueryTimeout)

	moduleIDs, err := h.getModuleIDsBySetTemplateIDs(nCtx, scope.BizID, scope.SetTemplateIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get module IDs by set template IDs: %w", err)
	}

	if len(moduleIDs) == 0 {
		return []*types.Target{}, nil
	}

	var allServiceInstances []*ServiceInstanceDetailInfo
	for _, moduleID := range moduleIDs {
		detailFn := h.buildListServiceInstanceDetailByModuleIDFn(scope.BizID, moduleID)
		result, err := executor.Execute(nCtx, types.UnlimitedPage(), detailFn)
		if err != nil {
			return nil, fmt.Errorf("failed to query service instances by module ID %d: %w", moduleID, err)
		}
		allServiceInstances = append(allServiceInstances, result.Items...)
	}

	if len(allServiceInstances) == 0 {
		return []*types.Target{}, nil
	}

	hostIDs := conv.SliceUnique(conv.SliceToSlice(allServiceInstances, func(inst *ServiceInstanceDetailInfo) int64 {
		return inst.BKHostID
	}))
	hostCond := &types.HostStaticExactCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: hostIDs,
		},
	}
	hosts, err := h.FindHostWithCondition(nCtx, types.UnlimitedPage(), hostCond)
	if err != nil {
		return nil, fmt.Errorf("failed to find hosts: %w", err)
	}

	targets, err := convServiceInstanceDetailToTarget(allServiceInstances, hosts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert service instance detail to target: %w", err)
	}

	return targets, nil
}

// GetTargetByScopeServiceTemplate get target by scope service template.
func (h *Handler) GetTargetByScopeServiceTemplate(nCtx contextx.IContext, scope *types.ScopeServiceTemplate) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope service template, nCtx is nil")
	}

	// TODO: 考虑此处的权限代持问题
	nCtx = contextx.From(nCtx, contextx.WithBKUsername(h.cli.config.VirtualUser))

	if scope == nil {
		return nil, fmt.Errorf("failed to get target by scope service template, scope is nil")
	}

	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("failed to get target by scope service template: %w", err)
	}

	var (
		targets []*types.Target
		err     error
	)
	switch scope.Granularity {
	case types.TargetGranularityHost:
		targets, err = h.getHostTargetByScopeServiceTemplate(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope service template: %w", err)
		}
	case types.TargetGranularityServiceInstance:
		targets, err = h.getServiceTargetByScopeServiceTemplate(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope service template: %w", err)
		}
	default:
		return nil, fmt.Errorf("failed to get target by scope service template, granularity(%s), not supported", scope.Granularity)
	}

	return targets, nil
}

func (h *Handler) getHostTargetByScopeServiceTemplate(nCtx contextx.IContext, scope *types.ScopeServiceTemplate) ([]*types.Target, error) {
	hosts, err := h.FindHostByServiceTemplate(nCtx, scope.BizID, types.UnlimitedPage(), scope.ServiceTemplateIDs, scope.ModuleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by scope service template: %w", err)
	}

	targets := convHostToTarget(hosts)

	return targets, nil
}

func (h *Handler) getServiceTargetByScopeServiceTemplate(nCtx contextx.IContext, scope *types.ScopeServiceTemplate) ([]*types.Target, error) {
	executor := pageexecutor.NewPageExecutor[*ServiceInstanceDetailInfo](CCPageSizeLimit, ccQueryTimeout)

	moduleIDs, err := h.getModuleIDsByServiceTemplateIDs(nCtx, scope.BizID, scope.ServiceTemplateIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get module IDs by service template IDs: %w", err)
	}

	if len(moduleIDs) == 0 {
		return []*types.Target{}, nil
	}

	var allServiceInstances []*ServiceInstanceDetailInfo
	for _, moduleID := range moduleIDs {
		detailFn := h.buildListServiceInstanceDetailByModuleIDFn(scope.BizID, moduleID)
		result, err := executor.Execute(nCtx, types.UnlimitedPage(), detailFn)
		if err != nil {
			return nil, fmt.Errorf("failed to query service instances by module ID %d: %w", moduleID, err)
		}
		allServiceInstances = append(allServiceInstances, result.Items...)
	}

	if len(allServiceInstances) == 0 {
		return []*types.Target{}, nil
	}

	hostIDs := conv.SliceUnique(conv.SliceToSlice(allServiceInstances, func(inst *ServiceInstanceDetailInfo) int64 {
		return inst.BKHostID
	}))
	hostCond := &types.HostStaticExactCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: hostIDs,
		},
	}
	hosts, err := h.FindHostWithCondition(nCtx, types.UnlimitedPage(), hostCond)
	if err != nil {
		return nil, fmt.Errorf("failed to find hosts: %w", err)
	}

	targets, err := convServiceInstanceDetailToTarget(allServiceInstances, hosts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert service instance detail to target: %w", err)
	}

	return targets, nil
}

// GetTargetByScopeInstance get target by scope instance.
func (h *Handler) GetTargetByScopeInstance(nCtx contextx.IContext, scope *types.ScopeInstance) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope instance, nCtx is nil")
	}

	// TODO: 考虑此处的权限代持问题
	nCtx = contextx.From(nCtx, contextx.WithBKUsername(h.cli.config.VirtualUser))

	if scope == nil {
		return nil, fmt.Errorf("failed to get target by scope instance, scope is nil")
	}

	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("failed to get target by scope instance: %w", err)
	}

	var (
		targets []*types.Target
		err     error
	)
	switch scope.Granularity {
	case types.TargetGranularityHost:
		targets, err = h.getHostTargetByScopeInstance(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope instance: %w", err)
		}
	case types.TargetGranularityServiceInstance:
		targets, err = h.getServiceTargetByScopeInstance(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope instance: %w", err)
		}
	default:
		return nil, fmt.Errorf("failed to get target by scope instance, granularity(%s), not supported", scope.Granularity)
	}

	return targets, nil
}

func (h *Handler) getServiceTargetByScopeInstance(nCtx contextx.IContext, scope *types.ScopeInstance) ([]*types.Target, error) {
	executor := pageexecutor.NewPageExecutor[*ServiceInstanceDetailInfo](CCPageSizeLimit, ccQueryTimeout)

	detailFn := h.buildListServiceInstanceDetailByServiceInstanceIDFn(scope.BizID, scope.InstanceIDs)

	detailResult, err := executor.Execute(nCtx, types.UnlimitedPage(), detailFn)
	if err != nil {
		return nil, fmt.Errorf("failed to list service instance detail: %w", err)
	}

	if len(detailResult.Items) == 0 {
		return []*types.Target{}, nil
	}

	hostIDs := conv.SliceUnique(conv.SliceToSlice(detailResult.Items, func(detail *ServiceInstanceDetailInfo) int64 {
		return detail.BKHostID
	}))

	hostCond := &types.HostStaticExactCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: hostIDs,
		},
	}

	if scope.BizID != CCNoBusinessID {
		hostCond.StaticExactInclude.BizID = append(hostCond.StaticExactInclude.BizID, scope.BizID)
	}

	hosts, err := h.FindHostWithCondition(nCtx, types.UnlimitedPage(), hostCond)
	if err != nil {
		return nil, fmt.Errorf("failed to find hosts: %w", err)
	}

	// Convert service instance details and hosts to targets
	targets, err := convServiceInstanceDetailToTarget(detailResult.Items, hosts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert service instance detail to target: %w", err)
	}

	return targets, nil
}

func (h *Handler) getHostTargetByScopeInstance(nCtx contextx.IContext, scope *types.ScopeInstance) ([]*types.Target, error) {
	cond := &types.HostStaticExactCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: scope.InstanceIDs,
		},
	}

	if scope.BizID != CCNoBusinessID {
		cond.StaticExactInclude.BizID = append(cond.StaticExactInclude.BizID, scope.BizID)
	}

	hosts, err := h.FindHostWithCondition(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by scope instance: %w", err)
	}

	targets := convHostToTarget(hosts)

	return targets, nil
}

// GetTargetByScopeTopo get target by scope topo.
func (h *Handler) GetTargetByScopeTopo(nCtx contextx.IContext, scope *types.ScopeTopo) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope topo, nCtx is nil")
	}

	// TODO: 考虑此处的权限代持问题
	nCtx = contextx.From(nCtx, contextx.WithBKUsername(h.cli.config.VirtualUser))

	if scope == nil {
		return nil, fmt.Errorf("failed to get target by scope topo, scope is nil")
	}

	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("failed to get target by scope topo: %w", err)
	}

	var (
		targets []*types.Target
		err     error
	)
	switch scope.Granularity {
	case types.TargetGranularityHost:
		targets, err = h.getHostTargetByScopeTopo(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope topo: %w", err)
		}
	case types.TargetGranularityServiceInstance:
		targets, err = h.getServiceTargetByScopeTopo(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope topo: %w", err)
		}
	default:
		return nil, fmt.Errorf("failed to get target by scope topo, granularity(%s), not supported", scope.Granularity)
	}

	return targets, nil
}

func (h *Handler) getHostTargetByScopeTopo(nCtx contextx.IContext, scope *types.ScopeTopo) ([]*types.Target, error) {
	hosts, err := h.FindHostByTopo(nCtx, scope.BizID, scope.Paths...)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by scope topo: %w", err)
	}

	targets := convHostToTarget(hosts)

	return targets, nil
}

// nolint: gocognit
func (h *Handler) getServiceTargetByScopeTopo(nCtx contextx.IContext, scope *types.ScopeTopo) ([]*types.Target, error) {
	executor := pageexecutor.NewPageExecutor[*ServiceInstanceDetailInfo](CCPageSizeLimit, ccQueryTimeout)

	// Separate module nodes from other nodes
	moduleNodes := make([]*types.ScopeTopoNode, 0)
	otherNodes := make([]*types.ScopeTopoNode, 0)
	for _, path := range scope.Paths {
		if path.TopoObjID == TopoNodeObjIDModule {
			moduleNodes = append(moduleNodes, path)
		} else {
			otherNodes = append(otherNodes, path)
		}
	}

	var allServiceInstances []*ServiceInstanceDetailInfo
	var allHosts []*types.Host

	// Strategy 1: Query service instances directly by module ID (more efficient)
	if len(moduleNodes) > 0 {
		moduleIDs := conv.SliceToSlice(moduleNodes, func(node *types.ScopeTopoNode) int64 {
			return node.TopoInstID
		})
		moduleIDs = conv.SliceUnique(moduleIDs)

		// Query each module separately
		for _, moduleID := range moduleIDs {
			detailFn := h.buildListServiceInstanceDetailByModuleIDFn(scope.BizID, moduleID)
			result, err := executor.Execute(nCtx, types.UnlimitedPage(), detailFn)
			if err != nil {
				return nil, fmt.Errorf("failed to query service instances by module ID %d: %w", moduleID, err)
			}
			allServiceInstances = append(allServiceInstances, result.Items...)
		}

		// Get hosts for these service instances
		if len(allServiceInstances) > 0 {
			hostIDs := conv.SliceUnique(conv.SliceToSlice(allServiceInstances, func(inst *ServiceInstanceDetailInfo) int64 {
				return inst.BKHostID
			}))
			hostCond := &types.HostStaticExactCondition{
				StaticExactInclude: &types.HostStaticExactFields{
					HostID: hostIDs,
				},
			}
			hosts, err := h.FindHostWithCondition(nCtx, types.UnlimitedPage(), hostCond)
			if err != nil {
				return nil, fmt.Errorf("failed to find hosts: %w", err)
			}
			allHosts = append(allHosts, hosts...)
		}
	}

	// Strategy 2: Query service instances by host IDs (for non-module nodes)
	if len(otherNodes) > 0 {
		hosts, err := h.FindHostByTopo(nCtx, scope.BizID, otherNodes...)
		if err != nil {
			return nil, fmt.Errorf("failed to find hosts by topo: %w", err)
		}

		if len(hosts) > 0 {
			hostIDs := conv.SliceUnique(conv.SliceToSlice(hosts, func(host *types.Host) int64 {
				return host.HostID
			}))

			detailFn := h.buildListServiceInstanceDetailByHostFn(scope.BizID, hostIDs)
			result, err := executor.Execute(nCtx, types.UnlimitedPage(), detailFn)
			if err != nil {
				return nil, fmt.Errorf("failed to list service instance detail: %w", err)
			}
			allServiceInstances = append(allServiceInstances, result.Items...)
			allHosts = append(allHosts, hosts...)
		}
	}

	if len(allServiceInstances) == 0 {
		return []*types.Target{}, nil
	}

	// Deduplicate hosts by host ID using conv package
	// Extract unique host IDs first
	uniqueHostIDs := conv.SliceUnique(conv.SliceToSlice(allHosts, func(host *types.Host) int64 {
		return host.HostID
	}))

	// Build host map for fast lookup (keep first occurrence)
	hostMap := make(map[int64]*types.Host, len(uniqueHostIDs))
	for _, host := range allHosts {
		if _, exists := hostMap[host.HostID]; !exists {
			hostMap[host.HostID] = host
		}
	}

	// Filter hosts by unique host IDs using conv package
	uniqueHosts := conv.SliceToSlice(uniqueHostIDs, func(hostID int64) *types.Host {
		return hostMap[hostID]
	})

	// Convert service instance details and hosts to targets
	targets, err := convServiceInstanceDetailToTarget(allServiceInstances, uniqueHosts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert service instance detail to target: %w", err)
	}

	return targets, nil
}

// GetTargetByScopeDynamicGroup get target by scope dynamic group.
func (h *Handler) GetTargetByScopeDynamicGroup(nCtx contextx.IContext, scope *types.ScopeDynamicGroup) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope dynamic group, nCtx is nil")
	}

	// TODO: 考虑此处的权限代持问题
	nCtx = contextx.From(nCtx, contextx.WithBKUsername(h.cli.config.VirtualUser))

	if scope == nil {
		return nil, fmt.Errorf("failed to get target by scope dynamic group, scope is nil")
	}

	if err := scope.Validate(); err != nil {
		return nil, fmt.Errorf("failed to get target by scope dynamic group: %w", err)
	}

	var (
		targets []*types.Target
		err     error
	)

	switch scope.Granularity {
	case types.TargetGranularityHost:
		targets, err = h.getHostTargetByScopeDynamicGroup(nCtx, scope)
		if err != nil {
			return nil, fmt.Errorf("failed to get target by scope dynamic group: %w", err)
		}
	default:
		return nil, fmt.Errorf("failed to get target by scope dynamic group, granularity(%s), not supported", scope.Granularity)
	}

	return targets, nil
}

func (h *Handler) getHostTargetByScopeDynamicGroup(nCtx contextx.IContext, scope *types.ScopeDynamicGroup) ([]*types.Target, error) {
	hosts, err := h.FindHostByDynamicGroup(nCtx, scope.BizID, scope.DynamicGroupIDs, types.UnlimitedPage())
	if err != nil {
		return nil, fmt.Errorf("failed to get target by scope dynamic group: %w", err)
	}

	targets := convHostToTarget(hosts)

	return targets, nil
}

func (h *Handler) buildListServiceInstanceDetailByHostFn(bizID int64, hostIDs []int64,
) func(contextx.IContext, types.Page) ([]*ServiceInstanceDetailInfo, error) {

	return func(nCtx contextx.IContext, p types.Page) ([]*ServiceInstanceDetailInfo, error) {
		req := &ListServiceInstanceDetailReq{
			BKBizID:    bizID,
			BKHostList: hostIDs,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}

		detailResp, err := h.cli.listServiceInstanceDetail(nCtx, req)
		if err != nil {
			return nil, err
		}

		return detailResp.Info, nil
	}
}

func (h *Handler) buildListServiceInstanceDetailByServiceInstanceIDFn(bizID int64, serviceInstanceIDs []int64,
) func(contextx.IContext, types.Page) ([]*ServiceInstanceDetailInfo, error) {

	return func(nCtx contextx.IContext, p types.Page) ([]*ServiceInstanceDetailInfo, error) {
		req := &ListServiceInstanceDetailReq{
			BKBizID:            bizID,
			ServiceInstanceIDs: serviceInstanceIDs,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}

		detailResp, err := h.cli.listServiceInstanceDetail(nCtx, req)
		if err != nil {
			return nil, err
		}

		return detailResp.Info, nil
	}
}

// buildListServiceInstanceDetailByModuleIDFn builds a function to query service instances by module ID.
func (h *Handler) buildListServiceInstanceDetailByModuleIDFn(bizID int64, moduleID int64,
) func(contextx.IContext, types.Page) ([]*ServiceInstanceDetailInfo, error) {

	return func(nCtx contextx.IContext, p types.Page) ([]*ServiceInstanceDetailInfo, error) {
		req := &ListServiceInstanceDetailReq{
			BKBizID:    bizID,
			BKModuleID: moduleID,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}

		detailResp, err := h.cli.listServiceInstanceDetail(nCtx, req)
		if err != nil {
			return nil, err
		}

		return detailResp.Info, nil
	}
}

// getModuleIDsByServiceTemplateIDs gets module IDs by service template IDs.
func (h *Handler) getModuleIDsByServiceTemplateIDs(nCtx contextx.IContext, bizID int64, serviceTemplateIDs []int64) ([]int64, error) {
	if len(serviceTemplateIDs) == 0 {
		return []int64{}, nil
	}

	// Query all modules in the business.
	allModules, err := h.searchAllModules(nCtx, bizID)
	if err != nil {
		return nil, fmt.Errorf("failed to search all modules: %w", err)
	}

	// Build template ID set for fast lookup
	templateIDSet := make(map[int64]struct{}, len(serviceTemplateIDs))
	for _, id := range serviceTemplateIDs {
		templateIDSet[id] = struct{}{}
	}

	// Filter modules matching service_template_id
	moduleIDs := make([]int64, 0)
	for _, module := range allModules {
		if module.ServiceTemplateID > CCInvalidID {
			if _, exists := templateIDSet[module.ServiceTemplateID]; exists {
				moduleIDs = append(moduleIDs, module.BKModuleID)
			}
		}
	}

	return conv.SliceUnique(moduleIDs), nil
}

// getModuleIDsBySetTemplateIDs gets module IDs by set template IDs.
func (h *Handler) getModuleIDsBySetTemplateIDs(nCtx contextx.IContext, bizID int64, setTemplateIDs []int64) ([]int64, error) {
	if len(setTemplateIDs) == 0 {
		return []int64{}, nil
	}

	allModules, err := h.searchAllModules(nCtx, bizID)
	if err != nil {
		return nil, fmt.Errorf("failed to search all modules: %w", err)
	}

	// Build template ID set for fast lookup
	templateIDSet := make(map[int64]struct{}, len(setTemplateIDs))
	for _, id := range setTemplateIDs {
		templateIDSet[id] = struct{}{}
	}

	// Filter modules matching set_template_id
	moduleIDs := make([]int64, 0)
	for _, module := range allModules {
		if module.SetTemplateID > CCInvalidID {
			if _, exists := templateIDSet[module.SetTemplateID]; exists {
				moduleIDs = append(moduleIDs, module.BKModuleID)
			}
		}
	}

	return conv.SliceUnique(moduleIDs), nil
}

// searchAllModules searches all modules in a business.
func (h *Handler) searchAllModules(nCtx contextx.IContext, bizID int64) ([]*ModuleInfo, error) {
	// Create executors at function start for reuse
	setExecutor := pageexecutor.NewPageExecutor[*SetInfo](CCPageSizeLimit, ccQueryTimeout)

	setFn := func(nCtx contextx.IContext, p types.Page) ([]*SetInfo, error) {
		req := &SearchSetReq{
			BKBizID: bizID,
			Page: Page{
				Start: p.Offset,
				Limit: p.Limit,
				Sort:  p.Sort,
			},
		}
		resp, err := h.cli.searchSet(nCtx, req)
		if err != nil {
			return nil, err
		}

		return resp.Info, nil
	}

	setResult, err := setExecutor.Execute(nCtx, types.UnlimitedPage(), setFn)
	if err != nil {
		return nil, fmt.Errorf("failed to search sets: %w", err)
	}

	if len(setResult.Items) == 0 {
		return []*ModuleInfo{}, nil
	}

	allModules := make([]*ModuleInfo, 0)
	moduleExecutor := pageexecutor.NewPageExecutor[*ModuleInfo](CCPageSizeLimit, ccQueryTimeout)
	for _, set := range setResult.Items {
		moduleFn := func(nCtx contextx.IContext, p types.Page) ([]*ModuleInfo, error) {
			req := &SearchModuleReq{
				BKBizID: bizID,
				BKSetID: set.BKSetID,
				Page: Page{
					Start: p.Offset,
					Limit: p.Limit,
					Sort:  p.Sort,
				},
			}
			resp, err := h.cli.searchModule(nCtx, req)
			if err != nil {
				return nil, err
			}

			return resp.Info, nil
		}

		moduleResult, err := moduleExecutor.Execute(nCtx, types.UnlimitedPage(), moduleFn)
		if err != nil {
			return nil, fmt.Errorf("failed to search modules for set %d: %w", set.BKSetID, err)
		}
		allModules = append(allModules, moduleResult.Items...)
	}

	return allModules, nil
}

func convHostToTarget(hosts []*types.Host) []*types.Target {
	targets := make([]*types.Target, len(hosts))
	for idx := range hosts {
		targets[idx] = &types.Target{
			Host: *hosts[idx],
		}
	}

	return targets
}

func convServiceInstanceDetailToTarget(
	serviceInstanceDetails []*ServiceInstanceDetailInfo,
	hosts []*types.Host,
) ([]*types.Target, error) {

	hostMap, err := conv.SliceToMap(hosts, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build host map: %w", err)
	}

	targets := make([]*types.Target, 0, len(serviceInstanceDetails))
	for _, detail := range serviceInstanceDetails {
		host, ok := hostMap[detail.BKHostID]
		if !ok {
			return nil, fmt.Errorf("host not found for service instance detail, host_id(%d), service_instance_id(%d)",
				detail.BKHostID, detail.ID)
		}

		targets = append(targets, &types.Target{
			Host: *host,
			ServiceInstance: types.ServiceInstance{
				ID:                detail.ID,
				Name:              detail.Name,
				BizID:             detail.BKBizID,
				HostID:            detail.BKHostID,
				ModuleID:          detail.BKModuleID,
				ServiceTemplateID: detail.ServiceTemplateID,
				ServiceCategoryID: detail.ServiceCategoryID,
			},
		})
	}

	return targets, nil
}
