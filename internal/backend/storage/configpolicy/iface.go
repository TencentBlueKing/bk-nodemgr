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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines the config policy storage interface.
type IStorage interface {
	basestorage.Interface
	IDaoConfigPolicy
}

// IDaoConfigPolicy defines the interface for config policy.
type IDaoConfigPolicy interface {
	// MatchConfigPolicy matches the config policy.
	MatchConfigPolicy(ctx context.Context,
		bizID, networkAreaID, networkUnitID int64,
		osType criteria.OSType, cpuArch criteria.CPUArch) (*types.ConfigPolicy, bool, error)

	// CountConfigPolicy counts the config policy by conditions.
	CountConfigPolicy(ctx context.Context, conditions ...*types.ConfigPolicyCondition) (int64, error)

	// ListConfigPolicy lists the config policy by page and conditions.
	ListConfigPolicy(ctx context.Context, page types.Page, conditions ...*types.ConfigPolicyCondition) (
		[]*types.ConfigPolicy, int64, error)

	// GetConfigPolicy gets the config policy.
	GetConfigPolicy(ctx context.Context, configPolicyID int64) (*types.ConfigPolicy, error)

	// CreateConfigPolicy creates the config policy.
	CreateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) (int64, error)

	// UpdateConfigPolicy updates the config policy.
	UpdateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) error

	// DeleteManyConfigPolicy deletes the config policies.
	DeleteManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error

	// EnableManyConfigPolicy enables the config policies.
	EnableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error

	// DisableManyConfigPolicy disables the config policies.
	DisableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error
}
