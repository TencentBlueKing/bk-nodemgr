/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package opluginpkg provides functions for formatting official plugin package names.
package opluginpkg

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// PkgExtension defines the package file extension.
const PkgExtension = "tgz"

// FormatPkgName formats the package name based on the node role, generation, and version.
func FormatPkgName(pluginName string, releaseType types.ReleaseType, gen types.Generation, plat platform.Platform, version string) (string, error) {

	if err := releaseType.Validate(); err != nil {
		return "", fmt.Errorf("format official plugin pkg name failed, err: %w", err)
	}

	if version == "" {
		return "", errors.New("format official plugin pkg name failed, version is empty")
	}

	// origin agent pkg contains all platforms in one pkg.
	// so the pkg-name should not contain platform info.
	switch releaseType {
	case types.ReleaseTypeOriginOfficialPlugin:
		pkgName := fmt.Sprintf(
			"bk-nodemgr_%s_%d_%s-%s-all.%s",
			releaseType,
			gen,
			pluginName,
			version,
			PkgExtension)

		return pkgName, nil
	case types.ReleaseTypeOfficialPlugin:
		if !plat.Validate() {
			return "", fmt.Errorf("format official plugin pkg name failed, platform is invalid: %s", plat.String())
		}

		pkgName := fmt.Sprintf(
			"bk-nodemgr_%s_%d_%s-%s-%s_%s.%s",
			releaseType,
			gen,
			pluginName,
			version,
			plat.OS,
			plat.Arch,
			PkgExtension)

		return pkgName, nil
	default:
		return "", fmt.Errorf("format official plugin pkg name failed, release type is invalid: %s", releaseType)
	}
}
