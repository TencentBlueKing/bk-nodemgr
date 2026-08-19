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

// Package configpolicy implements the dao for config policy events.
// nolint: nonamedreturns
package configpolicy

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoConfigPolicyEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy-event"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// countConfigPolicyEvent counts policy events.
func (s *Storage) countConfigPolicyEvent(nCtx contextx.IContext, conditions ...*types.ConfigPolicyEventCondition) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	opts, err := convertConfigPolicyEventConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	num, err := s.daoConfigPolicyEvent.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count config policy event: %w", err)
	}

	return num, nil
}

// listConfigPolicyEvent lists policy events.
func (s *Storage) listConfigPolicyEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyEventCondition) (
	[]*types.ConfigPolicyEvent, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts, err := convertConfigPolicyEventConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(daoConfigPolicyEvent.FieldKeyOperateTime))

	events, num, err := s.daoConfigPolicyEvent.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list config policy event: %w", err)
	}

	return events, num, nil
}

func (s *Storage) createManyConfigPolicyEvent(nCtx contextx.IContext, events ...*types.ConfigPolicyEvent) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if events == nil {
		return base.ErrInvalidParam(fmt.Errorf("events is nil"))
	}

	if err := s.daoConfigPolicyEvent.CreateMany(nCtx, events...); err != nil {
		return fmt.Errorf("failed to create many config policy event: %w", err)
	}

	return nil
}

func (s *Storage) distinctConfigPolicyEvent(
	nCtx contextx.IContext, request types.ConfigPolicyEventDistinctRequest, conditions ...*types.ConfigPolicyEventCondition) (
	data *types.ConfigPolicyEventDistinctResult, err error) {

	opts, err := convertConfigPolicyEventConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	data = new(types.ConfigPolicyEventDistinctResult)

	gp := gopool.NewPool()
	if request.BizID {
		gp.Go(func() error {
			var err error
			data.BizID, err = s.daoConfigPolicyEvent.DistinctBizID(nCtx, opts...)

			return err
		})
	}

	if request.ConfigPolicyType {
		gp.Go(func() error {
			var err error
			data.ConfigPolicyType, err = s.daoConfigPolicyEvent.DistinctConfigPolicyType(nCtx, opts...)

			return err
		})
	}

	if request.Type {
		gp.Go(func() error {
			var err error
			data.Type, err = s.daoConfigPolicyEvent.DistinctEventType(nCtx, opts...)

			return err
		})
	}

	if request.Version {
		gp.Go(func() error {
			var err error
			data.Version, err = s.daoConfigPolicyEvent.DistinctVersion(nCtx, opts...)

			return err
		})
	}

	if request.Operator {
		gp.Go(func() error {
			var err error
			data.Operator, err = s.daoConfigPolicyEvent.DistinctOperator(nCtx, opts...)

			return err
		})
	}

	if request.ConfigPolicyID {
		gp.Go(func() error {
			var err error
			data.ConfigPolicyID, err = s.daoConfigPolicyEvent.DistinctConfigPolicyID(nCtx, opts...)

			return err
		})
	}

	if request.ConfigPolicyName {
		gp.Go(func() error {
			var err error
			data.ConfigPolicyName, err = s.daoConfigPolicyEvent.DistinctConfigPolicyName(nCtx, opts...)

			return err
		})
	}
	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

func convertConfigPolicyEventConditionsToOptions(conditions ...*types.ConfigPolicyEventCondition) ([]daoConfigPolicyEvent.OptFn, error) {
	opts := make([]daoConfigPolicyEvent.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, daoConfigPolicyEvent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoConfigPolicyEvent.WithBizID(condition.ExactInclude.BizID...),
				daoConfigPolicyEvent.WithType(condition.ExactInclude.Type...),
				daoConfigPolicyEvent.WithVersion(condition.ExactInclude.Version...),
				daoConfigPolicyEvent.WithConfigPolicyType(condition.ExactInclude.ConfigPolicyType...),
				daoConfigPolicyEvent.WithConfigPolicyID(condition.ExactInclude.ConfigPolicyID...),
			)
		}

		if condition.ExactExclude != nil {
			return nil, errors.New("exact exclude is not supported")
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				daoConfigPolicyEvent.WithConfigPolicyName(condition.FuzzyInclude.ConfigPolicyName...),
				daoConfigPolicyEvent.WithOperator(condition.FuzzyInclude.Operator...),
			)
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
