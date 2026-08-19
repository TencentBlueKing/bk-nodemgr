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

// Validate check request body.
func (x *PublishReleaseAgentReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleaseAgentReq) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleaseProxyReq) Validate() error {
	if x.GetUploadOriginPkgType() == "" {
		return errors.New("upload_origin_pkg_type is required")
	}

	switch types.UploadCategory(x.GetUploadOriginPkgType()) {
	case types.UploadCategoryOriginProxy, types.UploadCategoryOriginServer:
		return nil
	default:
		return fmt.Errorf("unsupported upload origin pkg type: %s", x.GetUploadOriginPkgType())
	}
}

// AutoConvert auto convert.
func (x *PublishReleaseProxyReq) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleaseCertReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleaseCertReq) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleaseBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleaseBinToolReq) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleasePluginBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleasePluginBinToolReq) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleasePluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleasePluginV2Req) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleaseExternalPluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleaseExternalPluginV2Req) AutoConvert() {
}

// Validate check request body.
func (x *PublishReleasePluginV3Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PublishReleasePluginV3Req) AutoConvert() {
}
