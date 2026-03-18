/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tool this package provide stander format for the nodemgr tools.
package tool

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
)

const (
	// NamePrefixInstaller this is the prefix for the tools tool name.
	NamePrefixInstaller = "installer"
)

// FormatInstallerName formats the tools name based on the OS type and CPU architecture.
func FormatInstallerName(osType criteria.OSType, cpuArch criteria.CPUArch) (string, error) {
	if err := osType.Validate(); err != nil {
		return "", fmt.Errorf("invalid os type for installer name: %w", err)
	}

	if err := cpuArch.Validate(); err != nil {
		return "", fmt.Errorf("invalid cpu arch for installer name: %w", err)
	}

	toolName := fmt.Sprintf("%s_%s_%s", NamePrefixInstaller, osType, cpuArch)

	if osType == criteria.OSWindows {
		toolName += ".exe"
	}

	return toolName, nil
}

// JoinPath joins the path parts based on the OS type.
func JoinPath(osType criteria.OSType, parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	if osType == criteria.OSWindows {
		return winpath.Join(parts...)
	}

	return filepath.Join(parts...)
}
