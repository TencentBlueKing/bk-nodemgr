/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package config ...
package config

import (
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"gopkg.in/yaml.v2"
)

const (
	// backend service config default values.
	defaultBackendRunMode        = RunModeRelease
	defaultBackendTenantMode     = tenant.ModeSingle
	defaultBackendHTTPBindIP     = "127.0.0.1"
	defaultBackendHTTPPort       = 8000
	defaultBackendAdminBindIP    = "127.0.0.1"
	defaultBackendAdminPort      = 8001
	defaultBackendCallbackBindIP = "127.0.0.1"
	defaultBackendCallbackPort   = 8002
	defaultBackendProxyBindIP    = "127.0.0.1"
	defaultBackendProxyPort      = 8003
	defaultBackendLogDir         = "/bk-nodeman/log/"
	defaultBackendLogMaxNum      = 10
	defaultBackendLogMaxSizeMB   = 200
	defaultBackendLogLevel       = "INFO"
	defaultBackendEncryptKey     = "1234567890abcdef"
	defaultBackendSystemEnv      = "dev"
	defaultBackendSystemEdition  = "ce"
	defaultBackendAdvertiseIPv4  = "127.0.0.1"
	defaultBackendAdvertiseIPv6  = "::1"

	defaultInstallerFileGroup = "/bk-nodeman/file/tools"

	defaultGseDeployConfLinuxGeneration    = 2
	defaultGseDeployConfLinuxOsType        = string(criteria.OSLinux)
	defaultGseDeployConfLinuxBaseDeployDir = "/usr/local/"
	defaultGseDeployConfLinuxBaseWorkDir   = "/tmp/bknm/"

	defaultGseDeployConfWindowsGeneration    = 2
	defaultGseDeployConfWindowsOsType        = string(criteria.OSWindows)
	defaultGseDeployConfWindowsBaseDeployDir = `c:\`
	defaultGseDeployConfWindowsBaseWorkDir   = `c:\tmp\bknm\`
)

// BackendService the config of backend service.
type BackendService struct {
	RunMode            RunMode         `yaml:"runMode" usage:"run mode of service"`
	TenantMode         tenant.Mode     `yaml:"tenantMode" usage:"tenant mode of service"`
	CMDB               CMDB            `yaml:"cmdb" usage:"cmdb config of backend service"`
	GSE                GSE             `yaml:"gse" usage:"gse config of backend service"`
	Workflow           Workflow        `yaml:"workflow" usage:"workflow config of backend service"`
	HTTPServer         HTTPServer      `yaml:"httpServer" usage:"http server config of backend service"`
	AdminServer        HTTPServer      `yaml:"adminServer" usage:"admin server config of backend service"`
	CallbackServer     CallbackServer  `yaml:"callbackServer" usage:"callback server config of backend service"`
	ProxyServer        ProxyServer     `yaml:"proxyServer" usage:"proxy server config of backend service"`
	Etcd               Etcd            `yaml:"etcd" usage:"etcd config of backend service"`
	Redis              Redis           `yaml:"redis" usage:"redis config of backend service"`
	MongoDB            MongoDB         `yaml:"mongodb" usage:"mongodb config of backend service"`
	Log                Log             `yaml:"log" usage:"log config of backend service"`
	System             System          `yaml:"system" usage:"system config of backend service"`
	EncryptKey         string          `yaml:"encryptKey" usage:"encrypt key of backend service"`
	GSEDeployConfs     []GSEDeployConf `yaml:"gseDeployConfs" usage:"gse deploy config of backend service"`
	InstallerFileGroup FileGroup       `yaml:"installerFileGroup" usage:"tools file group config of backend service"`
}

// NewBackendService generates a new BackendService with default values.
func NewBackendService() *BackendService {
	return &BackendService{
		RunMode:    defaultBackendRunMode,
		TenantMode: defaultBackendTenantMode,
		HTTPServer: HTTPServer{
			BindIP:        defaultBackendHTTPBindIP,
			Port:          defaultBackendHTTPPort,
			AdvertiseIPV4: defaultBackendAdvertiseIPv4,
			AdvertiseIPV6: defaultBackendAdvertiseIPv6,
		},
		AdminServer: HTTPServer{
			BindIP: defaultBackendAdminBindIP,
			Port:   defaultBackendAdminPort,
		},
		CallbackServer: CallbackServer{
			HTTPServer: HTTPServer{
				BindIP:        defaultBackendCallbackBindIP,
				Port:          defaultBackendCallbackPort,
				AdvertiseIPV4: defaultBackendAdvertiseIPv4,
				AdvertiseIPV6: defaultBackendAdvertiseIPv6,
			},
		},
		ProxyServer: ProxyServer{
			HTTPServer: HTTPServer{
				BindIP:        defaultBackendProxyBindIP,
				Port:          defaultBackendProxyPort,
				AdvertiseIPV4: defaultBackendAdvertiseIPv4,
				AdvertiseIPV6: defaultBackendAdvertiseIPv6,
			},
		},
		Log: Log{
			Dir:       defaultBackendLogDir,
			MaxSizeMB: defaultBackendLogMaxSizeMB,
			MaxNum:    defaultBackendLogMaxNum,
			Level:     defaultBackendLogLevel,
		},
		EncryptKey: defaultBackendEncryptKey,
		System: System{
			Env:     defaultBackendSystemEnv,
			Edition: defaultBackendSystemEdition,
		},
		GSEDeployConfs: []GSEDeployConf{
			{
				Generation:    defaultGseDeployConfLinuxGeneration,
				OsType:        defaultGseDeployConfLinuxOsType,
				BaseWorkDir:   defaultGseDeployConfLinuxBaseWorkDir,
				BaseDeployDir: defaultGseDeployConfLinuxBaseDeployDir,
			},
			{
				Generation:    defaultGseDeployConfWindowsGeneration,
				OsType:        defaultGseDeployConfWindowsOsType,
				BaseWorkDir:   defaultGseDeployConfWindowsBaseDeployDir,
				BaseDeployDir: defaultGseDeployConfWindowsBaseWorkDir,
			},
		},
		InstallerFileGroup: FileGroup{
			FullPath: defaultInstallerFileGroup,
		},
	}
}

// LoadFromFile loads config from file.
func (svc *BackendService) LoadFromFile(path string) error {
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
func (svc *BackendService) Validate() error {
	if err := svc.Workflow.Validate(); err != nil {
		return err
	}

	return nil
}

// GSEDeployConf defines the deployment configuration for gse node.
type GSEDeployConf struct {
	Generation    int64           `yaml:"generation" usage:"generation of deploy"`
	OsType        string          `yaml:"osType" usage:"os type"`
	BaseWorkDir   string          `yaml:"baseWorkDir" usage:"base work dir"`
	BaseDeployDir string          `yaml:"baseDeployDir" usage:"base deploy dir"`
	Custom        GSEDeployCustom `yaml:"custom" usage:"custom deploy conf"`
}

// GSEDeployCustom defines the custom deployment configuration for gse node.
type GSEDeployCustom struct {
	LogDir             string `yaml:"logDir" usage:"log dir"`
	HostIDPath         string `yaml:"hostIDPath" usage:"host id path"`
	AgentDataIPCPath   string `yaml:"agentDataIPCPath" usage:"data ipc path"`
	AgentPluginIPCPath string `yaml:"agentPluginIPCPath" usage:"plugin ipc path"`
	EnvironDir         string `yaml:"environDir" usage:"environ dir"`
}
