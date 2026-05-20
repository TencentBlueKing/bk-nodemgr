/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodedeployment

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node deployment Handler interface.
type IHandler interface {
	CreateNodeDeployment(nCtx contextx.IContext, nodeDeployment *types.NodeDeployment) error
	ListNodeDeployment(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.NodeDeployment, int64, error)
	GetNodeDeploymentInfo(nCtx contextx.IContext, token string) (*types.DeploymentInfo, error)
	GetNodeDeploymentNodeConf(nCtx contextx.IContext, token string) (*types.NodeConf, error)
	SetNodeDeploymentNodeConf(nCtx contextx.IContext, token string, nodeConf *types.NodeConf) error
	UpdateNodeDeploymentInfo(nCtx contextx.IContext, token string, info *types.DeploymentInfo) error
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
		logger.G.Sys().WithErr(err).Warn("failed to ensure nodedeployment indexes")
	}

	return h
}

// GetNodeDeploymentInfo get a node deployment info.
func (h *Handler) GetNodeDeploymentInfo(nCtx contextx.IContext, token string) (*types.DeploymentInfo, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(nCtx, filter, FieldKeyInfo)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertDeploymentInfoToTypes(data.Info)
}

// ListNodeDeployment list node deployment by page and conditions.
func (h *Handler) ListNodeDeployment(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.NodeDeployment, int64, error) {

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

	data := make([]*types.NodeDeployment, len(deployments))
	for idx, deployment := range deployments {
		depTypes, err := convertDeploymentToTypes(deployment)
		if err != nil {
			return nil, 0, err
		}

		data[idx] = depTypes
	}

	return data, num, nil
}

func convertDeploymentToTypes(deployment *Data) (*types.NodeDeployment, error) {
	if deployment == nil {
		return nil, errors.New("deployment is nil")
	}

	Info, err := convertDeploymentInfoToTypes(deployment.Info)
	if err != nil {
		return nil, err
	}

	NodeConf, err := convertNodeConfToTypes(deployment.NodeConf)
	if err != nil {
		return nil, err
	}

	return &types.NodeDeployment{
		Token:    deployment.Token,
		Info:     Info,
		NodeConf: NodeConf,
	}, nil
}

// nolint: funlen
func convertDeploymentInfoToTypes(info *Info) (*types.DeploymentInfo, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	return &types.DeploymentInfo{
		BlockingActionName: info.ActionName,
		Host: types.Host{
			HostID:   info.HostID,
			TenantID: info.TenantID,
			Static: &types.HostStatic{
				BizID:         info.BizID,
				NetworkAreaID: info.NetworkAreaID,
				InnerIPList:   info.InnerIPList,
				OSType:        info.OSType,
				Addressing:    types.Addressing(info.Addressing),
			},
			Dynamic: &types.HostDynamic{
				NodeRole:       types.NodeRole(info.NodeRole),
				NodeStatus:     types.NodeStatus(info.NodeStatus),
				NodeVersion:    info.NodeVersion,
				NodeGeneration: types.Generation(info.NodeGeneration),
				NodeCPUArch:    criteria.CPUArch(info.NodeCPUArch),
				NodeOsType:     criteria.OSType(info.NodeOsType),
				AgentID:        info.AgentID,
				NetworkUnitID:  info.NetworkUnitID,
				ExportIP:       info.ExportIP,
				ExportIPV6:     info.ExportIPV6,
				AdvertiseIP:    info.AdvertiseIP,
				AdvertiseIPV6:  info.AdvertiseIPV6,
				ProxyTags: func() []types.ProxyTag {
					tags := make([]types.ProxyTag, len(info.ProxyTags))
					for i, tag := range info.ProxyTags {
						tags[i] = types.ProxyTag(tag)
					}

					return tags
				}(),
				ProxyClusterPort:         info.ProxyClusterPort,
				ProxyDataPort:            info.ProxyDataPort,
				ProxyFilePort:            info.ProxyFilePort,
				RelayDownloadPort:        info.RelayDownloadPort,
				RelayCallbackPort:        info.RelayCallbackPort,
				ProxyInstallOriginUnitID: info.ProxyInstallOriginUnitID,
				LoginIP:                  info.LoginIP,
				LoginPort:                info.LoginPort,
				LoginUser:                info.LoginUser,
				LoginMode:                types.LoginMode(info.LoginMode),
				LoginCreditID:            info.LoginCreditID,
			},
		},
		InstallerRuntime: types.DeploymentInstallerRuntime{
			BaseWorkDir: info.InstallerRuntime.BaseWorkDir,
			WorkDir:     info.InstallerRuntime.WorkDir,
		},
		BaseRuntime: types.DeploymentBaseRuntime{
			BaseDeployDir:  info.BaseRuntime.BaseDeployDir,
			DeployDir:      info.BaseRuntime.DeployDir,
			HomeDir:        info.BaseRuntime.HomeDir,
			DataIPC:        info.BaseRuntime.DataIPC,
			PluginIPC:      info.BaseRuntime.PluginIPC,
			ExtraConfigDir: info.BaseRuntime.ExtraConfigDir,
			LogDir:         info.BaseRuntime.LogDir,
		},
		InstallOptions: types.DeploymentInstallOptions{
			ReRegister:               info.InstallOptions.ReRegister,
			RenewGSETask:             info.InstallOptions.RenewGSETask,
			RenewGSEProc:             info.InstallOptions.RenewGSEProc,
			InstallPreOrderedPlugins: info.InstallOptions.InstallPreOrderedPlugins,
			DirectInstall:            info.InstallOptions.DirectInstall,
			IsManual:                 info.InstallOptions.IsManual,
			IsOffline:                info.InstallOptions.IsOffline,
			EnableCompatibilityMode:  info.InstallOptions.EnableCompatibilityMode,
		},
		ReconfigOptions: types.DeploymentReconfigOptions{
			DirectLink:           info.ReconfigOptions.DirectLink,
			AllowReleaseFallback: info.ReconfigOptions.AllowReleaseFallback,
		},
		RestartOptions: types.DeploymentRestartOptions{
			ForceRestart:           info.RestartOptions.ForceRestart,
			GracefulRestartTimeout: info.RestartOptions.GracefulRestartTimeout,
		},
		TransferOptions: types.DeploymentTransferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		CurrentVersionSupports: types.DeploymentVersionSupports{
			OperateAgentRestart: info.CurrentVersionSupports.OperateAgentRestart,
		},
		TargetVersion: func() []types.TargetVersion {
			versions := make([]types.TargetVersion, len(info.TargetVersion))
			for i, version := range info.TargetVersion {
				versions[i] = types.TargetVersion{
					OsType:  criteria.OSType(version.OsType),
					CPUArch: criteria.CPUArch(version.CPUArch),
					Version: version.Version,
				}
			}

			return versions
		}(),
		RelayInfo: types.RelayInfo{
			HostID:          info.RelayInfo.HostID,
			AgentID:         info.RelayInfo.AgentID,
			AdvertiseIP:     info.RelayInfo.InnerIP,
			AdvertiseIPV6:   info.RelayInfo.InnerIPV6,
			PackageDestDir:  info.RelayInfo.PackageDestDir,
			DownloadSvcPort: info.RelayInfo.DownloadSvcPort,
			CallbackSvcPort: info.RelayInfo.CallbackSvcPort,
		},
	}, nil
}

// CreateNodeDeployment create a new node deployment.
func (h *Handler) CreateNodeDeployment(nCtx contextx.IContext, nodeDeployment *types.NodeDeployment) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if nodeDeployment == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertNodeDeploymentFromTypes(nodeDeployment)
	if err != nil {
		return err
	}

	return h.dao.Create(nCtx, data)
}

func convertNodeDeploymentFromTypes(data *types.NodeDeployment) (*Data, error) {
	nodeDeployment := &Data{
		Token:    data.Token,
		Info:     new(Info),
		NodeConf: new(NodeConf),
	}

	var err error

	nodeDeployment.Info, err = convertDeploymentInfoFromTypes(data.Info)
	if err != nil {
		return nil, err
	}

	nodeDeployment.NodeConf, err = convertNodeConfFromTypes(data.NodeConf)
	if err != nil {
		return nil, err
	}

	return nodeDeployment, nil
}

func convertNodeConfFromTypes(nodeConf *types.NodeConf) (*NodeConf, error) {
	if nodeConf == nil {
		return nil, errors.New("node conf is nil")
	}

	return &NodeConf{
		ConfigTemplate: nodeConf.ConfigTemplate,
		PreSetting:     nodeConf.PreSetting,
		CustomSetting:  nodeConf.CustomSetting,
	}, nil
}

// GetNodeDeploymentNodeConf get a node deployment node conf.
func (h *Handler) GetNodeDeploymentNodeConf(nCtx contextx.IContext, token string) (*types.NodeConf, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(nCtx, filter, FieldKeyNodeConf)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertNodeConfToTypes(data.NodeConf)
}

func convertNodeConfToTypes(nodeConf *NodeConf) (*types.NodeConf, error) {
	if nodeConf == nil {
		return nil, errors.New("node conf is nil")
	}

	return &types.NodeConf{
		ConfigTemplate: nodeConf.ConfigTemplate,
		PreSetting:     nodeConf.PreSetting,
		CustomSetting:  nodeConf.CustomSetting,
	}, nil
}

// SetNodeDeploymentNodeConf set a node deployment node conf.
func (h *Handler) SetNodeDeploymentNodeConf(nCtx contextx.IContext, token string, nodeConf *types.NodeConf) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return base.ErrInvalidID()
	}

	if nodeConf == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertNodeConfFromTypes(nodeConf)
	if err != nil {
		return err
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	if err := h.dao.UpdateField(nCtx, filter, FieldKeyNodeConf, data); err != nil {
		return err
	}

	return nil
}

// UpdateNodeDeploymentInfo update a node deployment info.
func (h *Handler) UpdateNodeDeploymentInfo(nCtx contextx.IContext, token string, info *types.DeploymentInfo) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return base.ErrInvalidID()
	}

	if info == nil {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := convertDeploymentInfoFromTypes(info)
	if err != nil {
		return err
	}
	if err := h.dao.UpdateField(nCtx, filter, FieldKeyInfo, data); err != nil {
		return err
	}

	return nil
}

func convertDeploymentInfoFromTypes(info *types.DeploymentInfo) (*Info, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}
	data := &Info{
		ActionName:     info.BlockingActionName,
		HostID:         info.Host.HostID,
		OSType:         info.Host.Static.OSType,
		TenantID:       info.Host.TenantID,
		NodeRole:       string(info.Host.Dynamic.NodeRole),
		NodeStatus:     string(info.Host.Dynamic.NodeStatus),
		NodeVersion:    info.Host.Dynamic.NodeVersion,
		NodeGeneration: int64(info.Host.Dynamic.NodeGeneration),
		NodeCPUArch:    string(info.Host.Dynamic.NodeCPUArch),
		NodeOsType:     string(info.Host.Dynamic.NodeOsType),
		AgentID:        info.Host.Dynamic.AgentID,
		NetworkUnitID:  info.Host.Dynamic.NetworkUnitID,
		NetworkAreaID:  info.Host.Static.NetworkAreaID,
		BizID:          info.Host.Static.BizID,
		InnerIPList:    info.Host.Static.InnerIPList,
		Addressing:     string(info.Host.Static.Addressing),
		ExportIP:       info.Host.Dynamic.ExportIP,
		ExportIPV6:     info.Host.Dynamic.ExportIPV6,
		AdvertiseIP:    info.Host.Dynamic.AdvertiseIP,
		AdvertiseIPV6:  info.Host.Dynamic.AdvertiseIPV6,
		ProxyTags: func() []string {
			tags := make([]string, len(info.Host.Dynamic.ProxyTags))
			for i, tag := range info.Host.Dynamic.ProxyTags {
				tags[i] = string(tag)
			}

			return tags
		}(),
		LoginIP:                  info.Host.Dynamic.LoginIP,
		LoginPort:                info.Host.Dynamic.LoginPort,
		LoginUser:                info.Host.Dynamic.LoginUser,
		LoginMode:                string(info.Host.Dynamic.LoginMode),
		LoginCreditID:            info.Host.Dynamic.LoginCreditID,
		ProxyClusterPort:         info.Host.Dynamic.ProxyClusterPort,
		ProxyDataPort:            info.Host.Dynamic.ProxyDataPort,
		ProxyFilePort:            info.Host.Dynamic.ProxyFilePort,
		RelayDownloadPort:        info.Host.Dynamic.RelayDownloadPort,
		RelayCallbackPort:        info.Host.Dynamic.RelayCallbackPort,
		ProxyInstallOriginUnitID: info.Host.Dynamic.ProxyInstallOriginUnitID,
		InstallerRuntime: InstallerRuntime{
			BaseWorkDir: info.InstallerRuntime.BaseWorkDir,
			WorkDir:     info.InstallerRuntime.WorkDir,
		},
		BaseRuntime: BaseRuntime{
			BaseDeployDir:  info.BaseRuntime.BaseDeployDir,
			DeployDir:      info.BaseRuntime.DeployDir,
			HomeDir:        info.BaseRuntime.HomeDir,
			DataIPC:        info.BaseRuntime.DataIPC,
			PluginIPC:      info.BaseRuntime.PluginIPC,
			ExtraConfigDir: info.BaseRuntime.ExtraConfigDir,
			LogDir:         info.BaseRuntime.LogDir,
		},
		InstallOptions: InstallOptions{
			ReRegister:               info.InstallOptions.ReRegister,
			RenewGSETask:             info.InstallOptions.RenewGSETask,
			RenewGSEProc:             info.InstallOptions.RenewGSEProc,
			InstallPreOrderedPlugins: info.InstallOptions.InstallPreOrderedPlugins,
			DirectInstall:            info.InstallOptions.DirectInstall,
			IsManual:                 info.InstallOptions.IsManual,
			IsOffline:                info.InstallOptions.IsOffline,
			EnableCompatibilityMode:  info.InstallOptions.EnableCompatibilityMode,
		},
		ReconfigOptions: ReconfigOptions{
			DirectLink:           info.ReconfigOptions.DirectLink,
			AllowReleaseFallback: info.ReconfigOptions.AllowReleaseFallback,
		},
		RestartOptions: RestartOptions{
			ForceRestart:           info.RestartOptions.ForceRestart,
			GracefulRestartTimeout: info.RestartOptions.GracefulRestartTimeout,
		},
		TransferOptions: TransferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		CurrentVersionSupports: VersionSupports{
			OperateAgentRestart: info.CurrentVersionSupports.OperateAgentRestart,
		},
		TargetVersion: func() []TargetVersion {
			dbTargetVersion := make([]TargetVersion, len(info.TargetVersion))
			for i, version := range info.TargetVersion {
				dbTargetVersion[i] = TargetVersion{
					OsType:  string(version.OsType),
					CPUArch: string(version.CPUArch),
					Version: version.Version,
				}
			}

			return dbTargetVersion
		}(),
		RelayInfo: RelayInfo{
			HostID:          info.RelayInfo.HostID,
			AgentID:         info.RelayInfo.AgentID,
			PackageDestDir:  info.RelayInfo.PackageDestDir,
			InnerIP:         info.RelayInfo.AdvertiseIP,
			DownloadSvcPort: info.RelayInfo.DownloadSvcPort,
			CallbackSvcPort: info.RelayInfo.CallbackSvcPort,
		},
	}

	return data, nil
}
