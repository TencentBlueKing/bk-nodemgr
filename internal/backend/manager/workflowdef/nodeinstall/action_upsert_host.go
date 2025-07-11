/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"errors"
	"fmt"
	"time"

	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpsertHostToCMDB defines the action name.
	ActionNameUpsertHostToCMDB = "upsert_host_to_cmdb"
)

// NewActionUpsertHostToCMDB get a new action.
func NewActionUpsertHostToCMDB(
	cmdbHandler cmdb.IHandler,
	storageHost topo.IStorageHost,
	storageNodeDeployment nodedeployment.IStorageNodeDeployment,
) action.Definition {

	return &actionUpsertHostToCMDB{
		cmdbHandler:           cmdbHandler,
		storageHost:           storageHost,
		storageNodeDeployment: storageNodeDeployment,
	}
}

// ActParamUpsertHostToCMDB ...
type ActParamUpsertHostToCMDB struct {
	Token string `json:"token"`
}

type actionUpsertHostToCMDB struct {
	cmdbHandler           cmdb.IHost
	storageHost           topo.IStorageHost
	storageNodeDeployment nodedeployment.IStorageNodeDeployment
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
// nolint: funlen,fnsize
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionUpsertHostToCMDB) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUpsertHostToCMDB)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	info, err := act.storageNodeDeployment.GetInfo(ctx.Ctx, param.Token)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, info.Host.TenantID)
	if err != nil {
		return err
	}

	if err := act.checkHost(tenantCtx, info); err != nil {
		return err
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		if err := act.storageNodeDeployment.UpdateInfo(ctx.Ctx, param.Token, info); err != nil {
			return fmt.Errorf("update node deployment info failed, err: %w", err)
		}

		return nil
	})

	gp.Go(func() error {
		if err := act.storageHost.UpsertManyHost(tenantCtx, &info.Host); err != nil {
			return fmt.Errorf("upsert host to db failed, err: %w", err)
		}

		return nil
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("wait group failed, err: %w", err)
	}

	return nil
}

func (act *actionUpsertHostToCMDB) checkHost(ctx context.Context, info *types.DeploymentInfo) error {
	// nolint: nestif
	if info.Host.HostID < 0 {
		hosts, count, err := act.storageHost.ListHost(ctx, types.Page{
			Offset: 0,
			Limit:  1,
		}, &types.HostCondition{
			ExactInclude: &types.HostExactFields{
				NetworkAreaID: []int64{info.Host.Static.NetworkAreaID},
				Addressing:    []types.Addressing{info.Host.Static.Addressing},
				InnerIP:       []string{info.Host.Static.InnerIP},
			},
		})
		if err != nil {
			return err
		}

		if count > 1 {
			return fmt.Errorf("more than one host found, connect the system administrator to check the host, "+
				"network_area_id(%d), addressing(%s), inner_ip(%s)",
				info.Host.Static.NetworkAreaID, info.Host.Static.Addressing, info.Host.Static.InnerIP)
		}

		if len(hosts) == 0 {
			info.Host.HostID, err = act.insertHost(ctx, info)
			if err != nil {
				return err
			}
		} else {
			info.Host.HostID = hosts[0].HostID
		}
	} else {
		count, err := act.storageHost.CountHost(ctx, &types.HostCondition{
			ExactInclude: &types.HostExactFields{
				HostID:        []int64{info.Host.HostID},
				NetworkAreaID: []int64{info.Host.Static.NetworkAreaID},
				Addressing:    []types.Addressing{info.Host.Static.Addressing},
				InnerIP:       []string{info.Host.Static.InnerIP},
			},
		})
		if err != nil {
			return err
		}

		if count == 0 {
			return fmt.Errorf("no host found, connect the system administrator to check the host, "+
				"network_area_id(%d), addressing(%s), inner_ip(%s)",
				info.Host.Static.NetworkAreaID, info.Host.Static.Addressing, info.Host.Static.InnerIP)
		}
	}

	return nil
}

func (act *actionUpsertHostToCMDB) insertHost(ctx context.Context, info *types.DeploymentInfo) (int64, error) {
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

	hostIDs, err := act.cmdbHandler.AddHostToBusinessIdle(ctx, info.Host.Static.BizID, host)
	if err != nil {
		return 0, err
	}

	if len(hostIDs) == 0 {
		return 0, errors.New("no host id returned")
	}

	return hostIDs[0], nil
}
