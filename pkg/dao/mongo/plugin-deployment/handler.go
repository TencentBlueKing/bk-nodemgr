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

package plugindeployment

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node deployment Handler interface.
type IHandler interface {
	// Create create a node deployment.
	Create(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error

	// List list node deployments.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PluginDeployment, int64, error)

	// GetInfo get a plugin deployment info.
	GetInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error)

	// UpdateInfo update plugin deployment info.
	UpdateInfo(nCtx contextx.IContext, token string, info *types.PluginDeploymentInfo) error

	// GetPluginConf get plugin deployment plugin conf.
	GetPluginConf(nCtx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error)

	// UpdatePluginConf set plugin deployment plugin conf.
	UpdatePluginConf(nCtx contextx.IContext, token string, detail *types.PluginDeploymentPluginConf) error

	// GetPluginConfConfigFilesDetail get plugin conf config files detail.
	GetPluginConfConfigFilesDetail(nCtx contextx.IContext, token string) ([]*types.PluginConfigDetail, error)

	// UpsertPluginConfConfigFilesDetail update plugin conf config files detail.
	UpsertPluginConfConfigFilesDetail(nCtx contextx.IContext, token string, configDetails ...*types.PluginConfigDetail) error
}

// Handler this is a Handler to operate node deployment table.
type Handler struct {
	dao *dao
}

// New new a Handler.
func New(client *mongo.Database) *Handler {
	h := &Handler{
		dao: newDao(client),
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure plugin deployment indexes")
	}

	return h
}

// Create create a new node deployment.
func (h *Handler) Create(nCtx contextx.IContext, pluginDeployment *types.PluginDeployment) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if pluginDeployment == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertPluginDeploymentFromTypes(pluginDeployment)
	if err != nil {
		return err
	}

	return h.dao.Create(nCtx, data)
}

// List list node deployments.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PluginDeployment, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	deployments, err := h.dao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.PluginDeployment, len(deployments))
	for idx, deployment := range deployments {
		depTypes, err := convertPluginDeploymentToTypes(deployment)
		if err != nil {
			return nil, 0, err
		}

		data[idx] = depTypes
	}

	return data, num, nil
}

// GetInfo get a plugin deployment info.
func (h *Handler) GetInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(nCtx, filter, FieldKeyInfo)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertPluginDeploymentInfoToTypes(data.Info)
}

// UpdateInfo update a node deployment info.
func (h *Handler) UpdateInfo(nCtx contextx.IContext, token string, info *types.PluginDeploymentInfo) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return ErrInvalidToken()
	}

	if info == nil {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := convertPluginDeploymentInfoFromTypes(info)
	if err != nil {
		return err
	}
	if err := h.dao.UpdateField(nCtx, filter, FieldKeyInfo, data); err != nil {
		return err
	}

	return nil
}

// GetPluginConf get plugin deployment plugin conf.
func (h *Handler) GetPluginConf(nCtx contextx.IContext, token string) (*types.PluginDeploymentPluginConf, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(nCtx, filter, FieldKeyPluginConf)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertPluginDeploymentPluginConfToTypes(data.PluginConf), nil
}

// UpdatePluginConf set plugin deployment plugin conf.
func (h *Handler) UpdatePluginConf(nCtx contextx.IContext, token string, conf *types.PluginDeploymentPluginConf) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return ErrInvalidToken()
	}

	if conf == nil {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)

	data := convertPluginDeploymentPluginConfFromTypes(conf)
	if err := h.dao.UpdateField(nCtx, filter, FieldKeyPluginConf, data); err != nil {
		return err
	}

	return nil
}

// GetPluginConfConfigFilesDetail get plugin conf config files detail.
func (h *Handler) GetPluginConfConfigFilesDetail(nCtx contextx.IContext, token string) ([]*types.PluginConfigDetail, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(nCtx, filter, FieldKeyPluginConfConfigFilesDetail)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertPluginConfigDetailsToTypes(data.PluginConf.ConfigFilesDetail...), nil
}

// UpsertPluginConfConfigFilesDetail update plugin conf config files detail.
func (h *Handler) UpsertPluginConfConfigFilesDetail(nCtx contextx.IContext, token string, configDetails ...*types.PluginConfigDetail) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return ErrInvalidToken()
	}

	if len(configDetails) == 0 {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)

	dbData, err := h.dao.Get(nCtx, filter, FieldKeyPluginConfConfigFilesDetail)
	if err != nil {
		return err
	}

	confMap, err := conv.SliceToMap(dbData.PluginConf.ConfigFilesDetail, func(detail configDetail) string {
		return detail.Name
	})
	if err != nil {
		return err
	}

	data := convertPluginConfigDetailsFromTypes(configDetails...)
	for _, item := range data {
		confMap[item.Name] = item
	}

	mergedDetails := conv.MapValueToSlice(confMap)
	if err := h.dao.UpdateField(nCtx, filter, FieldKeyPluginConfConfigFilesDetail, mergedDetails); err != nil {
		return err
	}

	return nil
}

func convertPluginDeploymentFromTypes(data *types.PluginDeployment) (*Data, error) {
	pluginDeployment := &Data{
		Token:      data.Token,
		Info:       new(Info),
		PluginConf: new(PluginConf),
	}

	var err error

	pluginDeployment.Info, err = convertPluginDeploymentInfoFromTypes(data.Info)
	if err != nil {
		return nil, err
	}

	pluginDeployment.PluginConf = convertPluginDeploymentPluginConfFromTypes(data.PluginConf)

	return pluginDeployment, nil
}

func convertPluginDeploymentToTypes(data *Data) (*types.PluginDeployment, error) {
	pluginDeployment := &types.PluginDeployment{
		Token:      data.Token,
		Info:       nil,
		PluginConf: nil,
	}

	var err error
	pluginDeployment.Info, err = convertPluginDeploymentInfoToTypes(data.Info)
	if err != nil {
		return nil, err
	}

	pluginDeployment.PluginConf = convertPluginDeploymentPluginConfToTypes(data.PluginConf)

	return pluginDeployment, nil
}

// nolint: funlen
func convertPluginDeploymentInfoToTypes(info *Info) (*types.PluginDeploymentInfo, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	typesInfo := &types.PluginDeploymentInfo{
		BlockingActionName: info.ActionName,
		ConfigSource:       convertConfigSourceToTypes(info.ConfigSource),
		Process: types.Process{
			TenantID:      info.Process.TenantID,
			HostID:        info.Process.HostID,
			BizID:         info.Process.BizID,
			PluginName:    info.Process.Name,
			PluginGroup:   info.Process.Group,
			PluginPkgName: info.Process.PkgName,
			Platform: platfmt.Platform{
				OS:   criteria.OSType(info.Process.Platform.OS),
				Arch: criteria.CPUArch(info.Process.Platform.Arch),
			},
			Generation: types.Generation(info.Process.Generation),
			BindIP:     info.Process.BindIP,
			BindPort:   info.Process.BindPort,
			Info: types.ProcessInfo{
				Pid:        info.Process.Info.Pid,
				Version:    info.Process.Info.Version,
				AgentID:    info.Process.Info.AgentID,
				AutoStart:  info.Process.Info.AutoStart,
				Status:     types.ProcessStatus(info.Process.Info.Status),
				LastSyncAt: info.Process.Info.LastSyncAt,
			},
			Identity: types.ProcessIdentity{
				Name:       info.Process.Identity.Name,
				SetupPath:  info.Process.Identity.SetupPath,
				PidPath:    info.Process.Identity.PidPath,
				ConfigPath: info.Process.Identity.ConfigPath,
				LogPath:    info.Process.Identity.LogPath,
				User:       info.Process.Identity.User,
			},
			Controller: types.ProcessController{
				StartCmd:   info.Process.Controller.StartCmd,
				StopCmd:    info.Process.Controller.StopCmd,
				RestartCmd: info.Process.Controller.RestartCmd,
				ReloadCmd:  info.Process.Controller.ReloadCmd,
				DebugCmd:   info.Process.Controller.DebugCmd,
				KillCmd:    info.Process.Controller.KillCmd,
				VersionCmd: info.Process.Controller.VersionCmd,
				HealthCmd:  info.Process.Controller.HealthCmd,
			},
			Resource: types.ProcessResource{
				CPULimitPercent: info.Process.Resource.CPULimitPercent,
				MemLimitPercent: info.Process.Resource.MemLimitPercent,
			},
			MonitorPolicy: types.ProcessMonitorPolicy{
				RestartType:    types.ProcessRestartType(info.Process.MonitorPolicy.RestartType),
				StartCheckSecs: info.Process.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  info.Process.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  info.Process.MonitorPolicy.OpTimeoutSecs,
			},
		},
		InstallerRuntime: types.PluginDeploymentInstallerRuntime{
			BaseWorkDir: info.InstallerRuntime.BaseWorkDir,
			WorkDir:     info.InstallerRuntime.WorkDir,
		},
		BaseRuntime: types.PluginDeploymentBaseRuntime{
			BaseDeployDir:         info.BaseRuntime.BaseDeployDir,
			DeployDir:             info.BaseRuntime.DeployDir,
			GSEHomeDir:            info.BaseRuntime.GSEHomeDir,
			PluginHomeDir:         info.BaseRuntime.PluginHomeDir,
			DataIPC:               info.BaseRuntime.DataIPC,
			PluginIPC:             info.BaseRuntime.PluginIPC,
			HostIDPath:            info.BaseRuntime.HostIDPath,
			LogDir:                info.BaseRuntime.LogDir,
			DataDir:               info.BaseRuntime.DataDir,
			RunDir:                info.BaseRuntime.RunDir,
			ConfigDir:             info.BaseRuntime.ConfigDir,
			SubConfigDir:          info.BaseRuntime.SubConfigDir,
			PluginCommonConstants: info.BaseRuntime.PluginCommonConstants,
			GlobalCommonConstants: info.BaseRuntime.GlobalCommonConstants,
		},
		InstallOptions: types.PluginDeploymentInstallOptions{
			Version:                 info.InstallOptions.Version,
			IsOffline:               info.InstallOptions.IsOffline,
			EnableCompatibilityMode: info.InstallOptions.EnableCompatibilityMode,
		},
		TransferOptions: types.PluginDeploymentTransferOptions{
			DisableReleasePackage: info.TransferOptions.DisableReleasePackage,
			DisableInstaller:      info.TransferOptions.DisableInstaller,
		},
	}

	if info.InstallOptions.CustomSpec != nil {
		typesInfo.InstallOptions.CustomSpec = &types.PluginSpec{
			Resource: types.ProcessResource{
				CPULimitPercent: info.InstallOptions.CustomSpec.Resource.CPULimitPercent,
				MemLimitPercent: info.InstallOptions.CustomSpec.Resource.MemLimitPercent,
			},
			MonitorPolicy: types.ProcessMonitorPolicy{
				RestartType:    types.ProcessRestartType(info.InstallOptions.CustomSpec.MonitorPolicy.RestartType),
				StartCheckSecs: info.InstallOptions.CustomSpec.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  info.InstallOptions.CustomSpec.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  info.InstallOptions.CustomSpec.MonitorPolicy.OpTimeoutSecs,
			},
		}
	}

	return typesInfo, nil
}

// nolint: funlen
func convertPluginDeploymentInfoFromTypes(info *types.PluginDeploymentInfo) (*Info, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	data := &Info{
		ActionName:   info.BlockingActionName,
		ConfigSource: convertConfigSourceFromTypes(info.ConfigSource),
		Process: process{
			TenantID: info.Process.TenantID,
			HostID:   info.Process.HostID,
			BizID:    info.Process.BizID,
			Name:     info.Process.PluginName,
			Group:    info.Process.PluginGroup,
			PkgName:  info.Process.PluginPkgName,
			Platform: platform{
				OS:   string(info.Process.Platform.OS),
				Arch: string(info.Process.Platform.Arch),
			},
			Generation: int64(info.Process.Generation),
			BindIP:     info.Process.BindIP,
			BindPort:   info.Process.BindPort,
			Info: processInfo{
				Pid:        info.Process.Info.Pid,
				Version:    info.Process.Info.Version,
				AgentID:    info.Process.Info.AgentID,
				AutoStart:  info.Process.Info.AutoStart,
				Status:     string(info.Process.Info.Status),
				LastSyncAt: info.Process.Info.LastSyncAt,
			},
			Identity: processIdentity{
				Name:       info.Process.Identity.Name,
				SetupPath:  info.Process.Identity.SetupPath,
				PidPath:    info.Process.Identity.PidPath,
				ConfigPath: info.Process.Identity.ConfigPath,
				LogPath:    info.Process.Identity.LogPath,
				User:       info.Process.Identity.User,
			},
			Controller: processController{
				StartCmd:   info.Process.Controller.StartCmd,
				StopCmd:    info.Process.Controller.StopCmd,
				RestartCmd: info.Process.Controller.RestartCmd,
				ReloadCmd:  info.Process.Controller.ReloadCmd,
				DebugCmd:   info.Process.Controller.DebugCmd,
				KillCmd:    info.Process.Controller.KillCmd,
				VersionCmd: info.Process.Controller.VersionCmd,
				HealthCmd:  info.Process.Controller.HealthCmd,
			},
			Resource: processResource{
				CPULimitPercent: info.Process.Resource.CPULimitPercent,
				MemLimitPercent: info.Process.Resource.MemLimitPercent,
			},
			MonitorPolicy: processMonitorPolicy{
				RestartType:    string(info.Process.MonitorPolicy.RestartType),
				StartCheckSecs: info.Process.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  info.Process.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  info.Process.MonitorPolicy.OpTimeoutSecs,
			},
		},
		InstallerRuntime: installerRuntime{
			BaseWorkDir: info.InstallerRuntime.BaseWorkDir,
			WorkDir:     info.InstallerRuntime.WorkDir,
		},
		BaseRuntime: baseRuntime{
			BaseDeployDir:         info.BaseRuntime.BaseDeployDir,
			DeployDir:             info.BaseRuntime.DeployDir,
			GSEHomeDir:            info.BaseRuntime.GSEHomeDir,
			PluginHomeDir:         info.BaseRuntime.PluginHomeDir,
			DataIPC:               info.BaseRuntime.DataIPC,
			PluginIPC:             info.BaseRuntime.PluginIPC,
			HostIDPath:            info.BaseRuntime.HostIDPath,
			LogDir:                info.BaseRuntime.LogDir,
			DataDir:               info.BaseRuntime.DataDir,
			RunDir:                info.BaseRuntime.RunDir,
			ConfigDir:             info.BaseRuntime.ConfigDir,
			SubConfigDir:          info.BaseRuntime.SubConfigDir,
			PluginCommonConstants: info.BaseRuntime.PluginCommonConstants,
			GlobalCommonConstants: info.BaseRuntime.GlobalCommonConstants,
		},
		TransferOptions: transferOptions{
			DisableReleasePackage: info.TransferOptions.DisableReleasePackage,
			DisableInstaller:      info.TransferOptions.DisableInstaller,
		},
		InstallOptions: installOptions{
			Version:                 info.InstallOptions.Version,
			IsOffline:               info.InstallOptions.IsOffline,
			EnableCompatibilityMode: info.InstallOptions.EnableCompatibilityMode,
		},
	}

	if info.InstallOptions.CustomSpec != nil {
		data.InstallOptions.CustomSpec = &processSpec{
			Resource: processResource{
				CPULimitPercent: info.InstallOptions.CustomSpec.Resource.CPULimitPercent,
				MemLimitPercent: info.InstallOptions.CustomSpec.Resource.MemLimitPercent,
			},
			MonitorPolicy: processMonitorPolicy{
				RestartType:    string(info.InstallOptions.CustomSpec.MonitorPolicy.RestartType),
				StartCheckSecs: info.InstallOptions.CustomSpec.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  info.InstallOptions.CustomSpec.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  info.InstallOptions.CustomSpec.MonitorPolicy.OpTimeoutSecs,
			},
		}
	}

	return data, nil
}

func convertConfigSourceToTypes(configSource *target) types.Target {
	if configSource == nil {
		return types.Target{}
	}

	return types.Target{
		Host:                 convertTargetHostToTypes(configSource.Host),
		ServiceInstance:      convertServiceInstanceToTypes(configSource.ServiceInstance),
		MatchedTopoRelations: convertTargetMatchedTopoRelationsToTypes(configSource.MatchedTopoRelations),
	}
}

func convertConfigSourceFromTypes(configSource types.Target) *target {
	if configSource.Host.HostID <= 0 {
		return nil
	}

	return &target{
		Host:                 convertTargetHostFromTypes(configSource.Host),
		ServiceInstance:      convertServiceInstanceFromTypes(configSource.ServiceInstance),
		MatchedTopoRelations: convertTargetMatchedTopoRelationsFromTypes(configSource.MatchedTopoRelations),
	}
}

func convertTargetMatchedTopoRelationsToTypes(relations []targetMatchedTopoRelation) []types.TargetMatchedTopoRelation {
	return conv.SliceToSlice(relations, func(relation targetMatchedTopoRelation) types.TargetMatchedTopoRelation {
		return types.TargetMatchedTopoRelation{
			TopoObjID:  relation.TopoObjID,
			TopoInstID: relation.TopoInstID,
		}
	})
}

func convertTargetMatchedTopoRelationsFromTypes(relations []types.TargetMatchedTopoRelation) []targetMatchedTopoRelation {
	return conv.SliceToSlice(relations, func(relation types.TargetMatchedTopoRelation) targetMatchedTopoRelation {
		return targetMatchedTopoRelation{
			TopoObjID:  relation.TopoObjID,
			TopoInstID: relation.TopoInstID,
		}
	})
}

func convertServiceInstanceToTypes(serviceInstance serviceInstance) types.ServiceInstance {
	return types.ServiceInstance{
		ID:                serviceInstance.ID,
		Name:              serviceInstance.Name,
		Labels:            serviceInstance.Labels,
		Processes:         convertServiceInstanceProcessesToTypes(serviceInstance.Processes),
		BizID:             serviceInstance.BizID,
		HostID:            serviceInstance.HostID,
		ModuleID:          serviceInstance.ModuleID,
		ServiceTemplateID: serviceInstance.ServiceTemplateID,
		ServiceCategoryID: serviceInstance.ServiceCategoryID,
	}
}

func convertServiceInstanceFromTypes(instance types.ServiceInstance) serviceInstance {
	return serviceInstance{
		ID:                instance.ID,
		Name:              instance.Name,
		Labels:            instance.Labels,
		Processes:         convertServiceInstanceProcessesFromTypes(instance.Processes),
		BizID:             instance.BizID,
		HostID:            instance.HostID,
		ModuleID:          instance.ModuleID,
		ServiceTemplateID: instance.ServiceTemplateID,
		ServiceCategoryID: instance.ServiceCategoryID,
	}
}

func convertServiceInstanceProcessesToTypes(processes map[string]serviceInstanceProcess) map[string]types.ServiceInstanceProcess {
	typeProcesses := make(map[string]types.ServiceInstanceProcess, len(processes))
	for name, process := range processes {
		typeProcesses[name] = types.ServiceInstanceProcess{
			AutoStart:       process.AutoStart,
			BizID:           process.BizID,
			FuncName:        process.FuncName,
			ProcessID:       process.ProcessID,
			ProcessName:     process.ProcessName,
			StartParamRegex: process.StartParamRegex,
			SupplierAccount: process.SupplierAccount,
			CreateTime:      process.CreateTime,
			LastTime:        process.LastTime,
			Description:     process.Description,
			FaceStopCmd:     process.FaceStopCmd,
			PidFile:         process.PidFile,
			Priority:        process.Priority,
			ProcNum:         process.ProcNum,
			ReloadCmd:       process.ReloadCmd,
			RestartCmd:      process.RestartCmd,
			StartCmd:        process.StartCmd,
			StopCmd:         process.StopCmd,
			Timeout:         process.Timeout,
			User:            process.User,
			WorkPath:        process.WorkPath,
			CreateAt:        process.CreateAt,
			CreateBy:        process.CreateBy,
			UpdateAt:        process.UpdateAt,
			UpdateBy:        process.UpdateBy,
			BindInfo:        convertServiceInstanceProcessBindInfoToTypes(process.BindInfo),
		}
	}

	return typeProcesses
}

func convertServiceInstanceProcessesFromTypes(processes map[string]types.ServiceInstanceProcess) map[string]serviceInstanceProcess {
	daoProcesses := make(map[string]serviceInstanceProcess, len(processes))
	for name, process := range processes {
		daoProcesses[name] = serviceInstanceProcess{
			AutoStart:       process.AutoStart,
			BizID:           process.BizID,
			FuncName:        process.FuncName,
			ProcessID:       process.ProcessID,
			ProcessName:     process.ProcessName,
			StartParamRegex: process.StartParamRegex,
			SupplierAccount: process.SupplierAccount,
			CreateTime:      process.CreateTime,
			LastTime:        process.LastTime,
			Description:     process.Description,
			FaceStopCmd:     process.FaceStopCmd,
			PidFile:         process.PidFile,
			Priority:        process.Priority,
			ProcNum:         process.ProcNum,
			ReloadCmd:       process.ReloadCmd,
			RestartCmd:      process.RestartCmd,
			StartCmd:        process.StartCmd,
			StopCmd:         process.StopCmd,
			Timeout:         process.Timeout,
			User:            process.User,
			WorkPath:        process.WorkPath,
			CreateAt:        process.CreateAt,
			CreateBy:        process.CreateBy,
			UpdateAt:        process.UpdateAt,
			UpdateBy:        process.UpdateBy,
			BindInfo:        convertServiceInstanceProcessBindInfoFromTypes(process.BindInfo),
		}
	}

	return daoProcesses
}

func convertServiceInstanceProcessBindInfoToTypes(bindInfo []serviceInstanceProcessBindInfo) []types.ServiceInstanceProcessBindInfo {
	return conv.SliceToSlice(bindInfo, func(info serviceInstanceProcessBindInfo) types.ServiceInstanceProcessBindInfo {
		return types.ServiceInstanceProcessBindInfo{
			Enable:        info.Enable,
			IP:            info.IP,
			Port:          info.Port,
			Protocol:      info.Protocol,
			TemplateRowID: info.TemplateRowID,
		}
	})
}

func convertServiceInstanceProcessBindInfoFromTypes(bindInfo []types.ServiceInstanceProcessBindInfo) []serviceInstanceProcessBindInfo {
	return conv.SliceToSlice(bindInfo, func(info types.ServiceInstanceProcessBindInfo) serviceInstanceProcessBindInfo {
		return serviceInstanceProcessBindInfo{
			Enable:        info.Enable,
			IP:            info.IP,
			Port:          info.Port,
			Protocol:      info.Protocol,
			TemplateRowID: info.TemplateRowID,
		}
	})
}

func convertTargetHostToTypes(host targetHost) types.Host {
	return types.Host{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   convertTargetHostStaticToTypes(host.Static),
		Dynamic:  convertTargetHostDynamicToTypes(host.Dynamic),
	}
}

func convertTargetHostFromTypes(host types.Host) targetHost {
	return targetHost{
		HostID:   host.HostID,
		TenantID: host.TenantID,
		Static:   convertTargetHostStaticFromTypes(host.Static),
		Dynamic:  convertTargetHostDynamicFromTypes(host.Dynamic),
	}
}

func convertTargetHostStaticToTypes(static *targetHostStatic) *types.HostStatic {
	if static == nil {
		return &types.HostStatic{}
	}

	return &types.HostStatic{
		BizID:         static.BizID,
		Topo:          convertTargetHostToposToTypes(static.Topo),
		NetworkAreaID: static.NetworkAreaID,
		ZoneID:        static.ZoneID,
		CityID:        static.CityID,
		HostName:      static.HostName,
		DeptName:      static.DeptName,
		InnerIPList:   static.InnerIPList,
		InnerIPV6List: static.InnerIPV6List,
		OuterIPList:   static.OuterIPList,
		OuterIPV6List: static.OuterIPV6List,
		Operator:      static.Operator,
		Mac:           static.Mac,
		OSTypeCCID:    static.OSTypeCCID,
		OSType:        static.OSType,
		Arch:          static.Arch,
		Addressing:    types.Addressing(static.Addressing),
		CPUNum:        static.CPUNum,
		MemCap:        static.MemCap,
		SyncedAgentID: static.SyncedAgentID,
	}
}

func convertTargetHostStaticFromTypes(static *types.HostStatic) *targetHostStatic {
	if static == nil {
		return &targetHostStatic{}
	}

	return &targetHostStatic{
		BizID:         static.BizID,
		Topo:          convertTargetHostToposFromTypes(static.Topo),
		NetworkAreaID: static.NetworkAreaID,
		ZoneID:        static.ZoneID,
		CityID:        static.CityID,
		HostName:      static.HostName,
		DeptName:      static.DeptName,
		InnerIPList:   static.InnerIPList,
		InnerIPV6List: static.InnerIPV6List,
		OuterIPList:   static.OuterIPList,
		OuterIPV6List: static.OuterIPV6List,
		Operator:      static.Operator,
		Mac:           static.Mac,
		OSTypeCCID:    static.OSTypeCCID,
		OSType:        static.OSType,
		Arch:          static.Arch,
		Addressing:    string(static.Addressing),
		CPUNum:        static.CPUNum,
		MemCap:        static.MemCap,
		SyncedAgentID: static.SyncedAgentID,
	}
}

func convertTargetHostToposToTypes(topos []targetHostTopo) []*types.HostTopo {
	return conv.SliceToSlice(topos, func(topo targetHostTopo) *types.HostTopo {
		return &types.HostTopo{
			SetID:    topo.SetID,
			ModuleID: topo.ModuleID,
		}
	})
}

func convertTargetHostToposFromTypes(topos []*types.HostTopo) []targetHostTopo {
	return conv.SliceToSlice(topos, func(topo *types.HostTopo) targetHostTopo {
		return targetHostTopo{
			SetID:    topo.SetID,
			ModuleID: topo.ModuleID,
		}
	})
}

func convertTargetHostDynamicToTypes(dynamic *targetHostDynamic) *types.HostDynamic {
	if dynamic == nil {
		return &types.HostDynamic{}
	}

	return &types.HostDynamic{
		NodeRole:                 types.NodeRole(dynamic.NodeRole),
		NodeStatus:               types.NodeStatus(dynamic.NodeStatus),
		NodeVersion:              dynamic.NodeVersion,
		NodeGeneration:           types.Generation(dynamic.NodeGeneration),
		NodeCPUArch:              criteria.CPUArch(dynamic.NodeCPUArch),
		NodeOsType:               criteria.OSType(dynamic.NodeOsType),
		AgentID:                  dynamic.AgentID,
		NetworkUnitID:            dynamic.NetworkUnitID,
		ProxyAccessDisabled:      dynamic.ProxyAccessDisabled,
		ProxyTags:                convertProxyTagsToTypes(dynamic.ProxyTags),
		ProxyInstallOriginUnitID: dynamic.ProxyInstallOriginUnitID,
		ProxyClusterPort:         dynamic.ProxyClusterPort,
		ProxyDataPort:            dynamic.ProxyDataPort,
		ProxyFilePort:            dynamic.ProxyFilePort,
		LoginIP:                  dynamic.LoginIP,
		LoginUser:                dynamic.LoginUser,
		LoginMode:                types.LoginMode(dynamic.LoginMode),
		ExportIP:                 dynamic.ExportIP,
		ExportIPV6:               dynamic.ExportIPV6,
		AdvertiseIP:              dynamic.AdvertiseIP,
		AdvertiseIPV6:            dynamic.AdvertiseIPV6,
		RelayDownloadPort:        dynamic.RelayDownloadPort,
		RelayCallbackPort:        dynamic.RelayCallbackPort,
	}
}

func convertTargetHostDynamicFromTypes(dynamic *types.HostDynamic) *targetHostDynamic {
	if dynamic == nil {
		return &targetHostDynamic{}
	}

	return &targetHostDynamic{
		NodeRole:                 string(dynamic.NodeRole),
		NodeStatus:               string(dynamic.NodeStatus),
		NodeVersion:              dynamic.NodeVersion,
		NodeGeneration:           int64(dynamic.NodeGeneration),
		NodeCPUArch:              string(dynamic.NodeCPUArch),
		NodeOsType:               string(dynamic.NodeOsType),
		AgentID:                  dynamic.AgentID,
		NetworkUnitID:            dynamic.NetworkUnitID,
		ProxyAccessDisabled:      dynamic.ProxyAccessDisabled,
		ProxyTags:                convertProxyTagsFromTypes(dynamic.ProxyTags),
		ProxyInstallOriginUnitID: dynamic.ProxyInstallOriginUnitID,
		ProxyClusterPort:         dynamic.ProxyClusterPort,
		ProxyDataPort:            dynamic.ProxyDataPort,
		ProxyFilePort:            dynamic.ProxyFilePort,
		LoginIP:                  dynamic.LoginIP,
		LoginUser:                dynamic.LoginUser,
		LoginMode:                string(dynamic.LoginMode),
		ExportIP:                 dynamic.ExportIP,
		ExportIPV6:               dynamic.ExportIPV6,
		AdvertiseIP:              dynamic.AdvertiseIP,
		AdvertiseIPV6:            dynamic.AdvertiseIPV6,
		RelayDownloadPort:        dynamic.RelayDownloadPort,
		RelayCallbackPort:        dynamic.RelayCallbackPort,
	}
}

func convertProxyTagsToTypes(tags []string) []types.ProxyTag {
	return conv.SliceToSlice(tags, func(tag string) types.ProxyTag {
		return types.ProxyTag(tag)
	})
}

func convertProxyTagsFromTypes(tags []types.ProxyTag) []string {
	return conv.SliceToSlice(tags, func(tag types.ProxyTag) string {
		return string(tag)
	})
}

func convertPluginDeploymentPluginConfToTypes(conf *PluginConf) *types.PluginDeploymentPluginConf {
	if conf == nil {
		return nil
	}

	pluginConf := &types.PluginDeploymentPluginConf{
		Set:                  conf.Set,
		TemplateRenderer:     types.TemplateRendererType(conf.TemplateRenderer),
		ConfigFilesDetail:    convertPluginConfigDetailsToTypes(conf.ConfigFilesDetail...),
		SystemConfigContext:  conf.SystemConfigContext,
		CustomConfigContext:  conf.CustomConfigContext,
		RemoveConfigFileName: conf.RemoveConfigFileName,
		RemoveAllConfigs:     conf.RemoveAllConfigs,
	}

	return pluginConf
}

func convertPluginDeploymentPluginConfFromTypes(conf *types.PluginDeploymentPluginConf) *PluginConf {
	if conf == nil {
		return nil
	}

	pluginConf := &PluginConf{
		Set:                  conf.Set,
		TemplateRenderer:     string(conf.TemplateRenderer),
		ConfigFilesDetail:    convertPluginConfigDetailsFromTypes(conf.ConfigFilesDetail...),
		SystemConfigContext:  conf.SystemConfigContext,
		CustomConfigContext:  conf.CustomConfigContext,
		RemoveConfigFileName: conf.RemoveConfigFileName,
		RemoveAllConfigs:     conf.RemoveAllConfigs,
	}

	return pluginConf
}

func convertPluginConfigDetailsFromTypes(details ...*types.PluginConfigDetail) []configDetail {
	configDetails := make([]configDetail, 0, len(details))
	for _, detail := range details {
		if detail == nil {
			continue
		}

		configDetails = append(configDetails, configDetail{
			Name:         detail.Name,
			TemplateName: detail.TemplateName,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
			FilePath:     detail.FilePath,
		})
	}

	return configDetails
}

func convertPluginConfigDetailsToTypes(details ...configDetail) []*types.PluginConfigDetail {
	configDetails := make([]*types.PluginConfigDetail, 0, len(details))
	for _, detail := range details {
		configDetails = append(configDetails, &types.PluginConfigDetail{
			Name:         detail.Name,
			TemplateName: detail.TemplateName,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
			FilePath:     detail.FilePath,
		})
	}

	return configDetails
}
