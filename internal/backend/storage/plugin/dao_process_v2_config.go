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
	daoProcessV2Config "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process-v2-config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createProcessV2Config(nCtx contextx.IContext, processConfig *types.ProcessConfig) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processConfig == nil {
		return base.ErrInvalidParam(fmt.Errorf("processConfig is nil"))
	}

	err := s.daoProcessV2Config.Create(nCtx, processConfig)
	if err != nil {
		return fmt.Errorf("failed to create processConfig: %w", err)
	}

	return nil
}

func (s *Storage) upsertProcessV2Configs(nCtx contextx.IContext, processConfigs ...*types.ProcessConfig) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processConfigs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("processConfigs is empty"))
	}

	err := s.daoProcessV2Config.UpsertMany(nCtx, processConfigs...)
	if err != nil {
		return fmt.Errorf("failed to upsert processConfigs: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessV2ConfigsByProcessUniqueKey(nCtx contextx.IContext, processUniqueKeys ...*types.ProcessUniqueKey) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processUniqueKeys) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("processUniqueKeys is empty"))
	}

	for _, key := range processUniqueKeys {
		if key == nil {
			return base.ErrInvalidParam(errors.New("processUniqueKey is nil"))
		}
	}

	err := s.daoProcessV2Config.DeleteMany(
		nCtx, daoProcessV2Config.WithProcessUniqueKeys(processUniqueKeys...))
	if err != nil {
		return fmt.Errorf("failed to delete processConfigs by processUniqueKey: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessV2Configs(nCtx contextx.IContext, processUniqueKey *types.ProcessUniqueKey, names ...string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if processUniqueKey == nil {
		return base.ErrInvalidParam(fmt.Errorf("processUniqueKey is nil"))
	}

	if len(names) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("names is empty"))
	}

	err := s.daoProcessV2Config.DeleteMany(
		nCtx,
		daoProcessV2Config.WithProcessName(processUniqueKey.Name),
		daoProcessV2Config.WithHostID(processUniqueKey.HostID),
		daoProcessV2Config.WithName(names...),
	)
	if err != nil {
		return fmt.Errorf("failed to delete processConfigs: %w", err)
	}

	return nil
}

func (s *Storage) getProcessV2Config(nCtx contextx.IContext, processUniqueKey *types.ProcessUniqueKey, name string) (*types.ProcessConfig, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if processUniqueKey == nil {
		return nil, base.ErrInvalidParam(fmt.Errorf("processUniqueKey is nil"))
	}

	if name == "" {
		return nil, base.ErrInvalidParam(fmt.Errorf("name is empty"))
	}

	config, err := s.daoProcessV2Config.Get(nCtx, processUniqueKey.HostID, processUniqueKey.Name, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get processConfig: %w", err)
	}

	return config, nil
}

func (s *Storage) countProcessV2Configs(nCtx contextx.IContext, conditions ...*types.ProcessConfigCondition) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConfigConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	cnt, err := s.daoProcessV2Config.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count processConfigs: %w", err)
	}

	return cnt, nil
}

func (s *Storage) listProcessV2Configs(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessConfigCondition) (
	[]*types.ProcessConfig, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConfigConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	configs, cnt, err := s.daoProcessV2Config.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list processConfigs: %w", err)
	}

	return configs, cnt, nil
}

func convertProcessV2ConfigConditionsToOptions(conditions ...*types.ProcessConfigCondition) ([]daoProcessV2Config.OptFn, error) {
	opts := make([]daoProcessV2Config.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoProcessV2Config.WithName(condition.ExactInclude.Name...),
				daoProcessV2Config.WithProcessName(condition.ExactInclude.ProcessName...),
				daoProcessV2Config.WithHostID(condition.ExactInclude.HostID...),
				daoProcessV2Config.WithIsMainConfig(condition.ExactInclude.IsMainConfig...),
			)
		}

		if condition.FuzzyInclude != nil {
			return nil, errors.New("fuzzy include is not supported")
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				daoProcessV2Config.WithoutName(condition.ExactExclude.Name...),
				daoProcessV2Config.WithoutProcessName(condition.ExactExclude.ProcessName...),
				daoProcessV2Config.WithoutHostID(condition.ExactExclude.HostID...),
			)
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
