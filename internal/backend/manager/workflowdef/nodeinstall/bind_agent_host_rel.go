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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionBindAgentHostRel ...
func NewActionBindAgentHostRel(cmdbHandler cmdb.IHandler, hostDao topo.IDaoHost,
	nodeDeploymentDao nodedeployment.IDaoNodeDeployment, logger logger.Logger) *BindAgentHostRel {

	return &BindAgentHostRel{
		cmdbHandler:       cmdbHandler,
		hostDao:           hostDao,
		nodeDeploymentDao: nodeDeploymentDao,
		logger:            logger,
	}
}

// BindAgentHostRelParam ...
type BindAgentHostRelParam struct {
	Token string `json:"token"`
}

// BindAgentHostRel ...
type BindAgentHostRel struct {
	cmdbHandler       cmdb.IHandler
	hostDao           topo.IDaoHost
	nodeDeploymentDao nodedeployment.IDaoNodeDeployment
	logger            logger.Logger
}

// Name returns the name of the action.
func (action *BindAgentHostRel) Name() string {
	return ActionNameBindAgentHostRel
}

// Version returns the version of the action.
func (action *BindAgentHostRel) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (action *BindAgentHostRel) Description() string {
	return "bind agent host relation"
}

// Timeout returns the timeout of the action.
func (action *BindAgentHostRel) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (action *BindAgentHostRel) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the max retry count of the action.
func (action *BindAgentHostRel) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (action *BindAgentHostRel) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (action *BindAgentHostRel) Do(ctx *operengine.ActionInstContext) error {
	param := new(BindAgentHostRelParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := action.nodeDeploymentDao.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return fmt.Errorf("get node deployment info failed, err: %w", err)
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, info.TenantID)
	if err != nil {
		return err
	}

	host, err := action.hostDao.GetHostByID(tenantCtx, info.HostID)
	if err != nil {
		return fmt.Errorf("get host info failed, err: %w", err)
	}

	host.Dynamic = &types.HostDynamic{
		NodeRole:       info.NodeRole,
		NodeStatus:     info.NodeStatus,
		NodeVersion:    info.NodeVersion,
		NodeGeneration: info.NodeGeneration,
		AgentID:        info.AgentID,
		NetworkUnitID:  info.NetworkUnitID,
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		if err := action.cmdbHandler.BindHostAgent(tenantCtx, host); err != nil {
			return err
		}

		action.logger.Infof("successfully bind host agent relation to cmdb, host-id(%d), agent-id(%s)",
			host.HostID, host.Dynamic.AgentID)

		return nil
	})

	gp.Go(func() error {
		if err := action.hostDao.UpdateManyHostDynamic(tenantCtx, host); err != nil {
			return err
		}

		action.logger.Infof("successfully bind host agent relation to db, host-id(%d), agent-id(%s)",
			host.HostID, host.Dynamic.AgentID)

		return nil
	})

	if err = gp.Wait(); err != nil {
		return fmt.Errorf("bind host agent relation failed, err: %w", err)
	}

	ctx.Data.Log(fmt.Sprintf("successfully bind agent host rel, host-id(%d), agent-id(%s)", host.HostID,
		host.Dynamic.AgentID))

	return nil
}
