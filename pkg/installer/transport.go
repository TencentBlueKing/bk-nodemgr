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

package installer

// UnixMakeExecutableCommand makes a transferred installer executable through a Unix shell.
func UnixMakeExecutableCommand(path string) (string, error) {
	return "chmod +x " + path, nil
}

// UnixEnsureDirectoryCommand creates an installer directory through a Unix shell.
func UnixEnsureDirectoryCommand(path string) (string, error) {
	return "mkdir -p " + path, nil
}

// WindowsEnsureDirectoryCommand creates an installer directory through the WMI command channel.
// Unlike EnsureWindowsWorkDirCommand, it uses the existing mkdir command rather than PowerShell.
func WindowsEnsureDirectoryCommand(path string) (string, error) {
	return "mkdir " + path, nil
}

// WindowsBatchCommand returns a transferred installer batch path for the WMI command channel.
// SYNC: WMI RunSilentCommand passes this directly to Win32_Process.Create without a shell.
func WindowsBatchCommand(path string) (string, error) {
	return path, nil
}
