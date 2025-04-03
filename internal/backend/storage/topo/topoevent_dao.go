/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CountTopoEvent counts topo events.
func (s *Storage) CountTopoEvent(ctx context.Context, conditions ...types.TopoEventCondition) (int64, error) {
	opts, err := convertTopoEventConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	return s.daoTopoEvent.Count(ctx, opts...)
}

// ListTopoEvent lists topo events.
func (s *Storage) ListTopoEvent(ctx context.Context, page types.Page, conditions ...types.TopoEventCondition) (
	[]*types.TopoEvent, int64, error) {

	opts, err := convertTopoEventConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	return s.daoTopoEvent.List(ctx, page, opts...)
}

// CreateManyTopoEvent creates topo events.
func (s *Storage) CreateManyTopoEvent(ctx context.Context, events ...*types.TopoEvent) error {
	return s.daoTopoEvent.CreateMany(ctx, events...)
}

// DistinctTopoEvent distincts topo events.
func (s *Storage) DistinctTopoEvent(
	ctx context.Context, request types.TopoEventDistinctRequest, conditions ...types.TopoEventCondition) (
	*types.TopoEventDistinctResult, error) {

	opts, err := convertTopoEventConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	result := new(types.TopoEventDistinctResult)

	gp := gopool.NewPool()
	if request.Type {
		gp.Go(func() error {
			var err error
			result.Type, err = s.daoTopoEvent.DistinctType(ctx, opts...)

			return err
		})
	}
	if request.NetworkAreaID {
		gp.Go(func() error {
			var err error
			result.NetworkAreaID, err = s.daoTopoEvent.DistinctNetworkAreaID(ctx, opts...)

			return err
		})
	}
	if request.NetworkUnitID {
		gp.Go(func() error {
			var err error
			result.NetworkUnitID, err = s.daoTopoEvent.DistinctNetworkUnitID(ctx, opts...)

			return err
		})
	}
	if request.AccessPointID {
		gp.Go(func() error {
			var err error
			result.AccessPointID, err = s.daoTopoEvent.DistinctAccessPointID(ctx, opts...)

			return err
		})
	}
	if request.Operator {
		gp.Go(func() error {
			var err error
			result.Operator, err = s.daoTopoEvent.DistinctOperator(ctx, opts...)

			return err
		})
	}
	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func convertTopoEventConditionsToOptions(conditions ...types.TopoEventCondition) ([]topoevent.OptFn, error) {
	opts := make([]topoevent.OptFn, 0)
	for _, condition := range conditions {
		opts = append(opts, topoevent.WithOperateTimeRange(condition.OperateTimeRange))

		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					topoevent.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
					topoevent.WithNetworkUnitID(condition.Exact.NetworkUnitID...),
					topoevent.WithAccessPointID(condition.Exact.AccessPointID...),
					topoevent.WithType(condition.Exact.Type...),
					topoevent.WithOperator(condition.Exact.Operator...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					topoevent.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
					topoevent.WithoutNetworkUnitID(condition.Exact.NetworkUnitID...),
					topoevent.WithoutAccessPointID(condition.Exact.AccessPointID...),
					topoevent.WithoutType(condition.Exact.Type...),
					topoevent.WithoutOperator(condition.Exact.Operator...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
		case types.ConditionTypeFuzzyExclude:

		default:
			return nil, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return opts, nil
}
