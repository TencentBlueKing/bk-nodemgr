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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate validates the cleanup upload category.
func (x *CleanupUploadReq) Validate() error {
	if x.GetCategory() == "" {
		return errors.New("category is required")
	}

	switch x.ConvertToTypes() {
	case types.UploadCategoryOriginAgent,
		types.UploadCategoryOriginServer,
		types.UploadCategoryOriginProxy,
		types.UploadCategoryOriginCert,
		types.UploadCategoryOriginBinTool,
		types.UploadCategoryOriginPluginBinTool,
		types.UploadCategoryOriginPluginV2,
		types.UploadCategoryOriginExternalPluginV2,
		types.UploadCategoryOriginPluginV3:
		return nil
	default:
		return fmt.Errorf("unsupported upload category: %s", x.GetCategory())
	}
}

// AutoConvert applies cleanup upload request defaults.
func (x *CleanupUploadReq) AutoConvert() {}

// ConvertToTypes converts the cleanup upload category to the domain type.
func (x *CleanupUploadReq) ConvertToTypes() types.UploadCategory {
	return types.UploadCategory(x.GetCategory())
}

// Validate validates the cleanup release type.
func (x *CleanupReleaseReq) Validate() error {
	if x.GetReleaseType() == "" {
		return errors.New("release_type is required")
	}

	switch x.ConvertToTypes() {
	case types.ReleaseTypeAgent, types.ReleaseTypeProxy, types.ReleaseTypePlugin:
		return nil
	default:
		return fmt.Errorf("unsupported cleanup release type: %s", x.GetReleaseType())
	}
}

// AutoConvert applies cleanup release request defaults.
func (x *CleanupReleaseReq) AutoConvert() {}

// ConvertToTypes converts the cleanup release type to the domain type.
func (x *CleanupReleaseReq) ConvertToTypes() types.ReleaseType {
	return types.ReleaseType(x.GetReleaseType())
}

// Validate validates the empty cleanup export request.
func (x *CleanupExportReq) Validate() error { return nil }

// AutoConvert applies cleanup export request defaults.
func (x *CleanupExportReq) AutoConvert() {}
