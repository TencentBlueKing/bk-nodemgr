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
	"math/rand"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSelectRelayHost defines the action name.
	ActionNameSelectRelayHost = "select_relay_host"
)

// NewActionSelectRelayHost get a new action.
func NewActionSelectRelayHost(capability *Capability) action.Definition {
	return &actionSelectRelayHost{
		storageHost:           capability.StorageTopo,
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamSelectRelayHost ...
type ActParamSelectRelayHost struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// actionSelectRelayHost ...
type actionSelectRelayHost struct {
	storageHost           topoStg.IStorageHost
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionSelectRelayHost) Name() string {
	return ActionNameSelectRelayHost
}

// Version returns the version of the action.
func (act *actionSelectRelayHost) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSelectRelayHost) Description() string {
	return "select relay host by random"
}

// Timeout returns the timeout of the action.
func (act *actionSelectRelayHost) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionSelectRelayHost) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionSelectRelayHost) MaxRetryCount() uint {
	return 3 //nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionSelectRelayHost) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionSelectRelayHost) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamSelectRelayHost)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		err = fmt.Errorf("failed to convert param: %w", err)

		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	relayHost, err := act.selectDedicatedInstallerHost(std)
	if err != nil {
		return err
	}

	std.DeployInfo().RelayInfo = relayHost

	std.InstanceData().LogI(fmt.Sprintf("select relay host success. relay-host-id(%d)", relayHost.HostID))

	return nil
}

func (act *actionSelectRelayHost) selectDedicatedInstallerHost(
	std *nodeUtils.NodeActionStandarder) (types.RelayInfo, error) {

	// for agent, we select relay host by random in its own network unit.
	networkunitID := std.DeployInfo().Host.Dynamic.NetworkUnitID
	// for proxy, we should use user selected of origin network unit id to select relay host.
	if std.DeployInfo().Host.Dynamic.NodeRole == types.NodeRoleProxy {
		networkunitID = std.DeployInfo().Host.Dynamic.ProxyInstallOriginUnitID
	}

	hosts, num, err := act.storageHost.ListHost(std.Context(), types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			NetworkUnitID: []int64{networkunitID},
			NodeRole:      []types.NodeRole{types.NodeRoleProxy},
			NodeStatus:    []types.NodeStatus{types.NodeStatusRunning},
		},
	})
	if err != nil {
		return types.RelayInfo{}, err
	}

	if num == 0 {
		std.InstanceData().LogE(fmt.Sprintf("no proxy host in network unit. network-unit-id(%d)",
			std.DeployInfo().Host.Dynamic.NetworkUnitID))

		return types.RelayInfo{}, errors.New("no proxy host in network unit")
	}

	dedicatedHosts := make([]*types.Host, 0, num)
	for _, host := range hosts {
		for _, tag := range host.Dynamic.ProxyTags {
			if tag == types.ProxyTagDedicatedInstaller {
				dedicatedHosts = append(dedicatedHosts, host)
				break
			}
		}
	}

	if len(dedicatedHosts) == 0 {
		std.InstanceData().LogE("no dedicated installer host")
		return types.RelayInfo{}, errors.New("no dedicated installer host")
	}

	// this just is a simple random selector.
	// nolint: gosec
	relayHost := dedicatedHosts[rand.Intn(len(dedicatedHosts))]

	return types.RelayInfo{
		HostID:          relayHost.HostID,
		AgentID:         relayHost.Dynamic.AgentID,
		InnerIP:         relayHost.Static.InnerIP,
		DownloadSvcPort: relayHost.Dynamic.RelayDownloadPort,
		CallbackSvcPort: relayHost.Dynamic.RelayCallbackPort,
	}, nil
}
