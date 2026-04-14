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

package utils

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// ProcessInfo represents information about a process.
type ProcessInfo struct {
	PID        int
	Name       string
	FullPath   string
	Command    string
	Namespaces map[string]string
}

// GetProcessInfo returns a ProcessInfo struct containing information about the process.
func GetProcessInfo(pid int) (ProcessInfo, error) {
	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		uint32(pid),
	)
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("OpenProcess(%d) failed: %w", pid, err)
	}

	defer func() {
		if err := windows.CloseHandle(handle); err != nil {
			fmt.Printf("failed to close handle: %v\n", err)
		}
	}()

	var pathBuf [windows.MAX_PATH]uint16
	size := uint32(len(pathBuf))
	if err := windows.QueryFullProcessImageName(handle, 0, &pathBuf[0], &size); err != nil {
		return ProcessInfo{}, fmt.Errorf("QueryFullProcessImageName(%d) failed: %w", pid, err)
	}

	fullPath := windows.UTF16ToString(pathBuf[:size])

	return ProcessInfo{
		PID:      pid,
		FullPath: fullPath,
	}, nil
}
