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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DeployConf defines the deployment configuration for node.
type DeployConf struct {
	Generation    types.Generation
	OsType        criteria.OSType
	BaseDeployDir string
	BaseWorkDir   string

	ManualScriptPath string
}

// Validate checks if the deployment configuration is valid.
func (conf DeployConf) Validate() error {
	if err := conf.Generation.Validate(); err != nil {
		return fmt.Errorf("invalid generation: %w", err)
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
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("set deploy conf failed: %w", err)
	}

	if _, ok := deployConfMap[conf.Generation]; !ok {
		deployConfMap[conf.Generation] = make(map[criteria.OSType]DeployConf)
	}

	if _, ok := deployConfMap[conf.Generation][conf.OsType]; !ok {
		deployConfMap[conf.Generation][conf.OsType] = conf
	}

	return nil
}

// GenerateDeployDir generates the deploy dir for the deployment configuration.
func (conf DeployConf) GenerateDeployDir() string {
	return tool.JoinPath(conf.OsType, conf.BaseDeployDir, system.GetEnv())
}

// GenerateWorkDir generates the work dir for the deployment configuration.
func (conf DeployConf) GenerateWorkDir() string {
	return tool.JoinPath(conf.OsType, conf.BaseWorkDir, system.GetEnv())
}
