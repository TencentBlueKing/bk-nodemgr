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
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameStopPluginV2Process defines the action name for stopping plugin v2 process.
	ActionNameStopPluginV2Process = "stop_plugin_v2_process"

	pollingInterval = 10 * time.Second
)

// NewActionStopPluginV2Process creates an action to stop plugin v2 process.
func NewActionStopPluginV2Process(capability *Capability) action.Definition {
	return &actionStopPluginV2Process{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		pluginMgrIface:      capability.PluginIface,
		daoActionInstance:   capability.StorageWorkflow,
		daoPluginWorkflow:   capability.StoragePlugin,
	}
}

// ActParamStopPluginV2Process defines the parameters for actionStopPluginV2Process.
type ActParamStopPluginV2Process struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionStopPluginV2Process struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	daoHost             topoStg.IStorageHost
	pluginMgrIface      managerIface.IPluginManager
	daoActionInstance   workflow.IStorageActionInstance
	daoPluginWorkflow   pluginStg.IDaoPluginWorkflow
}

// Name returns the name of the action.
func (act *actionStopPluginV2Process) Name() string {
	return ActionNameStopPluginV2Process
}

// Version returns the version of the action.
func (act *actionStopPluginV2Process) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionStopPluginV2Process) Description() string {
	return "stop plugin v2 process"
}

// Timeout returns the timeout of the action.
func (act *actionStopPluginV2Process) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionStopPluginV2Process) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionStopPluginV2Process) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionStopPluginV2Process) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionStopPluginV2Process) DisplayNameZh() string {
	return "停止 V2 插件进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionStopPluginV2Process) DisplayNameEn() string {
	return "Stop V2 plugin process"
}

// Do this func define what the action will do.
func (act *actionStopPluginV2Process) Do(ctx *action.InstanceContext) error {
	param := new(ActParamStopPluginV2Process)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	workflowID, err := act.launchStopPluginV2Workflow(std)
	if err != nil {
		std.InstanceData().Log().
			Zh("停止 V2 插件进程失败, 错误(%v)", err).
			En("stop v2 plugin process failed, error(%v)", err).
			Error()

		return err
	}

	subWorkflowRefs := []types.SubWorkflowRef{
		{
			WorkflowID:     workflowID,
			WorkflowDomain: types.WorkflowDomainPlugin,
		},
	}

	if err := act.waitWorkflow(std, subWorkflowRefs); err != nil {
		return err
	}

	return nil
}

func (act *actionStopPluginV2Process) launchStopPluginV2Workflow(std *pluginUtils.PluginActionStandarder) (string, error) {
	nCtx := std.Context()
	deployInfo := std.DeployInfo()

	pluginName := deployInfo.Process.PluginName
	std.InstanceData().Log().
		Zh("开始停止 V2 插件进程(%v)", pluginName).
		En("start to stop v2 plugin process(%v)", pluginName).
		Info()

	stopProcessParam := types.StopProcessParam{
		Type:     types.PluginWorkflowTypeStopV2,
		HostIDs:  []int64{deployInfo.Process.HostID},
		BizIDs:   []int64{deployInfo.Process.BizID},
		Operator: std.Operator(),
		PluginDeployments: []*types.PluginDeployment{
			// stop plugin no need plugin conf.
			types.NewPluginDeployment(deployInfo, nil),
		},
	}

	workflowID, err := act.pluginMgrIface.LaunchStopPluginV2(nCtx, stopProcessParam)
	if err != nil {
		return "", fmt.Errorf("failed to launch stop plugin v2 process workflow: %w", err)
	}

	return workflowID, nil
}

func (act *actionStopPluginV2Process) saveSubWorkflowRefs(
	nCtx contextx.IContext,
	operInstID string,
	serializedRefs string,
) error {

	return act.daoActionInstance.UpsertActionInstancePrivateData(
		nCtx,
		operInstID,
		ActionNameStopPluginV2Process,
		map[string]any{
			types.PDKeySubWorkflowRefs: serializedRefs,
		},
	)
}

func (act *actionStopPluginV2Process) waitWorkflow(std *pluginUtils.PluginActionStandarder,
	subWorkflowRefs []types.SubWorkflowRef) error {

	nCtx := std.Context()
	serializedSubWorkflowRefs, err := types.SerializeSubWorkflowRefs(subWorkflowRefs)
	if err != nil {
		return fmt.Errorf("failed to serialize sub workflow refs: %w", err)
	}

	std.InstanceData().PrivateData[types.PDKeySubWorkflowRefs] = serializedSubWorkflowRefs
	if err = act.saveSubWorkflowRefs(
		nCtx,
		std.InstanceData().OperationInstanceID,
		serializedSubWorkflowRefs,
	); err != nil {
		return fmt.Errorf("failed to save sub workflow refs to private data: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功启动停止 V2 插件工作流, workflow(%+v)", subWorkflowRefs).
		En("succeeded in launching stop v2 plugin process workflow, workflow(%+v)", subWorkflowRefs).
		Info()

	std.InstanceData().Log().
		Zh("等待工作流完成").
		En("wait for workflow to finish").
		Info()

	gp := gopool.NewPool()
	for idx := range subWorkflowRefs {
		subWorkflowRef := subWorkflowRefs[idx]
		gp.Go(func() error {
			return act.checkPluginWorkflowStatus(std, subWorkflowRef)
		})
	}
	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to wait workflow: %w", err)
	}

	return nil
}

func (act *actionStopPluginV2Process) checkPluginWorkflowStatus(
	std *pluginUtils.PluginActionStandarder, subWorkflowRef types.SubWorkflowRef) error {

	workflowID := subWorkflowRef.WorkflowID

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: pollingInterval,
	})

	nCtx := std.Context()

	var workflowStatus types.PluginWorkflowStatus
	err := polling.Do(nCtx, func(_ int) error {
		var err error
		workflowStatus, err = act.daoPluginWorkflow.GetPluginWorkflowStatus(nCtx, workflowID)
		if err != nil {
			return fmt.Errorf("failed to get plugin workflow status: %w", err)
		}

		if workflowStatus == types.PluginWorkflowStatusRunning {
			std.InstanceData().Log().
				Zh("停止 V2 插件工作流仍在运行中, workflow-id: %s", workflowID).
				En("stop v2 plugin process workflow is still running, workflow-id: %s", workflowID).
				Info()

			return fmt.Errorf("plugin workflow is still running, workflow-id(%s)", workflowID)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("stop v2 plugin process workflow polling failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("停止 V2 插件工作流已结束, workflow-id: %s, 状态: %s", workflowID, workflowStatus).
		En("stop v2 plugin process workflow finished, workflow-id: %s, status: %s", workflowID, workflowStatus).
		Info()

	switch workflowStatus {
	case types.PluginWorkflowStatusSuccess:
		std.InstanceData().Log().
			Zh("停止 V2 插件成功, workflow-id: %s", workflowID).
			En("stop v2 plugin process succeeded, workflow-id: %s", workflowID).
			Info()

		return nil
	case types.PluginWorkflowStatusFailed:
		std.InstanceData().Log().
			Zh("停止 V2 插件失败, workflow-id: %s", workflowID).
			En("stop v2 plugin process failed, workflow-id: %s", workflowID).
			Info()

		return errors.New("stop v2 plugin process failed")
	case types.PluginWorkflowStatusPartialFailed:
		std.InstanceData().Log().
			Zh("停止 V2 插件部分失败, workflow-id: %s", workflowID).
			En("stop v2 plugin process partially failed, workflow-id: %s", workflowID).
			Info()

		return errors.New("stop v2 plugin process partially failed")
	default:
		return fmt.Errorf("unknown plugin workflow status: %s", workflowStatus)
	}
}
