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
	"io"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPluginBinTool defines the interface for bin tool.
type IPluginBinTool interface {
	// UploadOriginPluginBinTool upload origin plugin bintool package.
	UploadOriginPluginBinTool(nCtx contextx.IContext, binToolFile io.ReadCloser) (
		*types.OriginPluginBinToolPkgDetail, error)

	// PublishReleasePluginBinTool generate release plugin bintool package.
	PublishReleasePluginBinTool(nCtx contextx.IContext, uploadID string) error
}

const (
	originPluginBinToolFileName  = "plugin_bintool-all.tgz"
	releasePluginBinToolFileName = "plugin_bintool.tgz"
)

// UploadOriginPluginBinTool upload generation2 origin plugin bintool package.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPluginBinTool(
	nCtx contextx.IContext,
	binToolFile io.ReadCloser) (*types.OriginPluginBinToolPkgDetail, error) {

	// validation.
	if binToolFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin plugin bintool package. bin tool file is nil")

		return nil, errors.New("bin tool file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, binToolFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginPluginBinToolPkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error(
			"failed to upload origin plugin bintool package. failed to check origin plugin bintool package")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package. failed to get temp file")

		return nil, err
	}

	pkgName := m.wrapOriginPackageName(originPluginBinToolFileName)

	// upload to upstream.
	if err := m.upstreamOriginPluginBinTool.Store(nCtx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package, failed to upload to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginPluginBinTool.GetFile(nCtx, pkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePluginBinTool(nCtx, types.Generation2)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreatePluginBinToolUpload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginPluginBinTool,
		SavedName: pkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	logger.G.Biz(nCtx).With("filename", pkgName).Info("uploaded origin plugin bintool package to upstream")

	return detail, nil
}

const (
	pluginBinToolDirNameRoot             = "plugin_bintool"
	pluginBinToolDirNamePlatLinuxAmd64   = "linux_amd64"
	pluginBinToolDirNamePlatLinuxArm64   = "linux_arm64"
	pluginBinToolDirNamePlatDarwinAmd64  = "darwin_amd64"
	pluginBinToolDirNamePlatWindowsAmd64 = "windows_amd64"
)

func checkOriginPluginBinToolPkg(file io.ReadCloser) (*types.OriginPluginBinToolPkgDetail, error) {
	platforms := make(map[string]platfmt.Platform)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{buildFullMatchRegex(pluginBinToolDirNameRoot), ".*", ".*"},
			callback: func(path []string, _ io.Reader) error {
				plat := convPluginBinToolDirNameToPlat(path[1])
				platforms[plat.String()] = plat

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	detail := new(types.OriginPluginBinToolPkgDetail)
	detail.Platforms = make([]platfmt.Platform, 0)
	for _, plat := range platforms {
		detail.Platforms = append(detail.Platforms, plat)
	}

	return detail, nil
}

// PublishReleasePluginBinTool generates release plugin bintool by upload-id.
func (m *Manager) PublishReleasePluginBinTool(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginBinToolUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginPluginBinTool {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin bintool, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginPluginBinTool.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin bintool, failed to get file")

		return err
	}

	// get origin content.
	content, err := file.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin bintool, failed to get content")

		return err
	}

	// generate release file.
	generatedFile, err := m.generatePluginBinToolPkg(nCtx, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool, failed to generate plugin bintool pkg")

		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleasePluginBinTool.Store(
		nCtx, fileiface.FileInfo{Name: releasePluginBinToolFileName}, generatedFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool, failed to upload to upstream")

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleasePluginBinTool.GetFile(nCtx, releasePluginBinToolFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool, failed to get release file")

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	// upsert release plugin bintool.
	if err = m.storageRelease.UpsertReleasePluginBinTool(nCtx, types.ReleasePluginBinTool{
		Release: types.Release{
			Generation:   types.Generation2,
			Type:         types.ReleaseTypePluginBinTool,
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
	}); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool, failed to upsert release plugin bintool")

		return err
	}

	logger.G.Biz(nCtx).With("filename", releaseInfo.Name, "md5", releaseInfo.MD5).Info("generated and published release plugin bintool")

	return nil
}

func (m *Manager) generatePluginBinToolPkg(nCtx contextx.IContext, sourceFile io.ReadCloser) (io.ReadCloser, error) {
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
			{targetFilePath: []string{pluginBinToolDirNameRoot}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxArm64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatDarwinAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatWindowsAmd64}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules: []tgzWriteRuleFile{
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxArm64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxArm64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatDarwinAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatDarwinAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatWindowsAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatWindowsAmd64, tgzPathMatchingSegment1},
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

func convPlatToPluginBinToolDirName(plat platfmt.Platform) string {
	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSLinux {
		return pluginBinToolDirNamePlatLinuxAmd64
	}

	if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
		return pluginBinToolDirNamePlatLinuxArm64
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
		return pluginBinToolDirNamePlatDarwinAmd64
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
		return pluginBinToolDirNamePlatWindowsAmd64
	}

	return ""
}

func convPluginBinToolDirNameToPlat(dirName string) platfmt.Platform {
	switch dirName {
	case pluginBinToolDirNamePlatLinuxAmd64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case pluginBinToolDirNamePlatLinuxArm64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	case pluginBinToolDirNamePlatWindowsAmd64:
		{
			return platfmt.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
		}
	case pluginBinToolDirNamePlatDarwinAmd64:
		{
			return platfmt.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64)
		}
	default:
		{
			return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
