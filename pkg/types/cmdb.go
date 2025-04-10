/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"fmt"
)

// CCCloudVendor represents a cloud vendor option.
type CCCloudVendor struct {
	Key  string
	Name string
}

// CCOsType represents a os type option.
type CCOsType struct {
	Key  string
	Name string
}

// DynamicGroup represents the dynamic group of cmdb.
type DynamicGroup struct {
	ID    string
	BizID int64
	ObjID string
	Name  string
}

// ServiceTemplate represents a service template option.
type ServiceTemplate struct {
	ID                  int64
	BizID               int64
	ServiceTemplateName string
	ServiceCategoryID   int64
	HostApplyEnabled    bool
}

// HostRel represents the host relation of cmdb.
type HostRel struct {
	HostID   int64
	BizID    int64
	ModuleID int64
	SetID    int64
}

// ChangeType represents the change type of cmdb.
type ChangeType string

const (
	// ChangeTypeCreate represents the create change type.
	ChangeTypeCreate ChangeType = "create"

	// ChangeTypeUpdate represents the update change type.
	ChangeTypeUpdate ChangeType = "update"

	// ChangeTypeDelete represents the delete change type.
	ChangeTypeDelete ChangeType = "delete"
)

// Validate validates the change type.
func (changeType ChangeType) Validate() error {
	switch changeType {
	case ChangeTypeCreate, ChangeTypeUpdate, ChangeTypeDelete:
		return nil
	default:
		return fmt.Errorf("invalid change type, type(%s)", changeType)
	}
}

// ChangeEvent represents the change event of cmdb.
type ChangeEvent[T any] struct {
	ChangeType ChangeType
	Detail     T
}

// Validate validates the change event.
func (e *ChangeEvent[T]) Validate() error {
	if err := e.ChangeType.Validate(); err != nil {
		return fmt.Errorf("invalid change type, err: %w", err)
	}

	return nil
}
