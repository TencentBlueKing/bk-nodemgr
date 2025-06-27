/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package constant ...
package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
)

// nolint: gochecknoglobals
var node = struct {
	sync.Once
	nodeRole  constant.NodeRole
	gseRoot   string
	deployEnv struct {
		env string
		sync.Once
	}
	generation int
	version    string
	agentID    string
}{}

// GetNodeAgentID get node agent id.
func GetNodeAgentID() string {
	return node.agentID
}

// SetNodeAgentID set node agent id.
func SetNodeAgentID(agentID string) error {
	if agentID == "" {
		return errors.New("set node agent id failed, agent id is empty")
	}

	node.agentID = agentID

	return nil
}

// SetNodeGeneration set node generation.
func SetNodeGeneration(generation int) error {
	if generation <= 0 {
		return errors.New("set node generation failed, generation should be greater than 0")
	}

	node.generation = generation

	return nil
}

// SetNodeVersion set node version.
func SetNodeVersion(version string) error {
	if version == "" {
		return errors.New("set node version failed, version is empty")
	}

	node.version = version

	return nil
}

// GetDeployEnv get deploy env.
func GetDeployEnv() string {
	node.deployEnv.Once.Do(func() {
		node.deployEnv.env = filepath.Base(filepath.Clean(node.gseRoot))
		if node.deployEnv.env == "" {
			node.deployEnv.env = "unknown"
			return
		}
	})

	return node.deployEnv.env
}

// GetGseRoot get gse root.
func GetGseRoot() string {
	return node.gseRoot
}

// SetGseRoot set gse root.
func SetGseRoot(gseRoot string) error {
	if gseRoot == "" {
		return errors.New("set gse root failed, gse root is empty")
	}

	gseRoot, err := filepath.Abs(gseRoot)
	if err != nil {
		return fmt.Errorf("set gse root failed, err: %w", err)
	}

	node.gseRoot = filepath.Clean(gseRoot)

	return nil
}

// SetNodeRole set node role.
func SetNodeRole(nodeRole constant.NodeRole) error {
	if err := nodeRole.Validate(); err != nil {
		return fmt.Errorf("set node role failed, err: %w", err)
	}

	node.nodeRole = nodeRole

	return nil
}

// GetNodeRole get node role.
func GetNodeRole() constant.NodeRole {
	return node.nodeRole
}

// GetNodePkgGeneration get node pkg generation.
func GetNodePkgGeneration() int {
	return node.generation
}

// GetNodePkgVersion get node pkg version.
func GetNodePkgVersion() string {
	return node.version
}
