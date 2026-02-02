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
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
)

const (
	// DefaultNextHostID is the default next host id.
	DefaultNextHostID = 1

	taskIDPrefix             = "mock_task_"
	cursorHostResourcePrefix = "mock_cursor_host_resource_"
	cursorHostRelationPrefix = "mock_cursor_host_relation_"
)

// watchEvent represents a watch event for CMDB resources.
type watchEvent struct {
	Cursor    string
	Resource  ResourceType
	EventType EventType
	Detail    any
}

// storage provides thread-safe in-memory storage for CMDB mock data.
type storage struct {
	mu sync.RWMutex
	// business info map.
	businesses map[int64]*cmdb.BusinessInfo // biz id -> business info.

	// cloud area info map.
	cloudAreas map[int64]*cmdb.CloudArea // cloud id -> cloud area info.

	// host info map.
	hosts      map[int64]*cmdb.HostInfo // host id -> host info.
	hostBiz    map[int64]int64          // host id -> biz id.
	nextHostID int64                    // next host id.

	// identifier task map.
	identifierTasks map[string][]int64 // task id -> host ids.

	// watch event queues.
	hostEvents         []*watchEvent // host resource events.
	hostRelationEvents []*watchEvent // host relation resource events.
	cursorCounter      int64         // cursor counter for generating unique cursors.
}

// newStorage creates a new storage instance and loads data from config.
func newStorage(conf *Config) *storage {
	s := &storage{
		businesses:         make(map[int64]*cmdb.BusinessInfo),
		cloudAreas:         make(map[int64]*cmdb.CloudArea),
		hosts:              make(map[int64]*cmdb.HostInfo),
		hostBiz:            make(map[int64]int64),
		nextHostID:         DefaultNextHostID,
		identifierTasks:    make(map[string][]int64),
		hostEvents:         make([]*watchEvent, 0),
		hostRelationEvents: make([]*watchEvent, 0),
		cursorCounter:      1,
	}

	s.loadDataFromConfig(conf)

	return s
}

// loadDataFromConfig loads data from config.
func (s *storage) loadDataFromConfig(conf *Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loadBusinessesLocked(conf)
	s.loadCloudAreasLocked(conf)
	s.loadHostsLocked(conf)
}

// SearchBusiness searches businesses with pagination support.
func (s *storage) SearchBusiness(page cmdb.Page) ([]*cmdb.BusinessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allBusinesses := conv.MapValueToSlice(s.businesses)
	sort.Slice(allBusinesses, func(i, j int) bool {
		return allBusinesses[i].BKBizID < allBusinesses[j].BKBizID
	})

	return paginateSlice(allBusinesses, page), nil
}

// SearchCloudAreas searches cloud areas with pagination support.
func (s *storage) SearchCloudAreas(page cmdb.Page) ([]*cmdb.CloudArea, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allAreas := conv.MapValueToSlice(s.cloudAreas)
	sort.Slice(allAreas, func(i, j int) bool {
		return allAreas[i].BKCloudID < allAreas[j].BKCloudID
	})

	return paginateSlice(allAreas, page), nil
}

// ListBizHosts lists hosts by biz id with pagination support.
func (s *storage) ListBizHosts(bizID int64, page cmdb.Page) ([]*cmdb.HostInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hostList := make([]*cmdb.HostInfo, 0)
	for hostID, host := range s.hosts {
		if s.hostBiz[hostID] != bizID {
			continue
		}
		hostList = append(hostList, host)
	}

	sort.Slice(hostList, func(i, j int) bool {
		return hostList[i].BKHostID < hostList[j].BKHostID
	})

	return paginateSlice(hostList, page), nil
}

// GetBusinessCount returns the total count of businesses.
func (s *storage) GetBusinessCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.businesses)
}

// GetCloudAreaCount returns the total count of cloud areas.
func (s *storage) GetCloudAreaCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.cloudAreas)
}

// GetBizHostCount returns the total count of hosts for a biz.
func (s *storage) GetBizHostCount(bizID int64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for hostID := range s.hosts {
		if s.hostBiz[hostID] != bizID {
			continue
		}
		count++
	}

	return count
}

// AddHostToBusinessIdle adds hosts to business idle and returns host IDs.
func (s *storage) AddHostToBusinessIdle(bizID int64, hosts []*cmdb.CreateHostInfo) ([]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if bizID < 0 {
		return nil, fmt.Errorf("invalid biz id. biz-id(%d)", bizID)
	}

	// check if business exists.
	if _, exists := s.businesses[bizID]; !exists {
		return nil, fmt.Errorf("business not found, biz-id(%d)", bizID)
	}

	hostIDs := make([]int64, len(hosts))
	for idx, host := range hosts {
		// generate host id.
		hostID := s.nextHostID
		s.nextHostID++

		hostIDs[idx] = hostID
		s.hostBiz[hostID] = bizID

		hostInfo := &cmdb.HostInfo{
			BKHostID:          hostID,
			BKCloudID:         host.BKCloudID,
			BKHostInnerIPV4:   host.BKHostInnerIP,
			BKHostInnerIPV6:   host.BKHostInnerIPV6,
			BKHostOuterIPV4:   host.BKHostOuterIP,
			BKHostOuterIPV6:   host.BKHostOuterIPV6,
			BKOSType:          host.BKOSType,
			BKAddressing:      host.BKAddressing,
			BKCpuArchitecture: host.BKCpuArchitecture,
		}
		s.hosts[hostID] = hostInfo

		// generate host create event.
		hostCursor := fmt.Sprintf("%s%d", cursorHostResourcePrefix, s.cursorCounter)
		s.cursorCounter++
		s.hostEvents = append(s.hostEvents, &watchEvent{
			Cursor:    hostCursor,
			Resource:  ResourceTypeHost,
			EventType: EventTypeCreate,
			Detail:    hostInfo,
		})

		// generate host_relation create event.
		relationCursor := fmt.Sprintf("%s%d", cursorHostRelationPrefix, s.cursorCounter)
		s.cursorCounter++
		s.hostRelationEvents = append(s.hostRelationEvents, &watchEvent{
			Cursor:    relationCursor,
			Resource:  ResourceTypeHostRelation,
			EventType: EventTypeCreate,
			Detail: &cmdb.HostTopoRelation{
				BKBizID:  bizID,
				BKHostID: hostID,
			},
		})
	}

	return hostIDs, nil
}

// BindHostAgent binds agent id to hosts.
func (s *storage) BindHostAgent(hostInfos []*cmdb.HostAgentIDInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, host := range hostInfos {
		info, exists := s.hosts[host.BKHostID]
		if !exists {
			return fmt.Errorf("host not found, host-id(%d)", host.BKHostID)
		}
		info.BKAgentID = host.BKAgentID

		// generate host update event.
		hostCursor := fmt.Sprintf("%s%d", cursorHostResourcePrefix, s.cursorCounter)
		s.cursorCounter++
		s.hostEvents = append(s.hostEvents, &watchEvent{
			Cursor:    hostCursor,
			Resource:  ResourceTypeHost,
			EventType: EventTypeUpdate,
			Detail:    info,
		})
	}

	return nil
}

// PushHostIdentifier creates a task and returns the task id.
func (s *storage) PushHostIdentifier(hostIDs []int64) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskID := fmt.Sprintf("%s%d", taskIDPrefix, time.Now().UnixNano())
	s.identifierTasks[taskID] = hostIDs

	return taskID
}

// FindHostIdentifierPushResult returns success list for a task.
func (s *storage) FindHostIdentifierPushResult(taskID string) []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.identifierTasks[taskID]
	if !exists {
		return []int64{}
	}

	return result
}

// WatchResource returns watch events for the given resource type and cursor.
func (s *storage) WatchResource(resource ResourceType, cursor string) ([]*watchEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*watchEvent
	switch resource {
	case ResourceTypeHost:
		events = s.hostEvents
	case ResourceTypeHostRelation:
		events = s.hostRelationEvents
	default:
		return nil, fmt.Errorf("unsupported resource type. type(%s)", resource)
	}

	// if cursor is empty, return latest event.
	if cursor == "" {
		if len(events) == 0 {
			return []*watchEvent{}, nil
		}

		return []*watchEvent{events[len(events)-1]}, nil
	}

	// find events after the cursor.
	result := make([]*watchEvent, 0)
	found := false
	for _, event := range events {
		if found {
			result = append(result, event)
		}
		if event.Cursor == cursor {
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("event chain node not exist, cursor(%s)", cursor)
	}

	return result, nil
}

// paginateSlice applies pagination to a slice and returns the paginated result.
func paginateSlice[T any](items []T, page cmdb.Page) []T {
	start := page.Start
	if start >= len(items) {
		return []T{}
	}

	// use cc page size limit.
	limit := page.Limit
	if limit <= 0 {
		limit = cmdb.CCPageSizeLimit
	}

	end := min(start+limit, len(items))

	return items[start:end]
}

// loadBusinessesLocked loads businesses from config.
func (s *storage) loadBusinessesLocked(conf *Config) {
	if conf == nil || len(conf.Businesses) == 0 {
		return
	}

	for _, biz := range conf.Businesses {
		s.businesses[biz.BKBizID] = &cmdb.BusinessInfo{
			BKBizID:   biz.BKBizID,
			BKBizName: biz.BKBizName,
		}
	}

	logger.G.Sys().With("count", len(s.businesses)).Info("loaded businesses")
}

// loadCloudAreasLocked loads cloud areas from config.
func (s *storage) loadCloudAreasLocked(conf *Config) {
	if conf == nil || len(conf.Areas) == 0 {
		return
	}

	for _, area := range conf.Areas {
		s.cloudAreas[area.BKCloudID] = &cmdb.CloudArea{
			BKCloudID:   area.BKCloudID,
			BKCloudName: area.BKCloudName,
		}
	}

	logger.G.Sys().With("count", len(s.cloudAreas)).Info("loaded cloud areas")
}

// loadHostsLocked loads hosts from config.
func (s *storage) loadHostsLocked(conf *Config) {
	if conf == nil || len(conf.Hosts) == 0 {
		return
	}

	for _, host := range conf.Hosts {
		hostInfo := &cmdb.HostInfo{
			BKHostID:          host.BKHostID,
			BKCloudID:         host.BKCloudID,
			BKHostInnerIPV4:   host.BKInnerIP,
			BKHostInnerIPV6:   host.BKInnerIPv6,
			BKOSType:          host.BKOSType,
			BKCpuArchitecture: host.BKCpuArch,
		}
		s.hosts[host.BKHostID] = hostInfo
		s.hostBiz[host.BKHostID] = host.BKBizID

		// update nextHostID if needed.
		if host.BKHostID >= s.nextHostID {
			s.nextHostID = host.BKHostID + 1
		}
	}

	logger.G.Sys().With("count", len(s.hosts), "nextHostID", s.nextHostID).Info("loaded hosts")
}
