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

package plugin

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoProcessV2 "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process-v2"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createProcessV2(nCtx contextx.IContext, process *types.Process) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	if err := s.daoProcessV2.Create(nCtx, process); err != nil {
		return fmt.Errorf("failed to create process v2: %w", err)
	}

	return nil
}

func (s *Storage) existProcessV2(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	if pluginName == "" {
		return false, base.ErrInvalidParam(fmt.Errorf("plugin name is empty"))
	}

	exist, err := s.daoProcessV2.Exist(nCtx, hostID, pluginName)
	if err != nil {
		return false, fmt.Errorf("failed to check process v2 exist: %w", err)
	}

	return exist, nil
}

func (s *Storage) getProcessV2(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if pluginName == "" {
		return nil, base.ErrInvalidParam(fmt.Errorf("plugin name is empty"))
	}

	process, err := s.daoProcessV2.Get(nCtx, daoProcessV2.WithHostID(hostID), daoProcessV2.WithPluginName(pluginName))
	if err != nil {
		return nil, fmt.Errorf("failed to get process v2: %w", err)
	}

	return process, nil
}

func (s *Storage) updateProcessV2(nCtx contextx.IContext, process *types.Process) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	if err := s.daoProcessV2.Update(nCtx, process.HostID, process.PluginName, process); err != nil {
		return fmt.Errorf("failed to update process v2: %w", err)
	}

	return nil
}

func (s *Storage) updateProcessInfoV2(
	nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error {

	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processInfo == nil {
		return base.ErrInvalidParam(fmt.Errorf("process info is nil"))
	}

	if pluginName == "" {
		return base.ErrInvalidParam(fmt.Errorf("plugin name is empty"))
	}

	if err := s.daoProcessV2.UpdateInfo(nCtx, hostID, pluginName, processInfo); err != nil {
		return fmt.Errorf("failed to update process v2 info: %w", err)
	}

	return nil
}

func (s *Storage) listProcessesV2(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessCondition) ([]*types.Process, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	processes, total, err := s.daoProcessV2.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list processes v2: %w", err)
	}

	return processes, total, nil
}

func (s *Storage) countProcessesV2(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	count, err := s.daoProcessV2.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count processes v2: %w", err)
	}

	return count, nil
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
				daoProcessV2.WithInfoTrusteeship(condition.ExactInclude.InfoTrusteeship...),
				daoProcessV2.WithInfoAgentID(condition.ExactInclude.InfoAgentID...),
				daoProcessV2.WithInfoVersion(condition.ExactInclude.InfoVersion...),
				daoProcessV2.WithPluginName(condition.ExactInclude.PluginName...),
				daoProcessV2.WithPkgName(condition.ExactInclude.PluginPkgName...))
			if condition.ExactInclude.InfoLastSyncAtBefore != nil {
				opts = append(opts, daoProcessV2.WithInfoLastSyncAtBefore(*condition.ExactInclude.InfoLastSyncAtBefore))
			}
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

func (s *Storage) updateManyProcessInfoV2(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processInfoDeltas) == 0 {
		return nil
	}

	if err := s.daoProcessV2.UpdateManyInfo(nCtx, processInfoDeltas); err != nil {
		return fmt.Errorf("failed to batch update process v2 info: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessV2(nCtx contextx.IContext, hostID int64, pluginName string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if pluginName == "" {
		return base.ErrInvalidParam(fmt.Errorf("plugin name is empty"))
	}

	if err := s.daoProcessV2.Delete(nCtx, hostID, pluginName); err != nil {
		return fmt.Errorf("failed to delete process v2: %w", err)
	}

	return nil
}
