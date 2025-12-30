/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dpmgr

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IScopeCalculator define the scope calculator.
type IScopeCalculator interface {
	// Calculate calculate the target.
	Calculate(nCtx contextx.IContext, scopes ...*types.Scope) ([]*types.Target, error)
}

var _ IScopeCalculator = &Calculator{}

// Calculator defines the scope calculator.
type Calculator struct {
	cmdbClient cmdb.IHandler

	calculateConcurrency int
}

// CalculatorConfig defines the config of calculator.
type CalculatorConfig struct {
	CmdbClient           cmdb.IHandler
	CalculateConcurrency int
}

// NewScopeCalculator new scope calculator.
func NewScopeCalculator(conf *CalculatorConfig) *Calculator {
	return &Calculator{
		cmdbClient:           conf.CmdbClient,
		calculateConcurrency: conf.CalculateConcurrency,
	}
}

// Calculate calculate the target.
func (cal *Calculator) Calculate(nCtx contextx.IContext, scopes ...*types.Scope) ([]*types.Target, error) {
	targetsChan := make(chan []*types.Target, len(scopes))

	// 1. convert scopes to target

	gp := gopool.NewPool()
	gp.SetLimit(cal.calculateConcurrency)
	for idx := range scopes {
		scope := scopes[idx]
		fn := func() error {
			var scopeCalFn func(nCtx contextx.IContext, scope *types.Scope) ([]*types.Target, error)
			switch scope.Type() {
			case types.ScopeTypeTopo:
				scopeCalFn = cal.convScopeTopoToTargets
			case types.ScopeTypeServiceTemplate:
				scopeCalFn = cal.convScopeServiceTemplateToTargets
			case types.ScopeTypeSetTemplate:
				scopeCalFn = cal.convScopeSetTemplateToTargets
			case types.ScopeTypeInstance:
				scopeCalFn = cal.convScopeInstanceToTargets
			case types.ScopeTypeDynamicGroup:
				scopeCalFn = cal.convScopeDynamicGroupToTargets
			default:
				return fmt.Errorf("failed to calculate, scope-type(%s) not support", scope.Type())
			}

			targets, err := scopeCalFn(nCtx, scope)
			if err != nil {
				return fmt.Errorf("failed to calculate, scope(%v): %w", scope, err)
			}

			targetsChan <- targets

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to calculate: %w", err)
	}

	close(targetsChan)

	// 2. merge target
	targetMap := make(map[string]*types.Target)
	for targets := range targetsChan {
		for _, target := range targets {
			if _, ok := targetMap[target.UniqueID()]; ok {
				// we trust the some unique id target is same.
				continue
			}

			targetMap[target.UniqueID()] = target
		}
	}

	fullTargets := conv.MapValueToSlice(targetMap)

	return fullTargets, nil
}

// convScopeSetTemplateToTargets conv scope set template to Targets.
func (cal *Calculator) convScopeSetTemplateToTargets(nCtx contextx.IContext, scope *types.Scope) ([]*types.Target, error) {
	if scope.Type() != types.ScopeTypeSetTemplate {
		return nil, fmt.Errorf("failed to conv scope set template to Targets, scope-type(%s)", scope.Type())
	}

	scopeSetTemplate, err := scope.GetSetTemplateScope()
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to scope item, scope(%v): %w", scope, err)
	}

	targets, err := cal.cmdbClient.GetTargetByScopeSetTemplate(nCtx, scopeSetTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by set template, scope(%v): %w", scope, err)
	}

	return targets, nil
}

// convScopeServiceTemplateToTargets conv scope service template to Targets.
func (cal *Calculator) convScopeServiceTemplateToTargets(nCtx contextx.IContext, scope *types.Scope) ([]*types.Target, error) {
	if scope.Type() != types.ScopeTypeServiceTemplate {
		return nil, fmt.Errorf("failed to conv scope service template to Targets, scope-type(%s)", scope.Type())
	}

	scopeServiceTemplate, err := scope.GetServiceTemplateScope()
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to scope item, scope(%v): %w", scope, err)
	}

	targets, err := cal.cmdbClient.GetTargetByScopeServiceTemplate(nCtx, scopeServiceTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by service template, scope(%v): %w", scope, err)
	}

	return targets, nil
}

// convScopeTopoToTargets conv scope topo to Targets.
func (cal *Calculator) convScopeTopoToTargets(nCtx contextx.IContext, scope *types.Scope) ([]*types.Target, error) {
	if scope.Type() != types.ScopeTypeTopo {
		return nil, fmt.Errorf("failed to conv scope topo to Targets, scope-type(%s)", scope.Type())
	}

	scopeTopo, err := scope.GetTopoScope()
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to scope item, scope(%v): %w", scope, err)
	}

	targets, err := cal.cmdbClient.GetTargetByScopeTopo(nCtx, scopeTopo)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by topo, scope(%v): %w", scope, err)
	}

	return targets, nil
}

// convScopeInstanceToTargets conv scope instance to Targets.
func (cal *Calculator) convScopeInstanceToTargets(nCtx contextx.IContext, scope *types.Scope) ([]*types.Target, error) {
	if scope.Type() != types.ScopeTypeInstance {
		return nil, fmt.Errorf("failed to conv scope instance to Targets, scope-type(%s)", scope.Type())
	}

	scopeItemInstance, err := scope.GetInstanceScope()
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to scope item, scope(%v): %w", scope, err)
	}

	targets, err := cal.cmdbClient.GetTargetByScopeInstance(nCtx, scopeItemInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by instance, scope(%v): %w", scope, err)
	}

	return targets, nil
}

// convScopeDynamicGroupToTargets conv scope dynamic group to Targets.
func (cal *Calculator) convScopeDynamicGroupToTargets(nCtx contextx.IContext, scope *types.Scope) ([]*types.Target, error) {
	if scope.Type() != types.ScopeTypeDynamicGroup {
		return nil, fmt.Errorf("failed to conv scope dynamic group to Targets, scope-type(%s)", scope.Type())
	}

	scopeDynamicGroup, err := scope.GetDynamicGroupScope()
	if err != nil {
		return nil, fmt.Errorf("failed to convert scope to scope item, scope(%v): %w", scope, err)
	}

	targets, err := cal.cmdbClient.GetTargetByScopeDynamicGroup(nCtx, scopeDynamicGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get target by dynamic group, scope(%v): %w", scope, err)
	}

	return targets, nil
}
