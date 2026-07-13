/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoDeployPolicy "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createDeployPolicy(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error) {
	if nCtx == nil {
		return -1, base.ErrInvalidContext()
	}

	if deployPolicy == nil {
		return -1, base.ErrInvalidParam(errors.New("deploy policy is nil"))
	}

	deployPolicyID, err := s.daoDeployPolicy.Create(nCtx, deployPolicy)
	if err != nil {
		return -1, fmt.Errorf("failed to create deploy policy: %w", err)
	}

	return deployPolicyID, nil
}

func (s *Storage) listDeployPolicies(nCtx contextx.IContext, page types.Page, condition *types.DeployPolicyCondition) (
	[]*types.DeployPolicy, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	optFns := make([]daoDeployPolicy.OptFn, 0)
	if condition != nil {
		optFns = append(optFns, convDeployPolicyConditionsToOptions(condition)...)
	}

	deployPolicies, total, err := s.daoDeployPolicy.List(nCtx, page, optFns...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list deploy policies: %w", err)
	}

	return deployPolicies, total, nil
}

// listDeployPoliciesWithoutCount lists deploy policies without count.
func (s *Storage) listDeployPoliciesWithoutCount(nCtx contextx.IContext, page types.Page, condition *types.DeployPolicyCondition) (
	[]*types.DeployPolicy, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	optFns := make([]daoDeployPolicy.OptFn, 0)
	if condition != nil {
		optFns = append(optFns, convDeployPolicyConditionsToOptions(condition)...)
	}

	deployPolicies, err := s.daoDeployPolicy.ListWithoutCount(nCtx, page, optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deploy policies: %w", err)
	}

	return deployPolicies, nil
}

func (s *Storage) getDeployPolicyByID(nCtx contextx.IContext, deployPolicyID int64) (*types.DeployPolicy, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	deployPolicy, err := s.daoDeployPolicy.Get(nCtx, daoDeployPolicy.WithDeployPolicyID(deployPolicyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy policy: %w", err)
	}

	return deployPolicy, nil
}

func (s *Storage) updateDeployPolicyFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(deployPolicy) == 0 {
		return base.ErrInvalidParam(errors.New("deploy policy list is empty"))
	}

	if err := s.daoDeployPolicy.UpdateFields(nCtx, fields, deployPolicy...); err != nil {
		return fmt.Errorf("failed to update deploy policy fields: %w", err)
	}

	return nil
}

func (s *Storage) deleteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := s.daoDeployPolicy.Delete(nCtx, deployPolicyID); err != nil {
		return fmt.Errorf("failed to delete deploy policy: %w", err)
	}

	return nil
}

func (s *Storage) existDeployPolicy(nCtx contextx.IContext, condition *types.DeployPolicyCondition) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	optFns := make([]daoDeployPolicy.OptFn, 0)
	if condition != nil {
		optFns = append(optFns, convDeployPolicyConditionsToOptions(condition)...)
	}

	exist, err := s.daoDeployPolicy.Exist(nCtx, optFns...)
	if err != nil {
		return false, fmt.Errorf("failed to check deploy policy exist: %w", err)
	}

	return exist, nil
}

// convDeployPolicyConditionsToOptions converts deploy policy conditions to options.
func convDeployPolicyConditionsToOptions(condition *types.DeployPolicyCondition) []daoDeployPolicy.OptFn {
	opts := make([]daoDeployPolicy.OptFn, 0)

	if condition == nil {
		return opts
	}

	if condition.ExecutedTimeRange != nil {
		opts = append(opts, daoDeployPolicy.WithExecutedAtTimeRange(*condition.ExecutedTimeRange))
	}

	if condition.ExactInclude != nil {
		opts = append(opts, daoDeployPolicy.WithDeployPolicyID(condition.ExactInclude.DeployPolicyID...))
		opts = append(opts, daoDeployPolicy.WithDsuID(condition.ExactInclude.DsuID...))
		opts = append(opts, daoDeployPolicy.WithEnabled(condition.ExactInclude.Enabled...))
		opts = append(opts, daoDeployPolicy.WithMetaName(condition.ExactInclude.DeployPolicyName...))
		opts = append(opts, daoDeployPolicy.WithOperator(condition.ExactInclude.Operator...))
	}

	if condition.ExactExclude != nil {
		opts = append(opts, daoDeployPolicy.WithoutDeployPolicyID(condition.ExactExclude.DeployPolicyID...))
		opts = append(opts, daoDeployPolicy.WithoutDsuID(condition.ExactExclude.DsuID...))
		opts = append(opts, daoDeployPolicy.WithoutEnabled(condition.ExactExclude.Enabled...))
		opts = append(opts, daoDeployPolicy.WithoutMetaName(condition.ExactExclude.DeployPolicyName...))
		opts = append(opts, daoDeployPolicy.WithoutOperator(condition.ExactExclude.Operator...))
	}

	if condition.FuzzyInclude != nil {
		opts = append(opts, daoDeployPolicy.WithFuzzyMetaName(condition.FuzzyInclude.DeployPolicyName...))
		opts = append(opts, daoDeployPolicy.WithFuzzyOperator(condition.FuzzyInclude.Operator...))
	}

	if condition.FuzzyExclude != nil {
		opts = append(opts, daoDeployPolicy.WithoutFuzzyMetaName(condition.FuzzyExclude.DeployPolicyName...))
		opts = append(opts, daoDeployPolicy.WithoutFuzzyOperator(condition.FuzzyExclude.Operator...))
	}

	return opts
}
