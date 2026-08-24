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
	// OperDefNamePackagePluginV2Import is the official v2 package import operation name.
	OperDefNamePackagePluginV2Import = "package_plugin_v2_import"
)

// NewOperPackagePluginV2Import creates an official v2 package import operation.
func NewOperPackagePluginV2Import(param OperParamPackagePluginV2Import) operation.Definition {
	return &operPackagePluginV2Import{param: param}
}

type operPackagePluginV2Import struct {
	param OperParamPackagePluginV2Import
}

// OperParamPackagePluginV2Import defines official v2 package import parameters.
type OperParamPackagePluginV2Import struct {
	TenantID string `json:"tenant_id"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

func (oper *operPackagePluginV2Import) Name() string {
	return OperDefNamePackagePluginV2Import
}

func (oper *operPackagePluginV2Import) DisplayNameZh() string {
	return "导入插件V2资源包"
}

func (oper *operPackagePluginV2Import) DisplayNameEn() string {
	return "Import Plugin V2 Package"
}

func (oper *operPackagePluginV2Import) Version() string {
	return "1.0.0" // nolint: goconst
}

func (oper *operPackagePluginV2Import) Description() string {
	return "download, upload, publish and update package visibility"
}

func (oper *operPackagePluginV2Import) ActionDefNames() []string {
	return []string{
		ActionNamePackageImportPluginV2PkgFetchAndUpload,
		ActionNamePackagePublishPluginV2Pkg,
		ActionNamePackageReleasePluginEnable,
		ActionNamePackageReleasePluginHidden,
	}
}

func (oper *operPackagePluginV2Import) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNamePackageImportPluginV2PkgFetchAndUpload: true,
			ActionNamePackagePublishPluginV2Pkg:              true,
			ActionNamePackageReleasePluginEnable:             true,
			ActionNamePackageReleasePluginHidden:             true,
		},
	}
}

func (oper *operPackagePluginV2Import) ExtraExecutionName() string {
	return ""
}
