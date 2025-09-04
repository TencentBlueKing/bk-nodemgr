/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall/utils"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSelectRelayHost defines the action name.
	ActionNameSelectRelayHost = "select_relay_host"
)

// NewActionSelectRelayHost get a new action.
func NewActionSelectRelayHost(
	storageHost topo.IStorageHost,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
	logger logger.ILogger,
) action.Definition {

	return &actionSelectRelayHost{
		storageHost:           storageHost,
		storageNodeDeployment: storageNodeDeployment,
		logger:                logger,
	}
}

// ActParamSelectRelayHost ...
type ActParamSelectRelayHost struct {
	utils.NodeActionStandardParam `json:",inline"`
}

// actionSelectRelayHost ...
type actionSelectRelayHost struct {
	storageHost           topo.IStorageHost
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
	logger                logger.ILogger
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
	std := utils.NewNodeActionStandarder(act.storageNodeDeployment)
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

	ctx.Data.LogI(fmt.Sprintf("select relay host success. host-id(%d)", relayHost.HostID))

	return nil
}

func (act *actionSelectRelayHost) selectDedicatedInstallerHost(
	std *utils.NodeActionStandarder) (types.RelayInfo, error) {

	hosts, num, err := act.storageHost.ListHost(std.Context(), types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			NetworkUnitID: []int64{std.DeployInfo().Host.Dynamic.NetworkUnitID},
			NodeRole:      []types.NodeRole{types.NodeRoleProxy},
			NodeStatus:    []types.NodeStatus{types.NodeStatusRunning},
		},
	})
	if err != nil {
		return types.RelayInfo{}, err
	}

	if num == 0 {
		std.InstanceData().LogE("no proxy host in network unit")
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
		FileSvcPort:     relayHost.Dynamic.RelayFilePort,
		CallbackSvcPort: relayHost.Dynamic.RelayCallbackPort,
	}, nil
}
