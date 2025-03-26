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
)

// DeployConf defines the deployment configuration for agent.
type DeployConf struct {
	HostIDPath    string `json:"host_id_path"`
	GseDataIPC    string `json:"gse_data_ipc"`
	GsePluginIPC  string `json:"gse_plugin_ipc"`
	GseHomeDir    string `json:"gse_home_dir"`
	GseDataDir    string `json:"gse_data_dir"`
	GseRunDir     string `json:"gse_run_dir"`
	GseLogDir     string `json:"gse_log_dir"`
	GseEnvironDir string `json:"gse_environ_dir"`
}

// Validate checks if the deployment configuration is valid.
func (conf DeployConf) Validate() error {
	if conf.HostIDPath == "" {
		return errors.New("host_id_path is empty")
	}

	if conf.GseDataIPC == "" {
		return errors.New("gse_data_ipc is empty")
	}

	if conf.GsePluginIPC == "" {
		return errors.New("gse_plugin_ipc is empty")
	}

	if conf.GseHomeDir == "" {
		return errors.New("gse_home_dir is empty")
	}

	if conf.GseDataDir == "" {
		return errors.New("gse_data_dir is empty")
	}

	if conf.GseRunDir == "" {
		return errors.New("gse_run_dir is empty")
	}

	if conf.GseLogDir == "" {
		return errors.New("gse_log_dir is empty")
	}

	if conf.GseEnvironDir == "" {
		return errors.New("gse_environ_dir is empty")
	}

	return nil
}

// nolint: gochecknoglobals
var deployConfMap = make(map[string]DeployConf)

// GetDeployConf returns the deployment configuration for the specified OS type.
func GetDeployConf(osType string) (DeployConf, error) {
	if conf, ok := deployConfMap[osType]; ok {
		return conf, nil
	}

	return DeployConf{}, fmt.Errorf("deploy conf not found for os type: %s", osType)
}

// SetDeployConf sets the deployment configuration for the specified OS type.
// this map only set once, if the osType already exists, it will not be set again.
func SetDeployConf(osType string, conf DeployConf) error {
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed, err: %w", err)
	}

	if _, ok := deployConfMap[osType]; !ok {
		deployConfMap[osType] = conf
	}

	return nil
}
