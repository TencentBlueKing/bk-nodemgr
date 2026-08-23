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

package usermanager

import "fmt"

const (
	lookupFieldLoginName = "login_name"
)

// PermissionActionsRelatedResourceTypesInstance describe the instance of permission actions.
type PermissionActionsRelatedResourceTypesInstance struct {
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

// PermissionActionsRelatedResourceTypes describe the related resource types of permission actions.
type PermissionActionsRelatedResourceTypes struct {
	SystemID   string                                            `json:"system_id"`
	SystemName string                                            `json:"system_name"`
	Type       string                                            `json:"type"`
	TypeName   string                                            `json:"type_name"`
	Instances  [][]PermissionActionsRelatedResourceTypesInstance `json:"instances"`
}

// PermissionActions describe the actions of permission.
type PermissionActions struct {
	ID                   string                                  `json:"id"`
	Name                 string                                  `json:"name"`
	RelatedResourceTypes []PermissionActionsRelatedResourceTypes `json:"related_resource_types"`
}

// Permission describe the permission of user.
type Permission struct {
	SystemID   string              `json:"system_id"`
	SystemName string              `json:"system_name"`
	Actions    []PermissionActions `json:"actions"`
}

// RespCommon describe the common part of response data.
type RespCommon struct {
	Result     *bool       `json:"result"`
	Code       *int        `json:"code"`
	Message    string      `json:"message"`
	Permission *Permission `json:"permission"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data *T `json:"data"`
}

const (
	// CodeOK define the success code.
	CodeOK = 0
)

// IsFailed check the response is ok.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Result == nil && resp.Code == nil {
		return nil
	}

	result := false
	if resp.Result != nil {
		result = *resp.Result
	}

	code := CodeOK
	if resp.Code != nil {
		code = *resp.Code
	}

	switch {
	case result && code == CodeOK:
		return nil
	default:
		return fmt.Errorf("result(%v), code(%d) , msg(%s)", result, code, resp.Message)
	}
}

type listTenantResp = []userManagerTenant

type batchLookupVirtualUserResp = []virtualUser

type virtualUser struct {
	BKUsername  string `json:"bk_username"`
	LoginName   string `json:"login_name"`
	FullName    string `json:"full_name"`
	DisplayName string `json:"display_name"`
}

type tenantStatus string

const (
	tenantStatusEnabled  tenantStatus = "enabled"
	tenantStatusDisabled tenantStatus = "disabled"
)

// Bool convert the tenant status to bool.
func (status tenantStatus) Bool() (bool, error) {
	switch status {
	case tenantStatusEnabled:
		return true, nil
	case tenantStatusDisabled:
		return false, nil
	default:
		return false, fmt.Errorf("invalid tenant status: %s", status)
	}
}

type userManagerTenant struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Status tenantStatus `json:"status"`
}
