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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/nodeinstall"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// Manager defines the manager interface.
// nolint: interfacebloat
type Manager interface {
	// Start starts the manager.
	Start(ctx context.Context) error

	// CheckHealth checks the health of manager.
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	// LaunchSyncBizAndHost launch a task to sync biz and host. returns the trigger-id.
	LaunchSyncBizAndHost(ctx context.Context) (string, error)

	// LaunchSyncHostByBizID launch a task to sync host by biz-id. returns the trigger-id.
	LaunchSyncHostByBizID(ctx context.Context, bizID int64) (string, error)

	// LaunchSyncNetworkArea launch a task to sync networkarea. returns the trigger-id.
	LaunchSyncNetworkArea(ctx context.Context) (string, error)

	// LaunchInstallNode launch a task to install node. returns the workflow-id.
	LaunchInstallNode(ctx context.Context, param InstallNodeParam) (string, error)

	// RetryOperationNode launch a task to retry operation instance
	RetryOperationNode(ctx context.Context, param RetryOperationNodeParam) ([]string, error)

	// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
	LaunchUpgradeNode(ctx context.Context, param UpgradeNodeParam) (string, error)

	// LaunchReconfigNode launch a task to reconfig node. returns the workflow-id.
	LaunchReconfigNode(ctx context.Context, param ReconfigNodeParam) (string, error)

	// LaunchRestartNode launch a task to restart node. returns the workflow-id.
	LaunchRestartNode(ctx context.Context, param RestartNodeParam) (string, error)

	// LaunchSyncAgentState launch a task to sync agent state from gse. returns the workflow-id.
	LaunchSyncAgentState(ctx context.Context, hostIDs ...int64) (string, error)

	// LaunchSyncAllAgentState launch a task to sync all agent state from gse. returns the workflow-id.
	LaunchSyncAllAgentState(ctx context.Context) (string, error)
}

// InstallNodeParam install node param.
type InstallNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// UpgradeNodeParam upgrade node param.
type UpgradeNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// ReconfigNodeParam reconfig node param.
type ReconfigNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// RestartNodeParam restart node param.
type RestartNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// RetryOperationNodeParam retry node param.
type RetryOperationNodeParam struct {
	WorkflowID string

	RetryMod     types.NodeOperationRetryMode
	OperationIDs []string
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
		workflow.WithStorageTrigger(conf.StorageTrigger),
		workflow.WithStorageOperation(conf.StorageOperation),
		workflow.WithStorageOperationInstance(conf.StorageOperInst),
		workflow.WithStorageActionInstance(conf.StorageOperInst),
		workflow.WithLocker(conf.LockerFactory),
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

	if err := mgr.startScheduleWorkflow(ctx); err != nil {
		return fmt.Errorf("failed to start schedule workflow, err: %w", err)
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

	if err := mgr.conf.StorageTopo.CheckHealthz(); err != nil {
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

// registerActionDefs init action defs.
func (mgr *manager) registerActionDefs() error {
	if err := mgr.registerOnceOperationActions(); err != nil {
		return fmt.Errorf("register once operation actions failed, err: %w", err)
	}

	if err := mgr.registerPeriodicOperationActions(); err != nil {
		return fmt.Errorf("register periodic operation actions failed, err: %w", err)
	}

	return nil
}

// registerOnceTriggerActions registers the action definitions for once trigger actions.
func (mgr *manager) registerOnceOperationActions() error {
	if err := mgr.registerActionDefNodeInstall(); err != nil {
		return fmt.Errorf("register action def node install failed, err: %w", err)
	}

	if err := mgr.registerActionDefSyncData(); err != nil {
		return fmt.Errorf("register action def sync data failed, err: %w", err)
	}

	return nil
}

// registerPeriodicTriggerActions registers the action definitions for periodic trigger actions.
func (mgr *manager) registerPeriodicOperationActions() error {
	if err := mgr.registerActionDefSchedule(); err != nil {
		return fmt.Errorf("register action def schedule failed, err: %w", err)
	}

	return nil
}

// registerActionDefSyncData registers the action definitions for sync data operations.
func (mgr *manager) registerActionDefSyncData() error {
	return mgr.workflowMgr.RegisterActions(
		syncdata.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.logger),
		syncdata.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionSyncNetworkAreaFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionGenOperSyncHost(mgr.conf.StorageTopo, mgr.workflowMgr),
		syncdata.NewActionSyncAgentState(mgr.conf.GSEHandler, mgr.conf.StorageTopo, mgr.logger),
		syncdata.NewActionGenOperSyncAgentState(mgr.conf.StorageTopo, mgr.workflowMgr),
	)
}

// registerActionDefNodeInstall registers the action definitions for node installation operations.
// nolint: lll
func (mgr *manager) registerActionDefNodeInstall() error {
	return mgr.workflowMgr.RegisterActions(
		nodeinstall.NewActionTryReuseAgentID(mgr.conf.StorageTopo, mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionBindAgentHostRel(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionInstallNodeBySSH(mgr.conf.InstallerFileGroup, mgr.logger, mgr.conf.StorageNodeDeployment, mgr.conf.Provider, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		nodeinstall.NewActionInstallNodeByWMI(mgr.conf.InstallerFileGroup, mgr.logger, mgr.conf.StorageNodeDeployment, mgr.conf.Provider, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		nodeinstall.NewActionWaitGseReady(mgr.conf.GSEHandler, mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionSyncNodeInfo(mgr.conf.GSEHandler, mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionPushHostIdentifier(mgr.conf.CmdbHandler, mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionRenderNodeDeployment(mgr.conf.StorageNodeDeployment, mgr.conf.StorageTopo, mgr.conf.StorageTopo, mgr.conf.StorageRelease, mgr.conf.StorageConfigPolicy, mgr.logger),
		nodeinstall.NewActionUpsertHostToCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.conf.StorageNodeDeployment),
		nodeinstall.NewActionWaitInstallerComplete(mgr.conf.StorageOperInst, mgr.logger),
		nodeinstall.NewActionUpdateHost(mgr.conf.StorageTopo, mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionTransferPkgToNode(mgr.conf.StorageNodeDeployment, mgr.conf.FileHandler, mgr.logger),
		nodeinstall.NewActionDetectInfoBySSH(mgr.logger, mgr.conf.StorageNodeDeployment, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		nodeinstall.NewActionDetectInfoByWMI(mgr.logger, mgr.conf.StorageNodeDeployment, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		nodeinstall.NewActionUpgradeNode(mgr.conf.StorageNodeDeployment, mgr.conf.GSEHandler, mgr.logger, mgr.conf.Provider),
		nodeinstall.NewActionCleanInstaller(mgr.conf.StorageNodeDeployment, mgr.conf.GSEHandler, mgr.logger),
		nodeinstall.NewActionVersionCompatCheck(mgr.conf.StorageNodeDeployment, mgr.logger),
		nodeinstall.NewActionReconfigNode(mgr.conf.StorageNodeDeployment, mgr.conf.GSEHandler, mgr.logger, mgr.conf.Provider),
		nodeinstall.NewActionRestartNode(mgr.conf.StorageNodeDeployment, mgr.conf.GSEHandler, mgr.logger),
		nodeinstall.NewActionEnsurePkgToRelay(mgr.conf.InstallerFileGroup, mgr.conf.StorageRelease, mgr.conf.StorageOperInst, mgr.conf.StorageNodeDeployment, mgr.conf.FileHandler, mgr.conf.ProxyMessager, mgr.logger),
	)
}

// registerActionDefSchedule registers the action definitions for schedule operations.
// nolint: lll
func (mgr *manager) registerActionDefSchedule() error {
	return mgr.workflowMgr.RegisterActions(
		schedule.NewActionGenScheduleOnceTrigger(SyncCmdbHostWorkflowName, mgr.conf.StorageOperInst, mgr.LaunchSyncBizAndHost),
		schedule.NewActionGenScheduleOnceTrigger(SyncGseAgentStateWorkflowName, mgr.conf.StorageOperInst, mgr.LaunchSyncAllAgentState),
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

	operationDef := syncdata.NewOperSyncBizAndHostFromCMDB(syncdata.OperParamSyncBizAndHostFromCMDB{TenantID: tenantID})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync biz and host task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

func (mgr *manager) LaunchSyncHostByBizID(ctx context.Context, bizID int64) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncHostFromCMDB(syncdata.OperParamSyncHostFromCMDB{TenantID: tenantID, BizID: bizID})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync host task. tenant-id(%s), biz-id(%d), trigger-id(%s), operation-id(%s)",
		tenantID, bizID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

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

	operationDef := syncdata.NewOperSyncNetworkAreaFromCMDB(syncdata.OperParamSyncNetworkAreaFromCMDB{TenantID: tenantID})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync networkarea task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchInstallNode launch a task to install node.
func (mgr *manager) LaunchInstallNode(ctx context.Context, param InstallNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNodeWorkflow.CreateNodeWorkflow(ctx, &types.NodeWorkflow{
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy

		gp.Go(func() error {
			return mgr.createOper(ctx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch install node task. err: %w", err)
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// RetryOperationNode launch a task to retry operation instance.
func (mgr *manager) RetryOperationNode(ctx context.Context, param RetryOperationNodeParam) ([]string, error) {
	instanceIDs := make([]string, 0)

	nodeWorkflow, err := mgr.conf.StorageNodeWorkflow.GetNodeWorkflow(ctx, param.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("get trigger failed, err: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.GetTrigger(ctx, nodeWorkflow.TriggerID)
	if err != nil {
		return nil, fmt.Errorf("get trigger failed, err: %w", err)
	}

	operCtls, err := triggerCtl.ListOperation(ctx, param.OperationIDs...)
	if err != nil {
		return nil, fmt.Errorf("get operation failed, err: %w", err)
	}

	err = mgr.conf.StorageNodeWorkflow.UpdateNodeWorkflowStatus(ctx, param.WorkflowID, types.NodeWorkflowStatusRunning)
	if err != nil {
		return nil, fmt.Errorf("update node workflow status failed, err: %w", err)
	}

	for _, operCtl := range operCtls {
		instanceCtl, err := operCtl.CreateRetryOperationInstance(ctx, param.RetryMod)
		if err != nil {
			return nil, fmt.Errorf("create operation instance failed, err: %w", err)
		}

		if err := instanceCtl.LaunchOperationInstance(ctx); err != nil {
			return nil, fmt.Errorf("launch operation instance failed, err: %w", err)
		}

		instanceIDs = append(instanceIDs, instanceCtl.GetOperationInstanceID())
	}

	return instanceIDs, nil
}
func (mgr *manager) createOper(
	ctx context.Context,
	operator string,
	triggerCtl workflow.ITriggerCtl,
	deploy *types.NodeDeployment,
) error {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if err := mgr.conf.StorageNodeDeployment.Create(ctx, deploy); err != nil {
		mgr.logger.ErrorCtxf(ctx,
			"failed to create node deployment. "+
				"tenant-id(%s), trigger-id(%s), node-deployment-token(%s), err(%v)",
			tenantID, triggerCtl.GetTriggerID(), deploy.Token, err)

		return err
	}

	// TODO: distinguish between pagent and agent based on workunitID
	var operationDef operation.Definition

	// [debug]
	if deploy.Info.Host.Static.InnerIP == "192.168.9.100" {
		relayhost, _, _ := mgr.conf.StorageTopo.ListHost(ctx, types.UnlimitedPage(), &types.HostCondition{
			ExactInclude: &types.HostExactFields{
				HostID: []int64{52287},
			},
		})
		relay := types.RelayInfo{
			HostID:     relayhost[0].HostID,
			AgentID:    relayhost[0].Dynamic.AgentID,
			NodeOsType: relayhost[0].Dynamic.NodeOsType,
		}
		deploy.Info.RelayInfo = relay

		mgr.logger.Infof("createOper: %s", deploy.Info.RelayInfo)

		deploy.Info.Host.Dynamic.NodeOsType = criteria.OSLinux
		deploy.Info.Host.Dynamic.NodeCPUArch = criteria.CPUArchAmd64
		deploy.Info.Host.Dynamic.NodeVersion = "v2.1.6-beta.59"
		deploy.Info.Host.Dynamic.NodeGeneration = 2

		err := mgr.conf.StorageNodeDeployment.UpdateInfo(ctx, deploy.Token, deploy.Info)
		if err != nil {
			mgr.logger.Errorf("failed to update node deployment: %v", err)
			return err
		}
		operationDef = nodeinstall.NewOperInstallPagentNodeBySSH(nodeinstall.OperParamInstallPagentNodeBySSH{
			Token:    deploy.Token,
			Operator: operator,
		})
		// end [debug]
	} else {
		switch deploy.Info.Host.Static.OSType {
		case string(criteria.OSLinux), string(criteria.OSDarwin):
			operationDef = nodeinstall.NewOperInstallNodeBySSH(nodeinstall.OperParamInstallNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			})
		case string(criteria.OSWindows):
			operationDef = nodeinstall.NewOperInstallNodeByWMI(nodeinstall.OperParamInstallNodeByWMI{
				Token:    deploy.Token,
				Operator: operator,
			})
		default:
			operationDef = nodeinstall.NewOperInstallNodeBySSH(nodeinstall.OperParamInstallNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			})
		}
	}

	// switch deploy.Info.Host.Static.OSType {
	// case string(criteria.OSLinux), string(criteria.OSDarwin):
	// 	operationDef = nodeinstall.NewOperInstallNodeBySSH(nodeinstall.OperParamInstallNodeBySSH{
	// 		Token:    deploy.Token,
	// 		Operator: operator,
	// 	})
	// case string(criteria.OSWindows):
	// 	operationDef = nodeinstall.NewOperInstallNodeByWMI(nodeinstall.OperParamInstallNodeByWMI{
	// 		Token:    deploy.Token,
	// 		Operator: operator,
	// 	})
	// default:
	// 	operationDef = nodeinstall.NewOperInstallNodeBySSH(nodeinstall.OperParamInstallNodeBySSH{
	// 		Token:    deploy.Token,
	// 		Operator: operator,
	// 	})
	// }

	operationParam := operationDef.DefaultParameters()
	operationParam.ExtraContent = deploymentInfoToMap(deploy.Info)

	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationParam)
	if err != nil {
		mgr.logger.ErrorCtxf(ctx,
			"failed to launch install node task. "+
				"tenant-id(%s), trigger-id(%s), operation-id(%s), node-deployment-token(%s), err(%v)",
			tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token, err)

		return err
	}

	mgr.logger.InfoCtxf(ctx,
		"launched install node task. tenant-id(%s), trigger-id(%s), operation-id(%s),"+
			" node-deployment-token(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token)

	return nil
}

// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
func (mgr *manager) LaunchUpgradeNode(ctx context.Context, param UpgradeNodeParam) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNodeWorkflow.CreateNodeWorkflow(ctx, &types.NodeWorkflow{
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launching upgrade node task. tenant-id(%s), trigger-id(%s), node-deployments(%d)",
		tenantID, triggerCtl.GetTriggerID(), len(param.NodeDeployments))

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy

		gp.Go(func() error {
			if err := mgr.conf.StorageNodeDeployment.Create(ctx, deploy); err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to create node deployment. "+
						"tenant-id(%s), trigger-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), deploy.Token, err)

				return err
			}

			operationDef := nodeinstall.NewOperUpgradeNode(nodeinstall.OperParamUpgradeNode{
				Token:    deploy.Token,
				Operator: param.Operator,
			})
			operationParam := operationDef.DefaultParameters()
			operationParam.ExtraContent = deploymentInfoToMap(deploy.Info)

			operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationParam)
			if err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to launch upgrade node task. "+
						"tenant-id(%s), trigger-id(%s), operation-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token, err)

				return err
			}

			mgr.logger.InfoCtxf(ctx,
				"launched upgrade node task. tenant-id(%s), trigger-id(%s), operation-id(%s),"+
					" node-deployment-token(%s)",
				tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token)

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch upgrade node task. err: %w", err)
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchReconfigNode launch a task to reconfig node. returns the workflow-id.
func (mgr *manager) LaunchReconfigNode(ctx context.Context, param ReconfigNodeParam) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNodeWorkflow.CreateNodeWorkflow(ctx, &types.NodeWorkflow{
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launching reconfig node task. tenant-id(%s), trigger-id(%s), node-deployments(%d)",
		tenantID, triggerCtl.GetTriggerID(), len(param.NodeDeployments))

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions.SelectDownloads = true
		deploy.Info.TransferOptions.EnableInstaller = true

		gp.Go(func() error {
			if err := mgr.conf.StorageNodeDeployment.Create(ctx, deploy); err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to create node deployment. "+
						"tenant-id(%s), trigger-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), deploy.Token, err)

				return err
			}

			operationDef := nodeinstall.NewOperReconfigNode(nodeinstall.OperParamReconfigNode{
				Token:    deploy.Token,
				Operator: param.Operator,
			})
			operationParam := operationDef.DefaultParameters()
			operationParam.ExtraContent = deploymentInfoToMap(deploy.Info)

			operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationParam)
			if err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to launch reconfig node task. "+
						"tenant-id(%s), trigger-id(%s), operation-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token, err)

				return err
			}

			mgr.logger.InfoCtxf(ctx,
				"launched reconfig node task. tenant-id(%s), trigger-id(%s), operation-id(%s),"+
					" node-deployment-token(%s)",
				tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token)

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch reconfig node task. err: %w", err)
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchRestartNode launch a task to restart node. returns the workflow-id.
func (mgr *manager) LaunchRestartNode(ctx context.Context, param RestartNodeParam) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNodeWorkflow.CreateNodeWorkflow(ctx, &types.NodeWorkflow{
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launching restart node task. tenant-id(%s), trigger-id(%s), node-deployments(%d)",
		tenantID, triggerCtl.GetTriggerID(), len(param.NodeDeployments))

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions.SelectDownloads = true
		deploy.Info.TransferOptions.EnableInstaller = true

		gp.Go(func() error {
			if err := mgr.conf.StorageNodeDeployment.Create(ctx, deploy); err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to create node deployment. "+
						"tenant-id(%s), trigger-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), deploy.Token, err)

				return err
			}

			operationDef := nodeinstall.NewOperRestartNode(nodeinstall.OperParamRestartNode{
				Token:    deploy.Token,
				Operator: param.Operator,
			})
			operationParam := operationDef.DefaultParameters()
			operationParam.ExtraContent = deploymentInfoToMap(deploy.Info)

			operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationParam)
			if err != nil {
				mgr.logger.ErrorCtxf(ctx,
					"failed to launch restart node task. "+
						"tenant-id(%s), trigger-id(%s), operation-id(%s), node-deployment-token(%s), err(%v)",
					tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token, err)

				return err
			}

			mgr.logger.InfoCtxf(ctx,
				"launched restart node task. tenant-id(%s), trigger-id(%s), operation-id(%s),"+
					" node-deployment-token(%s)",
				tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID(), deploy.Token)

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch restart node task. err: %w", err)
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func deploymentInfoToMap(info *types.DeploymentInfo) map[string]any {
	return map[string]any{
		"networkarea_id": info.Host.Static.NetworkAreaID,
		"biz_id":         info.Host.Static.BizID,
		"inner_ip":       info.Host.Static.InnerIP,
		"inner_ipv6":     info.Host.Static.InnerIPV6,
		"node_version":   info.Host.Dynamic.NodeVersion,
	}
}

// LaunchSyncAgentState launch a task to sync agent state.
func (mgr *manager) LaunchSyncAgentState(ctx context.Context, hostIDs ...int64) (string, error) {
	if len(hostIDs) == 0 {
		return "", errors.New("hostIDs cannot be empty")
	}

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	hosts, err := mgr.conf.StorageTopo.FindHostWithDynamic(ctx, types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			HostID: hostIDs,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to find hosts with dynamic info: %w", err)
	}

	hostAgentID := make([]*syncdata.HostIDAgentID, 0, len(hosts))
	for _, host := range hosts {
		hostAgentID = append(hostAgentID, &syncdata.HostIDAgentID{
			HostID:  host.HostID,
			AgentID: host.Dynamic.AgentID,
		})
	}

	operationDef := syncdata.NewOperSyncAgentStateFromGSE(syncdata.OperParamSyncAgentStateFromGSE{
		TenantID: tenantID,
		Hosts:    hostAgentID,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync agent state task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}

// LaunchSyncAllAgentState launch a task to sync all agent state.
func (mgr *manager) LaunchSyncAllAgentState(ctx context.Context) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	operationDef := syncdata.NewOperSyncAllAgentStateFromGSE(syncdata.OperParamSyncAllAgentStateFromGSE{
		TenantID: tenantID,
	})
	operCtl, err := triggerCtl.CreateOperation(ctx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		return "", err
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	mgr.logger.InfoCtxf(ctx, "launched sync all agent state task. tenant-id(%s), trigger-id(%s), operation-id(%s)",
		tenantID, triggerCtl.GetTriggerID(), operCtl.GetOperationID())

	return triggerCtl.GetTriggerID(), nil
}
