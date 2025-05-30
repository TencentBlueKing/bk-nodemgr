//go:build linux || darwin || freebsd || aix

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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
)

// ProcessInfo represents information about a process.
type ProcessInfo struct {
	PID        int
	Name       string
	FullPath   string
	Command    string
	Namespaces map[string]string
}

// GetSameSpaceProcesses returns a list of processes running in the same namespace as the current process.
func GetSameSpaceProcesses() ([]ProcessInfo, error) {
	selfProcessInfo, err := GetProcessInfo(os.Getpid())
	if err != nil {
		return nil, fmt.Errorf("failed to get self process info, err:%v", err)
	}

	processInfos, err := GetAllProcessInfos()
	if err != nil {
		return nil, fmt.Errorf("failed to get all processes, err:%v", err)
	}

	sameSpaceProcesses := make([]ProcessInfo, 0, len(processInfos))
	for _, process := range processInfos {
		// ignore self process
		if selfProcessInfo.PID == process.PID {
			continue
		}

		if err := areInSameNamespaces(selfProcessInfo, process); err != nil {
			continue
		}

		sameSpaceProcesses = append(sameSpaceProcesses, process)
	}

	return sameSpaceProcesses, nil
}

func areInSameNamespaces(self, other ProcessInfo) error {
	// If the number of namespaces is different, they are not in the same environment
	if len(self.Namespaces) != len(other.Namespaces) {
		return errors.New("different number of namespaces")
	}

	for nsType, selfNs := range self.Namespaces {
		otherNs, exists := other.Namespaces[nsType]
		if !exists {
			return fmt.Errorf("namespace %s does not exist for process %d", nsType, other.PID)
		}

		// Skip namespaces that read the wrong thing
		if strings.HasPrefix(selfNs, "err") || strings.HasPrefix(otherNs, "err") {
			continue
		}

		if selfNs != otherNs {
			return fmt.Errorf("namespace %s is not the same for process %d and %d", nsType, self.PID, other.PID)
		}
	}

	return nil
}

// GetAllProcessInfos returns a list of all processes running on the system.
// notice: this func shouldn't return a process map, because same name process is normal in linux system.
// notice: this func result's process may be not in the same namespace as the current process.
func GetAllProcessInfos() ([]ProcessInfo, error) {
	procDirs, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc directory: %v", err)
	}

	processes := make([]ProcessInfo, 0, len(procDirs))
	for _, dir := range procDirs {
		if !dir.IsDir() {
			continue
		}

		// filter all pid directories.
		pid, err := strconv.Atoi(dir.Name())
		if err != nil {
			continue
		}

		procInfo, err := GetProcessInfo(pid)
		if err != nil {
			// process may have exited between the ReadDir and Stat calls
			continue
		}

		processes = append(processes, procInfo)
	}

	return processes, nil
}

// GetProcessInfo returns a ProcessInfo struct containing information about the process.
func GetProcessInfo(pid int) (ProcessInfo, error) {
	var procInfo ProcessInfo

	gp := gopool.NewPool()

	gp.Go(func() error {
		procInfo.PID = pid

		return nil
	})

	gp.Go(func() error {
		binNamePath := fmt.Sprintf("/proc/%d/comm", pid)
		binNameBytes, err := os.ReadFile(filepath.Clean(binNamePath))
		if err != nil {
			return fmt.Errorf("failed to read %s, err: %v", binNamePath, err)
		}

		procInfo.Name = strings.TrimSpace(string(binNameBytes))

		return nil
	})

	gp.Go(func() error {
		exePath := fmt.Sprintf("/proc/%d/exe", pid)
		fullPath, err := os.Readlink(exePath)
		if err != nil {
			// should not use cmdline to get process name, because it may be empty.
			// must make the full path available,
			// otherwise if you use this path to kill the process in the future, it will be deleted by mistake.

			return fmt.Errorf("failed to readlink %s, err: %v", exePath, err)
		}
		procInfo.FullPath = fullPath

		return nil
	})

	gp.Go(func() error {
		cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
		cmdlineData, err := os.ReadFile(filepath.Clean(cmdlinePath))
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", cmdlinePath, err)
		}

		// Convert \0-delimited parameters to space-separated strings
		if len(cmdlineData) > 0 {
			cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")
			procInfo.Command = strings.TrimSpace(cmdline)
		}

		return nil
	})

	gp.Go(func() error {
		nsDir := fmt.Sprintf("/proc/%d/ns", pid)
		entries, err := os.ReadDir(nsDir)
		if err != nil {
			return fmt.Errorf("failed to read namespace directory %s: %v", nsDir, err)
		}

		namespaces := make(map[string]string)
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				nsPath := filepath.Join(nsDir, entry.Name())
				nsLink, err := os.Readlink(nsPath)
				if err != nil {
					namespaces[entry.Name()] = fmt.Sprintf("error: %v", err)
					continue
				}
				namespaces[entry.Name()] = nsLink
			}
		}
		procInfo.Namespaces = namespaces

		return nil
	})

	if err := gp.Wait(); err != nil {
		return procInfo, fmt.Errorf("failed to get process info, err: %v", err)
	}

	return procInfo, nil
}
