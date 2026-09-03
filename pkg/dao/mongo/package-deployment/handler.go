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

package packagedeployment

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler package deployment handler interface.
type IHandler interface {
	// CreatePackageDeployment creates package deployment data.
	CreatePackageDeployment(nCtx contextx.IContext, deployment *types.PackageDeployment) error

	// ListPackageDeployment lists package deployment data by page and conditions.
	ListPackageDeployment(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageDeployment, int64, error)

	// GetPackageDeploymentInfo gets package deployment info by token.
	GetPackageDeploymentInfo(nCtx contextx.IContext, token string) (*types.PackageDeploymentInfo, error)

	// UpdatePackageDeploymentInfo updates package deployment info by token.
	UpdatePackageDeploymentInfo(nCtx contextx.IContext, token string, info *types.PackageDeploymentInfo) error
}

var _ IHandler = &Handler{}

// Handler is the handler of package deployment.
type Handler struct {
	dao *dao
}

// New creates a package deployment handler.
func New(client *mongo.Database) *Handler {
	handler := &Handler{dao: newDao(client)}
	if err := handler.dao.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to ensure package deployment indexes")
	}

	return handler
}

// CreatePackageDeployment creates a new package deployment.
func (h *Handler) CreatePackageDeployment(nCtx contextx.IContext, deployment *types.PackageDeployment) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if deployment == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertPackageDeploymentFromTypes(deployment)
	if err != nil {
		return fmt.Errorf("failed to convert package deployment: %w", err)
	}

	return h.dao.Create(nCtx, data)
}

// ListPackageDeployment lists package deployment data by page and conditions.
func (h *Handler) ListPackageDeployment(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.PackageDeployment, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.dao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count package deployments: %w", err)
	}

	data, err := h.dao.List(nCtx, filter, base.ParsePage(page))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list package deployments: %w", err)
	}

	deployments := make([]*types.PackageDeployment, len(data))
	for idx, deploy := range data {
		deployments[idx], err = convertPackageDeploymentToTypes(deploy)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert package deployment: %w", err)
		}
	}

	return deployments, num, nil
}

// GetPackageDeploymentInfo gets package deployment info by token.
func (h *Handler) GetPackageDeploymentInfo(nCtx contextx.IContext, token string) (*types.PackageDeploymentInfo, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if token == "" {
		return nil, base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)

	data, err := h.dao.Get(nCtx, filter, FieldKeyInfo)
	if err != nil {
		return nil, err
	}

	return convertInfoToTypes(data.Info)
}

// UpdatePackageDeploymentInfo updates package deployment info by token.
func (h *Handler) UpdatePackageDeploymentInfo(nCtx contextx.IContext, token string, info *types.PackageDeploymentInfo) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if token == "" {
		return base.ErrEmptyParamData()
	}

	if info == nil {
		return base.ErrEmptyParamData()
	}

	data, err := convertInfoFromTypes(info)
	if err != nil {
		return fmt.Errorf("failed to convert package deployment info: %w", err)
	}

	filter := base.AliveFilter()
	filter = WithToken(token)(filter)

	return h.dao.UpdateField(nCtx, filter, FieldKeyInfo, data)
}

func convertPackageDeploymentFromTypes(deployment *types.PackageDeployment) (*Data, error) {
	if deployment == nil {
		return nil, errors.New("package deployment is nil")
	}

	if deployment.Token == "" {
		return nil, base.ErrEmptyParamData()
	}

	info, err := convertInfoFromTypes(deployment.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert package deployment info: %w", err)
	}

	return &Data{
		Token: deployment.Token,
		Info:  info,
	}, nil
}

func convertPackageDeploymentToTypes(data *Data) (*types.PackageDeployment, error) {
	if data == nil {
		return nil, errors.New("package deployment is nil")
	}

	info, err := convertInfoToTypes(data.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to convert package deployment info: %w", err)
	}

	return &types.PackageDeployment{
		Token: data.Token,
		Info:  info,
	}, nil
}

func convertInfoFromTypes(info *types.PackageDeploymentInfo) (*Info, error) {
	if info == nil {
		return nil, errors.New("package deployment info is nil")
	}

	return &Info{
		Release: convertReleasesFromTypes(info.Release),
		Upload:  convertUploadInfoFromTypes(info.Upload),
		ImportPluginPkgOptions: importPluginPkgOptions{
			FileSourceType: string(info.ImportPluginPkgOptions.FileSourceType),
			FileSource:     info.ImportPluginPkgOptions.FileSource,
			FileName:       info.ImportPluginPkgOptions.FileName,
			MD5:            info.ImportPluginPkgOptions.MD5,
		},
		ExportPluginPkgOptions: exportPluginPkgOptions{
			PluginPkgName:    info.ExportPluginPkgOptions.PluginPkgName,
			PluginPkgVersion: info.ExportPluginPkgOptions.PluginPkgVersion,
		},
	}, nil
}

func convertInfoToTypes(info *Info) (*types.PackageDeploymentInfo, error) {
	if info == nil {
		return nil, errors.New("package deployment info is nil")
	}

	return &types.PackageDeploymentInfo{
		Release: convertReleasesToTypes(info.Release),
		Upload:  convertUploadInfoToTypes(info.Upload),
		ImportPluginPkgOptions: types.PackageImportPluginPkgOptions{
			FileSourceType: types.FileSourceType(info.ImportPluginPkgOptions.FileSourceType),
			FileSource:     info.ImportPluginPkgOptions.FileSource,
			FileName:       info.ImportPluginPkgOptions.FileName,
			MD5:            info.ImportPluginPkgOptions.MD5,
		},
		ExportPluginPkgOptions: types.PackageExportPluginPkgOptions{
			PluginPkgName:    info.ExportPluginPkgOptions.PluginPkgName,
			PluginPkgVersion: info.ExportPluginPkgOptions.PluginPkgVersion,
		},
	}, nil
}

func convertUploadInfoFromTypes(upload types.PackageDeploymentUploadInfo) uploadInfo {
	return uploadInfo{
		UploadID:  upload.UploadID,
		Name:      upload.Name,
		Version:   upload.Version,
		Platforms: convertPlatformsFromTypes(upload.Platforms),
	}
}

func convertUploadInfoToTypes(upload uploadInfo) types.PackageDeploymentUploadInfo {
	return types.PackageDeploymentUploadInfo{
		UploadID:  upload.UploadID,
		Name:      upload.Name,
		Version:   upload.Version,
		Platforms: convertPlatformsToTypes(upload.Platforms),
	}
}

func convertReleasesFromTypes(releases []types.Release) []release {
	if releases == nil {
		return nil
	}

	data := make([]release, len(releases))
	for idx, value := range releases {
		data[idx] = release{
			Name:         value.Name,
			Generation:   int64(value.Generation),
			Type:         string(value.Type),
			Version:      value.Version,
			CPUArch:      string(value.Platform.Arch),
			OSType:       string(value.Platform.OS),
			Labels:       value.Labels,
			FileName:     value.FileName,
			MD5:          value.MD5,
			Enabled:      value.Enabled,
			IsHidden:     value.IsHidden,
			AsDefault:    value.AsDefault,
			UpdatedAt:    value.UpdatedAt,
			Operator:     value.Operator,
			AdditionInfo: value.AdditionInfo,
		}
	}

	return data
}

func convertReleasesToTypes(releases []release) []types.Release {
	if releases == nil {
		return nil
	}

	data := make([]types.Release, len(releases))
	for idx, value := range releases {
		data[idx] = types.Release{
			Name:       value.Name,
			Generation: types.Generation(value.Generation),
			Type:       types.ReleaseType(value.Type),
			Version:    value.Version,
			Platform: platfmt.Platform{
				OS:   criteria.OSType(value.OSType),
				Arch: criteria.CPUArch(value.CPUArch),
			},
			Labels:       value.Labels,
			FileName:     value.FileName,
			MD5:          value.MD5,
			Enabled:      value.Enabled,
			IsHidden:     value.IsHidden,
			AsDefault:    value.AsDefault,
			UpdatedAt:    value.UpdatedAt,
			Operator:     value.Operator,
			AdditionInfo: value.AdditionInfo,
		}
	}

	return data
}

func convertPlatformsFromTypes(platforms []platfmt.Platform) []platform {
	if platforms == nil {
		return nil
	}

	data := make([]platform, len(platforms))
	for idx, value := range platforms {
		data[idx] = platform{
			OS:   string(value.OS),
			Arch: string(value.Arch),
		}
	}

	return data
}

func convertPlatformsToTypes(platforms []platform) []platfmt.Platform {
	if platforms == nil {
		return nil
	}

	data := make([]platfmt.Platform, len(platforms))
	for idx, value := range platforms {
		data[idx] = platfmt.Platform{
			OS:   criteria.OSType(value.OS),
			Arch: criteria.CPUArch(value.Arch),
		}
	}

	return data
}
