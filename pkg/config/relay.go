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
	defaultRelayPluginName = "bk-nodemgr-relay"

	defaultRelayInfoBindIP           = "127.0.0.1"
	defaultRelayInfoBindIPV6         = "::1"
	defaultRelayInfoPort             = 28300
	defaultRelayInfoTraceServiceName = "relay-server-info"
	defaultRelayInfoIdentity         = AuthIdentityNone

	defaultRelayAdminBindIP           = "127.0.0.1"
	defaultRelayAdminBindIPV6         = "::1"
	defaultRelayAdminPort             = 28301
	defaultRelayAdminTraceServiceName = "relay-server-admin"
	defaultRelayAdminIdentity         = AuthIdentityNone

	defaultRelayCallbackBindIP           = "127.0.0.1"
	defaultRelayCallbackBindIPV6         = "::1"
	defaultRelayCallbackPort             = 28302
	defaultRelayCallbackTraceServiceName = "relay-server-callback"
	defaultRelayCallbackIdentity         = AuthIdentityNone

	defaultRelayDownloadBindIP           = "127.0.0.1"
	defaultRelayDownloadBindIPV6         = "::1"
	defaultRelayDownloadPort             = 28303
	defaultRelayDownloadTraceServiceName = "relay-server-download"
	defaultRelayDownloadIdentity         = AuthIdentityNone

	defaultRelayAdvertiseIPv4 = "127.0.0.1"
	defaultRelayAdvertiseIPv6 = "::1"

	defaultRelayLogDir       = "/data/plugin-relay/log"
	defaultRelayLogMaxNum    = 10
	defaultRelayLogMaxSizeMB = 200
	defaultRelayLogLevel     = "INFO"

	defaultRelayWorkspaceGroupFullPath = "/data/plugin-relay"

	defaultRelayTracingExporterType = "stdout"
)

// RelayService the config of relay service.
type RelayService struct {
	PluginName PluginName `yaml:"pluginName" usage:"gse agent plugin name of relay service"`
	Plugin     GSEPlugin  `yaml:"plugin" usage:"gse agent plugin config of relay service"`

	InfoServer     HTTPServer     `yaml:"infoServer" usage:"info server config of relay service"`
	AdminServer    HTTPServer     `yaml:"adminServer" usage:"admin server config of relay service"`
	CallbackServer CallbackServer `yaml:"callbackServer" usage:"callback server config of relay service"`
	DownloadServer HTTPServer     `yaml:"downloadServer" usage:"download server config of download service"`

	RelayWorkspaceFileGroup FileGroup `yaml:"relayWorkspaceFileGroup" usage:"relay workspace file group config of relay service"`

	Tracing Tracing `yaml:"tracing" usage:"tracing config of relay service"`
	Log     Log     `yaml:"log" usage:"log config of relay service"`
}

// NewRelayService generates a new RelayService with default value.
func NewRelayService() *RelayService {
	return &RelayService{
		PluginName: defaultRelayPluginName,
		Plugin: GSEPlugin{
			PidFile:                 "bk-nodemgr-relay.pid",
			MessageDomainSocketPath: "",
			MessageLocalSocketPort:  0,
		},
		InfoServer: HTTPServer{
			BindIP:   defaultRelayInfoBindIP,
			BindIPV6: defaultRelayInfoBindIPV6,
			Port:     defaultRelayInfoPort,
			TraceService: TraceService{
				TraceServiceName: defaultRelayInfoTraceServiceName,
			},
			AuthIdentity:  defaultRelayInfoIdentity,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		AdminServer: HTTPServer{
			BindIP:   defaultRelayAdminBindIP,
			BindIPV6: defaultRelayAdminBindIPV6,
			Port:     defaultRelayAdminPort,
			TraceService: TraceService{
				TraceServiceName: defaultRelayAdminTraceServiceName,
			},
			AuthIdentity:  defaultRelayAdminIdentity,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		CallbackServer: CallbackServer{
			HTTPServer{
				BindIP:   defaultRelayCallbackBindIP,
				BindIPV6: defaultRelayCallbackBindIPV6,
				Port:     defaultRelayCallbackPort,
				TraceService: TraceService{
					TraceServiceName: defaultRelayCallbackTraceServiceName,
				},
				AuthIdentity:  defaultRelayCallbackIdentity,
				AdvertiseIPV4: defaultRelayAdvertiseIPv4,
				AdvertiseIPV6: defaultRelayAdvertiseIPv6,
			}},
		DownloadServer: HTTPServer{
			BindIP:   defaultRelayDownloadBindIP,
			BindIPV6: defaultRelayDownloadBindIPV6,
			Port:     defaultRelayDownloadPort,
			TraceService: TraceService{
				TraceServiceName: defaultRelayDownloadTraceServiceName,
			},
			AuthIdentity:  defaultRelayDownloadIdentity,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		Log: Log{
			Dir:       defaultRelayLogDir,
			MaxSizeMB: defaultRelayLogMaxSizeMB,
			MaxNum:    defaultRelayLogMaxNum,
			Level:     defaultRelayLogLevel,
		},
		RelayWorkspaceFileGroup: FileGroup{
			FullPath: defaultRelayWorkspaceGroupFullPath,
		},
		Tracing: Tracing{
			ExporterType: defaultRelayTracingExporterType,
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
	if err := svc.PluginName.Validate(); err != nil {
		return fmt.Errorf("failed to validate plugin name config: %w", err)
	}

	if err := svc.InfoServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate info service config: %w", err)
	}
	if err := svc.AdminServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate admin service config: %w", err)
	}
	if err := svc.DownloadServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate node service config: %w", err)
	}
	if err := svc.CallbackServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate callback service config: %w", err)
	}

	if err := svc.RelayWorkspaceFileGroup.Validate(); err != nil {
		return fmt.Errorf("failed to validate workspace file group config: %w", err)
	}

	if err := svc.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if err := svc.Tracing.Validate(); err != nil {
		return fmt.Errorf("failed to validate tracing config: %w", err)
	}

	return nil
}
