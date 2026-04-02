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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	nodeLibDirName           = "lib"
	nodeUnixDataIPCName      = "ipc.state.report"
	nodeUnixPluginIPCName    = "ipc.state.message"
	nodeWindowsDataIPCPort   = "27000"
	nodeWindowsPluginIPCPort = "26000"

	nodeUnixLogDirFormat            = "/var/log/%s/"
	nodeUnixExtraConfigDirFormat    = "/etc/sysconfig/gse/%s/user_conf"
	nodeWindowsLogDirFormat         = "C:\\%s\\logs\\"
	nodeWindowsExtraConfigDirFormat = "C:\\Windows\\System32\\config\\gse\\%s\\user_conf"
)

// NodeDeployConf defines the deployment configuration for agent.
type NodeDeployConf struct {
	DeployConf

	// custom.
	LogDir         string
	ExtraConfigDir string
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
	conf.LogDir = conv.NonEmptyOr(conf.LogDir, conf.generateDefaultNodeLogDir())
	conf.ExtraConfigDir = conv.NonEmptyOr(conf.ExtraConfigDir, conf.generateDefaultExtraConfigDir())
}

// generateDefaultNodeLogDir generates the default log directory for the deployment configuration.
func (conf NodeDeployConf) generateDefaultNodeLogDir() string {
	if conf.OsType == criteria.OSWindows {
		return fmt.Sprintf(nodeWindowsLogDirFormat, system.GetEnv())
	}

	return fmt.Sprintf(nodeUnixLogDirFormat, system.GetEnv())
}

// generateDefaultExtraConfigDir generates the default extra config directory for the deployment configuration.
func (conf NodeDeployConf) generateDefaultExtraConfigDir() string {
	if conf.OsType == criteria.OSWindows {
		return fmt.Sprintf(nodeWindowsExtraConfigDirFormat, system.GetEnv())
	}

	return fmt.Sprintf(nodeUnixExtraConfigDirFormat, system.GetEnv())
}

// GenerateNodeHomeDir generates the home directory path for the given node role.
func (conf NodeDeployConf) GenerateNodeHomeDir(role types.NodeRole) string {
	return tool.JoinPath(conf.OsType, conf.GenerateDeployDir(), string(role))
}

// GenerateDataIPCPath generates the data IPC path based on the OS type.
func (conf NodeDeployConf) GenerateDataIPCPath(role types.NodeRole) string {
	if conf.OsType == criteria.OSWindows {
		return nodeWindowsDataIPCPort
	}

	return filepath.Join(conf.GenerateDeployDir(), string(role), nodeLibDirName, nodeUnixDataIPCName)
}

// GeneratePluginIPCPath generates the plugin IPC path based on the OS type.
func (conf NodeDeployConf) GeneratePluginIPCPath(role types.NodeRole) string {
	if conf.OsType == criteria.OSWindows {
		return nodeWindowsPluginIPCPort
	}

	return filepath.Join(conf.GenerateDeployDir(), string(role), nodeLibDirName, nodeUnixPluginIPCName)
}
