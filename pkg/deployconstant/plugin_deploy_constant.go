/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package deployconstant provides constants and configuration for deployment.
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
	pluginBaseDirName   = "plugin"
	pluginConfigDirName = "etc"
	pluginRunDirName    = "run"
	pluginDataDirName   = "data"

	pluginUnixLogDirFormat        = "/var/log/%s/plugin/"
	pluginUnixHostIDPathFormat    = "/var/lib/%s/host/hostid"
	pluginWindowsLogDirFormat     = "C:\\%s\\logs\\plugin\\"
	pluginWindowsHostIDPathFormat = "C:\\%s\\data\\host\\hostid"
)

// PluginDeployConf defines the deployment configuration for agent.
type PluginDeployConf struct {
	DeployConf

	// custom.
	LogDir          string
	HostIDPath      string
	CommonConstants map[string]map[string]any
}

// Validate checks if the deployment configuration is valid.
func (conf PluginDeployConf) Validate() error {
	if err := conf.DeployConf.Validate(); err != nil {
		return fmt.Errorf("invalid deploy conf: %w", err)
	}

	return nil
}

// nolint: gochecknoglobals
var pluginDeployConfMap = make(map[types.Generation]map[criteria.OSType]PluginDeployConf)

// GetPluginDeployConf returns the deployment configuration for the specified OS type.
func GetPluginDeployConf(generation types.Generation, osType criteria.OSType) (PluginDeployConf, error) {
	confMap, ok := pluginDeployConfMap[generation]
	if !ok {
		return PluginDeployConf{}, fmt.Errorf("deploy conf not found for generation: %d", generation)
	}

	conf, ok := confMap[osType]
	if !ok {
		return PluginDeployConf{}, fmt.Errorf("deploy conf not found for os type: %s", osType)
	}

	return conf, nil
}

// SetPluginDeployConf sets the deployment configuration for the specified OS type.
// this map only set once, if the osType already exists, it will not be set again.
func SetPluginDeployConf(conf PluginDeployConf) error {
	// Populate default values if not set
	populatePluginDefaultValues(&conf)

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed: %w", err)
	}

	if _, ok := pluginDeployConfMap[conf.Generation]; !ok {
		pluginDeployConfMap[conf.Generation] = make(map[criteria.OSType]PluginDeployConf)
	}

	if _, ok := pluginDeployConfMap[conf.Generation][conf.OsType]; !ok {
		pluginDeployConfMap[conf.Generation][conf.OsType] = conf
	}

	return nil
}

func populatePluginDefaultValues(conf *PluginDeployConf) {
	conf.LogDir = conv.NonEmptyOr(conf.LogDir, conf.generateDefaultPluginLogDir())
	conf.HostIDPath = conv.NonEmptyOr(conf.HostIDPath, conf.generateDefaultHostIDPath())
}

// generateDefaultPluginLogDir generates the default log directory for the deployment configuration.
func (conf PluginDeployConf) generateDefaultPluginLogDir() string {
	if conf.OsType == criteria.OSWindows {
		return fmt.Sprintf(pluginWindowsLogDirFormat, system.GetEnv())
	}

	return fmt.Sprintf(pluginUnixLogDirFormat, system.GetEnv())
}

// GenerateDefaultHostIDPath generates the default host ID path for the deployment configuration.
// HostIDPath is a special logic of CMDB that cannot be modified, so we need to provide a default value for it.
func (conf PluginDeployConf) generateDefaultHostIDPath() string {
	if conf.OsType == criteria.OSWindows {
		return fmt.Sprintf(pluginWindowsHostIDPathFormat, system.GetEnv())
	}

	return fmt.Sprintf(pluginUnixHostIDPathFormat, system.GetEnv())
}

// GeneratePluginDeployDir generates the deploy dir for the deployment configuration.
func (conf PluginDeployConf) GeneratePluginDeployDir() string {
	return tool.JoinPath(conf.OsType, conf.BaseDeployDir, system.GetEnv(), pluginBaseDirName)
}

// GeneratePluginHomeDir generates the plugin home directory based on plugin group and plugin name.
func (conf PluginDeployConf) GeneratePluginHomeDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginDeployDir(), pluginGroup, pluginName)
}

// GeneratePluginRunDir generates the run directory based on plugin group and plugin name.
func (conf PluginDeployConf) GeneratePluginRunDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginHomeDir(pluginGroup, pluginName), pluginRunDirName)
}

// GeneratePluginDataDir generates the data directory based on plugin group and plugin name.
func (conf PluginDeployConf) GeneratePluginDataDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginHomeDir(pluginGroup, pluginName), pluginDataDirName)
}

// GeneratePluginConfigDir generates the configuration directory based on plugin group and plugin name.
func (conf PluginDeployConf) GeneratePluginConfigDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginHomeDir(pluginGroup, pluginName), pluginConfigDirName)
}

// GeneratePluginSubConfigDir generates the sub-configuration directory based on plugin group and plugin name.
func (conf PluginDeployConf) GeneratePluginSubConfigDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginConfigDir(pluginGroup, pluginName), pluginName)
}

// CommonConstants structure:
//
//	{
//	    "plugin_name_1": {
//	        "constant_key_1": "constant_value_1",
//	        "constant_key_2": "constant_value_2"
//	    },
//	    "plugin_name_2": {
//	        "constant_key_3": "constant_value_3",
//	        "constant_key_4": "constant_value_4"
//	    },
//	    "global": {
//	        "constant_key_5": "constant_value_5"
//	    }
//	}.

// GetPluginCommonConstants returns the common constants by target plugin name.
func (conf PluginDeployConf) GetPluginCommonConstants(targetPluginName string) map[string]any {
	if conf.CommonConstants == nil {
		return map[string]any{}
	}

	pluginConstants, ok := conf.CommonConstants[targetPluginName]
	if !ok {
		return map[string]any{}
	}

	return pluginConstants
}

// GetGlobalCommonConstants returns the global common constants.
func (conf PluginDeployConf) GetGlobalCommonConstants() map[string]any {
	if conf.CommonConstants == nil {
		return map[string]any{}
	}

	globalConstants, ok := conf.CommonConstants["global"]
	if !ok {
		return map[string]any{}
	}

	return globalConstants
}
