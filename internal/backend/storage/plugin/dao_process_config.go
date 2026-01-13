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
	daoProcessConfig "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process-config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createProcessConfig(nCtx contextx.IContext, processConfig *types.ProcessConfig) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processConfig == nil {
		return base.ErrInvalidParam(fmt.Errorf("processConfig is nil"))
	}

	err := s.daoProcessConfig.Create(nCtx, processConfig)
	if err != nil {
		return fmt.Errorf("failed to create processConfig: %w", err)
	}

	return nil
}

func (s *Storage) upsertProcessConfigs(nCtx contextx.IContext, processConfigs ...*types.ProcessConfig) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processConfigs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("processConfigs is empty"))
	}

	err := s.daoProcessConfig.UpsertMany(nCtx, processConfigs...)
	if err != nil {
		return fmt.Errorf("failed to upsert processConfigs: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessConfigsByProcessID(nCtx contextx.IContext, processID ...string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processID) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("processID is empty"))
	}

	err := s.daoProcessConfig.DeleteMany(nCtx, daoProcessConfig.WithProcessID(processID...))
	if err != nil {
		return fmt.Errorf("failed to delete processConfigs by processID: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessConfigs(nCtx contextx.IContext, processID string, names ...string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processID == "" {
		return base.ErrInvalidParam(fmt.Errorf("processID is empty"))
	}

	if len(names) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("names is empty"))
	}

	err := s.daoProcessConfig.DeleteMany(nCtx, daoProcessConfig.WithProcessID(processID), daoProcessConfig.WithName(names...))
	if err != nil {
		return fmt.Errorf("failed to delete processConfigs: %w", err)
	}

	return nil
}

func (s *Storage) getProcessConfig(nCtx contextx.IContext, processID, name string) (*types.ProcessConfig, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if processID == "" {
		return nil, base.ErrInvalidParam(fmt.Errorf("processID is empty"))
	}

	if name == "" {
		return nil, base.ErrInvalidParam(fmt.Errorf("name is empty"))
	}

	config, err := s.daoProcessConfig.Get(nCtx, name, processID)
	if err != nil {
		return nil, fmt.Errorf("failed to get processConfig: %w", err)
	}

	return config, nil
}

func (s *Storage) countProcessConfigs(nCtx contextx.IContext, conditions ...*types.ProcessConfigCondition) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessConfigConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	cnt, err := s.daoProcessConfig.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count processConfigs: %w", err)
	}

	return cnt, nil
}

func (s *Storage) listProcessConfigs(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessConfigCondition) (
	[]*types.ProcessConfig, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessConfigConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	configs, cnt, err := s.daoProcessConfig.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list processConfigs: %w", err)
	}

	return configs, cnt, nil
}

func convertProcessConfigConditionsToOptions(conditions ...*types.ProcessConfigCondition) ([]daoProcessConfig.OptFn, error) {
	opts := make([]daoProcess.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoProcessConfig.WithName(condition.ExactInclude.Name...),
				daoProcessConfig.WithProcessID(condition.ExactInclude.ProcessID...),
			)
		}

		if condition.FuzzyInclude != nil {
			return nil, errors.New("fuzzy include is not supported")
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				daoProcessConfig.WithoutName(condition.ExactInclude.Name...),
				daoProcessConfig.WithoutProcessID(condition.ExactInclude.ProcessID...),
			)
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
