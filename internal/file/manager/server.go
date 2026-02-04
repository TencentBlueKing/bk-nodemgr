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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IServer defines the interface for server.
type IServer interface {
	// UploadOriginServer uploads the origin server.
	UploadOriginServer(nCtx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error)

	// PublishReleaseProxyFromServerPkg generates release proxy packages from server packages by upload-id.
	PublishReleaseProxyFromServerPkg(nCtx contextx.IContext, uploadID string) error

	// PublishReleaseServerFromLocalDir generates release server packages from local source directory.
	PublishReleaseServerFromLocalDir(nCtx contextx.IContext, sourceDir string) error
}

// UploadOriginServer uploads the origin server.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginServer(nCtx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error) {
	if pkgFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin server package. file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, pkgFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package. failed to get temp file")

		return nil, err
	}

	detail, err := checkGSE2OriginServerPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package. failed to check origin server package")

		return nil, err
	}

	// origin server package only have one platform.
	if len(detail.Platforms) == 0 {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package. failed to get platform")

		return nil, errors.New("failed to get platform")
	}
	plat := detail.Platforms[0]

	pkgFileName, err := nodepkg.FormatPkgFileName(
		types.Generation2,
		types.ReleaseTypeOriginServer,
		plat,
		detail.Version,
	)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package, failed to format package")

		return nil, err
	}
	pkgFileName = m.wrapOriginPackageName(pkgFileName)

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package. failed to get temp file")

		return nil, err
	}

	// upload to upstream.
	if err := m.upstreamOriginServer.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package, failed to upload to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginServer.GetFile(nCtx, pkgFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	uploadID, err := m.storageUpload.CreateServerUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginServer,
		SavedName: pkgFileName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin server package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// record event.
	m.recordUploadEvent(nCtx, types.ReleaseTypeOriginServer, pkgFileName, detail.Version, detail.Platforms)

	logger.G.Biz(nCtx).With("platform", plat, "version", detail.Version, "filename", pkgFileName).Info("uploaded origin server package to upstream")

	return detail, nil
}

const (
	originalServerDirNameRoot        = "server"
	originalServerDirNameBin         = "bin"
	originalServerDirNameEnv         = "env"
	originalServerDirNameSupportFile = "support-files"
	originalServerDirNameTemplate    = "templates"

	originalServerFileNameVersion                       = "VERSION"
	originalServerFileNameProxyEnv                      = "gse_proxy.env"
	originalServerFileNameFileServer                    = "gse_file"
	originalServerFileNameDataServer                    = "gse_data"
	originalServerFileNameConfTemplateFileServerTypeOne = "#etc#gse#gse_file_proxy.conf"
	originalServerFileNameConfTemplateDataServerTypeOne = "#etc#gse#gse_data_proxy.conf"
	originalServerFileNameConfTemplateFileServerTypeTwo = "gse_file_proxy.conf.template"
	originalServerFileNameConfTemplateDataServerTypeTwo = "gse_data_proxy.conf.template"
)

// checkGSE2OriginServerPkg check gse2 origin server package.
// nolint: gocognit,gocyclo,cyclop,funlen
// NOCC: golint/gocyclo,cyclop (this function should be complex).
func checkGSE2OriginServerPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	detail := types.NewOriginPkgDetail()
	var seenFile, seenData bool
	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerFileNameVersion)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read version file: %w", err)
				}

				detail.Version = strings.Trim(string(content), "\n\r\t ")

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameRoot),
				buildFullMatchRegex(originalServerDirNameBin), buildFullMatchRegex(originalServerFileNameFileServer)},
			callback: func(_ []string, r io.Reader) error {
				seenFile = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check server binary platform: %w", err)
				}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameRoot),
				buildFullMatchRegex(originalServerDirNameBin), buildFullMatchRegex(originalServerFileNameDataServer)},
			callback: func(_ []string, r io.Reader) error {
				seenData = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check server binary platform: %w", err)
				}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameSupportFile),
				buildFullMatchRegex(originalServerDirNameTemplate),
				buildFullMatchRegex(originalServerFileNameConfTemplateFileServerTypeOne)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read #etc#gse#gse_file_proxy.conf template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyFile] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameSupportFile),
				buildFullMatchRegex(originalServerDirNameTemplate), buildFullMatchRegex(originalServerFileNameConfTemplateDataServerTypeOne)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read #etc#gse#gse_data_proxy.conf template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyData] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameSupportFile),
				buildFullMatchRegex(originalServerDirNameTemplate), buildFullMatchRegex(originalServerFileNameConfTemplateFileServerTypeTwo)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_file_proxy.conf.template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyData] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameSupportFile),
				buildFullMatchRegex(originalServerDirNameTemplate), buildFullMatchRegex(originalServerFileNameConfTemplateDataServerTypeTwo)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_data_proxy.conf.template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyData] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalServerDirNameSupportFile),
				buildFullMatchRegex(originalServerDirNameEnv), buildFullMatchRegex(originalServerFileNameProxyEnv)},
			callback: func(_ []string, r io.Reader) error {
				environ, err := parseEnvFile(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_proxy.env file: %w", err)
				}

				detail.ConfigEnviron = environ

				return nil
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to check origin server package: %w", err)
	}

	if !seenFile || !seenData || detail.Version == "" {
		return nil, fmt.Errorf("invalid origin server package. gse-file(%t) gse-data(%t) version(%s) platform(%v)",
			seenFile, seenData, detail.Version, detail.Platforms)
	}

	return detail, nil
}

// PublishReleaseProxyFromServerPkg generates release proxy from server packages by upload-id.
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseProxyFromServerPkg(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetServerUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release server, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginServer {
		logger.G.Biz(nCtx).With("category", up.Category).Error("failed to publish release server, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginServer.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release server, failed to get file")

		return err
	}

	if err := m.handleServerPkg(nCtx, originFile); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release server, failed to handle server pkg")

		return err
	}

	return nil
}

// PublishReleaseServerFromLocalDir generates release server from local directory.
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseServerFromLocalDir(nCtx contextx.IContext, sourceDir string) error {
	localGroup, err := local.NewLocalDir(sourceDir)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to create local group")

		return err
	}

	files, err := localGroup.AllFiles(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to get all local files")

		return err
	}

	if len(files) == 0 {
		logger.G.Biz(nCtx).With("source-dir", sourceDir).
			Info("no server package files found in source dir, skip publishing")

		return nil
	}

	gp := gopool.NewPool()
	for _, file := range files {
		f := file
		if !strings.HasSuffix(f.Info().Name, ".tgz") &&
			!strings.HasSuffix(f.Info().Name, ".gz") &&
			!strings.HasSuffix(f.Info().Name, ".tar") {

			logger.G.Biz(nCtx).With("filename", f.Info().Name).Info("skip non-tgz file for publishing local release server package")

			continue
		}

		logger.G.Biz(nCtx).With("filename", f.Info().Name).Info("start to publish local release server package")

		gp.Go(func() error {
			logger.G.Biz(nCtx).With("filename", f.Info().Name).Info("publishing local release server package")

			if err := m.handleServerPkg(nCtx, f); err != nil {
				logger.G.Biz(nCtx).WithErr(err).With("filename", f.Info().Name).Error("failed to publish release server, failed to handle server pkg")

				return err
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to handle server pkgs")

		return err
	}

	logger.G.Biz(nCtx).With("source-dir", sourceDir).Info("successfully published release server packages")

	return nil
}

// nolint: funlen, gocognit,cyclop
func (m *Manager) handleServerPkg(nCtx contextx.IContext, pkg fileiface.File) error {
	// get origin content.
	content, err := pkg.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", pkg.Info().Name).Error("failed to publish release server, failed to get content. file")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release server package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release server package. failed to get temp file")

		return err
	}

	detail, err := checkGSE2OriginServerPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release server package. failed to check origin server package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateProxyPkgByServer(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to generate server pkg")

		return err
	}

	gp := gopool.NewPool()

	releasesMap := make(map[string]*types.ReleaseProxy)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgFileName, err := nodepkg.FormatPkgFileName(
				types.Generation2,
				types.ReleaseTypeProxy,
				pkg.platform,
				detail.Version,
			)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release server package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to get temp file")

				return err
			}

			if err = m.upstreamReleaseProxy.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleaseProxy.GetFile(nCtx, pkgFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to get temp file")

				return err
			}

			agentRelease, err := m.storageRelease.GetReleaseAgent(nCtx, types.Generation2, pkg.platform, detail.Version)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to get agent release")

				return err
			}

			proxyRelease := &types.ReleaseProxy{
				Release: types.Release{
					Name:         types.ReleaseNameProxy,
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

			// record publish event.
			m.recordPublishEvent(nCtx, &proxyRelease.Release)

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to upload to upstream")

		return err
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertManyReleaseProxy(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release server, failed to upsert release server")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release servers")

	return nil
}

// generateProxyPkg generates proxy package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateProxyPkgByServer(nCtx contextx.IContext, originDetail *types.OriginPkgDetail, originLocalFileName string) (
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
								targetFilePath: []string{proxyPkgDirNameBin, platfmt.FormatBinaryFileName(proxyPkgFileNameFile, plat.OS)},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, originalServerDirNameRoot, originalServerDirNameBin,
									originalServerFileNameDataServer},
								targetFilePath: []string{proxyPkgDirNameBin, platfmt.FormatBinaryFileName(proxyPkgFileNameData, plat.OS)},
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
								sourceFilePath: []string{agentPkgDirNameBin, platfmt.FormatBinaryFileName(originalAgentFileNameAgent, plat.OS)},
								targetFilePath: []string{proxyPkgDirNameBin, platfmt.FormatBinaryFileName(proxyPkgFileNameAgent, plat.OS)},
								targetFileMode: tgzModeExe,
							},
						},
					},
				},
			); err != nil {
				return fmt.Errorf("failed to generate tgz from origin packages by server: %w", err)
			}

			result[plat.String()] = &releaseProxyPkg{
				platform:     plat,
				tempFileName: tempFileName,
			}

			return nil
		})
	}

	if err = gp.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate proxy packages by server: %w", err)
	}

	return conv.MapValueToSlice(result), nil
}
