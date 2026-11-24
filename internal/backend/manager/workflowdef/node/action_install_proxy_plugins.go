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

package node

import (
	"errors"
	"fmt"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNameInstallProxyPlugins defines the action name for installing Proxy-related plugins.
const ActionNameInstallProxyPlugins = "install_proxy_plugins"

// NewActionInstallProxyPlugins creates an action that installs only Proxy-related plugins.
func NewActionInstallProxyPlugins(capability *Capability) action.Definition {
	return &actionInstallProxyPlugins{
		actionInstallPreOrderedPlugins: newActionInstallPreOrderedPlugins(capability),
	}
}

type actionInstallProxyPlugins struct {
	*actionInstallPreOrderedPlugins
}

// Name returns the name of the action.
func (act *actionInstallProxyPlugins) Name() string {
	return ActionNameInstallProxyPlugins
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallProxyPlugins) DisplayNameZh() string {
	return "安装 Proxy 相关插件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallProxyPlugins) DisplayNameEn() string {
	return "Install Proxy Plugins"
}

// Description returns the description of the action.
func (act *actionInstallProxyPlugins) Description() string {
	return "Install Proxy-related plugins on target node."
}

// DelayFn returns the retry delay function.
func (act *actionInstallProxyPlugins) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do installs the relay plugin without installing pre-ordered monitoring plugins.
func (act *actionInstallProxyPlugins) Do(ctx *action.InstanceContext) (err error) {
	param := new(nodeUtils.NodeActionStandardParam)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to decode Proxy plugin install parameters: %w", err)
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err := std.Initialize(ctx, *param); err != nil {
		return fmt.Errorf("failed to initialize Proxy plugin installation: %w", err)
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	workflowID, err := act.installPlugins(std, []string{"bk-nodemgr-relay"})
	if err != nil {
		return fmt.Errorf("failed to install Proxy plugins: %w", err)
	}

	subWorkflowRefs := []types.SubWorkflowRef{{
		WorkflowID:     workflowID,
		WorkflowDomain: types.WorkflowDomainPlugin,
	}}
	if _, err := act.waitWorkflow(std, subWorkflowRefs); err != nil {
		return fmt.Errorf("failed to wait for Proxy plugin installation: %w", err)
	}

	return nil
}
