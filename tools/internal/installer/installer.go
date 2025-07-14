/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package installer provides the installer common types.
package installer

// Step define the ctl step.
type Step string

const (
	// StepInstallNode this is the step to install node.
	StepInstallNode Step = "install_node"

	// StepUpgradeNode this is the step to upgrade node.
	StepUpgradeNode Step = "upgrade_node"

	// StepDownloadFiles this is the step to download files.
	StepDownloadFiles Step = "download_files"

	// StepPreCheck this is the step to pre check.
	StepPreCheck Step = "pre_check"

	// StepReportData this is the step to report data.
	StepReportData Step = "report_data"

	// StepReportStatus this is the step to report status.
	StepReportStatus Step = "report_status"

	// StepStartNode this is the step to start node.
	StepStartNode Step = "start_node"

	// StepStopNode this is the step to stop node.
	StepStopNode Step = "stop_node"

	// StepRestartNode this is the step to restart node.
	StepRestartNode Step = "restart_node"

	// StepCheckDeploy this is the step to check this gse node is deploy or not.
	StepCheckDeploy Step = "check_deploy"

	// StepUninstallNode this is the step to uninstall node.
	StepUninstallNode Step = "uninstall_node"
)
