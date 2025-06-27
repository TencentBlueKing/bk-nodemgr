/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// CmdFlagPkgVersion this flag is used to specify the version of the package.
	CmdFlagPkgVersion = "pkg_version"
	// CmdFlagDownloadEndpoint this flag is used to specify the download endpoint.
	CmdFlagDownloadEndpoint = "download_endpoint"
	// CmdFlagCallbackEndpoint this flag is used to specify the callback endpoint.
	CmdFlagCallbackEndpoint = "callback_endpoint"
	// CmdFlagPkgGeneration this flag is used to specify the generation of the package.
	CmdFlagPkgGeneration = "pkg_generation"
	// CmdFlagGseRoot this flag is used to specify the gse root.
	CmdFlagGseRoot = "gse_root"
	// CmdFlagNodeRole this flag is used to specify the node role.
	CmdFlagNodeRole = "node_role"
	// CmdFlagDebug this flag is used to specify the debug.
	CmdFlagDebug = "debug"
	// CmdFlagWorkspace this flag is used to specify the workspace dir.
	CmdFlagWorkspace = "workspace"
	// CmdFlagLogFilePath this flag is used to specify the log file path.
	CmdFlagLogFilePath = "log_file_path"
	// CmdFlagPkgName this flag is used to specify the pkg name.
	CmdFlagPkgName = "pkg_name"
	// CmdFlagPreCheckListPath this flag is used to specify the pre check list path.
	CmdFlagPreCheckListPath = "pre_check_list_path"
	// CmdFlagSetupDirPath this flag is used to specify the setup dir path.
	CmdFlagSetupDirPath = "setup_dir_path"
	// 	CmdFlagRunDirPath this flag is used to specify the run dir path.
	CmdFlagRunDirPath = "run_dir_path"
	// CmdFlagToken this flag is used to specify the token.
	CmdFlagToken = "token"
	// CmdFlagReinstall this flag is used to specify the reinstall.
	CmdFlagReinstall = "reinstall"
	// CmdFlagReRegisterAgentID this flag is used to specify the reinstall.
	CmdFlagReRegisterAgentID = "re_register_agent_id"
	// CmdFlagAgentID this flag is used to specify the agent id.
	CmdFlagAgentID = "agent_id"
)

const (
	// CmdDefaultPkgGeneration this flag is used to specify the default generation.
	CmdDefaultPkgGeneration int = 2
	// CmdDefaultWorkspace this flag is used to specify the default workspace dir.
	CmdDefaultWorkspace = "/tmp"
)

// CmdDefaultGseRoot get default gse root.
func CmdDefaultGseRoot() string {
	switch runtime.GOOS {
	// nolint: goconst
	case "windows":
		return "C:\\gse2"
	default:
		// nolint: goconst
		return "/usr/local/gse2"
	}
}

// CmdDefaultLogFilePath get default log file path.
func CmdDefaultLogFilePath() string {
	return filepath.Join(GetTmpDir(), "logs", fmt.Sprintf("installer_%s.log", time.Now().Format("2006-01-02T15-04-05")))
}
