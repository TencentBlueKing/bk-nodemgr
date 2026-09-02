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

// Package config ...
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"gopkg.in/yaml.v2"
)

const (
	// backend service config default values.
	defaultBackendRunMode                        = RunModeRelease
	defaultBackendTenantMode                     = tenant.ModeSingle
	defaultBackendInfoBindIP                     = "127.0.0.1"
	defaultBackendInfoBindIPV6                   = "::1"
	defaultBackendInfoPort                       = 28100
	defaultBackendInfoGracefulShutdownTimeoutSec = 60
	defaultBackendInfoTraceServiceName           = "backend-server-info"
	defaultBackendInfoAuthIdentity               = AuthIdentityNone

	defaultBackendAdminBindIP                     = "127.0.0.1"
	defaultBackendAdminBindIPV6                   = "::1"
	defaultBackendAdminPort                       = 28101
	defaultBackendAdminGracefulShutdownTimeoutSec = 60
	defaultBackendAdminTraceServiceName           = "backend-server-admin"
	defaultBackendAdminAuthIdentity               = AuthIdentityRestServer
	defaultBackendAdminJWTCryptoType              = JWTCryptoTypeSymmetric

	defaultBackendBasicBindIP                     = "127.0.0.1"
	defaultBackendBasicBindIPV6                   = "::1"
	defaultBackendBasicPort                       = 28102
	defaultBackendBasicGracefulShutdownTimeoutSec = 60
	defaultBackendBasicTraceServiceName           = "backend-server-basic"
	defaultBackendBasicAuthIdentity               = AuthIdentityAPIGW
	defaultBackendBasicJWTCryptoType              = JWTCryptoTypeAsymmetric

	defaultBackendCallbackBindIP                     = "127.0.0.1"
	defaultBackendCallbackBindIPV6                   = "::1"
	defaultBackendCallbackPort                       = 28103
	defaultBackendCallbackGracefulShutdownTimeoutSec = 60
	defaultBackendCallbackTraceServiceName           = "backend-server-callback"
	defaultBackendCallbackAuthIdentity               = AuthIdentityNone

	defaultBackendProxyBindIP                     = "127.0.0.1"
	defaultBackendProxyBindIPV6                   = "::1"
	defaultBackendProxyPort                       = 28104
	defaultBackendProxyGracefulShutdownTimeoutSec = 60
	defaultBackendProxyTraceServiceName           = "backend-server-proxy"
	defaultBackendProxyAuthIdentity               = AuthIdentityNone

	defaultBackendLogDir       = "/bk-nodemgr/log/"
	defaultBackendLogMaxNum    = 10
	defaultBackendLogMaxSizeMB = 200
	defaultBackendLogLevel     = "INFO"

	defaultBackendEncryptKey = "1234567890abcdef"

	defaultBackendSystemEnv     = "dev"
	defaultBackendSystemEdition = "ce"

	defaultBackendAdvertiseIPv4 = "127.0.0.1"
	defaultBackendAdvertiseIPv6 = "::1"

	defaultBackendTracingExporterType    = "stdout"
	defaultBackendGlobalTraceServiceName = "backend"

	defaultBackendRedisTraceServiceName = "bk-nodemgr_redis"

	defaultBackendMongoDBAppName          = "bk_nodemgr_backend"
	defaultBackendMongoDBTraceServiceName = "bk_nodemgr_mongo"

	defaultBackendWorkflowTraceServiceName            = "workflow"
	defaultBackendWorkflowWorkerNum                   = 10
	defaultBackendWorkflowGracefulShutdownTimeoutSecs = 60

	defaultInstallerFileGroup = "/bk-nodemgr/file/tools"

	defaultGseDeployConfLinuxGeneration    = 2
	defaultGseDeployConfLinuxOsType        = string(criteria.OSLinux)
	defaultGseDeployConfLinuxBaseDeployDir = "/usr/local/"
	defaultGseDeployConfLinuxBaseWorkDir   = "/tmp/bknm/"

	defaultGseDeployConfWindowsGeneration    = 2
	defaultGseDeployConfWindowsOsType        = string(criteria.OSWindows)
	defaultGseDeployConfWindowsBaseDeployDir = `c:\`
	defaultGseDeployConfWindowsBaseWorkDir   = `c:\tmp\bknm\`

	defaultBackendFileCacheDir             = "/bk-nodemgr/filecache"
	defaultBackendFileCacheExpirationHours = 72
	defaultBackendFileCacheGCIntervalHours = 1
	defaultBackendFileCacheRestoreOnStart  = false

	defaultBackendFileTraceServiceName        = "backend-client-file"
	defaultBackendCMDBTraceServiceName        = "backend-client-cmdb"
	defaultBackendGSETraceServiceName         = "backend-client-gse"
	defaultBackendUserManagerTraceServiceName = "backend-client-usermanager"
	defaultBackendIEGTJJTraceServiceName      = "backend-client-iegtjj"
	defaultBackendIAMV3TraceServiceName       = "backend-client-iam-v3"
	defaultBackendMonitorTraceServiceName     = "backend-client-monitor"

	defaultBackendAgentBaseAlarmEventDataID = 1000
	defaultBackendTaskProcEventDataID       = 1100008

	defaultBackendNetworkUnitDefaultDirectUnitEnabled = false
	defaultBackendNetworkUnitDefaultDirectUnitName    = "default"
)

// BackendService the config of backend service.
type BackendService struct {
	RunMode            RunMode          `yaml:"runMode" usage:"run mode of service"`
	TenantMode         tenant.Mode      `yaml:"tenantMode" usage:"tenant mode of service"`
	CMDB               CMDB             `yaml:"cmdb" usage:"cmdb config of backend service"`
	File               File             `yaml:"file" usage:"file config of backend service"`
	GSE                GSE              `yaml:"gse" usage:"gse config of backend service"`
	UserManager        UserManager      `yaml:"userManager" usage:"user manager config of backend service"`
	IAMV3              IAMV3            `yaml:"iamV3" usage:"IAM v3 gateway config"`
	Monitor            Monitor          `yaml:"monitor" usage:"monitor gateway config"`
	NodeEventDataID    NodeEventDataID  `yaml:"nodeEventDataID" usage:"node event data-id config"`
	Workflow           Workflow         `yaml:"workflow" usage:"workflow config of backend service"`
	InfoServer         HTTPServer       `yaml:"infoServer" usage:"info server config of backend service"`
	AdminServer        HTTPServer       `yaml:"adminServer" usage:"admin server config of backend service"`
	BasicServer        HTTPServer       `yaml:"basicServer" usage:"basic server config of backend service"`
	CallbackServer     CallbackServer   `yaml:"callbackServer" usage:"callback server config of backend service"`
	ProxyServer        ProxyServer      `yaml:"proxyServer" usage:"proxy server config of backend service"`
	Etcd               Etcd             `yaml:"etcd" usage:"etcd config of backend service"`
	Redis              Redis            `yaml:"redis" usage:"redis config of backend service"`
	MongoDB            MongoDB          `yaml:"mongodb" usage:"mongodb config of backend service"`
	Log                Log              `yaml:"log" usage:"log config of backend service"`
	System             System           `yaml:"system" usage:"system config of backend service"`
	EncryptKey         string           `yaml:"encryptKey" usage:"encrypt key of backend service"`
	GSEDeployConfs     []GSEDeployConf  `yaml:"gseDeployConfs" usage:"gse deploy config of backend service"`
	InstallerFileGroup FileGroup        `yaml:"installerFileGroup" usage:"tools file group config of backend service"`
	FileCache          BackendFileCache `yaml:"fileCache" usage:"local file cache config of backend service"`
	CreditVault        CreditVault      `yaml:"creditVault" usage:"credit vault config of backend service"`
	Access             Access           `yaml:"access" usage:"access config of backend service"`
	NetworkUnit        NetworkUnit      `yaml:"networkUnit" usage:"network unit config of backend service"`
	Tracing            Tracing          `yaml:"tracing" usage:"tracing config of backend service"`
	Profiling          Profiling        `yaml:"profiling" usage:"profiling config of backend service"`
}

// BackendFileCache configures the local artifact file cache used by backend SSH install flows.
type BackendFileCache struct {
	// Dir is the absolute path for the local file cache directory.
	// Defaults to /bk-nodemgr/filecache when empty.
	Dir string `yaml:"dir" usage:"absolute path for the local file cache directory"`

	// ExpirationHours is the number of hours after which an unused cache entry is evicted by GC.
	// Defaults to 72 hours.
	ExpirationHours int `yaml:"expirationHours" usage:"number of hours an unused cache entry is retained"`

	// GCIntervalHours is the number of hours between garbage collection runs.
	// Defaults to 1 hour.
	GCIntervalHours int `yaml:"gcIntervalHours" usage:"number of hours between garbage collection runs"`

	// RestoreOnStart controls whether existing cache entries on disk are loaded into
	// the in-memory index at startup. Defaults to false (cache starts empty).
	RestoreOnStart bool `yaml:"restoreOnStart" usage:"restore cache entries from disk on startup"`
}

// NewBackendService generates a new BackendService with default values.
// nolint: funlen, fnsize
// NOCC: golint/fnsize(default backend configuration is clearer as one initializer).
func NewBackendService() *BackendService {
	return &BackendService{
		RunMode:    defaultBackendRunMode,
		TenantMode: defaultBackendTenantMode,
		CMDB: CMDB{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultBackendCMDBTraceServiceName,
				},
			},
		},
		File: File{
			TraceService: TraceService{
				TraceServiceName: defaultBackendFileTraceServiceName,
			},
		},
		GSE: GSE{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultBackendGSETraceServiceName,
				},
			},
		},
		UserManager: UserManager{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultBackendUserManagerTraceServiceName,
				},
			},
		},
		IAMV3: IAMV3{
			SystemID:     defaultIAMV3SystemID,
			CMDBSystemID: defaultIAMV3CMDBSystemID,
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultBackendIAMV3TraceServiceName,
				},
			},
		},
		Monitor: Monitor{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultBackendMonitorTraceServiceName,
				},
			},
		},
		NodeEventDataID: NodeEventDataID{
			Default: NodeEventDataIDConf{
				AgentBaseAlarmEventDataID: defaultBackendAgentBaseAlarmEventDataID,
				TaskProcEventDataID:       defaultBackendTaskProcEventDataID,
			},
		},
		Workflow: Workflow{
			TraceService: TraceService{
				TraceServiceName: defaultBackendWorkflowTraceServiceName,
			},
			WorkerNum:                      defaultBackendWorkflowWorkerNum,
			GracefulShutdownTimeoutSeconds: defaultBackendWorkflowGracefulShutdownTimeoutSecs,
		},
		InfoServer: HTTPServer{
			BindIP:                     defaultBackendInfoBindIP,
			BindIPV6:                   defaultBackendInfoBindIPV6,
			Port:                       defaultBackendInfoPort,
			GracefulShutdownTimeoutSec: defaultBackendInfoGracefulShutdownTimeoutSec,
			TraceService: TraceService{
				TraceServiceName: defaultBackendInfoTraceServiceName,
			},
			AuthIdentity:  defaultBackendInfoAuthIdentity,
			AdvertiseIPV4: defaultBackendAdvertiseIPv4,
			AdvertiseIPV6: defaultBackendAdvertiseIPv6,
		},
		AdminServer: HTTPServer{
			BindIP:                     defaultBackendAdminBindIP,
			BindIPV6:                   defaultBackendAdminBindIPV6,
			Port:                       defaultBackendAdminPort,
			GracefulShutdownTimeoutSec: defaultBackendAdminGracefulShutdownTimeoutSec,
			TraceService: TraceService{
				TraceServiceName: defaultBackendAdminTraceServiceName,
			},
			AuthIdentity:    defaultBackendAdminAuthIdentity,
			AdvertiseIPV4:   defaultBackendAdvertiseIPv4,
			AdvertiseIPV6:   defaultBackendAdvertiseIPv6,
			JWTServerConfig: JWTServerConfig{CryptoType: defaultBackendAdminJWTCryptoType},
		},
		BasicServer: HTTPServer{
			BindIP:                     defaultBackendBasicBindIP,
			BindIPV6:                   defaultBackendBasicBindIPV6,
			Port:                       defaultBackendBasicPort,
			GracefulShutdownTimeoutSec: defaultBackendBasicGracefulShutdownTimeoutSec,
			TraceService: TraceService{
				TraceServiceName: defaultBackendBasicTraceServiceName,
			},
			AuthIdentity:    defaultBackendBasicAuthIdentity,
			AdvertiseIPV4:   defaultBackendAdvertiseIPv4,
			AdvertiseIPV6:   defaultBackendAdvertiseIPv6,
			JWTServerConfig: JWTServerConfig{CryptoType: defaultBackendBasicJWTCryptoType},
		},
		CallbackServer: CallbackServer{
			HTTPServer: HTTPServer{
				BindIP:                     defaultBackendCallbackBindIP,
				BindIPV6:                   defaultBackendCallbackBindIPV6,
				Port:                       defaultBackendCallbackPort,
				GracefulShutdownTimeoutSec: defaultBackendCallbackGracefulShutdownTimeoutSec,
				TraceService: TraceService{
					TraceServiceName: defaultBackendCallbackTraceServiceName,
				},
				AuthIdentity:  defaultBackendCallbackAuthIdentity,
				AdvertiseIPV4: defaultBackendAdvertiseIPv4,
				AdvertiseIPV6: defaultBackendAdvertiseIPv6,
			},
		},
		ProxyServer: ProxyServer{
			HTTPServer: HTTPServer{
				BindIP:                     defaultBackendProxyBindIP,
				BindIPV6:                   defaultBackendProxyBindIPV6,
				Port:                       defaultBackendProxyPort,
				GracefulShutdownTimeoutSec: defaultBackendProxyGracefulShutdownTimeoutSec,
				TraceService: TraceService{
					TraceServiceName: defaultBackendProxyTraceServiceName,
				},
				AuthIdentity:  defaultBackendProxyAuthIdentity,
				AdvertiseIPV4: defaultBackendAdvertiseIPv4,
				AdvertiseIPV6: defaultBackendAdvertiseIPv6,
			},
		},
		Etcd: Etcd{},
		Redis: Redis{
			TraceService: TraceService{
				TraceServiceName: defaultBackendRedisTraceServiceName,
			},
		},
		MongoDB: MongoDB{
			AppName: defaultBackendMongoDBAppName,
			TraceService: TraceService{
				TraceServiceName: defaultBackendMongoDBTraceServiceName,
			},
		},
		Log: Log{
			Dir:       defaultBackendLogDir,
			MaxSizeMB: defaultBackendLogMaxSizeMB,
			MaxNum:    defaultBackendLogMaxNum,
			Level:     defaultBackendLogLevel,
		},
		System: System{
			Env:     defaultBackendSystemEnv,
			Edition: defaultBackendSystemEdition,
		},
		EncryptKey: defaultBackendEncryptKey,
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
		FileCache: BackendFileCache{
			Dir:             defaultBackendFileCacheDir,
			ExpirationHours: defaultBackendFileCacheExpirationHours,
			GCIntervalHours: defaultBackendFileCacheGCIntervalHours,
			RestoreOnStart:  defaultBackendFileCacheRestoreOnStart,
		},
		CreditVault: CreditVault{
			HostCreditVault: HostCreditVault{
				IEGTJJ: IEGTJJ{
					APIGatewayClient: APIGatewayClient{
						TraceService: TraceService{
							TraceServiceName: defaultBackendIEGTJJTraceServiceName,
						},
					},
				},
			},
		},
		Access: Access{
			VirtualUser: defaultAccessVirtualUser,
		},
		NetworkUnit: NetworkUnit{
			DefaultDirectUnit: DefaultDirectUnit{
				Enabled: defaultBackendNetworkUnitDefaultDirectUnitEnabled,
				Name:    defaultBackendNetworkUnitDefaultDirectUnitName,
			},
		},
		Tracing: Tracing{
			ExporterType: defaultBackendTracingExporterType,
			GlobalService: TraceService{
				TraceServiceName: defaultBackendGlobalTraceServiceName,
				TraceSampleRate:  0,
			},
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
// nolint: funlen, gocognit, gocyclo, cyclop
func (svc *BackendService) Validate() error {
	if err := svc.RunMode.Validate(); err != nil {
		return fmt.Errorf("failed to validate run mode config: %w", err)
	}

	if err := svc.TenantMode.Validate(); err != nil {
		return fmt.Errorf("failed to validate tenant mode config: %w", err)
	}

	if err := svc.Workflow.Validate(); err != nil {
		return fmt.Errorf("failed to validate workflow service config: %w", err)
	}

	if err := svc.CreditVault.Validate(); err != nil {
		return fmt.Errorf("failed to validate credit vault service config: %w", err)
	}

	if err := svc.InfoServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate info service config: %w", err)
	}

	if err := svc.AdminServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate admin service config: %w", err)
	}

	if err := svc.BasicServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate basic service config: %w", err)
	}

	if err := svc.CallbackServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate callback service config: %w", err)
	}

	if err := svc.ProxyServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate proxy service config: %w", err)
	}

	if err := svc.Access.Validate(); err != nil {
		return fmt.Errorf("failed to validate access service config: %w", err)
	}

	if err := svc.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if err := svc.MongoDB.Validate(); err != nil {
		return fmt.Errorf("failed to validate mongodb config: %w", err)
	}

	if err := svc.Redis.Validate(); err != nil {
		return fmt.Errorf("failed to validate redis config: %w", err)
	}

	if err := svc.Etcd.Validate(); err != nil {
		return fmt.Errorf("failed to validate etcd config: %w", err)
	}

	if err := svc.System.Validate(); err != nil {
		return fmt.Errorf("failed to validate system config: %w", err)
	}

	if err := svc.File.Validate(); err != nil {
		return fmt.Errorf("failed to validate file config: %w", err)
	}

	if err := svc.CMDB.Validate(); err != nil {
		return fmt.Errorf("failed to validate cmdb config: %w", err)
	}

	if err := svc.GSE.Validate(); err != nil {
		return fmt.Errorf("failed to validate gse service config: %w", err)
	}

	if err := svc.UserManager.Validate(); err != nil {
		return fmt.Errorf("failed to validate user manager config: %w", err)
	}

	if err := svc.IAMV3.Validate(); err != nil {
		return fmt.Errorf("failed to validate IAM v3 config: %w", err)
	}

	if err := svc.Monitor.Validate(); err != nil {
		return fmt.Errorf("failed to validate monitor config: %w", err)
	}

	if err := svc.NodeEventDataID.Validate(); err != nil {
		return fmt.Errorf("failed to validate node event data-id config: %w", err)
	}

	if svc.EncryptKey == "" {
		return fmt.Errorf("failed to validate encrypt key config: encrypt key is empty")
	}

	for _, conf := range svc.GSEDeployConfs {
		if err := conf.Validate(); err != nil {
			return fmt.Errorf("failed to validate gse deploy conf: %w", err)
		}
	}

	if err := svc.InstallerFileGroup.Validate(); err != nil {
		return fmt.Errorf("failed to validate installer file group config: %w", err)
	}

	if err := svc.Tracing.Validate(); err != nil {
		return fmt.Errorf("failed to validate tracing config: %w", err)
	}

	if err := svc.Profiling.Validate(); err != nil {
		return fmt.Errorf("failed to validate profiling config: %w", err)
	}

	return nil
}

// NetworkUnit defines the network unit configuration.
type NetworkUnit struct {
	DefaultDirectUnit DefaultDirectUnit `yaml:"defaultDirectUnit" usage:"default direct network unit config"`
}

// DefaultDirectUnit defines the default direct network unit config.
type DefaultDirectUnit struct {
	Enabled          bool     `yaml:"enabled" usage:"enable auto-create default direct network unit"`
	Name             string   `yaml:"name" usage:"name of the default direct network unit"`
	ClusterEndpoints []string `yaml:"clusterEndpoints" usage:"cluster endpoints of the default direct network unit"`
	FileEndpoints    []string `yaml:"fileEndpoints" usage:"file endpoints of the default direct network unit"`
	DataEndpoints    []string `yaml:"dataEndpoints" usage:"data endpoints of the default direct network unit"`
}

// Monitor defines the monitor gateway configuration.
type Monitor struct {
	Enabled          bool `yaml:"enabled" usage:"enable monitor event data-id integration"`
	APIGatewayClient `yaml:",inline" usage:"api-gateway config of monitor"`
}

// Validate validates the monitor config.
func (conf Monitor) Validate() error {
	if !conf.Enabled {
		return nil
	}

	if err := conf.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("failed to validate monitor api gateway config: %w", err)
	}

	return nil
}

// NodeEventDataID defines node event data-id configuration.
type NodeEventDataID struct {
	Default NodeEventDataIDConf `yaml:"default" usage:"global default node event data-id config"`
}

// Validate validates the node event data-id config.
func (conf NodeEventDataID) Validate() error {
	if err := conf.Default.Validate(); err != nil {
		return fmt.Errorf("failed to validate default node event data-id config: %w", err)
	}

	return nil
}

// NodeEventDataIDConf defines a pair of node event data IDs.
type NodeEventDataIDConf struct {
	AgentBaseAlarmEventDataID int64 `yaml:"agentBaseAlarmEventDataID" usage:"agent base alarm event data-id"`
	TaskProcEventDataID       int64 `yaml:"taskProcEventDataID" usage:"task process event data-id"`
}

// Validate validates the node event data-id pair.
func (conf NodeEventDataIDConf) Validate() error {
	if conf.AgentBaseAlarmEventDataID <= 0 {
		return fmt.Errorf("agentBaseAlarmEventDataID must be positive, got %d", conf.AgentBaseAlarmEventDataID)
	}

	if conf.TaskProcEventDataID <= 0 {
		return fmt.Errorf("taskProcEventDataID must be positive, got %d", conf.TaskProcEventDataID)
	}

	return nil
}

// GSEDeployConf defines the deployment configuration for gse node.
type GSEDeployConf struct {
	Generation       int64                 `yaml:"generation" usage:"generation of deploy"`
	OsType           string                `yaml:"osType" usage:"os type"`
	BaseWorkDir      string                `yaml:"baseWorkDir" usage:"base work dir"`
	BaseDeployDir    string                `yaml:"baseDeployDir" usage:"base deploy dir"`
	Custom           GSEDeployCustom       `yaml:"custom" usage:"custom deploy conf"`
	PluginCustom     GSEDeployPluginCustom `yaml:"pluginCustom" usage:"plugin custom deploy conf"`
	ManualScriptPath string                `yaml:"manualScriptPath" usage:"manual script"`
}

// Validate validates the config.
func (conf GSEDeployConf) Validate() error {
	if err := types.Generation(conf.Generation).Validate(); err != nil {
		return fmt.Errorf("failed to validate generation config: %w", err)
	}

	if err := criteria.OSType(conf.OsType).Validate(); err != nil {
		return fmt.Errorf("failed to validate os type config: %w", err)
	}

	if conf.BaseDeployDir == "" {
		return fmt.Errorf("failed to validate base deploy dir config: base deploy dir is empty")
	}

	if err := conf.Custom.Validate(); err != nil {
		return fmt.Errorf("failed to validate custom deploy conf: %w", err)
	}

	if err := conf.PluginCustom.Validate(); err != nil {
		return fmt.Errorf("failed to validate plugin custom deploy conf: %w", err)
	}

	if _, err := os.Stat(conf.ManualScriptPath); err != nil {
		return fmt.Errorf("failed to validate manual script path config: %w", err)
	}

	return nil
}

// GSEDeployCustom defines the custom deployment configuration for gse node.
type GSEDeployCustom struct {
	LogDir            string `yaml:"logDir" usage:"log dir"`
	ExtraConfigDir    string `yaml:"extraConfigDir" usage:"extra config dir"`
	DataIPC           string `yaml:"dataIPC" usage:"data ipc, in linux is path, in windows is port"`
	PluginIPC         string `yaml:"pluginIPC" usage:"plugin ipc, in linux is path, in windows is port"`
	ProxyFileCacheDir string `yaml:"proxyFileCacheDir" usage:"proxy file cache dir, only for linux and proxy node"`
	ZoneID            string `yaml:"zoneID" usage:"zone id of gse deploy node"`
	CityID            string `yaml:"cityID" usage:"city id of gse deploy node"`
}

// Validate validates the config.
func (conf *GSEDeployCustom) Validate() error {
	return nil
}

// GSEDeployPluginCustom defines the custom deployment configuration for plugin.
type GSEDeployPluginCustom struct {
	LogDir          string                    `yaml:"logDir" usage:"log dir"`
	HostIDPath      string                    `yaml:"hostIDPath" usage:"host id path"`
	CommonConstants map[string]map[string]any `yaml:"commonConstants" usage:"common constants for plugin"`
}

// Validate validates the config.
func (conf GSEDeployPluginCustom) Validate() error {
	return nil
}

// Tracing defines the tracing configuration for nodemgr system.
type Tracing struct {
	InstanceID    string            `yaml:"instanceID" usage:"instance id of tracing system"`
	ExporterType  string            `yaml:"exporterType" usage:"exporter type of tracing system"`
	OTLPEndpoint  string            `yaml:"otlpEndpoint" usage:"otlp endpoint of tracing system"`
	OTLPInsecure  bool              `yaml:"otlpInsecure" usage:"otlp insecure of tracing system"`
	OTLPHeaders   map[string]string `yaml:"otlpHeaders" usage:"otlp headers of tracing system"`
	GlobalService TraceService      `yaml:"globalService" usage:"global fallback trace service config"`
}

// Validate validates the config.
func (conf Tracing) Validate() error {
	if conf.ExporterType == "" {
		return fmt.Errorf("exporter type is empty")
	}

	if err := conf.GlobalService.Validate(); err != nil {
		return fmt.Errorf("failed to validate global service trace config: %w", err)
	}

	return nil
}

// Validate validates the config.
func (conf IAMV3) Validate() error {
	// Skip validation if IAM v3 is disabled
	if !conf.Enable {
		return nil
	}

	if err := conf.APIGatewayClient.Validate(); err != nil {
		return fmt.Errorf("failed to validate IAM v3 client config: %w", err)
	}

	if conf.SystemID == "" {
		return fmt.Errorf("system ID is required for IAM v3")
	}

	if conf.CallbackPath == "" {
		return fmt.Errorf("callback path is required for IAM v3")
	}

	return nil
}
