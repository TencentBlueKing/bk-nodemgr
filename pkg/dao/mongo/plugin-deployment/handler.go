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

	// GetInfo get a node deployment info.
	GetInfo(nCtx contextx.IContext, token string) (*types.PluginDeploymentInfo, error)

	// UpdateInfo update plugin deployment info.
	UpdateInfo(nCtx contextx.IContext, token string, info *types.PluginDeploymentInfo) error

	// GetMainConfig get main config.
	GetMainConfig(nCtx contextx.IContext, token string) ([]byte, error)

	// UpdateMainConfig set main config.
	UpdateMainConfig(nCtx contextx.IContext, token string, mainConfig []byte) error
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

// GetInfo get a node deployment info.
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

// GetMainConfig get main config.
func (h *Handler) GetMainConfig(nCtx contextx.IContext, token string) ([]byte, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(nCtx, filter, FieldKeyMainConfig)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return data.MainConfig, nil
}

// nolint: funlen
func convertPluginDeploymentInfoToTypes(info *Info) (*types.PluginDeploymentInfo, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	typesInfo := &types.PluginDeploymentInfo{
		BlockingActionName: info.ActionName,
		Plugin: types.Plugin{
			PluginID: info.Plugin.PluginID,
			TenantID: info.Plugin.TenantID,
			Static: types.PluginStatic{
				Identity: types.ProcessIdentity{
					Name:       info.Plugin.Static.Identity.Name,
					SetupPath:  info.Plugin.Static.Identity.SetupPath,
					PidPath:    info.Plugin.Static.Identity.PidPath,
					ConfigPath: info.Plugin.Static.Identity.ConfigPath,
					LogPath:    info.Plugin.Static.Identity.LogPath,
					User:       info.Plugin.Static.Identity.User,
				},
				Controller: types.ProcessController{
					StartCmd:   info.Plugin.Static.Controller.StartCmd,
					StopCmd:    info.Plugin.Static.Controller.StopCmd,
					RestartCmd: info.Plugin.Static.Controller.RestartCmd,
					ReloadCmd:  info.Plugin.Static.Controller.ReloadCmd,
					KillCmd:    info.Plugin.Static.Controller.KillCmd,
					VersionCmd: info.Plugin.Static.Controller.VersionCmd,
					HealthCmd:  info.Plugin.Static.Controller.HealthCmd,
				},
				Resource: types.ProcessResource{
					CPULimitPercent: info.Plugin.Static.Resource.CPULimitPercent,
					MemLimitPercent: info.Plugin.Static.Resource.MemLimitPercent,
				},
				MonitorPolicy: types.ProcessMonitorPolicy{
					AutoType:       types.ProcessAutoType(info.Plugin.Static.MonitorPolicy.AutoType),
					StartCheckSecs: info.Plugin.Static.MonitorPolicy.StartCheckSecs,
					StopCheckSecs:  info.Plugin.Static.MonitorPolicy.StopCheckSecs,
					OpTimeoutSecs:  info.Plugin.Static.MonitorPolicy.OpTimeoutSecs,
				},
			},
			Dynamic: types.PluginDynamic{
				Name:       info.Plugin.Dynamic.Name,
				Type:       types.PluginType(info.Plugin.Dynamic.Type),
				Generation: types.Generation(info.Plugin.Dynamic.Generation),
				Platform: platfmt.Platform{
					OS:   criteria.OSType(info.Plugin.Dynamic.Platform.OS),
					Arch: criteria.CPUArch(info.Plugin.Dynamic.Platform.Arch),
				},
				Version: info.Plugin.Dynamic.Version,
				HostID:  info.Plugin.Dynamic.HostID,
				Status:  types.ProcessStatus(info.Plugin.Dynamic.Status),
			},
		},
		InstallerWorkDir: info.InstallerWorkDir,
		InstallOptions:   types.PluginDeploymentInstallOptions{},
		TransferOptions: types.PluginDeploymentTransferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		TargetVersion: make([]types.TargetPluginVersion, 0),
	}

	for _, targetVersion := range info.TargetVersion {
		typesInfo.TargetVersion = append(typesInfo.TargetVersion, types.TargetPluginVersion{
			Platform: platfmt.Platform{
				OS:   criteria.OSType(targetVersion.Platform.OS),
				Arch: criteria.CPUArch(targetVersion.Platform.Arch),
			},
			Version: targetVersion.Version,
		})
	}

	return typesInfo, nil
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

func convertPluginDeploymentFromTypes(data *types.PluginDeployment) (*Data, error) {
	pluginDeployment := &Data{
		Token: data.Token,
		Info:  new(Info),
	}

	var err error

	pluginDeployment.Info, err = convertPluginDeploymentInfoFromTypes(data.Info)
	if err != nil {
		return nil, err
	}

	return pluginDeployment, nil
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

// nolint: funlen
func convertPluginDeploymentInfoFromTypes(info *types.PluginDeploymentInfo) (*Info, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	var data = &Info{
		ActionName:       info.BlockingActionName,
		InstallerWorkDir: info.InstallerWorkDir,
		Plugin: plugin{
			TenantID: info.Plugin.TenantID,
			PluginID: info.Plugin.PluginID,
			Static: pluginStatic{
				Info: ProcessInfo{
					Pid:         info.Plugin.Static.Info.Pid,
					Version:     info.Plugin.Static.Info.Version,
					AgentID:     info.Plugin.Static.Info.AgentID,
					Trusteeship: info.Plugin.Static.Info.Trusteeship,
					Status:      string(info.Plugin.Static.Info.Status),
				},
				Identity: processIdentity{
					Name:       info.Plugin.Static.Identity.Name,
					SetupPath:  info.Plugin.Static.Identity.SetupPath,
					PidPath:    info.Plugin.Static.Identity.PidPath,
					ConfigPath: info.Plugin.Static.Identity.ConfigPath,
					LogPath:    info.Plugin.Static.Identity.LogPath,
					User:       info.Plugin.Static.Identity.User,
				},
				Controller: processController{
					StartCmd:   info.Plugin.Static.Controller.StartCmd,
					StopCmd:    info.Plugin.Static.Controller.StopCmd,
					RestartCmd: info.Plugin.Static.Controller.RestartCmd,
					ReloadCmd:  info.Plugin.Static.Controller.ReloadCmd,
					KillCmd:    info.Plugin.Static.Controller.KillCmd,
					VersionCmd: info.Plugin.Static.Controller.VersionCmd,
					HealthCmd:  info.Plugin.Static.Controller.HealthCmd,
				},
				Resource: processResource{
					CPULimitPercent: info.Plugin.Static.Resource.CPULimitPercent,
					MemLimitPercent: info.Plugin.Static.Resource.MemLimitPercent,
				},
				MonitorPolicy: processMonitorPolicy{
					AutoType:       string(info.Plugin.Static.MonitorPolicy.AutoType),
					StartCheckSecs: info.Plugin.Static.MonitorPolicy.StartCheckSecs,
					StopCheckSecs:  info.Plugin.Static.MonitorPolicy.StopCheckSecs,
					OpTimeoutSecs:  info.Plugin.Static.MonitorPolicy.OpTimeoutSecs,
				},
			},
			Dynamic: pluginDynamic{
				Name:       info.Plugin.Dynamic.Name,
				Type:       string(info.Plugin.Dynamic.Type),
				Generation: int64(info.Plugin.Dynamic.Generation),
				Platform: Platform{
					OS:   string(info.Plugin.Dynamic.Platform.OS),
					Arch: string(info.Plugin.Dynamic.Platform.Arch),
				},
				Version: info.Plugin.Dynamic.Version,
				HostID:  info.Plugin.Dynamic.HostID,
				Status:  string(info.Plugin.Dynamic.Status),
			},
		},
		TransferOptions: transferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		InstallOptions: installOptions{},
		TargetVersion:  nil,
	}
	for _, item := range info.TargetVersion {
		data.TargetVersion = append(data.TargetVersion, targetVersion{
			Platform: Platform{
				OS:   string(item.Platform.OS),
				Arch: string(item.Platform.Arch),
			},
			Version: item.Version,
		})
	}

	return data, nil
}

// UpdateMainConfig set main config.
func (h *Handler) UpdateMainConfig(nCtx contextx.IContext, token string, mainConfig []byte) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	if err := h.dao.UpdateField(nCtx, filter, FieldKeyMainConfig, mainConfig); err != nil {
		return err
	}

	return nil
}
