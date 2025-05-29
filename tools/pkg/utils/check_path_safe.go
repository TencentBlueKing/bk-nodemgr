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
	"os"
	"path/filepath"
	"strings"
)

// CheckDirPathSafe checks if the given dir path is safe to use.
func CheckDirPathSafe(dirPath string) (bool, error) {
	if dirPath == "" {
		return false, fmt.Errorf("dirPath is empty")
	}

	cleanPath := filepath.Clean(dirPath)
	// compare the cleaned path with the original path
	if cleanPath != dirPath {
		return false, fmt.Errorf("dirPath is not a clean path, clean-path(%s), origin-path(%s)", cleanPath, dirPath)
	}

	fi, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("dirPath is not exist, dirPath(%s)", dirPath)
	}

	if !fi.IsDir() {
		return false, fmt.Errorf("dirPath is not a dir, dirPath(%s)", dirPath)
	}

	if err := isDangerousPath(dirPath); err != nil {
		return false, fmt.Errorf("dirPath is dangerous, dirPath(%s), err: %v", dirPath, err)
	}

	return true, nil
}

func isDangerousPath(path string) error {
	path = filepath.Clean(path)

	if path == "/" {
		return fmt.Errorf("dirPath is root path, too dangerous")
	}

	// Check if the path contains any dangerous patterns
	dangerousDirPrefixs := []string{
		"/proc/",
		"/sys/",
		"/dev/",
	}
	for _, dangerousDir := range dangerousDirPrefixs {
		if strings.HasPrefix(path, dangerousDir) {
			return fmt.Errorf("dirPath is dangerous, dirPath(%s)", path)
		}
	}

	return nil
}
