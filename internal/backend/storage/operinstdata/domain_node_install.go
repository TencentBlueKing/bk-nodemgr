/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package operinstdata

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

// IDomainNodeInstall defines the Storage interface for domain node install.
type IDomainNodeInstall interface {
	// UpdateOperInstActionStatus will update the oper inst action status.
	UpdateOperInstActionStatus(ctx context.Context, operInstID string, actionName string,
		status action.State) error

	// PushActInstMsgs will append the oper inst action log.
	PushActionInstanceMessage(ctx context.Context, operInstID string, actionName string,
		msgs ...common.Message) error

	// UpsertActionInstancePrivateData upserts action instance private data.
	UpsertActionInstancePrivateData(
		ctx context.Context, operInstID string, actionName string, privateData map[string]any) error
}

// UpdateOperInstActionStatus update the oper inst action status.
func (s *Storage) UpdateOperInstActionStatus(
	ctx context.Context, operInstID string, actionName string, status action.State) (err error) {

	// record metric.
	metric := s.metric().Start("update_action_status")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("oper inst id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	if err = s.daoOperinstdata.UpdateActionInstStatus(ctx, operInstID, actionName, status); err != nil {
		return fmt.Errorf("update oper inst action status error: %v", err)
	}

	return nil
}
