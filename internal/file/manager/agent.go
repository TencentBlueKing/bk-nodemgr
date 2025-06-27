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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UploadOriginAgent uploads the origin agent.
// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginAgent(ctx context.Context, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error) {
	// validation.
	if pkgFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, pkgFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkGen2OriginAgentPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to check origin agent package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get temp file. err: %v", err)

		return nil, err
	}

	pkgName, err := nodepkg.FormatPkgName(
		types.Generation2,
		types.ReleaseTypeOriginAgent,
		platform.EmptyPlatform(),
		detail.Version,
	)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package, failed to format package. err: %v", err)

		return nil, err
	}
	pkgName = m.wrapOriginPackageName(pkgName)

	// upload to upstream.
	if err := m.upstreamOriginAgent.Store(ctx, iface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package, failed to upload to upstream. err: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginAgent.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseAgent(ctx, detail.Version, detail.Platforms...)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginAgent,
		SavedName: pkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx,
		"uploaded origin agent package to upstream. version(%s), file-name(%s)", detail.Version, pkgName)

	return detail, nil
}

// checkGen2OriginAgentPkg check gen2 origin agent package.
// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func checkGen2OriginAgentPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	plats := make(map[string]platform.Platform)
	detail := new(types.OriginPkgDetail)
	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, "VERSION"},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read version file. err: %w", err)
				}

				detail.Version = strings.Trim(string(content), "\n\r\t ")

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "DESCRIPTION"},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read description file. err: %w", err)
				}

				detail.ChangeLogZH = string(content)

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "DESCRIPTION"},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read description-en file. err: %w", err)
				}

				detail.ChangeLogEN = string(content)

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "agent_linux_x86_64", "bin", "gse_agent"},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}
				plats[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "agent_linux_aarch64", "bin", "gse_agent"},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchArm64}
				plats[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "agent_windows_x86_64", "bin", "gse_agent.exe"},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSWindows, Arch: criteria.CPUArchAmd64}
				plats[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "agent_darwin_x86_64", "bin", "gse_agent"},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSDarwin, Arch: criteria.CPUArchAmd64}
				plats[plat.String()] = plat

				return nil
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to check origin agent package. err: %w", err)
	}

	detail.Platforms = conv.MapToSlice(plats)

	return detail, nil
}

// PublishReleaseAgent generates release agent packages by upload-id.
// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseAgent(ctx context.Context, uploadID string) error {
	up, err := m.storageUpload.GetUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginAgent {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginAgent.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(ctx, originContent)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release agent package. failed to save temp file. err: %v", err)

		return err
	}

	checkingFile, err := m.getTempFile(ctx, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release agent package. failed to get temp file. err: %v", err)

		return err
	}

	detail, err := checkGen2OriginAgentPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release agent package. failed to check origin agent package. err: %v", err)

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateAgentPkg(ctx, detail, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to generate agent pkg. err: %v", err)

		return err
	}

	gp := gopool.NewPool()

	releasesMap := make(map[string]*types.Release)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := nodepkg.FormatPkgName(
				types.Generation2,
				types.ReleaseTypeAgent,
				pkg.platform,
				detail.Version,
			)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to upload release agent package, failed to format package. err: %v", err)

				return err
			}

			generatedFile, err := m.getTempFile(ctx, pkg.tempFileName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to get temp file. err: %v", err)

				return err
			}

			if err = m.upstreamReleaseAgent.Store(ctx, iface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to upload to upstream. err: %v", err)

				return err
			}

			file, err := m.temp.GetFile(ctx, pkg.tempFileName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to get temp file. err: %v", err)

				return err
			}

			releasesMap[pkg.platform.String()] = &types.Release{
				Generation: types.Generation2,
				Type:       types.ReleaseTypeAgent,
				Platform:   pkg.platform,
				Version:    detail.Version,
				FileName:   pkgName,
				MD5:        file.Info().MD5,
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to upload to upstream. err: %v", err)

		return err
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertManyRelease(ctx, conv.MapToSlice(releasesMap)); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to upsert release agent: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release agents. platforms(%v)", detail.Platforms)

	return nil
}

type releaseAgentPkg struct {
	platform     platform.Platform
	tempFileName string
}

// generateAgentPkg generates agent package.
// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateAgentPkg(ctx context.Context,
	originDetail *types.OriginPkgDetail,
	originLocalFileName string) ([]*releaseAgentPkg, error) {

	// local origin agent.
	localOrigin, err := m.temp.GetFile(ctx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	// local cert.
	localCert, err := m.fetchReleaseCertToLocal(ctx)
	if err != nil {
		return nil, err
	}

	// local bintool.
	localBinTool, err := m.fetchReleaseBinToolToLocal(ctx)
	if err != nil {
		return nil, err
	}

	gp := gopool.NewPool()

	result := make(map[string]*releaseAgentPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(ctx)
			if err != nil {
				return fmt.Errorf("failed to create agent pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(ctx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			originAgentFile, err := localOrigin.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin agent file: %w", err)
			}
			originCertFile, err := localCert.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin cert file: %w", err)
			}
			originBinToolFile, err := localBinTool.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin bintool file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{targetFilePath: []string{"bin"}, targetFileMode: tgzModeDir},
					{targetFilePath: []string{"cert"}, targetFileMode: tgzModeDir},
				},
				[]*tgzWriteRuleStream{
					// get things from origin agent.
					{
						sourceFile: originAgentFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, func() string {
									if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
										return "agent_linux_x86_64"
									}

									if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
										return "agent_linux_aarch64"
									}

									if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
										return "agent_darwin_x86_64"
									}

									if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
										return "agent_windows_x86_64"
									}

									return ""
								}(), "bin", tgzPathNameAny2},
								targetFilePath: []string{"bin", tgzPathNameAny2},
								targetFileMode: tgzModeExe,
							},
						},
					},
					// get things from origin cert.
					{
						sourceFile: originCertFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, "gseca.crt"},
								targetFilePath: []string{"cert", "gseca.crt"},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "gse_agent.crt"},
								targetFilePath: []string{"cert", "gse_agent.crt"},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "gse_agent.key"},
								targetFilePath: []string{"cert", "gse_agent.key"},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "cert_encrypt.key"},
								targetFilePath: []string{"cert", "cert_encrypt.key"},
								targetFileMode: tgzModeFile,
							},
						},
					},
					// get things from origin bintool.
					{
						sourceFile: originBinToolFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, func() string {
									if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
										return "agent_linux_amd64"
									}

									if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
										return "agent_linux_arm64"
									}

									if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
										return "agent_darwin_amd64"
									}

									if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
										return "agent_windows_amd64"
									}

									return ""
								}(), tgzPathNameAny2},
								targetFilePath: []string{"bin", tgzPathNameAny2},
								targetFileMode: tgzModeExe,
							},
						},
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages: %w", err)
			}

			result[plat.String()] = &releaseAgentPkg{
				platform:     plat,
				tempFileName: tempFileName,
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate agent packages: %w", err)
	}

	return conv.MapToSlice(result), nil
}
