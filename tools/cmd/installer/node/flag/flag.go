/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package flag defines the flags for the installer.
package flag

const (
	// DeployEnv defines the deploy env flag.
	DeployEnv = "deploy_env"
	// DeployEnvS defines the deploy env short flag.
	DeployEnvS = "e"

	// Generation defines the generation flag.
	Generation = "generation"
	// GenerationS defines the generation short flag.
	GenerationS = "g"

	// NodeRole defines the node role flag.
	NodeRole = "node_role"
	// NodeRoleS defines the node role short flag.
	NodeRoleS = "r"

	// BaseDeployDir defines the base dir flag.
	BaseDeployDir = "base_deploy_dir"

	// BaseWorkDir defines the base work dir flag.
	BaseWorkDir = "base_work_dir"

	// LogDir defines the log dir flag.
	LogDir = "log_dir"

	// LogToStd defines the log std flag.
	LogToStd = "log_to_std"

	// PreCheckListConf defines the pre-check list conf flag.
	PreCheckListConf = "pre_check_list_conf"

	// PkgFile defines the pkg file flag.
	PkgFile = "pkg_file"

	// NodeVersion defines the node version flag.
	NodeVersion = "node_version"

	// AgentID defines the agent id flag.
	AgentID = "agent_id"

	// DeployToken defines the deploy token flag.
	DeployToken = "deploy_token"

	// DownloadSvrAddr defines the download server address flag.
	DownloadSvrAddr = "dlsvr_addr"

	// CallbackSvrAddr defines the callback server address flag.
	CallbackSvrAddr = "cbsvr_addr"

	// Status defines the status flag.
	Status = "status"

	// OperInstID defines the operation instance id flag.
	OperInstID = "oper_inst_id"

	// Force defines the force flag.
	Force = "force"

	// Restart defines the restart flag.
	Restart = "restart"

	// SkipDownload defines the skip download flag.
	SkipDownload = "skip_download"
)
