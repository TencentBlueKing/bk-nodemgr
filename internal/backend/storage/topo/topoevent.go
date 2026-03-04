/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology storage for backend.
package topo

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CountTopoEvent counts topo events.
func (s *Storage) CountTopoEvent(nCtx contextx.IContext, conditions ...*types.TopoEventCondition) (int64, error) {
	var (
		num int64
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCountTopoEvent, func(nCtx contextx.IContext) error {
		var err error
		opts := convertTopoEventConditionsToOptions(conditions...)
		num, err = s.daoTopoEvent.Count(nCtx, opts...)

		return err
	})

	return num, err
}

// ListTopoEvent lists topo events.
func (s *Storage) ListTopoEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.TopoEventCondition) (
	[]*types.TopoEvent, int64, error) {

	var (
		results []*types.TopoEvent
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListTopoEvent, func(nCtx contextx.IContext) error {
		var err error
		opts := convertTopoEventConditionsToOptions(conditions...)
		results, num, err = s.daoTopoEvent.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// CreateManyTopoEvent creates topo events.
func (s *Storage) CreateManyTopoEvent(nCtx contextx.IContext, events ...*types.TopoEvent) error {
	return s.WrapFn(nCtx, metricOperationCreateManyTopoEvent, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoTopoEvent.CreateMany(nCtx, events...)

		return err
	})
}

// DistinctTopoEvent distincts topo events.
func (s *Storage) DistinctTopoEvent(
	nCtx contextx.IContext, request types.TopoEventDistinctRequest, conditions ...*types.TopoEventCondition) (
	*types.TopoEventDistinctResult, error) {

	var (
		data *types.TopoEventDistinctResult
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctTopoEvent, func(nCtx contextx.IContext) error {
		opts := convertTopoEventConditionsToOptions(conditions...)
		data = new(types.TopoEventDistinctResult)

		gp := gopool.NewPool()
		if request.Type {
			gp.Go(func() error {
				var err error
				data.Type, err = s.daoTopoEvent.DistinctType(nCtx, opts...)

				return err
			})
		}
		if request.NetworkAreaID {
			gp.Go(func() error {
				var err error
				data.NetworkAreaID, err = s.daoTopoEvent.DistinctNetworkAreaID(nCtx, opts...)

				return err
			})
		}
		if request.NetworkUnitID {
			gp.Go(func() error {
				var err error
				data.NetworkUnitID, err = s.daoTopoEvent.DistinctNetworkUnitID(nCtx, opts...)

				return err
			})
		}
		if request.AccessPointID {
			gp.Go(func() error {
				var err error
				data.AccessPointID, err = s.daoTopoEvent.DistinctAccessPointID(nCtx, opts...)

				return err
			})
		}
		if request.Operator {
			gp.Go(func() error {
				var err error
				data.Operator, err = s.daoTopoEvent.DistinctOperator(nCtx, opts...)

				return err
			})
		}

		return gp.Wait()
	})

	return data, err
}

func convertTopoEventConditionsToOptions(conditions ...*types.TopoEventCondition) []topoevent.OptFn {
	opts := make([]topoevent.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, topoevent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				topoevent.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				topoevent.WithNetworkUnitID(condition.ExactInclude.NetworkUnitID...),
				topoevent.WithAccessPointID(condition.ExactInclude.AccessPointID...),
				topoevent.WithType(condition.ExactInclude.Type...),
				topoevent.WithOperator(condition.ExactInclude.Operator...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				topoevent.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
				topoevent.WithoutNetworkUnitID(condition.ExactExclude.NetworkUnitID...),
				topoevent.WithoutAccessPointID(condition.ExactExclude.AccessPointID...),
				topoevent.WithoutType(condition.ExactExclude.Type...),
				topoevent.WithoutOperator(condition.ExactExclude.Operator...),
			)
		}
		if condition.FuzzyInclude != nil {
			opts = append(opts,
				topoevent.WithFuzzyNetworkAreaName(condition.FuzzyInclude.NetworkAreaName...),
				topoevent.WithFuzzyNetworkUnitName(condition.FuzzyInclude.NetworkUnitName...),
			)
		}

		if condition.FuzzyExclude != nil {
			opts = append(opts,
				topoevent.WithoutFuzzyNetworkAreaName(condition.FuzzyExclude.NetworkAreaName...),
				topoevent.WithoutFuzzyNetworkUnitName(condition.FuzzyExclude.NetworkUnitName...),
			)
		}
	}

	return opts
}
