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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
)

// storage provides thread-safe in-memory storage for CMDB mock data.
type storage struct {
	mu         sync.RWMutex
	businesses map[int64]*cmdb.BusinessInfo
	// TODO: add more data here. such as areas, hosts, etc.
}

// newStorage creates a new storage instance and loads data from config.
func newStorage(conf *Config) *storage {
	s := &storage{
		businesses: make(map[int64]*cmdb.BusinessInfo),
	}

	// Load businesses from config.
	s.loadBusinesses(conf)

	return s
}

// SearchBusiness searches businesses with pagination support.
func (s *storage) SearchBusiness(page cmdb.Page) ([]*cmdb.BusinessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice using conv package.
	allBusinesses := conv.MapValueToSlice(s.businesses)

	start := page.Start
	if start >= len(allBusinesses) {
		return []*cmdb.BusinessInfo{}, nil
	}

	// use cc page size limit.
	limit := page.Limit
	if limit <= 0 {
		limit = cmdb.CCPageSizeLimit
	}

	end := start + limit
	if end > len(allBusinesses) {
		end = len(allBusinesses)
	}

	return allBusinesses[start:end], nil
}

// GetBusinessCount returns the total count of businesses.
func (s *storage) GetBusinessCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.businesses)
}

// loadBusinesses loads businesses from config.
func (s *storage) loadBusinesses(conf *Config) {
	if conf != nil && len(conf.Businesses) > 0 {
		for _, biz := range conf.Businesses {
			s.businesses[biz.BKBizID] = &cmdb.BusinessInfo{
				BKBizID:   biz.BKBizID,
				BKBizName: biz.BKBizName,
			}
		}
	}

	logger.G.Sys().With("count", len(s.businesses)).Info("loaded businesses")
}
