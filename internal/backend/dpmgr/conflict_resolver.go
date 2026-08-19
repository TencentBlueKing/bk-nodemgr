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

package dpmgr

import (
	"fmt"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IConflictResolver define the conflict resolver.
type IConflictResolver interface {
	ResolveConflict(originalUnits []*DeployUnit) ([]*DeployUnit, error)
}

var _ IConflictResolver = &ConflictResolver{}

// ConflictResolver the conflict resolver.
type ConflictResolver struct {
}

// NewConflictResolver create a new conflict resolver.
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{}
}

// ResolveConflict resolve the conflict.
func (resolver *ConflictResolver) ResolveConflict(originalUnits []*DeployUnit) ([]*DeployUnit, error) {
	// sort by create time
	sort.Slice(originalUnits, func(i, j int) bool {
		return originalUnits[i].LifeCycle.CreatedAt.Before(originalUnits[j].LifeCycle.CreatedAt)
	})

	resultUnits := make([]*DeployUnit, len(originalUnits))
	conflictGroupMap := map[string]*DeployUnitGroup{}

	for idx := range originalUnits {
		currentUnit := originalUnits[idx]
		availableTargets := currentUnit.Targets

		for _, spec := range currentUnit.Specs {
			groupID, err := spec.UniqueID()
			if err != nil {
				return nil, fmt.Errorf("failed to get group id: %w", err)
			}

			conflictGroup, ok := conflictGroupMap[groupID]
			if !ok {
				conflictGroup = &DeployUnitGroup{
					groupID:  groupID,
					specType: spec.Type(),
					items:    make([]*DeployUnitGroupItem, 0),
				}
			}

			availableTargets, err = getAvailableTargets(conflictGroup, spec, availableTargets)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve conflict: %w", err)
			}

			conflictGroup.items = append(conflictGroup.items, &DeployUnitGroupItem{
				targets: availableTargets,
				spec:    spec,
			})

			conflictGroupMap[groupID] = conflictGroup
		}

		resultUnit := &DeployUnit{
			DeployPolicyID: currentUnit.DeployPolicyID,
			LifeCycle:      currentUnit.LifeCycle,
			Targets:        availableTargets,
			Specs:          currentUnit.Specs,
		}

		resultUnits[idx] = resultUnit
	}

	return resultUnits, nil
}

// getAvailableTargets get the available Targets from conflict group.
func getAvailableTargets(conflictGroup *DeployUnitGroup, spec *types.DeploySpec, targets []*types.Target) ([]*types.Target, error) {
	availableTargetMap, err := conv.SliceToMap(targets, func(target *types.Target) string {
		return target.UniqueID()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert slice to map: %w", err)
	}

	for idx := range conflictGroup.items {
		item := conflictGroup.items[idx]
		if !item.spec.Type().IsConflict(spec.Type()) {
			// no conflict, we can use this target directly.
			continue
		}

		for _, target := range item.targets {
			// if this target already exist, we think this target not available, so we delete it.
			delete(availableTargetMap, target.UniqueID())
		}
	}

	availableTargets := conv.MapValueToSlice(availableTargetMap)

	return availableTargets, nil
}

// DeployUnitGroup defines the deploy unit group.
type DeployUnitGroup struct {
	groupID  string
	specType types.DeploySpecType
	items    []*DeployUnitGroupItem
}

// DeployUnitGroupItem defines the deploy unit group item.
type DeployUnitGroupItem struct {
	targets []*types.Target
	spec    *types.DeploySpec
}
