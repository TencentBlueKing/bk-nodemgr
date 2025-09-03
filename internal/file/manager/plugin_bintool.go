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

// IPluginBinTool defines the interface for bin tool.
type IPluginBinTool interface {
	// UploadOriginPluginBinTool upload origin plugin bintool package.
	UploadOriginPluginBinTool(ctx contextx.IContext, binToolFile io.ReadCloser) (
		*types.OriginPluginBinToolPkgDetail, error)

	// PublishReleasePluginBinTool generate release plugin bintool package.
	PublishReleasePluginBinTool(ctx contextx.IUserContext, uploadID string) error
}

const (
	originPluginBinToolFileName  = "plugin_bintool-all.tgz"
	releasePluginBinToolFileName = "plugin_bintool.tgz"
)

// UploadOriginPluginBinTool upload generation2 origin plugin bintool package.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (m *Manager) UploadOriginPluginBinTool(
	ctx contextx.IContext,
	binToolFile io.ReadCloser) (*types.OriginPluginBinToolPkgDetail, error) {

	// validation.
	if binToolFile == nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package. bin tool file is nil")

		return nil, errors.New("bin tool file is nil")
	}

	// store file to temp.
	tempFileName, err := m.saveTempFile(ctx, binToolFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package. failed to save temp file. err: %v", err)

		return nil, err
	}

	checkingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package. failed to get temp file. err: %v", err)

		return nil, err
	}

	detail, err := checkGen2OriginPluginBinToolPkg(checkingFile)
	if err != nil {
		m.logger.ErrorCtxf(ctx,
			"failed to upload origin plugin bintool package. failed to check origin plugin bintool package. err: %v", err)

		return nil, err
	}

	uploadingFile, err := m.getTempFile(ctx, tempFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package. failed to get temp file. err: %v", err)

		return nil, err
	}

	pkgName := m.wrapOriginPackageName(originPluginBinToolFileName)

	// upload to upstream.
	if err := m.upstreamOriginPluginBinTool.Store(ctx, fileiface.FileInfo{Name: pkgName}, uploadingFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package, failed to upload to upstream. err: %v", err)

		return nil, err
	}

	// get file.
	file, err := m.upstreamOriginPluginBinTool.GetFile(ctx, pkgName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package. failed to get file from upstream. err: %v", err)

		return nil, err
	}

	// get info.
	detail.FileInfo = file.Info()

	// check if release existed.
	existed, err := m.storageRelease.ExistReleasePluginBinTool(ctx, types.Generation2)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package. failed to check if release existed. err: %v", err)

		return nil, err
	}
	detail.Existed = existed

	// create the upload record.
	uploadID, err := m.storageUpload.CreatePluginBinToolUpload(ctx, &types.Upload{
		Category:  types.UploadCategoryOriginPluginBinTool,
		SavedName: pkgName,
	})
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to upload origin plugin bintool package, failed to create upload: %v", err)

		return nil, err
	}
	detail.UploadID = uploadID

	m.logger.InfoCtxf(ctx,
		"uploaded origin plugin bintool package to upstream. file-name(%s)", pkgName)

	return detail, nil
}

func checkGen2OriginPluginBinToolPkg(file io.ReadCloser) (*types.OriginPluginBinToolPkgDetail, error) {
	agentPlatforms := make(map[string]platform.Platform)
	proxyPlatforms := make(map[string]platform.Platform)

	if err := checkTgz(file, []tgzReadRule{
		{
			filePath: []string{tgzPathNameAny1, "linux_amd64", tgzPathNameAny2},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}
				agentPlatforms[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "linux_arm64", tgzPathNameAny2},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchArm64}
				agentPlatforms[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "windows_amd64", tgzPathNameAny2},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSWindows, Arch: criteria.CPUArchAmd64}
				agentPlatforms[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "darwin_amd64", tgzPathNameAny2},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSDarwin, Arch: criteria.CPUArchAmd64}
				agentPlatforms[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "linux_amd64", tgzPathNameAny2},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchAmd64}
				proxyPlatforms[plat.String()] = plat

				return nil
			},
		},
		{
			filePath: []string{tgzPathNameAny1, "linux_arm64", tgzPathNameAny2},
			callback: func(_ []string, _ io.Reader) error {
				plat := platform.Platform{OS: criteria.OSLinux, Arch: criteria.CPUArchArm64}
				proxyPlatforms[plat.String()] = plat

				return nil
			},
		},
	}); err != nil {
		return nil, err
	}

	detail := new(types.OriginPluginBinToolPkgDetail)
	detail.Platforms = make([]platform.Platform, 0)
	for _, plat := range agentPlatforms {
		detail.Platforms = append(detail.Platforms, plat)
	}

	return detail, nil
}

// PublishReleasePluginBinTool generates release plugin bintool by upload-id.
func (m *Manager) PublishReleasePluginBinTool(ctx contextx.IUserContext, uploadID string) error {
	up, err := m.storageUpload.GetPluginBinToolUpload(ctx, uploadID)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to get upload(%s). err: %v", uploadID, err)

		return err
	}

	if up.Category != types.UploadCategoryOriginPluginBinTool {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, invalid category. err: %s", up.Category)

		return errors.New("invalid category")
	}

	// get origin file.
	file, err := m.upstreamOriginPluginBinTool.GetFile(ctx, up.SavedName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to get file(%s). err: %v", up.SavedName, err)

		return err
	}

	// get origin content.
	content, err := file.Content(ctx)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to get content. file(%s). err: %v",
			up.SavedName, err)

		return err
	}

	// generate release file.
	generatedFile, err := m.generatePluginBinToolPkg(ctx, content)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to generate plugin bintool pkg. err: %v", err)

		return err
	}
	defer func() {
		_ = generatedFile.Close()
	}()

	// upload to upstream.
	if err = m.upstreamReleasePluginBinTool.Store(
		ctx, fileiface.FileInfo{Name: releasePluginBinToolFileName}, generatedFile, true); err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to upload to upstream. err: %v", err)

		return err
	}

	// get release file.
	releaseFile, err := m.upstreamReleasePluginBinTool.GetFile(ctx, releasePluginBinToolFileName)
	if err != nil {
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to get release file. err: %v", err)

		return err
	}

	// get release info.
	releaseInfo := releaseFile.Info()

	// upsert release plugin bintool.
	if err = m.storageRelease.UpsertReleasePluginBinTool(ctx, types.ReleasePluginBinTool{
		Release: types.Release{
			Generation:   types.Generation2,
			Type:         types.ReleaseTypePluginBinTool,
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
		m.logger.ErrorCtxf(ctx, "failed to publish release plugin bintool, failed to upsert release plugin bintool: %v", err)

		return err
	}

	m.logger.InfoCtxf(ctx, "generated and published release plugin bintool. file(%s), md5(%s)", releaseInfo.Name, releaseInfo.MD5)

	return nil
}

func (m *Manager) generatePluginBinToolPkg(ctx context.Context, sourceFile io.ReadCloser) (io.ReadCloser, error) {
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
			{targetFilePath: []string{"plugin_bintool"}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{"plugin_bintool", "linux_amd64"}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{"plugin_bintool", "linux_arm64"}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{"plugin_bintool", "darwin_amd64"}, targetFileMode: tgzModeDir},
			{targetFilePath: []string{"plugin_bintool", "windows_amd64"}, targetFileMode: tgzModeDir},
		},
		[]*tgzWriteRuleStream{{
			sourceFile: sourceFile,
			fileRules: []tgzWriteRuleFile{
				{
					sourceFilePath: []string{tgzPathNameAny1, "linux_amd64", tgzPathNameAny2},
					targetFilePath: []string{"plugin_bintool", "linux_amd64", tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, "linux_arm64", tgzPathNameAny2},
					targetFilePath: []string{"plugin_bintool", "linux_arm64", tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, "darwin_amd64", tgzPathNameAny2},
					targetFilePath: []string{"plugin_bintool", "darwin_amd64", tgzPathNameAny2},
					targetFileMode: tgzModeFile,
				},
				{
					sourceFilePath: []string{tgzPathNameAny1, "windows_amd64", tgzPathNameAny2},
					targetFilePath: []string{"plugin_bintool", "windows_amd64", tgzPathNameAny2},
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
