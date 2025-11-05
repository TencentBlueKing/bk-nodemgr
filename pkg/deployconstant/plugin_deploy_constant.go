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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	pluginBaseDirName   = "plugin"
	pluginConfigDirName = "etc"
	pluginRunDirName    = "run"
	pluginDataDirName   = "data"
)

// PluginDeployConf defines the deployment configuration for agent.
type PluginDeployConf struct {
	DeployConf

	// custom.
	LogDir          string
	HostIDPath      string
	CommonConstants map[string]any
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
	env := system.GetEnv()

	conf.DeployDir = tool.JoinPath(conf.OsType, conf.BaseDeployDir, env, pluginBaseDirName)
	conf.WorkDir = tool.JoinPath(conf.OsType, conf.BaseWorkDir, env, pluginBaseDirName)

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

// GenerateDefaultRunDir generates the default run directory based on plugin group and plugin name.
func (conf PluginDeployConf) GenerateDefaultRunDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.DeployDir, pluginGroup, pluginName, pluginRunDirName)
}

// GenerateDefaultDataDir generates the default data directory based on plugin group and plugin name.
func (conf PluginDeployConf) GenerateDefaultDataDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.DeployDir, pluginGroup, pluginName, pluginDataDirName)
}

// GenerateDefaultSetupPath generates the default setup path based on plugin group and plugin name.
func (conf PluginDeployConf) GenerateDefaultSetupPath(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.DeployDir, pluginGroup, pluginName)
}

// GenerateDefaultSubConfigDir generates the default sub-configuration directory based on plugin group and plugin name.
func (conf PluginDeployConf) GenerateDefaultSubConfigDir(pluginGroup, pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.DeployDir, pluginGroup, pluginName, pluginConfigDirName, pluginName)
}
