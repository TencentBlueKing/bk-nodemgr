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

// Package node provides the node installer common types.
package node

import "github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"

const (
	// StepGeneral this is the step to general.
	StepGeneral logger.Step = "general"

	// StepInstallNode this is the step to install node.
	StepInstallNode logger.Step = "install_node"

	// StepUpgradeNode this is the step to upgrade node.
	StepUpgradeNode logger.Step = "upgrade_node"

	// StepDownloadFiles this is the step to download files.
	StepDownloadFiles logger.Step = "download_files"

	// StepFetchConfigs this is the step to fetch configs.
	StepFetchConfigs logger.Step = "fetch_configs"

	// StepPreCheck this is the step to pre check.
	StepPreCheck logger.Step = "pre_check"

	// StepReportData this is the step to report data.
	StepReportData logger.Step = "report_data"

	// StepReportStatus this is the step to report status.
	StepReportStatus logger.Step = "report_status"

	// StepStartNode this is the step to start node.
	StepStartNode logger.Step = "start_node"

	// StepStopNode this is the step to stop node.
	StepStopNode logger.Step = "stop_node"

	// StepRestartNode this is the step to restart node.
	StepRestartNode logger.Step = "restart_node"

	// StepCheckDeploy this is the step to check this gse node is deploy or not.
	StepCheckDeploy logger.Step = "check_deploy"

	// StepUninstallNode this is the step to uninstall node.
	StepUninstallNode logger.Step = "uninstall_node"

	// StepCleanTmp this is the step to clean tmp.
	StepCleanTmp logger.Step = "clean_tmp"

	// StepCleanup this is the step to cleanup old release packages.
	StepCleanup logger.Step = "cleanup"
)
