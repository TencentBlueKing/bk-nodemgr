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

// Package v4 implements IAM V4 authorization and resource provider wiring.
package v4

import (
	"errors"
	"fmt"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v4/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv4"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	iamv4BatchLimit      = 20
	iamv4AuthorizedAnyID = "*"
)

type iamv4Authorizer struct {
	systemID          string
	handler           iamv4.IHandler
	attributeEnricher provider.IAttributeEnricher
	instanceLister    provider.IInstanceLister
}

// NewProviderHandler registers the IAM V4 resource providers.
func NewProviderHandler(topoStorage topo.IStorage, releaseStorage file.IPkgReleaseHandler) provider.IHandler {
	handler := provider.NewHandler()
	handler.RegisterProvider(provider.ResourceTypeBiz, provider.NewBizProvider(topoStorage))
	handler.RegisterProvider(provider.ResourceTypeNetworkArea, provider.NewNetworkAreaProvider(topoStorage))
	handler.RegisterProvider(provider.ResourceTypeNetworkUnit, provider.NewNetworkUnitProvider(topoStorage))
	handler.RegisterProvider(provider.ResourceTypePackageType, provider.NewPackageTypeProvider())
	handler.RegisterProvider(provider.ResourceTypePackage, provider.NewPackageProvider(releaseStorage))

	return handler
}

// NewIAMV4Authorizer creates an IAuthorizer backed by the IAM v4 handler.
func NewIAMV4Authorizer(
	systemID string,
	handler iamv4.IHandler,
	attributeEnricher provider.IAttributeEnricher,
	instanceLister provider.IInstanceLister,
) auth.IAuthorizer {

	return &iamv4Authorizer{
		systemID:          systemID,
		handler:           handler,
		attributeEnricher: attributeEnricher,
		instanceLister:    instanceLister,
	}
}

func (authorizer *iamv4Authorizer) enrichResourceAttributes(ctx contextx.IContext, resources []types.AuthResource) []types.AuthResource {
	if authorizer.attributeEnricher == nil {
		return resources
	}

	grouped := iamv4GroupResourcesForEnrichment(resources)
	for key, indices := range grouped {
		authorizer.fetchAndMergeAttributes(ctx, key.resType, indices, resources)
	}

	return resources
}

func (authorizer *iamv4Authorizer) fetchAndMergeAttributes(
	ctx contextx.IContext,
	resType string,
	indices []int,
	resources []types.AuthResource,
) {

	ids := make([]string, 0, len(indices))
	for _, idx := range indices {
		ids = append(ids, resources[idx].ID)
	}

	attrsMap, err := authorizer.attributeEnricher.FetchResourceAttributes(ctx, resType, ids)
	if err != nil {
		return
	}

	for _, idx := range indices {
		resID := resources[idx].ID
		attrs, ok := attrsMap[resID]
		if !ok || len(attrs) == 0 {
			continue
		}

		if resources[idx].Attributes == nil {
			resources[idx].Attributes = attrs
			continue
		}

		for k, v := range attrs {
			resources[idx].Attributes[k] = v
		}
	}
}

func (authorizer *iamv4Authorizer) newCheckRequest(ctx contextx.IContext, action auth.Action, resources []types.AuthResource) types.IAMCheckRequest {
	return types.IAMCheckRequest{
		SystemID:  authorizer.systemID,
		Username:  ctx.BKUsername(),
		ActionID:  string(action),
		Resources: iamv4ToIAMResources(resources),
	}
}

func (authorizer *iamv4Authorizer) newMultiActionCheckRequest(
	ctx contextx.IContext,
	actions []auth.Action,
	resources []types.AuthResource,
) types.IAMMultiActionCheckRequest {

	actionIDs := make([]string, 0, len(actions))
	for _, action := range actions {
		actionIDs = append(actionIDs, string(action))
	}

	return types.IAMMultiActionCheckRequest{
		SystemID:  authorizer.systemID,
		Username:  ctx.BKUsername(),
		ActionIDs: actionIDs,
		Resources: iamv4ToIAMResources(resources),
	}
}

func (authorizer *iamv4Authorizer) collectDeniedResources(
	ctx contextx.IContext, action auth.Action, resources []types.AuthResource,
) ([]types.AuthResource, bool, error) {

	if ctx == nil {
		return nil, false, errors.New("auth: Check called with nil context")
	}

	if len(resources) == 0 {
		allowed, err := authorizer.handler.IsAllowed(ctx, authorizer.newCheckRequest(ctx, action, nil))
		if err != nil {
			return nil, false, err
		}

		return nil, !allowed, nil
	}

	denied := make([]types.AuthResource, 0, len(resources))
	for start := 0; start < len(resources); start += iamv4BatchLimit {
		end := start + iamv4BatchLimit
		if end > len(resources) {
			end = len(resources)
		}

		batch := resources[start:end]
		allowedByResource, err := authorizer.handler.ResourcesAllowed(ctx, authorizer.newCheckRequest(ctx, action, batch))
		if err != nil {
			return nil, false, err
		}

		for _, resource := range batch {
			if allowed, ok := allowedByResource[resource.ID]; ok && allowed {
				continue
			}
			denied = append(denied, resource)
		}
	}

	return denied, len(denied) != 0, nil
}

func (authorizer *iamv4Authorizer) newPermissionDeniedError(
	ctx contextx.IContext, actionResources map[auth.Action][]types.AuthResource,
) auth.PermissionDeniedError {

	actions := iamv4SortedActions(actionResources)
	applyActions := make([]types.IAMApplyAction, 0, len(actions))
	deniedActions := make([]auth.ActionInfo, 0, len(actions))
	for _, action := range actions {
		rts := iamv4BuildIAMApplyResourceTypes(actionResources[action])
		applyActions = append(applyActions, types.IAMApplyAction{
			ID:                   string(action),
			RelatedResourceTypes: rts,
		})
		deniedActions = append(deniedActions, auth.ActionInfo{
			ID:                   string(action),
			Name:                 auth.ActionDisplayName(action),
			RelatedResourceTypes: iamv4BuildRelatedResourceTypes(rts),
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

	return auth.PermissionDeniedError{
		ApplyURL:   applyURL,
		SystemID:   authorizer.systemID,
		SystemName: types.SystemDisplayName(authorizer.systemID),
		Actions:    deniedActions,
	}
}

func (authorizer *iamv4Authorizer) collectDeniedActionResourcesByActions(
	ctx contextx.IContext,
	actionResources map[auth.Action][]types.AuthResource,
) (map[auth.Action][]types.AuthResource, error) {

	deniedActionResources := make(map[auth.Action][]types.AuthResource, len(actionResources))
	resourcesByKey, actionsByKey := iamv4GroupActionsByResource(actionResources)

	keys := make([]string, 0, len(actionsByKey))
	for key := range actionsByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		resources := resourcesByKey[key]
		actions := actionsByKey[key]
		for start := 0; start < len(actions); start += iamv4BatchLimit {
			end := start + iamv4BatchLimit
			if end > len(actions) {
				end = len(actions)
			}

			batchActions := actions[start:end]
			allowedByAction, err := authorizer.handler.ActionsAllowed(
				ctx,
				authorizer.newMultiActionCheckRequest(ctx, batchActions, resources),
			)
			if err != nil {
				return nil, err
			}

			for _, action := range batchActions {
				if allowed, ok := allowedByAction[string(action)]; ok && allowed {
					continue
				}
				deniedActionResources[action] = append(deniedActionResources[action], resources...)
			}
		}
	}

	return deniedActionResources, nil
}

func iamv4GroupActionsByResource(
	actionResources map[auth.Action][]types.AuthResource,
) (map[string][]types.AuthResource, map[string][]auth.Action) {

	resourcesByKey := make(map[string][]types.AuthResource)
	actionsByKey := make(map[string][]auth.Action)

	for _, action := range iamv4SortedActions(actionResources) {
		resources := actionResources[action]
		if len(resources) == 0 {
			resourcesByKey[""] = nil
			actionsByKey[""] = append(actionsByKey[""], action)

			continue
		}

		for _, resource := range resources {
			key := iamv4BuildIAMBatchLookupKey(iamv4ToIAMResources([]types.AuthResource{resource}))
			resourcesByKey[key] = []types.AuthResource{resource}
			actionsByKey[key] = append(actionsByKey[key], action)
		}
	}

	return resourcesByKey, actionsByKey
}

func (authorizer *iamv4Authorizer) Check(ctx contextx.IContext, action auth.Action, resources []types.AuthResource) error {
	enrichedResources := authorizer.enrichResourceAttributes(ctx, resources)

	denied, deniedAny, err := authorizer.collectDeniedResources(ctx, action, enrichedResources)
	if err != nil {
		return err
	}
	if !deniedAny {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, map[auth.Action][]types.AuthResource{
		action: denied,
	})
}

func (authorizer *iamv4Authorizer) CheckMany(
	ctx contextx.IContext,
	actionResources map[auth.Action][]types.AuthResource,
) error {

	enrichedActionResources := make(map[auth.Action][]types.AuthResource, len(actionResources))
	for action, resources := range actionResources {
		enrichedActionResources[action] = authorizer.enrichResourceAttributes(ctx, resources)
	}

	deniedActionResources, err := authorizer.collectDeniedActionResourcesByActions(ctx, enrichedActionResources)
	if err != nil {
		return err
	}
	if len(deniedActionResources) == 0 {
		return nil
	}

	return authorizer.newPermissionDeniedError(ctx, deniedActionResources)
}

func (authorizer *iamv4Authorizer) ListAuthorizedInstances(
	ctx contextx.IContext, action auth.Action, resourceType types.AuthResourceType,
) (auth.AuthorizedScope, error) {

	if ctx == nil {
		return auth.AuthorizedScope{}, errors.New("auth: ListAuthorizedInstances called with nil context")
	}

	results, err := authorizer.handler.ListAuthorizedResources(ctx, types.IAMAuthorizedInstancesRequest{
		SystemID:     authorizer.systemID,
		Username:     ctx.BKUsername(),
		ActionID:     string(action),
		ResourceType: string(resourceType),
	})
	if err != nil {
		return auth.AuthorizedScope{}, err
	}

	resources, isAny, err := authorizer.resolveAuthorizedResources(ctx, resourceType, results)
	if err != nil {
		return auth.AuthorizedScope{}, err
	}

	return auth.AuthorizedScope{IsAny: isAny, Resources: resources}, nil
}

func (authorizer *iamv4Authorizer) resolveAuthorizedResources(
	ctx contextx.IContext,
	resourceType types.AuthResourceType,
	results []iamv4.AuthorizedResourceResponse,
) ([]types.AuthResource, bool, error) {

	resources := make([]types.AuthResource, 0)
	seen := make(map[string]struct{})

	for _, result := range results {
		resultType := types.AuthResourceType(result.Type)
		if resultType == resourceType && hasAnyAuthorizedID(result.IDs) {
			return []types.AuthResource{}, true, nil
		}
		if resultType == resourceType {
			appendAuthorizedResources(&resources, seen, resourceType, result.IDs)
			continue
		}

		expanded, err := authorizer.expandAuthorizedResources(ctx, resultType, resourceType, result.IDs)
		if err != nil {
			return nil, false, err
		}
		appendAuthorizedResources(&resources, seen, resourceType, expanded)
	}

	return resources, false, nil
}

func appendAuthorizedResources(
	resources *[]types.AuthResource,
	seen map[string]struct{},
	resourceType types.AuthResourceType,
	ids []string,
) {

	for _, id := range ids {
		if id == "" || id == iamv4AuthorizedAnyID {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		*resources = append(*resources, types.AuthResource{
			SystemID: types.AuthResourceTypeToSystemID(resourceType),
			Type:     resourceType,
			ID:       id,
		})
	}
}

func hasAnyAuthorizedID(ids []string) bool {
	for _, id := range ids {
		if id == iamv4AuthorizedAnyID {
			return true
		}
	}

	return false
}

func (authorizer *iamv4Authorizer) expandAuthorizedResources(
	ctx contextx.IContext,
	parentType types.AuthResourceType,
	targetType types.AuthResourceType,
	parentIDs []string,
) ([]string, error) {

	if !canExpandAuthorizedParent(parentType, targetType) {
		return []string{}, nil
	}

	ids := make([]string, 0)
	for _, parentID := range parentIDs {
		if parentID == "" {
			continue
		}
		if parentID == iamv4AuthorizedAnyID {
			return authorizer.listAllResourceIDs(ctx, targetType, nil)
		}

		parent := &provider.ParentFilter{
			Type: string(parentType),
			ID:   parentID,
		}
		expanded, err := authorizer.listAllResourceIDs(ctx, targetType, parent)
		if err != nil {
			return nil, fmt.Errorf("failed to expand %s %s to %s: %w", parentType, parentID, targetType, err)
		}
		ids = append(ids, expanded...)
	}

	return conv.SliceUnique(ids), nil
}

func canExpandAuthorizedParent(parentType, targetType types.AuthResourceType) bool {
	switch targetType {
	case types.AuthResourceTypeNetworkUnit:
		return parentType == types.AuthResourceTypeNetworkArea
	case types.AuthResourceTypePackage:
		return parentType == types.AuthResourceTypePackageType
	default:
		return false
	}
}

func (authorizer *iamv4Authorizer) listAllResourceIDs(
	ctx contextx.IContext,
	resourceType types.AuthResourceType,
	parent *provider.ParentFilter,
) ([]string, error) {

	if authorizer.instanceLister == nil {
		return nil, fmt.Errorf("auth provider instance lister is nil")
	}

	ids := make([]string, 0)
	page := types.Page{Limit: provider.MaxListInstancePageSize}
	var expectedCount int64
	seen := make(map[string]struct{})
	for {
		data, err := authorizer.instanceLister.ListInstance(ctx, string(resourceType), &provider.Request[provider.ListInstanceFilter]{
			Filter: provider.ListInstanceFilter{Parent: parent}, Page: page,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to enumerate resource instances: %w", err)
		}
		if data == nil || data.Count < 0 {
			return nil, fmt.Errorf("invalid list_instance result")
		}
		if page.Offset == 0 {
			expectedCount = data.Count
		}
		remaining := expectedCount - int64(page.Offset)
		if data.Count != expectedCount || int64(len(data.Results)) != min(int64(page.Limit), remaining) {
			return nil, fmt.Errorf("incomplete or changing list_instance pages")
		}

		for _, instance := range data.Results {
			if _, duplicate := seen[instance.ID]; duplicate || instance.ID == "" {
				return nil, fmt.Errorf("invalid or duplicate list_instance ID")
			}
			seen[instance.ID] = struct{}{}
			ids = append(ids, instance.ID)
		}

		page.Offset += len(data.Results)
		if int64(page.Offset) == expectedCount {
			break
		}
	}

	return conv.SliceUnique(ids), nil
}
