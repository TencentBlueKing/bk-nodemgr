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

// IPluginBinToolV2 defines the interface for bin tool.
type IPluginBinToolV2 interface {
	// UploadOriginPluginBinToolV2 upload origin plugin bintool package v2.
	UploadOriginPluginBinToolV2(nCtx contextx.IContext, binToolFile io.ReadCloser) (
		*types.OriginPluginBinToolV2PkgDetail, error)

	// PublishReleasePluginBinToolV2 generate release plugin bintool package v2.
	PublishReleasePluginBinToolV2(nCtx contextx.IContext, uploadID string) error
}

const (
	originPluginBinToolV2FileName  = "plugin_bintool_v2-all.tgz"
	releasePluginBinToolV2FileName = "plugin_bintool_v2.tgz"
)

// UploadOriginPluginBinToolV2 upload generation2 origin plugin bintool package v2.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPluginBinToolV2(
	nCtx contextx.IContext,
	binToolFile io.ReadCloser) (*types.OriginPluginBinToolV2PkgDetail, error) {

	// validation.
	if binToolFile == nil {
		logger.G.Biz(nCtx).Error("failed to upload origin plugin bintool package v2. bin tool file is nil")

		return nil, errors.New("bin tool file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(nCtx, binToolFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2. failed to save temp file")

		return nil, err
	}

	checkingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2. failed to get temp file")

		return nil, err
	}

	detail, err := checkOriginPluginBinToolV2Pkg(checkingFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error(
			"failed to upload origin plugin bintool package v2. failed to check origin plugin bintool package v2")

		return nil, err
	}

	uploadingFile, err := m.getTempFile(nCtx, tempFileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2. failed to get temp file")

		return nil, err
	}

	pkgName := m.wrapOriginPackageName(originPluginBinToolV2FileName)

	// upload to upstream.
	if err := m.upstreamOriginPluginBinToolV2.Store(nCtx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2, failed to upload to upstream")

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginPluginBinToolV2.GetFile(nCtx, pkgName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2. failed to get file from upstream")

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePluginBinToolV2(nCtx, types.Generation2)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2. failed to check if release existed")

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreatePluginBinToolV2Upload(nCtx, &types.Upload{
		Category:  types.UploadCategoryOriginPluginBinToolV2,
		SavedName: pkgName,
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2, failed to create upload")

		return nil, err
	}
	detail.UploadID = uploadID

	logger.G.Biz(nCtx).With("filename", pkgName).Info("uploaded origin plugin bintool package v2 to upstream")

	return detail, nil
}

const (
	pluginBinToolV2DirNameRoot             = "plugin_bintool"
	pluginBinToolV2DirNamePlatLinuxAmd64   = "linux_amd64"
	pluginBinToolV2DirNamePlatLinuxArm64   = "linux_arm64"
	pluginBinToolV2DirNamePlatDarwinAmd64  = "darwin_amd64"
	pluginBinToolV2DirNamePlatWindowsAmd64 = "windows_amd64"
)

func checkOriginPluginBinToolV2Pkg(file io.ReadCloser) (*types.OriginPluginBinToolV2PkgDetail, error) {
	platforms := make(map[string]platfmt.Platform)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{buildFullMatchRegex(pluginBinToolV2DirNameRoot), ".*", ".*"},
			callback: func(path []string, _ io.Reader) error {
				plat := convPluginBinToolDirNameToPlat(path[1])
				platforms[plat.String()] = plat

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	detail := new(types.OriginPluginBinToolV2PkgDetail)
	detail.Platforms = make([]platfmt.Platform, 0)
	for _, plat := range platforms {
		detail.Platforms = append(detail.Platforms, plat)
	}

	return detail, nil
}

// PublishReleasePluginBinToolV2 generates release plugin bintool v2 by upload-id.
func (m *Manager) PublishReleasePluginBinToolV2(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginBinToolUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool v2, failed to get upload")

		return err
	}

	if up.Category != types.UploadCategoryOriginPluginBinToolV2 {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin bintool v2, invalid category")

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginPluginBinToolV2.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin bintool v2, failed to get file")

		return err
	}

	// get origin content.
	content, err := file.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin bintool v2, failed to get content")

		return err
	}

	// generate release file.
	generatedFile, err := m.generatePluginBinToolV2Pkg(nCtx, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to generate plugin bintool v2 pkg")

		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleasePluginBinToolV2.Store(
		nCtx, fileiface.FileInfo{Name: releasePluginBinToolV2FileName}, generatedFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to upload to upstream")

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleasePluginBinToolV2.GetFile(nCtx, releasePluginBinToolV2FileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to get release file")

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	// upsert release plugin bintool v2.
	if err = m.storageRelease.UpsertReleasePluginBinToolV2(nCtx, types.ReleasePluginBinToolV2{
		Release: types.Release{
			Generation:   types.Generation2,
			Type:         types.ReleaseTypePluginBinToolV2,
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
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to upsert release plugin bintool v2")

		return err
	}

	logger.G.Biz(nCtx).With("filename", releaseInfo.Name, "md5", releaseInfo.MD5).Info("generated and published release plugin bintool v2")

	return nil
}

func (m *Manager) generatePluginBinToolV2Pkg(nCtx contextx.IContext, sourceFile io.ReadCloser) (io.ReadCloser, error) {
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
			{targetFilePath: []string{pluginBinToolV2DirNameRoot}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatLinuxAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatLinuxArm64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatDarwinAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatWindowsAmd64}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules: []tgzWriteRuleFile{
				{
					sourceFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatLinuxAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatLinuxAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatLinuxArm64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatLinuxArm64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatDarwinAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatDarwinAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatWindowsAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolV2DirNameRoot, pluginBinToolV2DirNamePlatWindowsAmd64, tgzPathMatchingSegment1},
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
		return pluginBinToolV2DirNamePlatLinuxAmd64
	}

	if plat.Arch == criteria.CPUArchArm64 && plat.OS == criteria.OSLinux {
		return pluginBinToolV2DirNamePlatLinuxArm64
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSDarwin {
		return pluginBinToolV2DirNamePlatDarwinAmd64
	}

	if plat.Arch == criteria.CPUArchAmd64 && plat.OS == criteria.OSWindows {
		return pluginBinToolV2DirNamePlatWindowsAmd64
	}

	return ""
}

func convPluginBinToolDirNameToPlat(dirName string) platfmt.Platform {
	switch dirName {
	case pluginBinToolV2DirNamePlatLinuxAmd64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchAmd64)
		}
	case pluginBinToolV2DirNamePlatLinuxArm64:
		{
			return platfmt.NewPlatform(criteria.OSLinux, criteria.CPUArchArm64)
		}
	case pluginBinToolV2DirNamePlatWindowsAmd64:
		{
			return platfmt.NewPlatform(criteria.OSWindows, criteria.CPUArchAmd64)
		}
	case pluginBinToolV2DirNamePlatDarwinAmd64:
		{
			return platfmt.NewPlatform(criteria.OSDarwin, criteria.CPUArchAmd64)
		}
	default:
		{
			return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
