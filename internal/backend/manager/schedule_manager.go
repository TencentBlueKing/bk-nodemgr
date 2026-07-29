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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"github.com/google/uuid"
)

const (
	scheduledWorkflowSyncTenant                 = "sync_tenant"
	scheduledWorkflowSyncBizAndHost             = "sync_biz_and_host"
	scheduledWorkflowSyncNetworkArea            = "sync_networkarea"
	scheduledWorkflowSyncAgentState             = "sync_agent_state"
	scheduledWorkflowSyncAliveAgentInfo         = "sync_alive_agent_info"
	scheduledWorkflowSyncAlivePluginProcessInfo = "sync_alive_plugin_process_info"
	scheduledWorkflowWatchAndApplyCMDBResource  = "watch_and_apply_cmdb_resource"
	scheduledWorkflowExecuteDeployPolicy        = "execute_deploy_policy"
)

const (
	scheduledWorkflowMonitorTimeGap = 10 * time.Second
)

type initScheduledWorkflowFunc func(nCtx contextx.IContext, tenantID string) error
type syncScheduledWorkflowFunc func(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error

func (mgr *Manager) getInitScheduledWorkflowFuncs() map[string]initScheduledWorkflowFunc {
	return map[string]initScheduledWorkflowFunc{
		scheduledWorkflowSyncTenant:                 mgr.initSWSyncTenant,
		scheduledWorkflowSyncBizAndHost:             mgr.initSWSyncBizAndHost,
		scheduledWorkflowSyncNetworkArea:            mgr.initSWSyncNetworkArea,
		scheduledWorkflowSyncAgentState:             mgr.initSWSyncAgentState,
		scheduledWorkflowSyncAliveAgentInfo:         mgr.initSWSyncAliveAgentInfo,
		scheduledWorkflowSyncAlivePluginProcessInfo: mgr.initSWSyncAlivePluginProcessInfo,
		scheduledWorkflowWatchAndApplyCMDBResource:  mgr.initSWWatchAndApplyCMDBResource,
		scheduledWorkflowExecuteDeployPolicy:        mgr.initSWExecuteDeployPolicy,
	}
}

func (mgr *Manager) getSyncScheduledWorkflowFuncs() map[string]syncScheduledWorkflowFunc {
	return map[string]syncScheduledWorkflowFunc{
		scheduledWorkflowSyncTenant:                 mgr.syncSWSyncTenant,
		scheduledWorkflowSyncBizAndHost:             mgr.syncSWSyncBizAndHost,
		scheduledWorkflowSyncNetworkArea:            mgr.syncSWSyncNetworkArea,
		scheduledWorkflowSyncAgentState:             mgr.syncSWSyncAgentState,
		scheduledWorkflowSyncAliveAgentInfo:         mgr.syncSWSyncAliveAgentInfo,
		scheduledWorkflowSyncAlivePluginProcessInfo: mgr.syncSWSyncAlivePluginProcessInfo,
		scheduledWorkflowWatchAndApplyCMDBResource:  mgr.syncSWWatchAndApplyCMDBResource,
		scheduledWorkflowExecuteDeployPolicy:        mgr.syncSWExecuteDeployPolicy,
	}
}

// nolint: gocognit
func (mgr *Manager) startMonitoringScheduledWorkflow(nCtx contextx.IContext) {
	logger.G.Sys().With("time-gap", scheduledWorkflowMonitorTimeGap.String()).Info("start monitoring scheduled workflows")

	tenantIDs := tenant.GetAllTenantIDs()
	for _, tenantID := range tenantIDs {
		for name, f := range mgr.getInitScheduledWorkflowFuncs() {
			if err := mgr.initScheduleWorkflow(nCtx, tenantID, name, f); err != nil {
				logger.G.Sys().WithErr(err).With("tenant-id", tenantID, "workflow-name", name).Error("failed to initialize scheduled workflow")

				continue
			}

			logger.G.Sys().With("tenant-id", tenantID, "workflow-name", name).Info("initialized scheduled workflow")
		}
	}

	go func() {
		ticker := time.NewTicker(scheduledWorkflowMonitorTimeGap)
		for {
			select {
			case <-nCtx.Done():
				logger.G.Sys().Warn("stopping monitoring scheduled workflows")

				return

			case <-ticker.C:
				sws, _, err := mgr.conf.StorageWorkflow.ListScheduledWorkflow(nCtx, types.UnlimitedPage())
				if err != nil {
					logger.G.Sys().WithErr(err).Error("failed to list scheduled workflows")

					continue
				}

				for _, sw := range sws {
					if err = mgr.ensureScheduledWorkflow(nCtx, sw); err != nil {
						logger.G.Sys().
							WithErr(err).
							With("workflow-id", sw.WorkflowID, "trigger-id", sw.TriggerID).
							Error("failed to ensure scheduled workflow")

						continue
					}
				}
			}
		}
	}()
}

func (mgr *Manager) initScheduleWorkflow(nCtx contextx.IContext, tenantID string, workflowName string, initFunc initScheduledWorkflowFunc) error {
	nCtx = contextx.From(nCtx, contextx.WithTenantID(tenantID))

	locker := mgr.genScheduledWorkflowLocker(tenantID, workflowName)
	if err := locker.tryLock(nCtx); err != nil {
		return nil
	}
	defer func() {
		_ = locker.unlock(nCtx)
	}()

	// reload the scheduled workflow after get the lock.
	cond := &types.ScheduledWorkflowCondition{
		ExactInclude: &types.ScheduledWorkflowExactFields{
			WorkflowName: []string{workflowName},
		},
	}

	sws, _, err := mgr.conf.StorageWorkflow.ListScheduledWorkflow(nCtx, types.SingleItemPage(), cond)
	if err != nil {
		return err
	}

	switch len(sws) {
	case 0:
		if err = initFunc(nCtx, tenantID); err != nil {
			return fmt.Errorf("failed to initialize scheduled workflow: %w", err)
		}

		return nil
	case 1:
		break
	default:
		return fmt.Errorf("expected 0 or 1 scheduled workflow, but got %d", len(sws))
	}

	return nil
}

func (mgr *Manager) ensureScheduledWorkflow(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	locker := mgr.genScheduledWorkflowLocker(sw.TenantID, sw.WorkflowName)
	if err := locker.tryLock(nCtx); err != nil {
		return nil
	}
	defer func() {
		_ = locker.unlock(nCtx)
	}()

	nCtx = contextx.From(nCtx, contextx.WithTenantID(sw.TenantID))

	// reload the scheduled workflow after get the lock.
	var err error
	sw, err = mgr.conf.StorageWorkflow.GetScheduledWorkflow(nCtx, sw.WorkflowID)
	if err != nil {
		return err
	}

	// scheduled workflow has not been triggered yet.
	if sw.TriggerID == "" {
		return mgr.trySyncingScheduledWorkflow(nCtx, sw)
	}

	// check if the trigger has been created.
	ok, err := mgr.conf.StorageWorkflow.ExistTrigger(nCtx, sw.TriggerID)
	if err != nil {
		return err
	}
	if !ok {
		return mgr.trySyncingScheduledWorkflow(nCtx, sw)
	}

	if err := mgr.ensureTrigger(nCtx, sw); err != nil {
		return fmt.Errorf("failed to ensure scheduled workflow's trigger: %w", err)
	}

	return nil
}

func (mgr *Manager) ensureTrigger(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	trigCtl, err := mgr.workflowMgr.GetTrigger(nCtx, sw.TriggerID)
	if err != nil {
		return err
	}

	// check if the trigger state is changed.
	if !sw.Enabled {
		if err := trigCtl.InactivateTrigger(nCtx); err != nil {
			logger.G.Sys().
				With("workflow-name", sw.WorkflowName,
					"scheduled-workflow-id", sw.WorkflowID,
					"trigger-id", sw.TriggerID,
				).
				WithErr(err).Error("failed to inactivate scheduled workflow's trigger")

			return fmt.Errorf("failed to inactivate scheduled workflow's trigger")
		}
	} else {
		if err := trigCtl.ActivateTrigger(nCtx); err != nil {
			logger.G.Sys().
				With("workflow-name", sw.WorkflowName,
					"scheduled-workflow-id", sw.WorkflowID,
					"trigger-id", sw.TriggerID,
				).
				WithErr(err).Error("failed to activate scheduled workflow's trigger")

			return fmt.Errorf("failed to activate scheduled workflow's trigger")
		}
	}

	meta, ok := trigCtl.GetTriggerMetadata().(*trigger.MetadataPeriodic)
	if !ok {
		// notice: This is a case that should not have happened.
		logger.G.Sys().With("workflow-name", sw.WorkflowName).
			With("scheduled-workflow-id", sw.WorkflowID,
				"original-trigger-metadata", meta,
			).
			Warn("happened unexpected trigger metadata, scheduled workflow's trigger was rebuild")

		return mgr.trySyncingScheduledWorkflow(nCtx, sw)
	}

	var needUpdate bool
	if meta.Interval != sw.Interval {
		meta.Interval = sw.Interval
		needUpdate = true
	}

	if needUpdate {
		if err := trigCtl.UpdateMetadata(nCtx, meta); err != nil {
			logger.G.Sys().
				With("workflow-name", sw.WorkflowName,
					"scheduled-workflow-id", sw.WorkflowID,
					"original-trigger-metadata", meta,
				).
				WithErr(err).Error("failed to update scheduled workflow's trigger metadata")

			return fmt.Errorf("failed to update scheduled workflow's trigger metadata: %w", err)
		}
	}

	return nil
}

func (mgr *Manager) trySyncingScheduledWorkflow(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	funcs := mgr.getSyncScheduledWorkflowFuncs()

	f, ok := funcs[sw.WorkflowName]
	if !ok {
		logger.G.Sys().With("workflow-name", sw.WorkflowName).Warn("workflow function not found, skip")

		return nil
	}

	return f(nCtx, sw)
}

func (mgr *Manager) initScheduledWorkflow(nCtx contextx.IContext, tenantID, workflowName, interval string) error {
	sw := &types.ScheduledWorkflow{
		WorkflowID:   identifier.GenWorkflowID(),
		TenantID:     tenantID,
		Enabled:      true,
		WorkflowName: workflowName,
		Interval:     interval,
		Operator:     access.GetVirtualUser(),
		OperateTime:  time.Now(),
	}

	if err := mgr.conf.StorageWorkflow.CreateScheduledWorkflow(nCtx, sw); err != nil {
		logger.G.Sys().WithErr(err).With("workflow-name", sw.WorkflowName, "tenant-id", sw.TenantID).Error("failed to create scheduled workflow")

		return fmt.Errorf("failed to create scheduled workflow: %w", err)
	}

	logger.G.Sys().With("workflow-name", sw.WorkflowName, "tenant-id", sw.TenantID).Info("created scheduled workflow")

	return nil
}

func (mgr *Manager) syncScheduledWorkflow(nCtx contextx.IContext, sw *types.ScheduledWorkflow, operationDef operation.Definition) error {
	metadata, err := trigger.NewMetadataPeriodic(sw.Interval, false)
	if err != nil {
		logger.G.Sys().WithErr(err).With("workflow-name", sw.WorkflowName).Error("failed to create periodic metadata")

		return fmt.Errorf("failed to create periodic metadata: %w", err)
	}
	metadata.CleanPolicy.MaxOperInstNum = 100

	trigCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryPeriodic, metadata)
	if err != nil {
		logger.G.Sys().WithErr(err).With("workflow-name", sw.WorkflowName).Error("failed to create periodic trigger")

		return fmt.Errorf("failed to create periodic trigger: %w", err)
	}

	sw.TriggerID = trigCtl.GetTriggerID()
	if err = mgr.conf.StorageWorkflow.UpdateScheduledWorkflowTriggerID(nCtx, sw.WorkflowID, sw.TriggerID); err != nil {
		logger.G.Sys().
			WithErr(err).
			With("workflow-name", sw.WorkflowName, "trigger-id", sw.TriggerID).
			Error("failed to update scheduled workflow trigger id")

		return fmt.Errorf("failed to update scheduled workflow trigger id: %w", err)
	}

	operCtl, err := trigCtl.CreateOperation(nCtx, operationDef, operationDef.DefaultParameters())
	if err != nil {
		logger.G.Sys().
			WithErr(err).
			With("workflow-name", sw.WorkflowName, "trigger-id", trigCtl.GetTriggerID()).
			Error("failed to create operation for scheduled workflow")

		return fmt.Errorf("failed to create operation for scheduled workflow: %w", err)
	}

	logger.G.Sys().
		With("workflow-name", sw.WorkflowName, "trigger-id", sw.TriggerID, "operation-id", operCtl.GetOperationID()).
		Info("scheduled workflow with new trigger")

	if sw.Enabled {
		return trigCtl.ActivateTrigger(nCtx)
	}

	return nil
}

func (mgr *Manager) initSWSyncTenant(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowSyncTenant, scheduler.Every10m)
}

func (mgr *Manager) syncSWSyncTenant(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(nCtx, sw, schedule.NewOperSyncTenant(schedule.OperParamSyncTenant{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) initSWSyncBizAndHost(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowSyncBizAndHost, scheduler.Every30m)
}

func (mgr *Manager) syncSWSyncBizAndHost(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(nCtx, sw, schedule.NewOperSyncBizAndHost(schedule.OperParamSyncBizAndHost{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) initSWSyncNetworkArea(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowSyncNetworkArea, scheduler.Every10m)
}

func (mgr *Manager) syncSWSyncNetworkArea(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(nCtx, sw, schedule.NewOperSyncNetworkArea(schedule.OperParamSyncNetworkArea{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) initSWSyncAgentState(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowSyncAgentState, scheduler.Every10m)
}

func (mgr *Manager) syncSWSyncAgentState(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(nCtx, sw, schedule.NewOperSyncAgentState(schedule.OperParamSyncAgentState{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
		CompareCurrentState: true,
	}))
}

func (mgr *Manager) initSWSyncAliveAgentInfo(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowSyncAliveAgentInfo, scheduler.Every30m)
}

func (mgr *Manager) syncSWSyncAliveAgentInfo(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(nCtx, sw, schedule.NewOperSyncAliveAgentInfo(schedule.OperParamSyncAliveAgentInfo{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) initSWSyncAlivePluginProcessInfo(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowSyncAlivePluginProcessInfo, scheduler.Every10m)
}

func (mgr *Manager) syncSWSyncAlivePluginProcessInfo(nCtx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(nCtx, sw, schedule.NewOperSyncAlivePluginProcessInfo(schedule.OperParamSyncAlivePluginProcessInfo{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) initSWWatchAndApplyCMDBResource(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowWatchAndApplyCMDBResource, scheduler.Every10s)
}

func (mgr *Manager) syncSWWatchAndApplyCMDBResource(ctx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(ctx, sw, schedule.NewOperWatchAndApplyCMDBResource(schedule.OperParamWatchAndApplyCMDBResource{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) initSWExecuteDeployPolicy(nCtx contextx.IContext, tenantID string) error {
	return mgr.initScheduledWorkflow(nCtx, tenantID, scheduledWorkflowExecuteDeployPolicy, scheduler.Every1h)
}

func (mgr *Manager) syncSWExecuteDeployPolicy(ctx contextx.IContext, sw *types.ScheduledWorkflow) error {
	return mgr.syncScheduledWorkflow(ctx, sw, schedule.NewOperExecuteDeployPolicy(schedule.OperParamExecuteDeployPolicy{
		ScheduleActionStandardParam: utils.ScheduleActionStandardParam{
			WorkflowID: sw.WorkflowID,
			TenantID:   sw.TenantID,
			Operator:   access.GetVirtualUser(),
		},
	}))
}

func (mgr *Manager) genScheduledWorkflowLocker(tenantID, workflowName string) *scheduledWorkflowLocker {
	return &scheduledWorkflowLocker{
		tenantID:     tenantID,
		workflowName: workflowName,

		key:        fmt.Sprintf("bknm:backend:scheduledworkflow:%s:%s", tenantID, workflowName),
		uuid:       uuid.New().String(),
		expiration: 1 * time.Minute,

		cache: mgr.conf.Cache,
	}
}

type scheduledWorkflowLocker struct {
	tenantID     string
	workflowName string

	key        string
	uuid       string
	expiration time.Duration

	cache cache.ICache
}

func (locker *scheduledWorkflowLocker) tryLock(nCtx contextx.IContext) error {
	result, err := locker.cache.SetNXWithExpiration(nCtx, locker.key, []byte(locker.uuid), locker.expiration)
	if err != nil {
		return err
	}

	if !result {
		return errors.New("failed to lock scheduled workflow, maybe already locked")
	}

	return nil
}

func (locker *scheduledWorkflowLocker) unlock(nCtx contextx.IContext) error {
	result, err := locker.cache.Get(nCtx, locker.key)
	if err != nil {
		return err
	}

	if string(result) != locker.uuid {
		return errors.New("failed to unlock scheduled workflow, maybe already unlocked")
	}

	if _, err := locker.cache.Delete(nCtx, locker.key); err != nil {
		return err
	}

	return nil
}
