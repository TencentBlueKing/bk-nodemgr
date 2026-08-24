/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"errors"
	"fmt"
	"io"
	"time"

	pkgUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pkg/utils"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePackageImportExternalPluginV2PkgFetchAndUpload is the external v2 fetch/upload action name.
	ActionNamePackageImportExternalPluginV2PkgFetchAndUpload = "package_import_external_plugin_v2_pkg_fetch_and_upload"
)

// NewActionPackageImportExternalPluginV2PkgFetchAndUpload creates the external v2 upload action.
func NewActionPackageImportExternalPluginV2PkgFetchAndUpload(capability *Capability) action.Definition {
	return &actionPackageImportExternalPluginV2PkgFetchAndUpload{
		fileHandler:    capability.FileHandler,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackageImportExternalPluginV2PkgFetchAndUpload struct {
	fileHandler    file.IHandler
	storagePackage pkgStg.IStorage
}

// ActionPackageImportExternalPluginV2PkgFetchAndUpload defines the action's parameters.
type ActionPackageImportExternalPluginV2PkgFetchAndUpload struct {
	pkgUtils.PackageActionStandardParam
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) Name() string {
	return ActionNamePackageImportExternalPluginV2PkgFetchAndUpload
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) DisplayNameZh() string {
	return "获取并上传外部插件V2资源包"
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) DisplayNameEn() string {
	return "Fetch And Upload External Plugin V2 Package"
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) Version() string {
	return "v1.0.0" // nolint: goconst
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) Description() string {
	return "fetch package and upload origin external plugin v2 package"
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) Timeout() time.Duration {
	return 30 * time.Minute // nolint: mnd
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackageImportExternalPluginV2PkgFetchAndUpload) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionPackageImportExternalPluginV2PkgFetchAndUpload)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pkgUtils.NewPackageActionStandarder(act.storagePackage)
	if err := std.Initialize(ctx, param.PackageActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	nCtx := std.Context()
	opts := std.DeployInfo().ImportPluginPkgOptions
	if opts.FileSource == "" {
		return errors.New("file source is empty")
	}

	var fileContent io.Reader
	switch opts.FileSourceType {
	case types.FileSourceTypeDownload:
		std.InstanceData().Log().
			Zh("正在下载安装包，文件名: %s", opts.FileName).
			En("downloading package, file-name: %s", opts.FileName).
			Info()

		tmpFile, err := act.fileHandler.DownloadRemoteFile(nCtx, opts.FileName, opts.FileSource, opts.MD5)
		if err != nil {
			return fmt.Errorf("download remote file failed: %w", err)
		}
		defer func() { _ = tmpFile.Data.Close() }()

		std.InstanceData().Log().
			Zh("下载安装包成功，文件名: %s", opts.FileName).
			En("downloaded package successfully, file-name: %s", opts.FileName).
			Info()

		fileContent = tmpFile.Data
	default:
		return fmt.Errorf("unsupported file source type: %s", opts.FileSourceType)
	}

	std.InstanceData().Log().
		Zh("上传外部插件V2安装包, 文件名: %s", opts.FileName).
		En("upload external plugin v2 package, file-name: %s", opts.FileName).
		Info()

	detail, err := act.fileHandler.UploadOriginExternalPluginV2(nCtx, opts.FileName, fileContent, true)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("file-name", opts.FileName).
			Error("failed to upload origin external plugin v2 package")

		return fmt.Errorf("failed to upload origin external plugin v2 package: %w", err)
	}

	std.DeployInfo().Upload = types.PackageDeploymentUploadInfo{
		UploadID:  detail.UploadID,
		Name:      detail.PluginPkgName,
		Version:   detail.Version,
		Platforms: detail.Platforms,
	}

	logger.G.Sys().Ctx(nCtx).With("upload-id", detail.UploadID, "plugin-name", detail.PluginPkgName).
		Info("uploaded origin external plugin v2 package")

	std.InstanceData().Log().
		Zh("上传外部插件V2安装包成功，上传ID: %s，插件包: %s，版本: %s", detail.UploadID, detail.PluginPkgName, detail.Version).
		En("successfully uploaded external plugin v2 package, upload-id: %s, plugin-name: %s, version: %s",
			detail.UploadID, detail.PluginPkgName, detail.Version).
		Info()

	return nil
}
