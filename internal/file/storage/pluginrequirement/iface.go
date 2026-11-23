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

package pluginrequirement

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines requirement metadata storage and its lifecycle.
type IStorage interface {
	basestorage.Interface

	// CreatePluginRequirement creates an independent requirement file record.
	CreatePluginRequirement(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error
	// UpdatePluginRequirement updates an existing record and never inserts one.
	UpdatePluginRequirement(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error
	// UpdatePluginRequirementScope updates only applicability and audit metadata.
	UpdatePluginRequirementScope(nCtx contextx.IContext, record *types.PluginRequirementFileRecord) error
	// ListPluginRequirements lists exact matching file metadata with pagination.
	ListPluginRequirements(nCtx contextx.IContext, page types.Page,
		condition *types.PluginRequirementFileCondition) ([]*types.PluginRequirementFileRecord, int64, error)
	// GetPluginRequirement retrieves a record by ID within the current tenant.
	GetPluginRequirement(nCtx contextx.IContext, id string) (*types.PluginRequirementFileRecord, error)
}
