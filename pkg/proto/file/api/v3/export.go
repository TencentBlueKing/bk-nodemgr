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

package v3

import (
	"errors"
)

// Validate check request body.
func (x *ExportPrepareOriginPluginPackageReq) Validate() error {
	if x.GetPluginPkgName() == "" {
		return errors.New("plugin_pkg_name is required")
	}

	if x.GetPluginPkgVersion() == "" {
		return errors.New("plugin_pkg_version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ExportPrepareOriginPluginPackageReq) AutoConvert() {}

// Validate checks the request for an origin plugin package download address.
func (x *ExportGetOriginPluginPackageDownloadAddressReq) Validate() error {
	if x.GetExportId() == "" {
		return errors.New("export_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ExportGetOriginPluginPackageDownloadAddressReq) AutoConvert() {}

// Validate check request body.
func (x *ExportOriginPluginPackageReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ExportOriginPluginPackageReq) AutoConvert() {}

// ConvertResult converts export result to response.
func (x *ExportPrepareOriginPluginPackageResp) ConvertResult(exportID string) {
	x.Data = &ExportPrepareOriginPluginPackageResp_Data{
		ExportId: exportID,
	}
}

// ConvertResult converts a download address result to a response.
func (x *ExportGetOriginPluginPackageDownloadAddressResp) ConvertResult(downloadURL string, expiredAt int64) {
	x.Data = &ExportGetOriginPluginPackageDownloadAddressResp_Data{
		DownloadUrl:          downloadURL,
		DownloadUrlExpiredAt: expiredAt,
	}
}
