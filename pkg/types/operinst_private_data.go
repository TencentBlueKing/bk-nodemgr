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

const (
	// PDKeyOfflineInstallResult is key for offline install result data in oper inst private data.
	// The value is a JSON string of installer.data.json submitted by the user.
	PDKeyOfflineInstallResult string = "offline_install_result"

	// PDKeyInstallerReportStatus is key for installer report status in oper inst private data.
	PDKeyInstallerReportStatus string = "installer_report_status"

	// PDKeyInstallerReportAgentID is key for installer report agent id in oper inst private data.
	PDKeyInstallerReportAgentID string = "installer_report_agent_id"

	// PDKeyReportDetectInfo is key for report detect info in oper inst private data.
	// will be used in manual and relay-based install.
	PDKeyReportDetectInfo string = "report_detect_info"

	// PDKeyManualInstallBootstrapCommandBash is key for manual install bootstrap bash command in oper inst private data.
	PDKeyManualInstallBootstrapCommandBash string = "manual_install_bootstrap_command_bash"
	// PDKeyManualInstallBootstrapCommandBat is key for manual install bootstrap bat command in oper inst private data.
	PDKeyManualInstallBootstrapCommandBat string = "manual_install_bootstrap_command_bat"

	// PDKeyManualInstallInstallerPath is key for manual install installer path in oper inst private data.
	PDKeyManualInstallInstallerPath string = "manual_install_installer_path"

	// PDKeyManualInstallExecCommand is key for manual install exec command in oper inst private data.
	PDKeyManualInstallExecCommand string = "manual_install_exec_command"

	// PDKeyManualInstallCallbackAddress is key for manual install callback address in oper inst private data.
	PDKeyManualInstallCallbackAddress string = "manual_install_callback_address"

	// PDKeyManualInstallDownloadAddress is key for manual install download address in oper inst private data.
	PDKeyManualInstallDownloadAddress string = "manual_install_download_address"

	// PDKeyManualInstallActionNameReportDetectInfo is key for manual install action name report detect info in oper inst private data.
	PDKeyManualInstallActionNameReportDetectInfo = "manual_install_action_name_report_detect_info"

	// PDKeyManualInstallActionNameGetExecCommand is key for manual install action name get exec command in oper inst private data.
	PDKeyManualInstallActionNameGetExecCommand = "manual_install_action_name_get_exec_command"

	// PDKeySubWorkflowRefs is the oper-inst action private_data key for spawned child workflow refs.
	// The value is a JSON string of []SubWorkflowRef.
	PDKeySubWorkflowRefs string = "sub_workflow_refs"
)

// PDDetectInfo is report detect info.
type PDDetectInfo struct {
	OsType  string `json:"os_type"`
	CPUArch string `json:"cpu_arch"`
	RunDir  string `json:"run_dir"`
	ErrMsg  string `json:"err_msg"`
}

// WorkflowDomain identifies which workflow API domain a workflow belongs to.
type WorkflowDomain string

const (
	// WorkflowDomainNode indicates the node workflow domain.
	WorkflowDomainNode WorkflowDomain = "node"
	// WorkflowDomainPlugin indicates the plugin workflow domain.
	WorkflowDomainPlugin WorkflowDomain = "plugin"
)

// SubWorkflowRef describes a child workflow reference stored in action private_data.
type SubWorkflowRef struct {
	WorkflowID     string         `json:"workflow_id" bson:"workflow_id"`
	WorkflowDomain WorkflowDomain `json:"workflow_domain" bson:"workflow_domain"`
}
