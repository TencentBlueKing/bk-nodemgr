/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IScope define the scope interface
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
	// TODO: implement me.
	panic("implement me.")
}

// GetTargetByScopeServiceTemplate get target by scope service template.
func (h *Handler) GetTargetByScopeServiceTemplate(nCtx contextx.IContext, scope *types.ScopeServiceTemplate) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope service template, nCtx is nil")
	}

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
			return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
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
	// TODO: implement me.
	panic("implement me.")
}

// GetTargetByScopeInstance get target by scope instance.
func (h *Handler) GetTargetByScopeInstance(nCtx contextx.IContext, scope *types.ScopeInstance) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope instance, nCtx is nil")
	}

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
			return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
		}
	default:
		return nil, fmt.Errorf("failed to get target by scope instance, granularity(%s), not supported", scope.Granularity)
	}

	return targets, nil
}

func (h *Handler) getServiceTargetByScopeInstance(nCtx contextx.IContext, scope *types.ScopeInstance) ([]*types.Target, error) {
	// TODO: implement me.
	panic("implement me.")
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
			return nil, fmt.Errorf("failed to get target by scope set template: %w", err)
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

func (h *Handler) getServiceTargetByScopeTopo(nCtx contextx.IContext, scope *types.ScopeTopo) ([]*types.Target, error) {
	// TODO: implement me
	panic("implement me")
}

// GetTargetByScopeDynamicGroup get target by scope dynamic group.
func (h *Handler) GetTargetByScopeDynamicGroup(nCtx contextx.IContext, scope *types.ScopeDynamicGroup) ([]*types.Target, error) {
	if nCtx == nil {
		return nil, fmt.Errorf("failed to get target by scope dynamic group, nCtx is nil")
	}

	if scope == nil {
		return nil, fmt.Errorf("failed to get target by scope dynamic group, scope is nil")
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
	case types.TargetGranularityServiceInstance:
		targets, err = h.getServiceTargetByScopeDynamicGroup(nCtx, scope)
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

func (h *Handler) getServiceTargetByScopeDynamicGroup(nCtx contextx.IContext, scope *types.ScopeDynamicGroup) ([]*types.Target, error) {
	// TODO: implement me
	panic("implement me")
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
