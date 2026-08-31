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

package types

import (
	"errors"
)

// ExportPrepareParam defines the export prepare param.
type ExportPrepareParam struct {
	PluginPkgName       string
	PluginPkgVersion    string
	UploadOriginPkgType UploadCategory
	UploadIDs           []string
	FileNameSuffix      string
}

// ExportTokenPurpose identifies the authenticated export token contract.
const ExportTokenPurpose = "bk-nodemgr/package-export/v1"

// ExportDownloadTokenPayload defines the shared export download token payload.
// Short JSON tags keep the serialized payload, and therefore the Base64URL token, compact.
// They are part of the cross-service token contract and must remain stable.
type ExportDownloadTokenPayload struct {
	TenantID  string `json:"t"`
	Filename  string `json:"f"`
	ExpiresAt int64  `json:"e"`
}

// Validate checks the structural and value correctness of the payload.
func (payload ExportDownloadTokenPayload) Validate() error {
	if payload.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if payload.Filename == "" {
		return errors.New("filename is required")
	}
	if payload.ExpiresAt <= 0 {
		return errors.New("expires_at must be positive")
	}

	return nil
}
