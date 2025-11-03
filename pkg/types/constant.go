/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import "github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"

const (
	// ReleaseVersionPluginBinTool defines the version of plugin bintool.
	ReleaseVersionPluginBinTool = "default"

	// ReleaseVersionBinTool defines the version of bintool.
	ReleaseVersionBinTool = "default"

	// ReleaseVersionCert defines the version of cert.
	ReleaseVersionCert = "default"
)

const (
	// ReleaseOSTypePluginBinTool defines the os type of plugin bintool.
	ReleaseOSTypePluginBinTool = criteria.OSUnknown

	// ReleaseOSTypeBinTool defines the os type of bintool.
	ReleaseOSTypeBinTool = criteria.OSUnknown

	// ReleaseOSTypeCert defines the os type of cert.
	ReleaseOSTypeCert = criteria.OSUnknown
)

const (
	// ReleaseCPUArchPluginBinTool defines the cpu arch of plugin bintool.
	ReleaseCPUArchPluginBinTool = criteria.CPUArchUnknown

	// ReleaseCPUArchBinTool defines the cpu arch of bintool.
	ReleaseCPUArchBinTool = criteria.CPUArchUnknown

	// ReleaseCPUArchCert defines the cpu arch of cert.
	ReleaseCPUArchCert = criteria.CPUArchUnknown
)

const (
	// ReleaseNameAgent defines the name of agent.
	ReleaseNameAgent = "agent"

	// ReleaseNameProxy defines the name of proxy.
	ReleaseNameProxy = "proxy"

	// ReleaseNameBinTool defines the name of bintool.
	ReleaseNameBinTool = "bintool"

	// ReleaseNamePluginBinToolV2 defines the name of plugin bintool v2.
	ReleaseNamePluginBinToolV2 = "plugin_bintool_v2"

	// ReleaseNamePluginBinToolV3 defines the name of plugin bintool v3.
	ReleaseNamePluginBinToolV3 = "plugin_bintool_v3"

	// ReleaseNameCert defines the name of cert.
	ReleaseNameCert = "cert"
)
