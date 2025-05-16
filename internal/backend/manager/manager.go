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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/nodedeployment"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// Manager defines the manager interface.
type Manager interface {
	// Start starts the manager.
	Start(ctx context.Context) error

	// CheckHealth checks the health of manager.
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	// LaunchSyncBizAndHost launch a task to sync biz and host. returns the trigger-id.
	LaunchSyncBizAndHost(ctx context.Context) (string, error)

	// LaunchSyncNetworkArea launch a task to sync networkarea. returns the trigger-id.
	LaunchSyncNetworkArea(ctx context.Context) (string, error)

	// LaunchInstallNode launch a task to install node. returns the workflow-id.
	LaunchInstallNode(ctx context.Context, nodeDeploys []*types.NodeDeployment) (string, error)
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

	workflowMgr        workflow.IManager
	iDaoNodeDeployment nodedeployment.IDomainInit

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

	return mgr.workflowMgr.RegisterActions(
		syncdata.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.logger),
		syncdata.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		syncdata.NewActionSyncNetworkAreaFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
		syncdata.NewActionGenOperSyncHost(mgr.conf.TopoStorage, mgr.workflowMgr),
	)
}

func (mgr *manager) registerActionDefNodeInstall() error {
	return mgr.workflowMgr.RegisterActions(
		nodeinstall.NewActionBindAgentHostRel(
			mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionInstallNodeBySSH(mgr.conf.InstallerFileGroup, mgr.conf.Crypter, mgr.logger,
			mgr.conf.NodeDeploymentStorage, mgr.conf.Provider),
		nodeinstall.NewActionWaitGseReady(
			mgr.conf.GSEHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionSyncNodeInfo(mgr.conf.GSEHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionPushHostIdentifier(mgr.conf.CmdbHandler, mgr.conf.NodeDeploymentStorage, mgr.logger),
		nodeinstall.NewActionRenderNodeDeployment(
			mgr.conf.NodeDeploymentStorage, mgr.conf.TopoStorage, mgr.conf.TopoStorage, mgr.logger),
		nodeinstall.NewActionUpsertHostToCMDB(
			mgr.conf.CmdbHandler, mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage),
		nodeinstall.NewActionWaitInstallComplete(mgr.conf.OperInstStorage, mgr.logger),
		nodeinstall.NewActionUpdateHost(mgr.conf.TopoStorage, mgr.conf.NodeDeploymentStorage, mgr.logger),
	)
}

// LaunchSyncBizAndHost launch a task to sync biz and host.
func (mgr *manager) LaunchSyncBizAndHost(ctx context.Context) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncBizAndHostFromCMDB(syncdata.SyncBizFromCMDBParam{TenantID: tenantID})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync biz and host task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncNetworkArea launch a task to sync networkarea.
func (mgr *manager) LaunchSyncNetworkArea(ctx context.Context) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncNetworkAreaFromCMDB(syncdata.SyncNetworkAreaFromCMDBParam{TenantID: tenantID})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync networkarea task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchInstallNode launch a task to install node.
func (mgr *manager) LaunchInstallNode(ctx context.Context, nodeDeploys []*types.NodeDeployment) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, nodeDeploy := range nodeDeploys {
		deploy := nodeDeploy

		gp.Go(func() error {
			if err := mgr.iDaoNodeDeployment.Create(ctx, deploy); err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to create node deployment. "+
						"tenant-id(%s), trigger-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), deploy.Token, err)
			}

			operationDef := nodeinstall.NewOperInstallNodeBySSH(nodeinstall.UpsertHostToCMDBParam{Token: deploy.Token})
			operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
			if err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to launch install node task. "+
						"tenant-id(%s), trigger-id(%s), operation-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token, err)

				return err
			}

			mgr.logger.InfoCtxf(ctx,
				"launched install node task. tenant-id(%s), trigger-id(%s), operation-id(%s), node-deployment-token(%s)",
				tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token)
			return nil
		})
	}
	gp.Wait()

	return triggerCtl.GetTriggerID(), nil
}
