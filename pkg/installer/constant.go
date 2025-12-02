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
)

const (
	// WaitInstallerCompleteReportStatusKey defines the report status key.
	WaitInstallerCompleteReportStatusKey = "installer_report_status"

	// WaitInstallerCompleteReportAgentIDKey defines the report agent id key.
	WaitInstallerCompleteReportAgentIDKey = "installer_report_agent_id"
)
