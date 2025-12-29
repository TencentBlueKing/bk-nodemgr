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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the storage interface.
type IStorage interface {
	basestorage.Interface

	IDaoDeployPolicy
	IDomainDeployPolicyMgr
}

// IDaoDeployPolicy defines the deploy policy dao interface.
type IDaoDeployPolicy interface {
	// CreateDeployPolicy create deploy policy.
	CreateDeployPolicy(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error)

	// ListDeployPolicies list deploy policies.
	ListDeployPolicies(nCtx contextx.IContext, page types.Page, condition *types.DeployPolicyCondition) ([]*types.DeployPolicy, int64, error)

	// GetDeployPolicyByID get deploy policy by id.
	GetDeployPolicyByID(nCtx contextx.IContext, deployPolicyID int64) (*types.DeployPolicy, error)

	// UpdateDeployPolicyFields update deploy policy fields.
	UpdateDeployPolicyFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error

	// DeleteDeployPolicy delete deploy policy.
	DeleteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) error

	// ExistDeployPolicy check deploy policy exist.
	ExistDeployPolicy(nCtx contextx.IContext, condition *types.DeployPolicyCondition) (bool, error)
}

// IDomainDeployPolicyMgr defines the deploy policy manager interface.
type IDomainDeployPolicyMgr interface {
	// DiscoverPoliciesBySpecifyPlugin discover policies by specify plugin.
	DiscoverPoliciesBySpecifyPlugin(nCtx contextx.IContext, param *types.SpecifyPluginParam) ([]*types.DeployPolicy, error)

	// UpdateDeployPoliciesExecutedAt update deploy policies executed at.
	UpdateDeployPoliciesExecutedAt(nCtx contextx.IContext, deployPolicyIDs []int64, executedAt time.Time) error
}
