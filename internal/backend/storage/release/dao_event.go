/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package release provides the release storage interface.
// nolint: nonamedreturns
package release

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoPackageEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// countPakcageEvent counts package events.
func (s *Storage) countPakcageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	opts := convertPackageEventConditionsToOptions(conditions...)

	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	num, err := s.daoPackageEvent.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count package event: %w", err)
	}

	return num, nil
}

// listPackageEvent lists package events.
func (s *Storage) listPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) (
	[]*types.PackageEvent, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts := convertPackageEventConditionsToOptions(conditions...)

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(daoPackageEvent.FieldKeyOperateTime))

	events, num, err := s.daoPackageEvent.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list package event: %w", err)
	}

	return events, num, nil
}

// listPackageEventWithoutCount lists package events without count.
func (s *Storage) listPackageEventWithoutCount(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) (
	[]*types.PackageEvent, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	opts := convertPackageEventConditionsToOptions(conditions...)

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(daoPackageEvent.FieldKeyOperateTime))

	events, err := s.daoPackageEvent.ListWithoutCount(nCtx, page, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list package event: %w", err)
	}

	return events, nil
}

func (s *Storage) createManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if events == nil {
		return base.ErrInvalidParam(fmt.Errorf("events is nil"))
	}

	if err := s.daoPackageEvent.CreateMany(nCtx, events...); err != nil {
		return fmt.Errorf("failed to create many package event: %w", err)
	}

	return nil
}

func (s *Storage) distinctPackageEvent(
	nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
	data *types.PackageEventDistinctResult, err error) {

	opts := convertPackageEventConditionsToOptions(conditions...)

	data = new(types.PackageEventDistinctResult)

	gp := gopool.NewPool()
	if request.ReleaseType {
		gp.Go(func() error {
			var err error
			data.ReleaseType, err = s.daoPackageEvent.DistinctReleaseType(nCtx, opts...)

			return err
		})
	}

	if request.EventType {
		gp.Go(func() error {
			var err error
			data.EventType, err = s.daoPackageEvent.DistinctEventType(nCtx, opts...)

			return err
		})
	}

	if request.OSType {
		gp.Go(func() error {
			var err error
			data.OSType, err = s.daoPackageEvent.DistinctOsType(nCtx, opts...)

			return err
		})
	}

	if request.CPUArch {
		gp.Go(func() error {
			var err error
			data.CPUArch, err = s.daoPackageEvent.DistinctCPUArch(nCtx, opts...)

			return err
		})
	}

	if request.Version {
		gp.Go(func() error {
			var err error
			data.Version, err = s.daoPackageEvent.DistinctVersion(nCtx, opts...)

			return err
		})
	}

	if request.Operator {
		gp.Go(func() error {
			var err error
			data.Operator, err = s.daoPackageEvent.DistinctOperator(nCtx, opts...)

			return err
		})
	}
	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

func convertPackageEventConditionsToOptions(conditions ...*types.PackageEventCondition) []daoPackageEvent.OptFn {
	opts := make([]daoPackageEvent.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.OperateTimeRange != nil {
			opts = append(opts, daoPackageEvent.WithOperateTimeRange(*condition.OperateTimeRange))
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoPackageEvent.WithEventType(condition.ExactInclude.EventType...),
				daoPackageEvent.WithGeneration(condition.ExactInclude.Generation...),
				daoPackageEvent.WithReleaseType(condition.ExactInclude.ReleaseType...),
				daoPackageEvent.WithCPUArch(condition.ExactInclude.CPUArch...),
				daoPackageEvent.WithOSType(condition.ExactInclude.OSType...),
				daoPackageEvent.WithVersion(condition.ExactInclude.Version...),
				daoPackageEvent.WithOperator(condition.ExactInclude.Operator...),
				daoPackageEvent.WithVersion(condition.ExactInclude.Version...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				daoPackageEvent.WithEventType(condition.ExactInclude.EventType...),
				daoPackageEvent.WithGeneration(condition.ExactInclude.Generation...),
				daoPackageEvent.WithReleaseType(condition.ExactInclude.ReleaseType...),
				daoPackageEvent.WithCPUArch(condition.ExactInclude.CPUArch...),
				daoPackageEvent.WithOSType(condition.ExactInclude.OSType...),
				daoPackageEvent.WithVersion(condition.ExactInclude.Version...),
				daoPackageEvent.WithOperator(condition.ExactInclude.Operator...),
				daoPackageEvent.WithVersion(condition.ExactInclude.Version...),
			)
		}
	}

	return opts
}
