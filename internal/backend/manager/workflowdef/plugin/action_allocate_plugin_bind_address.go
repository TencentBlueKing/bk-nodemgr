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

package plugin

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameAllocatePluginBindAddress defines the action name.
	ActionNameAllocatePluginBindAddress = "allocate_plugin_bind_address"

	allocatePluginBindAddressScriptTimeout = 30 * time.Second
	allocatePluginBindAddressPollInterval  = time.Second
)

// NewActionAllocatePluginBindAddress creates an action that allocates a plugin bind address.
func NewActionAllocatePluginBindAddress(capability *Capability) action.Definition {
	return &actionAllocatePluginBindAddress{
		daoPluginDeployment: capability.StoragePlugin,
		daoPluginRelease:    capability.StorageRelease,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamAllocatePluginBindAddress defines the parameters for actionAllocatePluginBindAddress.
type ActParamAllocatePluginBindAddress struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionAllocatePluginBindAddress struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPluginRelease    release.IPlugin
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionAllocatePluginBindAddress) Name() string {
	return ActionNameAllocatePluginBindAddress
}

// Version returns the version of the action.
func (act *actionAllocatePluginBindAddress) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionAllocatePluginBindAddress) Description() string {
	return "allocate plugin bind address"
}

// Timeout returns the timeout of the action.
func (act *actionAllocatePluginBindAddress) Timeout() time.Duration {
	return time.Minute
}

// Tags returns the tags of the action.
func (act *actionAllocatePluginBindAddress) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionAllocatePluginBindAddress) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the retry delay function.
func (act *actionAllocatePluginBindAddress) DelayFn(_ int) func() {
	return func() {
		time.Sleep(time.Second)
	}
}

// Do allocates the first unused TCP port allowed by the plugin release.
func (act *actionAllocatePluginBindAddress) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamAllocatePluginBindAddress)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	process := &std.DeployInfo().Process
	pluginRelease, err := act.daoPluginRelease.GetReleasePlugin(std.Context(), types.ReleasePluginKey{
		Generation: process.Generation,
		Platform:   process.Platform,
		Version:    process.Info.Version,
		Name:       process.PluginPkgName,
	})
	if err != nil {
		return fmt.Errorf("failed to get plugin release info: %w", err)
	}
	if !pluginRelease.Enabled {
		return fmt.Errorf("plugin release is not enabled, plugin-name(%s), version(%s)", process.PluginName, process.Info.Version)
	}

	allocation := pluginRelease.BindAddressAllocated
	if !allocation.Enable {
		return nil
	}

	if net.ParseIP(allocation.BindIP) == nil {
		return fmt.Errorf("invalid plugin bind IP, bind-ip(%s)", allocation.BindIP)
	}

	if err := allocation.BindPortAvailableRange.Validate(); err != nil {
		return fmt.Errorf("invalid plugin bind port range: %w", err)
	}

	std.InstanceData().Log().
		Zh("分配插件监听端口要求目标主机已安装 netstat 命令").
		En("allocating a plugin listen port requires netstat on the target host").
		Info()

	scriptType, script := buildAllocatePluginBindAddressScript(process.Platform.OS, allocation.BindPortAvailableRange)
	endpoint := types.Endpoint{AgentID: process.Info.AgentID}
	taskID, err := act.gseHandler.ExecuteScript(
		std.Context(),
		scriptType,
		script,
		allocatePluginBindAddressScriptTimeout,
		&types.EndpointWithAuth{Endpoint: endpoint},
	)
	if err != nil {
		return fmt.Errorf("failed to execute plugin bind address allocation script: %w", err)
	}

	std.InstanceData().Log().
		Zh("开始等待插件监听地址分配结果, task-id(%s)", taskID).
		En("start waiting for plugin listen address allocation result, task-id(%s)", taskID).
		Info()

	port, err := act.waitAllocatedPluginBindPort(std, taskID, endpoint)
	if err != nil {
		return err
	}

	process.BindIP = allocation.BindIP
	process.BindPort = port

	std.InstanceData().Log().
		Zh("成功分配插件监听地址, plugin-name(%s), bind-ip(%s), bind-port(%d)",
			process.PluginName, process.BindIP, process.BindPort).
		En("successfully allocated plugin bind address, plugin-name(%s), bind-ip(%s), bind-port(%d)",
			process.PluginName, process.BindIP, process.BindPort).
		Info()

	return nil
}

func (act *actionAllocatePluginBindAddress) waitAllocatedPluginBindPort(
	std *pluginUtils.PluginActionStandarder, taskID string, endpoint types.Endpoint) (int, error) {

	ticker := time.NewTicker(allocatePluginBindAddressPollInterval)
	defer ticker.Stop()

	for {
		results, err := act.gseHandler.QueryScriptExecutionResult(
			std.Context(),
			taskID,
			&types.EndpointWithRestrict{Endpoint: endpoint},
		)
		if err != nil {
			return 0, fmt.Errorf("failed to query plugin bind address allocation result: %w", err)
		}

		if len(results) > 1 {
			return 0, fmt.Errorf("plugin bind address allocation result count is not 1, count(%d)", len(results))
		}

		if len(results) == 1 {
			result := results[0]
			switch result.Status {
			case types.ScriptStatusFinished:
				port, err := conv.StringToPort(strings.TrimSpace(result.ScreenLog))
				if err != nil {
					return 0, fmt.Errorf("invalid allocated plugin bind port: %w", err)
				}

				return int(port), nil
			case types.ScriptStatusFailed, types.ScriptStatusTimeout, types.ScriptStatusStopped, types.ScriptStatusAgentRestarted:
				return 0, fmt.Errorf(
					"plugin bind address allocation failed, status(%s), exit-code(%d), error-code(%d), error-message(%s), screen(%s)",
					result.Status, result.ExitCode, result.ErrorCode, result.ErrorMessage, result.ScreenLog,
				)
			case types.ScriptStatusUnknown, types.ScriptStatusReceived, types.ScriptStatusRunning:
				// The allocation task is still in progress.
			}
		}

		select {
		case <-std.Context().Done():
			return 0, fmt.Errorf("plugin bind address allocation interrupted: %w", std.Context().Err())
		case <-ticker.C:
		}
	}
}

func buildAllocatePluginBindAddressScript(osType criteria.OSType, portRange types.PluginPkgAvailablePortRange) (types.ScriptType, string) {
	if osType == criteria.OSWindows {
		return types.ScriptTypeBat, buildBatAllocatePluginBindAddressScript(string(portRange))
	}

	return types.ScriptTypeBash, fmt.Sprintf(unixAllocatePluginBindAddressScript, quoteShellArg(string(portRange)))
}

// buildBatAllocatePluginBindAddressScript composes a cmd.exe batch script that
// selects the first port absent from the local TCP socket table.
func buildBatAllocatePluginBindAddressScript(portRange string) string {
	commands := []string{
		"@echo off",
		"setlocal EnableExtensions EnableDelayedExpansion",
		fmt.Sprintf(`set "PORT_RANGE=%s"`, escapeBatPath(portRange)),
		strings.ReplaceAll(batProbeScript, "\n", "\r\n"),
	}

	return strings.Join(commands, "\r\n") + "\r\n"
}

// nolint: dupword
const unixAllocatePluginBindAddressScript = `#!/bin/bash
set -euo pipefail

port_range=%s
tcp_table=""

if ! command -v netstat >/dev/null 2>&1; then
    echo "netstat is required to allocate a plugin bind address" >&2
    exit 2
fi

tcp_table="$(netstat -an 2>/dev/null)" || {
    echo "failed to inspect local TCP ports with netstat" >&2
    exit 2
}

used_ports=()

mark_used_port() {
    local endpoint="$1"
    local port="${endpoint##*:}"

    if [[ "$port" == "$endpoint" ]]; then
        port="${endpoint##*.}"
    fi
    if [[ "$port" =~ ^[0-9]+$ ]]; then
        used_ports[port]=1
    fi
}

while read -r protocol _ _ local_endpoint _; do
    case "$protocol" in
        tcp*) mark_used_port "$local_endpoint" ;;
    esac
done <<< "$tcp_table"

IFS=',' read -r -a range_items <<< "$port_range"
for item in "${range_items[@]}"; do
    start_port_text="${item%%%%-*}"
    end_port_text="${item##*-}"
    start_port=$((10#$start_port_text))
    end_port=$((10#$end_port_text))

    for ((port = start_port; port <= end_port; port++)); do
        if [[ -z "${used_ports[port]+used}" ]]; then
            printf '%%d\n' "$port"
            exit 0
        fi
    done
done

echo "no available plugin port found" >&2
exit 1
`

// batProbeScript selects the first port absent from the local TCP socket table.
const batProbeScript = `if not exist "%SystemRoot%\System32\netstat.exe" (
    >&2 echo netstat.exe is required to allocate a plugin bind address
    exit /b 2
)

"%SystemRoot%\System32\netstat.exe" -an -p tcp >nul 2>&1
if errorlevel 1 (
    >&2 echo failed to inspect local TCP ports with netstat.exe
    exit /b 2
)

for /f "tokens=1,2" %%A in ('"%SystemRoot%\System32\netstat.exe" -an -p tcp 2^>nul') do (
    if /I "%%A"=="TCP" (
        set "LOCAL_ENDPOINT=%%B"
        call :mark_used_port
    )
)

for %%R in (%PORT_RANGE:,= %) do (
    set "START_PORT="
    set "END_PORT="
    for /f "tokens=1,2 delims=-" %%A in ("%%R") do (
        set "START_PORT=%%A"
        set "END_PORT=%%B"
    )
    if not defined END_PORT set "END_PORT=!START_PORT!"

    for /l %%P in (!START_PORT!,1,!END_PORT!) do (
        if not defined USED_%%P (
            echo %%P
            exit /b 0
        )
    )
)

goto :no_available_port

:mark_used_port
:strip_local_endpoint
for /f "tokens=1,* delims=:" %%A in ("%LOCAL_ENDPOINT%") do (
    if "%%B"=="" goto :mark_local_port
    set "LOCAL_ENDPOINT=%%B"
    goto :strip_local_endpoint
)

:mark_local_port
set "USED_%LOCAL_ENDPOINT%=1"
exit /b 0

:no_available_port
>&2 echo no available plugin port found
exit /b 1`

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionAllocatePluginBindAddress) DisplayNameZh() string {
	return "分配插件监听地址"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionAllocatePluginBindAddress) DisplayNameEn() string {
	return "Allocate Plugin Bind Address"
}
