/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugindeployment

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
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
		Process: types.Process{
			TenantID:      info.Process.TenantID,
			HostID:        info.Process.HostID,
			PluginName:    info.Process.Name,
			PluginGroup:   info.Process.Group,
			PluginPkgName: info.Process.PkgName,
			Platform: platfmt.Platform{
				OS:   criteria.OSType(info.Process.Platform.OS),
				Arch: criteria.CPUArch(info.Process.Platform.Arch),
			},
			Generation: types.Generation(info.Process.Generation),
			Info: types.ProcessInfo{
				Pid:       info.Process.Info.Pid,
				Version:   info.Process.Info.Version,
				AgentID:   info.Process.Info.AgentID,
				AutoStart: info.Process.Info.AutoStart,
				Status:    types.ProcessStatus(info.Process.Info.Status),
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
				KillCmd:    info.Process.Controller.KillCmd,
				VersionCmd: info.Process.Controller.VersionCmd,
				HealthCmd:  info.Process.Controller.HealthCmd,
			},
			Resource: types.ProcessResource{
				CPULimitPercent: info.Process.Resource.CPULimitPercent,
				MemLimitPercent: info.Process.Resource.MemLimitPercent,
			},
			MonitorPolicy: types.ProcessMonitorPolicy{
				AutoType:       types.ProcessAutoType(info.Process.MonitorPolicy.AutoType),
				StartCheckSecs: info.Process.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  info.Process.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  info.Process.MonitorPolicy.OpTimeoutSecs,
			},
		},
		InstallerWorkDir: info.InstallerWorkDir,
		InstallOptions: types.PluginDeploymentInstallOptions{
			Version: info.InstallOptions.Version,
		},
		TransferOptions: types.PluginDeploymentTransferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
	}

	return typesInfo, nil
}

// nolint: funlen
func convertPluginDeploymentInfoFromTypes(info *types.PluginDeploymentInfo) (*Info, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	data := &Info{
		ActionName:       info.BlockingActionName,
		InstallerWorkDir: info.InstallerWorkDir,
		Process: process{
			TenantID: info.Process.TenantID,
			HostID:   info.Process.HostID,
			Name:     info.Process.PluginName,
			Group:    info.Process.PluginGroup,
			PkgName:  info.Process.PluginPkgName,
			Platform: platform{
				OS:   string(info.Process.Platform.OS),
				Arch: string(info.Process.Platform.Arch),
			},
			Generation: int64(info.Process.Generation),
			Info: processInfo{
				Pid:       info.Process.Info.Pid,
				Version:   info.Process.Info.Version,
				AgentID:   info.Process.Info.AgentID,
				AutoStart: info.Process.Info.AutoStart,
				Status:    string(info.Process.Info.Status),
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
				KillCmd:    info.Process.Controller.KillCmd,
				VersionCmd: info.Process.Controller.VersionCmd,
				HealthCmd:  info.Process.Controller.HealthCmd,
			},
			Resource: processResource{
				CPULimitPercent: info.Process.Resource.CPULimitPercent,
				MemLimitPercent: info.Process.Resource.MemLimitPercent,
			},
			MonitorPolicy: processMonitorPolicy{
				AutoType:       string(info.Process.MonitorPolicy.AutoType),
				StartCheckSecs: info.Process.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  info.Process.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  info.Process.MonitorPolicy.OpTimeoutSecs,
			},
		},
		TransferOptions: transferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		InstallOptions: installOptions{
			Version: info.InstallOptions.Version,
		},
	}

	return data, nil
}

func convertPluginDeploymentPluginConfToTypes(conf *PluginConf) *types.PluginDeploymentPluginConf {
	if conf == nil {
		return nil
	}

	pluginConf := &types.PluginDeploymentPluginConf{
		TemplateRenderer:    types.TemplateRendererType(conf.TemplateRenderer),
		ConfigFilesDetail:   convertPluginConfigDetailsToTypes(conf.ConfigFilesDetail...),
		SystemConfigContext: conf.SystemConfigContext,
		CustomConfigContext: conf.CustomConfigContext,
	}

	return pluginConf
}

func convertPluginDeploymentPluginConfFromTypes(conf *types.PluginDeploymentPluginConf) *PluginConf {
	if conf == nil {
		return nil
	}

	pluginConf := &PluginConf{
		TemplateRenderer:    string(conf.TemplateRenderer),
		ConfigFilesDetail:   convertPluginConfigDetailsFromTypes(conf.ConfigFilesDetail...),
		SystemConfigContext: conf.SystemConfigContext,
		CustomConfigContext: conf.CustomConfigContext,
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
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
		})
	}

	return configDetails
}

func convertPluginConfigDetailsToTypes(details ...configDetail) []*types.PluginConfigDetail {
	configDetails := make([]*types.PluginConfigDetail, 0, len(details))
	for _, detail := range details {
		configDetails = append(configDetails, &types.PluginConfigDetail{
			Name:         detail.Name,
			Content:      detail.Content,
			IsMainConfig: detail.IsMainConfig,
		})
	}

	return configDetails
}
