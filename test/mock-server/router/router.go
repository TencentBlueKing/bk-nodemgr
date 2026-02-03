/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package router provides routing for mock server.
package router

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/cmdb"
	"github.com/gin-gonic/gin"
)

// MockData holds the optional preset mock data configuration for all mock routers.
type MockData struct {
	// CMDB preset mock data (businesses, areas, hosts, etc.).
	CMDB *cmdb.MockData `yaml:"cmdb"`
}

// Validate validates the MockData.
func (m *MockData) Validate() error {
	if m.CMDB != nil {
		if err := m.CMDB.Validate(); err != nil {
			return fmt.Errorf("failed to validate cmdb mock data: %w", err)
		}
	}

	return nil
}

// Load registers all mock service API routes with the given gin router group.
func Load(rg *gin.RouterGroup, cmdbConfig *cmdb.Config, bkrepoConfig *bkrepo.Config, mockData *MockData) {
	if mockData == nil {
		mockData = &MockData{}
	}

	// load cmdb mock API routes.
	cmdb.Load(rg, cmdbConfig, mockData.CMDB)

	// load bkrepo mock API routes.
	bkrepo.Load(rg, bkrepoConfig)
}
