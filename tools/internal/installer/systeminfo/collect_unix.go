//go:build !windows

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

package systeminfo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	osReleasePath     = "/etc/os-release"
	kernelReleasePath = "/proc/sys/kernel/osrelease"
)

func collectSystemVersion() (string, error) {
	content, err := readOptionalFile(osReleasePath)
	if err != nil || content == "" {
		return "", err
	}

	return parseOSReleasePrettyName(content), nil
}

func collectKernelVersion() (string, error) {
	if runtime.GOOS != "linux" {
		return "", nil
	}

	content, err := readOptionalFile(kernelReleasePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(content), nil
}

func readOptionalFile(path string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}

		return "", err
	}

	return string(content), nil
}

func parseOSReleasePrettyName(content string) string {
	for _, line := range strings.Split(content, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "PRETTY_NAME" {
			continue
		}

		return unquoteOSReleaseValue(value)
	}

	return ""
}

func unquoteOSReleaseValue(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return ""
	}

	unquotedValue, err := strconv.Unquote(trimmedValue)
	if err != nil {
		return trimmedValue
	}

	return strings.TrimSpace(unquotedValue)
}
