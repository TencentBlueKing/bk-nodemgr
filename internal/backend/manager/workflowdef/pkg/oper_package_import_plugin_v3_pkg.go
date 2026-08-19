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
	// OperDefNamePackagePluginV3Import is the package import operation definition name.
	OperDefNamePackagePluginV3Import = "package_plugin_v3_import"
)

// NewOperPackagePluginV3Import creates a package import operation definition.
func NewOperPackagePluginV3Import(param OperParamPackagePluginV3Import) operation.Definition {
	return &operPackagePluginV3Import{param: param}
}

type operPackagePluginV3Import struct {
	param OperParamPackagePluginV3Import
}

// OperParamPackagePluginV3Import defines package import operation parameters.
type OperParamPackagePluginV3Import struct {
	TenantID string `json:"tenant_id"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

func (oper *operPackagePluginV3Import) Name() string {
	return OperDefNamePackagePluginV3Import
}

func (oper *operPackagePluginV3Import) DisplayNameZh() string {
	return "导入插件V3资源包"
}

func (oper *operPackagePluginV3Import) DisplayNameEn() string {
	return "Import Plugin V3 Package"
}

func (oper *operPackagePluginV3Import) Version() string {
	return "1.0.0" // nolint: goconst
}

func (oper *operPackagePluginV3Import) Description() string {
	return "download, upload, publish and update package visibility"
}

func (oper *operPackagePluginV3Import) ActionDefNames() []string {
	return []string{
		ActionNamePackageImportPluginV3PkgFetchAndUpload,
		ActionNamePackagePublishPluginV3Pkg,
		ActionNamePackageReleasePluginEnable,
		ActionNamePackageReleasePluginHidden,
	}
}

func (oper *operPackagePluginV3Import) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNamePackageImportPluginV3PkgFetchAndUpload: true,
			ActionNamePackagePublishPluginV3Pkg:              true,
			ActionNamePackageReleasePluginEnable:             true,
			ActionNamePackageReleasePluginHidden:             true,
		},
	}
}

func (oper *operPackagePluginV3Import) ExtraExecutionName() string {
	return ""
}
