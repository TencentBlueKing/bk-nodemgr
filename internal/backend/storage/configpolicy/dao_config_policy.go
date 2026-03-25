/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicy

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// matchConfigPolicyNode matches enabled policies for the node and returns the cascading-merged result.
func (s *Storage) matchConfigPolicyNode(nCtx contextx.IContext,
	bizID, networkAreaID, networkUnitID int64,
	osType criteria.OSType, cpuArch criteria.CPUArch,
	nodeRole types.NodeRole, hostID int64) (*types.ConfigPolicyMatchResult, error) {

	configpolicyType, err := types.ConvertNodeRoleToConfigPolicyType(nodeRole)
	if err != nil {
		return nil, fmt.Errorf("failed to convert node role to config policy type: %w", err)
	}

	// sort by priority DESC.
	// deepMergeConfig iterates in order and later entries overlay earlier ones,
	// meaning the smallest-numbered (highest-importance) policy wins conflicts.
	page := types.Page{Limit: 0}
	page.Sort = types.WithFieldDesc(configpolicy.FieldKeyPriority)

	opts := []configpolicy.OptFn{
		configpolicy.WithEnabledScope(bizID, networkAreaID, networkUnitID, osType, cpuArch, hostID),
		configpolicy.WithConfigPolicyType(configpolicyType),
	}

	results, _, err := s.daoConfigPolicy.List(nCtx, page, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list config policy: %w", err)
	}

	if len(results) == 0 {
		return &types.ConfigPolicyMatchResult{}, nil
	}

	return buildMatchResult(results), nil
}

func buildMatchResult(policies []*types.ConfigPolicy) *types.ConfigPolicyMatchResult {
	matched := conv.SliceToSlice(policies, func(p *types.ConfigPolicy) types.ConfigPolicyMatchedPolicy {
		return types.ConfigPolicyMatchedPolicy{
			PolicyID:   p.ID,
			PolicyName: p.Name,
			Priority:   p.Priority,
		}
	})

	var merged map[string]any
	for _, p := range policies {
		merged = deepMergeConfig(merged, p.Configs)
	}

	return &types.ConfigPolicyMatchResult{
		MatchedPolicies: matched,
		MergedConfig:    merged,
	}
}

// countConfigPolicy counts the config policy by conditions.
func (s *Storage) countConfigPolicy(nCtx contextx.IContext, conditions ...*types.ConfigPolicyCondition) (int64, error) {
	var opts []configpolicy.OptFn
	var err error

	if opts, err = convertConfigPolicyConditionsToOptions(conditions...); err != nil {
		return 0, err
	}

	return s.daoConfigPolicy.Count(nCtx, opts...)
}

// listConfigPolicy lists the config policy by page and conditions.
func (s *Storage) listConfigPolicy(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyCondition) (
	[]*types.ConfigPolicy, int64, error) {

	// sort by priority ASC.
	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldAsc(configpolicy.FieldKeyPriority))

	var opts []configpolicy.OptFn
	var err error

	if opts, err = convertConfigPolicyConditionsToOptions(conditions...); err != nil {
		return nil, 0, err
	}

	return s.daoConfigPolicy.List(nCtx, page, opts...)
}

// getConfigPolicy gets the config policy.
func (s *Storage) getConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error) {
	var data *types.ConfigPolicy
	var err error

	if data, err = s.daoConfigPolicy.Get(nCtx, configPolicyID); err != nil {
		return nil, err
	}

	return data, nil
}

// createConfigPolicy creates the config policy.
func (s *Storage) createConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	configPolicy.UpdatedAt = time.Now()

	configPolicyID, err := s.daoConfigPolicy.Create(nCtx, configPolicy)
	if err != nil {
		return -1, err
	}

	return configPolicyID, nil
}

// updateConfigPolicy updates the config policy.
func (s *Storage) updateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) error {
	var err error

	configPolicy.UpdatedAt = time.Now()
	if err = s.daoConfigPolicy.UpdateMany(nCtx, configPolicy); err != nil {
		return err
	}

	return nil
}

// deleteManyConfigPolicy deletes the config policies.
func (s *Storage) deleteManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	var err error

	if err = s.daoConfigPolicy.DeleteMany(nCtx, configPolicyIDs...); err != nil {
		return err
	}

	return nil
}

// enableManyConfigPolicy enables the config policies and reassigns global priorities.
func (s *Storage) enableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	var err error

	if err = s.daoConfigPolicy.EnableMany(nCtx, configPolicyIDs...); err != nil {
		return err
	}

	return nil
}

// disableManyConfigPolicy disables the config policies and clears their priorities.
func (s *Storage) disableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	var err error

	if err = s.daoConfigPolicy.DisableMany(nCtx, configPolicyIDs...); err != nil {
		return err
	}

	return nil
}

// updatePriorityManyConfigPolicy batch-updates the priority field for the given policy IDs.
func (s *Storage) updatePriorityManyConfigPolicy(nCtx contextx.IContext, priorities map[int64]int64) error {
	var err error

	if err = s.daoConfigPolicy.UpdatePriorityMany(nCtx, priorities); err != nil {
		return err
	}

	return nil
}

func convertConfigPolicyConditionsToOptions(conditions ...*types.ConfigPolicyCondition) ([]configpolicy.OptFn, error) {
	opts := make([]configpolicy.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				configpolicy.WithConfigPolicyID(condition.ExactInclude.ConfigPolicyID...),
				configpolicy.WithConfigPolicyType(condition.ExactInclude.Type...),
				configpolicy.WithBizID(condition.ExactInclude.BizID...),
				configpolicy.WithEnabled(condition.ExactInclude.Enabled...))
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				configpolicy.WithFuzzyConfigPolicyName(condition.FuzzyInclude.ConfigPolicyName...),
				configpolicy.WithFuzzyOperator(condition.FuzzyInclude.Operator...))
		}

		if condition.ExactExclude != nil {
			return nil, errors.New("exact exclude is not supported")
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
