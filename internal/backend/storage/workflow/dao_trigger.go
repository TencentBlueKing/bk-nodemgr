/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoTrigger "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// createTrigger creates a new trigger.
func (s *Storage) createTrigger(ctx contextx.IContext, trig *trigger.Trigger) error {
	if trig == nil {
		return errors.New("trigger is nil")
	}

	return s.daoTrigger.Create(ctx, trig)
}

// updateTrigger updates a trigger.
func (s *Storage) updateTrigger(ctx contextx.IContext, trig *trigger.Trigger) error {
	if trig == nil {
		return errors.New("trigger is nil")
	}

	return s.daoTrigger.Update(ctx, trig)
}

// updateTriggerState updates a trigger's state.
func (s *Storage) updateTriggerState(ctx contextx.IContext, triggerID string, state trigger.State) error {
	return s.daoTrigger.UpdateState(ctx, triggerID, state)
}

// getTrigger gets a trigger by triggerID.
func (s *Storage) getTrigger(ctx contextx.IContext, triggerID string) (*trigger.Trigger, error) {
	return s.daoTrigger.Get(ctx, triggerID)
}

// listAliveTrigger lists alive triggers by category.
func (s *Storage) listAliveTrigger(ctx contextx.IContext, category trigger.Category) ([]*trigger.Trigger, error) {
	results, _, err := s.daoTrigger.List(ctx, types.UnlimitedPage(),
		daoTrigger.WithState(trigger.StateInit, trigger.StateRunning),
		daoTrigger.WithCategory(category))
	if err != nil {
		return nil, err
	}

	return results, err
}

// deleteTriggers deletes triggers by given trigger IDs.
func (s *Storage) deleteTriggers(ctx contextx.IContext, triggerIDs ...string) error {
	if len(triggerIDs) == 0 {
		return errors.New("triggerIDs is empty")
	}

	return s.daoTrigger.Delete(ctx, triggerIDs...)
}
