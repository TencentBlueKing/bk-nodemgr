/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deployconstant

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	nodeConfigDirName        = "etc"
	nodeLibDirName           = "lib"
	nodeUnixDataIPCName      = "ipc.state.report"
	nodeUnixPluginIPCName    = "ipc.state.message"
	nodeWindowsDataIPCPort   = "27000"
	nodeWindowsPluginIPCPort = "26000"
)

// NodeDeployConf defines the deployment configuration for agent.
type NodeDeployConf struct {
	DeployConf

	// custom.
	LogDir     string
	HostIDPath string
}

// Validate checks if the deployment configuration is valid.
func (conf NodeDeployConf) Validate() error {
	if err := conf.DeployConf.Validate(); err != nil {
		return fmt.Errorf("invalid deploy conf: %w", err)
	}

	return nil
}

// nolint: gochecknoglobals
var nodeDeployConfMap = make(map[types.Generation]map[criteria.OSType]NodeDeployConf)

// GetNodeDeployConf returns the deployment configuration for the specified OS type.
func GetNodeDeployConf(generation types.Generation, osType criteria.OSType) (NodeDeployConf, error) {
	confMap, ok := nodeDeployConfMap[generation]
	if !ok {
		return NodeDeployConf{}, fmt.Errorf("deploy conf not found for generation, generation(%d)", generation)
	}

	conf, ok := confMap[osType]
	if !ok {
		return NodeDeployConf{}, fmt.Errorf("deploy conf not found for os type, os-type(%s)", osType)
	}

	return conf, nil
}

// SetNodeDeployConf sets the deployment configuration for the specified OS type.
// this map only set once, if the osType already exists, it will not be set again.
func SetNodeDeployConf(conf NodeDeployConf) error {
	// Populate default values if not set
	populateNodeDefaultValues(&conf)

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed: %w", err)
	}

	if _, ok := nodeDeployConfMap[conf.Generation]; !ok {
		nodeDeployConfMap[conf.Generation] = make(map[criteria.OSType]NodeDeployConf)
	}

	if _, ok := nodeDeployConfMap[conf.Generation][conf.OsType]; !ok {
		nodeDeployConfMap[conf.Generation][conf.OsType] = conf
	}

	return nil
}

func populateNodeDefaultValues(conf *NodeDeployConf) {
	env := system.GetEnv()

	conf.DeployDir = tool.JoinPath(conf.OsType, conf.BaseDeployDir, env)
	conf.WorkDir = tool.JoinPath(conf.OsType, conf.BaseWorkDir, env)

	// HostIDPath is a special logic of CMDB that cannot be modified
	if conf.OsType == criteria.OSWindows {
		if conf.LogDir == "" {
			conf.LogDir = fmt.Sprintf("C:\\%s\\logs\\", env)
		}

		if conf.HostIDPath == "" {
			conf.HostIDPath = fmt.Sprintf("C:\\%s\\data\\host\\hostid", env)
		}

		return
	}

	if conf.LogDir == "" {
		conf.LogDir = fmt.Sprintf("/var/log/%s/", env)
	}

	if conf.HostIDPath == "" {
		conf.HostIDPath = fmt.Sprintf("/var/lib/%s/host/hostid", env)
	}
}

// GenerateNodeHomeDir generates the home directory path for the given node role.
func (conf NodeDeployConf) GenerateNodeHomeDir(role types.NodeRole) string {
	return tool.JoinPath(conf.OsType, conf.DeployDir, string(role))
}

// GenerateDefaultNodeConfigDir generates the default configuration directory path for the given node role.
func (conf NodeDeployConf) GenerateDefaultNodeConfigDir(role types.NodeRole) string {
	return tool.JoinPath(conf.OsType, conf.DeployDir, string(role), nodeConfigDirName)
}

// GenerateDefaultDataIPCPath generates the default data IPC path based on the OS type.
func (conf NodeDeployConf) GenerateDefaultDataIPCPath(role types.NodeRole) string {
	if conf.OsType == criteria.OSWindows {
		return nodeWindowsDataIPCPort
	}

	return filepath.Join(conf.DeployDir, string(role), nodeLibDirName, nodeUnixDataIPCName)
}

// GenerateDefaultPluginIPCPath generates the default plugin IPC path based on the OS type.
func (conf NodeDeployConf) GenerateDefaultPluginIPCPath(role types.NodeRole) string {
	if conf.OsType == criteria.OSWindows {
		return nodeWindowsPluginIPCPort
	}

	return filepath.Join(conf.DeployDir, string(role), nodeLibDirName, nodeUnixPluginIPCName)
}
