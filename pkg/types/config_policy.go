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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// ConfigPolicy defines the config policy.
type ConfigPolicy struct {
	TenantID  string
	Version   int
	ID        int64
	Name      string
	NodeRole  NodeRole
	BizID     []int64
	Remark    string
	Scopes    []ConfigPolicyScope
	Configs   map[string]any
	Enabled   bool
	UpdatedAt time.Time
	Operator  string
}

// ConfigPolicyScope defines the config policy scope.
type ConfigPolicyScope struct {
	NetworkAreaID int64
	NetworkUnitID int64
	NodeOsType    criteria.OSType
	NodeCPUArch   criteria.CPUArch
}

// ConfigPolicyTemplate defines the config policy template.
type ConfigPolicyTemplate struct {
	TenantID string
	ID       int64
	Template string
}

// ConfigPolicyTemplateBlock defines the config policy template block.
type ConfigPolicyTemplateBlock struct {
	ID      string                     `json:"id"`
	TitleEN string                     `json:"title_en"`
	TitleZH string                     `json:"title_zh"`
	Items   []ConfigPolicyTemplateItem `json:"items"`
}

// ConfigPolicyTemplateItem defines the config policy template item.
type ConfigPolicyTemplateItem struct {
	Enabled           bool                     `json:"enabled"`
	ID                string                   `json:"id"`
	NameEN            string                   `json:"name_en"`
	NameZH            string                   `json:"name_zh"`
	RemarkEN          string                   `json:"remark_en"`
	RemarkZH          string                   `json:"remark_zh"`
	Key               string                   `json:"key"`
	Type              ConfigPolicyTemplateType `json:"type"`
	ValueString       string                   `json:"value_string"`
	ValueInt          int64                    `json:"value_int"`
	ValueBool         bool                     `json:"value_bool"`
	ValueStringSelect []string                 `json:"value_string_select"`
	ValueIntSelect    []int64                  `json:"value_int_select"`

	/**
	 * ValueGroup is only for definition, will not return to frontend.
	**/

	// ValueGroupFixed defines the fixed value to set.
	ValueGroupFixed map[string]any `json:"value_group_fixed"`

	// ValueGroupAssigned defines the value assgined from Value(Type) to set.
	// if the value type is different from Type, it will be ignored.
	ValueGroupAssigned map[string]any `json:"value_group_assigned"`
}

// ConfigPolicyTemplateType defines the config policy config template type.
type ConfigPolicyTemplateType int

const (
	ConfigPolicyTemplateTypeString       ConfigPolicyTemplateType = 0
	ConfigPolicyTemplateTypeInt          ConfigPolicyTemplateType = 1
	ConfigPolicyTemplateTypeBool         ConfigPolicyTemplateType = 2
	ConfigPolicyTemplateTypeStringSelect ConfigPolicyTemplateType = 3
	ConfigPolicyTemplateTypeIntSelect    ConfigPolicyTemplateType = 4
)
