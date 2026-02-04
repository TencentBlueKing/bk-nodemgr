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
	"encoding/binary"
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IProxy defines the interface for proxy.
type IProxy interface {
	// UploadOriginProxy uploads the origin proxy.
	UploadOriginProxy(nCtx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error)

	// PublishReleaseProxyFromProxyPkg generates release proxy from origin proxy package by upload-id.
	PublishReleaseProxyFromProxyPkg(nCtx contextx.IContext, uploadID string) error

	// PublishReleaseProxyFromLocalDir generates release proxy packages from local source directory.
	PublishReleaseProxyFromLocalDir(nCtx contextx.IContext, sourceDir string) error
}

// UploadOriginProxy uploads the origin proxy.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginProxy(nCtx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error) {
	// validation.
	if pkgFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin proxy package. file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, pkgFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package. failed to get temp file")

		return nil, err
	}

	detail, err := checkGSE2OriginProxyPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package. failed to check origin proxy package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package. failed to get temp file")

		return nil, err
	}

	// origin proxy package only have one platform.
	if len(detail.Platforms) == 0 {
		logger.G.Biz(nCtx).Error("failed to upload origin proxy package. failed to get platform")

		return nil, errors.New("failed to get platform")
	}
	plat := detail.Platforms[0]
	gen := types.Generation2
	pkgFileName, err := nodepkg.FormatPkgFileName(
		gen,
		types.ReleaseTypeOriginProxy,
		plat,
		detail.Version,
	)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package, failed to format package")

		return nil, err
	}
	pkgFileName = m.wrapOriginPackageName(pkgFileName)

	// upload to upstream.
	if err := m.upstreamOriginProxy.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package, failed to upload to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginProxy.GetFile(nCtx, pkgFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseProxy(nCtx, gen, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateProxyUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginProxy,
		SavedName: pkgFileName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin proxy package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// record event.
	m.recordUploadEvent(nCtx, types.ReleaseTypeOriginProxy, types.ReleaseNameProxy, detail.Version, detail.Platforms)

	logger.G.Biz(nCtx).With("version", detail.Version, "filename", pkgFileName).Info("uploaded origin proxy package to upstream")

	return detail, nil
}

// checkGSE2OriginProxyPkg check origin proxy package.
// nolint:funlen,gocognit,gocyclo,cyclop,lll
// NOCC: golint/fnsize(func design is not suitable for splitting).
func checkGSE2OriginProxyPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	platSet := make(map[string]struct{})
	var seenFile, seenData, seenAgent bool
	detail := types.NewOriginPkgDetail()
	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyFileNameVersion)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyFileNameDescription)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read description file: %w", err)
				}

				detail.ChangeLogZH = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyFileNameDescriptionEN)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read description-en file: %w", err)
				}

				detail.ChangeLogEN = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameTemplates), buildFullMatchRegex(originalProxyFileNameConfTemplateDataProxyTypeOne)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameTemplates), buildFullMatchRegex(originalProxyFileNameConfTemplateDataProxyTypeTwo)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameTemplates), buildFullMatchRegex(originalProxyFileNameConfTemplateFileProxyTypeOne)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameTemplates), buildFullMatchRegex(originalProxyFileNameConfTemplateFileProxyTypeTwo)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_file_proxy.conf.template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyFile] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameTemplates), buildFullMatchRegex(originalProxyFileNameConfTemplateAgentTypeOne)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read #etc#gse#gse_agent.conf template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyAgent] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameTemplates), buildFullMatchRegex(originalProxyFileNameConfTemplateAgentTypeTwo)},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_agent.conf.template file: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyAgent] = string(content)

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameSupportFile), buildFullMatchRegex(originalProxyDirNameEnv), buildFullMatchRegex(originalProxyFileNameProxyEnv)},
			callback: func(_ []string, r io.Reader) error {
				environ, err := parseEnvFile(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_proxy.env file: %w", err)
				}

				detail.ConfigEnviron = environ

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameRoot), buildFullMatchRegex(originalProxyDirNameBin), buildFullMatchRegex(originalProxyFileNameAgent)},
			callback: func(_ []string, r io.Reader) error {
				seenAgent = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check gse_agent binary platform: %w", err)
				}

				platSet[plat.String()] = struct{}{}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameRoot), buildFullMatchRegex(originalProxyDirNameBin), buildFullMatchRegex(originalProxyFileNameData)},
			callback: func(_ []string, r io.Reader) error {
				seenData = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check gse_data binary platform: %w", err)
				}

				platSet[plat.String()] = struct{}{}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalProxyDirNameRoot), buildFullMatchRegex(originalProxyDirNameBin), buildFullMatchRegex(originalProxyFileNameFile)},
			callback: func(_ []string, r io.Reader) error {
				seenFile = true
				plat, err := checkServerBinaryPlatform(r)
				if err != nil {
					return fmt.Errorf("failed to check gse_file binary platform: %w", err)
				}

				platSet[plat.String()] = struct{}{}

				if len(detail.Platforms) == 0 {
					detail.Platforms = append(detail.Platforms, *plat)
				}

				return nil
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to check origin proxy package: %w", err)
	}

	if !seenFile || !seenData || !seenAgent || len(platSet) != 1 || detail.Version == "" {
		return nil, fmt.Errorf("invalid origin proxy package. gse-file(%t) gse-data(%t) gse-agent(%t) version(%s) platform(%v) platform-set-size(%d)",
			seenFile, seenData, seenAgent, detail.Version, detail.Platforms, len(platSet))
	}

	return detail, nil
}

// PublishReleaseProxyFromProxyPkg generates release proxy packages from origin proxy package by upload-id.
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseProxyFromProxyPkg(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetProxyUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release proxy, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginProxy {
		logger.G.Biz(nCtx).With("category", up.Category).Error("failed to publish release proxy, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginProxy.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release proxy, failed to get file")

		return err
	}

	if err := m.handleProxyPkg(nCtx, originFile); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release proxy, failed to handle proxy pkg")

		return err
	}

	logger.G.Biz(nCtx).With("upload-id", uploadID).Info("successfully published release proxy")

	return nil
}

// PublishReleaseProxyFromLocalDir generates release proxy packages from local source directory.
// nolint: gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseProxyFromLocalDir(nCtx contextx.IContext, sourceDir string) error {
	localFileGroup, err := local.NewLocalDir(sourceDir)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("source-dir", sourceDir).
			Error("failed to publish local release proxy package, failed to create local file group")

		return err
	}

	files, err := localFileGroup.AllFiles(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("source-dir", sourceDir).
			Error("failed to publish local release proxy package, failed to list local files")

		return err
	}

	if len(files) == 0 {
		logger.G.Biz(nCtx).With("source-dir", sourceDir).
			Info("no proxy package files found in source dir, skip publishing")

		return nil
	}

	gp := gopool.NewPool()
	for _, file := range files {
		f := file
		if !strings.HasSuffix(f.Info().Name, ".tgz") &&
			!strings.HasSuffix(f.Info().Name, ".gz") &&
			!strings.HasSuffix(f.Info().Name, ".tar") {

			logger.G.Biz(nCtx).With("filename", f.Info().Name).Info("skip non-tgz file for publishing local release proxy package")

			continue
		}

		logger.G.Biz(nCtx).With("filename", f.Info().Name).Info("start to publish local release proxy package")

		gp.Go(func() error {
			logger.G.Biz(nCtx).With("filename", f.Info().Name).Info("publishing local release proxy package")

			if err := m.handleProxyPkg(nCtx, f); err != nil {
				logger.G.Biz(nCtx).WithErr(err).
					With("filename", f.Info().Name).
					Error("failed to publish local release proxy package, failed to handle proxy pkg")

				return err
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("source-dir", sourceDir).
			Error("failed to publish local release proxy package, failed to handle proxy pkgs")

		return err
	}

	logger.G.Biz(nCtx).With("source-dir", sourceDir).Info("successfully published local release proxy packages")

	return nil
}

// nolint: funlen
func (m *Manager) handleProxyPkg(nCtx contextx.IContext, pkg fileiface.File) error {
	// get origin content.
	content, err := pkg.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", pkg.Info().Name).Error("failed to publish release proxy, failed to get content. file")

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release proxy package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release proxy package. failed to get temp file")

		return err
	}

	detail, err := checkGSE2OriginProxyPkg(checkingFile)
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
			pkgFileName, err := nodepkg.FormatPkgFileName(
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

			if err = m.upstreamReleaseProxy.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleaseProxy.GetFile(nCtx, pkgFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release proxy, failed to get temp file")

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

			releasesMap[pkg.platform.String()] = proxyRelease

			// record publish event.
			m.recordPublishEvent(nCtx, &proxyRelease.Release)

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
	originalProxyDirNameRoot        = "server"
	originalProxyDirNameEnv         = "env"
	originalProxyDirNameSupportFile = "support-files"
	originalProxyDirNameBin         = "bin"
	originalProxyDirNameTemplates   = "templates"

	originalProxyFileNameVersion       = "VERSION"
	originalProxyFileNameDescription   = "DESCRIPTION"
	originalProxyFileNameDescriptionEN = "DESCRIPTION_EN"
	originalProxyFileNameProxyEnv      = "gse_proxy.env"

	originalProxyFileNameAgent                    = "gse_agent"
	originalProxyFileNameConfTemplateAgentTypeOne = "#etc#gse#gse_agent.conf"
	originalProxyFileNameConfTemplateAgentTypeTwo = "gse_agent.conf.template"

	originalProxyFileNameFile                         = "gse_file"
	originalProxyFileNameConfTemplateFileProxyTypeOne = "#etc#gse#gse_file_proxy.conf"
	originalProxyFileNameConfTemplateFileProxyTypeTwo = "gse_file_proxy.conf.template"

	originalProxyFileNameData                         = "gse_data"
	originalProxyFileNameConfTemplateDataProxyTypeOne = "#etc#gse#gse_data_proxy.conf"
	originalProxyFileNameConfTemplateDataProxyTypeTwo = "gse_data_proxy.conf.template"

	proxyPkgDirNameBin  = "bin"
	proxyPkgDirNameCert = "cert"

	proxyPkgFileNameCaCrt          = "gseca.crt"
	proxyPkgFileNameAgentCrt       = "gse_agent.crt"
	proxyPkgFileNameAgentKey       = "gse_agent.key"
	proxyPkgFileNameServerCrt      = "gse_server.crt"
	proxyPkgFileNameServerKey      = "gse_server.key"
	proxyPkgFileNameAPIClientCrt   = "gse_api_client.crt"
	proxyPkgFileNameAPIClientKey   = "gse_api_client.key"
	proxyPkgFileNameCertEncryptKey = "cert_encrypt.key"

	proxyPkgFileNameAgent = "gse_agent"
	proxyPkgFileNameFile  = "gse_file"
	proxyPkgFileNameData  = "gse_data"
)

// generateProxyPkg generates proxy package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateProxyPkg(nCtx contextx.IContext, originDetail *types.OriginPkgDetail, originLocalFileName string) (
	[]*releaseProxyPkg, error) {

	// local origin proxy.
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
			originProxyFile, err := localOrigin.Content(nCtx)
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

			if err = generateTgz(targetFile,
				[]tgzWriteRuleDir{
					{targetFilePath: []string{proxyPkgDirNameBin}, targetFileMode: tgzModeDir},
					{targetFilePath: []string{proxyPkgDirNameCert}, targetFileMode: tgzModeDir},
				},
				[]*tgzWriteRuleStream{
					// get things from origin proxy.
					{
						sourceFile: originProxyFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, originalProxyDirNameRoot, originalProxyDirNameBin,
									originalProxyFileNameFile},
								targetFilePath: []string{proxyPkgDirNameBin, platfmt.FormatBinaryFileName(proxyPkgFileNameFile, plat.OS)},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, originalProxyDirNameRoot, originalProxyDirNameBin,
									originalProxyFileNameData},
								targetFilePath: []string{proxyPkgDirNameBin, platfmt.FormatBinaryFileName(proxyPkgFileNameData, plat.OS)},
								targetFileMode: tgzModeExe,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, originalProxyDirNameRoot, originalProxyDirNameBin,
									originalProxyFileNameAgent},
								targetFilePath: []string{proxyPkgDirNameBin, platfmt.FormatBinaryFileName(proxyPkgFileNameAgent, plat.OS)},
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

const (
	elfMagic   = "\x7FELF"
	machineAMD = 0x3E // EM_X86_64
	machineARM = 0xB7 // EM_AARCH64
)

func checkServerBinaryPlatform(r io.Reader) (*platfmt.Platform, error) {
	// read the ELF header.
	buf := make([]byte, 20) // nolint:mnd
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	// check ELF magic number.
	if string(buf[:4]) != elfMagic { // nolint:mnd
		return nil, errors.New("not an ELF file")
	}

	// check ELFCLASS64
	if buf[4] != 2 { // nolint:mnd
		return nil, errors.New("not a 64-bit ELF file")
	}

	// check ELFDATA2LSB
	if buf[5] != 1 { // nolint:mnd
		return nil, errors.New("not little-endian ELF file")
	}

	// check the machine type
	machine := binary.LittleEndian.Uint16(buf[18:20]) // nolint:mnd

	switch machine {
	case machineAMD:
		return &platfmt.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}, nil
	case machineARM:
		return &platfmt.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchArm64}, nil
	default:
		return nil, fmt.Errorf("unsupported architecture: 0x%X", machine)
	}
}
