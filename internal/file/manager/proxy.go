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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IProxy defines the interface for proxy.
type IProxy interface {
	// PublishReleaseProxy generates release proxy by upload-id.
	PublishReleaseProxy(nCtx contextx.IContext, uploadID string) error
}

// PublishReleaseProxy generates release proxy packages by upload-id.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseProxy(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetServerUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release proxy, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginServer {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release proxy, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginServer.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release proxy, failed to get file")

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release proxy, failed to get content. file")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release proxy package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release proxy package. failed to get temp file")

		return err
	}

	detail, err := checkGSE2OriginServerPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release proxy package. failed to check origin proxy package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateProxyPkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to generate proxy pkg")

		return err
	}

	gp := gopool.NewPool()

	releasesMap := make(map[string]*types.ReleaseProxy)
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
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release proxy package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to get temp file")

				return err
			}

			if err = m.upstreamReleaseProxy.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleaseProxy.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to get temp file")

				return err
			}

			agentRelease, err := m.storageRelease.GetReleaseAgent(nCtx, types.Generation2, pkg.platform, detail.Version)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to get agent release")

				return err
			}

			proxyRelease := &types.ReleaseProxy{
				Release: types.Release{
					Generation:   types.Generation2,
					Type:         types.ReleaseTypeProxy,
					Version:      detail.Version,
					Platform:     pkg.platform,
					Labels:       nil,
					FileName:     file.Info().Name,
					MD5:          file.Info().MD5,
					Enabled:      false,
					AsDefault:    false,
					UpdatedAt:    time.Now(),
					Operator:     nCtx.BKUsername(),
					AdditionInfo: nil,
				},
				ReleaseAdditionInfoProxy: types.ReleaseAdditionInfoProxy{
					ConfigTemplate: detail.ConfigTemplate,
					ConfigEnviron:  detail.ConfigEnviron,
					ChangeLogEN:    detail.ChangeLogEN,
					ChangeLogZH:    detail.ChangeLogZH,
				},
			}
			proxyRelease.ConfigTemplate[types.ConfigKeyAgent] = agentRelease.ConfigTemplate[types.ConfigKeyAgent]
			for k, v := range agentRelease.ConfigEnviron {
				if _, ok := proxyRelease.ConfigEnviron[k]; !ok {
					proxyRelease.ConfigEnviron[k] = v
				}
			}

			releasesMap[pkg.platform.String()] = proxyRelease

			go m.recordPublishEvent(nCtx, &proxyRelease.Release)

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to upload to upstream")

		return err
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertManyReleaseProxy(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to upsert release proxy")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release proxys")

	return nil
}

type releaseProxyPkg struct {
	platform     platfmt.Platform
	tempFileName string
}

const (
	proxyPkgDirNameBin             = "bin"
	proxyPkgDirNameCert            = "cert"
	proxyPkgFileNameFileServer     = "gse_file"
	proxyPkgFileNameDataServer     = "gse_data"
	proxyPkgFileNameCaCrt          = "gseca.crt"
	proxyPkgFileNameAgentCrt       = "gse_agent.crt"
	proxyPkgFileNameAgentKey       = "gse_agent.key"
	proxyPkgFileNameServerCrt      = "gse_server.crt"
	proxyPkgFileNameServerKey      = "gse_server.key"
	proxyPkgFileNameAPIClientCrt   = "gse_api_client.crt"
	proxyPkgFileNameAPIClientKey   = "gse_api_client.key"
	proxyPkgFileNameCertEncryptKey = "cert_encrypt.key"
)

func proxyPkgFileNameAgent(plat platfmt.Platform) string {
	if plat.OS == criteria.OSWindows {
		return "gse_agent.exe"
	}

	return "gse_agent"
}

// generateProxyPkg generates proxy package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateProxyPkg(nCtx contextx.IContext, originDetail *types.OriginPkgDetail, originLocalFileName string) (
	[]*releaseProxyPkg, error) {

	// local origin server.
	localOrigin, err := m.tempFileGroup.GetFile(nCtx, originLocalFileName)
	if err != nil {
		return nil, err
	}

	// local cert.
	localCert, err := m.fetchReleaseCertToLocal(nCtx)
	if err != nil {
		return nil, err
	}

	// local bintool.
	localBinTool, err := m.fetchReleaseBinToolToLocal(nCtx)
	if err != nil {
		return nil, err
	}

	gp := gopool.NewPool()

	result := make(map[string]*releaseProxyPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		// local release agent.
		localAgent, err := m.fetchReleaseAgentLocal(nCtx, plat, originDetail.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to generate release proxy, failed to fetch release agent: %w", err)
		}

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create proxy pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			originServerFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin proxy file: %w", err)
			}
			originCertFile, err := localCert.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin cert file: %w", err)
			}
			originBinToolFile, err := localBinTool.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin bintool file: %w", err)
			}
			releaseAgentFile, err := localAgent.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open release agent file: %w", err)
			}

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{targetFilePath: []string{proxyPkgDirNameBin}, targetFileMode: tgzModeDir},
					{targetFilePath: []string{proxyPkgDirNameCert}, targetFileMode: tgzModeDir},
				},
				[]*tgzWriteRuleStream{
					// get things from origin server.
					{
						sourceFile: originServerFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, originalServerDirNameRoot, originalServerDirNameBin,
									originalServerFileNameFileServer},
								targetFilePath: []string{proxyPkgDirNameBin, proxyPkgFileNameFileServer},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, originalServerDirNameRoot, originalServerDirNameBin,
									originalServerFileNameDataServer},
								targetFilePath: []string{proxyPkgDirNameBin, proxyPkgFileNameDataServer},
								targetFileMode: tgzModeExe,
							},
						},
					},
					// get things from origin cert.
					{
						sourceFile: originCertFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameCaCrt},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameCaCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameAgentCrt},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameAgentCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameAgentKey},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameAgentKey},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameServerCrt},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameServerCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameServerKey},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameServerKey},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameAPIClientCrt},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameAPIClientCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameAPIClientKey},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameAPIClientKey},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameCertEncryptKey},
								targetFilePath: []string{proxyPkgDirNameCert, proxyPkgFileNameCertEncryptKey},
								targetFileMode: tgzModeFile,
							},
						},
					},
					// get things from origin bintool.
					{
						sourceFile: originBinToolFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, convPlatToBinToolDirName(types.ReleaseTypeProxy, plat), tgzPathMatchingSegment2},
								targetFilePath: []string{proxyPkgDirNameBin, tgzPathMatchingSegment2},
								targetFileMode: tgzModeExe,
							},
						},
					},
					// get things from release agent.
					{
						sourceFile: releaseAgentFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{agentPkgDirNameBin, agentPkgFileNameAgent(plat)},
								targetFilePath: []string{proxyPkgDirNameBin, proxyPkgFileNameAgent(plat)},
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

	return conv.MapValueToSlice(result), nil
}
