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

package deployconstant

import (
	"fmt"

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

	nodeZoneID = "default"
	nodeCityID = "default"

	nodeDataAgentBindIP   = "::"
	nodeDataAgentBindPort = 28625

	nodeFileAgentBindIP   = "::"
	nodeFileAgentBindPort = 28925

	nodeProxyBindIP   = "::"
	nodeProxyBindPort = 28668

	nodeFileBittorrentBindIP             = "::"
	nodeFileBittorrentBindPort           = 10020
	nodeFileBittorrentTrackerBindPort    = 10030
	nodeFileBittorrentSpeedLimitMBPerSec = 10000

	nodeFileTopologyBindIP   = "::"
	nodeFileTopologyBindPort = 28930

	nodeFileAgentBindPortV1        = 58925
	nodeFileTopologyThriftBindPort = 58930
	nodeFileMetricExporterBindPort = 29404
	nodeDataMetricExporterBindPort = 29402
)

// NodeDeployConf defines the deployment configuration for agent.
type NodeDeployConf struct {
	DeployConf

	// custom.
	LogDir            string
	ExtraConfigDir    string
	DataIPC           string
	PluginIPC         string
	ProxyFileCacheDir string
	ZoneID            string
	CityID            string

	// Global Linux proxy network defaults, independent of per-network-unit deployment settings.
	DataAgentBindIP                  string
	DataAgentBindPort                int64
	FileAgentBindIP                  string
	FileAgentBindPort                int64
	ProxyBindIP                      string
	ProxyBindPort                    int64
	FileBittorrentBindIP             string
	FileBittorrentBindPort           int64
	FileBittorrentTrackerBindPort    int64
	FileBittorrentSpeedLimitMBPerSec int64
	FileTopologyBindIP               string
	FileTopologyBindPort             int64
	FileAgentBindPortV1              int64
	FileTopologyThriftBindPort       int64
	FileMetricExporterBindPort       int64
	DataMetricExporterBindPort       int64
}

// Validate checks if the deployment configuration is valid.
func (conf NodeDeployConf) Validate() error {
	if err := conf.DeployConf.Validate(); err != nil {
		return fmt.Errorf("invalid deploy conf: %w", err)
	}

	if conf.OsType != criteria.OSLinux {
		return nil
	}

	return conf.validateProxyBindPortsUnique()
}

// validateProxyBindPortsUnique rejects proxy bind ports that collide with each other.
// Proxy defaults are only populated on Linux, so callers must not use it for other OS types.
func (conf NodeDeployConf) validateProxyBindPortsUnique() error {
	ports := []struct {
		name  string
		value int64
	}{
		{"dataAgentBindPort", conf.DataAgentBindPort},
		{"fileAgentBindPort", conf.FileAgentBindPort},
		{"proxyBindPort", conf.ProxyBindPort},
		{"fileBittorrentBindPort", conf.FileBittorrentBindPort},
		{"fileBittorrentTrackerBindPort", conf.FileBittorrentTrackerBindPort},
		{"fileTopologyBindPort", conf.FileTopologyBindPort},
		{"fileAgentBindPortV1", conf.FileAgentBindPortV1},
		{"fileTopologyThriftBindPort", conf.FileTopologyThriftBindPort},
		{"fileMetricExporterBindPort", conf.FileMetricExporterBindPort},
		{"dataMetricExporterBindPort", conf.DataMetricExporterBindPort},
	}

	seen := make(map[int64]string, len(ports))
	for _, port := range ports {
		if prev, ok := seen[port.value]; ok {
			return fmt.Errorf("proxy bind port conflict: %s and %s both use port %d", prev, port.name, port.value)
		}

		seen[port.value] = port.name
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
	conf.ZoneID = conv.NonEmptyOr(conf.ZoneID, nodeZoneID)
	conf.CityID = conv.NonEmptyOr(conf.CityID, nodeCityID)

	if conf.OsType == criteria.OSLinux {
		populateProxyDefaultValues(conf)
	}
}

func populateProxyDefaultValues(conf *NodeDeployConf) {
	// Empty IPs and zero numeric values preserve the built-in proxy network defaults.
	conf.DataAgentBindIP = conv.NonEmptyOr(conf.DataAgentBindIP, nodeDataAgentBindIP)
	conf.DataAgentBindPort = conv.NonEmptyOr(conf.DataAgentBindPort, nodeDataAgentBindPort)
	conf.FileAgentBindIP = conv.NonEmptyOr(conf.FileAgentBindIP, nodeFileAgentBindIP)
	conf.FileAgentBindPort = conv.NonEmptyOr(conf.FileAgentBindPort, nodeFileAgentBindPort)
	conf.ProxyBindIP = conv.NonEmptyOr(conf.ProxyBindIP, nodeProxyBindIP)
	conf.ProxyBindPort = conv.NonEmptyOr(conf.ProxyBindPort, nodeProxyBindPort)
	conf.FileBittorrentBindIP = conv.NonEmptyOr(conf.FileBittorrentBindIP, nodeFileBittorrentBindIP)
	conf.FileBittorrentBindPort = conv.NonEmptyOr(conf.FileBittorrentBindPort, nodeFileBittorrentBindPort)
	conf.FileBittorrentTrackerBindPort = conv.NonEmptyOr(conf.FileBittorrentTrackerBindPort, nodeFileBittorrentTrackerBindPort)
	conf.FileBittorrentSpeedLimitMBPerSec = conv.NonEmptyOr(conf.FileBittorrentSpeedLimitMBPerSec, nodeFileBittorrentSpeedLimitMBPerSec)
	conf.FileTopologyBindIP = conv.NonEmptyOr(conf.FileTopologyBindIP, nodeFileTopologyBindIP)
	conf.FileTopologyBindPort = conv.NonEmptyOr(conf.FileTopologyBindPort, nodeFileTopologyBindPort)
	conf.FileAgentBindPortV1 = conv.NonEmptyOr(conf.FileAgentBindPortV1, nodeFileAgentBindPortV1)
	conf.FileTopologyThriftBindPort = conv.NonEmptyOr(conf.FileTopologyThriftBindPort, nodeFileTopologyThriftBindPort)
	conf.FileMetricExporterBindPort = conv.NonEmptyOr(conf.FileMetricExporterBindPort, nodeFileMetricExporterBindPort)
	conf.DataMetricExporterBindPort = conv.NonEmptyOr(conf.DataMetricExporterBindPort, nodeDataMetricExporterBindPort)
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

// GenerateDataDir generates the data dir path based on the OS type.
// only for unix system, for windows system, it will be empty.
func (conf NodeDeployConf) GenerateDataDir(role types.NodeRole) string {
	return tool.JoinPath(conf.OsType, conf.GenerateDeployDir(), string(role), nodeLibDirName)
}

// GenerateDataIPC generates the data IPC based on the OS type.
func (conf NodeDeployConf) GenerateDataIPC(role types.NodeRole) string {
	if conf.DataIPC != "" {
		return conf.DataIPC
	}

	if conf.OsType == criteria.OSWindows {
		return nodeWindowsDataIPCPort
	}

	return tool.JoinPath(conf.OsType, conf.GenerateDataDir(role), nodeUnixDataIPCName)
}

// GeneratePluginIPC generates the plugin IPC based on the OS type.
func (conf NodeDeployConf) GeneratePluginIPC(role types.NodeRole) string {
	if conf.PluginIPC != "" {
		return conf.PluginIPC
	}

	if conf.OsType == criteria.OSWindows {
		return nodeWindowsPluginIPCPort
	}

	return tool.JoinPath(conf.OsType, conf.GenerateDataDir(role), nodeUnixPluginIPCName)
}

// GetDefaultDataIPC returns the default data IPC based on the OS type.
func (conf NodeDeployConf) GetDefaultDataIPC() string {
	if conf.OsType == criteria.OSWindows {
		return conf.getWindowsDefaultDataIPCPort()
	}

	return conf.getUnixDefaultDataIPCPath()
}

// GetDefaultPluginIPC returns the default plugin IPC based on the OS type.
func (conf NodeDeployConf) GetDefaultPluginIPC() string {
	if conf.OsType == criteria.OSWindows {
		return conf.getWindowsDefaultPluginIPCPort()
	}

	return conf.getUnixDefaultPluginIPCPath()
}

// getUnixDefaultDataIPCPath returns the default data IPC path for Unix.
func (conf NodeDeployConf) getUnixDefaultDataIPCPath() string {
	return tool.JoinPath(conf.OsType, conf.GenerateDeployDir(), "{node_role}", nodeLibDirName, nodeUnixDataIPCName)
}

// getUnixDefaultPluginIPCPath returns the default plugin IPC path for Unix.
func (conf NodeDeployConf) getUnixDefaultPluginIPCPath() string {
	return tool.JoinPath(conf.OsType, conf.GenerateDeployDir(), "{node_role}", nodeLibDirName, nodeUnixPluginIPCName)
}

// getWindowsDefaultDataIPCPort returns the default data IPC port for Windows.
func (conf NodeDeployConf) getWindowsDefaultDataIPCPort() string {
	return nodeWindowsDataIPCPort
}

// getWindowsDefaultPluginIPCPort returns the default plugin IPC port for Windows.
func (conf NodeDeployConf) getWindowsDefaultPluginIPCPort() string {
	return nodeWindowsPluginIPCPort
}

// NodeEventDataIDConf defines event data-id configuration for node rendering.
type NodeEventDataIDConf struct {
	AgentBaseAlarmEventDataID int64
	TaskProcEventDataID       int64
}

// Validate checks if the event data-id configuration is valid.
func (conf NodeEventDataIDConf) Validate() error {
	if conf.AgentBaseAlarmEventDataID <= 0 {
		return fmt.Errorf("agentAlarmEventDataID must be positive, got %d", conf.AgentBaseAlarmEventDataID)
	}

	if conf.TaskProcEventDataID <= 0 {
		return fmt.Errorf("processEventDataID must be positive, got %d", conf.TaskProcEventDataID)
	}

	return nil
}
