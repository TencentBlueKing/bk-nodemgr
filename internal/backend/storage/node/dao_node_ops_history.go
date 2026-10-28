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

package node

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/node-ops-history"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) listNodeOpsHistory(
	nCtx contextx.IContext, page types.Page, selection *types.NodeOpsHistoryFieldSelection, conditions ...*types.NodeOpsHistoryCondition) (
	[]*types.NodeOpsHistory, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	opts := s.convertNodeOpsHistoryConditionsToOptions(conditions...)
	histories, total, err := s.daoNodeOpsHistory.ListWithFields(nCtx, page, selection, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list node ops history: %w", err)
	}

	return histories, total, nil
}

func (s *Storage) countNodeOpsHistory(nCtx contextx.IContext, conditions ...*types.NodeOpsHistoryCondition) (int64, error) {
	if nCtx == nil {
		return 0, basestorage.ErrNilContent()
	}
	opts := s.convertNodeOpsHistoryConditionsToOptions(conditions...)
	total, err := s.daoNodeOpsHistory.Count(nCtx, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to count node ops history: %w", err)
	}

	return total, nil
}

func (s *Storage) getNodeOpsHistory(nCtx contextx.IContext, operateID string, selection *types.NodeOpsHistoryFieldSelection) (
	*types.NodeOpsHistory, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}
	if operateID == "" {
		return nil, base.ErrEmptyParamData()
	}

	history, err := s.daoNodeOpsHistory.GetWithFields(nCtx, operateID, selection)
	if err != nil {
		return nil, fmt.Errorf("failed to get node ops history: %w", err)
	}

	return history, nil
}

func (s *Storage) convertNodeOpsHistoryConditionsToOptions(conditions ...*types.NodeOpsHistoryCondition) []nodeopshistory.OptFn {
	opts := make([]nodeopshistory.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}
		if condition.OperateTimeRange != nil {
			opts = append(opts, nodeopshistory.WithOperateTimeRange(*condition.OperateTimeRange))
		}
		if condition.ExactInclude != nil {
			opts = append(opts,
				nodeopshistory.WithBizID(condition.ExactInclude.BizID...),
				nodeopshistory.WithOperateID(condition.ExactInclude.OperateID...),
				nodeopshistory.WithOperator(condition.ExactInclude.Operator...),
				nodeopshistory.WithNodeRole(condition.ExactInclude.NodeRole...))
		}
		if condition.ExactExclude != nil {
			opts = append(opts,
				nodeopshistory.WithoutBizID(condition.ExactExclude.BizID...),
				nodeopshistory.WithoutOperateID(condition.ExactExclude.OperateID...),
				nodeopshistory.WithoutOperator(condition.ExactExclude.Operator...),
				nodeopshistory.WithoutNodeRole(condition.ExactExclude.NodeRole...))
		}
	}

	return opts
}

func (s *Storage) createNodeOpsHistory(nCtx contextx.IContext, history *types.NodeOpsHistory) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}
	if history == nil {
		return basestorage.ErrUpsertNilData()
	}
	if len(history.Snapshots) == 0 {
		return nil
	}
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}
	if err := base.CheckTenantIDMatched(nCtx.TenantID(), history.TenantID); err != nil {
		return err
	}
	if err := validateNodeOpsHistory(history); err != nil {
		return err
	}
	if err := s.daoNodeOpsHistory.Create(nCtx, history); err != nil {
		return fmt.Errorf("failed to create node operation history: %w", err)
	}

	return nil
}

func validateNodeOpsHistory(history *types.NodeOpsHistory) error {
	if history.Info.OperateID == "" || history.Info.OperateTime.IsZero() {
		return errors.New("operation ID and time are required")
	}
	if history.Info.NodeRole != types.NodeRoleBlank &&
		history.Info.NodeRole != types.NodeRoleAgent && history.Info.NodeRole != types.NodeRoleProxy {

		return fmt.Errorf("unsupported operation node role: %s", history.Info.NodeRole)
	}

	hostIDs := make(map[int64]struct{}, len(history.Snapshots))
	for _, snapshot := range history.Snapshots {
		if snapshot == nil || snapshot.HostID <= 0 {
			return base.ErrInvalidItemInParamList()
		}
		if _, ok := hostIDs[snapshot.HostID]; ok {
			return fmt.Errorf("duplicate snapshot host ID: %d", snapshot.HostID)
		}
		hostIDs[snapshot.HostID] = struct{}{}
	}

	return nil
}
