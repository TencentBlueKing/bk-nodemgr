/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const iamCacheTTL = 5 * time.Minute

type iamv3Authorizer struct {
	systemID string
	handler  iamv3.IHandler
}

// NewIAMV3Authorizer creates an IAuthorizer backed by the IAM v3 handler.
func NewIAMV3Authorizer(systemID string, handler iamv3.IHandler) IAuthorizer {
	return &iamv3Authorizer{systemID: systemID, handler: handler}
}

func toIAMResources(resources []Resource) []types.IAMResource {
	checkResources := make([]types.IAMResource, 0, len(resources))
	for _, r := range resources {
		checkResources = append(checkResources, types.IAMResource{
			SystemID:   r.SystemID,
			Type:       string(r.Type),
			ID:         r.ID,
			Attributes: map[string]interface{}{},
		})
	}

	return checkResources
}

func buildIAMBatchResultKey(resources []types.IAMResource) string {
	if len(resources) == 0 {
		return ""
	}
	if len(resources) == 1 {
		return resources[0].ID
	}

	nodeIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		nodeIDs = append(nodeIDs, fmt.Sprintf("%s,%s", resource.Type, resource.ID))
	}

	return strings.Join(nodeIDs, "/")
}

func (authorizer *iamv3Authorizer) newCheckRequest(ctx contextx.IContext, action Action, resources []Resource) types.IAMCheckRequest {
	return types.IAMCheckRequest{
		SystemID:  authorizer.systemID,
		Username:  ctx.BKUsername(),
		ActionID:  string(action),
		Resources: toIAMResources(resources),
	}
}

func (authorizer *iamv3Authorizer) newCheckRequestWithoutResource(ctx contextx.IContext, action Action) types.IAMCheckRequest {
	return types.IAMCheckRequest{
		SystemID: authorizer.systemID,
		Username: ctx.BKUsername(),
		ActionID: string(action),
	}
}

func buildIAMApplyResourceTypes(resources []Resource) []types.IAMApplyResourceType {
	rts := make([]types.IAMApplyResourceType, 0, len(resources))
	seen := map[string]bool{}
	for _, r := range resources {
		key := r.SystemID + "/" + string(r.Type)
		if seen[key] {
			continue
		}
		seen[key] = true
		rts = append(rts, types.IAMApplyResourceType{
			SystemID: r.SystemID,
			Type:     string(r.Type),
		})
	}

	return rts
}

func buildRelatedResourceTypes(rts []types.IAMApplyResourceType) []RelatedResourceType {
	relatedRTs := make([]RelatedResourceType, 0, len(rts))
	for _, rt := range rts {
		relatedRTs = append(relatedRTs, RelatedResourceType{
			SystemID: rt.SystemID,
			Type:     rt.Type,
			TypeName: ResourceTypeDisplayName(ResourceType(rt.Type)),
		})
	}

	return relatedRTs
}

func (authorizer *iamv3Authorizer) newPermissionDeniedError(
	ctx contextx.IContext, action Action, resources []Resource,
) PermissionDeniedError {

	rts := buildIAMApplyResourceTypes(resources)
	app := types.IAMApplyRequest{
		SystemID: authorizer.systemID,
		Actions: []types.IAMApplyAction{
			{ID: string(action), RelatedResourceTypes: rts},
		},
	}
	applyURL, urlErr := authorizer.handler.GetApplyURL(ctx, app)
	if urlErr != nil {
		applyURL = ""
	}

	return PermissionDeniedError{
		ApplyURL:   applyURL,
		SystemID:   authorizer.systemID,
		SystemName: SystemDisplayName(authorizer.systemID),
		Actions: []ActionInfo{{
			ID:                   string(action),
			Name:                 ActionDisplayName(action),
			RelatedResourceTypes: buildRelatedResourceTypes(rts),
		}},
	}
}

func (authorizer *iamv3Authorizer) Check(ctx contextx.IContext, action Action, resources []Resource) error {
	if ctx == nil {
		return fmt.Errorf("auth: Check called with nil context")
	}
	req := authorizer.newCheckRequest(ctx, action, resources)
	allowed, err := authorizer.handler.IsAllowedWithCache(ctx, req, iamCacheTTL)
	if err != nil {
		return err
	}
	if allowed {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, action, resources)
}

func (authorizer *iamv3Authorizer) BatchCheck(ctx contextx.IContext, action Action, resources []Resource) error {
	if len(resources) == 0 {
		return nil
	}
	if ctx == nil {
		return fmt.Errorf("auth: Check called with nil context")
	}

	resourcesList := make([][]types.IAMResource, 0, len(resources))
	for _, resource := range resources {
		resourcesList = append(resourcesList, toIAMResources([]Resource{resource}))
	}

	results, err := authorizer.handler.BatchIsAllowed(ctx, authorizer.newCheckRequestWithoutResource(ctx, action), resourcesList)
	if err != nil {
		return err
	}

	denied := make([]Resource, 0, len(resources))
	for i, resource := range resources {
		key := buildIAMBatchResultKey(resourcesList[i])
		if allowed, ok := results[key]; ok && allowed {
			continue
		}
		denied = append(denied, resource)
	}
	if len(denied) == 0 {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, action, denied)
}
