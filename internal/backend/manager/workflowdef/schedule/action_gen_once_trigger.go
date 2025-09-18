/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package schedule provides the action to generate a schedule once trigger for workflow.
package schedule

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameGenScheduleOnceTrigger defines the action name.
	ActionNameGenScheduleOnceTrigger = "gen_once_trigger_%s"
)

// OnceTriggerFunc defines the function type for generating a schedule once trigger.
type OnceTriggerFunc func(ctx contextx.ITenantUserContext) (string, error)

// NewActionGenScheduleOnceTrigger creates a new action to generate a schedule once trigger.
func NewActionGenScheduleOnceTrigger(
	name string,
	operInstCtl workflow.IStorageOperationInstance,
	operFunc OnceTriggerFunc) action.Definition {

	return &actionGenScheduleOnceTrigger{
		name:        name,
		operInstCtl: operInstCtl,
		operFunc:    operFunc,
	}
}

// GenScheduleOnceTriggerParam ...
type GenScheduleOnceTriggerParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// actionGenScheduleOnceTrigger implements the action.Definition interface.
type actionGenScheduleOnceTrigger struct {
	name        string
	operInstCtl workflow.IStorageOperationInstance
	operFunc    OnceTriggerFunc
}

// Name returns the name of the action.
func (act *actionGenScheduleOnceTrigger) Name() string {
	return fmt.Sprintf(ActionNameGenScheduleOnceTrigger, act.name)
}

// Version returns the version of the action.
func (act *actionGenScheduleOnceTrigger) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionGenScheduleOnceTrigger) Description() string {
	return "generate a schedule once trigger to do schedule workflow."
}

// Timeout returns the timeout of the action.
func (act *actionGenScheduleOnceTrigger) Timeout() time.Duration {
	return time.Second * 10 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionGenScheduleOnceTrigger) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionGenScheduleOnceTrigger) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionGenScheduleOnceTrigger) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionGenScheduleOnceTrigger) Do(ctx *action.InstanceContext) error {
	param := new(GenScheduleOnceTriggerParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	tenantUserCtx := contextx.NewTenantUserContext(tenantCtx, param.TenantID, param.Operator)
	triggerID, err := act.operFunc(tenantUserCtx)
	if err != nil {
		return fmt.Errorf("generate schedule once trigger action-name(%s) by tenant-id(%s) failed: %w",
			act.Name(), param.TenantID, err)
	}

	ctx.Data.PrivateData["child_trigger_id"] = triggerID

	return nil
}
