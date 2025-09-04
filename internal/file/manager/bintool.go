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
	"io"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IBinTool defines the interface for bin tool.
type IBinTool interface {
	// UploadOriginBinTool upload origin bintool package.
	UploadOriginBinTool(ctx contextx.IContext, binToolFile io.ReadCloser) (
		*types.OriginBinToolPkgDetail, error)

	// PublishReleaseBinTool generate release bintool package.
	PublishReleaseBinTool(ctx contextx.IUserContext, uploadID string) error
}

const (
	originBinToolFileName  = "bintool-all.tgz"
	releaseBinToolFileName = "bintool.tgz"
)

// UploadOriginBinTool upload generation2 origin bintool package.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginBinTool(
	ctx contextx.IContext,
	binToolFile io.ReadCloser) (*types.OriginBinToolPkgDetail, error) {

	// validation.
	if binToolFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package. bin tool file is nil")

		return nil, errors.New("bin tool file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, binToolFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkGen2OriginBinToolPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx,
			"failed to upload origin bintool package. failed to check origin bintool package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package. failed to get temp file. err: %v", err)

		return nil, err
	}

	pkgName := m.wrapOriginPackageName(originBinToolFileName)

	// upload to upstream.
	if err := m.upstreamOriginBinTool.Store(ctx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package, failed to upload to upstream. err: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginBinTool.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseBinTool(ctx, types.Generation2)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateBinToolUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginBinTool,
		SavedName: pkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin bintool package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx,
		"uploaded origin bintool package to upstream. file-name(%s)", pkgName)

	return detail, nil
}

func checkGen2OriginBinToolPkg(file io.ReadCloser) (*types.OriginBinToolPkgDetail, error) {
	agentPlatforms := make(map[string]platform.Platform)
	proxyPlatforms := make(map[string]platform.Platform)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{binToolDirNameRoot, tgzPathNameAny1, tgzPathNameAny2},
			callback: func(path []string, _ io.Reader) error {
				plat := convBinToolDirNameToPlat(path[1])
				agentPlatforms[plat.String()] = plat

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	detail := new(types.OriginBinToolPkgDetail)
	detail.AgentPlatforms = make([]platform.Platform, 0)
	detail.ProxyPlatforms = make([]platform.Platform, 0)
	for _, plat := range agentPlatforms {
		detail.AgentPlatforms = append(detail.AgentPlatforms, plat)
	}
	for _, plat := range proxyPlatforms {
		detail.ProxyPlatforms = append(detail.ProxyPlatforms, plat)
	}

	return detail, nil
}

// PublishReleaseBinTool generates release bintool by upload-id.
func (m *Manager) PublishReleaseBinTool(ctx contextx.IUserContext, uploadID string) error {
	up, err := m.storageUpload.GetBinToolUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginBinTool {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginBinTool.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	content, err := file.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// generate release file.
	generatedFile, err := m.generateBinToolPkg(ctx, content)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to generate bintool pkg. err: %v", err)

		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleaseBinTool.Store(
		ctx, fileiface.FileInfo{Name: releaseBinToolFileName}, generatedFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to upload to upstream. err: %v", err)

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleaseBinTool.GetFile(ctx, releaseBinToolFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to get release file. err: %v", err)

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	// upsert release bintool.
	if err = m.storageRelease.UpsertReleaseBinTool(ctx, types.ReleaseBinTool{
		Release: types.Release{
			Generation:   types.Generation2,
			Type:         types.ReleaseTypeBinTool,
			Version:      "",
			Platform:     platform.Platform{},
			Labels:       nil,
			FileName:     releaseInfo.Name,
			MD5:          releaseInfo.MD5,
			Enabled:      true,
			AsDefault:    true,
			UpdatedAt:    time.Now(),
			Operator:     ctx.BKUsername(),
			AdditionInfo: nil,
		},
	}); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release bintool, failed to upsert release bintool: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release bintool. file(%s), md5(%s)", releaseInfo.Name, releaseInfo.MD5)

	return nil
}

const (
	binToolDirNameRoot                  = "bintool"
	binToolDirNameAgentPlatLinuxAmd64   = "agent_linux_amd64"
	binToolDirNameAgentPlatLinuxArm64   = "agent_linux_arm64"
	binToolDirNameAgentPlatDarwinAmd64  = "agent_darwin_amd64"
	binToolDirNameAgentPlatWindowsAmd64 = "agent_windows_amd64"
	binToolDirNameProxyPlatLinuxAmd64   = "proxy_linux_amd64"
	binToolDirNameProxyPlatLinuxArm64   = "proxy_linux_arm64"
)

func (m *Manager) generateBinToolPkg(ctx context.Context, sourceFile io.ReadCloser) (io.ReadCloser, error) {
	tempFileName, err := m.createTempFile(ctx)
	if err != nil {
		return nil, err
	}

	targetFile, err := m.openTempFile(ctx, tempFileName)
	if err != nil {
		return nil, err
	}

	if err = generateTgz(targetFile,
		[]tgzWriteRuleDir{
			{targetFilePath: []string{binToolDirNameRoot}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatLinuxAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatLinuxArm64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatDarwinAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatWindowsAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{binToolDirNameRoot, binToolDirNameProxyPlatLinuxAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{binToolDirNameRoot, binToolDirNameProxyPlatLinuxArm64}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules: []tgzWriteRuleFile{
				{
					sourceFilePath: []string{tgzPathNameAny1, binToolDirNameAgentPlatLinuxAmd64, tgzPathNameAny2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatLinuxAmd64, tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, binToolDirNameAgentPlatLinuxArm64, tgzPathNameAny2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatLinuxArm64, tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, binToolDirNameAgentPlatDarwinAmd64, tgzPathNameAny2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatDarwinAmd64, tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, binToolDirNameAgentPlatWindowsAmd64, tgzPathNameAny2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatWindowsAmd64, tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, binToolDirNameProxyPlatLinuxAmd64, tgzPathNameAny2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameProxyPlatLinuxAmd64, tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, binToolDirNameProxyPlatLinuxArm64, tgzPathNameAny2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameProxyPlatLinuxArm64, tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
			},
		}},
	); err != nil {
		return nil, err
	}

	file, err := m.tempFileGroup.GetFile(ctx, tempFileName)
	if err != nil {
		return nil, err
	}

	return file.Content(ctx)
}

func convPlatToBinToolDirName(releaseType types.ReleaseType, plat platform.Platform) string {
	if releaseType == types.ReleaseTypeAgent {
		if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
			return binToolDirNameAgentPlatLinuxAmd64
		}

		if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
			return binToolDirNameAgentPlatLinuxArm64
		}

		if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
			return binToolDirNameAgentPlatDarwinAmd64
		}

		if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
			return binToolDirNameAgentPlatWindowsAmd64
		}
	}

	if releaseType == types.ReleaseTypeProxy {
		if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
			return binToolDirNameProxyPlatLinuxAmd64
		}

		if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
			return binToolDirNameProxyPlatLinuxArm64
		}
	}

	return ""
}

func convBinToolDirNameToPlat(dirName string) platform.Platform {
	switch dirName {
	case binToolDirNameAgentPlatLinuxAmd64:
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case binToolDirNameAgentPlatLinuxArm64:
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	case binToolDirNameAgentPlatWindowsAmd64:
		{
			return platform.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
		}
	case binToolDirNameAgentPlatDarwinAmd64:
		{
			return platform.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64)
		}
	case binToolDirNameProxyPlatLinuxAmd64:
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case binToolDirNameProxyPlatLinuxArm64:
		{
			return platform.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	default:
		{
			return platform.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
