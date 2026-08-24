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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNamePackageExternalPluginV2Import is the external v2 package import operation name.
	OperDefNamePackageExternalPluginV2Import = "package_external_plugin_v2_import"
)

// NewOperPackageExternalPluginV2Import creates an external v2 package import operation.
func NewOperPackageExternalPluginV2Import(param OperParamPackageExternalPluginV2Import) operation.Definition {
	return &operPackageExternalPluginV2Import{param: param}
}

type operPackageExternalPluginV2Import struct {
	param OperParamPackageExternalPluginV2Import
}

// OperParamPackageExternalPluginV2Import defines external v2 package import parameters.
type OperParamPackageExternalPluginV2Import struct {
	TenantID string `json:"tenant_id"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

func (oper *operPackageExternalPluginV2Import) Name() string {
	return OperDefNamePackageExternalPluginV2Import
}

func (oper *operPackageExternalPluginV2Import) DisplayNameZh() string {
	return "导入外部插件V2资源包"
}

func (oper *operPackageExternalPluginV2Import) DisplayNameEn() string {
	return "Import External Plugin V2 Package"
}

func (oper *operPackageExternalPluginV2Import) Version() string {
	return "1.0.0" // nolint: goconst
}

func (oper *operPackageExternalPluginV2Import) Description() string {
	return "download, upload, publish and update package visibility" // nolint: goconst
}

func (oper *operPackageExternalPluginV2Import) ActionDefNames() []string {
	return []string{
		ActionNamePackageImportExternalPluginV2PkgFetchAndUpload,
		ActionNamePackagePublishExternalPluginV2Pkg,
		ActionNamePackageReleasePluginEnable,
		ActionNamePackageReleasePluginHidden,
	}
}

func (oper *operPackageExternalPluginV2Import) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNamePackageImportExternalPluginV2PkgFetchAndUpload: true,
			ActionNamePackagePublishExternalPluginV2Pkg:              true,
			ActionNamePackageReleasePluginEnable:                     true,
			ActionNamePackageReleasePluginHidden:                     true,
		},
	}
}

func (oper *operPackageExternalPluginV2Import) ExtraExecutionName() string {
	return ""
}
