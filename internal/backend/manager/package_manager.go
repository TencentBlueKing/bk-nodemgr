/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// LaunchPackageImportPluginV3Pkg launches the package import plugin v3 pkg workflow.
func (mgr *Manager) LaunchPackageImportPluginV3Pkg(nCtx contextx.IContext, param types.PackageImportParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePackage.CreatePackageWorkflow(nCtx, &types.PackageWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PackageWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, pkgDeploy := range param.PackageDeployments {
		deploy := pkgDeploy

		gp.Go(func() error {
			return mgr.createPackageImportPluginPkgOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch import plugin v3 package task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) createPackageImportPluginPkgOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PackageDeployment) error {

	if err := mgr.conf.StoragePackage.CreatePackageDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("trigger-id", triggerCtl.GetTriggerID(), "package-token", deploy.Token).
			Error("failed to create package deployment.")

		return err
	}

	operationDef := mgr.getPackageImportPluginV3(nCtx, deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("package-token", deploy.Token).
			Error("failed to create package import plugin v3 package operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("package-token", deploy.Token).
		Info("launched import plugin v3 package task.")

	return nil
}

func (mgr *Manager) getPackageImportPluginV3(nCtx contextx.IContext, deploy *types.PackageDeployment, operator string) operation.Definition {
	return pkg.NewOperPackagePluginV3Import(pkg.OperParamPackagePluginV3Import{
		TenantID: nCtx.TenantID(),
		Token:    deploy.Token,
		Operator: operator,
	})
}

// LaunchPackageImportPluginV2Pkg launches the package import plugin v2 pkg workflow.
func (mgr *Manager) LaunchPackageImportPluginV2Pkg(nCtx contextx.IContext, param types.PackageImportParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePackage.CreatePackageWorkflow(nCtx, &types.PackageWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PackageWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, pkgDeploy := range param.PackageDeployments {
		deploy := pkgDeploy

		gp.Go(func() error {
			return mgr.createPackageImportPluginV2PkgOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch import plugin v2 package task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) createPackageImportPluginV2PkgOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PackageDeployment) error {

	if err := mgr.conf.StoragePackage.CreatePackageDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("trigger-id", triggerCtl.GetTriggerID(), "package-token", deploy.Token).
			Error("failed to create package deployment.")

		return err
	}

	operationDef := mgr.getPackageImportPluginV2(nCtx, deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("package-token", deploy.Token).
			Error("failed to create package import plugin v2 package operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("package-token", deploy.Token).
		Info("launched import plugin v2 package task.")

	return nil
}

func (mgr *Manager) getPackageImportPluginV2(nCtx contextx.IContext, deploy *types.PackageDeployment, operator string) operation.Definition {
	return pkg.NewOperPackagePluginV2Import(pkg.OperParamPackagePluginV2Import{
		TenantID: nCtx.TenantID(),
		Token:    deploy.Token,
		Operator: operator,
	})
}

// LaunchPackageImportExternalPluginV2Pkg launches the package import external plugin v2 pkg workflow.
func (mgr *Manager) LaunchPackageImportExternalPluginV2Pkg(nCtx contextx.IContext, param types.PackageImportParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePackage.CreatePackageWorkflow(nCtx, &types.PackageWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PackageWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, pkgDeploy := range param.PackageDeployments {
		deploy := pkgDeploy

		gp.Go(func() error {
			return mgr.createPackageImportExternalPluginV2PkgOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch import external plugin v2 package task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) createPackageImportExternalPluginV2PkgOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PackageDeployment) error {

	if err := mgr.conf.StoragePackage.CreatePackageDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("trigger-id", triggerCtl.GetTriggerID(), "package-token", deploy.Token).
			Error("failed to create package deployment.")

		return err
	}

	operationDef := mgr.getPackageImportExternalPluginV2(nCtx, deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("package-token", deploy.Token).
			Error("failed to create package import external plugin v2 package operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("package-token", deploy.Token).
		Info("launched import external plugin v2 package task.")

	return nil
}

func (mgr *Manager) getPackageImportExternalPluginV2(nCtx contextx.IContext, deploy *types.PackageDeployment, operator string) operation.Definition {
	return pkg.NewOperPackageExternalPluginV2Import(pkg.OperParamPackageExternalPluginV2Import{
		TenantID: nCtx.TenantID(),
		Token:    deploy.Token,
		Operator: operator,
	})
}

// LaunchPackageExportPlugin launches a plugin package export workflow.
func (mgr *Manager) LaunchPackageExportPlugin(nCtx contextx.IContext, param types.PackageExportParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePackage.CreatePackageWorkflow(nCtx, &types.PackageWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PackageWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, pkgDeploy := range param.PackageDeployments {
		deploy := pkgDeploy

		gp.Go(func() error {
			return mgr.createPackageExportPluginOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch export plugin task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) createPackageExportPluginOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PackageDeployment) error {

	if err := mgr.conf.StoragePackage.CreatePackageDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("trigger-id", triggerCtl.GetTriggerID(), "package-token", deploy.Token).
			Error("failed to create package deployment.")

		return err
	}

	operationDef := mgr.getPackageExportPlugin(nCtx, deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("package-token", deploy.Token).
			Error("failed to create package export plugin operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("package-token", deploy.Token).
		Info("launched export plugin task.")

	return nil
}

func (mgr *Manager) getPackageExportPlugin(nCtx contextx.IContext, deploy *types.PackageDeployment, operator string) operation.Definition {
	return pkg.NewOperPackageExportPlugin(pkg.OperParamPackageExportPlugin{
		TenantID: nCtx.TenantID(),
		Token:    deploy.Token,
		Operator: operator,
	})
}
