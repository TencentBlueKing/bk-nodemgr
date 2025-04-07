/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionGenAllBizHostSyncOper this action will create host sync operation for all business.
func NewActionGenAllBizHostSyncOper(topoStorage topo.IStorage, operMgr operengine.OperationMgr) operengine.ActionDef {
	return &genAllBizHostSyncOper{
		topoStorage: topoStorage,
		operMgr:     operMgr,
	}
}

// GenAllBizHostSyncOperParam ...
type GenAllBizHostSyncOperParam struct {
	TenantID string `json:"tenant_id"`
}

// genAllBizHostSyncOper ...
type genAllBizHostSyncOper struct {
	topoStorage topo.IStorage
	operMgr     operengine.OperationMgr
}

// Name returns the name of the action.
func (c *genAllBizHostSyncOper) Name() string {
	return GenAllBizHostSyncOper
}

// Version returns the version of the action.
func (c *genAllBizHostSyncOper) Version() string {
	return "v1.0.0"
}

// Description returns the description of the action.
func (c *genAllBizHostSyncOper) Description() string {
	return "reads all business information from the database," +
		"and creates host synchronization tasks on a business-by-business basis."
}

// Timeout returns the timeout of the action.
func (c *genAllBizHostSyncOper) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (c *genAllBizHostSyncOper) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (c *genAllBizHostSyncOper) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (c *genAllBizHostSyncOper) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (c *genAllBizHostSyncOper) Do(ctx *operengine.ActionInstContext) error {
	param := new(GenAllBizHostSyncOperParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	bizs, _, err := c.topoStorage.ListBusinesses(tenantCtx, types.Page{})
	if err != nil {
		return err
	}

	ctx.Data.Log(fmt.Sprintf("found %d business", len(bizs)))

	for idx := range bizs {
		biz := bizs[idx]

		if err = c.executeOper(ctx.Data, biz); err != nil {
			return err
		}
	}

	return nil
}

// executeOper create an operation to sync all host from cmdb and then execute it.
func (c *genAllBizHostSyncOper) executeOper(data *operengine.ActionInstData, biz *types.Business) error {
	operation := newOperSyncHostFromCMDB(data.TriggerID)
	err := c.operMgr.ExecuteOperation(operation, &operengine.OperInstParam{
		Timeout: time.Minute * 10, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(SyncHostFromCMDBParam{
			BizID:    biz.BizID,
			TenantID: biz.TenantID,
		}),
		ParentOperationID: data.OperationID,
	})
	if err != nil {
		msg := fmt.Sprintf("failed to create sync host operation for business, tenant-id(%s), biz-name(%s), biz-id(%d)",
			biz.TenantID, biz.BizName, biz.BizID)
		data.Log(msg)

		return err
	}

	msg := fmt.Sprintf("successfully create sync host operation for business, tenant-id(%s), biz-name(%s), biz-id(%d)",
		biz.TenantID, biz.BizName, biz.BizID)
	data.Log(msg)

	return nil
}
