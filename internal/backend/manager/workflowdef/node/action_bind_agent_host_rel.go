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
	"context"
	"fmt"
	"time"

	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameBindAgentHostRel defines the action name.
	ActionNameBindAgentHostRel = "bind_agent_host_rel"
)

// NewActionBindAgentHostRel get a new action.
func NewActionBindAgentHostRel(
	bindHostAgent cmdb.IBindHostAgent,
	storageHost topo.IStorageHost,
	storageNodeDeployment nodeStg.IDaoNodeDeployment,
	logger logger.ILogger) action.Definition {

	return &actionBindAgentHostRel{
		IBindHostAgent:        bindHostAgent,
		storageHost:           storageHost,
		storageNodeDeployment: storageNodeDeployment,
		logger:                logger,
	}
}

// ActParamBindAgentHostRel ...
type ActParamBindAgentHostRel struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

type actionBindAgentHostRel struct {
	cmdb.IBindHostAgent
	storageHost           topo.IStorageHost
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	logger                logger.ILogger
}

// Name returns the name of the action.
func (act *actionBindAgentHostRel) Name() string {
	return ActionNameBindAgentHostRel
}

// Version returns the version of the action.
func (act *actionBindAgentHostRel) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionBindAgentHostRel) Description() string {
	return "bind agent host relation"
}

// Timeout returns the timeout of the action.
func (act *actionBindAgentHostRel) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionBindAgentHostRel) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionBindAgentHostRel) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionBindAgentHostRel) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionBindAgentHostRel) Do(ctx *action.InstanceContext) error {
	param := new(ActParamBindAgentHostRel)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := act.storageNodeDeployment.GetNodeDeploymentInfo(ctx.Ctx, param.Token)
	if err != nil {
		return fmt.Errorf("get node deployment info failed, err: %w", err)
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, info.Host.TenantID)
	if err != nil {
		return err
	}

	// this is a special case, when the deployment is reverted, the host id is not in the host table.
	if err := act.checkHostExist(tenantCtx, info); err != nil {
		return err
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		tenantUserCtx := contextx.NewTenantUserContext(ctx.Ctx, info.Host.TenantID, param.Operator)
		if err := act.BindHostAgent(tenantUserCtx, &info.Host); err != nil {
			return err
		}

		act.logger.Infof("successfully bind host agent relation to cmdb, host-id(%d), agent-id(%s)",
			info.Host.HostID, info.Host.Dynamic.AgentID)

		return nil
	})

	gp.Go(func() error {
		if err := act.storageHost.UpdateManyHostDynamic(tenantCtx, &info.Host); err != nil {
			return err
		}

		act.logger.Infof("successfully bind host agent relation to db, host-id(%d), agent-id(%s)",
			info.Host.HostID, info.Host.Dynamic.AgentID)

		return nil
	})

	if err = gp.Wait(); err != nil {
		return fmt.Errorf("bind host agent relation failed, err: %w", err)
	}

	ctx.Data.LogI(fmt.Sprintf("successfully bind agent host rel, host-id(%d), agent-id(%s)", info.Host.HostID,
		info.Host.Dynamic.AgentID))

	return nil
}

func (act *actionBindAgentHostRel) checkHostExist(ctx context.Context, info *types.DeploymentInfo) error {
	daoHost, err := act.storageHost.GetHostByID(ctx, info.Host.HostID)
	if err != nil {
		return fmt.Errorf("get host info failed, err: %w", err)
	}

	if daoHost == nil {
		return fmt.Errorf("host not found, host-id(%d)", info.Host.HostID)
	}

	return nil
}
