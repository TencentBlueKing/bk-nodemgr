/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package plugin provides the plugin installer common types.
package plugin

import "github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"

const (
	// StepGeneral this is the step to general.
	StepGeneral logger.Step = "general"

	// StepInstallPlugin this is the step to install plugin.
	StepInstallPlugin logger.Step = "install_plugin"

	// StepUninstallPlugin this is the step to uninstall plugin.
	StepUninstallPlugin logger.Step = "uninstall_plugin"

	// StepUpgradePlugin this is the step to upgrade plugin.
	StepUpgradePlugin logger.Step = "upgrade_plugin"

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

	// StepCheckDeploy this is the step to check this nodemgr plugin is deploy or not.
	StepCheckDeploy logger.Step = "check_deploy"

	// StepCleanTmp this is the step to clean tmp.
	StepCleanTmp logger.Step = "clean_tmp"

	// StepDebugPlugin this is the step to start a debug plugin process.
	StepDebugPlugin logger.Step = "debug_plugin"
)
