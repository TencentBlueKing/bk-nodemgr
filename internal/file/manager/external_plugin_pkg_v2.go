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
	"slices"
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IExternalPluginV2 defines the interface for external plugin.
type IExternalPluginV2 interface {
	// UploadOriginExternalPlugin uploads the origin external plugin.
	UploadOriginExternalPlugin(nCtx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginExternalPluginV2PkgDetail, error)

	// PublishReleaseExternalPlugin generates release external plugin by upload-id.
	PublishReleaseExternalPlugin(nCtx contextx.IContext, uploadID string) error
}

// UploadOriginExternalPlugin uploads origin external plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginExternalPlugin(nCtx contextx.IContext, externalPluginFile io.ReadCloser) (
	*types.OriginExternalPluginV2PkgDetail, error) {

	if externalPluginFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin external plugin package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, externalPluginFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginExternalPluginPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to check origin external plugin package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to get temp file")

		return nil, err
	}

	gen := types.Generation2
	pkgFileName, err := pluginpkg.FormatPkgFileName(
		detail.PluginPkgName, types.ReleaseTypeOriginExternalPluginV2, gen, platfmt.UnknownPlatform(), detail.Version)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package, failed to format package")

		return nil, fmt.Errorf("failed to upload origin external plugin package: %w", err)
	}

	logger.G.Biz(nCtx).With("filename", pkgFileName).Info("formatting origin external plugin package")

	pkgFileName = m.wrapOriginPackageName(pkgFileName)

	// upload to upstream.
	if err := m.upstreamOriginExternalPluginV2.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginExternalPluginV2.GetFile(nCtx, pkgFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePlugin(nCtx, detail.PluginPkgName, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateExternalPluginV2Upload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginExternalPluginV2,
		SavedName: pkgFileName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// record upload.
	m.recordUploadEvent(nCtx, types.ReleaseTypeOriginExternalPluginV2, pkgFileName, detail.Version,
		detail.Platforms)

	logger.G.Biz(nCtx).With("filename", pkgFileName).Info("uploaded origin external plugin package to upstream")

	return detail, nil
}

// ExternalPluginProject represents the project.yml file.
type ExternalPluginProject struct {
	Name          string `yaml:"name"`
	Version       string `yaml:"version"`
	Description   string `yaml:"description"`
	DescriptionEn string `yaml:"description_en"`
	Scenario      string `yaml:"scenario"`
	ScenarioEn    string `yaml:"scenario_en"`
	ConfigFile    string `yaml:"config_file"`
	ConfigFormat  string `yaml:"config_format"`
	LaunchNode    string `yaml:"launch_node"`

	Control PluginV2Control `yaml:"control"`

	ConfigTemplates []ExternalPluginConfigTemplate `yaml:"config_templates"`
}

// ExternalPluginConfigTemplate represents the project.yml file's config_templates field.
type ExternalPluginConfigTemplate struct {
	PluginVersion string            `yaml:"plugin_version"`
	Name          string            `yaml:"name"`
	Version       string            `yaml:"version"`
	FilePath      string            `yaml:"file_path"`
	Format        string            `yaml:"format"`
	IsMainConfig  string            `yaml:"is_main_config"`
	SourcePath    string            `yaml:"source_path"`
	Variables     *PluginV2Property `yaml:"variables"`
}

/**
 * plugin files contains:
 * bin/
 * etc/
 * project.yml
 * others can be ignored.
 */
// nolint: funlen, gocognit
func checkOriginExternalPluginPkg(file io.ReadCloser) (*types.OriginExternalPluginV2PkgDetail, error) {
	detail := types.NewOriginExternalPluginV2PkgDetail()
	multiPlatConfigTplSourceContent := make(map[string]map[string]string)

	tgzReadRules := []tgzReadRule{
		{
			filePathRegex: []string{
				buildPrefixMatchRegex(originalExternalPluginDirNamePlatPrefix),
				".*",
				buildFullMatchRegex(originalExternalPluginFileNameProject)},
			callback: func(path []string, projectFile io.Reader) error {
				plat := convExternalPluginDirNameToPlat(path[0])
				detail.Platforms = append(detail.Platforms, plat)

				pluginProject := new(ExternalPluginProject)
				if err := yaml.NewDecoder(projectFile).Decode(pluginProject); err != nil {
					return fmt.Errorf("failed to decode project.yaml")
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

				if _, ok := detail.ConfigTemplates[plat.String()]; !ok {
					detail.ConfigTemplates[plat.String()] = make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))
				}

				pkgConfigTemplates, err := parseExternalPluginPkgConfigTemplateFromProject(pluginProject)
				if err != nil {
					return fmt.Errorf("failed to parse external plugin project config templates: %w", err)
				}
				detail.ConfigTemplates[plat.String()] = pkgConfigTemplates

				detail.Controller[plat.String()] = buildExternalPluginPkgController(pluginProject)

				return nil
			},
		},
		{
			filePathRegex: []string{buildPrefixMatchRegex(originalExternalPluginDirNamePlatPrefix), ".*", ".*"},
			callback: func(path []string, _ io.Reader) error {
				plat := convExternalPluginDirNameToPlat(path[0])
				if plat.Arch == criteria.CPUArchUnknown || plat.OS == criteria.OSUnknown {
					// this not a platform directory.
					return nil
				}

				subDirPath := filepath.Join(path[2 : len(path)-1]...)

				platStr := plat.String()
				if _, ok := detail.SubDirPaths[platStr]; !ok {
					detail.SubDirPaths[platStr] = make(map[string]struct{})
				}

				detail.SubDirPaths[platStr][subDirPath] = struct{}{}

				return nil
			},
		},
		{
			filePathRegex: []string{
				buildPrefixMatchRegex(originalExternalPluginDirNamePlatPrefix),
				".*",
				buildFullMatchRegex(originalExternalPluginDirNameEtc),
				buildSuffixMatchRegex(originalExternalPluginFileNameEtcExt)},
			callback: func(path []string, tplFile io.Reader) error {
				plat := convExternalPluginDirNameToPlat(path[0])

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
	if err := checkTgz(file, tgzReadRules); err != nil {
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

func parseExternalPluginPkgConfigTemplateFromProject(pluginProject *ExternalPluginProject) ([]types.PluginPkgConfigTemplate, error) {
	pkgConfigTemplates := make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))

	for idx, configTemplate := range pluginProject.ConfigTemplates {
		isMainConfig, err := conv.StringToBool(configTemplate.IsMainConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parseExternalPluginPkgConfigTemplateFromProject is_main_config: %w", err)
		}

		pkgConfigTemplates[idx] = types.PluginPkgConfigTemplate{
			Name:         configTemplate.Name,
			FilePath:     configTemplate.FilePath,
			IsMainConfig: isMainConfig,
			SourcePath:   configTemplate.SourcePath,
			Variables:    make(map[string]*types.PluginPkgConfigTemplateProperty),
		}

		variables := convPluginV2PropertyToTypes(configTemplate.Variables)
		if variables != nil {
			pkgConfigTemplates[idx].Variables[variables.Title] = variables
		}
	}

	return pkgConfigTemplates, nil
}

func buildExternalPluginPkgController(pluginProject *ExternalPluginProject) types.ProcessController {
	return types.ProcessController{
		StartCmd:   pluginProject.Control.StartCmd,
		StopCmd:    pluginProject.Control.StopCmd,
		RestartCmd: pluginProject.Control.RestartCmd,
		ReloadCmd:  pluginProject.Control.ReloadCmd,
		KillCmd:    pluginProject.Control.KillCmd,
		VersionCmd: pluginProject.Control.VersionCmd,
		HealthCmd:  pluginProject.Control.HealthCmd,
	}
}

// PublishReleaseExternalPlugin generates release external plugin by upload-id.
// nolint: funlen,gocognit
func (m *Manager) PublishReleaseExternalPlugin(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetExternalPluginV2Upload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release external plugin, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginExternalPluginV2 {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release external plugin, invalid category")
		return fmt.Errorf("failed to publish release external plugin, invalid category, category(%s)", up.Category)
	}

	// get origin file.
	originFile, err := m.upstreamOriginExternalPluginV2.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release external plugin, failed to get file")

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release external plugin, failed to get content")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release external plugin package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release external plugin package. failed to get temp file")

		return err
	}

	detail, err := checkOriginExternalPluginPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release external plugin package. failed to check origin external plugin package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateExternalPluginPkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to generate external plugin pkg")

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
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release external plugin package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to get temp file")

				return err
			}

			if err = m.upstreamReleasePlugin.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleasePlugin.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to get temp file")

				return err
			}

			pluginInfo := &types.ReleasePlugin{
				Release: types.Release{
					Name:         pluginPkgName,
					Generation:   gen,
					Type:         types.ReleaseTypePlugin,
					Version:      pluginPkgVersion,
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
					TemplateRendererType: types.TemplateRendererTypeJinja2,
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
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to upload to upstream")

		return err
	}

	// upsert release external plugin.
	if err = m.storageRelease.UpsertManyReleasePlugin(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to upsert release external plugin")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release external plugins")

	return nil
}

type releaseExternalPluginPkg struct {
	platform     platfmt.Platform
	tempFileName string
}

// generateExternalPluginPkg generates external plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateExternalPluginPkg(nCtx contextx.IContext,
	originDetail *types.OriginExternalPluginV2PkgDetail,
	originLocalFileName string) ([]*releaseExternalPluginPkg, error) {

	// local origin external plugin.
	localOrigin, err := m.tempFileGroup.GetFile(nCtx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	pluginPkgName := originDetail.PluginPkgName

	gp := gopool.NewPool()

	result := make(map[string]*releaseExternalPluginPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create external plugin pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			origiExternalPluginFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin external plugin file: %w", err)
			}

			// set sub dir paths.
			originDetail.SubDirPaths[plat.String()][externalPluginPkgDirNameBin] = struct{}{}

			subDirPaths := conv.MapKeyToSlice(originDetail.SubDirPaths[plat.String()])
			dirs := make([]tgzWriteRuleDir, 0, len(subDirPaths))
			files := make([]tgzWriteRuleFile, 0, len(dirs))

			for _, subDirPath := range subDirPaths {
				dirPaths := filepath.SplitList(subDirPath)

				if len(dirPaths) > 0 {
					dirs = append(dirs, tgzWriteRuleDir{
						targetFilePath: dirPaths,
						targetFileMode: tgzModeDir,
					})
				}

				subFilePaths := slices.Concat(dirPaths, []string{tgzPathMatchingSegment1})
				subFileMode := int64(tgzModeFile)
				if len(dirPaths) > 0 && dirPaths[0] == externalPluginPkgDirNameBin {
					subFileMode = tgzModeExe
				}

				files = append(files, tgzWriteRuleFile{
					sourceFilePath: append([]string{convPlatToExternalPluginDirName(plat), pluginPkgName}, subFilePaths...),
					targetFilePath: subFilePaths,
					targetFileMode: subFileMode,
				})
			}

			if err = generateTgz(targetFile,
				dirs,
				[]*tgzWriteRuleStream{
					// get things from origin external plugin.
					{
						sourceFile: origiExternalPluginFile,
						fileRules:  files,
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages: %w", err)
			}

			result[plat.String()] = &releaseExternalPluginPkg{
				platform:     plat,
				tempFileName: tempFileName,
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate external plugin packages: %w", err)
	}

	return conv.MapValueToSlice(result), nil
}

const (
	originalExternalPluginDirNamePlatPrefix     = "external_plugins_"
	originalExternalPluginDirNamePlatSplitTimes = 4
	originalExternalPluginDirNameEtc            = "etc"
	originalExternalPluginFileNameEtcExt        = ".tpl"

	originalExternalPluginFileNameProject = "project.yaml"

	externalPluginPkgDirNameBin = "bin"
)

func convPlatToExternalPluginDirName(plat platfmt.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalExternalPluginDirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convExternalPluginDirNameToPlat(dirName string) platfmt.Platform {
	if !strings.HasPrefix(dirName, originalExternalPluginDirNamePlatPrefix) {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalExternalPluginDirNamePlatSplitTimes)
	if len(dirSplit) != originalExternalPluginDirNamePlatSplitTimes {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platfmt.Normalize(dirSplit[2], dirSplit[3])
	if err != nil {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
