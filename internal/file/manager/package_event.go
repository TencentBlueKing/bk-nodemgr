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

package manager

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IPackageEvent defines package event query operations.
type IPackageEvent interface {
	// CountPackageEvent counts package event records by conditions.
	CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error)

	// ListPackageEvent lists package event records by page and conditions.
	ListPackageEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.PackageEventCondition) ([]*types.PackageEvent, int64, error)

	// DistinctPackageEvent gets distinct package event fields.
	DistinctPackageEvent(nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
		*types.PackageEventDistinctResult, error)
}

// CountPackageEvent counts package event records by conditions.
func (m *Manager) CountPackageEvent(nCtx contextx.IContext, conditions ...*types.PackageEventCondition) (int64, error) {
	return m.storageEvent.CountPackageEvent(nCtx, conditions...)
}

// ListPackageEvent lists package events by page and conditions.
func (m *Manager) ListPackageEvent(nCtx contextx.IContext, page types.Page,
	conditions ...*types.PackageEventCondition) ([]*types.PackageEvent, int64, error) {

	return m.storageEvent.ListPackageEvent(nCtx, page, conditions...)
}

// DistinctPackageEvent gets distinct package event fields.
func (m *Manager) DistinctPackageEvent(
	nCtx contextx.IContext, request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) (
	*types.PackageEventDistinctResult, error) {

	return m.storageEvent.DistinctPackageEvent(nCtx, request, conditions...)
}
