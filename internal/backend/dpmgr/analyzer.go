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

	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAnalyzer defines the analyzer interface.
type IAnalyzer interface {
	Analyze(nCtx contextx.IContext, units ...*DeployUnit) ([]*ChangeTask, error)
}

var _ IAnalyzer = &Analyzer{}

// Analyzer defines the analyzer.
type Analyzer struct {
	daoProcess pluginStg.IDaoProcess
	daoHost    topo.IStorageHost
}

// AnalyzerConfig defines the config of analyzer.
type AnalyzerConfig struct {
	DaoProcess pluginStg.IDaoProcess
	DaoHost    topo.IStorageHost
}

// Analyze analyzes the deploy units and design the change tasks.
func (analyzer *Analyzer) Analyze(nCtx contextx.IContext, units ...*DeployUnit) ([]*ChangeTask, error) {
	changeTasks := make([]*ChangeTask, 0)

	for _, unit := range units {
		for _, spec := range unit.Specs {
			params := &AnalyzeParams{
				DeployPolicyID: unit.DeployPolicyID,
				Spec:           spec,
				Targets:        unit.Targets,
			}
			unitChangeTasks, err := analyzer.analyze(nCtx, params)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze deploy unit, deploy-unit(%+v): %w", unit, err)
			}

			changeTasks = append(changeTasks, unitChangeTasks...)
		}
	}

	return changeTasks, nil
}

// NewAnalyzer creates a new analyzer.
func NewAnalyzer(conf *AnalyzerConfig) *Analyzer {
	return &Analyzer{
		daoProcess: conf.DaoProcess,
		daoHost:    conf.DaoHost,
	}
}

// analyze analyzes the deploy unit and design the change task.
func (analyzer *Analyzer) analyze(nCtx contextx.IContext, params *AnalyzeParams) (
	[]*ChangeTask, error) {

	switch params.Spec.Type() {
	case types.DeploySpecTypeSpecifyPlugin:
		return analyzer.analyzeSpecifyPlugin(nCtx, params)
	case types.DeploySpecTypeSpecifyPluginPkg:
		return analyzer.analyzeSpecifyPluginPkg(nCtx, params)
	default:
		return nil, fmt.Errorf("failed to analyze deploy unit, spec(%+v)", params.Spec)
	}
}

// analyzeSpecifyPlugin analyze specify plugin.
func (analyzer *Analyzer) analyzeSpecifyPlugin(nCtx contextx.IContext, params *AnalyzeParams) (
	[]*ChangeTask, error) {

	param, err := params.Spec.GetSpecifyPluginParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin param, spec(%+v): %w", params.Spec, err)
	}

	hostIDs := conv.SliceToSlice(params.Targets, func(target *types.Target) int64 {
		return target.Host.HostID
	})

	hostIDs = conv.SliceUnique(hostIDs)

	cond := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:     hostIDs,
			PluginName: []string{param.PluginName},
			InfoStatus: []types.ProcessStatus{types.ProcessStatusRunning},
		},
	}

	processes, _, err := analyzer.daoProcess.ListProcesses(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list processes, cond(%+v): %w", cond, err)
	}

	processMap, err := conv.SliceToMap[string, *types.Process](processes, func(process *types.Process) string {
		return genProcessUniqueID(process.HostID, process.PluginName)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert processes to map, processes(%+v): %w", processes, err)
	}

	changeTasks := make([]*ChangeTask, 0)
	for _, target := range params.Targets {
		//  Only one process with the same name will be used on a host.
		process, ok := processMap[genProcessUniqueID(target.Host.HostID, param.PluginName)]
		// if process not exist, install it.
		if !ok {
			changeTasks = append(changeTasks, &ChangeTask{
				DeployPolicyID: params.DeployPolicyID,
				Action:         ChangeActionPluginInstall,
				Spec:           params.Spec,
				Target:         target,
			})

			continue
		}

		// if process version not match, upgrade it.
		if process.Info.Version != param.Version {
			changeTasks = append(changeTasks, &ChangeTask{
				DeployPolicyID: params.DeployPolicyID,
				Action:         ChangeActionPluginUpgrade,
				Spec:           params.Spec,
				Target:         target,
			})

			continue
		}

		// everything ok, do nothing.
	}

	return changeTasks, nil
}

// analyzeSpecifyPluginPkg analyze specify plugin pkg.
func (analyzer *Analyzer) analyzeSpecifyPluginPkg(nCtx contextx.IContext, params *AnalyzeParams) ([]*ChangeTask, error) {
	param, err := params.Spec.GetSpecifyPluginPkgParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin pkg param, spec(%+v): %w", params.Spec, err)
	}

	hostIDs := conv.SliceToSlice(params.Targets, func(target *types.Target) int64 {
		return target.Host.HostID
	})

	hostIDs = conv.SliceUnique(hostIDs)

	cond := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:        hostIDs,
			PluginPkgName: []string{param.PluginPkgName},
			InfoStatus:    []types.ProcessStatus{types.ProcessStatusRunning},
		},
	}

	processes, _, err := analyzer.daoProcess.ListProcesses(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list processes, cond(%+v): %w", cond, err)
	}

	processMap, err := conv.SliceToMap[string, *types.Process](processes, func(process *types.Process) string {
		return genProcessUniqueID(process.HostID, process.PluginName)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert processes to map, processes(%+v): %w", processes, err)
	}

	changeTasks := make([]*ChangeTask, 0)
	for _, target := range params.Targets {
		pluginName := genPluginNameForSpecifyPluginPkg(param.PluginPkgName, params.DeployPolicyID, target.ServiceInstance.ModuleID)

		//  Only one process with the same name will be used on a host.
		process, ok := processMap[genProcessUniqueID(target.Host.HostID, pluginName)]
		// if process not exist, install it.
		if !ok {
			changeTasks = append(changeTasks, &ChangeTask{
				DeployPolicyID: params.DeployPolicyID,
				Action:         ChangeActionPluginPkgInstall,
				Spec:           params.Spec,
				Target:         target,
			})

			continue
		}

		// if process version not match, upgrade it.
		if process.Info.Version != param.Version {
			changeTasks = append(changeTasks, &ChangeTask{
				DeployPolicyID: params.DeployPolicyID,
				Action:         ChangeActionPluginPkgUpgrade,
				Spec:           params.Spec,
				Target:         target,
			})

			continue
		}

		// everything ok, do nothing.
	}

	return changeTasks, nil
}
