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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef"
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
	ExecuteOperation(name workflowdef.OperDefName, triggerID string, param *operengine.OperInstParam) error

	// RetryOperation retries the operation.
	RetryOperation(operationID string, param *operengine.OperInstParam) error
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

	mgr.logger.Info("successfully started manager")

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
	return mgr.operInstMgr.RegisterActions(
		workflowdef.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.logger),
		workflowdef.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		workflowdef.NewActionGenAllBizHostSyncOper(mgr.conf.TopoStorage, mgr.operMgr),
		workflowdef.NewActionSshHostExecCmd(mgr.conf.Crypter, mgr.logger),
	)
}

// ExecuteOperation execute an operation.
func (mgr *manager) ExecuteOperation(name workflowdef.OperDefName, triggerID string, param *operengine.OperInstParam) error {
	builder, ok := workflowdef.OperBuilderRegistry()[name]
	if !ok {
		return fmt.Errorf("operation builder not found, name: %s", name)
	}

	operation := builder(triggerID)
	err := mgr.operMgr.ExecuteOperation(operation, param)
	if err != nil {
		return fmt.Errorf("execute operation failed, name: %s, err: %v", name, err)
	}

	return nil
}

// RetryOperation ...
func (mgr *manager) RetryOperation(operationID string, param *operengine.OperInstParam) error {
	if len(operationID) == 0 {
		return errors.New("operation id is empty")
	}

	if param == nil {
		return errors.New("param is nil")
	}

	if err := mgr.operMgr.RetryOperation(operationID, param); err != nil {
		return err
	}

	return nil
}
