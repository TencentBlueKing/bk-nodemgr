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
	// UploadOriginPluginBinTool upload origin plugin bintool package.
	UploadOriginPluginBinTool(nCtx contextx.IContext, binToolFile io.ReadCloser) (
		*types.OriginPluginBinToolPkgDetail, error)

	// PublishReleasePluginBinToolV2 generate release plugin bintool package v2.
	PublishReleasePluginBinToolV2(nCtx contextx.IContext, uploadID string) error

	// PublishReleasePluginBinToolV3 generate release plugin bintool package v3.
	PublishReleasePluginBinToolV3(nCtx contextx.IContext, uploadID string) error
}

const (
	originPluginBinToolFileName    = "plugin_bintool-all.tgz"
	releasePluginBinToolV2FileName = "plugin_bintool_v2.tgz"
	releasePluginBinToolV3FileName = "plugin_bintool_v3.tgz"
)

// UploadOriginPluginBinTool upload generation2 origin plugin bintool package.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPluginBinTool(nCtx contextx.IContext, binToolFile io.ReadCloser) (*types.OriginPluginBinToolPkgDetail, error) {
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
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package. failed to check origin plugin bintool package")
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
	existedV2, err := m.storageRelease.ExistReleasePluginBinToolV2(nCtx, types.Generation2)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v2. failed to check if release existed")
		return nil, err
	}
	existedV3, err := m.storageRelease.ExistReleasePluginBinToolV3(nCtx, types.Generation2)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to upload origin plugin bintool package v3. failed to check if release existed")
		return nil, err
	}

	detail.Existed = existedV3 || existedV2

	uploadID, err := m.storageUpload.CreatePluginBinToolV2Upload(nCtx, &types.Upload{
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
	pluginBinToolDirNameGenerationV2     = "v2"
	pluginBinToolDirNameGenerationV3     = "v3"
	pluginBinToolDirNamePlatLinuxAmd64   = "linux_amd64"
	pluginBinToolDirNamePlatLinuxArm64   = "linux_arm64"
	pluginBinToolDirNamePlatDarwinAmd64  = "darwin_amd64"
	pluginBinToolDirNamePlatWindowsAmd64 = "windows_amd64"
	pluginBinToolDirNamePlatAix6Ppc64    = "aix6_ppc64"
	pluginBinToolDirNamePlatAix7Ppc64    = "aix7_ppc64"
)

func checkOriginPluginBinToolPkg(file io.ReadCloser) (*types.OriginPluginBinToolPkgDetail, error) {
	platformsV2 := make(map[string]platfmt.Platform)
	platformsV3 := make(map[string]platfmt.Platform)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePathRegex: []string{buildFullMatchRegex(pluginBinToolDirNameRoot), buildFullMatchRegex(pluginBinToolDirNameGenerationV2), ".*", ".*"},
			callback: func(path []string, _ io.Reader) error {
				plat := convPluginBinToolDirNameToPlat(path[2])
				platformsV2[plat.String()] = plat

				return nil
			},
		},
		{
			filePathRegex: []string{buildFullMatchRegex(pluginBinToolDirNameRoot), buildFullMatchRegex(pluginBinToolDirNameGenerationV3), ".*", ".*"},
			callback: func(path []string, _ io.Reader) error {
				plat := convPluginBinToolDirNameToPlat(path[2])
				platformsV3[plat.String()] = plat

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	detail := new(types.OriginPluginBinToolPkgDetail)
	detail.V2.Platforms = make([]platfmt.Platform, 0)
	detail.V3.Platforms = make([]platfmt.Platform, 0)
	for _, plat := range platformsV2 {
		detail.V2.Platforms = append(detail.V2.Platforms, plat)
	}

	for _, plat := range platformsV3 {
		detail.V3.Platforms = append(detail.V3.Platforms, plat)
	}

	return detail, nil
}

// PublishReleasePluginBinToolV2 generates release plugin bintool v2 by upload-id.
// nolint: funlen
func (m *Manager) PublishReleasePluginBinToolV2(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginBinToolUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool v2, failed to get upload")
		return err
	}

	if up.Category != types.UploadCategoryOriginPluginBinTool {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin bintool v2, invalid category")
		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginPluginBinTool.GetFile(nCtx, up.SavedName)
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
	generatedFile, err := m.generatePluginBinToolPkg(nCtx, content, types.ReleaseTypePluginBinToolV2)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to generate plugin bintool v2 pkg")
		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleasePluginBinTool.Store(
		nCtx, fileiface.FileInfo{Name: releasePluginBinToolV2FileName}, generatedFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to upload to upstream")
		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleasePluginBinTool.GetFile(nCtx, releasePluginBinToolV2FileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to get release file")
		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	pluginBinToolInfo := &types.ReleasePluginBinToolV2{
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
	}
	// upsert release plugin bintool v2.
	if err = m.storageRelease.UpsertReleasePluginBinToolV2(nCtx, *pluginBinToolInfo); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v2, failed to upsert release plugin bintool v2")
		return err
	}

	go m.recordBinToolEvent(nCtx, &pluginBinToolInfo.Release, types.PackageEventTypePublish)

	logger.G.Biz(nCtx).With("filename", releaseInfo.Name, "md5", releaseInfo.MD5).Info("generated and published release plugin bintool v2")

	return nil
}

// PublishReleasePluginBinToolV3 generates release plugin bintool v3 by upload-id.
// nolint: funlen
func (m *Manager) PublishReleasePluginBinToolV3(nCtx contextx.IContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginBinToolUpload(nCtx, uploadID)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("upload-id", uploadID).Error("failed to publish release plugin bintool v3, failed to get upload")
		return err
	}

	if up.Category != types.UploadCategoryOriginPluginBinTool {
		logger.G.Biz(nCtx).WithErr(err).With("category", up.Category).Error("failed to publish release plugin bintool v3, invalid category")
		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginPluginBinTool.GetFile(nCtx, up.SavedName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin bintool v3, failed to get file")
		return err
	}

	// get origin content.
	content, err := file.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("filename", up.SavedName).Error("failed to publish release plugin bintool v3, failed to get content")
		return err
	}

	// generate release file.
	generatedFile, err := m.generatePluginBinToolPkg(nCtx, content, types.ReleaseTypePluginBinToolV3)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v3, failed to generate plugin bintool v3 pkg")
		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleasePluginBinTool.Store(
		nCtx, fileiface.FileInfo{Name: releasePluginBinToolV3FileName}, generatedFile, true); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v3, failed to upload to upstream")
		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleasePluginBinTool.GetFile(nCtx, releasePluginBinToolV3FileName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v3, failed to get release file")
		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	pluginBinToolInfo := &types.ReleasePluginBinToolV3{
		Release: types.Release{
			Generation:   types.Generation2,
			Type:         types.ReleaseTypePluginBinToolV3,
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
	// upsert release plugin bintool v3.
	if err = m.storageRelease.UpsertReleasePluginBinToolV3(nCtx, *pluginBinToolInfo); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to publish release plugin bintool v3, failed to upsert release plugin bintool v3")
		return err
	}

	go m.recordBinToolEvent(nCtx, &pluginBinToolInfo.Release, types.PackageEventTypePublish)

	logger.G.Biz(nCtx).With("filename", releaseInfo.Name, "md5", releaseInfo.MD5).Info("generated and published release plugin bintool v3")

	return nil
}

// nolint: lll
func (m *Manager) generatePluginBinToolPkg(nCtx contextx.IContext, sourceFile io.ReadCloser, releaseType types.ReleaseType) (io.ReadCloser, error) {
	tempFileName, err := m.createTempFile(nCtx)
	if err != nil {
		return nil, err
	}

	targetFile, err := m.openTempFile(nCtx, tempFileName)
	if err != nil {
		return nil, err
	}

	pkgGeneration := pluginBinToolDirNameGenerationV2
	if releaseType == types.ReleaseTypePluginBinToolV3 {
		pkgGeneration = pluginBinToolDirNameGenerationV3
	}

	if err = generateTgz(targetFile,
		[]tgzWriteRuleDir{
			{targetFilePath: []string{pluginBinToolDirNameRoot}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxArm64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatDarwinAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatWindowsAmd64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatAix6Ppc64}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatAix7Ppc64}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules: []tgzWriteRuleFile{
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pkgGeneration, pluginBinToolDirNamePlatLinuxAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pkgGeneration, pluginBinToolDirNamePlatLinuxArm64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatLinuxArm64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pkgGeneration, pluginBinToolDirNamePlatDarwinAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatDarwinAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pkgGeneration, pluginBinToolDirNamePlatWindowsAmd64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatWindowsAmd64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pkgGeneration, pluginBinToolDirNamePlatAix6Ppc64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatAix6Ppc64, tgzPathMatchingSegment1},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{pluginBinToolDirNameRoot, pkgGeneration, pluginBinToolDirNamePlatAix7Ppc64, tgzPathMatchingSegment1},
					targetFilePath: []string{pluginBinToolDirNameRoot, pluginBinToolDirNamePlatAix7Ppc64, tgzPathMatchingSegment1},
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
	case pluginBinToolDirNamePlatAix6Ppc64:
		{
			return platfmt.NewPlatform(criteria.OSAix6, criteria.CPUArchPpc64)
		}
	case pluginBinToolDirNamePlatAix7Ppc64:
		{
			return platfmt.NewPlatform(criteria.OSAix7, criteria.CPUArchPpc64)
		}
	default:
		{
			return platfmt.NewPlatform(criteria.OSUnknown, criteria.CPUArchUnknown)
		}
	}
}
