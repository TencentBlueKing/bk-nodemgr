/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpsertHostToCMDB defines the action name.
	ActionNameUpsertHostToCMDB = "upsert_host_to_cmdb"
)

// NewActionUpsertHostToCMDB get a new action.
func NewActionUpsertHostToCMDB(capability *Capability) action.Definition {
	return &actionUpsertHostToCMDB{
		cmdbHandler:           capability.CMDBHandler,
		storageHost:           capability.StorageTopo,
		storageNodeDeployment: capability.StorageNode,
	}
}

// ActParamUpsertHostToCMDB ...
type ActParamUpsertHostToCMDB struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionUpsertHostToCMDB struct {
	cmdbHandler           cmdb.IHost
	storageHost           topoStg.IStorageHost
	storageNodeDeployment nodeStg.IDaoNodeDeployment
}

// Name returns the name of the action.
func (act *actionUpsertHostToCMDB) Name() string {
	return ActionNameUpsertHostToCMDB
}

// Version returns the version of the action.
func (act *actionUpsertHostToCMDB) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionUpsertHostToCMDB) Description() string {
	return "insert or update host to cmdb"
}

// Timeout returns the timeout of the action.
func (act *actionUpsertHostToCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpsertHostToCMDB) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpsertHostToCMDB) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpsertHostToCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpsertHostToCMDB) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUpsertHostToCMDB)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	if err := act.checkHost(std.Context(), std.DeployInfo()); err != nil {
		return err
	}

	return nil
}

func (act *actionUpsertHostToCMDB) checkHost(nCtx contextx.IContext, info *types.DeploymentInfo) error {
	// nolint: nestif
	// host-id not specified.
	if info.Host.HostID < 0 {
		count, err := act.storageHost.CountHost(nCtx, &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				NetworkAreaID: []int64{info.Host.Static.NetworkAreaID},
				Addressing:    []types.Addressing{info.Host.Static.Addressing},
				InnerIP:       info.Host.Static.InnerIPList,
			},
		})
		if err != nil {
			return err
		}

		if count > 0 {
			return fmt.Errorf("duplicated host with same ip found. networkarea-id(%d), addressing(%s), inner-ip(%v)",
				info.Host.Static.NetworkAreaID, info.Host.Static.Addressing, info.Host.Static.InnerIPList)
		}

		// new host should be inserted into cmdb.
		hostID, err := act.insertHost(nCtx, info)
		if err != nil {
			return fmt.Errorf("insert host to cmdb failed: %w", err)
		}

		info.Host.HostID = hostID
		if err := act.storageHost.UpsertManyHost(nCtx, &info.Host); err != nil {
			return fmt.Errorf("upsert host to db failed: %w", err)
		}

		return nil
	}

	// host-id specified.
	count, err := act.storageHost.CountHost(nCtx, &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID:        []int64{info.Host.HostID},
			NetworkAreaID: []int64{info.Host.Static.NetworkAreaID},
			Addressing:    []types.Addressing{info.Host.Static.Addressing},
			InnerIP:       info.Host.Static.InnerIPList,
		},
	})
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("no host found, contact the system administrator to check the host, "+
			"host_id(%d), networkarea_id(%d), addressing(%s), inner_ip(%v)",
			info.Host.HostID, info.Host.Static.NetworkAreaID, info.Host.Static.Addressing, info.Host.Static.InnerIPList)
	}

	return nil
}

func (act *actionUpsertHostToCMDB) insertHost(nCtx contextx.IContext, info *types.DeploymentInfo) (int64, error) {
	host := &info.Host

	// inorder to check the interface of cc, and set the default architecture at the beginning
	// Here is the historical reason for cc, and can only support x86 architecture and arm architecture
	switch host.Static.OSType {
	case string(criteria.OSWindows), string(criteria.OSLinux):
		host.Static.Arch = string(criteria.CPUArch386)
	case string(criteria.OSDarwin):
		host.Static.Arch = string(criteria.CPUArchArm)
	default:
		host.Static.Arch = string(criteria.CPUArch386)
	}

	hostIDs, err := act.cmdbHandler.AddHostToBusinessIdle(nCtx, info.Host.Static.BizID, host)
	if err != nil {
		return -1, err
	}

	if len(hostIDs) == 0 {
		return -1, errors.New("no host id returned")
	}

	return hostIDs[0], nil
}
