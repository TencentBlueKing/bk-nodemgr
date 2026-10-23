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

package database

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// execCommand replaces the current process with the given command, so that
// interactive sessions and exit codes pass through transparently.
// extraEnv entries are appended to the current environment.
func execCommand(name string, args []string, extraEnv []string) error {
	binary, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("failed to find %s in PATH: %w", name, err)
	}

	// nolint: gosec // the binary is resolved from PATH and args are built from trusted config.
	return syscall.Exec(binary, append([]string{name}, args...), mergeEnv(extraEnv))
}

// mergeEnv appends extraEnv to the current environment, dropping inherited entries
// with the same keys first: on duplicate keys the child process reads the first value,
// so an inherited ETCDCTL_USER/REDISCLI_AUTH would otherwise shadow the configured
// credential.
func mergeEnv(extraEnv []string) []string {
	extraKeys := make(map[string]struct{}, len(extraEnv))
	for _, entry := range extraEnv {
		key, _, _ := strings.Cut(entry, "=")
		extraKeys[key] = struct{}{}
	}

	environ := os.Environ()
	env := make([]string, 0, len(environ)+len(extraEnv))
	for _, entry := range environ {
		key, _, _ := strings.Cut(entry, "=")
		if _, overridden := extraKeys[key]; overridden {
			continue
		}
		env = append(env, entry)
	}

	return append(env, extraEnv...)
}

// fileExists reports whether the path exists as a regular file.
// TLS arguments are injected only for existing files: a configured but missing
// cert path means the current environment has no TLS material mounted.
func fileExists(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)

	return err == nil && info.Mode().IsRegular()
}

// splitHostPort splits "host:port" into host and port, using defaultPort when the
// address has no port component. Bracketed IPv6 addresses with a port
// ("[::1]:6379") are supported; a bracketed address without a port ("[::1]")
// is returned as-is, including the brackets.
func splitHostPort(addr, defaultPort string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, defaultPort
	}

	return host, port
}
