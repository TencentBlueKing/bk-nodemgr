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

package process

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// pluginNameMonitorbeat is the v2 plugin name of monitor collection.
	pluginNameMonitorbeat = "bkmonitorbeat"
)

// ListV2Compatible defines the handler to list the plugin processes of hosts for node manager v2 compatibility.
func (h *handler) ListV2Compatible(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ProcessV2CompatibleListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list v2 plugin processes, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForProcessView(rCtx, req.GetBkBizId())
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list v2 plugin processes, permission denied.")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	resp := new(protoBackend.ProcessV2CompatibleListResp)
	if !scopeIsAny && len(narrowedBizIDs) == 0 {
		resp.ConvertHostPluginProcessFromTypes(0, nil)

		return resp.GetData(), nil
	}

	condition := req.ConvertConditionsToTypes()
	condition.StaticExactInclude.BizID = narrowedBizIDs
	hosts, total, err := h.daoHost.ListHost(rCtx, types.UnlimitedPage(), condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list v2 plugin processes, failed to list hosts.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	processes, err := h.listV2Processes(rCtx, hosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list v2 plugin processes, failed to list plugin processes.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	setupPaths, err := h.queryAgentSetupPaths(rCtx, hosts)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list v2 plugin processes, failed to query agent setup paths.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	items := buildHostPluginProcesses(hosts, processes, setupPaths)

	resp.ConvertHostPluginProcessFromTypes(total, items)

	return resp.GetData(), nil
}

// listV2Processes lists the recorded v2 processes of the given hosts.
func (h *handler) listV2Processes(nCtx contextx.IContext, hosts []*types.Host) (map[int64]types.ProcessInfo, error) {
	hostIDs := make([]int64, 0, len(hosts))
	for _, host := range hosts {
		hostIDs = append(hostIDs, host.HostID)
	}

	if len(hostIDs) == 0 {
		return map[int64]types.ProcessInfo{}, nil
	}

	processes, _, err := h.daoProcessV2.ListProcessesV2(nCtx, types.UnlimitedPage(), hostIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list v2 plugin processes: %w", err)
	}

	processInfos := make(map[int64]types.ProcessInfo, len(processes))
	for _, process := range processes {
		if process.PluginName != pluginNameMonitorbeat {
			continue
		}

		processInfos[process.HostID] = process.Info
	}

	return processInfos, nil
}

// setupPathKey identifies the hosts that share one agent install dir.
type setupPathKey struct {
	Generation    types.Generation
	OSType        criteria.OSType
	NetworkUnitID int64
}

// queryAgentSetupPaths resolves the agent install dir once for each distinct host group.
func (h *handler) queryAgentSetupPaths(rCtx restserver.IContext, hosts []*types.Host) (map[setupPathKey]string, error) {
	setupPaths := make(map[setupPathKey]string, len(hosts))
	for _, host := range hosts {
		key := getHostSetupPathKey(host)
		if _, ok := setupPaths[key]; ok {
			continue
		}

		nodeDeployConf, err := deployconstant.GetNodeDeployConf(key.Generation, key.OSType)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).
				With("generation", key.Generation).
				With("os-type", key.OSType).
				Warn("failed to get node deploy conf, the setup path is empty.")

			return nil, err
		}

		if key.NetworkUnitID >= 0 {
			customDeployConfig, err := h.daoDomainGse.GetNetworkUnitCustomDeployConfig(rCtx, key.NetworkUnitID, key.OSType)
			if err != nil {
				return nil, fmt.Errorf("failed to get networkunit custom deploy config: %w", err)
			}

			if customDeployConfig != nil && customDeployConfig.NodeRuntime.BaseDeployDir != "" {
				nodeDeployConf.BaseDeployDir = customDeployConfig.NodeRuntime.BaseDeployDir
			}
		}

		setupPaths[key] = nodeDeployConf.GenerateDeployDir()
	}

	return setupPaths, nil
}

// buildHostPluginProcesses builds the plugin process status of each host.
func buildHostPluginProcesses(
	hosts []*types.Host, processes map[int64]types.ProcessInfo, setupPaths map[setupPathKey]string) []*types.HostPluginProcess {

	results := make([]*types.HostPluginProcess, 0, len(hosts))
	for _, host := range hosts {
		item := &types.HostPluginProcess{
			HostID: host.HostID,
			Plugins: []types.HostPluginProcessStatus{{
				Name:   pluginNameMonitorbeat,
				Status: types.ProcessStatusUnRegister,
			}},
		}

		if host.Static != nil {
			item.InnerIP = strings.Join(host.Static.InnerIPList, types.IPSeparator)
			item.InnerIPv6 = strings.Join(host.Static.InnerIPV6List, types.IPSeparator)
			item.NetworkAreaID = host.Static.NetworkAreaID
		}

		item.SetupPath = setupPaths[getHostSetupPathKey(host)]

		if process, ok := processes[host.HostID]; ok {
			item.Plugins[0].Version = process.Version
			item.Plugins[0].Status = process.Status
		}

		results = append(results, item)
	}

	return results
}

// getHostSetupPathKey builds the normalized lookup key of the agent install dir for one host.
// The user-set dynamic fields win, the CMDB-synced static fields fill the missing os type,
// and the generation falls back to generation 2 because that is the only deployable one.
func getHostSetupPathKey(host *types.Host) setupPathKey {
	// -1 means the host belongs to no network unit, so no custom deploy config applies.
	key := setupPathKey{NetworkUnitID: -1}

	if host.Dynamic != nil {
		key.Generation = host.Dynamic.NodeGeneration
		key.OSType = host.Dynamic.NodeOsType
		key.NetworkUnitID = host.Dynamic.NetworkUnitID
	}

	if host.Static != nil && (key.OSType == "" || key.OSType == criteria.OSUnknown) {
		// the synced os type is already normalized to the go standard value.
		key.OSType = criteria.OSType(host.Static.OSType)
	}

	if key.Generation == 0 {
		key.Generation = types.Generation2
	}
	if key.OSType == "" || key.OSType == criteria.OSUnknown {
		key.OSType = criteria.OSLinux
	}

	return key
}
