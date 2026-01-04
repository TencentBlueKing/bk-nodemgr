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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IPluginV3 defines the interface of plugin.
type IPluginV3 interface {
	// UploadOriginPluginV3 uploads origin plugin package v3.
	UploadOriginPluginV3(nCtx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginPluginV3PkgDetail, error)

	// PublishReleasePluginV3 generates release plugin package v3 by upload-id.
	PublishReleasePluginV3(nCtx contextx.IContext, uploadID string) error
}

// UploadOriginPluginV3 uploads origin plugin package v3.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPluginV3(nCtx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginPluginV3PkgDetail, error) {
	if pluginFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin plugin package v3 package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, pluginFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginPluginV3Pkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package. failed to check origin plugin package v3 package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package. failed to get temp file")

		return nil, err
	}

	pkgFileName, err := pluginpkg.FormatPkgFileName(detail.PluginPkgName, types.ReleaseTypeOriginPluginV3, types.Generation2,
		platfmt.UnknownPlatform(), detail.Version)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package, failed to format package")

		return nil, fmt.Errorf("failed to upload origin plugin package v3 package: %w", err)
	}

	logger.G.Biz(nCtx).With("filename", pkgFileName).Info("formatting origin plugin package v3 package")

	pkgFileName = m.wrapOriginPackageName(pkgFileName)

	// upload to upstream.
	if err := m.upstreamOriginPluginV3.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginPluginV3.GetFile(nCtx, pkgFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePlugin(nCtx, detail.PluginPkgName, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	uploadID, err := m.storageUpload.CreatePluginV3Upload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginPluginV3,
		SavedName: pkgFileName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v3 package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// upload event.
	m.recordUploadEvent(nCtx, types.ReleaseTypeOriginPluginV3, detail.PluginPkgName, detail.Version, detail.Platforms)

	logger.G.Biz(nCtx).With("filename", pkgFileName).Info("uploaded origin plugin package v3 package to upstream")

	return detail, nil
}

// PluginV3Project represents the project.yaml file.
type PluginV3Project struct {
	Name             string `yaml:"name"`
	Version          string `yaml:"version"`
	Description      string `yaml:"description"`
	DescriptionEn    string `yaml:"descriptionEn"`
	Scenario         string `yaml:"scenario"`
	ScenarioEn       string `yaml:"scenarioEn"`
	LaunchNode       string `yaml:"launchNode"`
	TemplateRenderer string `yaml:"templateRenderer"`
}

// PluginV3Definition represents the definition.yaml file's definition field.
type PluginV3Definition struct {
	ConfigTemplates []PluginV3ConfigTemplate `yaml:"configTemplates"`
	Control         PluginV3Control          `yaml:"control"`
}

// PluginV3Control represents the definition.yaml file's control field.
type PluginV3Control struct {
	StartCmd   string `yaml:"start"`
	StopCmd    string `yaml:"stop"`
	RestartCmd string `yaml:"restart"`
	ReloadCmd  string `yaml:"reload"`
	VersionCmd string `yaml:"version"`
	HealthCmd  string `yaml:"health"`
	KillCmd    string `yaml:"kill"`
}

// PluginV3ConfigTemplate represents the definition.yaml file's configTemplates field.
type PluginV3ConfigTemplate struct {
	Name         string              `yaml:"name"`
	FilePath     string              `yaml:"filePath"`
	SourcePath   string              `yaml:"sourcePath"`
	IsMainConfig bool                `yaml:"isMainConfig"`
	Variables    []*PluginV3Property `yaml:"variables"`
}

// PluginV3Property represents the definition.yaml file's configTemplates variables field.
type PluginV3Property struct {
	Title         string              `yaml:"title,omitempty"`
	Type          string              `yaml:"type,omitempty"`
	Required      bool                `yaml:"required,omitempty"`
	Default       any                 `yaml:"default,omitempty"`
	Description   string              `yaml:"description,omitempty"`
	DescriptionEn string              `yaml:"descriptionEn,omitempty"`
	Properties    []*PluginV3Property `yaml:"properties,omitempty"`
}

/**
 * plugin files contains:
 * bin/
 * etc/
 * project.yml
 * others can be ignored.
 */
// nolint: funlen
func checkOriginPluginV3Pkg(file io.ReadCloser) (*types.OriginPluginV3PkgDetail, error) {
	detail := types.NewOriginPluginV3PkgDetail()
	multiPlatConfigTplSourceContent := make(map[string]map[string]string)

	tgzReadRules := []tgzReadRule{
		{
			filePathRegex: []string{
				".*",
				buildFullMatchRegex(originalPluginPkgV3FileNameProject)},
			callback: func(_ []string, projectFile io.Reader) error {
				pluginProject := new(PluginV3Project)
				if err := yaml.NewDecoder(projectFile).Decode(pluginProject); err != nil {
					return fmt.Errorf("failed to decode project.yaml")
				}

				detail.PluginPkgName = pluginProject.Name
				detail.Version = pluginProject.Version
				detail.Description = pluginProject.Description
				detail.DescriptionEn = pluginProject.DescriptionEn
				detail.Scenario = pluginProject.Scenario
				detail.ScenarioEn = pluginProject.ScenarioEn
				detail.LaunchNode = pluginProject.LaunchNode
				detail.TemplateRenderer = types.TemplateRendererType(pluginProject.TemplateRenderer)

				return nil
			},
		},
		{
			filePathRegex: []string{
				".*",
				buildPrefixMatchRegex(originalPluginPkgV3DirNamePlatPrefix),
				buildFullMatchRegex(originalPluginPkgV3FileNameDefinition)},
			callback: func(path []string, definitionFile io.Reader) error {
				plat := convPluginV3DirNameToPlat(path[1])

				pluginDefinition := new(PluginV3Definition)
				if err := yaml.NewDecoder(definitionFile).Decode(pluginDefinition); err != nil {
					return fmt.Errorf("failed to decode definition.yaml for platform(%s): %w", plat.String(), err)
				}

				detail.Platforms = append(detail.Platforms, plat)
				detail.ConfigTemplates[plat.String()] = parsePluginV3PkgConfigTemplateFromDefinition(pluginDefinition)
				detail.Controller[plat.String()] = buildPluginV3PkgController(plat, pluginDefinition)

				return nil
			},
		},
		{
			filePathRegex: []string{
				".*",
				buildPrefixMatchRegex(originalPluginPkgV3DirNamePlatPrefix),
				buildFullMatchRegex(originalPluginPkgV3DirNameTemplates),
				buildSuffixMatchRegex(originalPluginPkgV3FileNameTemplatesExt)},
			callback: func(path []string, tplFile io.Reader) error {
				plat := convPluginV3DirNameToPlat(path[1])

				content, err := io.ReadAll(tplFile)
				if err != nil {
					return fmt.Errorf("failed to read (%s) template file: %w", path[len(path)-1], err)
				}

				if _, ok := multiPlatConfigTplSourceContent[plat.String()]; !ok {
					multiPlatConfigTplSourceContent[plat.String()] = make(map[string]string)
				}

				sourcePath := filepath.Join(path[2:]...)
				multiPlatConfigTplSourceContent[plat.String()][sourcePath] = string(content)

				return nil
			},
		},
	}

	err := checkTgz(file, tgzReadRules)
	if err != nil {
		return nil, err
	}

	for platStr, templates := range multiPlatConfigTplSourceContent {
		for idx, tpl := range detail.ConfigTemplates[platStr] {
			sourceContent, ok := templates[tpl.SourcePath]
			if !ok {
				continue
			}
			detail.ConfigTemplates[platStr][idx].SourceContent = sourceContent
		}
	}

	return detail, nil
}

func parsePluginV3PkgConfigTemplateFromDefinition(pluginDefinition *PluginV3Definition) []types.PluginPkgConfigTemplate {
	pkgConfigTemplates := make([]types.PluginPkgConfigTemplate, len(pluginDefinition.ConfigTemplates))

	for idx, configTemplate := range pluginDefinition.ConfigTemplates {
		pkgConfigTemplates[idx] = types.PluginPkgConfigTemplate{
			Name:         configTemplate.Name,
			FilePath:     configTemplate.FilePath,
			SourcePath:   configTemplate.SourcePath,
			IsMainConfig: configTemplate.IsMainConfig,
			Variables:    convPluginV3PropertyToTypes(configTemplate.Variables...),
		}
	}

	return pkgConfigTemplates
}

// notice: because the official plugin pkg's control cmd is provided by the nodemgr, so we specify the script path.
func buildPluginV3PkgController(plat platfmt.Platform, pluginDefinition *PluginV3Definition) types.ProcessController {
	fn := func(path ...string) []string {
		if len(path) == 0 || path[len(path)-1] == "" {
			return []string{}
		}

		return path
	}

	return types.ProcessController{
		StartCmd:   tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.StartCmd)...),
		StopCmd:    tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.StopCmd)...),
		RestartCmd: tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.RestartCmd)...),
		ReloadCmd:  tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.ReloadCmd)...),
		KillCmd:    tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.KillCmd)...),
		VersionCmd: tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.VersionCmd)...),
		HealthCmd:  tool.JoinPath(plat.OS, fn(pluginV3PkgDirNameBin, pluginDefinition.Control.HealthCmd)...),
	}
}

func convPluginV3PropertyToTypes(properties ...*PluginV3Property) map[string]*types.PluginPkgConfigTemplateProperty {
	if properties == nil {
		return nil
	}

	result := make(map[string]*types.PluginPkgConfigTemplateProperty)
	for _, property := range properties {
		if property == nil {
			continue
		}

		pluginPkgConfigTemplateProperty := &types.PluginPkgConfigTemplateProperty{
			Title:         property.Title,
			Type:          property.Type,
			Default:       property.Default,
			Required:      property.Required,
			Description:   property.Description,
			DescriptionEn: property.DescriptionEn,
		}

		if property.Properties != nil {
			pluginPkgConfigTemplateProperty.Properties = convPluginV3PropertyToTypes(property.Properties...)
		}

		result[property.Title] = pluginPkgConfigTemplateProperty
	}

	return result
}

const releasePluginV3Label = "v3"

// PublishReleasePluginV3 generates release plugin by upload-id.
// nolint: funlen,gocognit
func (m *Manager) PublishReleasePluginV3(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginV3Upload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin package v3, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginPluginV3 {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin package v3, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginPluginV3.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin package v3, failed to get file")

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin package v3, failed to get content")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v3 package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v3 package. failed to get temp file")

		return err
	}

	detail, err := checkOriginPluginV3Pkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v3 package. failed to check origin plugin package v3 package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generatePluginV3Pkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v3, failed to generate plugin pkg")

		return err
	}

	gp := gopool.NewPool()
	gen := types.Generation2
	pluginPkgName := detail.PluginPkgName
	pluginPkgVersion := detail.Version

	releasesMap := make(map[string]*types.ReleasePlugin)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := pluginpkg.FormatPkgFileName(pluginPkgName, types.ReleaseTypePlugin, gen, pkg.platform, pluginPkgVersion)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v3 package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v3, failed to get temp file")

				return err
			}

			if err = m.upstreamReleasePlugin.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v3, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleasePlugin.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v3, failed to get temp file")

				return err
			}

			pluginInfo := &types.ReleasePlugin{
				Release: types.Release{
					Name:       pluginPkgName,
					Generation: gen,
					Type:       types.ReleaseTypePlugin,
					Version:    pluginPkgVersion,
					Platform:   pkg.platform,
					Labels: []string{
						releasePluginV3Label,
					},
					FileName:  file.Info().Name,
					MD5:       file.Info().MD5,
					Enabled:   false,
					AsDefault: false,
					UpdatedAt: time.Now(),
					Operator:  nCtx.BKUsername(),
				},
				ReleaseAdditionInfoPlugin: types.ReleaseAdditionInfoPlugin{
					TemplateRendererType: detail.TemplateRenderer,
					LaunchNodeType:       types.LaunchNodeType(strings.ToLower(detail.LaunchNode)),
					ConfigTemplates:      detail.ConfigTemplates[pkg.platform.String()],
					PluginController:     detail.Controller[pkg.platform.String()],
					Description:          detail.Description,
					DescriptionEn:        detail.DescriptionEn,
					Scenario:             detail.Scenario,
					ScenarioEn:           detail.ScenarioEn,
				},
			}

			releasesMap[pkg.platform.String()] = pluginInfo

			// record publish event.
			m.recordPublishEvent(nCtx, &pluginInfo.Release)

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v3, failed to upload to upstream")

		return err
	}

	// upsert release plugin package v3.
	if err = m.storageRelease.UpsertManyReleasePlugin(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v3, failed to upsert release plugin package v3")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release plugins package v3")

	return nil
}

type releasePluginPkg struct {
	platform     platfmt.Platform
	tempFileName string
}

// generatePluginV3Pkg generates plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generatePluginV3Pkg(nCtx contextx.IContext,
	originDetail *types.OriginPluginV3PkgDetail,
	originLocalFileName string) ([]*releasePluginPkg, error) {

	// local origin plugin package v3.
	localOrigin, err := m.tempFileGroup.GetFile(nCtx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	// local plugin bintool v3.
	localPluginBinTool, err := m.fetchReleasePluginBinToolToLocal(nCtx, types.ReleaseNamePluginBinToolV3)
	if err != nil {
		return nil, err
	}

	pluginName := originDetail.PluginPkgName

	gp := gopool.NewPool()

	result := make(map[string]*releasePluginPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create plugin package temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			origiPluginFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin package file: %w", err)
			}

			originPluginBinToolFile, err := localPluginBinTool.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin bintool file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{
						targetFilePath: []string{pluginV3PkgDirNameBin},
						targetFileMode: tgzModeDir,
					},
					{
						targetFilePath: []string{pluginV3PkgDirNameEtc},
						targetFileMode: tgzModeDir,
					},
				},
				[]*tgzWriteRuleStream{
					// get things from origin plugin.
					{
						sourceFile: origiPluginFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{pluginName, convPlatToPluginV3DirName(plat), originalPluginPkgV3DirNameBin,
									tgzPathMatchingSegment2},
								targetFilePath: []string{
									pluginV3PkgDirNameBin, tgzPathMatchingSegment2,
								},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{pluginName, originalPluginPkgV3FileNameProject},
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
								targetFilePath: []string{pluginV3PkgDirNameBin, tgzPathMatchingSegment2},
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
	originalPluginPkgV3FileNameProject       = "project.yaml"
	originalPluginPkgV3FileNameDefinition    = "definition.yaml"
	originalPluginPkgV3FileNameTemplatesExt  = ".template"
	originalPluginPkgV3DirNameBin            = "bin"
	originalPluginPkgV3DirNameTemplates      = "templates"
	originalPluginPkgV3DirNamePlatPrefix     = "plugins_"
	originalPluginPkgV3DirNamePlatSplitTimes = 3

	pluginV3PkgDirNameBin = "bin"
	pluginV3PkgDirNameEtc = "etc"
)

func convPlatToPluginV3DirName(plat platfmt.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalPluginPkgV3DirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convPluginV3DirNameToPlat(dirName string) platfmt.Platform {
	if !strings.HasPrefix(dirName, originalPluginPkgV3DirNamePlatPrefix) {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalPluginPkgV3DirNamePlatSplitTimes)
	if len(dirSplit) != originalPluginPkgV3DirNamePlatSplitTimes {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platfmt.Normalize(dirSplit[1], dirSplit[2])
	if err != nil {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
