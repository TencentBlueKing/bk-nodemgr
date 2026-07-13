/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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

	// ListConfigPolicyWithoutCount lists the config policy by page and conditions without count.
	ListConfigPolicyWithoutCount(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyCondition) (
		[]*types.ConfigPolicy, error)

	// GetConfigPolicy gets the config policy.
	GetConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error)

	// CreateConfigPolicy creates the config policy.
	CreateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error)

	// UpdateConfigPolicy updates the config policy.
	UpdateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) error

	// DeleteManyConfigPolicy deletes the config policies.
	DeleteManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// EnableManyConfigPolicy enables the config policies.
	EnableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// DisableManyConfigPolicy disables the config policies.
	DisableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// UpdatePriorityManyConfigPolicy batch-updates the priority field for the given policy IDs.
	UpdatePriorityManyConfigPolicy(nCtx contextx.IContext, priorities map[int64]int64) error
}

// IDaoConfigPolicyNode defines the interface for config policy node.
type IDaoConfigPolicyNode interface {
	// MatchConfigPolicyNode matches enabled policies for the node, merges them by priority.
	MatchConfigPolicyNode(nCtx contextx.IContext,
		bizID, networkAreaID, networkUnitID int64,
		osType criteria.OSType, cpuArch criteria.CPUArch,
		nodeRole types.NodeRole, hostID int64) (*types.ConfigPolicyMatchResult, error)

	// PreviewConfigPolicy previews the merged config for each host by matching enabled policies.
	PreviewConfigPolicy(nCtx contextx.IContext,
		bizID int64, policyType types.ConfigPolicyType,
		hosts []types.ConfigPolicyPreviewHost) (*types.ConfigPolicyPreviewResult, error)
}

// IDaoConfigPolicyEvent defines the interface for policy event.
type IDaoConfigPolicyEvent interface {
	// CountConfigPolicyEvent counts topo events by conditions.
	CountConfigPolicyEvent(nCtx contextx.IContext, conditions ...*types.ConfigPolicyEventCondition) (int64, error)

	// ListConfigPolicyEvent lists topo events by page and conditions.
	ListConfigPolicyEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyEventCondition) (
		[]*types.ConfigPolicyEvent, int64, error)

	// ListConfigPolicyEventWithoutCount lists policy events by page and conditions without count.
	ListConfigPolicyEventWithoutCount(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyEventCondition) (
		[]*types.ConfigPolicyEvent, error)

	// CreateManyConfigPolicyEvent creates multiple topo events.
	CreateManyConfigPolicyEvent(nCtx contextx.IContext, events ...*types.ConfigPolicyEvent) error

	// DistinctConfigPolicyEvent distincts topoevent fields.
	DistinctConfigPolicyEvent(
		nCtx contextx.IContext, request types.ConfigPolicyEventDistinctRequest, conditions ...*types.ConfigPolicyEventCondition) (
		*types.ConfigPolicyEventDistinctResult, error)
}
