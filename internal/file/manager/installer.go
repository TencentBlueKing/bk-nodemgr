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

package manager

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// IInstaller is interface for file installer.
type IInstaller interface {
	// GetInstaller get installer.
	GetInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (fileiface.File, error)
}

// GetInstaller get installer.
// The installer binary is a locally pre-built artifact that is present on disk at startup
// (via installerFileGroup, which maps to a local directory path). It does not pass through
// the filecache because it is already stored locally and does not require remote fetching
// or MD5-based deduplication.
func (m *Manager) GetInstaller(nCtx contextx.IContext, osType criteria.OSType, cpuArch criteria.CPUArch) (fileiface.File, error) {
	toolName, err := tool.FormatInstallerName(osType, cpuArch)
	if err != nil {
		return nil, err
	}

	return m.installerFileGroup.GetFile(nCtx, toolName)
}
