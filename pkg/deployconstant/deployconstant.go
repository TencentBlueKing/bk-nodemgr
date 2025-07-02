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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DeployConf defines the deployment configuration for agent.
type DeployConf struct {
	Generation         types.Generation
	OsType             criteria.OSType
	HostIDPath         string
	InstallerWorkspace string
	GseDataIPC         string
	GsePluginIPC       string
	GseHomeDir         string
	GseDataDir         string
	GseRunDir          string
	GseLogDir          string
	GseEnvironDir      string
}

// Validate checks if the deployment configuration is valid.
func (conf DeployConf) Validate() error {
	if err := conf.Generation.Validate(); err != nil {
		return fmt.Errorf("invalid generation, err: %w", err)
	}

	if err := conf.OsType.Validate(); err != nil {
		return err
	}

	if conf.GseHomeDir == "" {
		return errors.New("gseHomeDir is empty")
	}

	if conf.GseDataDir == "" {
		return errors.New("gseDataDir is empty")
	}

	if conf.GseRunDir == "" {
		return errors.New("gseRunDir is empty")
	}

	if conf.GseLogDir == "" {
		return errors.New("gseLogDir is empty")
	}

	if conf.GseEnvironDir == "" {
		return errors.New("gseEnvironDir is empty")
	}

	return nil
}

// nolint: gochecknoglobals
var deployConfMap = make(map[types.Generation]map[criteria.OSType]DeployConf)

// GetDeployConf returns the deployment configuration for the specified OS type.
func GetDeployConf(generation types.Generation, osType criteria.OSType) (DeployConf, error) {
	confMap, ok := deployConfMap[generation]
	if !ok {
		return DeployConf{}, fmt.Errorf("deploy conf not found for generation: %d", generation)
	}

	conf, ok := confMap[osType]
	if !ok {
		return DeployConf{}, fmt.Errorf("deploy conf not found for os type: %s", osType)
	}

	return conf, nil
}

// SetDeployConf sets the deployment configuration for the specified OS type.
// this map only set once, if the osType already exists, it will not be set again.
func SetDeployConf(conf DeployConf) error {
	// Populate default values if not set
	populateDefaultValues(&conf)

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed, err: %w", err)
	}

	if _, ok := deployConfMap[conf.Generation]; !ok {
		deployConfMap[conf.Generation] = make(map[criteria.OSType]DeployConf)
	}

	if _, ok := deployConfMap[conf.Generation][conf.OsType]; !ok {
		deployConfMap[conf.Generation][conf.OsType] = conf
	}

	return nil
}

func populateDefaultValues(conf *DeployConf) {
	if conf.OsType == criteria.OSWindows {
		populateDefaultValuesWindows(conf)

		return
	}

	populateDefaultValuesUnix(conf)
}

func populateDefaultValuesUnix(conf *DeployConf) {
	if conf.HostIDPath == "" {
		conf.HostIDPath = fmt.Sprintf("/var/lib/%s/host/hostid", system.GetEnv())
	}
	if conf.InstallerWorkspace == "" {
		conf.InstallerWorkspace = fmt.Sprintf("/tmp/bknm/%s", system.GetEnv())
	}
	if conf.GseDataIPC == "" {
		conf.GseDataIPC = fmt.Sprintf("/var/run/%s/ipc.state.report", system.GetEnv())
	}
	if conf.GsePluginIPC == "" {
		conf.GsePluginIPC = fmt.Sprintf("/var/run/%s/ipc.state.message", system.GetEnv())
	}
	if conf.GseHomeDir == "" {
		//conf.GseHomeDir = "/usr/local/gse2"
		conf.GseHomeDir = fmt.Sprintf("/usr/local/%s", system.GetEnv())
	}
	if conf.GseDataDir == "" {
		conf.GseDataDir = fmt.Sprintf("/var/lib/%s", system.GetEnv())
	}
	if conf.GseRunDir == "" {
		conf.GseRunDir = fmt.Sprintf("/var/run/%s", system.GetEnv())
	}
	if conf.GseLogDir == "" {
		conf.GseLogDir = fmt.Sprintf("/var/log/%s", system.GetEnv())
	}
	if conf.GseEnvironDir == "" {
		conf.GseEnvironDir = fmt.Sprintf("/etc/sysconfig/%s", system.GetEnv())
	}
}

func populateDefaultValuesWindows(conf *DeployConf) {
	if conf.HostIDPath == "" {
		conf.HostIDPath = fmt.Sprintf("C:\\\\%s\\\\data\\\\host\\\\hostid", system.GetEnv())
	}
	if conf.InstallerWorkspace == "" {
		conf.InstallerWorkspace = fmt.Sprintf("C:\\\\tmp\\\\bknm\\\\%s", system.GetEnv())
	}
	if conf.GseDataIPC == "" {
		conf.GseDataIPC = "27000"
	}
	if conf.GsePluginIPC == "" {
		conf.GsePluginIPC = "26000"
	}
	if conf.GseHomeDir == "" {
		conf.GseHomeDir = fmt.Sprintf("C:\\\\%s", system.GetEnv())
	}
	if conf.GseDataDir == "" {
		conf.GseDataDir = fmt.Sprintf("C:\\\\%s\\\\data", system.GetEnv())
	}
	if conf.GseRunDir == "" {
		conf.GseRunDir = fmt.Sprintf("C:\\\\%s\\\\data", system.GetEnv())
	}
	if conf.GseLogDir == "" {
		conf.GseLogDir = fmt.Sprintf("C:\\\\%s\\\\log", system.GetEnv())
	}
	if conf.GseEnvironDir == "" {
		conf.GseEnvironDir = fmt.Sprintf("C:\\\\Windows\\\\System32\\\\config\\\\gse\\\\%s", system.GetEnv())
	}
}
