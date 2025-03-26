/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides handlers to operate cmd api.
package cmdb

const (
	// ResourceWatchResourceHost describe the resource watch's resource host.
	ResourceWatchResourceHost = "host"

	// ResourceWatchResourceHostRelation describe the resource watch's resource host relation.
	ResourceWatchResourceHostRelation = "host_relation"

	// ResourceWatchResourceProcess describe the resource watch's resource process.
	ResourceWatchResourceProcess = "process"
)

// ResourceWatchEventType represents the event type of watch event.
type ResourceWatchEventType string

const (
	// ResourceWatchEventTypeCreate represents the create event type.
	ResourceWatchEventTypeCreate ResourceWatchEventType = "create"

	// ResourceWatchEventTypeUpdate represents the update event type.
	ResourceWatchEventTypeUpdate ResourceWatchEventType = "update"

	// ResourceWatchEventTypeDelete represents the delete event type.
	ResourceWatchEventTypeDelete ResourceWatchEventType = "delete"
)

// HostEventInfo describe the host event info define by cmdb.
type HostEventInfo struct {
	BKCursor    string    `json:"bk_cursor,omitempty"`
	BKResource  string    `json:"bk_resource"`
	BKEventType string    `json:"bk_event_type,omitempty"`
	BKDetail    *HostInfo `json:"bk_detail"`
}

// HostRelationEventInfo describe the host relation event info define by cmdb.
type HostRelationEventInfo struct {
	BKCursor    string            `json:"bk_cursor,omitempty"`
	BKResource  string            `json:"bk_resource"`
	BKEventType string            `json:"bk_event_type,omitempty"`
	BKDetail    *HostTopoRelation `json:"bk_detail"`
}

// ProcessEventInfo describe the host relation event info define by cmdb.
type ProcessEventInfo struct {
	BKCursor    string           `json:"bk_cursor,omitempty"`
	BKResource  string           `json:"bk_resource"`
	BKEventType string           `json:"bk_event_type,omitempty"`
	BKDetail    *ProcessProperty `json:"bk_detail"`
}
