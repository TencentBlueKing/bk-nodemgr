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
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/provider"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3/policy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const iamCacheTTL = 5 * time.Minute

const (
	iamTypeKeySep             = "/"
	iamBatchResultPairFmt     = "%s,%s"
	iamBatchResultResourceSep = "/"
)

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
	switch systemID + iamTypeKeySep + typ {
	case types.SystemIDCMDB + iamTypeKeySep + string(types.AuthResourceTypeBiz):
		return iamOrderBiz
	case types.SystemIDNodeMgr + iamTypeKeySep + string(types.AuthResourceTypeNetworkArea):
		return iamOrderNetworkArea
	case types.SystemIDNodeMgr + iamTypeKeySep + string(types.AuthResourceTypeNetworkUnit):
		return iamOrderNetworkUnit
	case types.SystemIDNodeMgr + iamTypeKeySep + string(types.AuthResourceTypePackageType):
		return iamOrderPackageType
	case types.SystemIDNodeMgr + iamTypeKeySep + string(types.AuthResourceTypePackage):
		return iamOrderPackage
	default:
		return iamOrderUnknown
	}
}

type iamv3Authorizer struct {
	systemID          string
	handler           iamv3.IHandler
	attributeEnricher provider.IAttributeEnricher
	resolver          provider.IResolver
}

// NewIAMV3Authorizer creates an IAuthorizer backed by the IAM v3 handler.
// If attributeEnricher is provided, it will automatically enrich resource attributes during authorization.
// If resolver is provided, it will be used as fallback when discrete policy parsing fails.
func NewIAMV3Authorizer(
	systemID string,
	handler iamv3.IHandler,
	attributeEnricher provider.IAttributeEnricher,
	resolver provider.IResolver,
) IAuthorizer {

	return &iamv3Authorizer{
		systemID:          systemID,
		handler:           handler,
		attributeEnricher: attributeEnricher,
		resolver:          resolver,
	}
}

// enrichResourceAttributes enriches resources with attributes from provider.
// It groups resources by (systemID, type) and batch-fetches attributes for each group.
func (authorizer *iamv3Authorizer) enrichResourceAttributes(ctx contextx.IContext, resources []types.AuthResource) []types.AuthResource {
	// If no enricher configured, return resources as-is
	if authorizer.attributeEnricher == nil {
		return resources
	}

	// Group resources by (systemID, type) for batch fetching
	grouped := groupResourcesForEnrichment(resources)

	// Fetch and merge attributes for each group
	for key, indices := range grouped {
		authorizer.fetchAndMergeAttributes(ctx, key.resType, indices, resources)
	}

	return resources
}

// resourceKey identifies a unique (systemID, resourceType) pair for grouping.
type resourceKey struct {
	systemID string
	resType  string
}

// groupResourcesForEnrichment groups resources by (systemID, type) for batch fetching.
// Returns a map from resourceKey to slice indices in the original resources slice.
func groupResourcesForEnrichment(resources []types.AuthResource) map[resourceKey][]int {
	grouped := make(map[resourceKey][]int)
	for i, r := range resources {
		// Only enrich bk_nodemgr resources (skip CMDB resources)
		if r.SystemID != types.SystemIDNodeMgr {
			continue
		}
		key := resourceKey{systemID: r.SystemID, resType: string(r.Type)}
		grouped[key] = append(grouped[key], i)
	}

	return grouped
}

// fetchAndMergeAttributes fetches attributes for a group of resources and merges them back.
func (authorizer *iamv3Authorizer) fetchAndMergeAttributes(
	ctx contextx.IContext,
	resType string,
	indices []int,
	resources []types.AuthResource,
) {
	// Collect resource IDs
	ids := make([]string, 0, len(indices))
	for _, idx := range indices {
		ids = append(ids, resources[idx].ID)
	}

	// Fetch attributes from provider
	attrsMap, err := authorizer.attributeEnricher.FetchResourceAttributes(ctx, resType, ids)
	if err != nil {
		// Log error but don't fail authorization - continue without attributes
		return
	}

	// Merge fetched attributes into resources
	for _, idx := range indices {
		resID := resources[idx].ID
		attrs, ok := attrsMap[resID]
		if !ok || len(attrs) == 0 {
			continue
		}

		// Initialize Attributes map if nil
		if resources[idx].Attributes == nil {
			resources[idx].Attributes = attrs
			continue
		}

		// Merge attributes into existing map
		for k, v := range attrs {
			resources[idx].Attributes[k] = v
		}
	}
}
func toIAMResources(resources []types.AuthResource) []types.IAMResource {
	checkResources := make([]types.IAMResource, 0, len(resources))
	for _, r := range resources {
		checkResources = append(checkResources, types.IAMResource{
			SystemID:   r.SystemID,
			Type:       string(r.Type),
			ID:         r.ID,
			Attributes: r.Attributes,
		})
	}

	return checkResources
}

func toAuthorizedScope(isAny bool, iamResources []types.IAMResource) AuthorizedScope {
	resources := make([]types.AuthResource, 0, len(iamResources))
	for _, r := range iamResources {
		resources = append(resources, types.AuthResource{
			SystemID: r.SystemID,
			Type:     types.AuthResourceType(r.Type),
			ID:       r.ID,
		})
	}

	return AuthorizedScope{IsAny: isAny, Resources: resources}
}

func buildIAMBatchLookupKey(resources []types.IAMResource) string {
	if len(resources) == 0 {
		return ""
	}
	if len(resources) == 1 {
		return resources[0].ID
	}

	nodeIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		nodeIDs = append(nodeIDs, fmt.Sprintf(iamBatchResultPairFmt, resource.Type, resource.ID))
	}

	return strings.Join(nodeIDs, iamBatchResultResourceSep)
}

func sortedActions(actionResources map[Action][]types.AuthResource) []Action {
	actions := make([]Action, 0, len(actionResources))
	for action := range actionResources {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(idx, jdx int) bool {
		return actions[idx] < actions[jdx]
	})

	return actions
}

func (authorizer *iamv3Authorizer) newCheckRequest(ctx contextx.IContext, action Action, resources []types.AuthResource) types.IAMCheckRequest {
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

func buildIAMApplyResourceTypes(resources []types.AuthResource) []types.IAMApplyResourceType {
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

		return rts[idx].SystemID+iamTypeKeySep+rts[idx].Type < rts[jdx].SystemID+iamTypeKeySep+rts[jdx].Type
	})

	return rts
}

func buildRelatedResourceTypes(rts []types.IAMApplyResourceType) []RelatedResourceType {
	return conv.SliceToSlice(rts, func(rt types.IAMApplyResourceType) RelatedResourceType {
		return RelatedResourceType{
			SystemID:   rt.SystemID,
			SystemName: types.SystemDisplayName(rt.SystemID),
			Type:       rt.Type,
			TypeName:   types.AuthResourceTypeDisplayName(types.AuthResourceType(rt.Type)),
			Instances:  buildResourceNodes(rt.Instances),
		}
	})
}

func buildResourceNodes(instances []types.IAMApplyResourceInstance) []ResourceNode {
	resourceNodes := make([]ResourceNode, 0, len(instances))
	for _, instance := range instances {
		nodes := conv.SliceToSlice(instance, func(node types.IAMApplyResourceNode) ResourceNode {
			return ResourceNode{
				Type:     node.Type,
				TypeName: types.AuthResourceTypeDisplayName(types.AuthResourceType(node.Type)),
				ID:       node.ID,
			}
		})
		resourceNodes = append(resourceNodes, nodes...)
	}

	return resourceNodes
}

func (authorizer *iamv3Authorizer) collectDeniedResources(
	ctx contextx.IContext, action Action, resources []types.AuthResource,
) ([]types.AuthResource, bool, error) {

	if ctx == nil {
		return nil, false, errors.New("auth: Check called with nil context")
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
		resourcesList = append(resourcesList, toIAMResources([]types.AuthResource{resource}))
	}

	results, err := authorizer.handler.BatchIsAllowed(ctx, authorizer.newCheckRequestWithoutResource(ctx, action), resourcesList)
	if err != nil {
		return nil, false, err
	}

	denied := make([]types.AuthResource, 0, len(resources))
	for i, resource := range resources {
		key := buildIAMBatchLookupKey(resourcesList[i])
		if allowed, ok := results[key]; ok && allowed {
			continue
		}
		denied = append(denied, resource)
	}

	return denied, len(denied) != 0, nil
}

func (authorizer *iamv3Authorizer) newPermissionDeniedError(
	ctx contextx.IContext, actionResources map[Action][]types.AuthResource,
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
		SystemName: types.SystemDisplayName(authorizer.systemID),
		Actions:    deniedActions,
	}
}

func (authorizer *iamv3Authorizer) Check(ctx contextx.IContext, action Action, resources []types.AuthResource) error {
	// Enrich resources with attributes from provider
	enrichedResources := authorizer.enrichResourceAttributes(ctx, resources)

	denied, deniedAny, err := authorizer.collectDeniedResources(ctx, action, enrichedResources)
	if err != nil {
		return err
	}
	if !deniedAny {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, map[Action][]types.AuthResource{
		action: denied,
	})
}

func (authorizer *iamv3Authorizer) CheckMany(
	ctx contextx.IContext, actionResources map[Action][]types.AuthResource,
) error {
	// Enrich all resources first
	enrichedActionResources := make(map[Action][]types.AuthResource, len(actionResources))
	for action, resources := range actionResources {
		enrichedActionResources[action] = authorizer.enrichResourceAttributes(ctx, resources)
	}

	deniedActionResources := make(map[Action][]types.AuthResource, len(enrichedActionResources))
	for _, action := range sortedActions(enrichedActionResources) {
		denied, deniedAny, err := authorizer.collectDeniedResources(ctx, action, enrichedActionResources[action])
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

func (authorizer *iamv3Authorizer) ListAuthorizedInstances(
	ctx contextx.IContext, action Action, resourceType types.AuthResourceType,
) (AuthorizedScope, error) {

	// Get raw policy expression from IAM
	policyExpr, err := authorizer.handler.GetPolicyExpression(ctx, types.IAMAuthorizedInstancesRequest{
		SystemID:     authorizer.systemID,
		Username:     ctx.BKUsername(),
		ActionID:     string(action),
		ResourceType: string(resourceType),
	})
	if err != nil {
		return AuthorizedScope{}, err
	}

	// Try fast path: parse discrete policy
	isAny, iamResources, parseErr := policy.Parse(policyExpr, authorizer.systemID, string(resourceType))
	if parseErr == nil {
		// Fast path succeeded
		return toAuthorizedScope(isAny, iamResources), nil
	}

	// Only topology resources (networkunit, package) support non-discrete policy fallback
	switch resourceType {
	case types.AuthResourceTypeNetworkUnit, types.AuthResourceTypePackage:
		// These resource types support fallback to expression evaluation
	default:
		return AuthorizedScope{}, fmt.Errorf(
			"resource type %s does not support non-discrete policy: %w",
			resourceType,
			parseErr,
		)
	}

	// Call provider to evaluate expression against all instances
	result, err := authorizer.resolver.ListInstancesByExpression(
		ctx,
		string(resourceType),
		policyExpr,
		types.Page{Limit: provider.MaxListInstanceByPolicyLimit},
	)
	if err != nil {
		return AuthorizedScope{}, fmt.Errorf("failed to evaluate non-discrete policy: %w", err)
	}

	// Convert provider results to AuthorizedScope
	resources := make([]types.AuthResource, 0, len(result.Results))
	for _, inst := range result.Results {
		resources = append(resources, types.AuthResource{
			SystemID: authorizer.systemID,
			Type:     resourceType,
			ID:       inst.ID,
		})
	}

	return AuthorizedScope{
		IsAny:     false,
		Resources: resources,
	}, nil
}
