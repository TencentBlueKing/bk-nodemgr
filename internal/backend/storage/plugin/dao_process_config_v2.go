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
	daoProcessV2Config "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process-v2-config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) getProcessConfigV2(nCtx contextx.IContext, processConfigKey *types.ProcessConfigUniqueKey) (
	*types.ProcessConfig, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if processConfigKey == nil {
		return nil, base.ErrInvalidParam(errors.New("processConfigKey is nil"))
	}

	if processConfigKey.Name == "" {
		return nil, base.ErrInvalidParam(errors.New("processConfigKey name is empty"))
	}

	config, err := s.daoProcessV2Config.Get(nCtx, processConfigKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get processConfig v2: %w", err)
	}

	return config, nil
}

func (s *Storage) countProcessConfigsV2(nCtx contextx.IContext, conditions ...*types.ProcessConfigCondition) (
	int64, error) {

	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConfigConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	cnt, err := s.daoProcessV2Config.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count processConfigs v2: %w", err)
	}

	return cnt, nil
}

func (s *Storage) listProcessConfigsV2(nCtx contextx.IContext, page types.Page,
	conditions ...*types.ProcessConfigCondition) ([]*types.ProcessConfig, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts, err := convertProcessV2ConfigConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	configs, cnt, err := s.daoProcessV2Config.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list processConfigs v2: %w", err)
	}

	return configs, cnt, nil
}

// convertProcessV2ConfigConditionsToOptions converts the v2 process config conditions to the v2 dao options.
// The v2 dao shares the v3 option helpers because the two config tables keep the same schema.
func convertProcessV2ConfigConditionsToOptions(conditions ...*types.ProcessConfigCondition) (
	[]daoProcessV2Config.OptFn, error) {

	opts := make([]daoProcessV2Config.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoProcessV2Config.WithName(condition.ExactInclude.Name...),
				daoProcessV2Config.WithTemplateName(condition.ExactInclude.TemplateName...),
				daoProcessV2Config.WithProcessName(condition.ExactInclude.ProcessName...),
				daoProcessV2Config.WithHostID(condition.ExactInclude.HostID...),
				daoProcessV2Config.WithSet(condition.ExactInclude.Set...),
				daoProcessV2Config.WithIsMainConfig(condition.ExactInclude.IsMainConfig...),
			)
		}

		if condition.FuzzyInclude != nil {
			return nil, errors.New("fuzzy include is not supported")
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				daoProcessV2Config.WithoutName(condition.ExactExclude.Name...),
				daoProcessV2Config.WithoutTemplateName(condition.ExactExclude.TemplateName...),
				daoProcessV2Config.WithoutProcessName(condition.ExactExclude.ProcessName...),
				daoProcessV2Config.WithoutHostID(condition.ExactExclude.HostID...),
				daoProcessV2Config.WithoutSet(condition.ExactExclude.Set...),
			)
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}

func (s *Storage) upsertProcessConfigsV2(nCtx contextx.IContext, processConfigs ...*types.ProcessConfig) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processConfigs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("processConfigs is empty"))
	}

	for _, processConfig := range processConfigs {
		if processConfig == nil {
			return base.ErrInvalidParam(errors.New("processConfig is nil"))
		}
	}

	if err := s.daoProcessV2Config.UpsertMany(nCtx, processConfigs...); err != nil {
		return fmt.Errorf("failed to upsert processConfigs v2: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessConfigsV2ByProcessUniqueKey(
	nCtx contextx.IContext, processUniqueKeys ...*types.ProcessUniqueKey) error {

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

	err := s.daoProcessV2Config.DeleteMany(nCtx, daoProcessV2Config.WithProcessUniqueKeys(processUniqueKeys...))
	if err != nil {
		return fmt.Errorf("failed to delete processConfigs v2 by processUniqueKey: %w", err)
	}

	return nil
}

func (s *Storage) deleteProcessConfigsV2(nCtx contextx.IContext, processConfigKeys ...*types.ProcessConfigUniqueKey) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(processConfigKeys) == 0 {
		return base.ErrInvalidParam(errors.New("processConfigKeys is empty"))
	}

	for _, key := range processConfigKeys {
		if key == nil {
			return base.ErrInvalidParam(errors.New("processConfigKey is nil"))
		}

		if key.Name == "" {
			return base.ErrInvalidParam(errors.New("processConfigKey name is empty"))
		}
	}

	err := s.daoProcessV2Config.DeleteMany(nCtx, daoProcessV2Config.WithProcessConfigUniqueKeys(processConfigKeys...))
	if err != nil {
		return fmt.Errorf("failed to delete processConfigs v2: %w", err)
	}

	return nil
}
