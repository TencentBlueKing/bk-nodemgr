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
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	defaultRelayCallbackBindIP = "127.0.0.1"
	defaultRelayCallbackPort   = 8001
	defaultRelayFileBindIP     = "127.0.0.1"
	defaultRelayFilePort       = 8002
	defaultRelayAdvertiseIPv4  = "127.0.0.1"
	defaultRelayAdvertiseIPv6  = "::1"

	defaultRelayLogDir       = "/bk-nodemgr/log/"
	defaultRelayLogMaxNum    = 10
	defaultRelayLogMaxSizeMB = 200
	defaultRelayLogLevel     = "INFO"
	defaultRelayPluginName   = "bk-nodemgr-relay"

	defaultRelayMessageIDPath     = "/usr/local/gse2/proxy/bin/tmp/message_id"
	defaultRelayStorageTmpDirPath = "/usr/local/gse2/proxy/bin/tmp/transfer_files"

	defaultRelayFileManagerDirPath = "/usr/local/gse2/proxy/bin/file_manager"
)

// RelayService the config of relay service.
type RelayService struct {
	Plugin GSEPlugin `yaml:"plugin" usage:"gse agent plugin config of relay service"`

	AgentFileGroup FileGroup `yaml:"agentFileGroup" usage:"agent file group config of relay service"`
	ProxyFileGroup FileGroup `yaml:"proxyFileGroup" usage:"proxy file group config of relay service"`

	CallbackServer CallbackServer `yaml:"callbackServer" usage:"callback server config of relay service"`
	FileServer     HTTPServer     `yaml:"fileServer" usage:"file server config of relay service"`

	PluginName string `yaml:"pluginName" usage:"gse agent plugin name of relay service"`

	MessageIDPath string `yaml:"messageIDPath" usage:"message id full path of relay service"`

	FileManagerDirPath string `yaml:"fileManagerDirPath" usage:"file manager dir path of relay service"`
	StorageTmpDirPath  string `yaml:"storageTmpDirPath" usage:"storage tmp dir path of relay service"`

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
		FileServer: HTTPServer{
			BindIP:        defaultRelayFileBindIP,
			Port:          defaultRelayFilePort,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		Log: Log{
			Dir:       defaultRelayLogDir,
			MaxSizeMB: defaultRelayLogMaxSizeMB,
			MaxNum:    defaultRelayLogMaxNum,
			Level:     defaultRelayLogLevel,
		},
		MessageIDPath:      defaultRelayMessageIDPath,
		PluginName:         defaultRelayPluginName,
		FileManagerDirPath: defaultRelayFileManagerDirPath,
		StorageTmpDirPath:  defaultRelayStorageTmpDirPath,
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
	if err := svc.AgentFileGroup.Validate(); err != nil {
		return err
	}

	if err := svc.ProxyFileGroup.Validate(); err != nil {
		return err
	}

	return nil
}
