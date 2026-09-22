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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IPluginV2 defines the interface of plugin.
type IPluginV2 interface {
	IReleasePlugin

	// UploadOriginPluginV2 uploads the origin plugin package v2.
	UploadOriginPluginV2(nCtx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginPluginV2PkgDetail, error)

	// PublishReleasePluginV2 generates release plugin package v2 by upload-id and is-shared.
	PublishReleasePluginV2(nCtx contextx.IContext, uploadID string, isShared bool) error
}

// UploadOriginPluginV2 uploads origin plugin package v2.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPluginV2(nCtx contextx.IContext, pluginFile io.ReadCloser) (_ *types.OriginPluginV2PkgDetail, retErr error) {
	if pluginFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin plugin package v2 package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, pluginFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package. failed to get temp file")

		return nil, err
	}
	defer func() {
		if errClose := checkingFile.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	detail, err := checkOriginPluginV2Pkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package. failed to check origin plugin package v2 package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package. failed to get temp file")

		return nil, err
	}
	defer func() {
		if errClose := uploadingFile.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	pkgFileName, err := pluginpkg.FormatPkgFileName(detail.PluginPkgName, types.ReleaseTypeOriginPluginV2, types.Generation2,
		platfmt.UnknownPlatform(), detail.Version)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package, failed to format package")

		return nil, fmt.Errorf("failed to upload origin plugin package v2 package: %w", err)
	}

	logger.G.Biz(nCtx).With("filename", pkgFileName).Info("formatting origin plugin package v2 package")

	pkgFileName = m.wrapOriginPackageName(pkgFileName)

	// upload to upstream.
	if err := m.upstreamOriginPluginV2.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginPluginV2.GetFile(nCtx, pkgFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePlugin(nCtx, detail.PluginPkgName, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	uploadID, err := m.storageUpload.CreatePluginV2Upload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginPluginV2,
		SavedName: pkgFileName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin package v2 package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// upload event.
	m.recordUploadEvent(nCtx, types.ReleaseTypeOriginPluginV2, detail.PluginPkgName, detail.Version, detail.Platforms)

	logger.G.Biz(nCtx).With("filename", pkgFileName).Info("uploaded origin plugin package v2 package to upstream")

	return detail, nil
}

// PluginV2Project represents the project.yml file.
type PluginV2Project struct {
	Name          string `yaml:"name"`
	Version       string `yaml:"version"`
	Description   string `yaml:"description"`
	DescriptionEn string `yaml:"description_en"`
	Scenario      string `yaml:"scenario"`
	ScenarioEn    string `yaml:"scenario_en"`
	ConfigFile    string `yaml:"config_file"`
	ConfigFormat  string `yaml:"config_format"`
	LaunchNode    string `yaml:"launch_node"`

	PortRange       types.PluginPkgAvailablePortRange `yaml:"port_range"`
	Control         PluginV2Control                   `yaml:"control"`
	ConfigTemplates []PluginV2ConfigTemplate          `yaml:"config_templates"`
}

func validateOptionalPluginPkgPortRange(portRange types.PluginPkgAvailablePortRange) error {
	if portRange == "" {
		return nil
	}

	return portRange.Validate()
}

func buildPluginV2PkgBindAddressAllocated(portRange types.PluginPkgAvailablePortRange) types.BindAddressAllocated {
	if portRange == "" {
		return types.BindAddressAllocated{}
	}

	return types.BindAddressAllocated{
		Enable:                 true,
		BindIP:                 "127.0.0.1",
		BindPortAvailableRange: portRange,
	}
}

// PluginV2Control represents the project.yml file's control field.
type PluginV2Control struct {
	StartCmd   string `yaml:"start"`
	StopCmd    string `yaml:"stop"`
	RestartCmd string `yaml:"restart"`
	ReloadCmd  string `yaml:"reload"`
	DebugCmd   string `yaml:"debug"`
	VersionCmd string `yaml:"version"`
	HealthCmd  string `yaml:"health"`
	KillCmd    string `yaml:"kill"`
}

// PluginV2ConfigTemplate represents the project.yml file's config_templates field.
type PluginV2ConfigTemplate struct {
	PluginVersion string            `yaml:"plugin_version"`
	Name          string            `yaml:"name"`
	Version       string            `yaml:"version"`
	FilePath      string            `yaml:"file_path"`
	Format        string            `yaml:"format"`
	IsMainConfig  string            `yaml:"is_main_config"`
	SourcePath    string            `yaml:"source_path"`
	Variables     *PluginV2Property `yaml:"variables"`
}

// PluginV2Property represents the project.yml file's variables field.
type PluginV2Property struct {
	Title      string                       `yaml:"title,omitempty"`
	Type       string                       `yaml:"type,omitempty"`
	Required   bool                         `yaml:"required,omitempty"`
	Default    any                          `yaml:"default,omitempty"`
	Items      *PluginV2Property            `yaml:"items,omitempty"`
	Properties map[string]*PluginV2Property `yaml:"properties,omitempty"`
}

/**
 * plugin files contains:
 * bin/
 * etc/
 * project.yml
 * others can be ignored.
 */
// nolint: funlen
func checkOriginPluginV2Pkg(file io.ReadCloser) (*types.OriginPluginV2PkgDetail, error) {
	detail := types.NewOriginPluginV2PkgDetail()
	multiPlatConfigTplSourceContent := make(map[string]map[string]string)

	tgzReadRules := []tgzReadRule{
		{
			filePathRegex: []string{
				buildPrefixMatchRegex(originalPluginPkgV2DirNamePlatPrefix),
				".*",
				buildFullMatchRegex(originalPluginPkgV2FileNameProject)},
			callback: func(path []string, projectFile io.Reader) error {
				plat := convPluginV2DirNameToPlat(path[0])
				detail.Platforms = append(detail.Platforms, plat)

				pluginProject := new(PluginV2Project)
				if err := yaml.NewDecoder(projectFile).Decode(pluginProject); err != nil {
					return fmt.Errorf("failed to decode project.yaml")
				}
				if err := validateOptionalPluginPkgPortRange(pluginProject.PortRange); err != nil {
					return fmt.Errorf("invalid port_range for platform(%s): %w", plat.String(), err)
				}

				detail.PluginPkgName = pluginProject.Name
				detail.Version = pluginProject.Version
				detail.Description = pluginProject.Description
				detail.DescriptionEn = pluginProject.DescriptionEn
				detail.Scenario = pluginProject.Scenario
				detail.ScenarioEn = pluginProject.ScenarioEn
				detail.ConfigFile = pluginProject.ConfigFile
				detail.ConfigFormat = pluginProject.ConfigFormat
				detail.LaunchNode = pluginProject.LaunchNode
				detail.BindAddressAllocated[plat.String()] = buildPluginV2PkgBindAddressAllocated(pluginProject.PortRange)

				if _, ok := detail.ConfigTemplates[plat.String()]; !ok {
					detail.ConfigTemplates[plat.String()] = make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))
				}

				pkgConfigTemplates, err := parsePluginV2PkgConfigTemplateFromProject(pluginProject)
				if err != nil {
					return fmt.Errorf("failed to get config templates: %w", err)
				}
				detail.ConfigTemplates[plat.String()] = pkgConfigTemplates

				detail.Controller[plat.String()] = buildPluginV2PkgController(plat, pluginProject)

				return nil
			},
		},
		{
			filePathRegex: []string{
				buildPrefixMatchRegex(originalPluginPkgV2DirNamePlatPrefix),
				".*",
				buildFullMatchRegex(originalPluginPkgV2DirNameEtc),
				buildSuffixMatchRegex(originalPluginPkgV2FileNameEtcExt)},
			callback: func(path []string, tplFile io.Reader) error {
				plat := convPluginV2DirNameToPlat(path[0])

				content, err := io.ReadAll(tplFile)
				if err != nil {
					return fmt.Errorf("failed to read (%s) template file: %w", path[len(path)-1], err)
				}

				if _, ok := multiPlatConfigTplSourceContent[plat.String()]; !ok {
					multiPlatConfigTplSourceContent[plat.String()] = make(map[string]string)
				}

				// Notice: The plugin package is read and processed on the Linux system
				// so the sourcePath here should also be processed using the Linux separator.
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

func parsePluginV2PkgConfigTemplateFromProject(pluginProject *PluginV2Project) ([]types.PluginPkgConfigTemplate, error) {
	pkgConfigTemplates := make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))

	for idx, configTemplate := range pluginProject.ConfigTemplates {
		isMainConfig, err := conv.StringToBool(configTemplate.IsMainConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parseExternalPluginPkgConfigTemplateFromProject is_main_config: %w", err)
		}

		pkgConfigTemplates[idx] = types.PluginPkgConfigTemplate{
			Name:         configTemplate.Name,
			FilePath:     configTemplate.FilePath,
			SourcePath:   configTemplate.SourcePath,
			IsMainConfig: isMainConfig,
			Variables:    make(map[string]*types.PluginPkgConfigTemplateProperty),
		}

		variables := convPluginV2PropertyToTypes(configTemplate.Variables)
		if variables != nil {
			pkgConfigTemplates[idx].Variables[variables.Title] = variables
		}
	}

	return pkgConfigTemplates, nil
}

// notice: because the official plugin pkg's control cmd is provided by the nodemgr, so we specify the script path.
func buildPluginV2PkgController(plat platfmt.Platform, pluginProject *PluginV2Project) types.ProcessController {
	fn := func(path ...string) []string {
		if len(path) == 0 || path[len(path)-1] == "" {
			return []string{}
		}

		return path
	}

	return types.ProcessController{
		StartCmd:   tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.StartCmd)...),
		StopCmd:    tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.StopCmd)...),
		RestartCmd: tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.RestartCmd)...),
		ReloadCmd:  tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.ReloadCmd)...),
		DebugCmd:   tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.DebugCmd)...),
		KillCmd:    tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.KillCmd)...),
		VersionCmd: tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.VersionCmd)...),
		HealthCmd:  tool.JoinPath(plat.OS, fn(pluginV2PkgDirNameBin, pluginProject.Control.HealthCmd)...),
	}
}

func convPluginV2PropertyToTypes(property *PluginV2Property) *types.PluginPkgConfigTemplateProperty {
	if property == nil {
		return nil
	}

	pluginPkgConfigTemplateProperty := &types.PluginPkgConfigTemplateProperty{
		Title:      property.Title,
		Type:       property.Type,
		Required:   property.Required,
		Default:    property.Default,
		Properties: make(map[string]*types.PluginPkgConfigTemplateProperty),
	}

	if property.Items != nil {
		pluginPkgConfigTemplateProperty.Properties[property.Items.Title] = convPluginV2PropertyToTypes(property.Items)
	}

	if property.Properties != nil {
		pluginPkgConfigTemplateProperty.Properties = make(map[string]*types.PluginPkgConfigTemplateProperty)
		for k, v := range property.Properties {
			pluginPkgConfigTemplateProperty.Properties[k] = convPluginV2PropertyToTypes(v)
		}
	}

	return pluginPkgConfigTemplateProperty
}

const releasePluginV2Label = "v2"

// PublishReleasePluginV2 generates release plugin by upload-id and is-shared.
// nolint: funlen,gocognit
func (m *Manager) PublishReleasePluginV2(nCtx contextx.IContext, uploadID string, isShared bool) (retErr error) {
	up, err := m.storageUpload.GetPluginV2Upload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin package v2, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginPluginV2 {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin package v2, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginPluginV2.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin package v2, failed to get file")

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin package v2, failed to get content")

		return err
	}
	defer func() {
		if errClose := originContent.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v2 package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v2 package. failed to get temp file")

		return err
	}
	defer func() {
		if errClose := checkingFile.Close(); errClose != nil {
			retErr = errors.Join(retErr, errClose)
		}
	}()

	detail, err := checkOriginPluginV2Pkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v2 package. failed to check origin plugin package v2 package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generatePluginV2Pkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v2, failed to generate plugin pkg")

		return err
	}

	gp := gopool.NewPool()
	gen := types.Generation2
	pluginPkgName := detail.PluginPkgName
	pluginPkgVersion := detail.Version

	releasesMap := make(map[string]*types.ReleasePlugin)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() (retErr error) {
			// generate package name.
			pkgName, err := pluginpkg.FormatPkgFileName(pluginPkgName, types.ReleaseTypePlugin, gen, pkg.platform, pluginPkgVersion)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release plugin package v2 package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v2, failed to get temp file")

				return err
			}
			defer func() {
				if errClose := generatedFile.Close(); errClose != nil {
					retErr = errors.Join(retErr, errClose)
				}
			}()

			if err = m.upstreamReleasePlugin.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v2, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleasePlugin.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v2, failed to get temp file")

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
						releasePluginV2Label,
					},
					FileName:  file.Info().Name,
					MD5:       file.Info().MD5,
					Enabled:   false,
					AsDefault: false,
					IsShared:  isShared && nCtx.TenantID() == tenant.SystemTenantID,
					IsSynced:  false,
					UpdatedAt: time.Now(),
					Operator:  nCtx.BKUsername(),
				},
				ReleaseAdditionInfoPlugin: types.ReleaseAdditionInfoPlugin{
					OriginUploadID:       up.UploadID,
					OriginUploadCategory: up.Category,
					TemplateRendererType: types.TemplateRendererTypeJinja2,
					LaunchNodeType:       types.LaunchNodeType(strings.ToLower(detail.LaunchNode)),
					ConfigTemplates:      detail.ConfigTemplates[pkg.platform.String()],
					PluginController:     detail.Controller[pkg.platform.String()],
					Description:          detail.Description,
					DescriptionEn:        detail.DescriptionEn,
					Scenario:             detail.Scenario,
					ScenarioEn:           detail.ScenarioEn,
					BindAddressAllocated: detail.BindAddressAllocated[pkg.platform.String()],
				},
			}

			releasesMap[pkg.platform.String()] = pluginInfo

			// record publish event.
			m.recordPublishEvent(nCtx, &pluginInfo.Release)

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v2, failed to upload to upstream")

		return err
	}

	// upsert release plugin package v2.
	if err = m.storageRelease.UpsertManyReleasePlugin(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin package v2, failed to upsert release plugin package v2")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release plugins package v2")

	return nil
}

type releasePluginV2Pkg struct {
	platform     platfmt.Platform
	tempFileName string
}

// generatePluginV2Pkg generates plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generatePluginV2Pkg(nCtx contextx.IContext,
	originDetail *types.OriginPluginV2PkgDetail,
	originLocalFileName string) ([]*releasePluginV2Pkg, error) {

	// local origin plugin package v2.
	localOrigin, err := m.tempFileGroup.GetFile(nCtx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	// local plugin bintool v2.
	localPluginBinTool, err := m.fetchReleasePluginBinToolToLocal(nCtx, types.ReleaseNamePluginBinToolV2)
	if err != nil {
		return nil, err
	}

	pluginName := originDetail.PluginPkgName

	gp := gopool.NewPool()

	result := make(map[string]*releasePluginV2Pkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() (err error) {
			// create target file.
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create plugin package temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}
			defer func() {
				if errClose := targetFile.Close(); errClose != nil {
					err = errors.Join(err, errClose)
				}
			}()

			// open all source files.
			origiPluginFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin package file: %w", err)
			}
			defer func() {
				if errClose := origiPluginFile.Close(); errClose != nil {
					err = errors.Join(err, errClose)
				}
			}()

			originPluginBinToolFile, err := localPluginBinTool.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin bintool file: %w", err)
			}
			defer func() {
				if errClose := originPluginBinToolFile.Close(); errClose != nil {
					err = errors.Join(err, errClose)
				}
			}()

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{
						targetFilePath: []string{pluginV2PkgDirNameBin},
						targetFileMode: tgzModeDir,
					},
					{
						targetFilePath: []string{pluginV2PkgDirNameEtc},
						targetFileMode: tgzModeDir,
					},
				},
				[]*tgzWriteRuleStream{
					// get things from origin plugin.
					{
						sourceFile: origiPluginFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{convPlatToPluginV2DirName(plat), pluginName, originalPluginPkgV2DirNameBin,
									tgzPathMatchingSegment2},
								targetFilePath: []string{
									pluginV2PkgDirNameBin, tgzPathMatchingSegment2,
								},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{convPlatToPluginV2DirName(plat), pluginName, originalPluginPkgV2FileNameProject},
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
								targetFilePath: []string{pluginV2PkgDirNameBin, tgzPathMatchingSegment2},
								targetFileMode: tgzModeExe,
							},
						},
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages: %w", err)
			}

			result[plat.String()] = &releasePluginV2Pkg{
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
	originalPluginPkgV2FileNameProject       = "project.yaml"
	originalPluginPkgV2FileNameEtcExt        = ".tpl"
	originalPluginPkgV2DirNameBin            = "bin"
	originalPluginPkgV2DirNameEtc            = "etc"
	originalPluginPkgV2DirNamePlatPrefix     = "plugins_"
	originalPluginPkgV2DirNamePlatSplitTimes = 3

	pluginV2PkgDirNameBin = "bin"
	pluginV2PkgDirNameEtc = "etc"
)

func convPlatToPluginV2DirName(plat platfmt.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalPluginPkgV2DirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convPluginV2DirNameToPlat(dirName string) platfmt.Platform {
	if !strings.HasPrefix(dirName, originalPluginPkgV2DirNamePlatPrefix) {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalPluginPkgV2DirNamePlatSplitTimes)
	if len(dirSplit) != originalPluginPkgV2DirNamePlatSplitTimes {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platfmt.Normalize(dirSplit[1], dirSplit[2])
	if err != nil {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
