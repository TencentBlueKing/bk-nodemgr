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
	"path/filepath"
	"slices"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/epluginpkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

// IExternalPlugin defines the interface for external plugin.
type IExternalPlugin interface {
	// UploadOriginExternalPlugin uploads the origin external plugin.
	UploadOriginExternalPlugin(ctx contextx.IContext, pluginFile io.ReadCloser) (*types.OriginExternalPluginPkgDetail, error)

	// PublishReleaseExternalPlugin generates release external plugin by upload-id.
	PublishReleaseExternalPlugin(ctx contextx.IUserContext, uploadID string) error
}

// UploadOriginExternalPlugin uploads origin external plugin.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginExternalPlugin(ctx contextx.IContext, externalPluginFile io.ReadCloser) (
	*types.OriginExternalPluginPkgDetail, error) {

	if externalPluginFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package, file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, externalPluginFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkOriginExternalPluginPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx,
			"failed to upload origin external plugin package. failed to check origin external plugin package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package. failed to get temp file. err: %v", err)

		return nil, err
	}

	gen := types.Generation2
	originalPkgName, err := epluginpkg.FormatPkgName(detail.Name, types.ReleaseTypeOriginExternalPlugin, gen, platform.EmptyPlatform(), detail.Version)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package, failed to format package. err: %v", err)

		return nil, fmt.Errorf("failed to upload origin external plugin package: %w", err)
	}

	m.logger.InfoCtxf(ctx, "formatting origin external plugin package: %s", originalPkgName)

	originalPkgName = m.wrapOriginPackageName(originalPkgName)

	// upload to upstream.
	if err := m.upstreamOriginExternalPlugin.Store(ctx, fileiface.FileInfo{Name: originalPkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package, failed to store to upstream: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginExternalPlugin.GetFile(ctx, originalPkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseExternalPluginGen2(ctx, detail.Name, detail.Version, detail.Platforms...)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateExternalPluginUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginExternalPlugin,
		SavedName: originalPkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin external plugin package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx, "uploaded origin external plugin package to upstream. file-name(%s)", originalPkgName)

	return detail, nil
}

// ExternalPluginProject represents the project.yml file.
type ExternalPluginProject struct {
	Name            string                         `yaml:"name"`
	Version         string                         `yaml:"version"`
	Description     string                         `yaml:"description"`
	Scenario        string                         `yaml:"scenario"`
	ConfigFile      string                         `yaml:"config_file"`
	ConfigFormat    string                         `yaml:"config_format"`
	LaunchMode      string                         `yaml:"launch_mode"`
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
func checkOriginExternalPluginPkg(file io.ReadCloser) (*types.OriginExternalPluginPkgDetail, error) {
	detail := new(types.OriginExternalPluginPkgDetail)
	detail.SubDirPaths = make(map[string]map[string]struct{})

	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, "project.yaml"},
			callback: func(path []string, projectFile io.Reader) error {
				detail.Platforms = append(detail.Platforms, convExternalPluginDirNameToPlat(path[0]))

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
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, tgzPathNameAny3},
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
func (m *Manager) PublishReleaseExternalPlugin(ctx contextx.IUserContext, uploadID string) error {
	up, err := m.storageUpload.GetExternalPluginUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginExternalPlugin {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, invalid category, category(%s)", up.Category)

		return fmt.Errorf("failed to publish release external plugin, invalid category, category(%s)", up.Category)
	}

	// get origin file.
	originFile, err := m.upstreamOriginExternalPlugin.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(ctx, originContent)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release external plugin package. failed to save temp file. err: %v", err)

		return err
	}

	checkingFile, err := m.getTempFile(ctx, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release external plugin package. failed to get temp file. err: %v", err)

		return err
	}

	detail, err := checkOriginExternalPluginPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release external plugin package. failed to check origin external plugin package. err: %v", err)

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateExternalPluginPkg(ctx, detail, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to generate external plugin pkg. err: %v", err)

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
				m.logger.ErrorCtxf(ctx, "failed to upload release external plugin package, failed to format package. err: %v", err)

				return err
			}

			generatedFile, err := m.getTempFile(ctx, pkg.tempFileName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to get temp file. err: %v", err)

				return err
			}

			if err = m.upstreamReleaseExternalPlugin.Store(ctx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to upload to upstream. err: %v", err)

				return err
			}

			file, err := m.upstreamReleaseExternalPlugin.GetFile(ctx, pkgName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to get temp file. err: %v", err)

				return err
			}

			releasesMap[pkg.platform.String()] = &types.ReleaseExternalPlugin{
				Release: types.Release{
					Generation:   gen,
					Type:         types.ReleaseTypeExternalPlugin,
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
				ReleaseAdditionInfoExternalPlugin: types.ReleaseAdditionInfoExternalPlugin{
					ConfigTemplates: detail.ConfigTemplates,
				},
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to upload to upstream. err: %v", err)

		return err
	}

	// upsert release external plugin.
	if err = m.storageRelease.UpsertManyReleaseExternalPluginGen2(ctx, conv.MapValueToSlice(releasesMap)); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release external plugin, failed to upsert release external plugin: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release external plugins. platforms(%v)", detail.Platforms)

	return nil
}

type releaseExternalPluginPkg struct {
	platform     platform.Platform
	tempFileName string
}

// generateExternalPluginPkg generates external plugin package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateExternalPluginPkg(ctx context.Context,
	originDetail *types.OriginExternalPluginPkgDetail,
	originLocalFileName string) ([]*releaseExternalPluginPkg, error) {

	// local origin external plugin.
	localOrigin, err := m.tempFileGroup.GetFile(ctx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	// local plugin bintool.
	localPluginBinTool, err := m.fetchReleasePluginBinToolToLocal(ctx)
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
			tempFileName, err := m.createTempFile(ctx)
			if err != nil {
				return fmt.Errorf("failed to create external plugin pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(ctx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			origiExternalPluginFile, err := localOrigin.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin external plugin file: %w", err)
			}

			originPluginBinToolFile, err := localPluginBinTool.Content(ctx)
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

				subFilePaths := slices.Concat(dirPaths, []string{tgzPathNameAny1})
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
								sourceFilePath: []string{tgzPathNameAny1, convPlatToPluginBinToolDirName(plat), tgzPathNameAny2},
								targetFilePath: []string{externalPluginPkgDirNameBin, tgzPathNameAny2},
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
	originalExternalPluginPlatLinuxAmd64   = "external_plugins_linux_x86_64"
	originalExternalPluginPlatLinuxArm64   = "external_plugins_linux_aarch64"
	originalExternalPluginPlatWindowsAmd64 = "external_plugins_windows_x86_64"
	originalExternalPluginPlatDarwinAmd64  = "external_plugins_darwin_x86_64"

	externalPluginPkgDirNameBin = "bin"
)

func convPlatToExternalPluginDirName(plat platform.Platform) string {
	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
		return originalExternalPluginPlatLinuxAmd64
	}

	if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
		return originalExternalPluginPlatLinuxArm64
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
		return originalExternalPluginPlatDarwinAmd64
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
		return originalExternalPluginPlatWindowsAmd64
	}

	return ""
}

func convExternalPluginDirNameToPlat(dirName string) platform.Platform {
	switch dirName {
	case originalExternalPluginPlatLinuxAmd64:
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case originalExternalPluginPlatLinuxArm64:
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	case originalExternalPluginPlatWindowsAmd64:
		{
			return platform.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
		}
	case originalExternalPluginPlatDarwinAmd64:
		{
			return platform.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64)
		}
	default:
		{
			return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
