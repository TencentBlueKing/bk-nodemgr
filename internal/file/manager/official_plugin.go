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
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/opluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IOfficialPlugin defines the interface of official plugin.
type IOfficialPlugin interface {
	// UploadOriginOfficialPlugin uploads the origin official plugin.
	UploadOriginOfficialPlugin(nCtx contextx.IContext, pluginFile io.ReadCloser) (
		*types.OriginOfficialPluginPkgDetail, error)

	// PublishReleaseOfficialPlugin generates release official plugin by upload-id.
	PublishReleaseOfficialPlugin(nCtx contextx.IContext, uploadID string) error
}

// UploadOriginOfficialPlugin uploads origin official plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginOfficialPlugin(nCtx contextx.IContext, officialPluginFile io.ReadCloser) (
	*types.OriginOfficialPluginPkgDetail, error) {

	if officialPluginFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin official plugin package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, officialPluginFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginOfficialPluginPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package. failed to check origin official plugin package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package. failed to get temp file")

		return nil, err
	}

	gen := types.Generation2
	originalPkgName, err := opluginpkg.FormatPkgName(detail.Name, types.ReleaseTypeOriginOfficialPlugin, gen, platform.EmptyPlatform(), detail.Version)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package, failed to format package")

		return nil, fmt.Errorf("failed to upload origin official plugin package: %w", err)
	}

	logger.G.Biz(nCtx).With("filename", originalPkgName).Info("formatting origin official plugin package")

	originalPkgName = m.wrapOriginPackageName(originalPkgName)

	// upload to upstream.
	if err := m.upstreamOriginOfficialPlugin.Store(nCtx, fileiface.FileInfo{Name: originalPkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginOfficialPlugin.GetFile(nCtx, originalPkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseOfficialPlugin(nCtx, detail.Name, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateOfficialPluginUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginOfficialPlugin,
		SavedName: originalPkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin official plugin package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	logger.G.Biz(nCtx).With("filename", originalPkgName).Info("uploaded origin official plugin package to upstream")

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
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, originalOfficialPluginFileNameProject},
			callback: func(path []string, projectFile io.Reader) error {
				detail.Platforms = append(detail.Platforms, convOfficialPluginDirNameToPlat(path[0]))

				pluginProject := new(OfficialPluginProject)
				if err := yaml.NewDecoder(projectFile).Decode(pluginProject); err != nil {
					return fmt.Errorf("failed to decode project.yaml")
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
		{
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, originalOfficialPluginDirNameEtc, originalOfficialPluginFileNameEtcExt},
			callback: func(path []string, tplFile io.Reader) error {
				for idx := range detail.ConfigTemplates {
					if !strings.Contains(detail.ConfigTemplates[idx].SourcePath, path[len(path)-1]) {
						continue
					}

					content, err := io.ReadAll(tplFile)
					if err != nil {
						return fmt.Errorf("failed to read (%s) template file: %w", path[len(path)-1], err)
					}
					detail.ConfigTemplates[idx].SourceContent = string(content)
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
// nolint: funlen,gocognit
func (m *Manager) PublishReleaseOfficialPlugin(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetOfficialPluginUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release official plugin, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginOfficialPlugin {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release official plugin, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginOfficialPlugin.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release official plugin, failed to get file")

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release official plugin, failed to get content")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release official plugin package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release official plugin package. failed to get temp file")

		return err
	}

	detail, err := checkOriginOfficialPluginPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release official plugin package. failed to check origin official plugin package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateOfficialPluginPkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release official plugin, failed to generate official plugin pkg")

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
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release official plugin package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release official plugin, failed to get temp file")

				return err
			}

			if err = m.upstreamReleaseOfficialPlugin.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release official plugin, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleaseOfficialPlugin.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release official plugin, failed to get temp file")

				return err
			}

			releasesMap[pkg.platform.String()] = &types.ReleaseOfficialPlugin{
				Release: types.Release{
					Name:         detail.Name,
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
					Operator:     nCtx.BKUsername(),
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
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release official plugin, failed to upload to upstream")

		return err
	}

	// upsert release official plugin.
	if err = m.storageRelease.UpsertManyReleaseOfficialPlugin(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release official plugin, failed to upsert release official plugin")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release official plugins")

	return nil
}

type releaseOfficialPluginPkg struct {
	platform     platform.Platform
	tempFileName string
}

// generateOfficialPluginPkg generates official plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateOfficialPluginPkg(nCtx contextx.IContext,
	originDetail *types.OriginOfficialPluginPkgDetail,
	originLocalFileName string) ([]*releaseOfficialPluginPkg, error) {

	// local origin official plugin.
	localOrigin, err := m.tempFileGroup.GetFile(nCtx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	// local plugin bintool.
	localPluginBinTool, err := m.fetchReleasePluginBinToolToLocal(nCtx)
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
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create official plugin pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			origiOfficialPluginFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin official plugin file: %w", err)
			}

			originPluginBinToolFile, err := localPluginBinTool.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin bintool file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{
						targetFilePath: []string{officialPluginPkgDirNameBin},
						targetFileMode: tgzModeDir,
					},
					{
						targetFilePath: []string{officialPluginPkgDirNameEtc},
						targetFileMode: tgzModeDir,
					},
				},
				[]*tgzWriteRuleStream{
					// get things from origin official plugin.
					{
						sourceFile: origiOfficialPluginFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{convPlatToOfficialPluginDirName(plat), pluginName, originalOfficialPluginDirNameBin,
									tgzPathNameAny2},
								targetFilePath: []string{
									officialPluginPkgDirNameBin, tgzPathNameAny2,
								},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{convPlatToOfficialPluginDirName(plat), pluginName, originalOfficialPluginFileNameProject},
								targetFilePath: []string{
									fmt.Sprintf("project_%s.yaml", pluginName),
								},
								targetFileMode: tgzModeFile,
							},
						},
					},
					// get things from origin plugin bintool.
					{
						sourceFile: originPluginBinToolFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, convPlatToPluginBinToolDirName(plat), tgzPathNameAny2},
								targetFilePath: []string{officialPluginPkgDirNameBin, tgzPathNameAny2},
								targetFileMode: tgzModeExe,
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

const (
	originalOfficialPluginFileNameProject       = "project.yaml"
	originalOfficialPluginFileNameEtcExt        = ".tpl"
	originalOfficialPluginDirNameBin            = "bin"
	originalOfficialPluginDirNameEtc            = "etc"
	originalOfficialPluginDirNamePlatPrefix     = "plugins_"
	originalOfficialPluginDirNamePlatSplitTimes = 3

	officialPluginPkgDirNameBin = "bin"
	officialPluginPkgDirNameEtc = "etc"
)

func convPlatToOfficialPluginDirName(plat platform.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalOfficialPluginDirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convOfficialPluginDirNameToPlat(dirName string) platform.Platform {
	if !strings.HasPrefix(dirName, originalOfficialPluginDirNamePlatPrefix) {
		return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalOfficialPluginDirNamePlatSplitTimes)
	if len(dirSplit) != originalOfficialPluginDirNamePlatSplitTimes {
		return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platform.Normalize(dirSplit[1], dirSplit[2])
	if err != nil {
		return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
