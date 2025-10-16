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
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PluginDeployConf defines the deployment configuration for agent.
type PluginDeployConf struct {
	DeployConf

	// custom.
	LogDir             string
	RunDir             string
	DataDir            string
	HostIDPath         string
	AgentDataIPCPath   string
	AgentPluginIPCPath string
	SubConfigBaseDir   string
	CommonConstants    map[string]any
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
	nodeDeployConf, err := GetNodeDeployConf(conf.Generation, conf.OsType)
	if err != nil {
		return fmt.Errorf("failed to get node deploy conf: %w", err)
	}

	// Populate default values if not set
	populatePluginDefaultValues(&conf, nodeDeployConf)

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

func populatePluginDefaultValues(conf *PluginDeployConf, nodeConf NodeDeployConf) {
	if conf.OsType == criteria.OSWindows {
		populatePluginDefaultValuesWindows(conf, nodeConf)

		return
	}

	populatePluginDefaultValuesUnix(conf, nodeConf)
}

func populatePluginDefaultValuesUnix(conf *PluginDeployConf, nodeConf NodeDeployConf) {
	env := system.GetEnv()

	if conf.LogDir == "" {
		conf.LogDir = fmt.Sprintf("/var/log/%s/", env)
	}

	if conf.DataDir == "" {
		conf.DataDir = fmt.Sprintf("/var/lib/%s/", env)
	}

	if conf.RunDir == "" {
		conf.RunDir = fmt.Sprintf("/var/run/%s/", env)
	}

	if conf.HostIDPath == "" {
		if nodeConf.HostIDPath != "" {
			conf.HostIDPath = nodeConf.HostIDPath
		} else {
			conf.HostIDPath = fmt.Sprintf("/var/lib/%s/host/hostid", env)
		}
	}

	if conf.AgentDataIPCPath == "" {
		if nodeConf.AgentDataIPCPath != "" {
			conf.AgentDataIPCPath = nodeConf.AgentDataIPCPath
		} else {
			conf.AgentDataIPCPath = fmt.Sprintf("/var/run/%s/ipc.state.report", env)
		}
	}

	if conf.AgentPluginIPCPath == "" {
		if nodeConf.AgentPluginIPCPath != "" {
			conf.AgentPluginIPCPath = nodeConf.AgentPluginIPCPath
		} else {
			conf.AgentPluginIPCPath = fmt.Sprintf("/var/run/%s/ipc.state.message", env)
		}
	}

	conf.DeployDir = filepath.Join(conf.BaseDeployDir, env)
	conf.WorkDir = filepath.Join(conf.BaseWorkDir, env)
	conf.SubConfigBaseDir = filepath.Join(conf.DeployDir, "plugins", "etc")
}

func populatePluginDefaultValuesWindows(conf *PluginDeployConf, nodeConf NodeDeployConf) {
	env := system.GetEnv()

	if conf.LogDir == "" {
		conf.LogDir = fmt.Sprintf("C:\\%s\\logs\\", env)
	}

	if conf.DataDir == "" {
		conf.DataDir = fmt.Sprintf("C:\\%s\\data\\", env)
	}

	if conf.RunDir == "" {
		conf.RunDir = fmt.Sprintf("C:\\%s\\run\\", env)
	}

	if conf.HostIDPath == "" {
		if nodeConf.HostIDPath != "" {
			conf.HostIDPath = nodeConf.HostIDPath
		} else {
			conf.HostIDPath = fmt.Sprintf("C:\\%s\\data\\host\\hostid", env)
		}
	}

	if conf.AgentDataIPCPath == "" {
		if nodeConf.AgentDataIPCPath != "" {
			conf.AgentDataIPCPath = nodeConf.AgentDataIPCPath
		} else {
			conf.AgentDataIPCPath = "27000"
		}
	}
	if conf.AgentPluginIPCPath == "" {
		if nodeConf.AgentPluginIPCPath != "" {
			conf.AgentPluginIPCPath = nodeConf.AgentPluginIPCPath
		} else {
			conf.AgentPluginIPCPath = "26000"
		}
	}

	conf.DeployDir = winpath.Join(conf.BaseDeployDir, env)
	conf.WorkDir = winpath.Join(conf.BaseWorkDir, env)
	conf.SubConfigBaseDir = winpath.Join(conf.DeployDir, "plugins", "etc")
}
