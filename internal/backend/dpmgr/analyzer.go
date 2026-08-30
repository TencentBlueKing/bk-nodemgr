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
	daoProcess       pluginStg.IDaoProcess
	daoProcessConfig pluginStg.IDaoProcessConfig
	daoHost          topo.IStorageHost
}

// AnalyzerConfig defines the config of analyzer.
type AnalyzerConfig struct {
	DaoProcess       pluginStg.IDaoProcess
	DaoProcessConfig pluginStg.IDaoProcessConfig
	DaoHost          topo.IStorageHost
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
		daoProcess:       conf.DaoProcess,
		daoProcessConfig: conf.DaoProcessConfig,
		daoHost:          conf.DaoHost,
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
	case types.DeploySpecTypeSpecifyPluginSubConfig:
		return analyzer.analyzeSpecifyPluginSubConfig(nCtx, params)
	case types.DeploySpecTypeSpecifyPluginSubConfigTemplate:
		return analyzer.analyzeSpecifyPluginSubConfigTemplate(nCtx, params)
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

	processMap, err := conv.SliceToMap(processes, func(process *types.Process) string {
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

	processMap, err := conv.SliceToMap(processes, func(process *types.Process) string {
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

func (analyzer *Analyzer) analyzeSpecifyPluginSubConfig(nCtx contextx.IContext, params *AnalyzeParams) ([]*ChangeTask, error) {
	param, err := params.Spec.GetSpecifyPluginSubConfigParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin sub config param, spec(%+v): %w", params.Spec, err)
	}

	desiredDetails := genDeployPolicySubConfigDetails(params.DeployPolicyID, param.ConfigFilesDetail)
	desiredDetailMap, err := conv.SliceToMap(
		desiredDetails,
		func(detail *types.PluginConfigDetail) string { return detail.Name },
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plugin config details to map, count(%d): %w", len(desiredDetails), err)
	}

	hostIDs := conv.SliceToSlice(params.Targets, func(target *types.Target) int64 {
		return target.Host.HostID
	})
	hostIDs = conv.SliceUnique(hostIDs)
	targetMap := groupTargetsByHost(params.Targets)

	processMap, err := analyzer.listRunningProcessMap(nCtx, param.PluginName, hostIDs)
	if err != nil {
		return nil, err
	}

	deployPolicyConfigMap, err := analyzer.listDeployPolicySubConfigMap(nCtx, param.PluginName, params.DeployPolicyID)
	if err != nil {
		return nil, err
	}

	changeTasks := make([]*ChangeTask, 0)
	for _, target := range params.Targets {
		_, ok := processMap[genProcessUniqueID(target.Host.HostID, param.PluginName)]
		if !ok {
			continue
		}

		missingDetails := findMissingSubConfigDetails(desiredDetails, deployPolicyConfigMap[target.Host.HostID])
		if len(missingDetails) == 0 {
			continue
		}

		applySpec, err := newSpecifyPluginSubConfigSpec(param.PluginName, missingDetails, param.CustomConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create specify plugin sub config spec: %w", err)
		}

		changeTasks = append(changeTasks, &ChangeTask{
			DeployPolicyID: params.DeployPolicyID,
			Action:         ChangeActionPluginApplySubConfig,
			Spec:           applySpec,
			Target:         target,
		})
	}

	deleteDetailsByHost := collectDeleteSubConfigDetails(
		desiredDetailMap,
		targetMap,
		deployPolicyConfigMap,
	)
	deleteTasks, err := genDeleteSubConfigTasks(
		params.DeployPolicyID,
		param.PluginName,
		param.CustomConfigContext,
		targetMap,
		deleteDetailsByHost,
	)
	if err != nil {
		return nil, err
	}

	changeTasks = append(changeTasks, deleteTasks...)

	return changeTasks, nil
}

func (analyzer *Analyzer) analyzeSpecifyPluginSubConfigTemplate(nCtx contextx.IContext, params *AnalyzeParams) (
	[]*ChangeTask, error) {

	param, err := params.Spec.GetSpecifyPluginSubConfigTemplateParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin sub config template param, spec(%+v): %w", params.Spec, err)
	}

	desiredDetails := genDeployPolicySubConfigDetails(params.DeployPolicyID, param.ConfigFilesDetail)

	hostIDs := conv.SliceToSlice(params.Targets, func(target *types.Target) int64 {
		return target.Host.HostID
	})
	hostIDs = conv.SliceUnique(hostIDs)

	processMap, err := analyzer.listRunningProcessMap(nCtx, param.PluginName, hostIDs)
	if err != nil {
		return nil, err
	}

	deployPolicyConfigMap, err := analyzer.listDeployPolicySubConfigMap(nCtx, param.PluginName, params.DeployPolicyID)
	if err != nil {
		return nil, err
	}

	changeTasks := make([]*ChangeTask, 0)
	for _, target := range params.Targets {
		_, ok := processMap[genProcessUniqueID(target.Host.HostID, param.PluginName)]
		if !ok {
			continue
		}

		missingDetails := findMissingSubConfigDetails(desiredDetails, deployPolicyConfigMap[target.Host.HostID])
		if len(missingDetails) == 0 {
			continue
		}

		applySpec, err := newSpecifyPluginSubConfigSpec(param.PluginName, missingDetails, param.CustomConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create specify plugin sub config spec: %w", err)
		}

		changeTasks = append(changeTasks, &ChangeTask{
			DeployPolicyID: params.DeployPolicyID,
			Action:         ChangeActionPluginApplySubConfig,
			Spec:           applySpec,
			Target:         target,
		})
	}

	return changeTasks, nil
}

func (analyzer *Analyzer) listRunningProcessMap(
	nCtx contextx.IContext, pluginName string, hostIDs []int64,
) (map[string]*types.Process, error) {
	if len(hostIDs) == 0 {
		return map[string]*types.Process{}, nil
	}

	cond := &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID:     hostIDs,
			PluginName: []string{pluginName},
			InfoStatus: []types.ProcessStatus{types.ProcessStatusRunning},
		},
	}
	processes, _, err := analyzer.daoProcess.ListProcesses(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list processes, cond(%+v): %w", cond, err)
	}

	processMap, err := conv.SliceToMap(processes, func(process *types.Process) string {
		return genProcessUniqueID(process.HostID, process.PluginName)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert processes to map, processes(%+v): %w", processes, err)
	}

	return processMap, nil
}

func (analyzer *Analyzer) listDeployPolicySubConfigMap(
	nCtx contextx.IContext, pluginName string, deployPolicyID int64,
) (map[int64]map[string]*types.ProcessConfig, error) {
	cond := &types.ProcessConfigCondition{
		ExactInclude: &types.ProcessConfigExactFields{
			ProcessName:  []string{pluginName},
			Set:          []string{genDeployPolicyProcessConfigSet(deployPolicyID)},
			IsMainConfig: []bool{false},
		},
	}
	configs, _, err := analyzer.daoProcessConfig.ListProcessConfigs(nCtx, types.UnlimitedPage(), cond)
	if err != nil {
		return nil, fmt.Errorf("failed to list deploy policy process configs, cond(%+v): %w", cond, err)
	}

	return groupProcessConfigsByHost(configs), nil
}

func genDeployPolicySubConfigDetails(deployPolicyID int64, details []*types.PluginConfigDetail) []*types.PluginConfigDetail {
	desiredDetails := make([]*types.PluginConfigDetail, 0, len(details))
	for _, detail := range details {
		if detail == nil {
			continue
		}

		templateName := detail.TemplateName
		if templateName == "" {
			templateName = detail.Name
		}
		if templateName == "" {
			continue
		}

		desiredDetails = append(desiredDetails, &types.PluginConfigDetail{
			Name:         genDeployPolicySubConfigName(templateName, deployPolicyID),
			TemplateName: templateName,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
			FilePath:     detail.FilePath,
		})
	}

	return desiredDetails
}

func groupProcessConfigsByHost(configs []*types.ProcessConfig) map[int64]map[string]*types.ProcessConfig {
	configMap := make(map[int64]map[string]*types.ProcessConfig)
	for _, config := range configs {
		if config == nil {
			continue
		}

		if _, ok := configMap[config.HostID]; !ok {
			configMap[config.HostID] = make(map[string]*types.ProcessConfig)
		}
		configMap[config.HostID][config.Name] = config
	}

	return configMap
}

func groupTargetsByHost(targets []*types.Target) map[int64]*types.Target {
	targetMap := make(map[int64]*types.Target)
	for _, target := range targets {
		if target == nil {
			continue
		}
		targetMap[target.Host.HostID] = target
	}

	return targetMap
}

func findMissingSubConfigDetails(
	desiredDetails []*types.PluginConfigDetail, currentConfigMap map[string]*types.ProcessConfig,
) []*types.PluginConfigDetail {
	missingDetails := make([]*types.PluginConfigDetail, 0)
	for _, detail := range desiredDetails {
		if detail == nil {
			continue
		}
		if _, ok := currentConfigMap[detail.Name]; ok {
			continue
		}

		missingDetails = append(missingDetails, detail)
	}

	return missingDetails
}

func findStaleSubConfigDetails(
	desiredDetailMap map[string]*types.PluginConfigDetail, currentConfigMap map[string]*types.ProcessConfig,
) []*types.PluginConfigDetail {
	staleDetails := make([]*types.PluginConfigDetail, 0)
	for _, config := range currentConfigMap {
		if config == nil {
			continue
		}
		if _, ok := desiredDetailMap[config.Name]; ok {
			continue
		}

		staleDetails = append(staleDetails, convProcessConfigToPluginConfigDetail(config))
	}

	return staleDetails
}

func collectDeleteSubConfigDetails(
	desiredDetailMap map[string]*types.PluginConfigDetail,
	targetMap map[int64]*types.Target,
	deployPolicyConfigMap map[int64]map[string]*types.ProcessConfig,
) map[int64][]*types.PluginConfigDetail {
	deleteDetailsByHost := make(map[int64][]*types.PluginConfigDetail)
	seen := make(map[string]struct{})

	for hostID, currentConfigMap := range deployPolicyConfigMap {
		if _, ok := targetMap[hostID]; !ok {
			for _, config := range currentConfigMap {
				addDeleteSubConfigDetail(deleteDetailsByHost, seen, hostID, convProcessConfigToPluginConfigDetail(config))
			}
			continue
		}

		staleDetails := findStaleSubConfigDetails(desiredDetailMap, currentConfigMap)
		for _, detail := range staleDetails {
			addDeleteSubConfigDetail(deleteDetailsByHost, seen, hostID, detail)
		}
	}

	return deleteDetailsByHost
}

func addDeleteSubConfigDetail(
	deleteDetailsByHost map[int64][]*types.PluginConfigDetail,
	seen map[string]struct{},
	hostID int64,
	detail *types.PluginConfigDetail,
) {
	if detail == nil {
		return
	}

	key := fmt.Sprintf("%d_%s", hostID, detail.Name)
	if _, ok := seen[key]; ok {
		return
	}

	seen[key] = struct{}{}
	deleteDetailsByHost[hostID] = append(deleteDetailsByHost[hostID], detail)
}

func genDeleteSubConfigTasks(
	deployPolicyID int64,
	pluginName string,
	customConfigContext map[string]any,
	targetMap map[int64]*types.Target,
	deleteDetailsByHost map[int64][]*types.PluginConfigDetail,
) ([]*ChangeTask, error) {
	changeTasks := make([]*ChangeTask, 0)
	for hostID, details := range deleteDetailsByHost {
		deleteSpec, err := newSpecifyPluginSubConfigSpec(pluginName, details, customConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create specify plugin sub config delete spec: %w", err)
		}

		changeTasks = append(changeTasks, &ChangeTask{
			DeployPolicyID: deployPolicyID,
			Action:         ChangeActionPluginDeleteSubConfig,
			Spec:           deleteSpec,
			Target:         selectSubConfigTarget(hostID, targetMap),
		})
	}

	return changeTasks, nil
}

func selectSubConfigTarget(hostID int64, targetMap map[int64]*types.Target) *types.Target {
	if target, ok := targetMap[hostID]; ok {
		return target
	}

	return &types.Target{
		Host: types.Host{
			HostID: hostID,
		},
	}
}

func convProcessConfigToPluginConfigDetail(config *types.ProcessConfig) *types.PluginConfigDetail {
	if config == nil {
		return nil
	}

	return &types.PluginConfigDetail{
		Name:         config.Name,
		TemplateName: config.TemplateName,
		Content:      config.Content,
		IsMainConfig: config.IsMainConfig,
		FilePath:     config.FilePath,
	}
}

func newSpecifyPluginSubConfigSpec(
	pluginName string, details []*types.PluginConfigDetail, customConfigContext map[string]any,
) (*types.DeploySpec, error) {
	return types.NewDeploySpecWithSpecifyPluginSubConfig(&types.SpecifyPluginSubConfigParam{
		PluginName:          pluginName,
		ConfigFilesDetail:   details,
		CustomConfigContext: customConfigContext,
	})
}
