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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// NodeDeployConf defines the deployment configuration for agent.
type NodeDeployConf struct {
	DeployConf

	// base.
	BaseWorkDir string

	// custom.
	LogDir             string
	HostIDPath         string
	AgentDataIPCPath   string
	AgentPluginIPCPath string
	EnvironDir         string

	// generates by base and custom.
	WorkDir string
}

// Validate checks if the deployment configuration is valid.
func (conf NodeDeployConf) Validate() error {
	if err := conf.Generation.Validate(); err != nil {
		return fmt.Errorf("invalid generation, err: %w", err)
	}

	if err := conf.OsType.Validate(); err != nil {
		return err
	}

	if conf.BaseWorkDir == "" {
		return errors.New("baseWorkDir is empty")
	}

	if conf.BaseDeployDir == "" {
		return errors.New("baseDeployDir is empty")
	}

	if conf.LogDir == "" {
		return errors.New("logDir is empty")
	}

	if conf.HostIDPath == "" {
		return errors.New("hostIDPath is empty")
	}

	if conf.AgentDataIPCPath == "" {
		return errors.New("agentDataIPCPath is empty")
	}

	if conf.AgentPluginIPCPath == "" {
		return errors.New("agentPluginIPCPath is empty")
	}

	if conf.EnvironDir == "" {
		return errors.New("environDir is empty")
	}

	return nil
}

// nolint: gochecknoglobals
var nodeDeployConfMap = make(map[types.Generation]map[criteria.OSType]NodeDeployConf)

// GetNodeDeployConf returns the deployment configuration for the specified OS type.
func GetNodeDeployConf(generation types.Generation, osType criteria.OSType) (NodeDeployConf, error) {
	confMap, ok := nodeDeployConfMap[generation]
	if !ok {
		return NodeDeployConf{}, fmt.Errorf("deploy conf not found for generation: %d", generation)
	}

	conf, ok := confMap[osType]
	if !ok {
		return NodeDeployConf{}, fmt.Errorf("deploy conf not found for os type: %s", osType)
	}

	return conf, nil
}

// SetNodeDeployConf sets the deployment configuration for the specified OS type.
// this map only set once, if the osType already exists, it will not be set again.
func SetNodeDeployConf(conf NodeDeployConf) error {
	// Populate default values if not set
	populateNodeDefaultValues(&conf)

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed, err: %w", err)
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
	if conf.OsType == criteria.OSWindows {
		populateNodeDefaultValuesWindows(conf)

		return
	}

	populateNodeDefaultValuesUnix(conf)
}

func populateNodeDefaultValuesUnix(conf *NodeDeployConf) {
	env := system.GetEnv()

	if conf.LogDir == "" {
		conf.LogDir = fmt.Sprintf("/var/log/%s/", env)
	}

	if conf.HostIDPath == "" {
		conf.HostIDPath = fmt.Sprintf("/var/lib/%s/host/hostid", env)
	}

	if conf.AgentDataIPCPath == "" {
		conf.AgentDataIPCPath = fmt.Sprintf("/var/run/%s/ipc.state.report", env)
	}

	if conf.AgentPluginIPCPath == "" {
		conf.AgentPluginIPCPath = fmt.Sprintf("/var/run/%s/ipc.state.message", env)
	}

	if conf.EnvironDir == "" {
		conf.EnvironDir = fmt.Sprintf("/etc/sysconfig/%s/", env)
	}

	conf.WorkDir = filepath.Join(conf.BaseWorkDir, env)
	conf.DeployDir = filepath.Join(conf.BaseDeployDir, env)
}

func populateNodeDefaultValuesWindows(conf *NodeDeployConf) {
	env := system.GetEnv()

	if conf.LogDir == "" {
		conf.LogDir = fmt.Sprintf("C:\\%s\\logs\\", env)
	}

	if conf.HostIDPath == "" {
		conf.HostIDPath = fmt.Sprintf("C:\\%s\\data\\host\\hostid", env)
	}

	if conf.AgentDataIPCPath == "" {
		conf.AgentDataIPCPath = "27000"
	}
	if conf.AgentPluginIPCPath == "" {
		conf.AgentPluginIPCPath = "26000"
	}

	if conf.EnvironDir == "" {
		conf.EnvironDir = fmt.Sprintf("C:\\Windows\\System32\\config\\gse\\%s\\", env)
	}

	conf.WorkDir = winpath.Join(conf.BaseWorkDir, env)
	conf.DeployDir = winpath.Join(conf.BaseDeployDir, env)
}
