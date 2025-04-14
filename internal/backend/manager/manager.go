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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// Manager defines the manager interface.
type Manager interface {
	// Start starts the manager.
	Start(ctx context.Context) error

	// CheckHealth checks the health of manager.
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	// ExecuteOperation executes the operation.
	ExecuteOperation(ctx context.Context, name workflowdef.OperDefName, param *operengine.OperInstParam) (string, error)

	// RetryOperation retries the operation.
	RetryOperation(ctx context.Context, operationID string, param *operengine.OperInstParam) error
}

// NewManager creates a new manager.
func NewManager(conf Config, logger logger.Logger) (Manager, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	mgr := &manager{
		logger:      logger,
		isRunning:   false,
		operInstMgr: nil,
		conf:        conf,
	}

	var err error
	mgr.operInstMgr, err = operengine.NewOperInstMgr(
		mgr.conf.WorkflowConfig.WorkNodeNum,
		operengine.WithRedis(
			mgr.conf.WorkflowConfig.Redis.Addr,
			mgr.conf.WorkflowConfig.Redis.Password,
			mgr.conf.WorkflowConfig.Redis.DB),
		mgr.conf.OperInstStorage,
		operengine.WithLogger(mgr.logger))
	if err != nil {
		return nil, err
	}

	mgr.operMgr, err = operengine.NewOperationMgr(
		mgr.operInstMgr,
		mgr.conf.OperStorage,
		operengine.OperMgrWithLogger(mgr.logger))
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

// manager provides to operate nodeman tasks.
type manager struct {
	logger logger.Logger

	// state
	isRunning bool

	operInstMgr operengine.OperInstMgr
	operMgr     operengine.OperationMgr

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

	if err := mgr.startOperEngineManager(ctx); err != nil {
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

	if mgr.operInstMgr == nil {
		return errors.New("task engine manager is not initialized")
	}

	if err := mgr.operInstMgr.CheckHealth(); err != nil {
		return fmt.Errorf("operation instance engine manager is unhealthy, err: %v", err)
	}

	return nil
}

// GracefulShutdown ...
func (mgr *manager) GracefulShutdown() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.operInstMgr.GracefulShutdown(); err != nil {
		return err
	}

	return nil
}

func (mgr *manager) startOperEngineManager(ctx context.Context) error {
	// TODO: implement me
	if err := mgr.registerActionDefs(); err != nil {
		return err
	}

	if err := mgr.operInstMgr.Start(ctx); err != nil {
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
		workflowdef.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.logger),
		workflowdef.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		workflowdef.NewActionSyncNetworkAreaFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		workflowdef.NewActionGenAllBizHostSyncOper(mgr.conf.TopoStorage, mgr.operMgr),
	)
}

func (mgr *manager) registerActionDefNodeInstall() error {
	return mgr.operInstMgr.RegisterActions(
		nodeinstall.NewActionBindAgentHostRel(
			mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionInstallAgentBySSH(mgr.conf.InstallerFileGroup, mgr.conf.Crypter, mgr.logger, mgr.conf.NodeDeploymentStorage, mgr.conf.Provider),
		nodeinstall.NewActionWaitGseRunning(
			mgr.conf.GSEHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionSyncNodeInfo(mgr.conf.GSEHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionPushHostIdentifier(mgr.conf.CmdbHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionRenderNodeDeployment(
			mgr.conf.NodeDeploymentStorage, mgr.conf.TopoStorage, mgr.conf.TopoStorage, mgr.logger),
		nodeinstall.NewActionUpsertHost(
			mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage),
		nodeinstall.NewActionWaitComplete(mgr.conf.OperInstStorage, mgr.logger),
	)
}

// ExecuteOperation execute an operation.
func (mgr *manager) ExecuteOperation(
	ctx context.Context, name workflowdef.OperDefName, param *operengine.OperInstParam) (string, error) {

	triggerID := identifier.GenTriggerID()
	mgr.logger.InfoCtxf(ctx, "try to execute operation. name(%s), trigger-id(%s), param(%v)",
		name, triggerID, param)

	builder, ok := workflowdef.OperBuilderRegistry()[name]
	if !ok {
		return triggerID, fmt.Errorf("operation builder not found, name: %s", name)
	}

	operation := builder(triggerID)
	err := mgr.operMgr.ExecuteOperation(ctx, operation, param)
	if err != nil {
		return triggerID, fmt.Errorf("execute operation failed, name: %s, err: %v", name, err)
	}

	mgr.logger.InfoCtxf(ctx, "dispatched execute operation. name(%s), trigger-id(%s), operation-id(%s)",
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
