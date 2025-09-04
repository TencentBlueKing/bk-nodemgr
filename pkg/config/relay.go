/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	defaultRelayCallbackBindIP = "127.0.0.1"
	defaultRelayCallbackPort   = 8001
	defaultRelayDownloadBindIP = "127.0.0.1"
	defaultRelayDownloadPort   = 8002
	defaultRelayAdvertiseIPv4  = "127.0.0.1"
	defaultRelayAdvertiseIPv6  = "::1"

	defaultRelayLogDir       = "/bk-nodemgr/log/"
	defaultRelayLogMaxNum    = 10
	defaultRelayLogMaxSizeMB = 200
	defaultRelayLogLevel     = "INFO"
	defaultRelayPluginName   = "bk-nodemgr-relay"

	defaultRelayWorkspaceGroupFullPath = "/data/plugin-relay"
)

// RelayService the config of relay service.
type RelayService struct {
	PluginName string    `yaml:"pluginName" usage:"gse agent plugin name of relay service"`
	Plugin     GSEPlugin `yaml:"plugin" usage:"gse agent plugin config of relay service"`

	CallbackServer CallbackServer `yaml:"callbackServer" usage:"callback server config of relay service"`
	DownloadServer HTTPServer     `yaml:"downloadServer" usage:"download server config of file service"`

	RelayWorkspaceFileGroup FileGroup `yaml:"relayWorkspaceFileGroup" usage:"workspace file group config of file service"`

	Log Log `yaml:"log" usage:"log config of relay service"`
}

// NewRelayService generates a new RelayService with default value.
func NewRelayService() *RelayService {
	return &RelayService{
		Plugin: GSEPlugin{
			PidFile:                 "bk-nodemgr-relay.pid",
			MessageDomainSocketPath: "",
			MessageLocalSocketPort:  0,
		},
		CallbackServer: CallbackServer{
			HTTPServer{
				BindIP:        defaultRelayCallbackBindIP,
				Port:          defaultRelayCallbackPort,
				AdvertiseIPV4: defaultRelayAdvertiseIPv4,
				AdvertiseIPV6: defaultRelayAdvertiseIPv6,
			}},
		DownloadServer: HTTPServer{
			BindIP:        defaultRelayDownloadBindIP,
			Port:          defaultRelayDownloadPort,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		Log: Log{
			Dir:       defaultRelayLogDir,
			MaxSizeMB: defaultRelayLogMaxSizeMB,
			MaxNum:    defaultRelayLogMaxNum,
			Level:     defaultRelayLogLevel,
		},
		PluginName: defaultRelayPluginName,
		RelayWorkspaceFileGroup: FileGroup{
			FullPath: defaultRelayWorkspaceGroupFullPath,
		},
	}
}

// LoadFromFile loads config from file.
func (svc *RelayService) LoadFromFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	absPath = filepath.Clean(absPath)
	configContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(configContent, svc); err != nil {
		return err
	}

	return nil
}

// Validate validates the config.
func (svc *RelayService) Validate() error {
	if err := svc.RelayWorkspaceFileGroup.Validate(); err != nil {
		return fmt.Errorf("failed to validate workspace file group config: %w", err)
	}

	if err := svc.DownloadServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate node service config: %w", err)
	}

	if err := svc.CallbackServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate callback service config: %w", err)
	}

	if err := svc.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	return nil
}
