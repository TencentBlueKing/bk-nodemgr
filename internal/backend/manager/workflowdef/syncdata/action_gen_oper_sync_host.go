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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameGenOperSyncHost defines the action name.
	ActionNameGenOperSyncHost = "gen_oper_sync_host"
)

// NewActionGenOperSyncHost this action will create host sync operation for all business.
func NewActionGenOperSyncHost(storageBusiness topo.IStorageBusiness, workflowCtl workflow.IController) action.Definition {
	return &actionGenOperSyncHost{
		storageBusiness: storageBusiness,
		workflowCtl:     workflowCtl,
	}
}

// GenOperSyncHostParam ...
type GenOperSyncHostParam struct {
	TenantID string `json:"tenant_id"`
}

type actionGenOperSyncHost struct {
	storageBusiness topo.IStorageBusiness
	workflowCtl     workflow.IController
}

// Name returns the name of the action.
func (act *actionGenOperSyncHost) Name() string {
	return ActionNameGenOperSyncHost
}

// Version returns the version of the action.
func (act *actionGenOperSyncHost) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenOperSyncHost) Description() string {
	return "reads all business information from the database," +
		"and creates host synchronization tasks on a business-by-business basis"
}

// Timeout returns the timeout of the action.
func (act *actionGenOperSyncHost) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenOperSyncHost) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenOperSyncHost) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenOperSyncHost) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenOperSyncHost) Do(ctx *action.InstanceContext) error {
	param := new(GenOperSyncHostParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	bizs, _, err := act.storageBusiness.ListBusinesses(tenantCtx, types.Page{})
	if err != nil {
		return err
	}

	ctx.Data.LogI(fmt.Sprintf("found business, lens(%d)", len(bizs)))

	for idx := range bizs {
		biz := bizs[idx]

		if err = act.executeOper(ctx, biz); err != nil {
			return err
		}
	}

	return nil
}

// executeOper create an operation to sync all host from cmdb and then execute it.
func (act *actionGenOperSyncHost) executeOper(
	ctx *action.InstanceContext,
	biz *types.Business) error {

	trigCtl, err := act.workflowCtl.GetTrigger(ctx.Ctx, ctx.Data.TriggerID)
	if err != nil {
		err = fmt.Errorf(
			"failed to get trigger. tenant-id(%s), trigger-id(%s), biz-name(%s), biz-id(%d), err: %w",
			biz.TenantID, ctx.Data.TriggerID, biz.BizName, biz.BizID, err)

		ctx.Data.LogE(err.Error())

		return err
	}

	operationDef := NewOperSyncHostFromCMDB(OperParamSyncHostFromCMDB{
		TenantID: biz.TenantID,
		BizID:    biz.BizID,
	})

	operationParam := operationDef.DefaultParameters()
	operationParam.ParentOperationID = ctx.Data.OperationID

	operCtl, err := trigCtl.CreateOperation(ctx.Ctx, operationDef, operationParam)
	if err != nil {
		err = fmt.Errorf(
			"failed to create sync host operation for business, tenant-id(%s), biz-name(%s), biz-id(%d), err: %w",
			biz.TenantID, biz.BizName, biz.BizID, err)

		ctx.Data.LogE(err.Error())

		return err
	}

	ctx.Data.LogI(
		fmt.Sprintf("created sync host operation for business, tenant-id(%s), operation-id(%s), biz-name(%s), biz-id(%d)",
			biz.TenantID, operCtl.GetOperationID(), biz.BizName, biz.BizID))

	return nil
}
