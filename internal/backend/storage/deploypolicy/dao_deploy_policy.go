/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoDeployPolicy "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) createDeployPolicy(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := s.daoDeployPolicy.Create(nCtx, deployPolicy); err != nil {
		return fmt.Errorf("failed to create deploy policy: %w", err)
	}

	return nil
}

func (s *Storage) listDeployPolicies(nCtx contextx.IContext, page types.Page) ([]*types.DeployPolicy, int64, error) {
	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	deployPolicies, total, err := s.daoDeployPolicy.List(nCtx, page)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list deploy policies: %w", err)
	}

	return deployPolicies, total, nil
}

func (s *Storage) getDeployPolicyByID(nCtx contextx.IContext, deployPolicyID int64) (*types.DeployPolicy, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	deployPolicy, err := s.daoDeployPolicy.Get(nCtx, daoDeployPolicy.WithDeployPolicyID(deployPolicyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy policy: %w", err)
	}

	return deployPolicy, nil
}

func (s *Storage) updateDeployPolicy(nCtx contextx.IContext, deployPolicyID int64, deployPolicy *types.DeployPolicy) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := s.daoDeployPolicy.Update(nCtx, deployPolicyID, deployPolicy); err != nil {
		return fmt.Errorf("failed to update deploy policy: %w", err)
	}

	return nil
}

func (s *Storage) deleteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := s.daoDeployPolicy.Delete(nCtx, deployPolicyID); err != nil {
		return fmt.Errorf("failed to delete deploy policy: %w", err)
	}

	return nil
}

func (s *Storage) existDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) (bool, error) {
	if nCtx == nil {
		return false, base.ErrInvalidContext()
	}

	exist, err := s.daoDeployPolicy.Exist(nCtx, deployPolicyID)
	if err != nil {
		return false, fmt.Errorf("failed to check deploy policy exist: %w", err)
	}

	return exist, nil
}
