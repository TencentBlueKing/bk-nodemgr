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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAgent defines the interface for agent.
type IAgent interface {
	// UploadOriginAgent uploads the origin agent.
	UploadOriginAgent(ctx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error)

	// PublishReleaseAgent generates release agent by upload-id.
	PublishReleaseAgent(ctx contextx.IUserContext, uploadID string) error
}

// UploadOriginAgent uploads the origin agent.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginAgent(ctx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error) {
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

	detail, err := checkGSE2OriginAgentPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to check origin agent package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to get temp file. err: %v", err)

		return nil, err
	}
	gen := types.Generation2

	pkgName, err := nodepkg.FormatPkgName(
		gen,
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
	if err := m.upstreamOriginAgent.Store(ctx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
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
	existed, err := m.storageRelease.ExistReleaseAgent(ctx, gen, detail.Version, detail.Platforms...)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin agent package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateAgentUpload(ctx, &types.Upload{
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

// checkGSE2OriginAgentPkg check origin agent package.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func checkGSE2OriginAgentPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	plats := make(map[string]platform.Platform)
	detail := types.NewOriginPkgDetail()
	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, originalAgentFileNameVersion},
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
			filePath: []string{tgzPathNameAny1, originalAgentFileNameDescription},
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
			filePath: []string{tgzPathNameAny1, originalAgentFileNameDescriptionEN},
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
			filePath: []string{tgzPathNameAny1, originalAgentDirNameSupportFile, originalAgentDirNameTemplates, originalAgentFileNameConfTemplateAgent},
			callback: func(_ []string, r io.Reader) error {
				content, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_agent.conf template file. err: %w", err)
				}

				detail.ConfigTemplate[types.ConfigKeyAgent] = string(content)

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, originalAgentDirNameSupportFile, originalAgentDirNameEnv, originalAgentFileNameAgentEnv},
			callback: func(_ []string, r io.Reader) error {
				environ, err := parseEnvFile(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_agent.env file. err: %w", err)
				}

				detail.ConfigEnviron = environ

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, tgzPathNameAny2, originalAgentDirNameBin, tgzPathNameAny3},
			callback: func(path []string, _ io.Reader) error {
				plat := convAgentDirNameToPlat(path[1])

				//  only has agent file to be considered a valid agent package.
				if path[3] == originalAgentFileNameAgent(plat) {
					plats[plat.String()] = plat
				}

				return nil
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to check origin agent package. err: %w", err)
	}

	detail.Platforms = conv.MapValueToSlice(plats)

	return detail, nil
}

// PublishReleaseAgent generates release agent packages by upload-id.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseAgent(ctx contextx.IUserContext, uploadID string) error {
	up, err := m.storageUpload.GetAgentUpload(ctx, uploadID)
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

	detail, err := checkGSE2OriginAgentPkg(checkingFile)
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
	gen := types.Generation2

	releasesMap := make(map[string]*types.ReleaseAgent)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := nodepkg.FormatPkgName(
				gen,
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

			if err = m.upstreamReleaseAgent.Store(ctx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to upload to upstream. err: %v", err)

				return err
			}

			file, err := m.upstreamReleaseAgent.GetFile(ctx, pkgName)
			if err != nil {
				m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to get temp file. err: %v", err)

				return err
			}

			releasesMap[pkg.platform.String()] = &types.ReleaseAgent{
				Release: types.Release{
					Generation:   gen,
					Type:         types.ReleaseTypeAgent,
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
				ReleaseAdditionInfoAgent: types.ReleaseAdditionInfoAgent{
					ConfigTemplate: detail.ConfigTemplate,
					ConfigEnviron:  detail.ConfigEnviron,
					ChangeLogEN:    detail.ChangeLogEN,
					ChangeLogZH:    detail.ChangeLogZH,
				},
			}

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release agent, failed to upload to upstream. err: %v", err)

		return err
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertManyReleaseAgent(ctx, conv.MapValueToSlice(releasesMap)); err != nil {
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
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateAgentPkg(ctx context.Context,
	originDetail *types.OriginPkgDetail,
	originLocalFileName string) ([]*releaseAgentPkg, error) {

	// local origin agent.
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
					{targetFilePath: []string{agentPkgDirNameBin}, targetFileMode: tgzModeDir},
					{targetFilePath: []string{agentPkgDirNameCert}, targetFileMode: tgzModeDir},
				},
				[]*tgzWriteRuleStream{
					// get things from origin agent.
					{
						sourceFile: originAgentFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, convPlatToAgentDirName(plat), originalAgentDirNameBin, tgzPathNameAny2},
								targetFilePath: []string{agentPkgDirNameBin, tgzPathNameAny2},
								targetFileMode: tgzModeExe,
							},
						},
					},
					// get things from origin cert.
					{
						sourceFile: originCertFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, certFileNameCaCrt},
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameCaCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, certFileNameAgentCrt},
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameAgentCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, certFileNameAgentKey},
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameAgentKey},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathNameAny1, certFileNameCertEncryptKey},
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameCertEncryptKey},
								targetFileMode: tgzModeFile,
							},
						},
					},
					// get things from origin bintool.
					{
						sourceFile: originBinToolFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathNameAny1, convPlatToBinToolDirName(types.ReleaseTypeAgent, plat), tgzPathNameAny2},
								targetFilePath: []string{agentPkgDirNameBin, tgzPathNameAny2},
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

	return conv.MapValueToSlice(result), nil
}

const (
	originalAgentDirNameEnv            = "env"
	originalAgentDirNameSupportFile    = "support-files"
	originalAgentDirNameBin            = "bin"
	originalAgentDirNameTemplates      = "templates"
	originalAgentDirNamePlatPrefix     = "agent_"
	originalAgentDirNamePlatSplitTimes = 3

	originalAgentFileNameVersion           = "VERSION"
	originalAgentFileNameDescription       = "DESCRIPTION"
	originalAgentFileNameDescriptionEN     = "DESCRIPTION_EN"
	originalAgentFileNameConfTemplateAgent = "#etc#gse#gse_agent.conf"
	originalAgentFileNameAgentEnv          = "gse_agent.env"

	agentPkgDirNameCert = "cert"
	agentPkgDirNameBin  = "bin"

	agentPkgFileNameCaCrt          = "gseca.crt"
	agentPkgFileNameAgentCrt       = "gse_agent.crt"
	agentPkgFileNameAgentKey       = "gse_agent.key"
	agentPkgFileNameCertEncryptKey = "cert_encrypt.key"
)

// nolint: goconst
func originalAgentFileNameAgent(plat platform.Platform) string {
	if plat.OS == criteria.OSWindows {
		return "gse_agent.exe"
	}

	return "gse_agent"
}

// nolint: goconst
func agentPkgFileNameAgent(plat platform.Platform) string {
	if plat.OS == criteria.OSWindows {
		return "gse_agent.exe"
	}

	return "gse_agent"
}

func convPlatToAgentDirName(plat platform.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalAgentDirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convAgentDirNameToPlat(dirName string) platform.Platform {
	if !strings.HasPrefix(dirName, originalAgentDirNamePlatPrefix) {
		return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalAgentDirNamePlatSplitTimes)
	if len(dirSplit) != originalAgentDirNamePlatSplitTimes {
		return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platform.Normalize(dirSplit[1], dirSplit[2])
	if err != nil {
		return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
