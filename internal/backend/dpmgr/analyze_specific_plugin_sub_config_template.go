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
	"reflect"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (analyzer *Analyzer) analyzeSpecifyPluginSubConfigTemplate(nCtx contextx.IContext, params *AnalyzeParams) (
	[]*ChangeTask, error) {

	param, err := params.Spec.GetSpecifyPluginSubConfigTemplateParam()
	if err != nil {
		return nil, fmt.Errorf("failed to get specify plugin sub config template param, spec(%+v): %w", params.Spec, err)
	}

	desiredDetails := genDeployPolicySubConfigDetailsByConfigTemplateName(params.DeployPolicyID, param.ConfigFilesDetail)

	hostIDs := conv.SliceToSlice(params.Targets, func(target *types.Target) int64 {
		return target.Host.HostID
	})

	deployPolicyConfigMap, err := analyzer.listDeployPolicySubConfigMap(nCtx, param.PluginName, params.DeployPolicyID)
	if err != nil {
		return nil, err
	}
	hostIDs = append(hostIDs, conv.MapKeyToSlice(deployPolicyConfigMap)...)
	hostIDs = conv.SliceUnique(hostIDs)

	runningProcessMap, err := analyzer.listRunningProcessMap(nCtx, param.PluginName, hostIDs)
	if err != nil {
		return nil, err
	}
	desiredDetailMap, err := conv.SliceToMap(
		desiredDetails,
		func(detail *types.PluginConfigDetail) string { return detail.Name },
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plugin config details to map, count(%d): %w", len(desiredDetails), err)
	}
	targetMap := groupTargetsByHost(params.Targets)

	changeTasks := make([]*ChangeTask, 0)

	deleteDetailsByHost := collectSpecifyPluginSubConfigTemplateDeleteDetails(
		desiredDetailMap,
		targetMap,
		deployPolicyConfigMap,
		runningProcessMap,
		param.PluginName,
	)
	deleteTasks, err := genSpecifyPluginSubConfigTemplateDeleteTasks(
		params.DeployPolicyID,
		param.PluginName,
		param.CustomConfigContext,
		targetMap,
		deleteDetailsByHost,
		runningProcessMap,
	)
	if err != nil {
		return nil, err
	}
	changeTasks = append(changeTasks, deleteTasks...)

	applyTasks, err := genSpecifyPluginSubConfigTemplateApplyOrUpdateTasks(
		params.DeployPolicyID,
		param.PluginName,
		param.CustomConfigContext,
		params.Targets,
		desiredDetails,
		deployPolicyConfigMap,
		runningProcessMap,
	)
	if err != nil {
		return nil, err
	}
	changeTasks = append(changeTasks, applyTasks...)

	return changeTasks, nil
}

func genSpecifyPluginSubConfigTemplateApplyOrUpdateTasks(
	deployPolicyID int64,
	pluginName string,
	customConfigContext map[string]any,
	targets []*types.Target,
	desiredDetails []*types.PluginConfigDetail,
	deployPolicyConfigMap map[int64]map[string]*types.ProcessConfig,
	runningProcessMap map[string]*types.Process,
) ([]*ChangeTask, error) {
	changeTasks := make([]*ChangeTask, 0)
	for _, target := range targets {
		_, ok := runningProcessMap[genProcessUniqueID(target.Host.HostID, pluginName)]
		if !ok {
			continue
		}

		applyDetails := findApplyOrUpdateSubConfigDetails(
			desiredDetails,
			deployPolicyConfigMap[target.Host.HostID],
			customConfigContext,
		)
		if len(applyDetails) == 0 {
			continue
		}

		applySpec, err := newSpecifyPluginSubConfigSpec(pluginName, applyDetails, customConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create specify plugin sub config template apply spec: %w", err)
		}

		changeTasks = append(changeTasks, &ChangeTask{
			DeployPolicyID: deployPolicyID,
			Action:         ChangeActionPluginApplySubConfig,
			Spec:           applySpec,
			Target:         target,
			ConfigSource:   target,
		})
	}

	return changeTasks, nil
}

func genDeployPolicySubConfigDetailsByConfigTemplateName(
	deployPolicyID int64, details []*types.PluginConfigDetail,
) []*types.PluginConfigDetail {
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
			Name:         genDeployPolicySubConfigNameByConfigTemplateName(templateName, deployPolicyID),
			TemplateName: templateName,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
			FilePath:     detail.FilePath,
		})
	}

	return desiredDetails
}

func findApplyOrUpdateSubConfigDetails(
	desiredDetails []*types.PluginConfigDetail,
	currentConfigMap map[string]*types.ProcessConfig,
	customConfigContext map[string]any,
) []*types.PluginConfigDetail {
	applyDetails := make([]*types.PluginConfigDetail, 0)
	for _, detail := range desiredDetails {
		if detail == nil {
			continue
		}

		currentConfig, ok := currentConfigMap[detail.Name]
		if !ok || isSpecifyPluginSubConfigTemplateDeclarationChanged(detail, currentConfig, customConfigContext) {
			applyDetails = append(applyDetails, detail)
		}
	}

	return applyDetails
}

func isSpecifyPluginSubConfigTemplateDeclarationChanged(
	desiredDetail *types.PluginConfigDetail,
	currentConfig *types.ProcessConfig,
	customConfigContext map[string]any,
) bool {
	if currentConfig == nil {
		return true
	}

	if currentConfig.TemplateName != desiredDetail.TemplateName {
		return true
	}

	return !sameCustomConfigContext(currentConfig.CustomConfigContext, customConfigContext)
}

func sameCustomConfigContext(currentConfigContext, desiredConfigContext map[string]any) bool {
	if len(currentConfigContext) == 0 && len(desiredConfigContext) == 0 {
		return true
	}

	return reflect.DeepEqual(currentConfigContext, desiredConfigContext)
}

func collectSpecifyPluginSubConfigTemplateDeleteDetails(
	desiredDetailMap map[string]*types.PluginConfigDetail,
	targetMap map[int64]*types.Target,
	deployPolicyConfigMap map[int64]map[string]*types.ProcessConfig,
	runningProcessMap map[string]*types.Process,
	pluginName string,
) map[int64][]*types.PluginConfigDetail {
	deleteDetailsByHost := make(map[int64][]*types.PluginConfigDetail)
	seen := make(map[string]struct{})

	for hostID, currentConfigMap := range deployPolicyConfigMap {
		_, inDeployScope := targetMap[hostID]
		_, hasRunningProcess := runningProcessMap[genProcessUniqueID(hostID, pluginName)]
		if !inDeployScope || !hasRunningProcess {
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

func genSpecifyPluginSubConfigTemplateDeleteTasks(
	deployPolicyID int64,
	pluginName string,
	customConfigContext map[string]any,
	targetMap map[int64]*types.Target,
	deleteDetailsByHost map[int64][]*types.PluginConfigDetail,
	runningProcessMap map[string]*types.Process,
) ([]*ChangeTask, error) {
	changeTasks := make([]*ChangeTask, 0)
	for hostID, details := range deleteDetailsByHost {
		deleteSpec, err := newSpecifyPluginSubConfigSpec(pluginName, details, customConfigContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create specify plugin sub config template delete spec: %w", err)
		}

		action := ChangeActionPluginDeleteSubConfigRecord
		target := selectSubConfigTarget(hostID, targetMap)
		process, ok := runningProcessMap[genProcessUniqueID(hostID, pluginName)]
		if ok {
			action = ChangeActionPluginDeleteSubConfig
			target = selectSubConfigTemplateDeleteTarget(hostID, targetMap, process)
		}
		changeTasks = append(changeTasks, &ChangeTask{
			DeployPolicyID: deployPolicyID,
			Action:         action,
			Spec:           deleteSpec,
			Target:         target,
			ConfigSource:   target,
		})
	}

	return changeTasks, nil
}

func selectSubConfigTemplateDeleteTarget(
	hostID int64, targetMap map[int64]*types.Target, process *types.Process,
) *types.Target {
	if target, ok := targetMap[hostID]; ok {
		return target
	}

	return &types.Target{
		Host: types.Host{
			HostID:   hostID,
			TenantID: process.TenantID,
			Static: &types.HostStatic{
				BizID: process.BizID,
			},
		},
	}
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
