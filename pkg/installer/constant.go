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

// Package installer defines the installer constants.
package installer

const (
	// ServerAddrSeparator is the separator used to join/split multiple server addresses.
	// This separator is used in the installer tools to handle comma-separated server addresses.
	ServerAddrSeparator = ","

	// LogFieldSeparator is the delimiter between fields in installer log lines.
	// Format: "YYYY/MM/DD HH:MM:SS | LEVEL | step_name | message"
	// Used by tools/pkg/logger to format and by the backend to parse installer logs.
	LogFieldSeparator = "|"

	// LogFieldCount is the expected number of fields in a well-formed installer log line
	// (timestamp | level | step | message).
	LogFieldCount = 4

	// StatusFileName is the name of the installer status file written by the tools and read by the backend.
	// SYNC: tools/internal/installer/node/statusreporter writes this file;
	//       internal/backend/manager/workflowdef/node/action_wait_installer_complete reads it via SSH.
	StatusFileName = "installer.status.json"

	// DataFileName is the name of the installer data file written by the tools and read by the backend.
	// SYNC: tools/internal/installer/node/datareporter writes this file;
	//       internal/backend/manager/workflowdef/node/action_wait_installer_complete reads it via SSH.
	DataFileName = "installer.data.json"

	// --- Offline proxy install bundle (tar.gz layout and filenames) ---
	// SYNC: internal/application/router/api-v3/node/workflow/offline.go (tar writer);
	//       internal/backend/router/api-v3/node/workflow/offline.go (install.sh, config map keys);
	//       front/src/pages/node/history/guide.vue (package stem prefix in UX; keep in sync manually).

	// OfflinePackageNamePrefix is the prefix for the offline bundle directory / download stem
	// ({prefix}-{network_area_id}-{ip_slug}).
	OfflinePackageNamePrefix = "bk-nodemgr-proxy-offline"

	// OfflinePkgRelPathData is the path segment under the bundle root for the data subtree: release package,
	// precheck JSON, and config dir. install.sh copies this tree into installer DataDir before --skip_download.
	OfflinePkgRelPathData = "data"

	// OfflinePkgRelPathConfig is the path under the bundle root for rendered GSE JSON configs
	// (tar prefix: {pkgName}/data/config).
	OfflinePkgRelPathConfig = "data/config"

	// OfflinePkgInstallScriptName is the entrypoint shell script at the bundle root.
	OfflinePkgInstallScriptName = "install.sh"

	// OfflinePkgMetadataFileName is instance metadata JSON at the bundle root.
	OfflinePkgMetadataFileName = "metadata.json"

	// OfflinePkgPrecheckFileName is the precheck list JSON inside OfflinePkgRelPathData.
	OfflinePkgPrecheckFileName = "precheck.json"

	// OfflineGse* file names match keys under OfflinePkgRelPathConfig and GSE tooling expectations.
	OfflineGseAgentConfFileName     = "gse_agent.conf"
	OfflineGseFileProxyConfFileName = "gse_file_proxy.conf"
	OfflineGseDataProxyConfFileName = "gse_data_proxy.conf"
)
