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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoProcess "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) getProcessByID(nCtx contextx.IContext, processID string) (*types.Process, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	process, err := s.daoProcess.Get(nCtx, daoProcess.WithProcessID(processID))
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	return process, nil
}

// nolint: nonamedreturns
func (s *Storage) createProcess(nCtx contextx.IContext, process *types.Process) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if process == nil {
		return base.ErrInvalidParam(fmt.Errorf("process is nil"))
	}

	err = s.daoProcess.Create(nCtx, process)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) updateProcess(nCtx contextx.IContext, processID string, processInfo *types.ProcessInfo) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processInfo == nil {
		return base.ErrInvalidParam(fmt.Errorf("process info is nil"))
	}

	err = s.daoProcess.UpdateInfo(nCtx, processID, processInfo)
	if err != nil {
		return fmt.Errorf("failed to update process info: %w", err)
	}

	return nil
}

// nolint: nonamedreturns
func (s *Storage) deleteProcess(nCtx contextx.IContext, processID string) (err error) {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	err = s.daoProcess.Delete(nCtx, processID)
	if err != nil {
		return fmt.Errorf("failed to delete process: %w", err)
	}

	return nil
}
