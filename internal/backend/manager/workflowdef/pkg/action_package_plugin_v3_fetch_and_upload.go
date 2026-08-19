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
	// ActionNamePackageImportPluginV3PkgFetchAndUpload is the package import upload action name.
	ActionNamePackageImportPluginV3PkgFetchAndUpload = "package_import_plugin_v3_pkg_fetch_and_upload"
)

// NewActionPackageImportPluginV3PkgFetchAndUpload creates an upload action for package import.
func NewActionPackageImportPluginV3PkgFetchAndUpload(capability *Capability) action.Definition {
	return &actionPackageImportPluginV3PkgFetchAndUpload{
		fileHandler:    capability.FileHandler,
		storagePackage: capability.StoragePackage,
	}
}

type actionPackageImportPluginV3PkgFetchAndUpload struct {
	fileHandler    file.IHandler
	storagePackage pkgStg.IStorage
}

// ActionPackageImportPluginV3PkgFetchAndUpload defines the action's param.
type ActionPackageImportPluginV3PkgFetchAndUpload struct {
	pkgUtils.PackageActionStandardParam
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) Name() string {
	return ActionNamePackageImportPluginV3PkgFetchAndUpload
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) DisplayNameZh() string {
	return "获取并上传插件V3资源包"
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) DisplayNameEn() string {
	return "Fetch And Upload Plugin V3 Package"
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) Version() string {
	return "v1.0.0" // nolint: goconst
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) Description() string {
	return "fetch package and upload origin plugin v3 package"
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) Timeout() time.Duration {
	return 30 * time.Minute // nolint: mnd
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) Tags() []action.Tag {
	return []action.Tag{}
}

func (act *actionPackageImportPluginV3PkgFetchAndUpload) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionPackageImportPluginV3PkgFetchAndUpload)
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
		Zh("上传插件V3安装包, 文件名: %s", opts.FileName).
		En("upload plugin v3 package, file-name: %s", opts.FileName).
		Info()

	detail, err := act.fileHandler.UploadOriginPluginV3(nCtx, opts.FileName, fileContent, true)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("file-name", opts.FileName).Error("failed to upload origin plugin v3 package")

		return fmt.Errorf("failed to upload origin plugin v3 package: %w", err)
	}

	// Persist the upload result into the package deployment record.
	info := std.DeployInfo()
	info.Upload = types.PackageDeploymentUploadInfo{
		UploadID:  detail.UploadID,
		Name:      detail.PluginPkgName,
		Version:   detail.Version,
		Platforms: detail.Platforms,
	}

	logger.G.Sys().Ctx(nCtx).With("upload-id", detail.UploadID, "plugin-name", detail.PluginPkgName).Info("uploaded origin plugin package")

	std.InstanceData().Log().
		Zh("上传V3安装包成功，上传ID: %s，插件包: %s，版本: %s", detail.UploadID, detail.PluginPkgName, detail.Version).
		En("successfully uploaded v3 package, upload-id: %s, plugin-name: %s, version: %s", detail.UploadID, detail.PluginPkgName, detail.Version).
		Info()

	return nil
}
