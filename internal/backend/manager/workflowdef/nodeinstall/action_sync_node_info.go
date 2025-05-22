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
	"fmt"
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncNodeInfo defines the action name.
	ActionNameSyncNodeInfo = "sync_node_info"
)

// NewActionSyncNodeInfo get a new action.
func NewActionSyncNodeInfo(
	gseClient gse.IHandler,
	storage nodedeployment.IStorage,
	logger logger.Logger,
) action.Definition {

	return &actionSyncNodeInfo{
		gseClient: gseClient,
		storage:   storage,
		logger:    logger,
	}
}

// SyncNodeInfoParam ...
type SyncNodeInfoParam struct {
	Token string `json:"token"`
}

type actionSyncNodeInfo struct {
	gseClient gse.IHandler
	storage   nodedeployment.IStorage
	logger    logger.Logger
}

// Name returns the name of the action.
func (act *actionSyncNodeInfo) Name() string {
	return ActionNameSyncNodeInfo
}

// Version returns the version of the action.
func (act *actionSyncNodeInfo) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncNodeInfo) Description() string {
	return "sync node info to db"
}

// Timeout returns the timeout of the action.
func (act *actionSyncNodeInfo) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionSyncNodeInfo) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionSyncNodeInfo) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionSyncNodeInfo) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionSyncNodeInfo) Do(ctx *action.InstanceContext) error {
	param := new(SyncNodeInfoParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := act.storage.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	agentInfos, err := act.gseClient.ListAgentInfo(ctx.Ctx, info.Dynamic.AgentID)
	if err != nil {
		return err
	}

	if len(agentInfos) != 1 {
		return fmt.Errorf("get agent info error, agent-id(%s), aget-infos(%v)", info.Dynamic.AgentID, agentInfos)
	}

	agentInfo := agentInfos[0]
	info.Dynamic.NodeCPUArch, err = platform.NormalizeArch(agentInfo.Arch)
	if err != nil {
		return fmt.Errorf("normalize arch error, agent-id(%s), arch(%s), err(%v)", info.Dynamic.AgentID, agentInfo.Arch, err)
	}

	info.Dynamic.NodeOsType, err = platform.NormalizeOS(agentInfo.OSType)
	if err != nil {
		return fmt.Errorf("normalize os error, agent-id(%s), os-type(%s), err(%v)",
			info.Dynamic.AgentID, agentInfo.OSType, err)
	}

	if err := act.storage.UpdateInfo(ctx.Ctx, param.Token, info); err != nil {
		return fmt.Errorf("update info error, agent-id(%s), info(%v)", info.Dynamic.AgentID, err)
	}

	return nil
}
