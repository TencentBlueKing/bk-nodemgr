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
	nodeType   constant.NodeType
	installEnv string
	gsePrefix  string
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
		return errors.New("set node generation failed, generation shuold be greater than 0")
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

// GetInstallEnv get install env.
func GetInstallEnv() string {
	return node.installEnv
}

// SetInstallEnv set install env.
func SetInstallEnv(installEnv string) error {
	if installEnv == "" {
		return errors.New("set install env failed, install env is empty")
	}

	node.installEnv = installEnv

	return nil
}

// GetGsePrefix get gse prefix.
func GetGsePrefix() string {
	return node.gsePrefix
}

// SetGsePrefix set gse prefix.
func SetGsePrefix(gsePrefix string) error {
	if gsePrefix == "" {
		return errors.New("set gse prefix failed, gse prefix is empty")
	}

	gsePrefix, err := filepath.Abs(gsePrefix)
	if err != nil {
		return fmt.Errorf("set gse prefix failed, err: %w", err)
	}

	node.gsePrefix = filepath.FromSlash(gsePrefix)

	return nil
}

// SetNodeType set node type.
func SetNodeType(nodeType constant.NodeType) error {
	if err := nodeType.Validate(); err != nil {
		return fmt.Errorf("set node type failed, err: %w", err)
	}

	node.nodeType = nodeType

	return nil
}

// GetNodeType get node type.
func GetNodeType() constant.NodeType {
	return node.nodeType
}

// GetNodePkgGeneration get node pkg generation.
func GetNodePkgGeneration() int {
	return node.generation
}

// GetNodePkgVersion get node pkg version.
func GetNodePkgVersion() string {
	return node.version
}
