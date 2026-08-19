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

// Package utils use to provide some common utils for plugin actions.
package utils

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"slices"
	"strings"

	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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

const (
	// DefaultEndpointSelectionCount is the default count when selecting endpoints.
	DefaultEndpointSelectionCount = 3
)

// BuildServerURLs builds a comma-separated string of multiple addresses in the format: http://addr1,http://addr2,http://addr3.
// The installer already supports this format and will use utils.SplitServerAddrs() to split and process it.
// If endpoints is empty, returns an empty string.
func BuildServerURLs(endpoints ...discover.Endpoint) string {
	if len(endpoints) == 0 {
		return ""
	}

	addrs := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		addr := ep.GetIPV4Address()
		if addr != "" {
			addrs = append(addrs, "http://"+addr)
		}
	}

	return strings.Join(addrs, installer.ServerAddrSeparator)
}

// GeneratePluginInstallerServerEndpoints generates the server URLs for installer.
// Returns: (callbackURL, downloadURL).
func GeneratePluginInstallerServerEndpoints(nCtx contextx.IContext, provider discover.Discover,
	storageHost topoStg.IStorageHost, storageNetworkUnit topoStg.IStorageNetworkUnit,
	storageProcess pluginStg.IDaoProcess, hostID int64) ([]discover.Endpoint, []discover.Endpoint, error) {

	// get host.
	host, err := storageHost.GetHostByID(nCtx, hostID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get host: %w", err)
	}
	if host.Dynamic.NetworkUnitID < 0 {
		return nil, nil, fmt.Errorf("invalid networkunit id: %d", host.Dynamic.NetworkUnitID)
	}

	// get networkunit.
	networkUnit, err := storageNetworkUnit.GetNetworkUnit(nCtx, host.Dynamic.NetworkUnitID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get networkunit: %w", err)
	}

	// if networkunit is direct, select endpoints from direct services with discover provider.
	if networkUnit.IsDirect {
		return generateDirectServerEndpoints(provider)
	}

	callbackEndpoints, err := generateInDirectServerEndpoints(nCtx, hostID, networkUnit, storageHost, storageProcess)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate in-direct server endpoints: %w", err)
	}

	// in-direct never use download endpoints.
	return callbackEndpoints, []discover.Endpoint{}, nil
}

func generateInDirectServerEndpoints(nCtx contextx.IContext, hostID int64, networkUnit *types.NetworkUnit,
	storageHost topoStg.IStorageHost, storageProcess pluginStg.IDaoProcess) ([]discover.Endpoint, error) {

	// select relay endpoints for callback.
	proxyHosts, _, err := storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NetworkUnitID: []int64{networkUnit.ID},
			NodeRole:      []types.NodeRole{types.NodeRoleProxy},
			NodeStatus:    []types.NodeStatus{types.NodeStatusRunning},
			ProxyTags:     []types.ProxyTag{types.ProxyTagDedicatedInstaller},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list running proxy hosts: %w", err)
	}

	if len(proxyHosts) == 0 {
		return []discover.Endpoint{}, nil
	}

	// check relay service is running.
	proxyHostIDs := conv.SliceToSlice[*types.Host, int64](proxyHosts, func(host *types.Host) int64 {
		return host.HostID
	})
	relayProcesses, _, err := storageProcess.ListProcesses(nCtx, types.UnlimitedPage(), &types.ProcessCondition{
		ExactInclude: &types.ProcessExactFields{
			HostID: proxyHostIDs,
			InfoStatus: []types.ProcessStatus{
				types.ProcessStatusRunning,
			},
			PluginName: []string{
				relayhandler.PluginName,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list running relay processes: %w", err)
	}

	runningRelayProcessMap, err := conv.SliceToMap[int64, *types.Process](relayProcesses,
		func(p *types.Process) int64 { return p.HostID },
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert running relay processes: %w", err)
	}

	callbackEndpoints := make([]discover.Endpoint, 0)
	for _, host := range proxyHosts {
		if _, ok := runningRelayProcessMap[host.HostID]; !ok {
			continue
		}

		if host.HostID == hostID {
			continue
		}

		if host.Dynamic.RelayCallbackPort <= 0 {
			continue
		}

		callbackEndpoints = append(callbackEndpoints, discover.Endpoint{
			IPV4: host.Dynamic.AdvertiseIP,
			IPV6: host.Dynamic.AdvertiseIPV6,
			Port: int(host.Dynamic.RelayCallbackPort),
		})
	}
	if len(callbackEndpoints) == 0 {
		return []discover.Endpoint{}, nil
	}

	// shuffle callback endpoints.
	rand.Shuffle(len(callbackEndpoints), func(i, j int) {
		callbackEndpoints[i], callbackEndpoints[j] = callbackEndpoints[j], callbackEndpoints[i]
	})

	if len(callbackEndpoints) > DefaultEndpointSelectionCount {
		callbackEndpoints = callbackEndpoints[:DefaultEndpointSelectionCount]
	}

	return callbackEndpoints, nil
}

func generateDirectServerEndpoints(provider discover.Discover) ([]discover.Endpoint, []discover.Endpoint, error) {
	callbackEndpoints, err := provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select backend callback endpoints: %w", err)
	}

	downloadEndpoints, err := provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select file endpoints: %w", err)
	}

	return callbackEndpoints, downloadEndpoints, nil
}
