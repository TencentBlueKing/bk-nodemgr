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

package packageevent

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoPackageEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/packageevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// countPackageEvent counts package events by conditions.
func (s *Storage) countPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	opts, err := convertPackageEventConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	return s.daoPackageEvent.Count(nCtx, opts...)
}

// listPackageEvent lists package events by page and conditions.
func (s *Storage) listPackageEvent(nCtx contextx.IContext, page types.Page,
	conditions ...*types.PackageEventCondition) ([]*types.PackageEvent, int64, error) {

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(daoPackageEvent.FieldKeyOperateTime),
		types.WithFieldDesc(daoPackageEvent.FieldKeyEventID))
	opts, err := convertPackageEventConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	return s.daoPackageEvent.List(nCtx, page, opts...)
}

// createManyPackageEvent creates package events.
func (s *Storage) createManyPackageEvent(nCtx contextx.IContext, events ...*types.PackageEvent) error {
	if err := s.daoPackageEvent.CreateMany(nCtx, events...); err != nil {
		return err
	}

	return nil
}

// distinctPackageEvent distincts package event fields.
func (s *Storage) distinctPackageEvent(nCtx contextx.IContext, request types.PackageEventDistinctRequest,
	conditions ...*types.PackageEventCondition) (*types.PackageEventDistinctResult, error) {

	opts, err := convertPackageEventConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}
	data := new(types.PackageEventDistinctResult)
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
	if err := gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

func convertPackageEventConditionsToOptions(conditions ...*types.PackageEventCondition) ([]daoPackageEvent.OptFn, error) {
	opts := make([]daoPackageEvent.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}
		if condition.FuzzyInclude != nil {
			return nil, errors.New("fuzzy include is not supported")
		}
		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
		if condition.OperateTimeRange != nil {
			opts = append(opts, daoPackageEvent.WithOperateTimeRange(*condition.OperateTimeRange))
		}
		if fields := condition.ExactInclude; fields != nil {
			opts = append(opts,
				daoPackageEvent.WithEventType(fields.EventType...),
				daoPackageEvent.WithGeneration(fields.Generation...),
				daoPackageEvent.WithReleaseType(fields.ReleaseType...),
				daoPackageEvent.WithCPUArch(fields.CPUArch...),
				daoPackageEvent.WithOSType(fields.OSType...),
				daoPackageEvent.WithVersion(fields.Version...),
				daoPackageEvent.WithOperator(fields.Operator...),
			)
		}
		if fields := condition.ExactExclude; fields != nil {
			opts = append(opts,
				daoPackageEvent.WithoutEventType(fields.EventType...),
				daoPackageEvent.WithoutGeneration(fields.Generation...),
				daoPackageEvent.WithoutReleaseType(fields.ReleaseType...),
				daoPackageEvent.WithoutCPUArch(fields.CPUArch...),
				daoPackageEvent.WithoutOSType(fields.OSType...),
				daoPackageEvent.WithoutVersion(fields.Version...),
				daoPackageEvent.WithoutOperator(fields.Operator...),
			)
		}
	}

	return opts, nil
}
