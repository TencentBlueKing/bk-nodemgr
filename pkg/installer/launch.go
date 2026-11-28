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

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
)

// RenderUnixLaunchScript renders an installer script and wraps its workspace, write, and launch commands.
func RenderUnixLaunchScript(workDir string, render func() (string, string, error)) (string, error) {
	command, err := RenderUnixScript(render)
	if err != nil {
		return "", err
	}

	return UnixWorkspaceCommand(workDir, command)
}

// RenderUnixScript renders the script-write and launch commands without preparing a workspace.
func RenderUnixScript(render func() (string, string, error)) (string, error) {
	name, content, err := render()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`echo "%s" > %s && sh %s`, content, name, name), nil
}

// UnixWorkspaceCommand creates and enters a workspace before executing command.
// The command must be a trusted command already rendered by this package, not raw input.
func UnixWorkspaceCommand(workDir, command string) (string, error) {
	return fmt.Sprintf("mkdir -p %s && cd %s && %s", workDir, workDir, command), nil
}

// RenderWindowsLaunchScript renders an installer batch script and wraps its workspace, write, and launch commands.
func RenderWindowsLaunchScript(workDir string, render func() (string, string, error)) (string, error) {
	name, content, err := render()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`mkdir "%s" 2>nul & cd "%s" && echo %s > "%s" && cmd /c "%s"`,
		workDir, workDir, content, name, name), nil
}

// RenderWindowsCommand renders an installer command executed in an existing workspace.
// Unlike RenderWindowsLaunchScript, it neither creates a workspace nor writes a script.
func RenderWindowsCommand(workDir string, render func() (string, string, error)) (string, error) {
	_, content, err := render()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`cd %s && %s`, workDir, content), nil
}

// WindowsSSHLaunchCommand starts a transferred installer batch script asynchronously through PowerShell.
func WindowsSSHLaunchCommand(workDir, scriptPath string) (string, error) {
	argumentList := fmt.Sprintf("/c \"%s\"", strings.ReplaceAll(scriptPath, "\"", "\"\""))

	return encodedPowerShellCommand(fmt.Sprintf(
		"Start-Process -FilePath %s -ArgumentList %s -WorkingDirectory %s",
		launchPowerShellSingleQuote("cmd.exe"),
		launchPowerShellSingleQuote(argumentList),
		launchPowerShellSingleQuote(workDir),
	)), nil
}

// WindowsShellLaunchCommand starts a transferred installer script through Cygwin sh.
func WindowsShellLaunchCommand(workDir, scriptName string) (string, error) {
	return fmt.Sprintf("cd %s && sh %s",
		launchShellDoubleQuote(winpath.ToSlash(winpath.Clean(workDir))),
		launchShellDoubleQuote(scriptName)), nil
}

// EnsureWindowsWorkDirCommand creates the installer workspace through an encoded PowerShell command.
func EnsureWindowsWorkDirCommand(workDir string) (string, error) {
	return encodedPowerShellCommand(fmt.Sprintf(
		"New-Item -ItemType Directory -Force -Path %s | Out-Null",
		launchPowerShellSingleQuote(workDir),
	)), nil
}

// EnsureWindowsShellWorkDirCommand creates a Windows installer workspace through Cygwin sh.
func EnsureWindowsShellWorkDirCommand(workDir string) (string, error) {
	return fmt.Sprintf("mkdir -p %s", launchShellDoubleQuote(workDir)), nil
}

func encodedPowerShellCommand(script string) string {
	const (
		bytesPerCodeUnit = 2
		bitsPerByte      = 8
	)
	codeUnits := utf16.Encode([]rune(script))
	encoded := make([]byte, len(codeUnits)*bytesPerCodeUnit)
	for idx, unit := range codeUnits {
		encoded[idx*bytesPerCodeUnit] = byte(unit)
		encoded[idx*bytesPerCodeUnit+1] = byte(unit >> bitsPerByte)
	}

	return "powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand " +
		base64.StdEncoding.EncodeToString(encoded)
}

func launchPowerShellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func launchShellDoubleQuote(value string) string {
	return "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
}
