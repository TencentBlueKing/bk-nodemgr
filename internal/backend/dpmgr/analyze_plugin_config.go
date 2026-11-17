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

package dpmgr

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (analyzer *Analyzer) listPluginMainConfigMap(nCtx contextx.IContext, processMap map[string]*types.Process) (
	map[string][]*types.ProcessConfig, error) {

	configMap := make(map[string][]*types.ProcessConfig)
	if len(processMap) == 0 {
		return configMap, nil
	}

	hostIDs := make([]int64, 0, len(processMap))
	pluginNames := make([]string, 0, len(processMap))
	for _, process := range processMap {
		hostIDs = append(hostIDs, process.HostID)
		pluginNames = append(pluginNames, process.PluginName)
	}
	cond := &types.ProcessConfigCondition{
		ExactInclude: &types.ProcessConfigExactFields{
			HostID:       conv.SliceUnique(hostIDs),
			ProcessName:  conv.SliceUnique(pluginNames),
			IsMainConfig: []bool{true},
		},
	}
	configs, _, err := analyzer.daoProcessConfig.ListProcessConfigs(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugin main configs, cond(%+v): %w", cond, err)
	}

	for _, config := range configs {
		key := genProcessUniqueID(config.HostID, config.ProcessName)
		configMap[key] = append(configMap[key], config)
	}

	return configMap, nil
}

func pluginMainConfigContextChanged(configs []*types.ProcessConfig, customConfigContext map[string]any) bool {
	for _, config := range configs {
		if !sameCustomConfigContext(config.CustomConfigContext, customConfigContext) {
			return true
		}
	}

	// No recorded main config also covers plugins that do not declare a main config template.
	return false
}
