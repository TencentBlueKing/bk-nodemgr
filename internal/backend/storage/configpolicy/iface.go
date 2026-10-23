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

package configpolicy

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the config policy storage interface.
type IStorage interface {
	basestorage.Interface
	IDaoConfigPolicy
	IDaoConfigPolicyEvent
	IDaoConfigPolicyNode
}

// IDaoConfigPolicy defines the interface for config policy.
type IDaoConfigPolicy interface {
	// CountConfigPolicy counts the config policy by conditions.
	CountConfigPolicy(nCtx contextx.IContext, conditions ...*types.ConfigPolicyCondition) (int64, error)

	// ListConfigPolicy lists the config policy by page and conditions.
	ListConfigPolicy(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyCondition) (
		[]*types.ConfigPolicy, int64, error)

	// GetConfigPolicy gets the config policy.
	GetConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error)

	// CreateConfigPolicy creates the policy and returns its inserted snapshot.
	CreateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (*types.ConfigPolicy, error)

	// UpdateConfigPolicy returns post-update snapshots of matched policies.
	UpdateConfigPolicy(nCtx contextx.IContext, fields types.ConfigPolicyFields, configPolicy *types.ConfigPolicy) ([]*types.ConfigPolicy, error)

	// DeleteManyConfigPolicy returns pre-delete snapshots, including successful mutations before an error.
	DeleteManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error)

	// EnableManyConfigPolicy returns post-enable snapshots, including successful mutations before an error.
	EnableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error)

	// DisableManyConfigPolicy returns post-disable snapshots, including successful mutations before an error.
	DisableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error)

	// UpdatePriorityManyConfigPolicy returns post-update snapshots, including successful mutations before an error.
	UpdatePriorityManyConfigPolicy(nCtx contextx.IContext, priorities map[int64]int64) ([]*types.ConfigPolicy, error)
}

// IDaoConfigPolicyNode defines the interface for config policy node.
type IDaoConfigPolicyNode interface {
	// MatchConfigPolicyNode matches enabled policies for the node, merges them by priority.
	MatchConfigPolicyNode(nCtx contextx.IContext,
		bizID, networkAreaID, networkUnitID int64,
		osType criteria.OSType, cpuArch criteria.CPUArch,
		nodeRole types.NodeRole, hostID int64) (*types.ConfigPolicyMatchResult, error)

	// MatchConfigPolicyPlugin matches and merges enabled policies for a plugin on the host.
	MatchConfigPolicyPlugin(nCtx contextx.IContext,
		bizID, networkAreaID, networkUnitID int64,
		osType criteria.OSType, cpuArch criteria.CPUArch,
		targetPluginName string, hostID int64) (*types.ConfigPolicyMatchResult, error)

	// PreviewConfigPolicy previews the merged config for each host by matching enabled policies.
	// orderedConfigPolicyID optionally overrides the DB priority order with the user-specified
	// priority order, highest priority first; empty falls back to the DB priority order.
	PreviewConfigPolicy(nCtx contextx.IContext,
		bizID int64, policyType types.ConfigPolicyType, targetPluginName string,
		hosts []types.ConfigPolicyPreviewHost,
		orderedConfigPolicyID []int64) (*types.ConfigPolicyPreviewResult, error)
}

// IDaoConfigPolicyEvent defines the interface for policy event.
type IDaoConfigPolicyEvent interface {
	// CountConfigPolicyEvent counts topo events by conditions.
	CountConfigPolicyEvent(nCtx contextx.IContext, conditions ...*types.ConfigPolicyEventCondition) (int64, error)

	// ListConfigPolicyEvent lists topo events by page and conditions.
	ListConfigPolicyEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyEventCondition) (
		[]*types.ConfigPolicyEvent, int64, error)

	// CreateManyConfigPolicyEvent creates multiple topo events.
	CreateManyConfigPolicyEvent(nCtx contextx.IContext, events ...*types.ConfigPolicyEvent) error

	// DistinctConfigPolicyEvent distincts topoevent fields.
	DistinctConfigPolicyEvent(
		nCtx contextx.IContext, request types.ConfigPolicyEventDistinctRequest, conditions ...*types.ConfigPolicyEventCondition) (
		*types.ConfigPolicyEventDistinctResult, error)
}
