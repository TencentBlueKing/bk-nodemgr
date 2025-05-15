/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package manager provides handlers to manage all the nodeman task operations
package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/old/operengine"
)

// OperInst defines the operation instance interface.
type OperInst interface {
	OperDef() operengine.OperDefSnapshot
	Param() operengine.OperInstParam
}

// Manager defines the manager interface.
type Manager interface {
	// Start starts the manager.
	Start(ctx context.Context) error

	// CheckHealth checks the health of manager.
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	// Execute operation.
	Execute(ctx context.Context, operInst OperInst) (string, error)

	// RetryOperation retries the operation.
	RetryOperation(ctx context.Context, operationID string, param *operengine.OperInstParam) error
}

// NewManager creates a new manager.
func NewManager(conf Config, logger logger.Logger) (Manager, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	mgr := &manager{
		logger:    logger,
		isRunning: false,
		conf:      conf,
	}

	mgr.workflowMgr = workflow.NewManager(
		mgr.conf.WorkflowConfig.WorkNodeNum,
		nil,
		workflow.WithRedis(
			mgr.conf.WorkflowConfig.Redis.Addr,
			mgr.conf.WorkflowConfig.Redis.Password,
			mgr.conf.WorkflowConfig.Redis.DB),
		workflow.WithLogger(mgr.logger))

	return mgr, nil
}

// manager provides to operate nodeman tasks.
type manager struct {
	logger logger.Logger

	// state
	isRunning bool

	workflowMgr workflow.IManager

	// config
	conf Config
}

// Start starts the manager.
func (mgr *manager) Start(ctx context.Context) error {
	mgr.logger.Info("starting manager")

	if mgr.isRunning {
		return errors.New("manager already started")
	}

	if ctx == nil {
		return errors.New("context is nil")
	}

	if err := mgr.conf.Validate(); err != nil {
		return fmt.Errorf("config is invalid, err: %v", err)
	}

	if err := mgr.startWorkflowManager(ctx); err != nil {
		return err
	}

	mgr.isRunning = true

	mgr.logger.Info("started manager")

	return nil
}

// CheckHealth checks the health of manager.
func (mgr *manager) CheckHealth() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.conf.TopoStorage.CheckHealthz(); err != nil {
		return fmt.Errorf("topo storage is unhealthy, err: %v", err)
	}

	if mgr.workflowMgr == nil {
		return errors.New("workflow manager is not initialized")
	}

	if err := mgr.workflowMgr.CheckHealth(); err != nil {
		return fmt.Errorf("workflow manager is unhealthy, err: %v", err)
	}

	return nil
}

// GracefulShutdown ...
func (mgr *manager) GracefulShutdown() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.workflowMgr.GracefulShutdown(); err != nil {
		return err
	}

	return nil
}

func (mgr *manager) startWorkflowManager(ctx context.Context) error {
	// TODO: implement me
	if err := mgr.registerActionDefs(); err != nil {
		return err
	}

	if err := mgr.workflowMgr.Start(ctx); err != nil {
		return err
	}

	return nil
}

// registerActionDefs init action defs
func (mgr *manager) registerActionDefs() error {
	if err := mgr.registerActionDefNodeInstall(); err != nil {
		return fmt.Errorf("register action def node install failed, err: %v", err)
	}

	return mgr.operInstMgr.RegisterActions(
		syncdata.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.logger),
		syncdata.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		syncdata.NewActionSyncNetworkAreaFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		syncdata.NewActionGenAllBizHostSyncOper(mgr.conf.TopoStorage, mgr.workflowMgr),
	)
}

func (mgr *manager) registerActionDefNodeInstall() error {
	return mgr.workflowMgr.RegisterActions(
		nodeinstall.NewActionBindAgentHostRel(
			mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionInstallAgentBySSH(mgr.conf.InstallerFileGroup, mgr.conf.Crypter, mgr.logger,
			mgr.conf.NodeDeploymentStorage, mgr.conf.Provider),
		nodeinstall.NewActionWaitGseRunning(
			mgr.conf.GSEHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionSyncNodeInfo(mgr.conf.GSEHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionPushHostIdentifier(mgr.conf.CmdbHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionRenderNodeDeployment(
			mgr.conf.NodeDeploymentStorage, mgr.conf.TopoStorage, mgr.conf.TopoStorage, mgr.logger),
		nodeinstall.NewActionUpsertHost(
			mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage),
		nodeinstall.NewActionWaitComplete(mgr.conf.OperInstStorage, mgr.logger),
		nodeinstall.NewActionUpdateHost(mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage, mgr.logger),
	)
}

// Execute try to execute an operation.
func (mgr *manager) Execute(ctx context.Context, operInst OperInst) (string, error) {
	triggerID := identifier.GenTriggerID()
	def := operInst.OperDef()
	param := operInst.Param()
	mgr.logger.InfoCtxf(ctx, "try to execute operation. name(%s), trigger-id(%s), param(%v)",
		def.OperDefName, triggerID, param)

	operation := operengine.NewOperation(triggerID, def)
	err := mgr.operMgr.ExecuteOperation(ctx, operation, &param)
	if err != nil {
		return triggerID, fmt.Errorf("execute operation failed, name(%s), trigger-id(%s), err: %v",
			def.OperDefName, triggerID, err)
	}

	mgr.logger.InfoCtxf(ctx, "dispatched execute operation. name(%s), trigger-id(%s), operation-id(%s)",
		def.OperDefName, triggerID, operation.OperationID)

	return triggerID, nil
}

// ExecuteOperations execute operations.
func (mgr *manager) ExecuteOperations(
	ctx context.Context, name workflowdef.OperDefName, params []*operengine.OperInstParam) (string, error) {

	triggerID := identifier.GenTriggerID()
	mgr.logger.InfoCtxf(ctx, "try to execute operations. name(%s), trigger-id(%s), params(%d)",
		name, triggerID, params)

	builder, ok := workflowdef.OperBuilderRegistry()[name]
	if !ok {
		return triggerID, fmt.Errorf("operation builder not found, name: %s", name)
	}

	operation := builder(triggerID)
	err := mgr.operMgr.ExecuteOperation(ctx, operation, params)
	if err != nil {
		return triggerID, fmt.Errorf("execute operations failed, name: %s, err: %v", name, err)
	}

	mgr.logger.InfoCtxf(ctx, "dispatched execute operations. name(%s), trigger-id(%s), operation-id(%s)",
		name, triggerID, operation.OperationID)

	return triggerID, nil
}

// RetryOperation ...
func (mgr *manager) RetryOperation(ctx context.Context, operationID string, param *operengine.OperInstParam) error {
	if len(operationID) == 0 {
		return errors.New("operation id is empty")
	}

	if param == nil {
		return errors.New("param is nil")
	}

	mgr.logger.InfoCtxf(ctx, "try to retry operation. operation-id(%s), param(%v)", operationID, param)

	if err := mgr.operMgr.RetryOperation(ctx, operationID, param); err != nil {
		return err
	}

	mgr.logger.InfoCtxf(ctx, "dispatched retry operation. operation-id(%s)", operationID)
	return nil
}
