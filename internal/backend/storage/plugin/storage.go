/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin provide plugin storage.
// nolint: nonamedreturns
package plugin

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoNodeDeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-deployment"
	daoOperation "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin"
	plugindeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-deployment"
	pluginworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-workflow"
	daoProcess "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process"
	daoProcessConfig "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process-config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName the name of storage.
	StorageName = "plugin"

	schedulerTaskObtainMonitoredWorkflows = "obtain_monitored_plugin_workflows"
	schedulerTaskMonitorWorkflowStatus    = "monitor_plugin_workflow_status"
	recentMonitoredTime                   = 5 * time.Minute
	missingOperationGraceTime             = 1 * time.Minute

	metricOperationGetPluginDeploymentInfo                           = "get_plugin_deployment_info"
	metricOperationCreatePluginDeployment                            = "create_plugin_deployment"
	metricOperationListPluginDeployment                              = "list_plugin_deployment"
	metricOperationUpdatePluginDeploymentInfo                        = "update_plugin_deployment_info"
	metricOperationGetPluginDeploymentPluginConf                     = "get_plugin_deployment_plugin_conf"
	metricOperationUpdatePluginDeploymentPluginConf                  = "update_plugin_deployment_plugin_conf"
	metricOperationGetPluginDeploymentPluginConfConfigFilesDetail    = "get_plugin_deployment_plugin_conf_config_files_detail"
	metricOperationUpsertPluginDeploymentPluginConfConfigFilesDetail = "upsert_plugin_deployment_plugin_conf_config_files_detail"

	metricOperationGetPluginWorkflow          = "get_plugin_workflow"
	metricOperationGetPluginWorkflowStatus    = "get_plugin_workflow_status"
	metricOperationCreatePluginWorkflow       = "create_plugin_workflow"
	metricOperationUpdatePluginWorkflowStatus = "update_plugin_workflow_status"
	metricOperationCountPluginWorkflow        = "count_plugin_workflow"
	metricOperationListPluginWorkflow         = "list_plugin_workflow"
	metricOperationDistinctPluginWorkflow     = "distinct_plugin_workflow"

	metricOperationGetPlugin                         = "get_plugin"
	metricOperationCountPlugins                      = "count_plugins"
	metricOperationListPlugins                       = "list_plugins"
	metricOperationExistPluginByPluginName           = "exist_plugin_by_plugin_name"
	metricOperationExistDefaultPluginByPluginPkgName = "exist_default_plugin_by_plugin_pkg_name"
	metricOperationSetPluginMemo                     = "set_plugin_memo"
	metricOperationCreatePlugin                      = "create_plugin"
	metricOperationUpsertManyPlugins                 = "upsert_many_plugins"
	metricOperationGetPluginVisibleBizIDs            = "get_plugin_visible_biz_ids"
	metricOperationListVisiblePluginByBizIDs         = "list_visible_plugin_by_biz_ids"

	metricOperationCountProcess                       = "count_process"
	metricOperationListProcess                        = "list_process"
	metricOperationCreateProcess                      = "create_process"
	metricOperationUpdateProcess                      = "update_process"
	metricOperationUpdateProcessInfo                  = "update_process_info"
	metricOperationUpdateManyProcessInfo              = "update_many_process_info"
	metricOperationUpdateProcessBizID                 = "update_process_biz_id"
	metricOperationDeleteProcess                      = "delete_process"
	metricOperationExistProcess                       = "exist_process"
	metricOperationGetProcess                         = "get_process"
	metricOperationGetProcessDistributionByHostID     = "get_process_distribution_by_host_id"
	metricOperationGetProcessDistributionByPluginName = "get_process_distribution_by_plugin_name"
	metricOperationDistinctProcess                    = "distinct_process"

	metricOperationGetProcessConfig                       = "get_process_config"
	metricOperationCreateProcessConfig                    = "create_process_config"
	metricOperationUpsertManyProcessConfigs               = "upsert_many_process_configs"
	metricOperationDeleteProcessConfigsByProcessUniqueKey = "delete_process_configs_by_process_unique_key"
	metricOperationDeleteProcessConfigs                   = "delete_process_configs"
	metricOperationListProcessConfig                      = "list_process_config"
	metricOperationCountProcessConfig                     = "count_process_config"
)

// NewStorage ...
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
		monitoredWorkflows:      make(map[string]*types.PluginWorkflow),
		monitoredWorkflowsMutex: sync.RWMutex{},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	err = s.registerScheduler()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register scheduler")

		return nil, fmt.Errorf("register scheduler failed: %w", err)
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage this is a storage to operate plugin deployment table.
type Storage struct {
	basestorage.Storage

	// dao
	daoPluginDeployment plugindeployment.IHandler
	daoPluginWorkflow   pluginworkflow.IHandler
	daoPlugin           plugin.IHandler
	daoProcess          daoProcess.IHandler
	daoProcessConfig    daoProcessConfig.IHandler
	daoOperation        daoOperation.IHandler
	daoNodeDeployment   daoNodeDeployment.IHandler

	monitoredWorkflows      map[string]*types.PluginWorkflow
	monitoredWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoPluginDeployment = plugindeployment.New(s.Database)
	s.daoPluginWorkflow = pluginworkflow.New(s.Database)
	s.daoPlugin = plugin.New(s.Database)
	s.daoProcess = daoProcess.New(s.Database)
	s.daoProcessConfig = daoProcessConfig.New(s.Database)
	s.daoOperation = daoOperation.New(s.Database)
	s.daoNodeDeployment = daoNodeDeployment.New(s.Database)

	return nil
}

func (s *Storage) registerScheduler() error {
	s.Scheduler = scheduler.NewScheduler()
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		schedulerTaskObtainMonitoredWorkflows,
		5*time.Second,  // nolint: mnd
		20*time.Second, // nolint: mnd
		s.obtainMonitoredWorkflows,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register obtain monitored plugin workflows task")

		return fmt.Errorf("register obtain monitored plugin workflows task failed: %w", err)
	}

	err = s.Scheduler.RegisterTask(scheduler.NewTask(
		schedulerTaskMonitorWorkflowStatus,
		1*time.Second,  // nolint: mnd
		10*time.Second, // nolint: mnd
		s.monitorWorkflowStatus,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register monitor plugin workflow status task")

		return fmt.Errorf("register monitor plugin workflow status task failed: %w", err)
	}

	return nil
}

// obtainMonitoredWorkflows Obtain a list of workflows that need to be listened to.
func (s *Storage) obtainMonitoredWorkflows(nCtx contextx.IContext) error {
	tenantIDs := tenant.GetAllTenantIDs()

	type tenantResult struct {
		running        []*types.PluginWorkflow
		recentFinished []*types.PluginWorkflow
	}
	results := make([]tenantResult, len(tenantIDs))

	gp := gopool.NewPool()
	for i, tenantID := range tenantIDs {
		slot := &results[i]
		tenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID))
		fn := func() error {
			runningWorkflows, _, err := s.daoPluginWorkflow.List(
				tenantCtx,
				types.UnlimitedPage(),
				pluginworkflow.WithStatus(types.PluginWorkflowStatusRunning))
			if err != nil {
				return fmt.Errorf("query running workflows failed: %w", err)
			}

			recentFinishedWorkflows, _, err := s.daoPluginWorkflow.List(
				tenantCtx,
				types.UnlimitedPage(),
				pluginworkflow.WithStatus(types.GetFinishedPluginWorkflowStatus()...),
				pluginworkflow.WithOperateTimeRange(types.RecentTimeRange(recentMonitoredTime)),
			)
			if err != nil {
				return fmt.Errorf("query recent finished plugin workflows failed: %w", err)
			}

			slot.running = runningWorkflows
			slot.recentFinished = recentFinishedWorkflows

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to obtain monitored workflows")
	}

	s.monitoredWorkflowsMutex.Lock()

	s.monitoredWorkflows = make(map[string]*types.PluginWorkflow)
	for i := range results {
		for _, workflow := range results[i].running {
			s.monitoredWorkflows[workflow.TriggerID] = workflow
		}
	}

	// recentFinished has higher priority and overwrites running entries
	for i := range results {
		for _, workflow := range results[i].recentFinished {
			s.monitoredWorkflows[workflow.TriggerID] = workflow
		}
	}

	s.monitoredWorkflowsMutex.Unlock()

	return nil
}

// nolint: gocognit
func (s *Storage) monitorWorkflowStatus(nCtx contextx.IContext) error {
	// Phase 1: RLock → deep copy snapshot → RUnlock
	s.monitoredWorkflowsMutex.RLock()
	if len(s.monitoredWorkflows) == 0 {
		s.monitoredWorkflowsMutex.RUnlock()
		return nil
	}
	snapshot := make(map[string]*types.PluginWorkflow, len(s.monitoredWorkflows))
	for k, v := range s.monitoredWorkflows {
		snapshot[k] = v
	}
	s.monitoredWorkflowsMutex.RUnlock()

	// Phase 2: work with snapshot, no lock held
	operations, _, err := s.daoOperation.List(nCtx,
		types.UnlimitedPage(),
		daoOperation.WithTriggerID(conv.MapKeyToSlice(snapshot)...))
	if err != nil {
		return fmt.Errorf("query operations failed: %w", err)
	}

	unfinishedTriggerMap := make(map[string]struct{})
	triggerOpers := make(map[string][]*operation.Operation)
	for _, oper := range operations {
		triggerOpers[oper.TriggerID] = append(triggerOpers[oper.TriggerID], oper)
		if oper.LatestInstBriefData == nil || !operation.CheckStateFinished(oper.LatestInstBriefData.Lifecycle.State) {
			unfinishedTriggerMap[oper.TriggerID] = struct{}{}
		}
	}

	triggersToDelete := make([]string, 0)
	for triggerID, pluginWorkflow := range snapshot {
		opers, hasOperations := triggerOpers[triggerID]
		if !hasOperations || len(opers) == 0 {
			// If the workflow is already in a finished state, its operations may have been cleaned
			// up by a scheduled purge job. Do not overwrite a legitimate terminal status.
			if pluginWorkflow.Status != types.PluginWorkflowStatusRunning {
				triggersToDelete = append(triggersToDelete, triggerID)
				continue
			}

			if !pluginWorkflow.OperateTime.IsZero() && time.Since(pluginWorkflow.OperateTime) < missingOperationGraceTime {
				continue
			}

			if err := s.updatePluginWorkflowResult(nCtx, pluginWorkflow, types.PluginWorkflowStatusFailed, time.Now()); err != nil {
				logger.G.Sys().WithErr(err).
					With("trigger-id", triggerID, "workflow-id", pluginWorkflow.WorkflowID).
					Error("failed to fallback update plugin workflow status when operation is missing")

				continue
			}

			logger.G.Sys().With("trigger-id", triggerID, "workflow-id", pluginWorkflow.WorkflowID).
				Warn("plugin workflow has no operation after grace time, fallback status to failed")

			triggersToDelete = append(triggersToDelete, triggerID)

			continue
		}

		if _, unfinished := unfinishedTriggerMap[triggerID]; unfinished {
			continue
		}
		status, finishTime, zeroEndTime := calWorkflowStatusAndTime(opers)
		if zeroEndTime {
			logger.G.Sys().With("trigger-id", triggerID).
				Warn("all operation instances have zero end time, using current time as fallback")
		}

		if err := s.updatePluginWorkflowResult(nCtx, pluginWorkflow, status, finishTime); err != nil {
			logger.G.Sys().WithErr(err).
				With("trigger-id", triggerID, "workflow-id", pluginWorkflow.WorkflowID).
				Error("failed to update plugin workflow status and finish time")

			continue
		}

		triggersToDelete = append(triggersToDelete, triggerID)
	}

	// Phase 3: Lock → batch delete → Unlock (only if needed)
	if len(triggersToDelete) > 0 {
		s.monitoredWorkflowsMutex.Lock()
		for _, triggerID := range triggersToDelete {
			delete(s.monitoredWorkflows, triggerID)
		}
		s.monitoredWorkflowsMutex.Unlock()
	}

	return nil
}

func (s *Storage) updatePluginWorkflowResult(
	nCtx contextx.IContext, pluginWorkflow *types.PluginWorkflow, status types.PluginWorkflowStatus, finishTime time.Time) error {

	updateCtx, cancel := contextx.WithTimeout(
		contextx.From(nCtx, contextx.WithTenantID(pluginWorkflow.TenantID)),
		30*time.Second, // nolint: mnd
	)
	defer cancel()

	err := s.daoPluginWorkflow.UpdateStatus(updateCtx, pluginWorkflow.WorkflowID, status)
	if err != nil {
		return fmt.Errorf("update plugin workflow status failed: %w", err)
	}

	err = s.daoPluginWorkflow.UpdateFinishTime(updateCtx, pluginWorkflow.WorkflowID, finishTime)
	if err != nil {
		return fmt.Errorf("update plugin workflow finish time failed: %w", err)
	}

	return nil
}

func calWorkflowStatusAndTime(opers []*operation.Operation) (types.PluginWorkflowStatus, time.Time, bool) {
	successCount := 0
	failedCount := 0
	var latestEndTime time.Time

	for _, oper := range opers {
		if oper.LatestInstBriefData == nil || oper.LatestInstBriefData.Lifecycle == nil {
			continue
		}
		lc := oper.LatestInstBriefData.Lifecycle
		if lc.EndedAt.After(latestEndTime) {
			latestEndTime = lc.EndedAt
		}

		switch lc.State {
		case operation.StateSuccess:
			successCount++
		case operation.StateFailed, operation.StateTimeout:
			failedCount++
		default:
		}
	}

	zeroEndTime := latestEndTime.IsZero()
	if zeroEndTime {
		latestEndTime = time.Now()
	}

	total := len(opers)
	switch {
	case successCount == total:
		return types.PluginWorkflowStatusSuccess, latestEndTime, zeroEndTime
	case failedCount == total:
		return types.PluginWorkflowStatusFailed, latestEndTime, zeroEndTime
	default:
		return types.PluginWorkflowStatusPartialFailed, latestEndTime, zeroEndTime
	}
}

func (s *Storage) check() error {
	if s.daoPluginDeployment == nil {
		return errors.New("dao plugin deployment is nil")
	}

	if s.daoPluginWorkflow == nil {
		return errors.New("dao plugin workflow is nil")
	}

	return nil
}

// ===============================================================================
// PluginDeploymentInfo Related Interface
// ===============================================================================

// GetPluginDeploymentInfo get plugin deployment info.
func (s *Storage) GetPluginDeploymentInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	var (
		info *types.PluginDeploymentInfo
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginDeploymentInfo, func(nCtx contextx.IContext) error {
		var err error
		info, err = s.getPluginDeploymentInfo(nCtx, token)

		return err
	})

	return info, err
}

// CreatePluginDeployment plugin deployment.
func (s *Storage) CreatePluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginDeployment, func(nCtx contextx.IContext) error {
		var err error
		err = s.createPluginDeployment(nCtx, pluginDeployment)

		return err
	})

	return err
}

// ListPluginDeployment list plugin deployment.
func (s *Storage) ListPluginDeployment(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginDeploymentCondition) (
	[]*types.PluginDeployment, int64, error) {

	var (
		pluginDeployments []*types.PluginDeployment
		total             int64
		err               error
	)

	err = s.WrapFn(nCtx, metricOperationListPluginDeployment, func(nCtx contextx.IContext) error {
		var err error
		pluginDeployments, total, err = s.listPluginDeployment(nCtx, page, conditions...)

		return err
	})

	return pluginDeployments, total, err
}

// UpdatePluginDeploymentInfo update a plugin deployment info.
func (s *Storage) UpdatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdatePluginDeploymentInfo, func(nCtx contextx.IContext) error {
		var err error
		err = s.updatePluginDeploymentInfo(nCtx, token, pluginDeploymentInfo)

		return err
	})

	return err
}

// GetPluginDeploymentPluginConf get plugin deployment plugin conf.
func (s *Storage) GetPluginDeploymentPluginConf(ctx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error) {
	var (
		pluginConf *types.PluginDeploymentPluginConf
		err        error
	)

	err = s.WrapFn(ctx, metricOperationGetPluginDeploymentPluginConf, func(ctx contextx.IContext) error {
		var err error
		pluginConf, err = s.getPluginDeploymentPluginConf(ctx, token)

		return err
	})

	return pluginConf, err
}

// UpdatePluginDeploymentPluginConf set plugin deployment plugin conf.
func (s *Storage) UpdatePluginDeploymentPluginConf(ctx contextx.IContext, token string, pluginConf *types.PluginDeploymentPluginConf) error {
	var (
		err error
	)

	err = s.WrapFn(ctx, metricOperationUpdatePluginDeploymentPluginConf, func(ctx contextx.IContext) error {
		var err error
		err = s.updatePluginDeploymentPluginConf(ctx, token, pluginConf)

		return err
	})

	return err
}

// GetPluginDeploymentPluginConfConfigFilesDetail get plugin deployment plugin conf config files detail.
func (s *Storage) GetPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string) ([]*types.PluginConfigDetail, error) {
	var (
		config []*types.PluginConfigDetail
		err    error
	)

	err = s.WrapFn(ctx, metricOperationGetPluginDeploymentPluginConfConfigFilesDetail, func(ctx contextx.IContext) error {
		var err error
		config, err = s.getPluginDeploymentPluginConfConfigFilesDetail(ctx, token)

		return err
	})

	return config, err
}

// UpsertPluginDeploymentPluginConfConfigFilesDetail update plugin deployment plugin conf config files detail.
func (s *Storage) UpsertPluginDeploymentPluginConfConfigFilesDetail(
	ctx contextx.IContext, token string, configDetails ...*types.PluginConfigDetail) error {

	var (
		err error
	)

	err = s.WrapFn(ctx, metricOperationUpsertPluginDeploymentPluginConfConfigFilesDetail, func(ctx contextx.IContext) error {
		var err error
		err = s.upsertPluginDeploymentPluginConfConfigFilesDetail(ctx, token, configDetails...)

		return err
	})

	return err
}

// ===============================================================================
// PluginWorkflow Related Interface
// ===============================================================================

// GetPluginWorkflow get plugin workflow.
func (s *Storage) GetPluginWorkflow(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error) {
	var (
		pluginWorkflow *types.PluginWorkflow
		err            error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginWorkflow, func(nCtx contextx.IContext) error {
		var err error
		pluginWorkflow, err = s.getPluginWorkflow(nCtx, workflowID)

		return err
	})

	return pluginWorkflow, err
}

// GetPluginWorkflowStatus get plugin workflow status.
func (s *Storage) GetPluginWorkflowStatus(nCtx contextx.IContext, workflowID string) (types.PluginWorkflowStatus, error) {
	var (
		status types.PluginWorkflowStatus
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginWorkflowStatus, func(nCtx contextx.IContext) error {
		var err error
		status, err = s.getPluginWorkflowStatus(nCtx, workflowID)

		return err
	})

	return status, err
}

// CreatePluginWorkflow createPluginDeployment plugin workflow.
func (s *Storage) CreatePluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginWorkflow, func(nCtx contextx.IContext) error {
		var err error
		err = s.createPluginWorkflow(nCtx, workflow)

		return err
	})

	return err
}

// UpdatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) UpdatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdatePluginWorkflowStatus, func(nCtx contextx.IContext) error {
		var err error
		err = s.updatePluginWorkflowStatus(nCtx, workflowID, status)

		return err
	})

	return err
}

// CountPluginWorkflow count plugin workflow.
func (s *Storage) CountPluginWorkflow(nCtx contextx.IContext, conditions ...*types.PluginWorkflowCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountPluginWorkflow, func(nCtx contextx.IContext) error {
		var err error
		count, err = s.countPluginWorkflow(nCtx, conditions...)

		return err
	})

	return count, err
}

// ListPluginWorkflow list plugin workflow.
func (s *Storage) ListPluginWorkflow(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginWorkflowCondition) (
	[]*types.PluginWorkflow, int64, error) {

	var (
		workflows []*types.PluginWorkflow
		total     int64
		err       error
	)

	err = s.WrapFn(nCtx, metricOperationListPluginWorkflow, func(nCtx contextx.IContext) error {
		var err error
		workflows, total, err = s.listPluginWorkflow(nCtx, page, conditions...)

		return err
	})

	return workflows, total, err
}

// DistinctPluginWorkflow distinct plugin workflow fields.
func (s *Storage) DistinctPluginWorkflow(
	nCtx contextx.IContext, request types.PluginWorkflowDistinctRequest, conditions ...*types.PluginWorkflowCondition) (
	*types.PluginWorkflowDistinctResult, error) {

	var (
		result *types.PluginWorkflowDistinctResult
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctPluginWorkflow, func(nCtx contextx.IContext) error {
		var err error
		result, err = s.distinctPluginWorkflow(nCtx, request, conditions...)

		return err
	})

	return result, err
}

// ===============================================================================
// Plugin Related Interface
// ===============================================================================

// GetPlugin get plugin by id.
func (s *Storage) GetPlugin(nCtx contextx.IContext, pluginName string) (*types.Plugin, error) {
	var (
		plugin *types.Plugin
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationGetPlugin, func(nCtx contextx.IContext) error {
		var err error
		plugin, err = s.getPlugin(nCtx, pluginName)

		return err
	})

	return plugin, err
}

// CountPlugins count plugins.
func (s *Storage) CountPlugins(nCtx contextx.IContext, conditions ...*types.PluginCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountPlugins, func(nCtx contextx.IContext) error {
		var err error
		count, err = s.countPlugins(nCtx, conditions...)

		return err
	})

	return count, err
}

// ListPlugins list plugins.
func (s *Storage) ListPlugins(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginCondition) ([]*types.Plugin, int64, error) {
	var (
		plugins []*types.Plugin
		cnt     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListPlugins, func(nCtx contextx.IContext) error {
		var err error
		plugins, cnt, err = s.listPlugins(nCtx, page, conditions...)

		return err
	})

	return plugins, cnt, err
}

// ExistPluginByPluginName check plugin exist by plugin name.
func (s *Storage) ExistPluginByPluginName(nCtx contextx.IContext, pluginName string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistPluginByPluginName, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.existPluginByPluginName(nCtx, pluginName)

		return err
	})

	return exist, err
}

// ExistDefaultPluginByPluginPkgName check default plugin exist by plugin package name.
func (s *Storage) ExistDefaultPluginByPluginPkgName(nCtx contextx.IContext, pluginPkgName string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistDefaultPluginByPluginPkgName, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.existPluginByPluginPkgName(nCtx, pluginPkgName)

		return err
	})

	return exist, err
}

// SetPluginMemo set plugin memo by plugin name.
func (s *Storage) SetPluginMemo(nCtx contextx.IContext, pluginName string, memo string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationSetPluginMemo, func(nCtx contextx.IContext) error {
		var err error
		err = s.setPluginMemo(nCtx, pluginName, memo)

		return err
	})

	return err
}

// CreatePlugin create plugin.
func (s *Storage) CreatePlugin(nCtx contextx.IContext, plugin *types.Plugin) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePlugin, func(nCtx contextx.IContext) error {
		var err error
		err = s.createPlugin(nCtx, plugin)

		return err
	})

	return err
}

// UpsertManyPlugins upsert many plugins.
func (s *Storage) UpsertManyPlugins(nCtx contextx.IContext, plugins ...*types.Plugin) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyPlugins, func(nCtx contextx.IContext) error {
		var err error
		err = s.upsertManyPlugins(nCtx, plugins...)

		return err
	})

	return err
}

// GetPluginVisibleBizIDs get plugin visible biz ids.
func (s *Storage) GetPluginVisibleBizIDs(nCtx contextx.IContext, pluginName string) ([]int64, error) {
	var (
		err    error
		bizIDs []int64
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginVisibleBizIDs, func(nCtx contextx.IContext) error {
		var err error
		bizIDs, err = s.getPluginVisibleBizIDs(nCtx, pluginName)

		return err
	})

	return bizIDs, err
}

// ListVisiblePluginByBizIDs list visible plugin by biz ids.
func (s *Storage) ListVisiblePluginByBizIDs(nCtx contextx.IContext, bizIDs []int64) ([]*types.Plugin, error) {
	var (
		err     error
		plugins []*types.Plugin
	)

	err = s.WrapFn(nCtx, metricOperationListVisiblePluginByBizIDs, func(nCtx contextx.IContext) error {
		var err error
		plugins, err = s.listVisiblePluginByBizIDs(nCtx, bizIDs)

		return err
	})

	return plugins, err
}

// ===============================================================================
// Process Related Interface
// ===============================================================================

// CountProcesses count processes.
func (s *Storage) CountProcesses(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountProcess, func(nCtx contextx.IContext) error {
		var err error
		count, err = s.countProcesses(nCtx, conditions...)

		return err
	})

	return count, err
}

// ListProcesses list processes.
func (s *Storage) ListProcesses(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessCondition) ([]*types.Process, int64, error) {
	var (
		processes []*types.Process
		total     int64
		err       error
	)

	err = s.WrapFn(nCtx, metricOperationListProcess, func(nCtx contextx.IContext) error {
		var err error
		processes, total, err = s.listProcesses(nCtx, page, conditions...)

		return err
	})

	return processes, total, err
}

// CreateProcess create process.
func (s *Storage) CreateProcess(nCtx contextx.IContext, process *types.Process) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreateProcess, func(nCtx contextx.IContext) error {
		var err error
		err = s.createProcess(nCtx, process)

		return err
	})

	return err
}

// UpdateProcess create process.
func (s *Storage) UpdateProcess(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateProcess, func(nCtx contextx.IContext) error {
		var err error
		err = s.updateProcess(nCtx, process, hostID, pluginName)

		return err
	})

	return err
}

// UpdateProcessInfo update process info.
func (s *Storage) UpdateProcessInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateProcessInfo, func(nCtx contextx.IContext) error {
		var err error
		err = s.updateProcessInfo(nCtx, hostID, pluginName, processInfo)

		return err
	})

	return err
}

// UpdateManyProcessInfo batch update process info by process ID.
func (s *Storage) UpdateManyProcessInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateManyProcessInfo, func(nCtx contextx.IContext) error {
		var err error
		err = s.updateManyProcessInfo(nCtx, processInfoDeltas)

		return err
	})

	return err
}

// UpdateProcessManyHostBizID update process biz id for many host.
func (s *Storage) UpdateProcessManyHostBizID(nCtx contextx.IContext, bizID int64, hostID ...int64) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateProcessBizID, func(nCtx contextx.IContext) error {
		var err error
		err = s.updateProcessManyHostBizID(nCtx, bizID, hostID...)

		return err
	})

	return err
}

// DeleteProcess delete process.
func (s *Storage) DeleteProcess(nCtx contextx.IContext, hostID int64, pluginName string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProcess, func(nCtx contextx.IContext) error {
		var err error
		err = s.deleteProcess(nCtx, hostID, pluginName)

		return err
	})

	return err
}

// ExistProcess exist process id.
func (s *Storage) ExistProcess(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistProcess, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.existProcess(nCtx, hostID, pluginName)

		return err
	})

	return exist, err
}

// GetProcess get process.
func (s *Storage) GetProcess(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error) {
	var (
		process *types.Process
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcess, func(nCtx contextx.IContext) error {
		var err error
		process, err = s.getProcess(nCtx, hostID, pluginName)

		return err
	})

	return process, err
}

// GetProcessDistributionByHostID get process distribution by host id.
func (s *Storage) GetProcessDistributionByHostID(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[int64]int64, error) {
	var (
		dist map[int64]int64
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessDistributionByHostID, func(nCtx contextx.IContext) error {
		var err error
		dist, err = s.getProcessDistributionByHostID(nCtx, condition...)

		return err
	})

	return dist, err
}

// GetProcessDistributionByPluginName get process distribution by plugin name.
func (s *Storage) GetProcessDistributionByPluginName(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[string]int64, error) {
	var (
		dist map[string]int64
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessDistributionByPluginName, func(nCtx contextx.IContext) error {
		var err error
		dist, err = s.getProcessDistributionByPluginName(nCtx, condition...)

		return err
	})

	return dist, err
}

// DistinctProcess distinct process.
func (s *Storage) DistinctProcess(nCtx contextx.IContext, request types.ProcessDistinctSelector, condition ...*types.ProcessCondition) (
	*types.ProcessDistinctResult, error) {

	var (
		result *types.ProcessDistinctResult
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctProcess, func(nCtx contextx.IContext) error {
		var err error
		result, err = s.distinctProcess(nCtx, request, condition...)

		return err
	})

	return result, err
}

// ===============================================================================
// ProcessConfig Related Interface
// ===============================================================================

// GetProcessConfig get process config.
func (s *Storage) GetProcessConfig(nCtx contextx.IContext, processUniqueKey *types.ProcessUniqueKey, name string) (*types.ProcessConfig, error) {
	var (
		processConfig *types.ProcessConfig
		err           error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessConfig, func(nCtx contextx.IContext) error {
		var err error
		processConfig, err = s.getProcessConfig(nCtx, processUniqueKey, name)

		return err
	})

	return processConfig, err
}

// CreateProcessConfig create process config.
func (s *Storage) CreateProcessConfig(nCtx contextx.IContext, processConfig *types.ProcessConfig) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreateProcessConfig, func(nCtx contextx.IContext) error {
		var err error
		err = s.createProcessConfig(nCtx, processConfig)

		return err
	})

	return err
}

// UpsertProcessConfigs upsert many process configs.
func (s *Storage) UpsertProcessConfigs(nCtx contextx.IContext, processConfigs ...*types.ProcessConfig) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyProcessConfigs, func(nCtx contextx.IContext) error {
		var err error
		err = s.upsertProcessConfigs(nCtx, processConfigs...)

		return err
	})

	return err
}

// DeleteProcessConfigsByProcessUniqueKey delete process configs by process unique key.
func (s *Storage) DeleteProcessConfigsByProcessUniqueKey(nCtx contextx.IContext, processUniqueKeys ...*types.ProcessUniqueKey) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProcessConfigsByProcessUniqueKey, func(nCtx contextx.IContext) error {
		var err error
		err = s.deleteProcessConfigsByProcessUniqueKey(nCtx, processUniqueKeys...)

		return err
	})

	return err
}

// DeleteProcessConfigs delete process configs.
func (s *Storage) DeleteProcessConfigs(nCtx contextx.IContext, processUniqueKey *types.ProcessUniqueKey, names ...string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProcessConfigs, func(nCtx contextx.IContext) error {
		var err error
		err = s.deleteProcessConfigs(nCtx, processUniqueKey, names...)

		return err
	})

	return err
}

// ListProcessConfigs list process configs.
func (s *Storage) ListProcessConfigs(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessConfigCondition) (
	[]*types.ProcessConfig, int64, error) {

	var (
		processConfigs []*types.ProcessConfig
		total          int64
		err            error
	)

	err = s.WrapFn(nCtx, metricOperationListProcessConfig, func(nCtx contextx.IContext) error {
		var err error
		processConfigs, total, err = s.listProcessConfigs(nCtx, page, conditions...)

		return err
	})

	return processConfigs, total, err
}

// CountProcessConfigs count process configs.
func (s *Storage) CountProcessConfigs(nCtx contextx.IContext, conditions ...*types.ProcessConfigCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountProcessConfig, func(nCtx contextx.IContext) error {
		var err error
		count, err = s.countProcessConfigs(nCtx, conditions...)

		return err
	})

	return count, err
}
