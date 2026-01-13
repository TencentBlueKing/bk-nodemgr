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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin"
	plugindeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-deployment"
	pluginworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process"
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

	metricOperationListPluginDeployment                           = "list_plugin_deployment"
	metricOperationGetPluginWorkflowStatus                        = "get_plugin_workflow_status"
	metricOperationCountPluginWorkflow                            = "count_plugin_workflow"
	metricOperationListPluginWorkflow                             = "list_plugin_workflow"
	metricOperationCountProcess                                   = "count_process"
	metricOperationListProcess                                    = "list_process"
	metricOperationGetProcessDistributionByHostID                 = "get_process_distribution_by_host_id"
	metricOperationGetProcessDistributionByPluginName             = "get_process_distribution_by_plugin_name"
	metricOperationDistinctProcess                                = "distinct_process"
	metricOperationCreateProcess                                  = "create_process"
	metricOperationUpdateProcess                                  = "update_process"
	metricOperationUpdateProcessInfo                              = "update_process_info"
	metricOperationDeleteProcess                                  = "delete_process"
	metricOperationExistProcess                                   = "exist_process"
	metricOperationUpdateManyProcessInfo                          = "update_many_process_info"
	metricOperationGetProcess                                     = "get_process"
	metricOperationGetPluginWorkflow                              = "get_plugin_workflow"
	metricOperationCreatePluginWorkflow                           = "create_plugin_workflow"
	metricOperationUpdatePluginWorkflowStatus                     = "update_plugin_workflow_status"
	metricOperationDistinctPluginWorkflow                         = "distinct_plugin_workflow"
	metricOperationGetPlugin                                      = "get_plugin"
	metricOperationCountPlugins                                   = "count_plugins"
	metricOperationListPlugins                                    = "list_plugins"
	metricOperationExistPluginByPluginName                        = "exist_plugin_by_plugin_name"
	metricOperationExistDefaultPluginByPluginPkgName              = "exist_default_plugin_by_plugin_pkg_name"
	metricOperationSetPluginMemo                                  = "set_plugin_memo"
	metricOperationCreatePlugin                                   = "create_plugin"
	metricOperationGetPluginDeploymentInfo                        = "get_plugin_deployment_info"
	metricOperationCreatePluginDeployment                         = "create_plugin_deployment"
	metricOperationUpdatePluginDeploymentInfo                     = "update_plugin_deployment_info"
	metricOperationGetPluginDeploymentPluginConf                  = "get_plugin_deployment_plugin_conf"
	metricOperationUpdatePluginDeploymentPluginConf               = "update_plugin_deployment_plugin_conf"
	metricOperationGetPluginDeploymentPluginConfConfigFilesDetail = "get_plugin_deployment_plugin_conf_config_files_detail"
	metricOperationUpsertManyPlugins                              = "upsert_many_plugins"
	metricOperationGetProcessConfig                               = "get_process_config"
	metricOperationCreateProcessConfig                            = "create_process_config"
	metricOperationUpsertManyProcessConfigs                       = "upsert_many_process_configs"
	metricOperationDeleteProcessConfigsByProcessID                = "delete_process_configs_by_process_id"
	metricOperationDeleteProcessConfigs                           = "delete_process_configs"
	metricOperationListProcessConfig                              = "list_process_config"
	metricOperationCountProcessConfig                             = "count_process_config"
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
	daoProcess          process.IHandler
	daoProcessConfig    daoProcessConfig.IHandler
	daoOperInstData     operinstdata.IHandler

	monitoredWorkflows      map[string]*types.PluginWorkflow
	monitoredWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoPluginDeployment = plugindeployment.New(s.Database)
	s.daoPluginWorkflow = pluginworkflow.New(s.Database)
	s.daoPlugin = plugin.New(s.Database)
	s.daoProcess = process.New(s.Database)
	s.daoProcessConfig = daoProcessConfig.New(s.Database)
	s.daoOperInstData = operinstdata.New(s.Database)

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

	runningWorkflowMap := map[string]*types.PluginWorkflow{}
	recentFinishedWorkflowMap := map[string]*types.PluginWorkflow{}

	gp := gopool.NewPool()
	for _, tenantID := range tenantIDs {
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

			// we can sure that the workflow id is unique, so we can use map in concurrency.
			for _, workflow := range runningWorkflows {
				runningWorkflowMap[workflow.WorkflowID] = workflow
			}

			for _, workflow := range recentFinishedWorkflows {
				recentFinishedWorkflowMap[workflow.WorkflowID] = workflow
			}

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to obtain monitored workflows")
	}

	s.monitoredWorkflowsMutex.Lock()

	s.monitoredWorkflows = make(map[string]*types.PluginWorkflow, len(s.monitoredWorkflows))
	for _, workflow := range runningWorkflowMap {
		s.monitoredWorkflows[workflow.TriggerID] = workflow
	}

	// notice: When these maps intersect,
	// you need to ensure that the workflow in the recentFinishedWorkflowMap has a higher priority
	for _, workflow := range recentFinishedWorkflowMap {
		s.monitoredWorkflows[workflow.TriggerID] = workflow
	}

	s.monitoredWorkflowsMutex.Unlock()

	return nil
}

func (s *Storage) monitorWorkflowStatus(nCtx contextx.IContext) error {
	s.monitoredWorkflowsMutex.RLock()
	defer s.monitoredWorkflowsMutex.RUnlock()

	if len(s.monitoredWorkflows) == 0 {
		return nil
	}

	operInst, err := s.daoOperInstData.ListAllLastOperInst(nCtx,
		operinstdata.WithTriggerID(conv.MapKeyToSlice(s.monitoredWorkflows)...))
	if err != nil {
		return fmt.Errorf("query last operation instance failed: %w", err)
	}

	unfinishedTriggerMap := make(map[string]struct{}, len(operInst))
	for _, inst := range operInst {
		if !operation.CheckStateFinished(inst.Lifecycle.State) {
			unfinishedTriggerMap[inst.Metadata.TriggerID] = struct{}{}
		}
	}

	finishedTriggerOperInstsMap := make(map[string][]*operation.InstanceBriefData, len(operInst))
	for _, inst := range operInst {
		if _, ok := unfinishedTriggerMap[inst.Metadata.TriggerID]; !ok {
			finishedTriggerOperInstsMap[inst.Metadata.TriggerID] =
				append(finishedTriggerOperInstsMap[inst.Metadata.TriggerID], inst)
		}
	}

	for triggerID, operInsts := range finishedTriggerOperInstsMap {
		status, finishTime := calWorkflowStatusAndTime(operInsts)

		pluginWorkflow, ok := s.monitoredWorkflows[triggerID]
		if !ok {
			continue
		}

		tenantNCtx := contextx.From(nCtx, contextx.WithTenantID(pluginWorkflow.TenantID))

		err = s.daoPluginWorkflow.UpdateStatus(tenantNCtx, pluginWorkflow.WorkflowID, status)
		if err != nil {
			return fmt.Errorf("update plugin workflow status failed: %w", err)
		}

		err = s.daoPluginWorkflow.UpdateFinishTime(
			tenantNCtx, pluginWorkflow.WorkflowID, finishTime)
		if err != nil {
			return fmt.Errorf("update plugin workflow finish time failed: %w", err)
		}

		delete(s.monitoredWorkflows, triggerID)
	}

	return nil
}

func calWorkflowStatusAndTime(operationInsts []*operation.InstanceBriefData) (types.PluginWorkflowStatus, time.Time) {
	successCount := 0
	failedCount := 0
	var latestEndTime time.Time

	for _, inst := range operationInsts {
		if inst.Lifecycle.EndedAt.After(latestEndTime) {
			latestEndTime = inst.Lifecycle.EndedAt
		}

		switch inst.Lifecycle.State {
		case operation.StateSuccess:
			successCount++
		case operation.StateFailed, operation.StateTimeout:
			failedCount++
		default:
		}
	}

	total := len(operationInsts)
	switch {
	case successCount == total:
		return types.PluginWorkflowStatusSuccess, latestEndTime
	case failedCount == total:
		return types.PluginWorkflowStatusFailed, latestEndTime
	default:
		return types.PluginWorkflowStatusPartialFailed, latestEndTime
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
		info, err = s.getPluginDeploymentInfo(nCtx, token)

		return err
	})
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CreatePluginDeployment plugin deployment.
func (s *Storage) CreatePluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginDeployment, func(nCtx contextx.IContext) error {
		err = s.createPluginDeployment(nCtx, pluginDeployment)

		return err
	})
	if err != nil {
		return err
	}

	return nil
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
		pluginDeployments, total, err = s.listPluginDeployment(nCtx, page, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return pluginDeployments, total, nil
}

// UpdatePluginDeploymentInfo update a plugin deployment info.
func (s *Storage) UpdatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdatePluginDeploymentInfo, func(nCtx contextx.IContext) error {
		err = s.updatePluginDeploymentInfo(nCtx, token, pluginDeploymentInfo)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// GetPluginDeploymentPluginConf get plugin deployment plugin conf.
func (s *Storage) GetPluginDeploymentPluginConf(ctx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error) {
	var (
		pluginConf *types.PluginDeploymentPluginConf
		err        error
	)

	err = s.WrapFn(ctx, metricOperationGetPluginDeploymentPluginConf, func(ctx contextx.IContext) error {
		pluginConf, err = s.getPluginDeploymentPluginConf(ctx, token)

		return err
	})
	if err != nil {
		return nil, err
	}

	return pluginConf, nil
}

// UpdatePluginDeploymentPluginConf set plugin deployment plugin conf.
func (s *Storage) UpdatePluginDeploymentPluginConf(ctx contextx.IContext, token string, pluginConf *types.PluginDeploymentPluginConf) error {
	var (
		err error
	)

	err = s.WrapFn(ctx, metricOperationUpdatePluginDeploymentPluginConf, func(ctx contextx.IContext) error {
		err = s.updatePluginDeploymentPluginConf(ctx, token, pluginConf)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// GetPluginDeploymentPluginConfConfigFilesDetail get plugin deployment plugin conf config files detail.
func (s *Storage) GetPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string) ([]*types.PluginConfigDetail, error) {
	var (
		config []*types.PluginConfigDetail
		err    error
	)

	err = s.WrapFn(ctx, metricOperationGetPluginDeploymentPluginConfConfigFilesDetail, func(ctx contextx.IContext) error {
		config, err = s.getPluginDeploymentPluginConfConfigFilesDetail(ctx, token)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return config, nil
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
		pluginWorkflow, err = s.getPluginWorkflow(nCtx, workflowID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pluginWorkflow, nil
}

// GetPluginWorkflowStatus get plugin workflow status.
func (s *Storage) GetPluginWorkflowStatus(nCtx contextx.IContext, workflowID string) (types.PluginWorkflowStatus, error) {
	var (
		status types.PluginWorkflowStatus
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationGetPluginWorkflowStatus, func(nCtx contextx.IContext) error {
		status, err = s.getPluginWorkflowStatus(nCtx, workflowID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return status, nil
}

// CreatePluginWorkflow createPluginDeployment plugin workflow.
func (s *Storage) CreatePluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePluginWorkflow, func(nCtx contextx.IContext) error {
		err = s.createPluginWorkflow(nCtx, workflow)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) UpdatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdatePluginWorkflowStatus, func(nCtx contextx.IContext) error {
		err = s.updatePluginWorkflowStatus(nCtx, workflowID, status)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// CountPluginWorkflow count plugin workflow.
func (s *Storage) CountPluginWorkflow(nCtx contextx.IContext, conditions ...*types.PluginWorkflowCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountPluginWorkflow, func(nCtx contextx.IContext) error {
		count, err = s.countPluginWorkflow(nCtx, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
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
		workflows, total, err = s.listPluginWorkflow(nCtx, page, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
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
		result, err = s.distinctPluginWorkflow(nCtx, request, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
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
		plugin, err = s.getPlugin(nCtx, pluginName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return plugin, nil
}

// CountPlugins count plugins.
func (s *Storage) CountPlugins(nCtx contextx.IContext, conditions ...*types.PluginCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountPlugins, func(nCtx contextx.IContext) error {
		count, err = s.countPlugins(nCtx, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListPlugins list plugins.
func (s *Storage) ListPlugins(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginCondition) ([]*types.Plugin, int64, error) {
	var (
		plugins []*types.Plugin
		cnt     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListPlugins, func(nCtx contextx.IContext) error {
		plugins, cnt, err = s.listPlugins(nCtx, page, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return plugins, cnt, nil
}

// ExistPluginByPluginName check plugin exist by plugin name.
func (s *Storage) ExistPluginByPluginName(nCtx contextx.IContext, pluginName string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistPluginByPluginName, func(nCtx contextx.IContext) error {
		exist, err = s.existPluginByPluginName(nCtx, pluginName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return exist, nil
}

// ExistDefaultPluginByPluginPkgName check default plugin exist by plugin package name.
func (s *Storage) ExistDefaultPluginByPluginPkgName(nCtx contextx.IContext, pluginPkgName string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistDefaultPluginByPluginPkgName, func(nCtx contextx.IContext) error {
		exist, err = s.existPluginByPluginPkgName(nCtx, pluginPkgName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return exist, nil
}

// SetPluginMemo set plugin memo by plugin name.
func (s *Storage) SetPluginMemo(nCtx contextx.IContext, pluginName string, memo string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationSetPluginMemo, func(nCtx contextx.IContext) error {
		err = s.setPluginMemo(nCtx, pluginName, memo)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// CreatePlugin create plugin.
func (s *Storage) CreatePlugin(nCtx contextx.IContext, plugin *types.Plugin) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreatePlugin, func(nCtx contextx.IContext) error {
		err = s.createPlugin(nCtx, plugin)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// UpsertManyPlugins upsert many plugins.
func (s *Storage) UpsertManyPlugins(nCtx contextx.IContext, plugins ...*types.Plugin) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyPlugins, func(nCtx contextx.IContext) error {
		err = s.upsertManyPlugins(nCtx, plugins...)
		return err
	})
	if err != nil {
		return err
	}

	return nil
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
		count, err = s.countProcesses(nCtx, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListProcesses list processes.
func (s *Storage) ListProcesses(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessCondition) ([]*types.Process, int64, error) {
	var (
		processes []*types.Process
		total     int64
		err       error
	)

	err = s.WrapFn(nCtx, metricOperationListProcess, func(nCtx contextx.IContext) error {
		processes, total, err = s.listProcesses(nCtx, page, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return processes, total, nil
}

// CreateProcess create process.
func (s *Storage) CreateProcess(nCtx contextx.IContext, process *types.Process) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreateProcess, func(nCtx contextx.IContext) error {
		err = s.createProcess(nCtx, process)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcess create process.
func (s *Storage) UpdateProcess(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateProcess, func(nCtx contextx.IContext) error {
		err = s.updateProcess(nCtx, process, hostID, pluginName)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessInfo update process info.
func (s *Storage) UpdateProcessInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateProcessInfo, func(nCtx contextx.IContext) error {
		err = s.updateProcessInfo(nCtx, hostID, pluginName, processInfo)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteProcess delete process.
func (s *Storage) DeleteProcess(nCtx contextx.IContext, hostID int64, pluginName string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProcess, func(nCtx contextx.IContext) error {
		err = s.deleteProcess(nCtx, hostID, pluginName)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// ExistProcess exist process id.
func (s *Storage) ExistProcess(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistProcess, func(nCtx contextx.IContext) error {
		exist, err = s.existProcess(nCtx, hostID, pluginName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return exist, nil
}

// UpdateManyProcessInfo batch update process info by process ID.
func (s *Storage) UpdateManyProcessInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateManyProcessInfo, func(nCtx contextx.IContext) error {
		err = s.updateManyProcessInfo(nCtx, processInfoDeltas)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// GetProcess get process.
func (s *Storage) GetProcess(nCtx contextx.IContext, hostID int64, pluginName string) (*types.Process, error) {
	var (
		process *types.Process
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcess, func(nCtx contextx.IContext) error {
		process, err = s.getProcess(nCtx, hostID, pluginName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return process, nil
}

// GetProcessDistributionByHostID get process distribution by host id.
func (s *Storage) GetProcessDistributionByHostID(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[int64]int64, error) {
	var (
		dist map[int64]int64
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessDistributionByHostID, func(nCtx contextx.IContext) error {
		dist, err = s.getProcessDistributionByHostID(nCtx, condition...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dist, nil
}

// GetProcessDistributionByPluginName get process distribution by plugin name.
func (s *Storage) GetProcessDistributionByPluginName(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[string]int64, error) {
	var (
		dist map[string]int64
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessDistributionByPluginName, func(nCtx contextx.IContext) error {
		dist, err = s.getProcessDistributionByPluginName(nCtx, condition...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dist, nil
}

// DistinctProcess distinct process.
func (s *Storage) DistinctProcess(nCtx contextx.IContext, request types.ProcessDistinctSelector, condition ...*types.ProcessCondition) (
	*types.ProcessDistinctResult, error) {

	var (
		result *types.ProcessDistinctResult
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctProcess, func(nCtx contextx.IContext) error {
		result, err = s.distinctProcess(nCtx, request, condition...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ===============================================================================
// ProcessConfig Related Interface
// ===============================================================================

// GetProcessConfig get process config.
func (s *Storage) GetProcessConfig(nCtx contextx.IContext, processID, name string) (*types.ProcessConfig, error) {
	var (
		processConfig *types.ProcessConfig
		err           error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessConfig, func(nCtx contextx.IContext) error {
		processConfig, err = s.getProcessConfig(nCtx, processID, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return processConfig, nil
}

// CreateProcessConfig create process config.
func (s *Storage) CreateProcessConfig(nCtx contextx.IContext, processConfig *types.ProcessConfig) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCreateProcessConfig, func(nCtx contextx.IContext) error {
		err = s.createProcessConfig(nCtx, processConfig)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// UpsertProcessConfigs upsert many process configs.
func (s *Storage) UpsertProcessConfigs(nCtx contextx.IContext, processConfigs ...*types.ProcessConfig) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationUpsertManyProcessConfigs, func(nCtx contextx.IContext) error {
		err = s.upsertProcessConfigs(nCtx, processConfigs...)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteProcessConfigsByProcessID delete process configs by process id.
func (s *Storage) DeleteProcessConfigsByProcessID(nCtx contextx.IContext, processID ...string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProcessConfigsByProcessID, func(nCtx contextx.IContext) error {
		err = s.deleteProcessConfigsByProcessID(nCtx, processID...)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteProcessConfigs delete process configs.
func (s *Storage) DeleteProcessConfigs(nCtx contextx.IContext, processID string, names ...string) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricOperationDeleteProcessConfigs, func(nCtx contextx.IContext) error {
		err = s.deleteProcessConfigs(nCtx, processID, names...)

		return err
	})
	if err != nil {
		return err
	}

	return nil
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
		processConfigs, total, err = s.listProcessConfigs(nCtx, page, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return processConfigs, total, nil
}

// CountProcessConfigs count process configs.
func (s *Storage) CountProcessConfigs(nCtx contextx.IContext, conditions ...*types.ProcessConfigCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationCountProcessConfig, func(nCtx contextx.IContext) error {
		count, err = s.countProcessConfigs(nCtx, conditions...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}
