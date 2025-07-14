//go:build windows

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package windows provides the agenthandler in windows.
package windows

import (
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(rootAbsDir, deployEnv string) *AgentHandler {
	handler := &AgentHandler{
		rootAbsDir: rootAbsDir,
		deployEnv:  deployEnv,
	}
	handler.initConfigs()

	return handler
}

// AgentHandler provides the windows agenthandler.
type AgentHandler struct {
	rootAbsDir string
	deployEnv  string

	setupDir  string
	backupDir string

	binDir  string
	etcDir  string
	certDir string

	gseCtlFilePath      string
	agentBinFilePath    string
	agentDaemonFilePath string
	agentConfigFilePath string

	agentDaemonServiceName string
}

// Role return agent role.
func (handler *AgentHandler) Role() types.NodeRole {
	return types.NodeRoleAgent
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
	handler.setupDir = string(types.NodeRoleAgent)
	handler.backupDir = "backup"

	handler.binDir = filepath.Join(handler.setupDir, "bin")
	handler.etcDir = filepath.Join(handler.setupDir, "etc")
	handler.certDir = filepath.Join(handler.setupDir, "cert")

	handler.gseCtlFilePath = filepath.Join(handler.binDir, "gsectl.bat")
	handler.agentBinFilePath = filepath.Join(handler.binDir, gseAgentBinName)
	handler.agentDaemonFilePath = filepath.Join(handler.binDir, gseAgentDaemonName)
	handler.agentConfigFilePath = filepath.Join(handler.etcDir, "gse_agent.conf")

	handler.agentDaemonServiceName = gseAgentDaemonName + "_" + handler.deployEnv
}

const (
	gseAgentBinName    = "gse_agent.exe"
	gseAgentDaemonName = "gse_agent_daemon.exe"
)

// isGseBin check if the name is a gse bin.
// notice: in order to kill the process which is not belong gse bin, we need to check the process name.
func isGseBin(name string) bool {
	switch name {
	case gseAgentBinName, gseAgentDaemonName:
		return true
	default:
		return false
	}
}
