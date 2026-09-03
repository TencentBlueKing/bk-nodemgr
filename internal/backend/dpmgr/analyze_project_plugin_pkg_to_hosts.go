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

func (analyzer *Analyzer) analyzeProjectPluginPkgToHosts(
	nCtx contextx.IContext, params *AnalyzeParams,
) ([]*ChangeTask, error) {
	param, err := params.Spec.GetProjectPluginPkgToHostsParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get project plugin pkg to hosts param, spec(%+v): %w", params.Spec, err)
	}

	placementHostMap, err := analyzer.listPlacementHostMap(nCtx, param.PlacementHostIDs)
	if err != nil {
		return nil, err
	}

	processMap, err := analyzer.listProjectPluginPkgProcessMap(nCtx, param.PluginPkgName, param.PlacementHostIDs)
	if err != nil {
		return nil, err
	}

	changeTasks := make([]*ChangeTask, 0)
	for _, sourceTarget := range params.Targets {
		if err := validateProjectPluginPkgSourceTarget(sourceTarget); err != nil {
			return nil, err
		}

		for _, placementHostID := range param.PlacementHostIDs {
			pluginName := genPluginNameForProjectPluginPkgToHosts(
				param.PluginPkgName,
				params.DeployPolicyID,
				sourceTarget.ServiceInstance.ModuleID,
				sourceTarget.ServiceInstance.HostID,
			)
			process, ok := processMap[genProcessUniqueID(placementHostID, pluginName)]
			target := newProjectPluginPkgTarget(placementHostMap[placementHostID], sourceTarget.ServiceInstance)

			if !ok {
				changeTasks = append(changeTasks, &ChangeTask{
					DeployPolicyID: params.DeployPolicyID,
					Action:         ChangeActionPluginPkgInstall,
					Spec:           params.Spec,
					Target:         target,
					ConfigSource:   sourceTarget,
				})

				continue
			}

			if process.Info.Version != param.Version {
				changeTasks = append(changeTasks, &ChangeTask{
					DeployPolicyID: params.DeployPolicyID,
					Action:         ChangeActionPluginPkgUpgrade,
					Spec:           params.Spec,
					Target:         target,
					ConfigSource:   sourceTarget,
				})
			}
		}
	}

	return changeTasks, nil
}

func (analyzer *Analyzer) listPlacementHostMap(
	nCtx contextx.IContext, hostIDs []int64,
) (map[int64]*types.Host, error) {
	cond := &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
	}
	hosts, _, err := analyzer.daoHost.ListHost(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list placement hosts, cond(%+v): %w", cond, err)
	}

	hostMap, err := conv.SliceToMap(hosts, func(host *types.Host) int64 { return host.HostID })
	if err != nil {
		return nil, fmt.Errorf("failed to convert placement hosts to map, hosts(%+v): %w", hosts, err)
	}

	for _, hostID := range hostIDs {
		if _, ok := hostMap[hostID]; !ok {
			return nil, fmt.Errorf("placement host not found, host-id(%d)", hostID)
		}
	}

	return hostMap, nil
}

func (analyzer *Analyzer) listProjectPluginPkgProcessMap(
	nCtx contextx.IContext, pluginPkgName string, hostIDs []int64,
) (map[string]*types.Process, error) {
	cond := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:        hostIDs,
			PluginPkgName: []string{pluginPkgName},
			InfoStatus:    []types.ProcessStatus{types.ProcessStatusRunning},
		},
	}
	processes, _, err := analyzer.daoProcess.ListProcesses(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list project plugin pkg processes, cond(%+v): %w", cond, err)
	}

	processMap, err := conv.SliceToMap(processes, func(process *types.Process) string {
		return genProcessUniqueID(process.HostID, process.PluginName)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert project plugin pkg processes to map, processes(%+v): %w", processes, err)
	}

	return processMap, nil
}

func validateProjectPluginPkgSourceTarget(target *types.Target) error {
	if target == nil {
		return fmt.Errorf("source target is nil")
	}

	if target.ServiceInstance.ModuleID <= 0 {
		return fmt.Errorf("source target module id is invalid, target(%+v)", target)
	}

	if target.ServiceInstance.HostID <= 0 {
		return fmt.Errorf("source target host id is invalid, target(%+v)", target)
	}

	return nil
}

func newProjectPluginPkgTarget(host *types.Host, serviceInstance types.ServiceInstance) *types.Target {
	return &types.Target{
		Host:            *host,
		ServiceInstance: serviceInstance,
	}
}
