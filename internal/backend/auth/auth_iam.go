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
	"sort"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const iamCacheTTL = 5 * time.Minute

// Canonical ordering constants for related_resource_types in IAM apply requests.
// Order must match action registration in support-files/bkiamv3/templates/0003_bk_nodemgr_actions.json.tpl.
const (
	iamOrderBiz         = 0
	iamOrderNetworkArea = 1
	iamOrderNetworkUnit = 2
	iamOrderPackageType = 3
	iamOrderPackage     = 4
	iamOrderUnknown     = 5
)

// iamResourceTypeOrderKey returns the canonical ordering index for a (systemID, resourceType) pair.
// Unknown pairs are assigned a position after all known pairs.
func iamResourceTypeOrderKey(systemID, typ string) int {
	switch systemID + "/" + typ {
	case SystemIDCMDB + "/" + string(ResourceTypeBiz):
		return iamOrderBiz
	case SystemIDNodeMgr + "/" + string(ResourceTypeNetworkArea):
		return iamOrderNetworkArea
	case SystemIDNodeMgr + "/" + string(ResourceTypeNetworkUnit):
		return iamOrderNetworkUnit
	case SystemIDNodeMgr + "/" + string(ResourceTypePackageType):
		return iamOrderPackageType
	case SystemIDNodeMgr + "/" + string(ResourceTypePackage):
		return iamOrderPackage
	default:
		return iamOrderUnknown
	}
}

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

func sortedActions(actionResources map[Action][]Resource) []Action {
	actions := make([]Action, 0, len(actionResources))
	for action := range actionResources {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(idx, jdx int) bool {
		return actions[idx] < actions[jdx]
	})

	return actions
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
	type rtKey struct{ systemID, typ string }
	order := make([]rtKey, 0, len(resources))
	instancesByKey := make(map[rtKey][]types.IAMApplyResourceInstance)

	for _, r := range resources {
		k := rtKey{r.SystemID, string(r.Type)}
		if _, exists := instancesByKey[k]; !exists {
			order = append(order, k)
			instancesByKey[k] = make([]types.IAMApplyResourceInstance, 0)
		}
		if r.ID != "" {
			instancesByKey[k] = append(instancesByKey[k], types.IAMApplyResourceInstance{
				{Type: string(r.Type), ID: r.ID},
			})
		}
	}

	rts := make([]types.IAMApplyResourceType, 0, len(order))
	for _, k := range order {
		rts = append(rts, types.IAMApplyResourceType{
			SystemID:  k.systemID,
			Type:      k.typ,
			Instances: instancesByKey[k],
		})
	}
	sort.Slice(rts, func(idx, jdx int) bool {
		ki := iamResourceTypeOrderKey(rts[idx].SystemID, rts[idx].Type)
		kj := iamResourceTypeOrderKey(rts[jdx].SystemID, rts[jdx].Type)
		if ki != kj {
			return ki < kj
		}

		return rts[idx].SystemID+"/"+rts[idx].Type < rts[jdx].SystemID+"/"+rts[jdx].Type
	})

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

func (authorizer *iamv3Authorizer) collectDeniedResources(
	ctx contextx.IContext, action Action, resources []Resource,
) ([]Resource, bool, error) {

	if ctx == nil {
		return nil, false, fmt.Errorf("auth: Check called with nil context")
	}

	if len(resources) == 0 {
		req := authorizer.newCheckRequest(ctx, action, nil)
		allowed, err := authorizer.handler.IsAllowedWithCache(ctx, req, iamCacheTTL)
		if err != nil {
			return nil, false, err
		}

		return nil, !allowed, nil
	}

	resourcesList := make([][]types.IAMResource, 0, len(resources))
	for _, resource := range resources {
		resourcesList = append(resourcesList, toIAMResources([]Resource{resource}))
	}

	results, err := authorizer.handler.BatchIsAllowed(ctx, authorizer.newCheckRequestWithoutResource(ctx, action), resourcesList)
	if err != nil {
		return nil, false, err
	}

	denied := make([]Resource, 0, len(resources))
	for i, resource := range resources {
		key := buildIAMBatchResultKey(resourcesList[i])
		if allowed, ok := results[key]; ok && allowed {
			continue
		}
		denied = append(denied, resource)
	}

	return denied, len(denied) != 0, nil
}

func (authorizer *iamv3Authorizer) newPermissionDeniedError(
	ctx contextx.IContext, actionResources map[Action][]Resource,
) PermissionDeniedError {

	actions := sortedActions(actionResources)
	applyActions := make([]types.IAMApplyAction, 0, len(actions))
	deniedActions := make([]ActionInfo, 0, len(actions))
	for _, action := range actions {
		rts := buildIAMApplyResourceTypes(actionResources[action])
		applyActions = append(applyActions, types.IAMApplyAction{
			ID:                   string(action),
			RelatedResourceTypes: rts,
		})
		deniedActions = append(deniedActions, ActionInfo{
			ID:                   string(action),
			Name:                 ActionDisplayName(action),
			RelatedResourceTypes: buildRelatedResourceTypes(rts),
		})
	}

	app := types.IAMApplyRequest{
		SystemID: authorizer.systemID,
		Actions:  applyActions,
	}
	applyURL, urlErr := authorizer.handler.GetApplyURL(ctx, app)
	if urlErr != nil {
		applyURL = ""
	}

	return PermissionDeniedError{
		ApplyURL:   applyURL,
		SystemID:   authorizer.systemID,
		SystemName: SystemDisplayName(authorizer.systemID),
		Actions:    deniedActions,
	}
}

func (authorizer *iamv3Authorizer) Check(ctx contextx.IContext, action Action, resources []Resource) error {
	denied, deniedAny, err := authorizer.collectDeniedResources(ctx, action, resources)
	if err != nil {
		return err
	}
	if !deniedAny {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, map[Action][]Resource{
		action: denied,
	})
}

func (authorizer *iamv3Authorizer) CheckMany(
	ctx contextx.IContext, actionResources map[Action][]Resource,
) error {

	deniedActionResources := make(map[Action][]Resource, len(actionResources))
	for _, action := range sortedActions(actionResources) {
		denied, deniedAny, err := authorizer.collectDeniedResources(ctx, action, actionResources[action])
		if err != nil {
			return err
		}
		if deniedAny {
			deniedActionResources[action] = denied
		}
	}
	if len(deniedActionResources) == 0 {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, deniedActionResources)
}
