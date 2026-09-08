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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	defaultRelayPluginName = "bk-nodemgr-relay"

	defaultRelayInfoBindIP                     = "127.0.0.1"
	defaultRelayInfoBindIPV6                   = "::1"
	defaultRelayInfoPort                       = 28300
	defaultRelayInfoGracefulShutdownTimeoutSec = 60
	defaultRelayInfoTraceServiceName           = "relay-server-info"
	defaultRelayInfoIdentity                   = AuthIdentityNone

	defaultRelayAdminBindIP                     = "127.0.0.1"
	defaultRelayAdminBindIPV6                   = "::1"
	defaultRelayAdminPort                       = 28301
	defaultRelayAdminGracefulShutdownTimeoutSec = 60
	defaultRelayAdminTraceServiceName           = "relay-server-admin"
	defaultRelayAdminIdentity                   = AuthIdentityNone

	defaultRelayCallbackBindIP                     = "127.0.0.1"
	defaultRelayCallbackBindIPV6                   = "::1"
	defaultRelayCallbackPort                       = 28302
	defaultRelayCallbackGracefulShutdownTimeoutSec = 60
	defaultRelayCallbackTraceServiceName           = "relay-server-callback"
	defaultRelayCallbackIdentity                   = AuthIdentityNone

	defaultRelayDownloadBindIP                     = "127.0.0.1"
	defaultRelayDownloadBindIPV6                   = "::1"
	defaultRelayDownloadPort                       = 28303
	defaultRelayDownloadGracefulShutdownTimeoutSec = 60
	defaultRelayDownloadTraceServiceName           = "relay-server-download"
	defaultRelayDownloadIdentity                   = AuthIdentityNone

	defaultRelayAdvertiseIPv4 = "127.0.0.1"
	defaultRelayAdvertiseIPv6 = "::1"

	defaultRelayLogDir       = "/data/plugin-relay/log"
	defaultRelayLogMaxNum    = 10
	defaultRelayLogMaxSizeMB = 200
	defaultRelayLogLevel     = "INFO"

	defaultRelayWorkspaceGroupFullPath = "/data/plugin-relay"

	// A relay caches installation packages shared by every node it installs, so entries are
	// kept for a week to survive cross-week reinstalls, with a size cap as the hard backstop.
	defaultRelayFileCacheExpirationHours = 168
	defaultRelayFileCacheGCIntervalHours = 1
	defaultRelayFileCacheMaxSizeMB       = 10240
	defaultRelayFileCacheRestoreOnStart  = true

	defaultRelayTracingExporterType    = "stdout"
	defaultRelayGlobalTraceServiceName = "relay"
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

	FileCache RelayFileCache `yaml:"fileCache" usage:"installation package cache config of relay service"`

	Tracing   Tracing   `yaml:"tracing" usage:"tracing config of relay service"`
	Profiling Profiling `yaml:"profiling" usage:"profiling config of relay service"`
	Log       Log       `yaml:"log" usage:"log config of relay service"`
}

// RelayFileCache configures the local installation package cache used by the relay.
// The relay serves these packages to every node it installs, so entries are shared across
// installations and only reclaimed once unused.
type RelayFileCache struct {
	// ExpirationHours is the number of hours after which an unused cache entry is evicted by GC.
	// Defaults to 168 hours (7 days).
	ExpirationHours int `yaml:"expirationHours" usage:"number of hours an unused cache entry is retained"`

	// GCIntervalHours is the number of hours between garbage collection runs.
	// Defaults to 1 hour.
	GCIntervalHours int `yaml:"gcIntervalHours" usage:"number of hours between garbage collection runs"`

	// MaxSizeMB caps the total size of cached packages. When exceeded, the least recently used
	// entries are evicted until the cache fits again. Defaults to 10240 MB (10 GB).
	// Set to 0 to disable the cap and rely on ExpirationHours alone.
	MaxSizeMB int64 `yaml:"maxSizeMB" usage:"maximum total size in MB of cached packages, 0 means unlimited"`

	// RestoreOnStart controls whether existing cache entries on disk are loaded into the
	// in-memory index at startup. Defaults to true: without it a relay restart would force
	// every package to be transferred again.
	RestoreOnStart bool `yaml:"restoreOnStart" usage:"restore cache entries from disk on startup"`
}

// Validate validates the relay file cache config.
// Negative values are rejected rather than normalized: the cache treats a non-positive
// MaxSizeMB as "unlimited", so silently accepting a typo would remove the very disk
// safeguard this config exists to provide.
func (conf RelayFileCache) Validate() error {
	if conf.ExpirationHours < 0 {
		return fmt.Errorf("file cache expiration hours must not be negative, got(%d)", conf.ExpirationHours)
	}

	if conf.GCIntervalHours < 0 {
		return fmt.Errorf("file cache gc interval hours must not be negative, got(%d)", conf.GCIntervalHours)
	}

	if conf.MaxSizeMB < 0 {
		return fmt.Errorf("file cache max size mb must not be negative, got(%d)", conf.MaxSizeMB)
	}

	return nil
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
			BindIP:                     defaultRelayInfoBindIP,
			BindIPV6:                   defaultRelayInfoBindIPV6,
			Port:                       defaultRelayInfoPort,
			GracefulShutdownTimeoutSec: defaultRelayInfoGracefulShutdownTimeoutSec,
			TraceService: TraceService{
				TraceServiceName: defaultRelayInfoTraceServiceName,
			},
			AuthIdentity:  defaultRelayInfoIdentity,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		AdminServer: HTTPServer{
			BindIP:                     defaultRelayAdminBindIP,
			BindIPV6:                   defaultRelayAdminBindIPV6,
			Port:                       defaultRelayAdminPort,
			GracefulShutdownTimeoutSec: defaultRelayAdminGracefulShutdownTimeoutSec,
			TraceService: TraceService{
				TraceServiceName: defaultRelayAdminTraceServiceName,
			},
			AuthIdentity:  defaultRelayAdminIdentity,
			AdvertiseIPV4: defaultRelayAdvertiseIPv4,
			AdvertiseIPV6: defaultRelayAdvertiseIPv6,
		},
		CallbackServer: CallbackServer{
			HTTPServer{
				BindIP:                     defaultRelayCallbackBindIP,
				BindIPV6:                   defaultRelayCallbackBindIPV6,
				Port:                       defaultRelayCallbackPort,
				GracefulShutdownTimeoutSec: defaultRelayCallbackGracefulShutdownTimeoutSec,
				TraceService: TraceService{
					TraceServiceName: defaultRelayCallbackTraceServiceName,
				},
				AuthIdentity:  defaultRelayCallbackIdentity,
				AdvertiseIPV4: defaultRelayAdvertiseIPv4,
				AdvertiseIPV6: defaultRelayAdvertiseIPv6,
			}},
		DownloadServer: HTTPServer{
			BindIP:                     defaultRelayDownloadBindIP,
			BindIPV6:                   defaultRelayDownloadBindIPV6,
			Port:                       defaultRelayDownloadPort,
			GracefulShutdownTimeoutSec: defaultRelayDownloadGracefulShutdownTimeoutSec,
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
		FileCache: RelayFileCache{
			ExpirationHours: defaultRelayFileCacheExpirationHours,
			GCIntervalHours: defaultRelayFileCacheGCIntervalHours,
			MaxSizeMB:       defaultRelayFileCacheMaxSizeMB,
			RestoreOnStart:  defaultRelayFileCacheRestoreOnStart,
		},
		Tracing: Tracing{
			ExporterType: defaultRelayTracingExporterType,
			GlobalService: TraceService{
				TraceServiceName: defaultRelayGlobalTraceServiceName,
				TraceSampleRate:  0,
			},
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

	if err := svc.FileCache.Validate(); err != nil {
		return fmt.Errorf("failed to validate file cache config: %w", err)
	}

	if err := svc.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if err := svc.Tracing.Validate(); err != nil {
		return fmt.Errorf("failed to validate tracing config: %w", err)
	}

	if err := svc.Profiling.Validate(); err != nil {
		return fmt.Errorf("failed to validate profiling config: %w", err)
	}

	return nil
}
