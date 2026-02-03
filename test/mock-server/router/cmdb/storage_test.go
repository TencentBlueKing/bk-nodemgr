/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides the CMDB mock API storage.
package cmdb

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/stretchr/testify/assert"
)

// newTestStorage creates a storage instance with test data.
func newTestStorage() *storage {
	conf := &MockData{
		Businesses: []BusinessConfig{
			{BKBizID: 1, BKBizName: "test-biz-1"},
			{BKBizID: 2, BKBizName: "test-biz-2"},
			{BKBizID: 3, BKBizName: "test-biz-3"},
		},
		Areas: []CloudAreaConfig{
			{BKCloudID: 0, BKCloudName: "default-area"},
			{BKCloudID: 1, BKCloudName: "area-1"},
			{BKCloudID: 2, BKCloudName: "area-2"},
		},
		Hosts: []HostConfig{
			{BKHostID: 100, BKBizID: 1, BKCloudID: 0, BKInnerIP: "127.0.0.1", BKOSType: "linux"},
			{BKHostID: 101, BKBizID: 1, BKCloudID: 1, BKInnerIP: "127.0.0.2", BKOSType: "linux"},
			{BKHostID: 102, BKBizID: 2, BKCloudID: 0, BKInnerIP: "127.0.0.3", BKOSType: "windows"},
		},
	}
	return newStorage(conf)
}

// TestStorage_SearchBusiness tests SearchBusiness.
func TestStorage_SearchBusiness(t *testing.T) {
	store := newTestStorage()

	tests := []struct {
		name    string
		page    cmdb.Page
		wantLen int
		wantErr bool
		checkFn func(t *testing.T, businesses []*cmdb.BusinessInfo)
	}{
		{
			name:    "normal_test_search_all_businesses",
			page:    cmdb.Page{Start: 0, Limit: 10},
			wantLen: 3,
			wantErr: false,
			checkFn: func(t *testing.T, businesses []*cmdb.BusinessInfo) {
				assert.Equal(t, 3, len(businesses))
				assert.Equal(t, int64(1), businesses[0].BKBizID)
				assert.Equal(t, int64(2), businesses[1].BKBizID)
				assert.Equal(t, int64(3), businesses[2].BKBizID)
				for idx, biz := range businesses {
					t.Logf("%d: biz-id(%d), biz-name(%s)", idx, biz.BKBizID, biz.BKBizName)
				}
			},
		},
		{
			name:    "normal_test_search_with_pagination",
			page:    cmdb.Page{Start: 0, Limit: 2},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "normal_test_search_with_offset",
			page:    cmdb.Page{Start: 1, Limit: 2},
			wantLen: 2,
			wantErr: false,
			checkFn: func(t *testing.T, businesses []*cmdb.BusinessInfo) {
				assert.Equal(t, int64(2), businesses[0].BKBizID)
			},
		},
		{
			name:    "normal_test_search_beyond_range",
			page:    cmdb.Page{Start: 10, Limit: 10},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			businesses, err := store.SearchBusiness(tt.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchBusiness() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(businesses) != tt.wantLen {
				t.Errorf("SearchBusiness() len = %v, wantLen %v", len(businesses), tt.wantLen)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, businesses)
			}
		})
	}
}

// TestStorage_GetBusinessCount tests GetBusinessCount.
func TestStorage_GetBusinessCount(t *testing.T) {
	store := newTestStorage()
	assert.Equal(t, 3, store.GetBusinessCount())
}

// TestStorage_SearchCloudAreas tests SearchCloudAreas.
func TestStorage_SearchCloudAreas(t *testing.T) {
	store := newTestStorage()

	tests := []struct {
		name    string
		page    cmdb.Page
		wantLen int
		wantErr bool
		checkFn func(t *testing.T, areas []*cmdb.CloudArea)
	}{
		{
			name:    "normal_test_search_all_cloud_areas",
			page:    cmdb.Page{Start: 0, Limit: 10},
			wantLen: 3,
			wantErr: false,
			checkFn: func(t *testing.T, areas []*cmdb.CloudArea) {
				assert.Equal(t, 3, len(areas))
				assert.Equal(t, int64(0), areas[0].BKCloudID)
				assert.Equal(t, int64(1), areas[1].BKCloudID)
				for idx, area := range areas {
					t.Logf("%d: cloud-id(%d), cloud-name(%s)", idx, area.BKCloudID, area.BKCloudName)
				}
			},
		},
		{
			name:    "normal_test_search_with_pagination",
			page:    cmdb.Page{Start: 0, Limit: 2},
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			areas, err := store.SearchCloudAreas(tt.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchCloudAreas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(areas) != tt.wantLen {
				t.Errorf("SearchCloudAreas() len = %v, wantLen %v", len(areas), tt.wantLen)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, areas)
			}
		})
	}
}

// TestStorage_GetCloudAreaCount tests GetCloudAreaCount.
func TestStorage_GetCloudAreaCount(t *testing.T) {
	store := newTestStorage()
	assert.Equal(t, 3, store.GetCloudAreaCount())
}

// TestStorage_ListBizHosts tests ListBizHosts.
func TestStorage_ListBizHosts(t *testing.T) {
	store := newTestStorage()

	tests := []struct {
		name    string
		bizID   int64
		page    cmdb.Page
		wantLen int
		wantErr bool
		checkFn func(t *testing.T, hosts []*cmdb.HostInfo)
	}{
		{
			name:    "normal_test_list_hosts_for_biz_1",
			bizID:   1,
			page:    cmdb.Page{Start: 0, Limit: 10},
			wantLen: 2,
			wantErr: false,
			checkFn: func(t *testing.T, hosts []*cmdb.HostInfo) {
				assert.Equal(t, 2, len(hosts))
				for _, host := range hosts {
					assert.Equal(t, int64(1), store.hostBiz[host.BKHostID])
					t.Logf("host-id(%d), biz-id(%d), inner-ip(%s)", host.BKHostID, store.hostBiz[host.BKHostID], host.BKHostInnerIPV4)
				}
			},
		},
		{
			name:    "normal_test_list_hosts_for_biz_2",
			bizID:   2,
			page:    cmdb.Page{Start: 0, Limit: 10},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "normal_test_list_hosts_for_non_existent_biz",
			bizID:   999,
			page:    cmdb.Page{Start: 0, Limit: 10},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hosts, err := store.ListBizHosts(tt.bizID, tt.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBizHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(hosts) != tt.wantLen {
				t.Errorf("ListBizHosts() len = %v, wantLen %v", len(hosts), tt.wantLen)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, hosts)
			}
		})
	}
}

// TestStorage_GetBizHostCount tests GetBizHostCount.
func TestStorage_GetBizHostCount(t *testing.T) {
	store := newTestStorage()
	assert.Equal(t, 2, store.GetBizHostCount(1))
	assert.Equal(t, 1, store.GetBizHostCount(2))
	assert.Equal(t, 0, store.GetBizHostCount(999))
}

// TestStorage_AddHostToBusinessIdle tests AddHostToBusinessIdle.
func TestStorage_AddHostToBusinessIdle(t *testing.T) {
	store := newTestStorage()

	tests := []struct {
		name      string
		bizID     int64
		hosts     []*cmdb.CreateHostInfo
		wantErr   bool
		wantCount int
		checkFn   func(t *testing.T, hostIDs []int64)
	}{
		{
			name:  "normal_test_add_host_successfully",
			bizID: 1,
			hosts: []*cmdb.CreateHostInfo{
				{
					BKCloudID:         0,
					BKHostInnerIP:     "192.168.1.1",
					BKOSType:          "linux",
					BKAddressing:      "static",
					BKCpuArchitecture: "x86_64",
				},
			},
			wantErr:   false,
			wantCount: 3, // 2 existing + 1 new
			checkFn: func(t *testing.T, hostIDs []int64) {
				assert.Equal(t, 1, len(hostIDs))
				assert.Greater(t, hostIDs[0], int64(0))
				for idx, hostID := range hostIDs {
					t.Logf("%d: host-id(%d)", idx, hostID)
				}
			},
		},
		{
			name:  "normal_test_add_multiple_hosts",
			bizID: 1,
			hosts: []*cmdb.CreateHostInfo{
				{BKCloudID: 0, BKHostInnerIP: "192.168.1.2", BKOSType: "linux"},
				{BKCloudID: 0, BKHostInnerIP: "192.168.1.3", BKOSType: "linux"},
			},
			wantErr:   false,
			wantCount: 5, // previous + 2 new
		},
		{
			name:  "normal_test_add_host_with_invalid_biz_id",
			bizID: -1,
			hosts: []*cmdb.CreateHostInfo{
				{BKCloudID: 0, BKHostInnerIP: "192.168.1.5", BKOSType: "linux"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostIDs, err := store.AddHostToBusinessIdle(tt.bizID, tt.hosts)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddHostToBusinessIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if hostIDs == nil {
					t.Errorf("AddHostToBusinessIdle() hostIDs = nil, want not nil")
					return
				}
				if tt.checkFn != nil {
					tt.checkFn(t, hostIDs)
				}
				if tt.wantCount > 0 {
					count := store.GetBizHostCount(tt.bizID)
					if count != tt.wantCount {
						t.Errorf("GetBizHostCount() = %v, wantCount %v", count, tt.wantCount)
					}
				}
			}
		})
	}
}

// TestStorage_BindHostAgent tests BindHostAgent.
func TestStorage_BindHostAgent(t *testing.T) {
	store := newTestStorage()

	// add a host
	hostIDs, err := store.AddHostToBusinessIdle(1, []*cmdb.CreateHostInfo{
		{BKCloudID: 0, BKHostInnerIP: "192.168.1.10", BKOSType: "linux"},
	})
	if err != nil {
		t.Fatalf("failed to add host for test: %v", err)
	}
	if len(hostIDs) != 1 {
		t.Fatalf("expected 1 host ID, got %d", len(hostIDs))
	}
	hostID := hostIDs[0]

	tests := []struct {
		name      string
		hostInfos []*cmdb.HostAgentIDInfo
		wantErr   bool
		checkFn   func(t *testing.T)
	}{
		{
			name: "normal_test_bind_agent_successfully",
			hostInfos: []*cmdb.HostAgentIDInfo{
				{BKHostID: hostID, BKAgentID: "agent-123"},
			},
			wantErr: false,
			checkFn: func(t *testing.T) {
				host, exists := store.hosts[hostID]
				if !exists {
					t.Errorf("host not found, host-id(%d)", hostID)
					return
				}
				if host.BKAgentID != "agent-123" {
					t.Errorf("BindHostAgent() agent-id = %v, want %v", host.BKAgentID, "agent-123")
				}
				t.Logf("host-id(%d), agent-id(%s)", hostID, host.BKAgentID)
			},
		},
		{
			name: "normal_test_bind_agent_to_non_existent_host",
			hostInfos: []*cmdb.HostAgentIDInfo{
				{BKHostID: 99999, BKAgentID: "agent-456"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.BindHostAgent(tt.hostInfos)
			if (err != nil) != tt.wantErr {
				t.Errorf("BindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t)
			}
		})
	}
}

// TestStorage_PushHostIdentifier tests PushHostIdentifier.
func TestStorage_PushHostIdentifier(t *testing.T) {
	store := newTestStorage()

	hostIDs := []int64{100, 101, 102}
	taskID := store.PushHostIdentifier(hostIDs)

	assert.NotEmpty(t, taskID)
	assert.Contains(t, taskID, taskIDPrefix)

	// Verify task is stored
	result := store.FindHostIdentifierPushResult(taskID)
	assert.Equal(t, hostIDs, result)
}

// TestStorage_FindHostIdentifierPushResult tests FindHostIdentifierPushResult.
func TestStorage_FindHostIdentifierPushResult(t *testing.T) {
	store := newTestStorage()

	hostIDs := []int64{200, 201}
	taskID := store.PushHostIdentifier(hostIDs)

	tests := []struct {
		name    string
		taskID  string
		wantLen int
		checkFn func(t *testing.T, result []int64)
	}{
		{
			name:    "find existing task",
			taskID:  taskID,
			wantLen: 2,
			checkFn: func(t *testing.T, result []int64) {
				assert.Equal(t, hostIDs, result)
			},
		},
		{
			name:    "find non-existent task",
			taskID:  "non-existent-task",
			wantLen: 0,
			checkFn: func(t *testing.T, result []int64) {
				// Should return empty slice instead of nil for consistency.
				assert.NotNil(t, result)
				assert.Equal(t, 0, len(result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.FindHostIdentifierPushResult(tt.taskID)
			assert.Equal(t, tt.wantLen, len(result))
			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

// TestStorage_WatchResource tests WatchResource.
func TestStorage_WatchResource(t *testing.T) {
	store := newTestStorage()

	// Add a host to generate events
	hostIDs, err := store.AddHostToBusinessIdle(1, []*cmdb.CreateHostInfo{
		{BKCloudID: 0, BKHostInnerIP: "192.168.1.20", BKOSType: "linux"},
	})
	if err != nil {
		t.Fatalf("failed to add host for test: %v", err)
	}
	if len(hostIDs) != 1 {
		t.Fatalf("expected 1 host ID, got %d", len(hostIDs))
	}

	tests := []struct {
		name     string
		resource ResourceType
		cursor   string
		wantLen  int
		wantErr  bool
		checkFn  func(t *testing.T, events []*watchEvent)
	}{
		{
			name:     "normal_test_watch_host_resource_with_empty_cursor",
			resource: ResourceTypeHost,
			cursor:   "",
			wantLen:  1, // latest event
			wantErr:  false,
			checkFn: func(t *testing.T, events []*watchEvent) {
				assert.Equal(t, 1, len(events))
				assert.Equal(t, ResourceTypeHost, events[0].Resource)
				assert.Equal(t, EventTypeCreate, events[0].EventType)
				for idx, event := range events {
					t.Logf("%d: cursor(%s), resource(%s), event-type(%s)", idx, event.Cursor, event.Resource, event.EventType)
				}
			},
		},
		{
			name:     "normal_test_watch_host_resource_with_cursor",
			resource: ResourceTypeHost,
			cursor:   store.hostEvents[0].Cursor,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "normal_test_watch_host_resource_with_nonexistent_cursor",
			resource: ResourceTypeHost,
			cursor:   "nonexistent_cursor_12345",
			wantLen:  0,
			wantErr:  true,
			checkFn: func(t *testing.T, events []*watchEvent) {
				// should return error when cursor not found
			},
		},
		{
			name:     "normal_test_watch_host_relation_resource",
			resource: ResourceTypeHostRelation,
			cursor:   "",
			wantLen:  1,
			wantErr:  false,
			checkFn: func(t *testing.T, events []*watchEvent) {
				assert.Equal(t, ResourceTypeHostRelation, events[0].Resource)
			},
		},
		{
			name:     "normal_test_watch_unsupported_resource",
			resource: ResourceType("unsupported"),
			cursor:   "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := store.WatchResource(tt.resource, tt.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("WatchResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(events) != tt.wantLen {
				t.Errorf("WatchResource() len = %v, wantLen %v", len(events), tt.wantLen)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, events)
			}
		})
	}
}
