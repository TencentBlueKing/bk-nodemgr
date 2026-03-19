/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package installer defines the installer constants.
package installer

const (
	// NodeCmdFullInstall defines the installer cmd.
	NodeCmdFullInstall = "node full-install"

	// NodeCmdFullUpgrade defines the installer cmd.
	NodeCmdFullUpgrade = "node full-upgrade"

	// NodeCmdFullReconfig defines the installer cmd.
	NodeCmdFullReconfig = "node full-reconfig"

	// NodeCmdFullUninstall defines the installer cmd.
	NodeCmdFullUninstall = "node full-uninstall"

	// NodeCmdStepCleanTmp defines the installer cmd.
	NodeCmdStepCleanTmp = "node step clean-tmp"

	// NodeCmdStepRestart defines the installer cmd.
	NodeCmdStepRestart = "node step restart"

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
)
