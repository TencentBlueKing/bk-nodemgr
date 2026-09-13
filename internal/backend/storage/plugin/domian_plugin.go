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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoPlugin "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) getPluginVisibleBizIDs(nCtx contextx.IContext, pluginName string) ([]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	plugin, err := s.daoPlugin.Get(nCtx, daoPlugin.WithName(pluginName))
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin visible biz ids, plugin-name(%s): %w", pluginName, err)
	}

	return plugin.VisibleBizIDs, nil
}

func (s *Storage) listVisiblePluginByBizIDs(nCtx contextx.IContext, bizIDs []int64) ([]*types.Plugin, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	plugin, _, err := s.daoPlugin.List(nCtx, types.UnlimitedPage(), daoPlugin.WithVisibleBizIDs(bizIDs...))
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin list, biz-ids(%v): %w", bizIDs, err)
	}

	return plugin, nil
}

// maxMemoFields covers the description and scenario in both languages.
const maxMemoFields = 4

func (s *Storage) ensureDefaultPlugin(nCtx contextx.IContext, release *types.ReleasePlugin) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if release == nil {
		return errors.New("release plugin is nil")
	}

	if release.Name == "" {
		return errors.New("plugin package name is empty")
	}

	exist, err := s.existPluginByPluginPkgName(nCtx, release.Name)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	plugin := &types.Plugin{
		Name:    release.Name,
		PkgName: release.Name,
		Group:   types.PluginGroupDefault,
		Memo:    buildPluginMemo(release),
	}
	if err := s.createPlugin(nCtx, plugin); err != nil {
		exist, existErr := s.existPluginByPluginPkgName(nCtx, release.Name)
		if existErr != nil {
			return errors.Join(err, fmt.Errorf("failed to recheck default plugin: %w", existErr))
		}

		if exist {
			return nil
		}

		return fmt.Errorf("plugin name(%s) is already occupied without a matching default plugin: %w", release.Name, err)
	}

	return nil
}

func buildPluginMemo(plugin *types.ReleasePlugin) string {
	memoParts := make([]string, 0, maxMemoFields)
	if plugin.Description != "" {
		memoParts = append(memoParts, fmt.Sprintf("描述: %s", plugin.Description))
	}
	if plugin.Scenario != "" {
		memoParts = append(memoParts, fmt.Sprintf("场景: %s", plugin.Scenario))
	}
	if plugin.DescriptionEn != "" {
		memoParts = append(memoParts, fmt.Sprintf("Description: %s", plugin.DescriptionEn))
	}
	if plugin.ScenarioEn != "" {
		memoParts = append(memoParts, fmt.Sprintf("Scene: %s", plugin.ScenarioEn))
	}

	return strings.Join(memoParts, "\n")
}
