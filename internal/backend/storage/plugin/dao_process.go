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
	daoProcess "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createProcess(nCtx contextx.IContext, process *types.Process) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	err := s.daoProcess.Create(nCtx, process)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	return nil
}

func (s *Storage) updateProcess(nCtx contextx.IContext, process *types.Process, hostID int64, pluginName string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	err := s.daoProcess.Update(nCtx, hostID, pluginName, process)
	if err != nil {
		return fmt.Errorf("failed to update process: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) updateProcessInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processInfo == nil {
		return base.ErrInvalidParam(fmt.Errorf("process info is nil"))
	}

	err = s.daoProcess.UpdateInfo(nCtx, hostID, pluginName, processInfo)
	if err != nil {
		return fmt.Errorf("failed to update process info: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) deleteProcess(nCtx contextx.IContext, hostID int64, pluginName string) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	err = s.daoProcess.Delete(nCtx, hostID, pluginName)
	if err != nil {
		return fmt.Errorf("failed to delete process: %w", err)
	}

	return nil
}

func (s *Storage) countProcesses(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (int64, error) {
	var opts []daoProcess.OptFn
	var err error

	if opts, err = convertProcessConditionsToOptions(conditions...); err != nil {
		return 0, err
	}

	return s.daoProcess.Count(nCtx, opts...)
}

func (s *Storage) listProcesses(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessCondition) ([]*types.Process, int64, error) {
	var opts []daoProcess.OptFn
	var err error

	if opts, err = convertProcessConditionsToOptions(conditions...); err != nil {
		return nil, 0, err
	}

	return s.daoProcess.List(nCtx, page, opts...)
}

func convertProcessConditionsToOptions(conditions ...*types.ProcessCondition) ([]daoProcess.OptFn, error) {
	opts := make([]daoProcess.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoProcess.WithHostID(condition.ExactInclude.HostID...),
				daoProcess.WithGroup(condition.ExactInclude.PluginGroup...),
				daoProcess.WithGeneration(condition.ExactInclude.NodeGeneration...),
				daoProcess.WithPlatformOS(condition.ExactInclude.PlatformOS...),
				daoProcess.WithPlatformArch(condition.ExactInclude.PlatformArch...),
				daoProcess.WithInfoStatus(condition.ExactInclude.InfoStatus...),
				daoProcess.WithInfoAgentID(condition.ExactInclude.InfoAgentID...),
				daoProcess.WithInfoVersion(condition.ExactInclude.InfoVersion...),
				daoProcess.WithPluginName(condition.ExactInclude.PluginName...),
				daoProcess.WithPkgName(condition.ExactInclude.PluginPkgName...))
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				daoProcess.WithFuzzyName(condition.FuzzyInclude.Name...),
				daoProcess.WithFuzzyPkgName(condition.FuzzyInclude.PkgName...))
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

// nolint: nonamedreturns
func (s *Storage) existProcess(nCtx contextx.IContext, hostID int64, pluginName string) (exist bool, err error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	exist, err = s.daoProcess.Exist(nCtx, hostID, pluginName)
	if err != nil {
		return false, fmt.Errorf("failed to count process: %w", err)
	}

	return exist, nil
}

// nolint: nonamedreturns
func (s *Storage) updateManyProcessInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processInfoDeltas) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("process info map is empty"))
	}

	err = s.daoProcess.UpdateManyInfo(nCtx, processInfoDeltas)
	if err != nil {
		return fmt.Errorf("failed to batch update process info: %w", err)
	}

	return nil
}

func (s *Storage) getProcess(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	process, err := s.daoProcess.Get(nCtx, daoProcess.WithHostID(hostID), daoProcess.WithPluginName(pluginName))
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	return process, nil
}

func (s *Storage) getProcessDistributionByHostID(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	opts, err := convertProcessConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	dist, err := s.daoProcess.GetProcessDistributionByHostID(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get process distribution: %w", err)
	}

	return dist, nil
}
func (s *Storage) getProcessDistributionByPluginName(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (map[string]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}
	opts, err := convertProcessConditionsToOptions(conditions...)
	if err != nil {
		return nil, err
	}

	dist, err := s.daoProcess.GetProcessDistributionByPluginName(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get process distribution: %w", err)
	}

	return dist, nil
}
