/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const operDefNamePackageExportPlugin = "package_export_plugin"

// NewOperPackageExportPlugin creates a plugin package export operation.
func NewOperPackageExportPlugin(param OperParamPackageExportPlugin) operation.Definition {
	return &operPackageExportPlugin{param: param}
}

type operPackageExportPlugin struct {
	param OperParamPackageExportPlugin
}

// OperParamPackageExportPlugin defines the plugin package export operation parameters.
type OperParamPackageExportPlugin struct {
	TenantID string `json:"tenant_id"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

func (oper *operPackageExportPlugin) Name() string {
	return operDefNamePackageExportPlugin
}

func (oper *operPackageExportPlugin) DisplayNameZh() string {
	return "导出插件资源包"
}

func (oper *operPackageExportPlugin) DisplayNameEn() string {
	return "Export Plugin Package"
}

func (oper *operPackageExportPlugin) Version() string {
	return "v1.0.0"
}

func (oper *operPackageExportPlugin) Description() string {
	return "export origin plugin package"
}

func (oper *operPackageExportPlugin) ActionDefNames() []string {
	return []string{
		ActionNamePackageExportPlugin,
	}
}

func (oper *operPackageExportPlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNamePackageExportPlugin: true,
		},
	}
}

func (oper *operPackageExportPlugin) ExtraExecutionName() string {
	return ""
}
