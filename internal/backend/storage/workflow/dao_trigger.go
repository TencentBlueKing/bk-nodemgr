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
func (s *Storage) createTrigger(nCtx contextx.IContext, trig *trigger.Trigger) error {
	if trig == nil {
		return errors.New("trigger is nil")
	}

	return s.daoTrigger.Create(nCtx, trig)
}

// updateTrigger updates a trigger.
func (s *Storage) updateTrigger(nCtx contextx.IContext, trig *trigger.Trigger) error {
	if trig == nil {
		return errors.New("trigger is nil")
	}

	return s.daoTrigger.Update(nCtx, trig)
}

// switchTriggerActive switches a trigger active status.
func (s *Storage) switchTriggerActive(nCtx contextx.IContext, triggerID string, active bool) error {
	return s.daoTrigger.SwitchActive(nCtx, triggerID, active)
}

// getTrigger gets a trigger by triggerID.
func (s *Storage) getTrigger(nCtx contextx.IContext, triggerID string) (*trigger.Trigger, error) {
	return s.daoTrigger.Get(nCtx, triggerID)
}

// listActiveTrigger lists active triggers by category.
func (s *Storage) listActiveTrigger(nCtx contextx.IContext, category trigger.Category) ([]*trigger.Trigger, error) {
	results, _, err := s.daoTrigger.List(nCtx, types.UnlimitedPage(),
		daoTrigger.WithActive(true),
		daoTrigger.WithCategory(category))
	if err != nil {
		return nil, err
	}

	return results, err
}

// listTrigger lists triggers by category.
func (s *Storage) listTrigger(nCtx contextx.IContext, page types.Page, category trigger.Category) ([]*trigger.Trigger, int64, error) {
	return s.daoTrigger.List(nCtx, page, daoTrigger.WithCategory(category))
}

// deleteTriggers deletes triggers by given trigger IDs.
func (s *Storage) deleteTriggers(nCtx contextx.IContext, triggerIDs ...string) error {
	if len(triggerIDs) == 0 {
		return errors.New("triggerIDs is empty")
	}

	return s.daoTrigger.Delete(nCtx, triggerIDs...)
}

// existTrigger checks if a trigger exists by triggerID.
func (s *Storage) existTrigger(nCtx contextx.IContext, triggerID string) (bool, error) {
	cnt, err := s.daoTrigger.Count(nCtx, daoTrigger.WithTriggerID(triggerID))
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}
