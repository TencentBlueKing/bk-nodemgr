/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodepkg provides functions to format package names for different node roles and generations.
package nodepkg

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PkgExtension defines the package file extension.
const PkgExtension = "tgz"

// FormatPkgName formats the package name based on the node role, generation, and version.
func FormatPkgName(
	nodeRole types.NodeRole,
	generation types.NodeGeneration,
	version string,
	osType string,
	cpuArch string,
) (string, error) {

	if err := nodeRole.Validate(); err != nil {
		return "", fmt.Errorf("format pkg name failed, err: %w", err)
	}

	if err := generation.Validate(); err != nil {
		return "", fmt.Errorf("format pkg name failed, err: %w", err)
	}

	if version == "" {
		return "", errors.New("format pkg name failed, version is empty")
	}

	if osType == "" {
		return "", errors.New("format pkg name failed, os_type is empty")
	}

	if cpuArch == "" {
		return "", errors.New("format pkg name failed, cpu_arch is empty")
	}

	pkgName := fmt.Sprintf(
		"gse_%s-%d-%s-%s_%s.%s",
		nodeRole,
		generation,
		version,
		osType,
		cpuArch,
		PkgExtension)

	return pkgName, nil
}
