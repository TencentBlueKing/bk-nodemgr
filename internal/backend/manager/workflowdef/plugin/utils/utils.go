/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils use to provide some common utils for node actions.
package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
)

const (
	// Dangerous directory paths for Unix systems.
	dangerousDirUnixRoot = "/"
	dangerousDirUnixProc = "/proc/"
	dangerousDirUnixSys  = "/sys/"
	dangerousDirUnixDev  = "/dev/"

	// Dangerous directory paths for Windows systems.
	dangerousDirWindowsCRoot            = "c:\\"
	dangerousDirWindowsCWindows         = "c:\\windows\\"
	dangerousDirWindowsCProgramFiles    = "c:\\program files\\"
	dangerousDirWindowsCProgramFilesX86 = "c:\\program files (x86)\\"
	dangerousDirWindowsCPrograms        = "c:\\programs\\"
	dangerousDirWindowsCRecovery        = "c:\\recovery\\"
)

// CheckDirPathSafe checks if the given dir path is safe to use based on the OS type.
func CheckDirPathSafe(dirPath string, osType criteria.OSType) error {
	if osType == criteria.OSWindows {
		return checkWindowsDirPathSafe(dirPath)
	}

	return checkUnixDirPathSafe(dirPath)
}

// checkUnixDirPathSafe checks if the given unix dir path is safe to use.
func checkUnixDirPathSafe(dirPath string) error {
	if dirPath == "" {
		return errors.New("dirPath is empty")
	}

	cleanPath := filepath.Clean(dirPath)
	// compare the cleaned path with the original path
	if cleanPath != dirPath {
		return fmt.Errorf("dirPath is not a clean path, clean-path(%s), origin-path(%s)", cleanPath, dirPath)
	}

	if containsParentDirSegment(cleanPath) {
		return fmt.Errorf("dirPath contains parent directory segment(..), dirPath(%s)", dirPath)
	}

	if err := isDangerousPath(dirPath); err != nil {
		return fmt.Errorf("dirPath is dangerous, dirPath(%s): %v", dirPath, err)
	}

	return nil
}

// checkWindowsDirPathSafe checks if the given windows dir path is safe to use.
func checkWindowsDirPathSafe(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("dirPath is empty")
	}

	cleanPath := filepath.Clean(dirPath)
	// compare the cleaned path with the original path
	if cleanPath != dirPath {
		return fmt.Errorf("dirPath is not a clean path, clean-path(%s), origin-path(%s)", cleanPath, dirPath)
	}

	if containsParentDirSegment(cleanPath) {
		return fmt.Errorf("dirPath contains parent directory segment(..), dirPath(%s)", dirPath)
	}

	if err := isDangerousPath(dirPath); err != nil {
		return fmt.Errorf("dirPath is dangerous, dirPath(%s): %v", dirPath, err)
	}

	return nil
}

func isDangerousPath(path string) error {
	path = filepath.Clean(strings.ToLower(path))

	if path == dangerousDirUnixRoot {
		return errors.New("dirPath is root path, too dangerous")
	}

	if path == dangerousDirWindowsCRoot {
		return fmt.Errorf("dirPath is dangerous, dirPath(%s)", path)
	}

	// Check if the path contains any dangerous patterns
	dangerousDirPrefixs := []string{
		dangerousDirUnixProc,
		dangerousDirUnixSys,
		dangerousDirUnixDev,
		dangerousDirWindowsCWindows,
		dangerousDirWindowsCProgramFiles,
		dangerousDirWindowsCProgramFilesX86,
		dangerousDirWindowsCPrograms,
		dangerousDirWindowsCRecovery,
	}
	for _, dangerousDir := range dangerousDirPrefixs {
		if strings.HasPrefix(path, filepath.Clean(dangerousDir)) {
			return fmt.Errorf("dirPath is dangerous, dirPath(%s)", path)
		}
	}

	return nil
}

func containsParentDirSegment(cleanPath string) bool {
	parts := strings.FieldsFunc(cleanPath, func(r rune) bool {
		return r == filepath.Separator || r == winpath.DirSeparator
	})

	return slices.Contains(parts, "..")
}
