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
	daoPlugin "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) getPlugin(nCtx contextx.IContext, pluginName string) (*types.Plugin, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	plugin, err := s.daoPlugin.Get(nCtx,
		daoPlugin.WithName(pluginName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin, plugin-name(%s): %w", pluginName, err)
	}

	return plugin, nil
}

func (s *Storage) countPlugins(nCtx contextx.IContext, condition ...*types.PluginCondition) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	opts, err := convertPluginConditionToOptions(condition)
	if err != nil {
		return 0, fmt.Errorf("failed to convert plugin condition to options: %w", err)
	}

	count, err := s.daoPlugin.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count plugins: %w", err)
	}

	return count, nil
}

func (s *Storage) listPlugins(nCtx contextx.IContext, page types.Page, condition ...*types.PluginCondition) ([]*types.Plugin, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	opts, err := convertPluginConditionToOptions(condition)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert plugin condition to options: %w", err)
	}

	plugins, cnt, err := s.daoPlugin.List(nCtx, page, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plugins: %w", err)
	}

	return plugins, cnt, nil
}

func convertPluginConditionToOptions(conditions []*types.PluginCondition) ([]daoPlugin.OptFn, error) {
	opts := make([]daoPlugin.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				daoPlugin.WithName(condition.ExactInclude.Name...),
				daoPlugin.WithGroup(condition.ExactInclude.Group...),
			)
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				daoPlugin.WithFuzzyName(condition.FuzzyInclude.Name...),
				daoPlugin.WithFuzzyPkgName(condition.FuzzyInclude.PkgName...),
			)
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

func (s *Storage) existPluginByPluginName(nCtx contextx.IContext, name string) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	exist, err := s.daoPlugin.Exist(nCtx, daoPlugin.WithName(name))
	if err != nil {
		return false, fmt.Errorf("failed to get plugin, plugin-name(%s): %w", name, err)
	}

	return exist, nil
}

func (s *Storage) existPluginByPluginPkgName(nCtx contextx.IContext, name string) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	exist, err := s.daoPlugin.Exist(nCtx, daoPlugin.WithPkgName(name), daoPlugin.WithGroup(types.PluginGroupDefault))
	if err != nil {
		return false, fmt.Errorf("failed to get plugin, plugin-pkg-name(%s): %w", name, err)
	}

	return exist, nil
}

func (s *Storage) createPlugin(nCtx contextx.IContext, p *types.Plugin) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if p == nil {
		return fmt.Errorf("plugin is nil")
	}

	if err := s.daoPlugin.Create(nCtx, p); err != nil {
		return fmt.Errorf("failed to create plugin, plugin-name(%s): %w", p.Name, err)
	}

	return nil
}
