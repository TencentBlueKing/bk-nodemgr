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

package auth

import (
	"fmt"
	"slices"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// PermissionDeniedError indicates the user has no permission for the requested actions.
type PermissionDeniedError struct {
	ApplyURL   string
	SystemID   string
	SystemName string
	Actions    []ActionInfo
}

// Error implements the error interface.
func (permErr PermissionDeniedError) Error() string {
	if permErr.ApplyURL != "" {
		return fmt.Sprintf("permission denied, apply at: %s", permErr.ApplyURL)
	}

	return "permission denied"
}

// ActionInfo represents action metadata used for permission-apply responses.
type ActionInfo struct {
	ID                   string
	Name                 string
	RelatedResourceTypes []RelatedResourceType
}

// ResourceNode represents one node on a denied resource instance path.
type ResourceNode struct {
	Type     string
	TypeName string
	ID       string
	Name     string
}

// RelatedResourceType represents an IAM related resource type.
type RelatedResourceType struct {
	SystemID   string
	SystemName string
	Type       string
	TypeName   string
	Instances  []ResourceNode
}

// FillResourceNames enriches denied instances without changing the permission decision.
// Missing instances keep their existing names so clients can fall back to their IDs.
func (permErr PermissionDeniedError) FillResourceNames(
	ctx contextx.IContext,
	fetch func(contextx.IContext, string, string, []string) (map[string]string, error),
) {
	// IAM V4 fetch_instance_info accepts at most 1000 IDs per request.
	const batchSize = 1000
	for key, nodesByID := range permErr.unnamedResourceNodes() {
		ids := make([]string, 0, len(nodesByID))
		for id := range nodesByID {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for batch := range slices.Chunk(ids, batchSize) {
			names, err := fetch(ctx, key.systemID, key.resourceType, batch)
			if err != nil {
				logger.G.Biz(ctx).WithErr(err).With("resource-type", key.resourceType).
					Error("failed to fetch denied resource names")

				continue
			}
			for id, name := range names {
				for _, node := range nodesByID[id] {
					node.Name = name
				}
			}
		}
	}
}

type resourceTypeKey struct{ systemID, resourceType string }

func (permErr PermissionDeniedError) unnamedResourceNodes() map[resourceTypeKey]map[string][]*ResourceNode {
	grouped := make(map[resourceTypeKey]map[string][]*ResourceNode)
	for _, action := range permErr.Actions {
		for _, rt := range action.RelatedResourceTypes {
			for index := range rt.Instances {
				node := &rt.Instances[index]
				if node.ID == "" || node.Name != "" {
					continue
				}
				// Parent nodes belong to their own resource type, not the leaf type.
				key := resourceTypeKey{rt.SystemID, node.Type}
				if grouped[key] == nil {
					grouped[key] = make(map[string][]*ResourceNode)
				}
				grouped[key][node.ID] = append(grouped[key][node.ID], node)
			}
		}
	}

	return grouped
}

// PermissionData converts PermissionDeniedError into a restserver.Permission payload.
// This satisfies the restserver.PermissionProvider interface.
func (permErr PermissionDeniedError) PermissionData() resterrf.Permission {
	actions := make([]resterrf.Action, 0, len(permErr.Actions))
	for _, act := range permErr.Actions {
		rts := make([]resterrf.RelatedResourceType, 0, len(act.RelatedResourceTypes))
		for _, rt := range act.RelatedResourceTypes {
			instances := make([]resterrf.ResourceNode, 0, len(rt.Instances))
			for _, node := range rt.Instances {
				instances = append(instances, resterrf.ResourceNode{
					Type:     node.Type,
					TypeName: node.TypeName,
					ID:       node.ID,
					Name:     node.Name,
				})
			}
			rts = append(rts, resterrf.RelatedResourceType{
				SystemID:   rt.SystemID,
				SystemName: rt.SystemName,
				Type:       rt.Type,
				TypeName:   rt.TypeName,
				Instances:  instances,
			})
		}
		actions = append(actions, resterrf.Action{
			ID:                   act.ID,
			Name:                 act.Name,
			RelatedResourceTypes: rts,
		})
	}

	return resterrf.Permission{
		System:     permErr.SystemID,
		SystemName: permErr.SystemName,
		ApplyURL:   permErr.ApplyURL,
		Actions:    actions,
	}
}
