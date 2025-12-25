/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
			unitChangeTasks, err := analyzer.analyze(nCtx, spec, unit.Targets)
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
func (analyzer *Analyzer) analyze(nCtx contextx.IContext, spec *types.DeploySpec, targets []*types.Target) ([]*ChangeTask, error) {
	switch spec.Type {
	case types.DeploySpecTypeSpecifyPlugin:
		return analyzer.analyzeSpecifyPlugin(nCtx, spec, targets)
	default:
		return nil, fmt.Errorf("failed to analyze deploy unit, spec(%+v)", spec)
	}
}

// analyzeSpecifyPlugin analyze specify plugin.
func (analyzer *Analyzer) analyzeSpecifyPlugin(nCtx contextx.IContext, spec *types.DeploySpec, targets []*types.Target) ([]*ChangeTask, error) {
	param, err := spec.GetSpecifyPluginParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin param, spec(%+v): %w", spec, err)
	}

	hostIDs := make([]int64, len(targets))
	for i, target := range targets {
		hostIDs[i] = target.Host.HostID
	}

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

	processMap, err := conv.SliceToMap[int64, *types.Process](processes, func(process *types.Process) int64 {
		return process.HostID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert processes to map, processes(%+v): %w", processes, err)
	}

	changeTasks := make([]*ChangeTask, 0)
	for _, target := range targets {
		//  Only one process with the same name will be used on a host.
		process, ok := processMap[target.Host.HostID]
		// if process not exist, install it.
		if !ok {
			changeTasks = append(changeTasks, &ChangeTask{
				Action: ChangeActionPluginInstall,
				Spec:   spec,
				Target: target,
			})

			continue
		}

		// if process version not match, upgrade it.
		if process.Info.Version != param.PluginVersion {
			changeTasks = append(changeTasks, &ChangeTask{
				Action: ChangeActionPluginUpgrade,
				Spec:   spec,
				Target: target,
			})

			continue
		}

		// everything ok, do nothing.
	}

	return changeTasks, nil
}
