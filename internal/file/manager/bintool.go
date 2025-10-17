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
	"bytes"
	"errors"
	"io"
	"runtime/debug"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IBinTool defines the interface for bin tool.
type IBinTool interface {
	// UploadOriginBinTool upload origin bintool package.
	UploadOriginBinTool(nCtx contextx.IContext, binToolFile io.ReadCloser) (
		*types.OriginBinToolPkgDetail, error)

	// PublishReleaseBinTool generate release bintool package.
	PublishReleaseBinTool(nCtx contextx.IContext, uploadID string) error
}

const (
	originBinToolFileName  = "bintool-all.tgz"
	releaseBinToolFileName = "bintool.tgz"
)

// UploadOriginBinTool upload generation2 origin bintool package.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginBinTool(
	nCtx contextx.IContext,
	binToolFile io.ReadCloser) (*types.OriginBinToolPkgDetail, error) {

	// validation.
	if binToolFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin bintool package. bin tool file is nil")

		return nil, errors.New("bin tool file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, binToolFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginBinToolPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error(
			"failed to upload origin bintool package. failed to check origin bintool package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package. failed to get temp file")

		return nil, err
	}

	pkgName := m.wrapOriginPackageName(originBinToolFileName)

	// upload to upstream.
	if err := m.upstreamOriginBinTool.Store(nCtx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package, failed to upload to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginBinTool.GetFile(nCtx, pkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleaseBinTool(nCtx, types.Generation2)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreateBinToolUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginBinTool,
		SavedName: pkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin bintool package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	logger.G.Biz(nCtx).With("filename", pkgName).Info("uploaded origin bintool package to upstream")

	return detail, nil
}

func checkOriginBinToolPkg(file io.ReadCloser) (*types.OriginBinToolPkgDetail, error) {
	agentPlatforms := make(map[string]platfmt.Platform)
	proxyPlatforms := make(map[string]platfmt.Platform)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{buildFullMatchRegex(binToolDirNameRoot), ".*", ".*"},
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
	detail.AgentPlatforms = make([]platfmt.Platform, 0)
	detail.ProxyPlatforms = make([]platfmt.Platform, 0)
	for _, plat := range agentPlatforms {
		detail.AgentPlatforms = append(detail.AgentPlatforms, plat)
	}
	for _, plat := range proxyPlatforms {
		detail.ProxyPlatforms = append(detail.ProxyPlatforms, plat)
	}

	return detail, nil
}

// PublishReleaseBinTool generates release bintool by upload-id.
// nolint:funlen
func (m *Manager) PublishReleaseBinTool(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetBinToolUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginBinTool {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginBinTool.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	content, err := file.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// generate release file.
	generatedFile, err := m.generateBinToolPkg(nCtx, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to generate bintool pkg")

		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleaseBinTool.Store(
		nCtx, fileiface.FileInfo{Name: releaseBinToolFileName}, generatedFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to upload to upstream")

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleaseBinTool.GetFile(nCtx, releaseBinToolFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to get release file")

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	bintoolInfo := &types.ReleaseBinTool{
		Release: types.Release{
			Generation:   types.Generation2,
			Type:         types.ReleaseTypeBinTool,
			Platform:     platfmt.EmptyPlatform(),
			Labels:       nil,
			FileName:     releaseInfo.Name,
			MD5:          releaseInfo.MD5,
			Enabled:      true,
			AsDefault:    true,
			UpdatedAt:    time.Now(),
			Operator:     nCtx.BKUsername(),
			AdditionInfo: nil,
		},
	}

	// upsert release bintool.
	if err = m.storageRelease.UpsertReleaseBinTool(nCtx, *bintoolInfo); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release bintool, failed to upsert release bintool")

		return err
	}

	go m.recordBinToolEvent(nCtx, &bintoolInfo.Release, types.PackageEventTypePublish, types.PackageEventTypeEnable, types.PackageEventTypeSetAsDefault)

	logger.G.Biz(nCtx).With("filename", releaseInfo.Name, "md5", releaseInfo.MD5).Info("generated and published release bintool")

	return nil
}

func (m *Manager) recordBinToolEvent(nCtx contextx.IContext, releaseInfo *types.Release, eventType ...types.PackageEventType) {
	// recover panic.
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			// The first line of the stack trace is of the form "goroutine N [status]:",
			// but by the time the panic reaches Do the goroutine may no longer exist,
			// and its status will have changed. Trim out the misleading line.
			if line := bytes.IndexByte(stack[:], '\n'); line >= 0 { //nolint: gocritic
				stack = stack[line+1:]
			}

			logger.G.Sys().With("event-type", eventType, "recover", r, "stack", stack).Error("failed to record package event")
		}
	}()

	events := make([]*types.PackageEvent, len(eventType))
	for idx, et := range eventType {
		events[idx] = &types.PackageEvent{
			EventType:   et,
			ReleaseType: releaseInfo.Type,
			Generation:  releaseInfo.Generation,
			Version:     releaseInfo.Version,
			OSType:      releaseInfo.Platform.OS,
			CPUArch:     releaseInfo.Platform.Arch,
			OperateTime: releaseInfo.UpdatedAt,
			Operator:    releaseInfo.Operator,
		}
	}

	if err := m.storageEvent.CreateManyPackageEvent(nCtx, events...); err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("event-type", eventType).
			Warn("failed to record package event event, failed to create package event")
	}
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

func (m *Manager) generateBinToolPkg(nCtx contextx.IContext, sourceFile io.ReadCloser) (io.ReadCloser, error) {
	tempFileName, err := m.createTempFile(nCtx)
	if err != nil {
		return nil, err
	}

	targetFile, err := m.openTempFile(nCtx, tempFileName)
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
					sourceFilePath: []string{tgzPathMatchingSegment1, binToolDirNameAgentPlatLinuxAmd64, tgzPathMatchingSegment2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatLinuxAmd64, tgzPathMatchingSegment2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathMatchingSegment1, binToolDirNameAgentPlatLinuxArm64, tgzPathMatchingSegment2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatLinuxArm64, tgzPathMatchingSegment2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathMatchingSegment1, binToolDirNameAgentPlatDarwinAmd64, tgzPathMatchingSegment2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatDarwinAmd64, tgzPathMatchingSegment2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathMatchingSegment1, binToolDirNameAgentPlatWindowsAmd64, tgzPathMatchingSegment2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameAgentPlatWindowsAmd64, tgzPathMatchingSegment2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathMatchingSegment1, binToolDirNameProxyPlatLinuxAmd64, tgzPathMatchingSegment2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameProxyPlatLinuxAmd64, tgzPathMatchingSegment2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathMatchingSegment1, binToolDirNameProxyPlatLinuxArm64, tgzPathMatchingSegment2},
					targetFilePath: []string{binToolDirNameRoot, binToolDirNameProxyPlatLinuxArm64, tgzPathMatchingSegment2},
					targetFileMode: tgzModeFile,
				},
			},
		}},
	); err != nil {
		return nil, err
	}

	file, err := m.tempFileGroup.GetFile(nCtx, tempFileName)
	if err != nil {
		return nil, err
	}

	return file.Content(nCtx)
}

func convPlatToBinToolDirName(releaseType types.ReleaseType, plat platfmt.Platform) string {
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

func convBinToolDirNameToPlat(dirName string) platfmt.Platform {
	switch dirName {
	case binToolDirNameAgentPlatLinuxAmd64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case binToolDirNameAgentPlatLinuxArm64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	case binToolDirNameAgentPlatWindowsAmd64:
		{
			return platfmt.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
		}
	case binToolDirNameAgentPlatDarwinAmd64:
		{
			return platfmt.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64)
		}
	case binToolDirNameProxyPlatLinuxAmd64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case binToolDirNameProxyPlatLinuxArm64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	default:
		{
			return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
