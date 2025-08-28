/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"errors"
	"fmt"
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/epluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IExternalPlugin defines the interface for external plugin.
type IExternalPlugin interface {
	// UploadOriginExternalPlugin uploads the origin external plugin.
	UploadOriginExternalPlugin(ctx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginExternalPluginPkgDetail, error)
}

// UploadOriginExternalPlugin uploads external plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginExternalPlugin(ctx contextx.IContext, externalPluginFile io.ReadCloser) (
	*types.OriginExternalPluginPkgDetail, error) {

	if externalPluginFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, externalPluginFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkOriginExternalPluginPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx,
			"failed to upload external plugin package. failed to check external plugin package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package. failed to get temp file. err: %v", err)

		return nil, err
	}

	pkgName, err := epluginpkg.FormatPkgName(
		detail.Name,
		types.ReleaseTypeOriginExternalPlugin,
		platform.EmptyPlatform(),
		detail.Version,
	)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package, failed to format package. err: %v", err)

		return nil, fmt.Errorf("failed to upload external plugin package: %w", err)
	}

	m.logger.InfoCtxf(ctx, "formatting external plugin package: %s", pkgName)

	pkgName = m.wrapOriginPackageName(pkgName)

	// upload to upstream.
	if err := m.upstreamOriginExternalPlugin.Store(ctx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package, failed to store to upstream: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginExternalPlugin.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseExternalPlugin(ctx, detail.Name, detail.Version, detail.Platforms...)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateExternalPluginUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginExternalPlugin,
		SavedName: pkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload external plugin package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx, "uploaded external plugin package to upstream. file-name(%s)", pkgName)

	return detail, nil
}

// ExternalPluginProject represents the project.yml file.
type ExternalPluginProject struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Description  string `yaml:"description"`
	Scenario     string `yaml:"scenario"`
	ConfigFile   string `yaml:"config_file"`
	ConfigFormat string `yaml:"config_format"`
	LaunchMode   string `yaml:"launch_mode"`
}

/**
 * plugin files contains:
 * bin/
 * etc/
 * project.yml
 * others can be ignored.
 */
func checkOriginExternalPluginPkg(file io.ReadCloser) (*types.OriginExternalPluginPkgDetail, error) {
	detail := new(types.OriginExternalPluginPkgDetail)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, "project.yaml"},
			callback: func(path []string, projectFile io.Reader) error {
				pluginProject := new(ExternalPluginProject)
				if err := yaml.NewDecoder(projectFile).Decode(pluginProject); err != nil {
					return fmt.Errorf("failed to decode project.yaml: %v", err)
				}

				detail.Name = pluginProject.Name
				detail.Version = pluginProject.Version
				detail.Description = pluginProject.Description
				detail.Scenario = pluginProject.Scenario
				detail.ConfigFile = pluginProject.ConfigFile
				detail.ConfigFormat = pluginProject.ConfigFormat
				detail.LaunchMode = pluginProject.LaunchMode

				return nil
			},
		},
		{
			filePath: []string{"external_plugins_linux_x86_64", tgzPathNameAny1, "project.yaml"},
			callback: func(_ []string, _ io.Reader) error {
				detail.Platforms = append(detail.Platforms, platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64))

				return nil
			},
		},
		{
			filePath: []string{"external_plugins_linux_aarch64", tgzPathNameAny1, "project.yaml"},
			callback: func(_ []string, _ io.Reader) error {
				detail.Platforms = append(detail.Platforms, platform.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64))

				return nil
			},
		},
		{
			filePath: []string{"external_plugins_windows_x86_64", tgzPathNameAny1, "project.yaml"},
			callback: func(_ []string, _ io.Reader) error {
				detail.Platforms = append(detail.Platforms, platform.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64))

				return nil
			},
		},
		{
			filePath: []string{"external_plugins_darwin_x86_64", tgzPathNameAny1, "project.yaml"},
			callback: func(_ []string, _ io.Reader) error {
				detail.Platforms = append(detail.Platforms, platform.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64))

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	return detail, nil
}
