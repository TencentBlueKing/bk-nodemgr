/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameAssignProxyInfo defines the action name.
	ActionNameAssignProxyInfo = "assign_proxy_info"

	// Default relay ports for proxy.
	defaultRelayCallbackPort = 28302
	defaultRelayDownloadPort = 28303
)

// NewActionAssignProxyInfo creates a new action.
func NewActionAssignProxyInfo(capability *Capability) action.Definition {
	return &actionAssignProxyInfo{
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamAssignProxyInfo defines the action parameter.
type ActParamAssignProxyInfo struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
	NetworkUnitID                     int64    `json:"network_unit_id"`
	RelayCallbackPort                 int64    `json:"relay_callback_port"`
	RelayDownloadPort                 int64    `json:"relay_download_port"`
	ProxyTags                         []string `json:"proxy_tags"`
}

type actionAssignProxyInfo struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionAssignProxyInfo) Name() string {
	return ActionNameAssignProxyInfo
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionAssignProxyInfo) DisplayNameZh() string {
	return "分配Proxy管控单元"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionAssignProxyInfo) DisplayNameEn() string {
	return "Assign Proxy Network Unit"
}

// Version returns the version of the action.
func (act *actionAssignProxyInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionAssignProxyInfo) Description() string {
	return "assign network unit and update proxy deployment info"
}

// Timeout returns the timeout of the action.
func (act *actionAssignProxyInfo) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionAssignProxyInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionAssignProxyInfo) MaxRetryCount() uint {
	return 3 //nolint: mnd
}

// DelayFn defines the delay before retrying on failure.
func (act *actionAssignProxyInfo) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do executes the action.
func (act *actionAssignProxyInfo) Do(ctx *action.InstanceContext) error {
	param := new(ActParamAssignProxyInfo)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// Initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, nil)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	deployInfo := std.DeployInfo()

	// Update network unit ID.
	deployInfo.Host.Dynamic.NetworkUnitID = param.NetworkUnitID

	// Set relay ports with defaults if not provided.
	if param.RelayCallbackPort > 0 {
		deployInfo.Host.Dynamic.RelayCallbackPort = param.RelayCallbackPort
	} else {
		deployInfo.Host.Dynamic.RelayCallbackPort = defaultRelayCallbackPort
	}

	if param.RelayDownloadPort > 0 {
		deployInfo.Host.Dynamic.RelayDownloadPort = param.RelayDownloadPort
	} else {
		deployInfo.Host.Dynamic.RelayDownloadPort = defaultRelayDownloadPort
	}

	// Set proxy tags.
	if len(param.ProxyTags) > 0 {
		deployInfo.Host.Dynamic.ProxyTags = types.StringListToProxyTagList(param.ProxyTags)
	} else {
		deployInfo.Host.Dynamic.ProxyTags = []types.ProxyTag{}
	}

	// Enable offline mode for plugin installation.
	deployInfo.InstallOptions.IsOffline = true

	// Save updated deployment info.
	if err := act.storageNodeDeployment.UpdateNodeDeploymentInfo(std.Context(), std.Token(), deployInfo); err != nil {
		return fmt.Errorf("failed to update node deployment info: %w", err)
	}

	std.InstanceData().Log().
		Zh("分配Proxy管控单元成功, 主机ID(%d), 管控单元ID(%d)", deployInfo.Host.HostID, param.NetworkUnitID).
		En("successfully assigned proxy network unit, host-id(%d), unit-id(%d)", deployInfo.Host.HostID, param.NetworkUnitID).
		Info()

	return nil
}
