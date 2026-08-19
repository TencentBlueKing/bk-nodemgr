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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/cmdb"
	"gopkg.in/yaml.v2"
)

const (
	// basic server config default values.
	defaultMockBasicBindIP   = "127.0.0.1"
	defaultMockBasicBindIPV6 = "::1"
	defaultMockBasicPort     = 28400
	defaultMockBasicIdentity = config.AuthIdentityNone

	// default log config.
	defaultMockLogDir       = "./logs"
	defaultMockLogMaxSizeMB = 200
	defaultMockLogMaxNum    = 10
	defaultMockLogLevel     = config.LogLevelInfo

	defaultMockAdvertiseIPv4 = "127.0.0.1"
	defaultMockAdvertiseIPv6 = "::1"

	defaultMockBKRepoBaseDir = "./bk-repo"
)

// MockService holds mock-server configuration.
type MockService struct {
	BasicServer config.HTTPServer `yaml:"basicServer" usage:"basic server config of mock-server"`
	Log         config.Log        `yaml:"log" usage:"log config of mock-server"`

	// CMDBConfig holds the CMDB router configuration.
	CMDBConfig *cmdb.Config `yaml:"cmdbConfig" usage:"cmdb router config"`

	// BKRepoConfig holds the BKRepo router configuration.
	BKRepoConfig *bkrepo.Config `yaml:"bkrepoConfig" usage:"bkrepo router config"`

	// MockData holds the optional preset mock data configuration.
	MockData router.MockData `yaml:"mockData" usage:"optional preset mock data config"`
}

// NewMockService generates a new MockService with default values.
func NewMockService() *MockService {
	return &MockService{
		BasicServer: config.HTTPServer{
			BindIP:        defaultMockBasicBindIP,
			BindIPV6:      defaultMockBasicBindIPV6,
			Port:          defaultMockBasicPort,
			AuthIdentity:  defaultMockBasicIdentity,
			AdvertiseIPV4: defaultMockAdvertiseIPv4,
			AdvertiseIPV6: defaultMockAdvertiseIPv6,
		},
		Log: config.Log{
			Dir:       defaultMockLogDir,
			MaxSizeMB: defaultMockLogMaxSizeMB,
			MaxNum:    defaultMockLogMaxNum,
			Level:     defaultMockLogLevel,
		},
		BKRepoConfig: &bkrepo.Config{
			BaseDir: defaultMockBKRepoBaseDir,
		},
	}
}

// GetMockData returns the MockData pointer.
func (conf *MockService) GetMockData() *router.MockData {
	return &conf.MockData
}

// LoadFromFile loads config from file.
func (conf *MockService) LoadFromFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	absPath = filepath.Clean(absPath)
	configContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(configContent, conf); err != nil {
		return err
	}

	return nil
}

// Validate validates the config.
func (conf *MockService) Validate() error {
	if err := conf.BasicServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate basic server config: %w", err)
	}

	if err := conf.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if conf.CMDBConfig != nil {
		if err := conf.CMDBConfig.Validate(); err != nil {
			return fmt.Errorf("failed to validate cmdb config: %w", err)
		}
	}

	if conf.BKRepoConfig != nil {
		if err := conf.BKRepoConfig.Validate(); err != nil {
			return fmt.Errorf("failed to validate bkrepo config: %w", err)
		}
	}

	if err := conf.MockData.Validate(); err != nil {
		return fmt.Errorf("failed to validate mock data config: %w", err)
	}

	return nil
}
