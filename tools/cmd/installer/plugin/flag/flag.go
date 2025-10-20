/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package flag defines the flags for plugins.
package flag

const (
	// DeployEnv defines the deploy env flag.
	DeployEnv = "deploy_env"
	// DeployEnvS defines the deploy env short flag.
	DeployEnvS = "e"

	// BaseDeployDir defines the base dir flag.
	BaseDeployDir = "base_deploy_dir"

	// BaseWorkDir defines the base work dir flag.
	BaseWorkDir = "base_work_dir"

	// LogDir defines the log dir flag.
	LogDir = "log_dir"

	// Status defines the status flag.
	Status = "status"

	// PluginGroup defines the plugin group flag.
	PluginGroup = "plugin_group"

	// PluginName defines the plugin name flag.
	PluginName = "plugin_name"

	// PkgFile defines the pkg file flag.
	PkgFile = "pkg_file"

	// DownloadSvrAddr defines the download server address flag.
	DownloadSvrAddr = "dlsvr_addr"

	// CallbackSvrAddr defines the callback server address flag.
	CallbackSvrAddr = "cbsvr_addr"

	// DeployToken defines the deploy token flag.
	DeployToken = "deploy_token"

	// OperInstID defines the operation instance id flag.
	OperInstID = "oper_inst_id"

	// PluginVersion defines the plugin version flag.
	PluginVersion = "plugin_version"
)
