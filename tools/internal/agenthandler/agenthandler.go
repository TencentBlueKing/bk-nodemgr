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

// Package agenthandler provides the agent handler.
package agenthandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

// IAgentHandler agent handler interface.
type IAgentHandler interface {
	// Role return agent role.
	Role() types.NodeRole

	// FS return agent file system handler.
	FS() IAgentFSHandler

	// Process return agent process handler.
	Process() IAgentProcessHandler
}

// PreserveOptions defines generated runtime files that should be renewed instead of preserved.
type PreserveOptions struct {
	// RenewGSEProc renews the generated .proc file instead of preserving it.
	RenewGSEProc bool

	// RenewGSETask renews the generated .task file instead of preserving it.
	RenewGSETask bool
}

// IAgentFSHandler agent file system handler interface.
type IAgentFSHandler interface {
	// CheckIntegrity check the integrity of agent file system.
	CheckIntegrity(ctx context.Context) error

	// Init init agent file system.
	Init() error

	// Backup backup agent file system.
	Backup(ctx context.Context) error

	// Purge purge agent file system.
	Purge(ctx context.Context) error

	// CopyConfigDir copy config files to agent file system.
	CopyConfigDir(ctx context.Context, configFileAbsDir string) error

	// UnpackReleasePackage unpack release package to agent file system.
	// if keepOldFileAsTmp is true, will move old-existing file to a tmp file in same directory,
	// to keep the process running. Call Clean() to clean all the tmp files.
	UnpackReleasePackage(ctx context.Context, releasePkgAbsPath string, keepOldFileAsTmp bool) error

	// SaveGSERuntimeFile saves generated GSE runtime files before reinstall or upgrade overwrites them.
	SaveGSERuntimeFile(ctx context.Context, opts PreserveOptions) error

	// RestoreGSERuntimeFile restores generated GSE runtime files saved by SaveGSERuntimeFile.
	RestoreGSERuntimeFile(ctx context.Context, opts PreserveOptions) error

	// Clean cleans all the tmp files and tools in agent file system.
	Clean(ctx context.Context) error
}

// AgentVersionDiagnostic contains raw gse_agent version diagnostic command metadata and outputs.
type AgentVersionDiagnostic struct {
	WorkDir    string
	Executable string
	Args       []string
	Stdout     string
	Stderr     string
}

// LogString formats diagnostic metadata and raw outputs as one physical log line.
func (diagnostic *AgentVersionDiagnostic) LogString() string {
	return fmt.Sprintf(
		"work_dir(%s) executable(%s) args(%v) stdout_raw(%s) stderr_raw(%s)",
		diagnostic.WorkDir,
		diagnostic.Executable,
		diagnostic.Args,
		escapeLogField(diagnostic.Stdout),
		escapeLogField(diagnostic.Stderr),
	)
}

func escapeLogField(value string) string {
	value = strings.TrimSpace(value)
	return strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(value)
}

// IAgentProcessHandler agent process handler interface.
type IAgentProcessHandler interface {
	// GetProcess get current node process status.
	// including file and data when it is proxy.
	GetProcess(ctx context.Context) (*NodeProcess, error)

	// DiagnoseVersion runs the agent binary version diagnostic command.
	DiagnoseVersion(ctx context.Context) (*AgentVersionDiagnostic, error)

	// ForceKill force kill the agent process.
	// including file and data when it is proxy.
	ForceKill(ctx context.Context) error

	// GetAgentID get the current running agent-id.
	GetAgentID(ctx context.Context) (string, error)

	// RegisterAgentID register agent-id.
	RegisterAgentID(ctx context.Context, existingAgentID string) (string, error)

	// UnregisterAgentID unregister agent-id.
	UnregisterAgentID(ctx context.Context) error

	// InstallAutoStartup install auto startup after host boot.
	InstallAutoStartup(ctx context.Context, opts *AsyncOutputOptions) error

	// UninstallAutoStartup uninstall auto startup after host boot.
	UninstallAutoStartup(ctx context.Context, opts *AsyncOutputOptions) error

	// Start start the agent process.
	Start(ctx context.Context, opts *AsyncOutputOptions) error

	// Stop stop the agent process.
	Stop(ctx context.Context, force bool, opts *AsyncOutputOptions) error

	// Restart restart the agent process.
	Restart(ctx context.Context, force bool, opts *AsyncOutputOptions) error

	// Reload reload the agent process.
	Reload(ctx context.Context, opts *AsyncOutputOptions) error
}

// NewNodeProcess creates a new node process.
func NewNodeProcess() *NodeProcess {
	return &NodeProcess{
		Running: make([]string, 0),
		Dead:    make([]string, 0),
	}
}

// NodeProcess defines the node process status.
type NodeProcess struct {
	Running []string
	Dead    []string
}

// IsAllRunning check if all node process is running.
func (np NodeProcess) IsAllRunning() bool {
	return len(np.Running) > 0 && len(np.Dead) == 0
}

// IsAllDead check if all node process is dead.
func (np NodeProcess) IsAllDead() bool {
	return len(np.Running) == 0 && len(np.Dead) > 0
}

// AsyncOutputFunc async output function.
type AsyncOutputFunc func(content string)

// AsyncOutputOptions async output options.
type AsyncOutputOptions struct {
	Stdout AsyncOutputFunc
	Stderr AsyncOutputFunc
}
