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

// Package manager provides handlers to manage all the nodeman task operations
package manager

import (
	"errors"
	"fmt"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
)

// IManager defines the Manager interface.
// nolint: interfacebloat
type IManager interface {
	// Start starts the Manager.
	Start(ctx contextx.IContext) error

	// CheckHealth checks the health of Manager.
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	managerIface.ISyncManager
	managerIface.INodeManager
	managerIface.IPluginManager
	managerIface.IDeployPolicyManager
	managerIface.IPackageManager
}

// NewManager creates a new Manager.
func NewManager(conf Config) (*Manager, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	mgr := &Manager{
		isRunning: false,
		conf:      conf,
	}

	var err error
	mgr.workflowMgr, err = workflow.NewManager(
		mgr.conf.WorkNodeNum,
		workflow.WithStorageTrigger(conf.StorageWorkflow),
		workflow.WithStorageOperation(conf.StorageWorkflow),
		workflow.WithStorageOperationInstance(conf.StorageWorkflow),
		workflow.WithStorageActionInstance(conf.StorageWorkflow),
		workflow.WithLocker(conf.LockerFactory),
		workflow.WithRedis(mgr.conf.Redis),
		workflow.WithTraceService(mgr.conf.TraceService),
		workflow.WithGracefulShutdownTimeout(mgr.conf.GracefulShutdownTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow manager: %w", err)
	}

	return mgr, nil
}

var _ IManager = &Manager{}

// Manager provides to operate nodeman tasks.
type Manager struct {
	// state
	isRunning bool

	workflowMgr workflow.IManager

	// config
	conf Config
}

// Start starts the manager.
func (mgr *Manager) Start(ctx contextx.IContext) error {
	logger.G.Sys().Info("starting manager")

	if mgr.isRunning {
		return errors.New("manager already started")
	}

	if ctx == nil {
		return errors.New("context is nil")
	}

	if err := mgr.conf.Validate(); err != nil {
		return fmt.Errorf("config is invalid: %v", err)
	}

	if err := mgr.startWorkflowManager(ctx); err != nil {
		return err
	}

	if err := mgr.startMonitoringScheduledWorkflow(ctx); err != nil {
		return err
	}

	mgr.isRunning = true

	logger.G.Sys().Info("started manager")

	return nil
}

// CheckHealth checks the health of manager.
func (mgr *Manager) CheckHealth() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.conf.StorageTopo.CheckHealthz(); err != nil {
		return fmt.Errorf("topo storage is unhealthy: %v", err)
	}

	if mgr.workflowMgr == nil {
		return errors.New("workflow manager is not initialized")
	}

	if err := mgr.workflowMgr.CheckHealth(); err != nil {
		return fmt.Errorf("workflow manager is unhealthy: %v", err)
	}

	return nil
}

// GracefulShutdown ...
func (mgr *Manager) GracefulShutdown() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.workflowMgr.GracefulShutdown(); err != nil {
		return err
	}

	return nil
}

func (mgr *Manager) startWorkflowManager(ctx contextx.IContext) error {
	if err := mgr.registerDefinitions(); err != nil {
		return err
	}

	if err := mgr.workflowMgr.Start(ctx); err != nil {
		return err
	}

	return nil
}
