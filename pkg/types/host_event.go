/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package types ...
package types

import (
	"errors"
	"fmt"
)

// ResourceType represents the type of event.
type ResourceType string

const (
	// ResourceTypeHost represents the host resource type.
	ResourceTypeHost ResourceType = "host"

	// ResourceTypeHostRelation represents the host relation resource type.
	ResourceTypeHostRelation ResourceType = "host_relation"
)

// EventType represents the event type of watch event.
type EventType string

const (
	// EventTypeCreate represents the create event type.
	EventTypeCreate EventType = "create"

	// EventTypeUpdate represents the update event type.
	EventTypeUpdate EventType = "update"

	// EventTypeDelete represents the delete event type.
	EventTypeDelete EventType = "delete"
)

// Validate validates the change type.
func (changeType EventType) Validate() error {
	switch changeType {
	case EventTypeCreate, EventTypeUpdate, EventTypeDelete:
		return nil
	default:
		return fmt.Errorf("invalid change type, type(%s)", changeType)
	}
}

// HostEvent represents the host event structure.
// ResourceType is the type of resource, such as host, host_relation.
// if ResourceType is host_relation, the content in Detail only needs to focus on the HostID and Static.BizID
// if ResourceType is host, the content in Detail needs to focus on the Host except Static.BizID.
type HostEvent struct {
	Cursor    string
	Resource  ResourceType
	EventType EventType
	Detail    *Host
}

// Validate validates the HostEvent structure.
func (he *HostEvent) Validate() error {
	switch he.Resource {
	case ResourceTypeHost, ResourceTypeHostRelation:
	default:
		return errors.New("invalid resource type")
	}

	switch he.EventType {
	case EventTypeCreate, EventTypeUpdate, EventTypeDelete:
	default:
		return errors.New("invalid event type")
	}

	return nil
}

const (
	// HostEventCursor is the cursor for host event.
	HostEventCursor string = "host_event_cursor"

	// HostRelationEventCursor is the cursor for host relation event.
	HostRelationEventCursor string = "host_relation_event_cursor"
)
