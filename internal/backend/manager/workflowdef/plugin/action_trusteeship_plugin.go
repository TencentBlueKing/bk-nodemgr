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
	"errors"
	"fmt"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTrusteeshipPlugin defines the action name.
	ActionNameTrusteeshipPlugin = "trusteeship_plugin"
)

// NewActionTrusteeshipPlugin ...
func NewActionTrusteeshipPlugin(capability *Capability) action.Definition {
	return &actTrusteeshipPlugin{
		gseHandlerProc: capability.GSEHandler,
	}
}

// ActionParamTrusteeshipPlugin ...
type ActionParamTrusteeshipPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actTrusteeshipPlugin ...
type actTrusteeshipPlugin struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actTrusteeshipPlugin) Name() string {
	return ActionNameTrusteeshipPlugin
}

// Version returns the version of the action.
func (act *actTrusteeshipPlugin) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actTrusteeshipPlugin) Description() string {
	return "trusteeship plugin to gse."
}

// Timeout returns the timeout of the action.
func (act *actTrusteeshipPlugin) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actTrusteeshipPlugin) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actTrusteeshipPlugin) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actTrusteeshipPlugin) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actTrusteeshipPlugin) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamTrusteeshipPlugin)
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

	nCtx := std.Context()
	processSpec := std.DeployInfo().Plugin.Static.Spec

	result, err := act.gseHandlerProc.TrusteeshipProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to trusteeship plugin: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("successfully execute trusteeship plugin operate, result(%s)", result))

	ctx.Data.LogI("wait process running")
	expoBackoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

	var processInfo *types.ProcessInfo
	err = expoBackoff.Do(nCtx, func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.Identity.Name, processSpec.AgentID)
		if err != nil {
			return fmt.Errorf("failed to query process info: %w", err)
		}

		if processInfo.Status != types.ProcessStatusRunning {
			ctx.Data.LogI(fmt.Sprintf("process status is not running, status(%s)", processInfo.Status))

			return fmt.Errorf("process status is not running, status(%s)", processInfo.Status)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait process running: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("process running, info(%+v)", processInfo))

	return nil
}
