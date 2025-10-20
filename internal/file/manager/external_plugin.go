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
	"slices"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/epluginpkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IExternalPlugin defines the interface for external plugin.
type IExternalPlugin interface {
	// UploadOriginExternalPlugin uploads the origin external plugin.
	UploadOriginExternalPlugin(nCtx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginExternalPluginPkgDetail, error)

	// PublishReleaseExternalPlugin generates release external plugin by upload-id.
	PublishReleaseExternalPlugin(nCtx contextx.IContext, uploadID string) error
}

// UploadOriginExternalPlugin uploads origin external plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginExternalPlugin(nCtx contextx.IContext, externalPluginFile io.ReadCloser) (
	*types.OriginExternalPluginPkgDetail, error) {

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
	originalPkgName, err := epluginpkg.FormatPkgName(
		detail.Name, types.ReleaseTypeOriginExternalPlugin, gen, platfmt.EmptyPlatform(), detail.Version)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package, failed to format package")

		return nil, fmt.Errorf("failed to upload origin external plugin package: %w", err)
	}

	logger.G.Biz(nCtx).With("filename", originalPkgName).Info("formatting origin external plugin package")

	originalPkgName = m.wrapOriginPackageName(originalPkgName)

	// upload to upstream.
	if err := m.upstreamOriginExternalPlugin.Store(nCtx, fileiface.FileInfo{Name: originalPkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package, failed to store to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginExternalPlugin.GetFile(nCtx, originalPkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseExternalPlugin(nCtx, detail.Name, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateExternalPluginUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginExternalPlugin,
		SavedName: originalPkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin external plugin package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	logger.G.Biz(nCtx).With("filename", originalPkgName).Info("uploaded origin external plugin package to upstream")

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
	LaunchNode   string `yaml:"launch_node"`

	Control PluginControl `yaml:"control"`

	ConfigTemplates []ExternalPluginConfigTemplate `yaml:"config_templates"`
}

// ExternalPluginConfigTemplate represents the project.yml file's config_templates field.
type ExternalPluginConfigTemplate struct {
	PluginVersion string    `yaml:"plugin_version"`
	Name          string    `yaml:"name"`
	Version       string    `yaml:"version"`
	FilePath      string    `yaml:"file_path"`
	Format        string    `yaml:"format"`
	IsMainConfig  string    `yaml:"is_main_config"`
	SourcePath    string    `yaml:"source_path"`
	Variables     *Property `yaml:"variables"`
}

/**
 * plugin files contains:
 * bin/
 * etc/
 * project.yml
 * others can be ignored.
 */
// nolint: funlen
func checkOriginExternalPluginPkg(file io.ReadCloser) (*types.OriginExternalPluginPkgDetail, error) {
	detail := types.NewOriginExternalPluginPkgDetail()

	if err := checkTgz(file, []tgzReadRule{
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

				detail.Name = pluginProject.Name
				detail.Version = pluginProject.Version
				detail.Description = pluginProject.Description
				detail.Scenario = pluginProject.Scenario
				detail.ConfigFile = pluginProject.ConfigFile
				detail.ConfigFormat = pluginProject.ConfigFormat
				detail.LaunchMode = pluginProject.LaunchNode

				if _, ok := detail.ConfigTemplates[plat.String()]; !ok {
					detail.ConfigTemplates[plat.String()] = make([]types.PluginPkgConfigTemplate, len(pluginProject.ConfigTemplates))
				}

				for idx, configTemplate := range pluginProject.ConfigTemplates {
					isMainConfig, err := conv.StringToBool(configTemplate.IsMainConfig)
					if err != nil {
						return fmt.Errorf("failed to parse is_main_config: %w", err)
					}

					detail.ConfigTemplates[plat.String()][idx] = types.PluginPkgConfigTemplate{
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

				detail.Controller[plat.String()] = types.ProcessController{
					StartCmd:   pluginProject.Control.StartCmd,
					StopCmd:    pluginProject.Control.StopCmd,
					RestartCmd: pluginProject.Control.RestartCmd,
					ReloadCmd:  pluginProject.Control.ReloadCmd,
					KillCmd:    pluginProject.Control.KillCmd,
					VersionCmd: pluginProject.Control.VersionCmd,
					HealthCmd:  pluginProject.Control.HealthCmd,
				}

				return nil
			},
		},
		{
			filePathRegex: []string{buildPrefixMatchRegex(originalOfficialPluginDirNamePlatPrefix), ".*", ".*"},
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
	}); err != nil {
		return nil, err
	}

	return detail, nil
}

// PublishReleaseExternalPlugin generates release external plugin by upload-id.
// nolint: funlen,gocognit
func (m *Manager) PublishReleaseExternalPlugin(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetExternalPluginUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release external plugin, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginExternalPlugin {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release external plugin, invalid category")
		return fmt.Errorf("failed to publish release external plugin, invalid category, category(%s)", up.Category)
	}

	// get origin file.
	originFile, err := m.upstreamOriginExternalPlugin.GetFile(nCtx, up.SavedName)
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

	releasesMap := make(map[string]*types.ReleaseExternalPlugin)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := epluginpkg.FormatPkgName(detail.Name, types.ReleaseTypeExternalPlugin, gen, pkg.platform, detail.Version)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release external plugin package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to get temp file")

				return err
			}

			if err = m.upstreamReleaseExternalPlugin.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleaseExternalPlugin.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to get temp file")

				return err
			}

			releasesMap[pkg.platform.String()] = &types.ReleaseExternalPlugin{
				Release: types.Release{
					Name:         detail.Name,
					Generation:   gen,
					Type:         types.ReleaseTypeExternalPlugin,
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
				ReleaseAdditionInfoExternalPlugin: types.ReleaseAdditionInfoExternalPlugin{
					ConfigTemplates:  detail.ConfigTemplates[pkg.platform.String()],
					PluginController: detail.Controller[pkg.platform.String()],
				},
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release external plugin, failed to upload to upstream")

		return err
	}

	// upsert release external plugin.
	if err = m.storageRelease.UpsertManyReleaseExternalPlugin(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
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
	originDetail *types.OriginExternalPluginPkgDetail,
	originLocalFileName string) ([]*releaseExternalPluginPkg, error) {

	// local origin external plugin.
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

			originPluginBinToolFile, err := localPluginBinTool.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin plugin bintool file: %w", err)
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
					sourceFilePath: append([]string{convPlatToExternalPluginDirName(plat), pluginName}, subFilePaths...),
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
					// get things from origin plugin bintool.
					{
						sourceFile: originPluginBinToolFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, convPlatToPluginBinToolDirName(plat), tgzPathMatchingSegment2},
								targetFilePath: []string{externalPluginPkgDirNameBin, tgzPathMatchingSegment2},
								targetFileMode: tgzModeExe,
							},
						},
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
