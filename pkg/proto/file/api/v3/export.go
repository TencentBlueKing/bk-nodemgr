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
	"fmt"

	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check request body.
func (x *ExportPrepareOriginPluginPackageReq) Validate() error {
	if x.GetPluginPkgName() == "" {
		return errors.New("plugin_pkg_name is required")
	}

	if x.GetPluginPkgVersion() == "" {
		return errors.New("plugin_pkg_version is required")
	}

	if x.GetFileNameSuffix() == "" {
		return errors.New("workflow_id is required")
	}

	if len(x.GetUploadIds()) == 0 {
		return errors.New("upload_ids is required")
	}

	switch types.UploadCategory(x.GetUploadOriginPkgType()) {
	case types.UploadCategoryOriginPluginV2,
		types.UploadCategoryOriginExternalPluginV2,
		types.UploadCategoryOriginPluginV3:
		return nil
	default:
		return fmt.Errorf("unsupported upload origin pkg type: %s", x.GetUploadOriginPkgType())
	}
}

// AutoConvert auto convert.
func (x *ExportPrepareOriginPluginPackageReq) AutoConvert() {}

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
func (x *ExportPrepareOriginPluginPackageResp) ConvertResult(info fileiface.FileInfo, address string) {
	x.Data = &ExportPrepareOriginPluginPackageResp_Data{
		Filename: info.Name,
		Size:     info.Size,
		Md5:      info.MD5,
		Address:  address,
	}
}
