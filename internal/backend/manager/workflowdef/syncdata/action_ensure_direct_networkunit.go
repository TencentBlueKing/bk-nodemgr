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

package syncdata

import (
	"fmt"
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnsureDirectNetworkUnit defines the action name.
	ActionNameEnsureDirectNetworkUnit = "ensure_direct_networkunit"
)

// NewActionEnsureDirectNetworkUnit get a new action.
func NewActionEnsureDirectNetworkUnit(capability *Capability) action.Definition {
	return &actionEnsureDirectNetworkUnit{
		storageNetworkArea: capability.StorageTopo,
		storageNetworkUnit: capability.StorageTopo,
		networkUnitConfig:  capability.NetworkUnitConfig,
	}
}

// ActionParamEnsureDirectNetworkUnit describes the parameters.
type ActionParamEnsureDirectNetworkUnit struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionEnsureDirectNetworkUnit struct {
	storageNetworkArea topoStg.IStorageNetworkArea
	storageNetworkUnit topoStg.IStorageNetworkUnit
	networkUnitConfig  config.NetworkUnit
}

// Name returns the name of the action.
func (act *actionEnsureDirectNetworkUnit) Name() string {
	return ActionNameEnsureDirectNetworkUnit
}

// Version returns the version of the action.
func (act *actionEnsureDirectNetworkUnit) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionEnsureDirectNetworkUnit) Description() string {
	return "ensure a default direct network unit exists in the default network area (id=0) for the tenant"
}

// Timeout returns the timeout of the action.
func (act *actionEnsureDirectNetworkUnit) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionEnsureDirectNetworkUnit) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the retry count of the action.
func (act *actionEnsureDirectNetworkUnit) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the delay function.
func (act *actionEnsureDirectNetworkUnit) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do does the action.
func (act *actionEnsureDirectNetworkUnit) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamEnsureDirectNetworkUnit)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	nCtx := std.Context()

	directUnitCfg := act.networkUnitConfig.DefaultDirectUnit
	if !directUnitCfg.Enabled {
		return nil
	}

	// ensure the default network area has been synced before creating the direct unit.
	_, areaNum, err := act.storageNetworkArea.ListNetworkArea(nCtx, types.SingleItemPage(), &types.NetworkAreaCondition{
		ExactInclude: &types.NetworkAreaExactFields{
			NetworkAreaID: []int64{types.DefaultNetworkAreaID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to list default network area: %w", err)
	}

	if areaNum == 0 {
		// default network area not synced yet, skip and wait for the next round.
		return nil
	}

	// idempotent check: skip when a direct unit already exists in the default area.
	_, num, err := act.storageNetworkUnit.ListNetworkUnit(nCtx, types.SingleItemPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkAreaID: []int64{types.DefaultNetworkAreaID},
			IsDirect:      []bool{true},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to list direct network unit in default area: %w", err)
	}

	if num > 0 {
		return nil
	}

	networkUnit := &types.NetworkUnit{
		TenantID:      std.TenantID(),
		NetworkAreaID: types.DefaultNetworkAreaID,
		Name:          directUnitCfg.Name,
		IsDirect:      true,
		DirectEndpoints: &types.Endpoints{
			Cluster: directUnitCfg.ClusterEndpoints,
			File:    directUnitCfg.FileEndpoints,
			Data:    directUnitCfg.DataEndpoints,
		},
		Generation: types.Generation2,
	}

	if _, _, err = act.storageNetworkUnit.CreateNetworkUnit(nCtx, networkUnit); err != nil {
		return fmt.Errorf("failed to create default direct network unit: %w", err)
	}

	ctx.Data.Log().
		Zh("已创建默认直连网络单元, name: %s", directUnitCfg.Name).
		En("created default direct network unit, name: %s", directUnitCfg.Name).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionEnsureDirectNetworkUnit) DisplayNameZh() string {
	return "确保默认直连网络单元"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionEnsureDirectNetworkUnit) DisplayNameEn() string {
	return "Ensure Default Direct Network Unit"
}
