/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package installer

import (
	"fmt"
	"path"
)

// workDirParentsMode is the permission of the base and env dirs above per-user workdirs:
// world-writable with the sticky bit, so every user can create its own workspace while
// only a dir owner can remove it.
const workDirParentsMode = "1777"

// EnsureUserWorkDirCmd builds the command that creates the per-user workdir as the login
// user and opens the parent base/env dirs it owns, so other users can create sibling
// workspaces later. Parent dirs owned by someone else (e.g. root from a previous install)
// are left untouched: creating the workdir inside them fails fast with a clear error.
func EnsureUserWorkDirCmd(workDir string) string {
	baseDir := path.Dir(path.Dir(workDir))
	envDir := path.Dir(workDir)

	return fmt.Sprintf(
		"mkdir -p '%s' && { [ ! -O '%s' ] || chmod %s '%s'; } && { [ ! -O '%s' ] || chmod %s '%s'; }",
		workDir, baseDir, workDirParentsMode, baseDir, envDir, workDirParentsMode, envDir)
}

// SudoEnsureUserWorkDirCmd builds the sudo-mode variant: it creates the workdir as root,
// hands it to the login user (a previous root-run install may have left it root-owned),
// and opens the parent base/env dirs via sudo, fixing legacy root-owned parents as well.
func SudoEnsureUserWorkDirCmd(workDir, loginUser string) string {
	baseDir := path.Dir(path.Dir(workDir))
	envDir := path.Dir(workDir)

	return fmt.Sprintf(
		"sudo -n mkdir -p '%s' && sudo -n chown '%s' '%s' && sudo -n chmod %s '%s' '%s'",
		workDir, loginUser, workDir, workDirParentsMode, baseDir, envDir)
}
