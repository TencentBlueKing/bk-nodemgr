/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package provider

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

const (
	// IAMAttrID is the IAM attribute key for resource ID.
	IAMAttrID = "id"

	// IAMAttrDisplayName is the IAM attribute key for resource display name.
	IAMAttrDisplayName = "display_name"

	// IAMAttrBkIAMPath is the IAM attribute key for BlueKing IAM path.
	IAMAttrBkIAMPath = "_bk_iam_path_"
)

// packageTypeDisplayNames maps ReleaseType to Chinese display names.
func packageTypeDisplayNames() map[types.ReleaseType]string {
	return map[types.ReleaseType]string{
		types.ReleaseTypeAgent:         "Agent 包",
		types.ReleaseTypeProxy:         "Proxy 包",
		types.ReleaseTypeCert:          "证书包",
		types.ReleaseTypeBinTool:       "工具包",
		types.ReleaseTypePluginBinTool: "插件工具包",
		types.ReleaseTypePlugin:        "插件包",
	}
}

// allPackageTypes returns the ordered list of all package types.
func allPackageTypes() []types.ReleaseType {
	return []types.ReleaseType{
		types.ReleaseTypeAgent,
		types.ReleaseTypeProxy,
		types.ReleaseTypeCert,
		types.ReleaseTypeBinTool,
		types.ReleaseTypePluginBinTool,
		types.ReleaseTypePlugin,
	}
}
