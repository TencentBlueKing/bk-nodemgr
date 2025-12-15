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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin"
	plugindeployment "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-deployment"
	pluginworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/plugin-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "plugin"

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
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage this is a storage to operate node deployment table.
type Storage struct {
	basestorage.Storage

	// dao
	daoPluginDeployment plugindeployment.IHandler
	daoPluginWorkflow   pluginworkflow.IHandler
	daoPlugin           plugin.IHandler
	daoProcess          process.IHandler
}

func (s *Storage) initDao() error {
	s.daoPluginDeployment = plugindeployment.New(s.Database)
	s.daoPluginWorkflow = pluginworkflow.New(s.Database)
	s.daoPlugin = plugin.New(s.Database)
	s.daoProcess = process.New(s.Database)

	return nil
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

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// GetPluginDeploymentInfo get plugin deployment info.
func (s *Storage) GetPluginDeploymentInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	var (
		info *types.PluginDeploymentInfo
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_info")
	defer metric.End(err)

	info, err = s.getPluginDeploymentInfo(nCtx, token)

	return info, err
}

// CreatePluginDeployment plugin deployment.
func (s *Storage) CreatePluginDeployment(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("createPluginDeployment")
	defer metric.End(err)

	err = s.createPluginDeployment(nCtx, pluginDeployment)

	return err
}

// UpdatePluginDeploymentInfo update a node deployment info.
func (s *Storage) UpdatePluginDeploymentInfo(nCtx contextx.IContext, token string, pluginDeploymentInfo *types.PluginDeploymentInfo) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_info")
	defer metric.End(err)

	err = s.updatePluginDeploymentInfo(nCtx, token, pluginDeploymentInfo)

	return err
}

// GetPluginDeploymentPluginConf get plugin deployment plugin conf.
func (s *Storage) GetPluginDeploymentPluginConf(ctx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error) {
	var (
		pluginConf *types.PluginDeploymentPluginConf
		err        error
	)

	// record metric.
	metric := s.metric().Start("get_plugin_deployment_plugin_conf")
	defer metric.End(err)

	pluginConf, err = s.getPluginDeploymentPluginConf(ctx, token)

	return pluginConf, err
}

// UpdatePluginDeploymentPluginConf set plugin deployment plugin conf.
func (s *Storage) UpdatePluginDeploymentPluginConf(ctx contextx.IContext, token string, pluginConf *types.PluginDeploymentPluginConf) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_plugin_deployment_plugin_conf")
	defer metric.End(err)

	err = s.updatePluginDeploymentPluginConf(ctx, token, pluginConf)

	return err
}

// GetPluginDeploymentPluginConfConfigFilesDetail get plugin deployment plugin conf config files detail.
func (s *Storage) GetPluginDeploymentPluginConfConfigFilesDetail(ctx contextx.IContext, token string) (
	config []*types.PluginConfigDetail, err error) {

	// record metric.
	metric := s.metric().Start("get_plugin_deployment_plugin_conf_config_files_detail")
	defer metric.End(err)

	config, err = s.getPluginDeploymentPluginConfConfigFilesDetail(ctx, token)

	return config, err
}

// GetPluginWorkflow get plugin workflow.
func (s *Storage) GetPluginWorkflow(nCtx contextx.IContext, workflowID string) (*types.PluginWorkflow, error) {
	var (
		pluginWorkflow *types.PluginWorkflow
		err            error
	)

	// record metric.
	metric := s.metric().Start("create_plugin_workflow")
	defer metric.End(err)

	pluginWorkflow, err = s.getPluginWorkflow(nCtx, workflowID)

	return pluginWorkflow, err
}

// CreatePluginWorkflow createPluginDeployment plugin workflow.
func (s *Storage) CreatePluginWorkflow(nCtx contextx.IContext, workflow *types.PluginWorkflow) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_plugin_workflow")
	defer metric.End(err)

	err = s.createPluginWorkflow(nCtx, workflow)

	return err
}

// UpdatePluginWorkflowStatus update plugin workflow status.
func (s *Storage) UpdatePluginWorkflowStatus(nCtx contextx.IContext, workflowID string, status types.PluginWorkflowStatus) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update_plugin_workflow_status")
	defer metric.End(err)

	err = s.updatePluginWorkflowStatus(nCtx, workflowID, status)

	return err
}

// CountProcesses count processes.
func (s *Storage) CountProcesses(nCtx contextx.IContext, conditions ...*types.ProcessCondition) (count int64, err error) {
	// record metric.
	metric := s.metric().Start("count_processes")
	defer metric.End(err)

	count, err = s.countProcesses(nCtx, conditions...)

	return count, err
}

// ListProcesses list processes.
func (s *Storage) ListProcesses(nCtx contextx.IContext, page types.Page, conditions ...*types.ProcessCondition) (
	processes []*types.Process, total int64, err error) {

	// record metric.
	metric := s.metric().Start("list_processes")
	defer metric.End(err)

	processes, total, err = s.listProcesses(nCtx, page, conditions...)

	return processes, total, err
}

// GetPlugin get plugin by id.
func (s *Storage) GetPlugin(nCtx contextx.IContext, pluginName string) (plugin *types.Plugin, err error) {
	// record metric.
	metric := s.metric().Start("get_plugin")
	defer metric.End(err)

	plugin, err = s.getPlugin(nCtx, pluginName)

	return plugin, err
}

// CountPlugins count plugins.
func (s *Storage) CountPlugins(nCtx contextx.IContext, conditions ...*types.PluginCondition) (count int64, err error) {
	// record metric.
	metric := s.metric().Start("count_plugins")
	defer metric.End(err)

	count, err = s.countPlugins(nCtx, conditions...)

	return count, err
}

// ListPlugins list plugins.
func (s *Storage) ListPlugins(nCtx contextx.IContext, page types.Page, conditions ...*types.PluginCondition) (plugins []*types.Plugin,
	cnt int64, err error) {
	// record metric.
	metric := s.metric().Start("list_plugins")
	defer metric.End(err)

	plugins, cnt, err = s.listPlugins(nCtx, page, conditions...)

	return plugins, cnt, err
}

// ExistPluginByPluginName check plugin exist by plugin name.
func (s *Storage) ExistPluginByPluginName(nCtx contextx.IContext, pluginName string) (exist bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_plugin_by_plugin_name")
	defer metric.End(err)

	exist, err = s.existPluginByPluginName(nCtx, pluginName)

	return exist, err
}

// ExistDefaultPluginByPluginPkgName check default plugin exist by plugin package name.
func (s *Storage) ExistDefaultPluginByPluginPkgName(nCtx contextx.IContext, pluginPkgName string) (exist bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_plugin_by_plugin_pkg_name")
	defer metric.End(err)

	exist, err = s.existPluginByPluginPkgName(nCtx, pluginPkgName)

	return exist, err
}

// CreatePlugin create plugin.
func (s *Storage) CreatePlugin(nCtx contextx.IContext, plugin *types.Plugin) (err error) {
	// record metric.
	metric := s.metric().Start("create_plugin")
	defer metric.End(err)

	err = s.createPlugin(nCtx, plugin)

	return err
}

// CreateProcess create process.
func (s *Storage) CreateProcess(nCtx contextx.IContext, process *types.Process) (err error) {
	// record metric.
	metric := s.metric().Start("create_process")
	defer metric.End(err)

	err = s.createProcess(nCtx, process)

	return err
}

// UpdateProcess create process.
func (s *Storage) UpdateProcess(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) (err error) {
	// record metric.
	metric := s.metric().Start("update_process")
	defer metric.End(err)

	err = s.updateProcess(nCtx, process, hostID, pluginName)

	return err
}

// UpdateProcessInfo update process info.
func (s *Storage) UpdateProcessInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) (err error) {
	// record metric.
	metric := s.metric().Start("update_process_info")
	defer metric.End(err)

	err = s.updateProcessInfo(nCtx, hostID, pluginName, processInfo)

	return err
}

// DeleteProcess delete process.
func (s *Storage) DeleteProcess(nCtx contextx.IContext, hostID int64, pluginName string) (err error) {
	// record metric.
	metric := s.metric().Start("delete_process")
	defer metric.End(err)

	err = s.deleteProcess(nCtx, hostID, pluginName)

	return err
}

// ExistProcess exist process id.
func (s *Storage) ExistProcess(nCtx contextx.IContext, hostID int64, pluginName string) (exist bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_process_id")
	defer metric.End(err)

	exist, err = s.existProcess(nCtx, hostID, pluginName)

	return exist, err
}

// UpdateManyProcessInfo batch update process info by process ID.
func (s *Storage) UpdateManyProcessInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) (err error) {
	// record metric.
	metric := s.metric().Start("update_many_process_info")
	defer metric.End(err)

	err = s.updateManyProcessInfo(nCtx, processInfoDeltas)

	return err
}

// GetProcess get process.
func (s *Storage) GetProcess(nCtx contextx.IContext, hostID int64, pluginName string) (process *types.Process, err error) {
	// record metric.
	metric := s.metric().Start("get_process")
	defer metric.End(err)

	process, err = s.getProcess(nCtx, hostID, pluginName)

	return process, err
}

const (
	metricOperationGetProcessDistributionByHostID     = "get_process_distribution_by_host_id"
	metricOperationGetProcessDistributionByPluginName = "get_process_distribution_by_plugin_name"
	metricOperationDistinctProcess                    = "distinct_process"
)

// GetProcessDistributionByHostID get process distribution by host id.
func (s *Storage) GetProcessDistributionByHostID(nCtx contextx.IContext, condition ...*types.ProcessCondition) (map[int64]int64, error) {
	var (
		dist map[int64]int64
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetProcessDistributionByHostID, func(nCtx contextx.IContext) error {
		// record metric.
		metric := s.metric().Start(metricOperationGetProcessDistributionByHostID)
		defer metric.End(err)

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
		// record metric.
		metric := s.metric().Start(metricOperationGetProcessDistributionByPluginName)
		defer metric.End(err)

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
		// record metric.
		metric := s.metric().Start(metricOperationDistinctProcess)
		defer metric.End(err)

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
