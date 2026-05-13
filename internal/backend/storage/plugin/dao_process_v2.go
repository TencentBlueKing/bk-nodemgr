/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoProcessV2 "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process-v2"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createProcessV2(nCtx contextx.IContext, process *types.Process) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	err := s.daoProcessV2.Create(nCtx, process)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	return nil
}

func (s *Storage) updateProcessV2(nCtx contextx.IContext, process *types.Process, hostID int64, pluginName string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	err := s.daoProcessV2.Update(nCtx, hostID, pluginName, process)
	if err != nil {
		return fmt.Errorf("failed to update process: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) updateProcessV2Info(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processInfo == nil {
		return base.ErrInvalidParam(fmt.Errorf("process info is nil"))
	}

	err = s.daoProcessV2.UpdateInfo(nCtx, hostID, pluginName, processInfo)
	if err != nil {
		return fmt.Errorf("failed to update process info: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) updateManyProcessV2Info(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processInfoDeltas) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("process info map is empty"))
	}

	err = s.daoProcessV2.UpdateManyInfo(nCtx, processInfoDeltas)
	if err != nil {
		return fmt.Errorf("failed to batch update process info: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) updateProcessV2ManyHostBizID(nCtx contextx.IContext, bizID int64, hostID ...int64) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	err = s.daoProcessV2.UpdateManyHostBizID(nCtx, bizID, hostID...)
	if err != nil {
		return fmt.Errorf("failed to update process biz id, host-id(%v): %w", hostID, err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) deleteProcessV2(nCtx contextx.IContext, hostID int64, pluginName string) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	err = s.daoProcessV2.Delete(nCtx, hostID, pluginName)
	if err != nil {
		return fmt.Errorf("failed to delete process: %w", err)
	}

	return nil
}

func (s *Storage) countProcessesV2(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (int64, error) {
	var opts []daoProcessV2.OptFn
	var err error

	if opts, err = convertProcessV2ConditionsToOptions(conditions...); err != nil {
		return 0, err
	}

	return s.daoProcessV2.Count(nCtx, opts...)
}

func (s *Storage) listProcessesV2(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessCondition) ([]*types.Process, int64, error) {
	var opts []daoProcessV2.OptFn
	var err error

	if opts, err = convertProcessV2ConditionsToOptions(conditions...); err != nil {
		return nil, 0, err
	}

	return s.daoProcessV2.List(nCtx, page, opts...)
}

// nolint: nonamedreturns
func (s *Storage) existProcessV2(nCtx contextx.IContext, hostID int64, pluginName string) (exist bool, err error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	exist, err = s.daoProcessV2.Exist(nCtx, hostID, pluginName)
	if err != nil {
		return false, fmt.Errorf("failed to count process: %w", err)
	}

	return exist, nil
}

func (s *Storage) getProcessV2(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	process, err := s.daoProcessV2.Get(nCtx, daoProcessV2.WithHostID(hostID), daoProcessV2.WithPluginName(pluginName))
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	return process, nil
}

func (s *Storage) getProcessV2DistributionByHostID(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	dist, err := s.daoProcessV2.GetProcessDistributionByHostID(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get process distribution: %w", err)
	}

	return dist, nil
}
func (s *Storage) getProcessV2DistributionByPluginName(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (map[string]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	opts, err := convertProcessV2ConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	dist, err := s.daoProcessV2.GetProcessDistributionByPluginName(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get process distribution: %w", err)
	}

	return dist, nil
}

// distinctProcessV2 distinct process.
func (s *Storage) distinctProcessV2(nCtx contextx.IContext, request types.ProcessDistinctSelector, conditions ...*types.ProcessCondition) (
	*types.ProcessDistinctResult, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	data := new(types.ProcessDistinctResult)
	gp := gopool.NewPool()
	if request.OSType {
		gp.Go(func() error {
			var err error
			data.OSType, err = s.daoProcessV2.DistinctPlatformOS(nCtx, opts...)

			return err
		})
	}
	if request.CPUArch {
		gp.Go(func() error {
			var err error
			data.CPUArch, err = s.daoProcessV2.DistinctCPUArch(nCtx, opts...)

			return err
		})
	}
	if request.Version {
		gp.Go(func() error {
			var err error
			data.Version, err = s.daoProcessV2.DistinctInfoVersion(nCtx, opts...)

			return err
		})
	}
	if request.Status {
		gp.Go(func() error {
			var err error
			data.Status, err = s.daoProcessV2.DistinctInfoStatus(nCtx, opts...)

			return err
		})
	}
	if request.PluginName {
		gp.Go(func() error {
			var err error
			data.PluginName, err = s.daoProcessV2.DistinctPluginName(nCtx, opts...)

			return err
		})
	}
	if request.PluginGroup {
		gp.Go(func() error {
			var err error
			data.PluginGroup, err = s.daoProcessV2.DistinctGroup(nCtx, opts...)

			return err
		})
	}
	if request.PluginPkgName {
		gp.Go(func() error {
			var err error
			data.PluginPkgName, err = s.daoProcessV2.DistinctPkgName(nCtx, opts...)

			return err
		})
	}
	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

func convertProcessV2ConditionsToOptions(conditions ...*types.ProcessCondition) ([]daoProcessV2.OptFn, error) {
	opts := make([]daoProcessV2.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoProcessV2.WithHostID(condition.ExactInclude.HostID...),
				daoProcessV2.WithBizID(condition.ExactInclude.BizID...),
				daoProcessV2.WithGroup(condition.ExactInclude.PluginGroup...),
				daoProcessV2.WithGeneration(condition.ExactInclude.Generation...),
				daoProcessV2.WithPlatformOS(condition.ExactInclude.PlatformOS...),
				daoProcessV2.WithPlatformArch(condition.ExactInclude.PlatformArch...),
				daoProcessV2.WithInfoStatus(condition.ExactInclude.InfoStatus...),
				daoProcessV2.WithInfoAgentID(condition.ExactInclude.InfoAgentID...),
				daoProcessV2.WithInfoVersion(condition.ExactInclude.InfoVersion...),
				daoProcessV2.WithPluginName(condition.ExactInclude.PluginName...),
				daoProcessV2.WithPkgName(condition.ExactInclude.PluginPkgName...))
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				daoProcessV2.WithFuzzyName(condition.FuzzyInclude.Name...),
				daoProcessV2.WithFuzzyPkgName(condition.FuzzyInclude.PkgName...))
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
