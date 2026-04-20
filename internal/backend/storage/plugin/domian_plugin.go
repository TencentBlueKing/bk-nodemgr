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
