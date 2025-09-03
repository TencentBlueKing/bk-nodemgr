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
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/opluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IOfficialPlugin defines the interface of official plugin.
type IOfficialPlugin interface {
	// UploadOriginOfficialPlugin uploads the origin official plugin.
	UploadOriginOfficialPlugin(ctx contextx.IContext, pluginFile io.ReadCloser) (
		*types.OriginOfficialPluginPkgDetail, error)

	// PublishReleaseOfficialPlugin generates release official plugin by upload-id.
	PublishReleaseOfficialPlugin(ctx contextx.IUserContext, uploadID string) error
}

// UploadOriginOfficialPlugin uploads origin official plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginOfficialPlugin(ctx contextx.IContext, officialPluginFile io.ReadCloser) (
	*types.OriginOfficialPluginPkgDetail, error) {

	if officialPluginFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, officialPluginFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkOriginOfficialPluginPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx,
			"failed to upload origin official plugin package. failed to check origin official plugin package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package. failed to get temp file. err: %v", err)

		return nil, err
	}

	gen := types.Generation2
	originalPkgName, err := opluginpkg.FormatPkgName(detail.Name, types.ReleaseTypeOriginOfficialPlugin, gen, platform.EmptyPlatform(), detail.Version)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package, failed to format package. err: %v", err)

		return nil, fmt.Errorf("failed to upload origin official plugin package: %w", err)
	}

	m.logger.InfoCtxf(ctx, "formatting origin official plugin package: %s", originalPkgName)

	originalPkgName = m.wrapOriginPackageName(originalPkgName)

	// upload to upstream.
	if err := m.upstreamOriginOfficialPlugin.Store(ctx, fileiface.FileInfo{Name: originalPkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package, failed to store to upstream: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginOfficialPlugin.GetFile(ctx, originalPkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseOfficialPluginGen2(ctx, detail.Name, detail.Version, detail.Platforms...)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateOfficialPluginUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginOfficialPlugin,
		SavedName: originalPkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin official plugin package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx, "uploaded origin official plugin package to upstream. file-name(%s)", originalPkgName)

	return detail, nil
}

// OfficialPluginProject represents the project.yml file.
type OfficialPluginProject struct {
	Name            string                         `yaml:"name"`
	Version         string                         `yaml:"version"`
	Description     string                         `yaml:"description"`
	Scenario        string                         `yaml:"scenario"`
	ConfigFile      string                         `yaml:"config_file"`
	ConfigFormat    string                         `yaml:"config_format"`
	LaunchMode      string                         `yaml:"launch_mode"`
	ConfigTemplates []OfficialPluginConfigTemplate `yaml:"config_templates"`
}

// OfficialPluginConfigTemplate represents the project.yml file's config_templates field.
type OfficialPluginConfigTemplate struct {
	PluginVersion string    `yaml:"plugin_version"`
	Name          string    `yaml:"name"`
	Version       string    `yaml:"version"`
	FilePath      string    `yaml:"file_path"`
	Format        string    `yaml:"format"`
	IsMainConfig  string    `yaml:"is_main_config"`
	SourcePath    string    `yaml:"source_path"`
	Variables     *Property `yaml:"variables"`
}

// Property represents the project.yml file's variables field.
type Property struct {
	Title      string               `yaml:"title,omitempty"`
	Type       string               `yaml:"type,omitempty"`
	Required   bool                 `yaml:"required,omitempty"`
	Default    any                  `yaml:"default,omitempty"`
	Items      *Property            `yaml:"items,omitempty"`
	Properties map[string]*Property `yaml:"properties,omitempty"`
}

/**
 * plugin files contains:
 * bin/
 * etc/
 * project.yml
 * others can be ignored.
 */
func checkOriginOfficialPluginPkg(file io.ReadCloser) (*types.OriginOfficialPluginPkgDetail, error) {
	detail := new(types.OriginOfficialPluginPkgDetail)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, "project.yaml"},
			callback: func(path []string, projectFile io.Reader) error {
				detail.Platforms = append(detail.Platforms, convOfficialPluginDirNameToPlat(path[0]))

				pluginProject := new(OfficialPluginProject)
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

				for _, configTemplate := range pluginProject.ConfigTemplates {
					detail.ConfigTemplates = append(detail.ConfigTemplates, types.PluginPkgConfigTemplate{
						PluginVersion: configTemplate.PluginVersion,
						Name:          configTemplate.Name,
						Version:       configTemplate.Version,
						FilePath:      configTemplate.FilePath,
						Format:        configTemplate.Format,
						IsMainConfig:  configTemplate.IsMainConfig,
						SourcePath:    configTemplate.SourcePath,
						Variables:     convPropertyToTypes(configTemplate.Variables),
					})
				}

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	return detail, nil
}

func convPropertyToTypes(property *Property) *types.PluginPkgConfigTemplateProperty {
	if property == nil {
		return nil
	}

	pluginPkgConfigTemplateProperty := &types.PluginPkgConfigTemplateProperty{
		Title:    property.Title,
		Type:     property.Type,
		Required: property.Required,
		Default:  property.Default,
	}

	if property.Items != nil {
		pluginPkgConfigTemplateProperty.Items = convPropertyToTypes(property.Items)
	}

	if property.Properties != nil {
		pluginPkgConfigTemplateProperty.Properties = make(map[string]*types.PluginPkgConfigTemplateProperty)
		for k, v := range property.Properties {
			pluginPkgConfigTemplateProperty.Properties[k] = convPropertyToTypes(v)
		}
	}

	return pluginPkgConfigTemplateProperty
}

// PublishReleaseOfficialPlugin generates release official plugin by upload-id.
func (m *Manager) PublishReleaseOfficialPlugin(ctx contextx.IUserContext, uploadID string) error {
	up, err := m.storageUpload.GetOfficialPluginUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginOfficialPlugin {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginOfficialPlugin.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(ctx, originContent)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release official plugin package. failed to save temp file. err: %v", err)

		return err
	}

	checkingFile, err := m.getTempFile(ctx, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release official plugin package. failed to get temp file. err: %v", err)

		return err
	}

	detail, err := checkOriginOfficialPluginPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release official plugin package. failed to check origin official plugin package. err: %v", err)

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateOfficialPluginPkg(ctx, detail, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to generate official plugin pkg. err: %v", err)

		return err
	}

	gp := gopool.NewPool()
	gen := types.Generation2

	releasesMap := make(map[string]*types.ReleaseOfficialPlugin)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := opluginpkg.FormatPkgName(detail.Name, types.ReleaseTypeOfficialPlugin, gen, pkg.platform, detail.Version)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to upload release official plugin package, failed to format package. err: %v", err)

				return err
			}

			generatedFile, err := m.getTempFile(ctx, pkg.tempFileName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to get temp file. err: %v", err)

				return err
			}

			if err = m.upstreamReleaseOfficialPlugin.Store(ctx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to upload to upstream. err: %v", err)

				return err
			}

			file, err := m.upstreamReleaseOfficialPlugin.GetFile(ctx, pkgName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to get temp file. err: %v", err)

				return err
			}

			releasesMap[pkg.platform.String()] = &types.ReleaseOfficialPlugin{
				Release: types.Release{
					Generation:   gen,
					Type:         types.ReleaseTypeOfficialPlugin,
					Version:      detail.Version,
					Platform:     pkg.platform,
					Labels:       []string{},
					FileName:     file.Info().Name,
					MD5:          file.Info().MD5,
					Enabled:      true,
					AsDefault:    false,
					UpdatedAt:    time.Now(),
					Operator:     ctx.BKUsername(),
					AdditionInfo: nil,
				},
				ReleaseAdditionInfoOfficialPlugin: types.ReleaseAdditionInfoOfficialPlugin{
					ConfigTemplates: detail.ConfigTemplates,
				},
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to upload to upstream. err: %v", err)

		return err
	}

	// upsert release official plugin.
	if err = m.storageRelease.UpsertManyReleaseOfficialPluginGen2(ctx, conv.MapValueToSlice(releasesMap)); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release official plugin, failed to upsert release official plugin: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release official plugins. platforms(%v)", detail.Platforms)

	return nil
}

type releaseOfficialPluginPkg struct {
	platform     platform.Platform
	tempFileName string
}

// generateOfficialPluginPkg generates official plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateOfficialPluginPkg(ctx context.Context,
	originDetail *types.OriginOfficialPluginPkgDetail,
	originLocalFileName string) ([]*releaseOfficialPluginPkg, error) {

	// local origin official plugin.
	localOrigin, err := m.tempFileGroup.GetFile(ctx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	pluginName := originDetail.Name

	gp := gopool.NewPool()

	result := make(map[string]*releaseOfficialPluginPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(ctx)
			if err != nil {
				return fmt.Errorf("failed to create official plugin pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(ctx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			origiOfficialPluginFile, err := localOrigin.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin official plugin file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{
						targetFilePath: []string{"bin"},
						targetFileMode: tgzModeDir,
					},
					{
						targetFilePath: []string{"etc"},
						targetFileMode: tgzModeDir,
					},
				},
				[]*tgzWriteRuleStream{
					// get things from origin official plugin.
					{
						sourceFile: origiOfficialPluginFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{convPlatToOfficialPluginDirName(plat), pluginName, "bin", tgzPathNameAny2},
								targetFilePath: []string{
									"bin", tgzPathNameAny2,
								},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{convPlatToOfficialPluginDirName(plat), pluginName, "etc", tgzPathNameAny2},
								targetFilePath: []string{
									"etc", tgzPathNameAny2,
								},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{convPlatToOfficialPluginDirName(plat), pluginName, "project.yaml"},
								targetFilePath: []string{
									fmt.Sprintf("%s_project.yaml", pluginName),
								},
								targetFileMode: tgzModeFile,
							},
						},
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages: %w", err)
			}

			result[plat.String()] = &releaseOfficialPluginPkg{
				platform:     plat,
				tempFileName: tempFileName,
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate official plugin packages: %w", err)
	}

	return conv.MapValueToSlice(result), nil
}

func convPlatToOfficialPluginDirName(plat platform.Platform) string {
	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
		return "plugins_linux_x86_64"
	}

	if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
		return "plugins_linux_aarch64"
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
		return "plugins_darwin_x86_64"
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
		return "plugins_windows_x86_64"
	}

	return ""
}

func convOfficialPluginDirNameToPlat(dirName string) platform.Platform {
	switch dirName {
	case "plugins_linux_x86_64":
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case "plugins_linux_aarch64":
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	case "plugins_windows_x86_64":
		{
			return platform.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
		}
	case "plugins_darwin_x86_64":
		{
			return platform.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64)
		}
	default:
		{
			return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
