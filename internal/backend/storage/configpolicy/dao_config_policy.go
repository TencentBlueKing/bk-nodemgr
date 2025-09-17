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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// matchConfigPolicy matches the config policy.
func (s *Storage) matchConfigPolicy(ctx context.Context,
	bizID, networkAreaID, networkUnitID int64,
	osType criteria.OSType, cpuArch criteria.CPUArch) (*types.ConfigPolicy, bool, error) {

	var results []*types.ConfigPolicy
	var err error

	if results, _, err = s.daoConfigPolicy.List(ctx,
		types.Page{
			Limit: 1,
			Sort:  "-" + configpolicy.FieldKeyUpdatedAt,
		},
		configpolicy.WithEnabledScope(bizID, networkAreaID, networkUnitID, osType, cpuArch),
	); err != nil {
		return nil, false, fmt.Errorf("list config policy failed: %w", err)
	}

	if len(results) == 0 {
		return nil, false, nil
	}

	return results[0], true, nil
}

// countConfigPolicy counts the config policy by conditions.
func (s *Storage) countConfigPolicy(ctx context.Context, conditions ...*types.ConfigPolicyCondition) (int64, error) {
	var opts []configpolicy.OptFn
	var err error

	if opts, err = convertConfigPolicyConditionsToOptions(conditions...); err != nil {
		return 0, err
	}

	return s.daoConfigPolicy.Count(ctx, opts...)
}

// listConfigPolicy lists the config policy by page and conditions.
func (s *Storage) listConfigPolicy(ctx context.Context, page types.Page, conditions ...*types.ConfigPolicyCondition) (
	[]*types.ConfigPolicy, int64, error) {

	var opts []configpolicy.OptFn
	var err error

	if opts, err = convertConfigPolicyConditionsToOptions(conditions...); err != nil {
		return nil, 0, err
	}

	return s.daoConfigPolicy.List(ctx, page, opts...)
}

// getConfigPolicy gets the config policy.
func (s *Storage) getConfigPolicy(ctx context.Context, configPolicyID int64) (*types.ConfigPolicy, error) {
	var data *types.ConfigPolicy
	var err error

	if data, err = s.daoConfigPolicy.Get(ctx, configPolicyID); err != nil {
		return nil, err
	}

	return data, nil
}

// createConfigPolicy creates the config policy.
func (s *Storage) createConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) (int64, error) {
	var configPolicyID int64
	var err error

	configPolicy.UpdatedAt = time.Now()
	if configPolicyID, err = s.daoConfigPolicy.Create(ctx, configPolicy); err != nil {
		return -1, err
	}

	return configPolicyID, nil
}

// updateConfigPolicy updates the config policy.
func (s *Storage) updateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) error {
	var err error

	configPolicy.UpdatedAt = time.Now()
	if err = s.daoConfigPolicy.UpdateMany(ctx, configPolicy); err != nil {
		return err
	}

	return nil
}

// deleteManyConfigPolicy deletes the config policies.
func (s *Storage) deleteManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error {
	var err error

	if err = s.daoConfigPolicy.DeleteMany(ctx, configPolicyIDs...); err != nil {
		return err
	}

	return nil
}

// enableManyConfigPolicy enables the config policies.
func (s *Storage) enableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error {
	var err error

	if err = s.daoConfigPolicy.EnableMany(ctx, configPolicyIDs...); err != nil {
		return err
	}

	return nil
}

// disableManyConfigPolicy disables the config policies.
func (s *Storage) disableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error {
	var err error

	if err = s.daoConfigPolicy.DisableMany(ctx, configPolicyIDs...); err != nil {
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
				configpolicy.WithBizID(condition.ExactInclude.BizID...),
				configpolicy.WithNodeRole(condition.ExactInclude.NodeRole...),
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
