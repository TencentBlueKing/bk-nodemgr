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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var _ policyCalculator = &absentCalculation{}

type absentCalculation struct {
	daoProcess       pluginStg.IDaoProcess
	daoProcessConfig pluginStg.IDaoProcessConfig
	daoHost          topo.IStorageHost
}

// Calculate derives reverse changes from specs without forward scope calculation.
func (calculation *absentCalculation) Calculate(
	nCtx contextx.IContext, policies []*types.DeployPolicy,
) (*policyCalculation, error) {

	changeTasks := make([]*ChangeTask, 0)
	units := make([]*DeployUnit, 0, len(policies))
	for _, policy := range policies {
		tasks, err := calculation.calculateAbsentChanges(nCtx, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate absent changes for policy %d: %w", policy.DeployPolicyID, err)
		}
		changeTasks = append(changeTasks, tasks...)
		targets := make([]*types.Target, 0, len(tasks))
		seenHosts := make(map[int64]struct{})
		for _, task := range tasks {
			if _, ok := seenHosts[task.Target.Host.HostID]; ok {
				continue
			}
			seenHosts[task.Target.Host.HostID] = struct{}{}
			targets = append(targets, task.Target)
		}
		units = append(units, &DeployUnit{
			DeployPolicyID: policy.DeployPolicyID,
			LifeCycle:      policy.LifeCycle,
			Targets:        targets,
			Specs:          policy.Specs,
		})
	}

	return &policyCalculation{
		originalUnits: units,
		resolvedUnits: units,
		changeTasks:   changeTasks,
	}, nil
}

func (calculation *absentCalculation) calculateAbsentChanges(
	nCtx contextx.IContext, policy *types.DeployPolicy,
) ([]*ChangeTask, error) {

	changeTasks := make([]*ChangeTask, 0)
	seen := make(map[string]struct{})
	for specIndex, spec := range policy.Specs {
		tasks, err := calculation.calculateSpecAbsentChanges(nCtx, policy.DeployPolicyID, spec)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate absent spec %d (%s): %w", specIndex, spec.Type(), err)
		}
		for _, task := range tasks {
			key, err := absentTaskKey(task)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			task.specIndex = specIndex
			changeTasks = append(changeTasks, task)
		}
	}

	return changeTasks, nil
}

func (calculation *absentCalculation) calculateSpecAbsentChanges(
	nCtx contextx.IContext, policyID int64, spec *types.DeploySpec,
) ([]*ChangeTask, error) {

	switch spec.Type() {
	case types.DeploySpecTypeSpecifyPluginPkg:
		param, err := spec.GetSpecifyPluginPkgParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin pkg param: %w", err)
		}

		return calculation.calculateAbsentPlugins(nCtx, policyID, param.PluginPkgName, param.Version)
	case types.DeploySpecTypeProjectPluginPkgToHosts:
		param, err := spec.GetProjectPluginPkgToHostsParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get projected plugin pkg param: %w", err)
		}

		return calculation.calculateAbsentPlugins(nCtx, policyID, param.PluginPkgName, param.Version)
	case types.DeploySpecTypeSpecifyPluginSubConfig:
		param, err := spec.GetSpecifyPluginSubConfigParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin sub config param: %w", err)
		}

		return calculation.calculateAbsentConfigs(nCtx, policyID, param.PluginName)
	case types.DeploySpecTypeSpecifyPluginSubConfigTemplate:
		param, err := spec.GetSpecifyPluginSubConfigTemplateParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin sub config template param: %w", err)
		}

		return calculation.calculateAbsentConfigs(nCtx, policyID, param.PluginName)
	default:
		return nil, fmt.Errorf("unsupported ensure_absent spec type (%s)", spec.Type())
	}
}

func (calculation *absentCalculation) calculateAbsentPlugins(
	nCtx contextx.IContext, policyID int64, pluginPkgName, version string,
) ([]*ChangeTask, error) {

	cond := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			PluginGroup:   []string{fmt.Sprintf("%d", policyID)},
			PluginPkgName: []string{pluginPkgName},
		},
	}
	processes, _, err := calculation.daoProcess.ListProcesses(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy plugins: %w", err)
	}

	tasks := make([]*ChangeTask, 0, len(processes))
	for _, process := range processes {
		// Uninstall uses the stored plugin name; the declared version does not filter removal.
		spec, err := types.NewDeploySpecWithSpecifyPlugin(&types.SpecifyPluginParam{
			PluginName: process.PluginName,
			Version:    version,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin uninstall spec: %w", err)
		}
		target, err := calculation.absentTarget(nCtx, process.HostID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &ChangeTask{
			DeployPolicyID: policyID,
			Action:         ChangeActionPluginUninstall,
			Spec:           spec,
			Target:         target,
		})
	}

	return tasks, nil
}

func (calculation *absentCalculation) calculateAbsentConfigs(
	nCtx contextx.IContext, policyID int64, pluginName string,
) ([]*ChangeTask, error) {

	cond := &types.ProcessConfigCondition{
		ExactInclude: &types.ProcessConfigExactFields{
			ProcessName:  []string{pluginName},
			Set:          []string{genDeployPolicyProcessConfigSet(policyID)},
			IsMainConfig: []bool{false},
		},
	}
	configs, _, err := calculation.daoProcessConfig.ListProcessConfigs(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy configs: %w", err)
	}

	configsByHost := groupProcessConfigsByHost(configs)
	tasks := make([]*ChangeTask, 0, len(configsByHost))
	for hostID, hostConfigs := range configsByHost {
		details := make([]*types.PluginConfigDetail, 0, len(hostConfigs))
		for _, config := range hostConfigs {
			details = append(details, convProcessConfigToPluginConfigDetail(config))
		}
		spec, err := newSpecifyPluginSubConfigSpec(pluginName, details, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create config deletion spec: %w", err)
		}
		target, err := calculation.absentTarget(nCtx, hostID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &ChangeTask{
			DeployPolicyID: policyID,
			Action:         ChangeActionPluginDeleteSubConfig,
			Spec:           spec,
			Target:         target,
		})
	}

	return tasks, nil
}

func absentTaskKey(task *ChangeTask) (string, error) {
	if task.Spec.Type() == types.DeploySpecTypeSpecifyPlugin {
		param, err := task.Spec.GetSpecifyPluginParam()
		if err != nil {
			return "", fmt.Errorf("failed to identify absent plugin task: %w", err)
		}

		return fmt.Sprintf("plugin/%d/%s", task.Target.Host.HostID, param.PluginName), nil
	}
	param, err := task.Spec.GetSpecifyPluginSubConfigParam()
	if err != nil {
		return "", fmt.Errorf("failed to identify absent config task: %w", err)
	}

	return fmt.Sprintf("config/%d/%s", task.Target.Host.HostID, param.PluginName), nil
}

func (calculation *absentCalculation) absentTarget(nCtx contextx.IContext, hostID int64) (*types.Target, error) {
	cond := &types.HostCondition{StaticExactInclude: &types.HostStaticExactFields{HostID: []int64{hostID}}}
	hosts, _, err := calculation.daoHost.ListHost(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to find absent target host %d: %w", hostID, err)
	}
	if len(hosts) != 1 {
		return nil, fmt.Errorf("absent target host %d not found", hostID)
	}

	return &types.Target{Host: *hosts[0]}, nil
}
