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

package v4

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/v4/provider"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	iamv4TypeKeySep             = "/"
	iamv4BatchResultPairFmt     = "%s,%s"
	iamv4BatchResultResourceSep = "/"
)

// Preserve the canonical related_resource_types ordering used by the IAM authorizers.
const (
	iamv4OrderBiz         = 0
	iamv4OrderNetworkArea = 1
	iamv4OrderNetworkUnit = 2
	iamv4OrderPackageType = 3
	iamv4OrderPackage     = 4
	iamv4OrderUnknown     = 5
)

// iamv4ResourceTypeOrderKey returns the canonical ordering index for a (systemID, resourceType) pair.
// Unknown pairs are assigned a position after all known pairs.
func iamv4ResourceTypeOrderKey(systemID, typ string) int {
	switch systemID + iamv4TypeKeySep + typ {
	case types.SystemIDCMDB + iamv4TypeKeySep + string(types.AuthResourceTypeBiz):
		return iamv4OrderBiz
	case types.SystemIDNodeMgr + iamv4TypeKeySep + string(types.AuthResourceTypeNetworkArea):
		return iamv4OrderNetworkArea
	case types.SystemIDNodeMgr + iamv4TypeKeySep + string(types.AuthResourceTypeNetworkUnit):
		return iamv4OrderNetworkUnit
	case types.SystemIDNodeMgr + iamv4TypeKeySep + string(types.AuthResourceTypePackageType):
		return iamv4OrderPackageType
	case types.SystemIDNodeMgr + iamv4TypeKeySep + string(types.AuthResourceTypePackage):
		return iamv4OrderPackage
	default:
		return iamv4OrderUnknown
	}
}

// iamv4ResourceKey identifies a unique (systemID, resourceType) pair for grouping.
type iamv4ResourceKey struct {
	systemID string
	resType  string
}

// iamv4GroupResourcesForEnrichment groups resources by (systemID, type) for batch fetching.
// Returns a map from iamv4ResourceKey to slice indices in the original resources slice.
func iamv4GroupResourcesForEnrichment(resources []types.AuthResource) map[iamv4ResourceKey][]int {
	grouped := make(map[iamv4ResourceKey][]int)
	for i, r := range resources {
		// Only enrich bk_nodemgr resources (skip CMDB resources).
		if r.SystemID != types.SystemIDNodeMgr {
			continue
		}
		key := iamv4ResourceKey{systemID: r.SystemID, resType: string(r.Type)}
		grouped[key] = append(grouped[key], i)
	}

	return grouped
}

func iamv4ToIAMResources(resources []types.AuthResource) []types.IAMResource {
	checkResources := make([]types.IAMResource, 0, len(resources))
	for _, r := range resources {
		checkResources = append(checkResources, types.IAMResource{
			SystemID:   r.SystemID,
			Type:       string(r.Type),
			ID:         r.ID,
			Attributes: iamv4BuildCheckAttributes(r.Attributes),
		})
	}

	return checkResources
}

func iamv4BuildCheckAttributes(attributes map[string]interface{}) map[string]interface{} {
	path, ok := iamv4SingleIAMPath(attributes[provider.AttrIAMPath])
	if !ok {
		return attributes
	}

	checkAttributes := make(map[string]interface{}, len(attributes))
	for key, value := range attributes {
		checkAttributes[key] = value
	}
	// Notice: IAM v4 auth-by-resources does not support _bk_iam_path_ arrays yet.
	// Remove this compatibility conversion after IAM supports path arrays.
	checkAttributes[provider.AttrIAMPath] = path

	return checkAttributes
}

func iamv4SingleIAMPath(pathValue interface{}) (string, bool) {
	switch paths := pathValue.(type) {
	case []string:
		if len(paths) != 1 {
			return "", false
		}

		return paths[0], true
	case []interface{}:
		if len(paths) != 1 {
			return "", false
		}
		path, ok := paths[0].(string)
		if !ok {
			return "", false
		}

		return path, true
	default:
		return "", false
	}
}

func iamv4BuildIAMBatchLookupKey(resources []types.IAMResource) string {
	if len(resources) == 0 {
		return ""
	}
	if len(resources) == 1 {
		return resources[0].ID
	}

	nodeIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		nodeIDs = append(nodeIDs, fmt.Sprintf(iamv4BatchResultPairFmt, resource.Type, resource.ID))
	}

	return strings.Join(nodeIDs, iamv4BatchResultResourceSep)
}

func iamv4SortedActions(actionResources map[auth.Action][]types.AuthResource) []auth.Action {
	actions := make([]auth.Action, 0, len(actionResources))
	for action := range actionResources {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(idx, jdx int) bool {
		return actions[idx] < actions[jdx]
	})

	return actions
}

func iamv4BuildIAMApplyResourceTypes(resources []types.AuthResource) []types.IAMApplyResourceType {
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
			// Build complete topology path.
			instance := iamv4BuildResourceInstancePath(r)
			instancesByKey[k] = append(instancesByKey[k], instance)
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
		ki := iamv4ResourceTypeOrderKey(rts[idx].SystemID, rts[idx].Type)
		kj := iamv4ResourceTypeOrderKey(rts[jdx].SystemID, rts[jdx].Type)
		if ki != kj {
			return ki < kj
		}

		return rts[idx].SystemID+iamv4TypeKeySep+rts[idx].Type < rts[jdx].SystemID+iamv4TypeKeySep+rts[jdx].Type
	})

	return rts
}

// iamv4BuildResourceInstancePath constructs the topology path using the optional _bk_iam_path_ parent.
func iamv4BuildResourceInstancePath(r types.AuthResource) types.IAMApplyResourceInstance {
	// Try to extract parent node from _bk_iam_path_ attribute.
	parentNode := provider.ParseParentFromIAMPath(r.Attributes)

	if parentNode == nil {
		// No parent: return single-node path.
		return types.IAMApplyResourceInstance{
			{Type: string(r.Type), ID: r.ID},
		}
	}

	// Has parent: return complete path [parent, current].
	return types.IAMApplyResourceInstance{
		*parentNode,
		{Type: string(r.Type), ID: r.ID},
	}
}

func iamv4BuildRelatedResourceTypes(rts []types.IAMApplyResourceType) []auth.RelatedResourceType {
	return conv.SliceToSlice(rts, func(rt types.IAMApplyResourceType) auth.RelatedResourceType {
		return auth.RelatedResourceType{
			SystemID:   rt.SystemID,
			SystemName: types.SystemDisplayName(rt.SystemID),
			Type:       rt.Type,
			TypeName:   types.AuthResourceTypeDisplayName(types.AuthResourceType(rt.Type)),
			Instances:  iamv4BuildResourceNodes(rt.Instances),
		}
	})
}

func iamv4BuildResourceNodes(instances []types.IAMApplyResourceInstance) []auth.ResourceNode {
	resourceNodes := make([]auth.ResourceNode, 0, len(instances))
	for _, instance := range instances {
		nodes := conv.SliceToSlice(instance, func(node types.IAMApplyResourceNode) auth.ResourceNode {
			return auth.ResourceNode{
				Type:     node.Type,
				TypeName: types.AuthResourceTypeDisplayName(types.AuthResourceType(node.Type)),
				ID:       node.ID,
			}
		})
		resourceNodes = append(resourceNodes, nodes...)
	}

	return resourceNodes
}
