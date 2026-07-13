/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"fmt"
	"slices"
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncCorrectAgentID defines the action name.
	ActionNameSyncCorrectAgentID = "sync_correct_agent_id"
)

const (
	syncCorrectAgentIDSkipHostNotFound        = "host_not_found"
	syncCorrectAgentIDSkipSyncedAgentIDEmpty  = "synced_agent_id_empty"
	syncCorrectAgentIDSkipDynamicAgentIDEmpty = "dynamic_agent_id_empty"
	syncCorrectAgentIDSkipAgentIDSame         = "agent_id_same"
	syncCorrectAgentIDSkipCurrentNodeRunning  = "current_node_running"
	syncCorrectAgentIDSkipGSENotFound         = "gse_not_found"
	syncCorrectAgentIDSkipGSENotRunning       = "gse_not_running"
	syncCorrectAgentIDSkipCloudIDMismatch     = "cloud_id_mismatch"
	syncCorrectAgentIDSkipHostIPMismatch      = "host_ip_mismatch"
)

// NewActionSyncCorrectAgentID creates a new sync correct agent id action.
func NewActionSyncCorrectAgentID(capability *Capability) action.Definition {
	return &actionSyncCorrectAgentID{
		gseHandler: capability.GSEHandler,
		hostStg:    capability.StorageTopo,
	}
}

// ActionParamSyncCorrectAgentID defines the action's param.
type ActionParamSyncCorrectAgentID struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionSyncCorrectAgentID struct {
	gseHandler gse.IHandler
	hostStg    topoStg.IStorageHost
}

type correctItem struct {
	host      *types.Host
	agentInfo *types.AgentInfo
}

type syncCorrectAgentIDStats struct {
	InputCount          int
	HostFoundCount      int
	NeedCorrectCount    int
	GSEFoundCount       int
	CorrectedCount      int
	SkippedCount        int
	RecheckSkippedCount int
	SkipReasons         map[string]int
}

func newSyncCorrectAgentIDStats(inputCount int) *syncCorrectAgentIDStats {
	return &syncCorrectAgentIDStats{
		InputCount:  inputCount,
		SkipReasons: make(map[string]int),
	}
}

func (stats *syncCorrectAgentIDStats) skip(reason string) {
	stats.SkippedCount++
	stats.SkipReasons[reason]++
}

func (stats *syncCorrectAgentIDStats) recheckSkip(reason string) {
	stats.RecheckSkippedCount++
	stats.skip(reason)
}

// Name returns the name of the action.
func (act *actionSyncCorrectAgentID) Name() string {
	return ActionNameSyncCorrectAgentID
}

// Version returns the version of the action.
func (act *actionSyncCorrectAgentID) Version() string {
	return "v1.0.0"
}

// Description returns the description of the action.
func (act *actionSyncCorrectAgentID) Description() string {
	return "correct stale agent id from synced agent id"
}

// Timeout returns the timeout of the action.
func (act *actionSyncCorrectAgentID) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionSyncCorrectAgentID) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionSyncCorrectAgentID) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionSyncCorrectAgentID) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionSyncCorrectAgentID) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncCorrectAgentID)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	stats := newSyncCorrectAgentIDStats(len(param.HostIDs))
	hosts, err := act.filterNeedCorrectHosts(std, stats, param.HostIDs)
	if err != nil {
		return err
	}

	correctedItems, err := act.correctHosts(std, stats, hosts)
	if err != nil {
		return err
	}

	recheckedItems, err := act.recheckItems(std, stats, correctedItems)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to correct agent id, failed to recheck hosts")
		return err
	}

	stats.CorrectedCount = len(recheckedItems)

	if err = act.updateHosts(std, recheckedItems); err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to correct agent id, failed to update host dynamic")
		return err
	}

	act.logStats(ctx, std, stats)

	return nil
}

func (act *actionSyncCorrectAgentID) correctHosts(std *syncDataUtils.SyncDataActionStandarder, stats *syncCorrectAgentIDStats, hosts []*types.Host) (
	[]*correctItem, error) {

	agentInfos, err := act.listAgentInfo(std, hosts)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to correct agent id, failed to list agent info")
		return nil, err
	}
	stats.GSEFoundCount = len(agentInfos)

	validatedItems := make([]*correctItem, 0, len(hosts))
	for _, host := range hosts {
		agentInfo, ok := agentInfos[host.Static.SyncedAgentID]
		if !ok {
			stats.skip(syncCorrectAgentIDSkipGSENotFound)
			continue
		}
		if reason, ok := matchAgentInfoAndHost(host, agentInfo); !ok {
			stats.skip(reason)
			continue
		}

		validatedItems = append(validatedItems, &correctItem{
			host:      host,
			agentInfo: agentInfo,
		})
	}

	return validatedItems, nil
}

func (act *actionSyncCorrectAgentID) filterNeedCorrectHosts(std *syncDataUtils.SyncDataActionStandarder,
	stats *syncCorrectAgentIDStats, hostIDs []int64) ([]*types.Host, error) {

	hosts, err := act.hostStg.ListHostWithoutCount(std.Context(), types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: hostIDs,
		},
	})
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to correct agent id, failed to find hosts")
		return nil, err
	}

	hostMap, err := conv.SliceToMapIgnore(hosts, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to correct agent id, failed to build host map")
		return nil, err
	}

	stats.HostFoundCount = len(hostMap)

	items := make([]*types.Host, 0, len(hosts))
	for _, hostID := range hostIDs {
		host, ok := hostMap[hostID]
		if !ok {
			stats.skip(syncCorrectAgentIDSkipHostNotFound)
			continue
		}

		if reason, need := checkHostNeedCorrect(host); !need {
			stats.skip(reason)
			continue
		}

		items = append(items, host)
	}
	stats.NeedCorrectCount = len(items)

	return items, nil
}

func checkHostNeedCorrect(host *types.Host) (string, bool) {
	if host.Static == nil || host.Static.SyncedAgentID == "" {
		return syncCorrectAgentIDSkipSyncedAgentIDEmpty, false
	}
	if host.Dynamic == nil || host.Dynamic.AgentID == "" {
		return syncCorrectAgentIDSkipDynamicAgentIDEmpty, false
	}
	if host.Static.SyncedAgentID == host.Dynamic.AgentID {
		return syncCorrectAgentIDSkipAgentIDSame, false
	}
	if host.Dynamic.NodeStatus == types.NodeStatusRunning {
		return syncCorrectAgentIDSkipCurrentNodeRunning, false
	}

	return "", true
}

func (act *actionSyncCorrectAgentID) listAgentInfo(std *syncDataUtils.SyncDataActionStandarder,
	items []*types.Host) (map[string]*types.AgentInfo, error) {

	agentIDs := conv.SliceToSlice(items, func(host *types.Host) string {
		return host.Static.SyncedAgentID
	})
	result, err := batchexecutor.Collect(std.Context(), agentIDs,
		func(nCtx contextx.IContext, batchAgentIDs []string) ([]*types.AgentInfo, error) {
			batchResult, err := act.gseHandler.ListAgentInfo(nCtx, batchAgentIDs...)
			if err != nil {
				return nil, fmt.Errorf("list agent info: %w", err)
			}

			return batchResult, nil
		},
		batchexecutor.WithBatchSize(gse.ListAgentInfoPageSize),
		batchexecutor.WithTimeout(act.Timeout()),
	)
	if err != nil {
		return nil, err
	}

	return conv.SliceToMap(result.Items, func(info *types.AgentInfo) string { return info.AgentID })
}

func matchAgentInfoAndHost(host *types.Host, agentInfo *types.AgentInfo) (string, bool) {
	if agentInfo.NodeStatus != types.NodeStatusRunning {
		return syncCorrectAgentIDSkipGSENotRunning, false
	}
	if int64(agentInfo.CloudID) != host.Static.NetworkAreaID {
		return syncCorrectAgentIDSkipCloudIDMismatch, false
	}
	if !slices.Contains(host.Static.GetInnerIPList(), agentInfo.HostIP) {
		return syncCorrectAgentIDSkipHostIPMismatch, false
	}

	return "", true
}

func (act *actionSyncCorrectAgentID) recheckItems(std *syncDataUtils.SyncDataActionStandarder,
	stats *syncCorrectAgentIDStats, items []*correctItem) ([]*correctItem, error) {

	if len(items) == 0 {
		return []*correctItem{}, nil
	}

	hostIDs := conv.SliceToSlice(items, func(item *correctItem) int64 {
		return item.host.HostID
	})
	hosts, err := act.hostStg.ListHostWithoutCount(std.Context(), types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: hostIDs,
		},
	})
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to correct agent id, failed to recheck hosts")
		return nil, err
	}

	hostMap := make(map[int64]*types.Host, len(hosts))
	for _, host := range hosts {
		hostMap[host.HostID] = host
	}

	recheckedItems := make([]*correctItem, 0, len(items))
	for _, item := range items {
		host, ok := hostMap[item.host.HostID]
		if !ok {
			stats.recheckSkip(syncCorrectAgentIDSkipHostNotFound)
			continue
		}
		if reason, need := checkHostNeedCorrect(host); !need {
			stats.recheckSkip(reason)
			continue
		}
		if reason, ok := matchAgentInfoAndHost(host, item.agentInfo); !ok {
			stats.recheckSkip(reason)
			continue
		}

		recheckedItems = append(recheckedItems, &correctItem{
			host:      host,
			agentInfo: item.agentInfo,
		})
	}

	return recheckedItems, nil
}

func (act *actionSyncCorrectAgentID) updateHosts(std *syncDataUtils.SyncDataActionStandarder,
	items []*correctItem) error {

	if len(items) == 0 {
		return nil
	}

	hosts := make([]*types.Host, 0, len(items))
	for _, item := range items {
		hosts = append(hosts, &types.Host{
			HostID: item.host.HostID,
			Dynamic: &types.HostDynamic{
				AgentID:    item.host.Static.SyncedAgentID,
				NodeStatus: item.agentInfo.NodeStatus,
			},
		})
	}

	return batchHandleHosts(std.Context(), hosts, func(hosts ...*types.Host) error {
		return act.hostStg.UpdateHostDynamicFields(std.Context(), types.HostDynamicFields{
			AgentID:    true,
			NodeStatus: true,
		}, hosts...)
	})
}

func (act *actionSyncCorrectAgentID) logStats(ctx *action.InstanceContext, std *syncDataUtils.SyncDataActionStandarder,
	stats *syncCorrectAgentIDStats) {

	logger.G.Sys().Ctx(std.Context()).With(
		"input-count", stats.InputCount,
		"host-found-count", stats.HostFoundCount,
		"need-correct-count", stats.NeedCorrectCount,
		"gse-found-count", stats.GSEFoundCount,
		"corrected-count", stats.CorrectedCount,
		"skipped-count", stats.SkippedCount,
		"recheck-skipped-count", stats.RecheckSkippedCount,
		"skip-reasons", stats.SkipReasons,
	).Info("sync correct agent id finished")

	ctx.Data.Log().
		Zh("Agent ID 修正完成，输入 %d 台，修正 %d 台，跳过 %d 台，重查跳过 %d 台", stats.InputCount,
			stats.CorrectedCount, stats.SkippedCount, stats.RecheckSkippedCount).
		En("sync correct agent id finished, input %d hosts, corrected %d hosts, skipped %d hosts, recheck skipped %d hosts",
			stats.InputCount, stats.CorrectedCount, stats.SkippedCount, stats.RecheckSkippedCount).
		Info()
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncCorrectAgentID) DisplayNameZh() string { return "修正 Agent ID" }

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncCorrectAgentID) DisplayNameEn() string { return "Correct Agent ID" }
