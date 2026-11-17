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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (executor *Executor) executeChangeActionPluginApplyConfig(nCtx contextx.IContext, execution ExecutionParam,
	tasks []*ChangeTask) error {

	operator, err := access.GetVirtualUserBKUsername(nCtx)
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin apply config: %w", err)
	}

	pluginDeployments := make([]*types.PluginDeployment, len(tasks))
	hostMap := make(map[int64]struct{})
	for idx, task := range tasks {
		deployment, err := newPluginConfigDeployment(nCtx, task)
		if err != nil {
			return fmt.Errorf("failed to build plugin config deployment: %w", err)
		}
		pluginDeployments[idx] = deployment
		hostMap[task.Target.Host.HostID] = struct{}{}
	}

	workflowID, err := executor.pluginManager.LaunchApplyPluginConfig(nCtx, types.ApplyPluginConfigParam{
		Type:              types.PluginWorkflowTypeApplyPluginConfig,
		HostIDs:           conv.MapKeyToSlice(hostMap),
		BizIDs:            collectTargetBizIDs(tasks),
		Operator:          operator,
		DeployPolicyIDs:   collectDeployPolicyIDs(tasks),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		return fmt.Errorf("failed to execute change action plugin apply config: %w", err)
	}
	executor.recordWorkflowChild(nCtx, execution, tasks, types.WorkflowDomainPlugin, workflowID)

	return nil
}

func newPluginConfigDeployment(nCtx contextx.IContext, task *ChangeTask) (*types.PluginDeployment, error) {
	var pluginName, version string
	var customConfigContext map[string]any
	switch task.Spec.Type() {
	case types.DeploySpecTypeSpecifyPlugin:
		param, err := task.Spec.GetSpecifyPluginParam()
		if err != nil {
			return nil, fmt.Errorf("failed to get specify plugin param: %w", err)
		}
		pluginName, version, customConfigContext = param.PluginName, param.Version, param.CustomConfigContext
	case types.DeploySpecTypeSpecifyPluginPkg, types.DeploySpecTypeProjectPluginPkgToHosts:
		param, err := getPluginPkgTaskParam(task)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin pkg task param: %w", err)
		}
		pluginName, version, customConfigContext = param.pluginName, param.version, param.customConfigContext
	default:
		return nil, fmt.Errorf("unsupported plugin config spec type, type(%s)", task.Spec.Type())
	}

	return types.NewPluginDeployment(&types.PluginDeploymentInfo{
		Process: types.Process{
			TenantID:   nCtx.TenantID(),
			HostID:     task.Target.Host.HostID,
			PluginName: pluginName,
		},
		ConfigSource: getTaskConfigSource(task),
		InstallOptions: types.PluginDeploymentInstallOptions{
			Version: version,
		},
	}, &types.PluginDeploymentPluginConf{
		CustomConfigContext: customConfigContext,
	}), nil
}
