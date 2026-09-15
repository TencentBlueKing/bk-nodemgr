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
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/nodepkg"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/goasync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IAgent defines the interface for agent.
type IAgent interface {
	IReleaseAgent

	// UploadOriginAgent uploads the origin agent.
	UploadOriginAgent(nCtx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error)

	// PublishReleaseAgent generates release agent by upload-id and is-shared.
	PublishReleaseAgent(nCtx contextx.IContext, uploadID string, isShared bool) error
}

// UploadOriginAgent uploads the origin agent.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginAgent(nCtx contextx.IContext, pkgFile io.ReadCloser) (*types.OriginPkgDetail, error) {
	// validation.
	if pkgFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin agent package. file is nil")

		return nil, errors.New("file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, pkgFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package. failed to get temp file")

		return nil, err
	}

	detail, err := checkGSE2OriginAgentPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package. failed to check origin agent package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package. failed to get temp file")

		return nil, err
	}
	gen := types.Generation2

	pkgFileName, err := nodepkg.FormatPkgFileName(
		gen,
		types.ReleaseTypeOriginAgent,
		platfmt.UnknownPlatform(),
		detail.Version,
	)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package, failed to format package")

		return nil, err
	}
	pkgFileName = m.wrapOriginPackageName(pkgFileName)

	// upload to upstream.
	if err := m.upstreamOriginAgent.Store(nCtx, fileiface.FileInfo{Name: pkgFileName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package, failed to upload to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginAgent.GetFile(nCtx, pkgFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseAgent(nCtx, gen, detail.Version, detail.Platforms...)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateAgentUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginAgent,
		SavedName: pkgFileName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin agent package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	// record event.
	m.recordUploadEvent(nCtx, types.ReleaseTypeOriginAgent, types.ReleaseNameAgent, detail.Version, detail.Platforms)

	logger.G.Biz(nCtx).With("version", detail.Version, "filename", pkgFileName).Info("uploaded origin agent package to upstream")

	return detail, nil
}

// checkGSE2OriginAgentPkg check origin agent package.
// nolint:funlen,gocognit,gocyclo,cyclop,lll
// NOCC: golint/fnsize(func design is not suitable for splitting).
func checkGSE2OriginAgentPkg(file io.ReadCloser) (*types.OriginPkgDetail, error) {
	plats := make(map[string]platfmt.Platform)
	detail := types.NewOriginPkgDetail()
	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{".*", buildFullMatchRegex(originalAgentFileNameVersion)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalAgentFileNameDescription)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalAgentFileNameDescriptionEN)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalAgentDirNameSupportFile), buildFullMatchRegex(originalAgentDirNameTemplates), buildFullMatchRegex(originalAgentFileNameConfTemplateAgentTypeOne)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalAgentDirNameSupportFile), buildFullMatchRegex(originalAgentDirNameTemplates), buildFullMatchRegex(originalAgentFileNameConfTemplateAgentTypeTwo)},
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
			filePathRegex: []string{".*", buildFullMatchRegex(originalAgentDirNameSupportFile), buildFullMatchRegex(originalAgentDirNameEnv), buildFullMatchRegex(originalAgentFileNameAgentEnv)},
			callback: func(_ []string, r io.Reader) error {
				environ, err := parseEnvFile(r)
				if err != nil {
					return fmt.Errorf("failed to read gse_agent.env file: %w", err)
				}

				detail.ConfigEnviron = environ

				return nil
			},
		},
		{
			filePathRegex: []string{".*", ".*", buildFullMatchRegex(originalAgentDirNameBin), ".*"},
			callback: func(path []string, _ io.Reader) error {
				plat := convAgentDirNameToPlat(path[1])

				//  only has agent file to be considered a valid agent package.
				if path[3] == platfmt.FormatBinaryFileName(agentPkgFileNameAgent, plat.OS) {
					plats[plat.String()] = plat
				}

				return nil
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to check origin agent package: %w", err)
	}

	detail.Platforms = conv.MapValueToSlice(plats)

	return detail, nil
}

// PublishReleaseAgent generates release agent packages by upload-id and is-shared.
// nolint:funlen,gocognit,gocyclo,cyclop
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) PublishReleaseAgent(nCtx contextx.IContext, uploadID string, isShared bool) error {
	up, err := m.storageUpload.GetAgentUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginAgent {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	originFile, err := m.upstreamOriginAgent.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	originContent, err := originFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// store file to temp.
	originTempFileName, err := m.saveTempFile(nCtx, originContent)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release agent package. failed to save temp file")

		return err
	}

	checkingFile, err := m.getTempFile(nCtx, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release agent package. failed to get temp file")

		return err
	}

	detail, err := checkGSE2OriginAgentPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release agent package. failed to check origin agent package")

		return err
	}

	// generate release packages.
	releasePkgs, err := m.generateAgentPkg(nCtx, detail, originTempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to generate agent pkg")

		return err
	}

	gp := gopool.NewPool()
	gen := types.Generation2

	releasesMap := make(map[string]*types.ReleaseAgent)
	for idx := range releasePkgs {
		pkg := releasePkgs[idx]
		gp.Go(func() error {
			// generate package name.
			pkgName, err := nodepkg.FormatPkgFileName(
				gen,
				types.ReleaseTypeAgent,
				pkg.platform,
				detail.Version,
			)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to upload release agent package, failed to format package")

				return err
			}

			generatedFile, err := m.getTempFile(nCtx, pkg.tempFileName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to get temp file")

				return err
			}

			if err = m.upstreamReleaseAgent.Store(nCtx, fileiface.FileInfo{Name: pkgName}, generatedFile, true); err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to upload to upstream")

				return err
			}

			file, err := m.upstreamReleaseAgent.GetFile(nCtx, pkgName)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to get temp file")

				return err
			}

			agentInfo := &types.ReleaseAgent{
				Release: types.Release{
					Name:       types.ReleaseNameAgent,
					Generation: gen,
					Type:       types.ReleaseTypeAgent,
					Version:    detail.Version,
					Platform:   pkg.platform,
					FileName:   file.Info().Name,
					MD5:        file.Info().MD5,
					Enabled:    false,
					AsDefault:  false,
					IsShared:   isShared && nCtx.TenantID() == tenant.SystemTenantID,
					IsSynced:   false,
					Operator:   nCtx.BKUsername(),
					UpdatedAt:  time.Now(),
				},
				ReleaseAdditionInfoAgent: types.ReleaseAdditionInfoAgent{
					ConfigTemplate: detail.ConfigTemplate,
					ConfigEnviron:  detail.ConfigEnviron,
					ChangeLogEN:    detail.ChangeLogEN,
					ChangeLogZH:    detail.ChangeLogZH,
				},
			}

			releasesMap[pkg.platform.String()] = agentInfo

			// record package events.
			m.recordPublishEvent(nCtx, &agentInfo.Release)

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to upload to upstream")

		return err
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertManyReleaseAgent(nCtx, conv.MapValueToSlice(releasesMap)); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release agent, failed to upsert release agent")

		return err
	}

	logger.G.Biz(nCtx).With("platforms", detail.Platforms).Info("generated and published release agents")

	return nil
}

func (m *Manager) recordUploadEvent(nCtx contextx.IContext, rt types.ReleaseType, name string, version string, platforms []platfmt.Platform) {
	events := make([]*types.PackageEvent, len(platforms))
	for idx := range platforms {
		plt := platforms[idx]
		events[idx] = &types.PackageEvent{
			Name:        name,
			EventType:   types.PackageEventTypeUpload,
			ReleaseType: rt,
			Generation:  types.Generation2,
			Version:     version,
			OSType:      plt.OS,
			CPUArch:     plt.Arch,
			Operator:    nCtx.BKUsername(),
			OperateTime: time.Now(),
		}
	}

	m.recordPackageEvents(nCtx, events...)
}

func (m *Manager) recordPublishEvent(nCtx contextx.IContext, releaseInfo *types.Release) {
	m.recordPackageEvents(nCtx, &types.PackageEvent{
		Name:        releaseInfo.Name,
		EventType:   types.PackageEventTypePublish,
		ReleaseType: releaseInfo.Type,
		Generation:  releaseInfo.Generation,
		Version:     releaseInfo.Version,
		OSType:      releaseInfo.Platform.OS,
		CPUArch:     releaseInfo.Platform.Arch,
		Operator:    nCtx.BKUsername(),
		OperateTime: time.Now(),
	})
}

func (m *Manager) recordPackageEvents(nCtx contextx.IContext, events ...*types.PackageEvent) {
	err := m.goAsyncPool.Run(
		nCtx,
		func(asyncCtx contextx.IContext) error {
			return m.storageEvent.CreateManyPackageEvent(asyncCtx, events...)
		},
		goasync.WithName("record_package_event"),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to submit package event recording task")
	}
}

type releaseAgentPkg struct {
	platform     platfmt.Platform
	tempFileName string
}

// generateAgentPkg generates agent package.
// nolint:funlen,gocognit,gocyclo,cyclop,lll
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) generateAgentPkg(nCtx contextx.IContext, originDetail *types.OriginPkgDetail, originLocalFileName string) (
	[]*releaseAgentPkg, error) {

	// local origin agent.
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

	result := make(map[string]*releaseAgentPkg)
	for idx := range originDetail.Platforms {
		plat := originDetail.Platforms[idx]

		gp.Go(func() error {
			// create target file.
			tempFileName, err := m.createTempFile(nCtx)
			if err != nil {
				return fmt.Errorf("failed to create agent pkg temp file. platform(%s): %w", plat.String(), err)
			}
			targetFile, err := m.openTempFile(nCtx, tempFileName)
			if err != nil {
				return fmt.Errorf("failed to open pkg file: %w", err)
			}

			// open all source files.
			originAgentFile, err := localOrigin.Content(nCtx)
			if err != nil {
				return fmt.Errorf("failed to open origin agent file: %w", err)
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
					{targetFilePath: []string{agentPkgDirNameBin}, targetFileMode: tgzModeDir},
					{targetFilePath: []string{agentPkgDirNameCert}, targetFileMode: tgzModeDir},
				},
				[]*tgzWriteRuleStream{
					// get things from origin agent.
					{
						sourceFile: originAgentFile,
						fileRules: []tgzWriteRuleFile{
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, convPlatToAgentDirName(plat), originalAgentDirNameBin, tgzPathMatchingSegment2},
								targetFilePath: []string{agentPkgDirNameBin, tgzPathMatchingSegment2},
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
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameCaCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameAgentCrt},
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameAgentCrt},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameAgentKey},
								targetFilePath: []string{agentPkgDirNameCert, agentPkgFileNameAgentKey},
								targetFileMode: tgzModeFile,
							},
							{
								sourceFilePath: []string{tgzPathMatchingSegment1, certFileNameCertEncryptKey},
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
								sourceFilePath: []string{tgzPathMatchingSegment1, convPlatToBinToolDirName(types.ReleaseTypeAgent, plat), tgzPathMatchingSegment2},
								targetFilePath: []string{agentPkgDirNameBin, tgzPathMatchingSegment2},
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

	originalAgentFileNameVersion                  = "VERSION"
	originalAgentFileNameDescription              = "DESCRIPTION"
	originalAgentFileNameDescriptionEN            = "DESCRIPTION_EN"
	originalAgentFileNameConfTemplateAgentTypeOne = "#etc#gse#gse_agent.conf"
	originalAgentFileNameConfTemplateAgentTypeTwo = "gse_agent.conf.template"
	originalAgentFileNameAgentEnv                 = "gse_agent.env"
	originalAgentFileNameAgent                    = "gse_agent"

	agentPkgDirNameCert = "cert"
	agentPkgDirNameBin  = "bin"

	agentPkgFileNameCaCrt          = "gseca.crt"
	agentPkgFileNameAgentCrt       = "gse_agent.crt"
	agentPkgFileNameAgentKey       = "gse_agent.key"
	agentPkgFileNameCertEncryptKey = "cert_encrypt.key"
	agentPkgFileNameAgent          = "gse_agent"
)

func convPlatToAgentDirName(plat platfmt.Platform) string {
	return fmt.Sprintf("%s%s_%s", originalAgentDirNamePlatPrefix, plat.OS.String(), plat.Arch.ToPkgArch())
}

func convAgentDirNameToPlat(dirName string) platfmt.Platform {
	if !strings.HasPrefix(dirName, originalAgentDirNamePlatPrefix) {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	dirSplit := strings.SplitN(dirName, "_", originalAgentDirNamePlatSplitTimes)
	if len(dirSplit) != originalAgentDirNamePlatSplitTimes {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	plat, err := platfmt.Normalize(dirSplit[1], dirSplit[2])
	if err != nil {
		return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
	}

	return plat
}
