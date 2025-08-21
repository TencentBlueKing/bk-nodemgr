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
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node deployment Handler interface.
type IHandler interface {
	Create(ctx context.Context, nodeDeployment *types.NodeDeployment) error
	GetInfo(ctx context.Context, token string) (*types.DeploymentInfo, error)
	GetNodeConf(ctx context.Context, token string) (*types.NodeConf, error)
	SetNodeConf(ctx context.Context, token string, nodeConf *types.NodeConf) error
	UpdateInfo(ctx context.Context, token string, info *types.DeploymentInfo) error
}

// Handler this is a Handler to operate node deployment table.
type Handler struct {
	dao    *dao
	logger logger.Logger
}

// New new a Handler.
func New(client *mongo.Database, logger logger.Logger) *Handler {
	h := &Handler{
		dao:    newDao(client, logger),
		logger: logger,
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure nodedeloyment indexes, err: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	return h
}

// GetInfo get a node deployment info.
func (h *Handler) GetInfo(ctx context.Context, token string) (*types.DeploymentInfo, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(ctx, filter, FieldKeyInfo)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertDeploymentInfoToTypes(data.Info)
}

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
				InnerIP:       info.InnerIP,
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
				ProxyTags: func() []types.ProxyTag {
					tags := make([]types.ProxyTag, len(info.ProxyTags))
					for i, tag := range info.ProxyTags {
						tags[i] = types.ProxyTag(tag)
					}

					return tags
				}(),
				ProxyClusterPort: info.ProxyClusterPort,
				ProxyDataPort:    info.ProxyDataPort,
				ProxyFilePort:    info.ProxyFilePort,
			},
		},
		LoginInfo: types.LoginInfo{
			IP:   info.LoginInfo.IP,
			Port: info.LoginInfo.Port,
			User: info.LoginInfo.User,
			Mode: types.LoginMode(info.LoginInfo.Mode),
		},
		InstallerWorkDir: info.InstallerWorkDir,
		InstallOptions: types.DeploymentInstallOptions{
			ReRegister: info.InstallOptions.ReRegister,
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
			HostID:         info.RelayHostID,
			AgentID:        info.RelayAgentID,
			NodeOsType:     criteria.OSType(info.NodeOsType),
			PackageDestDir: info.RelayPkgPath,
		},
	}, nil
}

// Create create a new node deployment.
func (h *Handler) Create(ctx context.Context, nodeDeployment *types.NodeDeployment) error {
	if ctx == nil {
		return base.ErrInvalidContext()
	}

	if nodeDeployment == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertNodeDeploymentFromTypes(nodeDeployment)
	if err != nil {
		return err
	}

	return h.dao.Create(ctx, data)
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

// GetNodeConf get a node deployment node conf.
func (h *Handler) GetNodeConf(ctx context.Context, token string) (*types.NodeConf, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(ctx, filter, FieldKeyNodeConf)
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

// SetNodeConf set a node deployment node conf.
func (h *Handler) SetNodeConf(ctx context.Context, token string, nodeConf *types.NodeConf) error {
	if ctx == nil {
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
	if err := h.dao.UpdateField(ctx, filter, FieldKeyNodeConf, data); err != nil {
		return err
	}

	return nil
}

// UpdateInfo update a node deployment info.
func (h *Handler) UpdateInfo(ctx context.Context, token string, info *types.DeploymentInfo) error {
	if ctx == nil {
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
	if err := h.dao.UpdateField(ctx, filter, FieldKeyInfo, data); err != nil {
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
		InnerIP:        info.Host.Static.InnerIP,
		Addressing:     string(info.Host.Static.Addressing),
		ProxyTags: func() []string {
			tags := make([]string, len(info.Host.Dynamic.ProxyTags))
			for i, tag := range info.Host.Dynamic.ProxyTags {
				tags[i] = string(tag)
			}

			return tags
		}(),
		ProxyClusterPort: info.Host.Dynamic.ProxyClusterPort,
		ProxyDataPort:    info.Host.Dynamic.ProxyDataPort,
		ProxyFilePort:    info.Host.Dynamic.ProxyFilePort,
		InstallerWorkDir: info.InstallerWorkDir,
		LoginInfo: LoginInfo{
			IP:   info.LoginInfo.IP,
			Port: info.LoginInfo.Port,
			User: info.LoginInfo.User,
			Mode: string(info.LoginInfo.Mode),
		},
		InstallOptions: InstallOptions{
			ReRegister: info.InstallOptions.ReRegister,
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
		RelayHostID:  info.RelayInfo.HostID,
		RelayAgentID: info.RelayInfo.AgentID,
		RelayOsType:  string(info.RelayInfo.NodeOsType),
		RelayPkgPath: info.RelayInfo.PackageDestDir,
	}

	return data, nil
}
