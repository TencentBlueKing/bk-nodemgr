/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package node

import (
	"os"
	"path/filepath"
	"strings"

	nodeStep "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

// CleanOldReleasePackages removes old release .tgz packages from dataDir,
// keeping only the one at currentPkgPath.
func CleanOldReleasePackages(dataDir, currentPkgPath string) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		logger.Warnf(nodeStep.StepCleanup, "failed to read data dir for cleanup: %v", err)
		return
	}

	absCurrentPkg, err := filepath.Abs(currentPkgPath)
	if err != nil {
		logger.Warnf(nodeStep.StepCleanup, "failed to resolve current pkg path: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".tgz") {
			continue
		}
		fullPath := filepath.Join(dataDir, entry.Name())
		absFullPath, err := filepath.Abs(fullPath)
		if err != nil {
			continue
		}
		if absFullPath == absCurrentPkg {
			continue
		}
		if err := os.Remove(fullPath); err != nil {
			logger.Warnf(nodeStep.StepCleanup, "failed to remove old release package %s: %v", fullPath, err)
		} else {
			logger.Infof(nodeStep.StepCleanup, "removed old release package: %s", fullPath)
		}
	}
}
