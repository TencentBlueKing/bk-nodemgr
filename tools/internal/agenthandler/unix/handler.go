//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package unix provides the agenthandler in unix.
package unix

import (
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(rootAbsDir string) *AgentHandler {
	handler := &AgentHandler{
		rootAbsDir: rootAbsDir,
		role:       types.NodeRoleAgent,
	}
	handler.initConfigs()

	return handler
}

// NewProxyAgentHandler creates a new AgentHandler for proxy agent.
func NewProxyAgentHandler(rootAbsDir string) *AgentHandler {
	handler := &AgentHandler{
		rootAbsDir: rootAbsDir,
		role:       types.NodeRoleProxy,
	}
	handler.initConfigs()

	return handler
}

// AgentHandler provides the unix agenthandler.
type AgentHandler struct {
	rootAbsDir string
	role       types.NodeRole

	setupDir  string
	backupDir string

	binDir  string
	etcDir  string
	certDir string

	gseCtlFilePath      string
	agentBinFilePath    string
	agentConfigFilePath string

	fileBinFilePath    string
	fileConfigFilePath string
	dataBinFilePath    string
	dataConfigFilePath string

	gseRuntimeFileProcFilePath string
	gseRuntimeFileTaskFilePath string
}

// Role return agent role.
func (handler *AgentHandler) Role() types.NodeRole {
	return handler.role
}

// FS return agent file system handler.
func (handler *AgentHandler) FS() agenthandler.IAgentFSHandler {
	return handler
}

// Process return agent process handler.
func (handler *AgentHandler) Process() agenthandler.IAgentProcessHandler {
	return handler
}

// initConfigs initializes the configurations via root-abs-dir.
func (handler *AgentHandler) initConfigs() {
	handler.setupDir = string(handler.role)
	handler.backupDir = "backup"

	handler.binDir = filepath.Join(handler.setupDir, "bin")
	handler.etcDir = filepath.Join(handler.setupDir, "etc")
	handler.certDir = filepath.Join(handler.setupDir, "cert")

	handler.gseCtlFilePath = filepath.Join(handler.binDir, "gsectl")
	handler.agentBinFilePath = filepath.Join(handler.binDir, gseAgentBinName)
	handler.agentConfigFilePath = filepath.Join(handler.etcDir, "gse_agent.conf")

	// proxy role will have extra modules: file and data
	if handler.role == types.NodeRoleProxy {
		handler.fileBinFilePath = filepath.Join(handler.binDir, gseFileBinName)
		handler.fileConfigFilePath = filepath.Join(handler.etcDir, "gse_file_proxy.conf")

		handler.dataBinFilePath = filepath.Join(handler.binDir, gseDataBinName)
		handler.dataConfigFilePath = filepath.Join(handler.etcDir, "gse_data_proxy.conf")
	}

	handler.gseRuntimeFileProcFilePath = filepath.Join(handler.etcDir, ".proc")
	handler.gseRuntimeFileTaskFilePath = filepath.Join(handler.etcDir, ".task")
}

const (
	gseAgentBinName = "gse_agent"
	gseFileBinName  = "gse_file"
	gseDataBinName  = "gse_data"
)

// isGseBin check if the name is a gse bin.
// notice: in order to kill the process which is not belong gse bin, we need to check the process name.
func isGseBin(name string) bool {
	switch name {
	case gseAgentBinName, gseFileBinName, gseDataBinName:
		return true
	default:
		return false
	}
}
