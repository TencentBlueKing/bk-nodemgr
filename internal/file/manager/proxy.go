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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PublishReleaseProxy generates release proxy packages by upload-id.
// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseProxy(ctx context.Context, uploadID string) error {
	up, err := m.storageUpload.GetUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginServer {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginServer.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(ctx, originContent)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release proxy package. failed to save temp file. err: %v", err)

		return err
	}

	checkingFile, err := m.getTempFile(ctx, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release proxy package. failed to get temp file. err: %v", err)

		return err
	}

	detail, err := checkGSE2OriginServerPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload release proxy package. failed to check origin proxy package. err: %v", err)

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateProxyPkg(ctx, detail, originTempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to generate proxy pkg. err: %v", err)

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
				types.ReleaseTypeProxy,
				pkg.platform,
				detail.Version,
			)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to upload release proxy package, failed to format package. err: %v", err)

				return err
			}

			generatedFile, err := m.getTempFile(ctx, pkg.tempFileName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to get temp file. err: %v", err)

				return err
			}

			if err = m.upstreamReleaseProxy.Store(ctx, iface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to upload to upstream. err: %v", err)

				return err
			}

			file, err := m.upstreamReleaseProxy.GetFile(ctx, pkgName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to get temp file. err: %v", err)

				return err
			}

			releasesMap[pkg.platform.String()] = &types.Release{
				Generation: types.Generation2,
				Type:       types.ReleaseTypeProxy,
				Platform:   pkg.platform,
				Version:    detail.Version,
				FileName:   file.Info().Name,
				MD5:        file.Info().MD5,
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to upload to upstream. err: %v", err)

		return err
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertManyRelease(ctx, conv.MapToSlice(releasesMap)); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release proxy, failed to upsert release proxy: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release proxys. platforms(%v)", detail.Platforms)

	return nil
}

type releaseProxyPkg struct {
	platform     platform.Platform
	tempFileName string
}

// generateProxyPkg generates proxy package.
// nolint:funlen,fnsize,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateProxyPkg(ctx context.Context,
	originDetail *types.OriginPkgDetail,
	originLocalFileName string) ([]*releaseProxyPkg, error) {

	// local origin server.
	localOrigin, err := m.tempFileGroup.GetFile(ctx, originLocalFileName)
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

	result := make(map[string]*releaseProxyPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		// local release agent.
		localAgent, err := m.fetchReleaseAgentLocal(ctx, plat, originDetail.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to generate release proxy, failed to fetch release agent: %w", err)
		}

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(ctx)
			if err != nil {
				return fmt.Errorf("failed to create proxy pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(ctx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			originServerFile, err := localOrigin.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin proxy file: %w", err)
			}
			originCertFile, err := localCert.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin cert file: %w", err)
			}
			originBinToolFile, err := localBinTool.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open origin bintool file: %w", err)
			}
			releaseAgentFile, err := localAgent.Content(ctx)
			if err != nil {
				return fmt.Errorf("failed to open release agent file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{targetFilePath: []string{"bin"}, targetFileMode: tgzModeDir},
					{targetFilePath: []string{"cert"}, targetFileMode: tgzModeDir},
				},
				[]*tgzWriteRuleStream{
					// get things from origin server.
					{
						sourceFile: originServerFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, "server", "bin", "gse_file"},
								targetFilePath: []string{"bin", "gse_file"},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "server", "bin", "gse_data"},
								targetFilePath: []string{"bin", "gse_data"},
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
								sourceFilePath: []string{tgzPathNameAny1, "gse_server.crt"},
								targetFilePath: []string{"cert", "gse_server.crt"},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "gse_server.key"},
								targetFilePath: []string{"cert", "gse_server.key"},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "gse_api_client.crt"},
								targetFilePath: []string{"cert", "gse_api_client.crt"},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, "gse_api_client.key"},
								targetFilePath: []string{"cert", "gse_api_client.key"},
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
										return "proxy_linux_amd64"
									}

									if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
										return "proxy_linux_arm64"
									}

									return ""
								}(), tgzPathNameAny2},
								targetFilePath: []string{"bin", tgzPathNameAny2},
								targetFileMode: tgzModeExe,
							},
						},
					},
					// get things from release agent.
					{
						sourceFile: releaseAgentFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{"bin", "gse_agent"},
								targetFilePath: []string{"bin", "gse_agent"},
								targetFileMode: tgzModeExe,
							},
						},
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages: %w", err)
			}

			result[plat.String()] = &releaseProxyPkg{
				platform:     plat,
				tempFileName: tempFileName,
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate proxy packages: %w", err)
	}

	return conv.MapToSlice(result), nil
}
