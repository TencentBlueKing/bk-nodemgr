/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// CodeOK defines the success code.
const CodeOK = 0

type backendBaseResp interface {
	GetCode() int32
	GetMessage() string
	GetRequestId() string
}

type backendPermissionResp interface {
	GetPermission() *protoBackend.Permission
}

type backendPermissionError struct {
	message    string
	permission *resterrf.Permission
}

func (err *backendPermissionError) Error() string {
	if err.message != "" {
		return err.message
	}

	return "permission denied"
}

func (err *backendPermissionError) PermissionData() resterrf.Permission {
	if err.permission != nil {
		return *err.permission
	}

	return resterrf.Permission{
		System:     auth.SystemIDNodeMgr,
		SystemName: "",
		ApplyURL:   "",
		Actions:    nil,
	}
}

func extractPermissionError(errInfo any) (resterrf.PermissionError, bool) {
	if errInfo == nil {
		return nil, false
	}

	permErr, ok := errInfo.(resterrf.PermissionError)
	if ok {
		return permErr, true
	}

	err, ok := errInfo.(error)
	if !ok {
		return nil, false
	}

	if errors.As(err, &permErr) {
		return permErr, true
	}

	return nil, false
}

func convertBackendRelatedResourceTypes(resourceTypes []*protoBackend.RelatedResourceType) []resterrf.RelatedResourceType {
	if len(resourceTypes) == 0 {
		return nil
	}

	relatedResourceTypes := make([]resterrf.RelatedResourceType, 0, len(resourceTypes))
	for _, resourceType := range resourceTypes {
		if resourceType == nil {
			continue
		}

		relatedResourceTypes = append(relatedResourceTypes, resterrf.RelatedResourceType{
			SystemID: resourceType.GetSystemId(),
			Type:     resourceType.GetType(),
			TypeName: resourceType.GetTypeName(),
		})
	}

	return relatedResourceTypes
}

func convertBackendActions(actions []*protoBackend.Action) []resterrf.Action {
	if len(actions) == 0 {
		return nil
	}

	convertedActions := make([]resterrf.Action, 0, len(actions))
	for _, action := range actions {
		if action == nil {
			continue
		}

		convertedActions = append(convertedActions, resterrf.Action{
			ID:                   action.GetId(),
			Name:                 action.GetName(),
			RelatedResourceTypes: convertBackendRelatedResourceTypes(action.GetRelatedResourceTypes()),
		})
	}

	return convertedActions
}

func convertBackendPermission(permission *protoBackend.Permission) resterrf.Permission {
	if permission == nil {
		return resterrf.Permission{}
	}

	return resterrf.Permission{
		System:     permission.GetSystem(),
		SystemName: permission.GetSystemName(),
		ApplyURL:   permission.GetApplyUrl(),
		Actions:    convertBackendActions(permission.GetActions()),
	}
}

func extractPermissionFromResp(resp backendBaseResp) (*resterrf.Permission, bool) {
	permissionResp, ok := resp.(backendPermissionResp)
	if !ok {
		return nil, false
	}

	permission := permissionResp.GetPermission()
	if permission == nil {
		return nil, false
	}

	convertedPermission := convertBackendPermission(permission)

	return &convertedPermission, true
}

func buildBackendResponseError(apiName string, resp backendBaseResp, errInfo any) error {
	baseErr := fmt.Errorf("%s failed. code(%d), message(%s), error(%v), request-id(%s)",
		apiName, resp.GetCode(), resp.GetMessage(), errInfo, resp.GetRequestId())

	respPermission, hasRespPermission := extractPermissionFromResp(resp)
	permDataErr, hasPermData := extractPermissionError(errInfo)

	isPermissionDenied := resterrf.Code(resp.GetCode()) == resterrf.PermissionDenied
	if !isPermissionDenied && !hasRespPermission && !hasPermData {
		return baseErr
	}

	permErr := &backendPermissionError{
		message: resp.GetMessage(),
	}
	if hasRespPermission {
		permErr.permission = respPermission
	} else if hasPermData {
		perm := permDataErr.PermissionData()
		permErr.permission = &perm
	}

	return fmt.Errorf("%w: %w", baseErr, permErr)
}

func convertPage(page types.Page) *protoBackend.Page {
	return &protoBackend.Page{
		Offset: int32(page.Offset),
		Limit:  int32(page.Limit),
	}
}
