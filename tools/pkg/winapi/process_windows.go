//go:build windows

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package winapi

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// KillProcessByNameAndPath terminates a process by its name and executable path.
func KillProcessByNameAndPath(processName, execPath string) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("create snapshot failed: %w", err)
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	err = windows.Process32First(snapshot, &pe)
	if err != nil {
		return fmt.Errorf("Process32First failed: %w", err)
	}

	for {
		exeName := windows.UTF16ToString(pe.ExeFile[:])
		if strings.EqualFold(exeName, processName) {
			process, err := windows.OpenProcess(
				windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_TERMINATE,
				false,
				pe.ProcessID)
			if err != nil {
				continue
			}
			defer windows.CloseHandle(process)

			var pathBuf [windows.MAX_PATH]uint16
			size := uint32(len(pathBuf))
			err = windows.QueryFullProcessImageName(process, 0, &pathBuf[0], &size)
			if err != nil {
				continue
			}

			processPath := windows.UTF16ToString(pathBuf[:size])

			if strings.EqualFold(processPath, execPath) {
				err = windows.TerminateProcess(process, 1)
				if err != nil {
					return fmt.Errorf("terminate process failed: %w", err)
				}
			}
		}

		err = windows.Process32Next(snapshot, &pe)
		if err != nil {
			break
		}
	}

	return nil
}
