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
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler node deployment Handler interface.
type IHandler interface {
	// Create create a node deployment.
	Create(ctx context.Context, pluginDeployment *types.PluginDeployment) error

	// GetInfo get a node deployment info.
	GetInfo(ctx context.Context, token string) (*types.PluginDeploymentInfo, error)

	// UpdateInfo update plugin deployment info.
	UpdateInfo(ctx context.Context, token string, info *types.PluginDeploymentInfo) error

	// GetMainConfig get main config.
	GetMainConfig(ctx context.Context, token string) ([]byte, error)
}

// Handler this is a Handler to operate node deployment table.
type Handler struct {
	dao    *dao
	logger logger.ILogger
}

// New new a Handler.
func New(client *mongo.Database, logger logger.ILogger) *Handler {
	h := &Handler{
		dao:    newDao(client, logger),
		logger: logger,
	}

	if err := h.dao.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure nodedeloyment indexes: %v",
			errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	return h
}

// GetInfo get a node deployment info.
func (h *Handler) GetInfo(ctx context.Context, token string) (*types.PluginDeploymentInfo, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(ctx, filter, FieldKeyInfo)
	if err != nil {
		return nil, base.ErrRecordNoFound()
	}

	return convertPluginDeploymentInfoToTypes(data.Info)
}

// GetMainConfig get main config.
func (h *Handler) GetMainConfig(ctx context.Context, token string) ([]byte, error) {
	if ctx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, ErrInvalidToken()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)
	data, err := h.dao.Get(ctx, filter, FieldKeyMainConfig)
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
			Name:       info.Plugin.Name,
			HostID:     info.Plugin.HostID,
			Type:       types.PluginType(info.Plugin.Type),
			Generation: types.Generation(info.Plugin.Generation),
			Platform: platform.Platform{
				OS:   criteria.OSType(info.Plugin.Platform.OS),
				Arch: criteria.CPUArch(info.Plugin.Platform.Arch),
			},
			Version: info.Plugin.Version,
		},
		InstallerWorkDir: info.InstallerWorkDir,
		InstallOptions:   types.PluginDeploymentInstallOptions{},
		UpgradeOptions:   types.PluginDeploymentUpgradeOptions{},
		RestartOptions:   types.PluginDeploymentRestartOptions{},
		TransferOptions: types.PluginDeploymentTransferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		TargetVersion: nil,
	}

	for _, targetVersion := range info.TargetVersion {
		typesInfo.TargetVersion = append(typesInfo.TargetVersion, types.TargetPluginVersion{
			Platform: platform.Platform{
				OS:   criteria.OSType(targetVersion.Platform.OS),
				Arch: criteria.CPUArch(targetVersion.Platform.Arch),
			},
			Version: targetVersion.Version,
		})
	}

	return typesInfo, nil
}

// Create create a new node deployment.
func (h *Handler) Create(ctx context.Context, pluginDeployment *types.PluginDeployment) error {
	if ctx == nil {
		return base.ErrInvalidContext()
	}

	if pluginDeployment == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertPluginDeploymentFromTypes(pluginDeployment)
	if err != nil {
		return err
	}

	return h.dao.Create(ctx, data)
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
func (h *Handler) UpdateInfo(ctx context.Context, token string, info *types.PluginDeploymentInfo) error {
	if ctx == nil {
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
	if err := h.dao.UpdateField(ctx, filter, FieldKeyInfo, data); err != nil {
		return err
	}

	return nil
}

func convertPluginDeploymentInfoFromTypes(info *types.PluginDeploymentInfo) (*Info, error) {
	if info == nil {
		return nil, errors.New("info is nil")
	}

	data := &Info{
		ActionName:       info.BlockingActionName,
		InstallerWorkDir: info.InstallerWorkDir,
		Plugin: Plugin{
			Name:       info.Plugin.Name,
			HostID:     info.Plugin.HostID,
			Type:       string(info.Plugin.Type),
			Generation: int64(info.Plugin.Generation),
			Platform: Platform{
				OS:   string(info.Plugin.Platform.OS),
				Arch: string(info.Plugin.Platform.Arch),
			},
			Version: info.Plugin.Version,
		},
		TransferOptions: TransferOptions{
			SelectDownloads:      info.TransferOptions.SelectDownloads,
			EnableReleasePackage: info.TransferOptions.EnableReleasePackage,
			EnableInstaller:      info.TransferOptions.EnableInstaller,
		},
		TargetVersion: nil,
	}

	for _, item := range info.TargetVersion {
		data.TargetVersion = append(data.TargetVersion, TargetVersion{
			Platform: Platform{
				OS:   string(item.Platform.OS),
				Arch: string(item.Platform.Arch),
			},
			Version: item.Version,
		})
	}

	return data, nil
}
