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

package deploypolicy

import (
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
	// DiscoverEnabledPoliciesBySpecifyPlugin discover enabled policies by specify plugin.
	DiscoverEnabledPoliciesBySpecifyPlugin(nCtx contextx.IContext, param *types.SpecifyPluginParam) ([]*types.DeployPolicy, error)

	// RefreshExecuteInfo refresh execute info.
	RefreshExecuteInfo(nCtx contextx.IContext, deployPolicy ...*types.DeployPolicy) error
}
