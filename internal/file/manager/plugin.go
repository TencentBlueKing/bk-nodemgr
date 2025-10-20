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
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/pluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IPlugin defines the interface of plugin.
type IPlugin interface {
	// UploadOriginPlugin uploads the origin plugin.
	UploadOriginPlugin(nCtx contextx.IContext, pluginFile io.ReadCloser) (
		*types.OriginPluginPkgDetail, error)

	// PublishReleasePlugin generates release plugin by upload-id.
	PublishReleasePlugin(nCtx contextx.IContext, uploadID string) error
}

// UploadOriginPlugin uploads origin plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPlugin(nCtx contextx.IContext, pluginFile io.ReadCloser) (
	*types.OriginPluginPkgDetail, error) {

	if pluginFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin plugin package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, pluginFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginPluginPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package. failed to check origin plugin package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package. failed to get temp file")

		return nil, err
	}

	gen := types.Generation2
	originalPkgName, err := pluginpkg.FormatPkgName(detail.Name, types.ReleaseTypeOriginPlugin, gen, platfmt.EmptyPlatform(), detail.Version)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package, failed to format package")

		return nil, fmt.Errorf("failed to upload origin plugin package: %w", err)
	}

	logger.G.Biz(nCtx).With("filename", originalPkgName).Info("formatting origin plugin package")

	originalPkgName = m.wrapOriginPackageName(originalPkgName)

	// upload to upstream.
	if err := m.upstreamOriginPlugin.Store(nCtx, fileiface.FileInfo{Name: originalPkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginPlugin.GetFile(nCtx, originalPkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePlugin(nCtx, detail.Name, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreatePluginUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginPlugin,
		SavedName: originalPkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	logger.G.Biz(nCtx).With("filename", originalPkgName).Info("uploaded origin plugin package to upstream")

	return detail, nil
}

// PluginProject represents the project.yml file.
type PluginProject struct {
	Name            string                 `yaml:"name"`
	Version         string                 `yaml:"version"`
	Description     string                 `yaml:"description"`
	Scenario        string                 `yaml:"scenario"`
	ConfigFile      string                 `yaml:"config_file"`
	ConfigFormat    string                 `yaml:"config_format"`
	LaunchNode      string                 `yaml:"launch_node"`
	Control         PluginControl          `yaml:"control"`
	ConfigTemplates []PluginConfigTemplate `yaml:"config_templates"`
}

// PluginControl represents the project.yml file's control field.
type PluginControl struct {
	StartCmd   string `yaml:"start"`
	StopCmd    string `yaml:"stop"`
	RestartCmd string `yaml:"restart"`
	ReloadCmd  string `yaml:"reload"`
	VersionCmd string `yaml:"version"`
	HealthCmd  string `yaml:"health"`
	KillCmd    string `yaml:"kill"`
}

// PluginConfigTemplate represents the project.yml file's config_templates field.
type PluginConfigTemplate struct {
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
// nolint: funlen
func checkOriginPluginPkg(file io.ReadCloser) (*types.OriginPluginPkgDetail, error) {
	detail := types.NewOriginPluginPkgDetail()
	multiPlatConfigTplSourceContent := make(map[string]map[string]string)

	tgzReadRules := []tgzReadRule{
		{
			filePathRegex: []string{
				buildPrefixMatchRegex(originalPluginDirNamePlatPrefix),
				".*",
				buildFullMatchRegex(originalPluginFileNameProject)},
			callback: func(path []string, projectFile io.Reader) error {
				plat := convPluginDirNameToPlat(path[0])
				detail.Platforms = append(detail.Platforms, plat)

				pluginProject := new(PluginProject)
				if err := yaml.NewDecoder(projectFile).Decode(pluginProject); err != nil {
					return fmt.Errorf("failed to decode project.yaml")
				}

				detail.Name = pluginProject.Name
				detail.Version = pluginProject.Version
				detail.Description = pluginProject.Description
				detail.Scenario = pluginProject.Scenario
				detail.ConfigFile = pluginProject.ConfigFile
				detail.ConfigFormat = pluginProject.ConfigFormat
				detail.LaunchNode = pluginProject.LaunchNode

				if _, ok := detail.ConfigTemplates[plat.String()]; !ok {
					detail.ConfigTemplates[plat.String()] = make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))
				}

				pkgConfigTemplates, err := parsePluginPkgConfigTemplateFromProject(pluginProject)
				if err != nil {
					return fmt.Errorf("failed to get config templates: %w", err)
				}
				detail.ConfigTemplates[plat.String()] = pkgConfigTemplates

				detail.Controller[plat.String()] = buildPluginPkgController(plat, pluginProject)

				return nil
			},
		},
		{
			filePathRegex: []string{
				buildPrefixMatchRegex(originalPluginDirNamePlatPrefix),
				".*",
				buildFullMatchRegex(originalPluginDirNameEtc),
				buildSuffixMatchRegex(originalPluginFileNameEtcExt)},
			callback: func(path []string, tplFile io.Reader) error {
				plat := convPluginDirNameToPlat(path[0])

				content, err := io.ReadAll(tplFile)
				if err != nil {
					return fmt.Errorf("failed to read (%s) template file: %w", path[len(path)-1], err)
				}

				if _, ok := multiPlatConfigTplSourceContent[plat.String()]; !ok {
					multiPlatConfigTplSourceContent[plat.String()] = make(map[string]string)
				}

				sourcePath := filepath.Clean(filepath.Join(originalPluginDirNameEtc, path[len(path)-1]))
				multiPlatConfigTplSourceContent[plat.String()][sourcePath] = string(content)

				return nil
			},
		},
	}

	err := checkTgz(file, tgzReadRules)
	if err != nil {
		return nil, err
	}

	for platStr, configTemplates := range detail.ConfigTemplates {
		configTplSourceContent := multiPlatConfigTplSourceContent[platStr]
		for idx, configTemplate := range configTemplates {
			if multiPlatConfigTplSourceContent[platStr] == nil {
				continue
			}

			sourcePath := filepath.Clean(configTemplate.SourcePath)
			if configTplSourceContent[sourcePath] == "" {
				continue
			}

			configTemplates[idx].SourceContent = configTplSourceContent[sourcePath]
		}
	}

	return detail, nil
}

func parsePluginPkgConfigTemplateFromProject(pluginProject *PluginProject) ([]types.PluginPkgConfigTemplate, error) {
	pkgConfigTemplates := make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))

	for idx, configTemplate := range pluginProject.ConfigTemplates {
		isMainConfig, err := conv.StringToBool(configTemplate.IsMainConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parseExternalPluginPkgConfigTemplateFromProject is_main_config: %w", err)
		}

		pkgConfigTemplates[idx] = types.PluginPkgConfigTemplate{
			PluginVersion: configTemplate.PluginVersion,
			Name:          configTemplate.Name,
			Version:       configTemplate.Version,
			FilePath:      configTemplate.FilePath,
			Format:        configTemplate.Format,
			IsMainConfig:  isMainConfig,
			SourcePath:    configTemplate.SourcePath,
			Variables:     convPropertyToTypes(configTemplate.Variables),
		}
	}

	return pkgConfigTemplates, nil
}

// notice: because the official plugin pkg's control cmd is provided by the nodemgr, so we specify the script path.
func buildPluginPkgController(plat platfmt.Platform, pluginProject *PluginProject) types.ProcessController {
	if plat.OS == criteria.OSWindows {
		return types.ProcessController{
			StartCmd:   winpath.Join(pluginPkgDirNameBin, pluginProject.Control.StartCmd),
			StopCmd:    winpath.Join(pluginPkgDirNameBin, pluginProject.Control.StopCmd),
			RestartCmd: winpath.Join(pluginPkgDirNameBin, pluginProject.Control.RestartCmd),
			ReloadCmd:  winpath.Join(pluginPkgDirNameBin, pluginProject.Control.ReloadCmd),
			KillCmd:    winpath.Join(pluginPkgDirNameBin, pluginProject.Control.KillCmd),
			VersionCmd: winpath.Join(pluginPkgDirNameBin, pluginProject.Control.VersionCmd),
			HealthCmd:  winpath.Join(pluginPkgDirNameBin, pluginProject.Control.HealthCmd),
		}
	}

	return types.ProcessController{
		StartCmd:   filepath.Join(pluginPkgDirNameBin, pluginProject.Control.StartCmd),
		StopCmd:    filepath.Join(pluginPkgDirNameBin, pluginProject.Control.StopCmd),
		RestartCmd: filepath.Join(pluginPkgDirNameBin, pluginProject.Control.RestartCmd),
		ReloadCmd:  filepath.Join(pluginPkgDirNameBin, pluginProject.Control.ReloadCmd),
		KillCmd:    filepath.Join(pluginPkgDirNameBin, pluginProject.Control.KillCmd),
		VersionCmd: filepath.Join(pluginPkgDirNameBin, pluginProject.Control.VersionCmd),
		HealthCmd:  filepath.Join(pluginPkgDirNameBin, pluginProject.Control.HealthCmd),
	}
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

// PublishReleasePlugin generates release plugin by upload-id.
// nolint: funlen,gocognit
func (m *Manager) PublishReleasePlugin(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginPlugin {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginPlugin.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin, failed to get file")

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin, failed to get content")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package. failed to get temp file")

		return err
	}

	detail, err := checkOriginPluginPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package. failed to check origin plugin package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generatePluginPkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin, failed to generate plugin pkg")

		return err
	}

	gp := gopool.NewPool()
	gen := types.Generation2

	releasesMap := make(map[string]*types.ReleasePlugin)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := pluginpkg.FormatPkgName(detail.Name, types.ReleaseTypePlugin, gen, pkg.platform, detail.Version)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin, failed to get temp file")

				return err
			}

			if err = m.upstreamReleasePlugin.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleasePlugin.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin, failed to get temp file")

				return err
			}

			releasesMap[pkg.platform.String()] = &types.ReleasePlugin{
				Release: types.Release{
					Name:         detail.Name,
					Generation:   gen,
					Type:         types.ReleaseTypePlugin,
					Version:      detail.Version,
					Platform:     pkg.platform,
					Labels:       []string{},
					FileName:     file.Info().Name,
					MD5:          file.Info().MD5,
					Enabled:      false,
					AsDefault:    false,
					UpdatedAt:    time.Now(),
					Operator:     nCtx.BKUsername(),
					AdditionInfo: nil,
				},
				ReleaseAdditionInfoPlugin: types.ReleaseAdditionInfoPlugin{
					ConfigTemplates:  detail.ConfigTemplates[pkg.platform.String()],
					PluginController: detail.Controller[pkg.platform.String()],
				},
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin, failed to upload to upstream")

		return err
	}

	// upsert release plugin.
	if err = m.storageRelease.UpsertManyReleasePlugin(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin, failed to upsert release plugin")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release plugins")

	return nil
}

type releasePluginPkg struct {
	platform     platfmt.Platform
	tempFileName string
}

// generatePluginPkg generates plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generatePluginPkg(nCtx contextx.IContext,
	originDetail *types.OriginPluginPkgDetail,
	originLocalFileName string) ([]*releasePluginPkg, error) {

	// local origin plugin.
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

	result := make(map[string]*releasePluginPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create plugin pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			origiPluginFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin file: %w", err)
			}

			originPluginBinToolFile, err := localPluginBinTool.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin bintool file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{
						targetFilePath: []string{pluginPkgDirNameBin},
						targetFileMode: tgzModeDir,
					},
					{
						targetFilePath: []string{pluginPkgDirNameEtc},
						targetFileMode: tgzModeDir,
					},
				},
				[]*tgzWriteRuleStream{
					// get things from origin plugin.
					{
						sourceFile: origiPluginFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{convPlatToPluginDirName(plat), pluginName, originalPluginDirNameBin,
									tgzPathMatchingSegment2},
								targetFilePath: []string{
									pluginPkgDirNameBin, tgzPathMatchingSegment2,
								},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{convPlatToPluginDirName(plat), pluginName, originalPluginFileNameProject},
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
								sourceFilePath: []string{tgzPathMatchingSegment1, convPlatToPluginBinToolDirName(plat), tgzPathMatchingSegment2},
								targetFilePath: []string{pluginPkgDirNameBin, tgzPathMatchingSegment2},
								targetFileMode: tgzModeExe,
							},
						},
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages: %w", err)
			}

			result[plat.String()] = &releasePluginPkg{
				platform:     plat,
				tempFileName: tempFileName,
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate plugin packages: %w", err)
	}

	return conv.MapValueToSlice(result), nil
}

const (
	originalPluginFileNameProject       = "project.yaml"
	originalPluginFileNameEtcExt        = ".tpl"
	originalPluginDirNameBin            = "bin"
	originalPluginDirNameEtc            = "etc"
	originalPluginDirNamePlatPrefix     = "plugins_"
	originalPluginDirNamePlatSplitTimes = 3

	pluginPkgDirNameBin = "bin"
	pluginPkgDirNameEtc = "etc"
)

func convPlatToPluginDirName(plat platfmt.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalPluginDirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convPluginDirNameToPlat(dirName string) platfmt.Platform {
	if !strings.HasPrefix(dirName, originalPluginDirNamePlatPrefix) {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalPluginDirNamePlatSplitTimes)
	if len(dirSplit) != originalPluginDirNamePlatSplitTimes {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platfmt.Normalize(dirSplit[1], dirSplit[2])
	if err != nil {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
