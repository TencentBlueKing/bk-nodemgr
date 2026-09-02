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
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"gopkg.in/yaml.v2"
)

const (
	exportServerHTTP  = "http"
	exportServerHTTPS = "https"

	// file service config default values.
	defaultFileRunMode                        = RunModeRelease
	defaultFileTenantMode                     = tenant.ModeSingle
	defaultFileInfoBindIP                     = "127.0.0.1"
	defaultFileInfoBindIPV6                   = "::1"
	defaultFileInfoPort                       = 28200
	defaultFileInfoGracefulShutdownTimeoutSec = 60
	defaultFileInfoTraceServiceName           = "file-server-info"
	defaultFileInfoIdentity                   = AuthIdentityNone

	defaultFileAdminBindIP                     = "127.0.0.1"
	defaultFileAdminBindIPV6                   = "::1"
	defaultFileAdminPort                       = 28201
	defaultFileAdminGracefulShutdownTimeoutSec = 60
	defaultFileAdminTraceServiceName           = "file-server-admin"
	defaultFileAdminIdentity                   = AuthIdentityRestServer

	defaultFileBasicBindIP                     = "127.0.0.1"
	defaultFileBasicBindIPV6                   = "::1"
	defaultFileBasicPort                       = 28202
	defaultFileBasicGracefulShutdownTimeoutSec = 60
	defaultFileBasicTraceServiceName           = "file-server-basic"
	defaultFileBasicIdentity                   = AuthIdentityNone

	defaultFileDownloadBindIP                     = "127.0.0.1"
	defaultFileDownloadBindIPV6                   = "::1"
	defaultFileDownloadPort                       = 28203
	defaultFileDownloadGracefulShutdownTimeoutSec = 60
	defaultFileDownloadTraceServiceName           = "file-server-download"
	defaultFileDownloadIdentity                   = AuthIdentityNone

	defaultFileExportBindIP                     = "127.0.0.1"
	defaultFileExportBindIPV6                   = "::1"
	defaultFileExportPort                       = 28204
	defaultFileExportGracefulShutdownTimeoutSec = 60
	defaultFileExportTraceServiceName           = "file-server-export"
	defaultFileExportIdentity                   = AuthIdentityNone

	defaultFileLogDir        = "/bk-nodemgr/log/"
	defaultFileLogMaxNum     = 10
	defaultFileLogMaxSizeMB  = 200
	defaultFileLogLevel      = "INFO"
	defaultFileAdvertiseIPv4 = "127.0.0.1"
	defaultFileAdvertiseIPv6 = "::1"

	defaultFileMongoDBAppName          = "bk_nodemgr_file"
	defaultFileMongoDBTraceServiceName = "bk_nodemgr_mongo"

	defaultFileEtcdUsername = "root"
	defaultFileEtcdPassword = ""

	defaultFileWorkspaceGroupFullPath = "/bk-nodemgr/file/"

	defaultFileCacheExpirationHours = 72
	defaultFileCacheGCIntervalHours = 1
	defaultFileCacheRestoreOnStart  = false

	defaultFileTempExpirationHours = 24
	defaultFileTempGCIntervalHours = 1

	defaultFileTracingExporterType    = "stdout"
	defaultFileGlobalTraceServiceName = "file"

	defaultFileGSETraceServiceName  = "file-client-gse"
	defaultFileRepoTraceServiceName = "file-client-bkrepo"
)

func defaultFileEtcdEndpoints() []string {
	return []string{"127.0.0.1:2379"}
}

// NewFileService generates a new FileService with default values.
func NewFileService() *FileService {
	return &FileService{
		RunMode:    defaultFileRunMode,
		TenantMode: defaultFileTenantMode,
		Etcd: Etcd{
			Endpoints: defaultFileEtcdEndpoints(),
			Username:  defaultFileEtcdUsername,
			Password:  defaultFileEtcdPassword,
		},
		GSE: GSE{
			APIGatewayClient: APIGatewayClient{
				TraceService: TraceService{
					TraceServiceName: defaultFileGSETraceServiceName,
				},
			},
		},
		Repo: Repo{
			TraceService: TraceService{
				TraceServiceName: defaultFileRepoTraceServiceName,
			},
		},
		InfoServer: HTTPServer{
			BindIP:                     defaultFileInfoBindIP,
			BindIPV6:                   defaultFileInfoBindIPV6,
			Port:                       defaultFileInfoPort,
			GracefulShutdownTimeoutSec: defaultFileInfoGracefulShutdownTimeoutSec,
			AuthIdentity:               defaultFileInfoIdentity,
			AdvertiseIPV4:              defaultFileAdvertiseIPv4,
			AdvertiseIPV6:              defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileInfoTraceServiceName,
			},
		},
		AdminServer: HTTPServer{
			BindIP:                     defaultFileAdminBindIP,
			BindIPV6:                   defaultFileAdminBindIPV6,
			Port:                       defaultFileAdminPort,
			GracefulShutdownTimeoutSec: defaultFileAdminGracefulShutdownTimeoutSec,
			AuthIdentity:               defaultFileAdminIdentity,
			AdvertiseIPV4:              defaultFileAdvertiseIPv4,
			AdvertiseIPV6:              defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileAdminTraceServiceName,
			},
		},
		BasicServer: HTTPServer{
			BindIP:                     defaultFileBasicBindIP,
			BindIPV6:                   defaultFileBasicBindIPV6,
			Port:                       defaultFileBasicPort,
			GracefulShutdownTimeoutSec: defaultFileBasicGracefulShutdownTimeoutSec,
			AuthIdentity:               defaultFileBasicIdentity,
			AdvertiseIPV4:              defaultFileAdvertiseIPv4,
			AdvertiseIPV6:              defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileBasicTraceServiceName,
			},
		},
		DownloadServer: HTTPServer{
			BindIP:                     defaultFileDownloadBindIP,
			BindIPV6:                   defaultFileDownloadBindIPV6,
			Port:                       defaultFileDownloadPort,
			GracefulShutdownTimeoutSec: defaultFileDownloadGracefulShutdownTimeoutSec,
			AuthIdentity:               defaultFileDownloadIdentity,
			AdvertiseIPV4:              defaultFileAdvertiseIPv4,
			AdvertiseIPV6:              defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileDownloadTraceServiceName,
			},
		},
		ExportServer: ExportServer{
			HTTPServer: HTTPServer{
				BindIP:                     defaultFileExportBindIP,
				BindIPV6:                   defaultFileExportBindIPV6,
				Port:                       defaultFileExportPort,
				GracefulShutdownTimeoutSec: defaultFileExportGracefulShutdownTimeoutSec,
				AuthIdentity:               defaultFileExportIdentity,
				AdvertiseIPV4:              defaultFileAdvertiseIPv4,
				AdvertiseIPV6:              defaultFileAdvertiseIPv6,
				JWTServerConfig:            JWTServerConfig{CryptoType: JWTCryptoTypeSymmetric},
				TraceService: TraceService{
					TraceServiceName: defaultFileExportTraceServiceName,
				},
			},
		},
		WorkspaceFileGroup: FileGroup{
			FullPath: defaultFileWorkspaceGroupFullPath,
		},
		FileCache: FileServiceFileCache{
			ExpirationHours: defaultFileCacheExpirationHours,
			GCIntervalHours: defaultFileCacheGCIntervalHours,
			RestoreOnStart:  defaultFileCacheRestoreOnStart,
		},
		TempFile: FileServiceTempFile{
			ExpirationHours: defaultFileTempExpirationHours,
			GCIntervalHours: defaultFileTempGCIntervalHours,
		},
		Log: Log{
			Dir:       defaultFileLogDir,
			MaxSizeMB: defaultFileLogMaxSizeMB,
			MaxNum:    defaultFileLogMaxNum,
			Level:     defaultFileLogLevel,
		},
		Tracing: Tracing{
			ExporterType: defaultFileTracingExporterType,
			GlobalService: TraceService{
				TraceServiceName: defaultFileGlobalTraceServiceName,
				TraceSampleRate:  0,
			},
		},
		MongoDB: MongoDB{
			AppName: defaultFileMongoDBAppName,
			TraceService: TraceService{
				TraceServiceName: defaultFileMongoDBTraceServiceName,
			},
		},
	}
}

// FileServiceFileCache configures the local file cache used by the file service.
// The cache directory is derived from WorkspaceFileGroup.FullPath and is not separately configurable.
type FileServiceFileCache struct {
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

// FileServiceTempFile configures the cleanup behavior of the temp file directory used by the
// file manager. The temp directory is derived from WorkspaceFileGroup.FullPath and is not
// separately configurable.
type FileServiceTempFile struct {
	// ExpirationHours is the number of hours after which an idle temp file (no access since)
	// becomes eligible for deletion by the GC. Defaults to 24 hours.
	ExpirationHours int `yaml:"expirationHours" usage:"number of hours an unused temp file is retained"`

	// GCIntervalHours is the number of hours between temp file GC runs. Defaults to 1 hour.
	GCIntervalHours int `yaml:"gcIntervalHours" usage:"number of hours between temp file GC runs"`
}

// FileService the config of file service.
type FileService struct {
	RunMode            RunMode              `yaml:"runMode" usage:"run mode of service"`
	TenantMode         tenant.Mode          `yaml:"tenantMode" usage:"tenant mode of service"`
	Etcd               Etcd                 `yaml:"etcd" usage:"etcd config of file service"`
	GSE                GSE                  `yaml:"gse" usage:"gse config of file service"`
	InfoServer         HTTPServer           `yaml:"infoServer" usage:"info server config of file service"`
	AdminServer        HTTPServer           `yaml:"adminServer" usage:"admin server config of file service"`
	BasicServer        HTTPServer           `yaml:"basicServer" usage:"basic server config of file service"`
	DownloadServer     HTTPServer           `yaml:"downloadServer" usage:"download server config of file service"`
	ExportServer       ExportServer         `yaml:"exportServer" usage:"export server config of file service"`
	Downloader         Downloader           `yaml:"downloader" usage:"remote package downloader config with host allow/block lists"`
	WorkspaceFileGroup FileGroup            `yaml:"workspaceFileGroup" usage:"workspace file group config of file service"`
	FileCache          FileServiceFileCache `yaml:"fileCache" usage:"local file cache config of file service"`
	TempFile           FileServiceTempFile  `yaml:"tempFile" usage:"temp file cleanup config of file service"`
	MountHostDir       string               `yaml:"mountHostDir" usage:"mount host dir of file service"`
	Repo               Repo                 `yaml:"repo" usage:"repo config of file service"`
	MongoDB            MongoDB              `yaml:"mongodb" usage:"mongodb config of file service"`
	Log                Log                  `yaml:"log" usage:"log config of file service"`
	Tracing            Tracing              `yaml:"tracing" usage:"tracing config of file service"`
	Profiling          Profiling            `yaml:"profiling" usage:"profiling config of file service"`
}

// ExportServer configures the HTTP server and public address used by exported package URLs.
type ExportServer struct {
	HTTPServer `yaml:",inline"`
	Address    string `yaml:"address" usage:"public HTTP or HTTPS URL of export server"`
}

// ToURL returns the public export server address as a URL.
func (svc *ExportServer) ToURL() (*url.URL, error) {
	if svc.Address == "" {
		if svc.AdvertiseIPV4 == "" || svc.Port <= 0 {
			return nil, errors.New("failed to create export server public URL, advertise-ip and port are empty")
		}

		return &url.URL{
			Scheme: exportServerURLScheme(svc),
			Host:   net.JoinHostPort(svc.AdvertiseIPV4, strconv.Itoa(svc.Port)),
		}, nil
	}

	parsedURL, err := url.Parse(svc.Address)
	if err != nil {
		return nil, err
	}
	if (parsedURL.Scheme != exportServerHTTP && parsedURL.Scheme != exportServerHTTPS) || parsedURL.Host == "" {
		return nil, errors.New("failed to create export server public URL, address must be an http or https URL")
	}

	return parsedURL, nil
}

func exportServerURLScheme(svc *ExportServer) string {
	if svc.TLSConfig.CertFile != "" && svc.TLSConfig.KeyFile != "" {
		return exportServerHTTPS
	}

	return exportServerHTTP
}

// LoadFromFile loads config from file.
func (svc *FileService) LoadFromFile(path string) error {
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
func (svc *FileService) Validate() error {
	if err := svc.RunMode.Validate(); err != nil {
		return fmt.Errorf("failed to validate run mode config: %w", err)
	}

	if err := svc.TenantMode.Validate(); err != nil {
		return fmt.Errorf("failed to validate tenant mode config: %w", err)
	}

	if err := svc.WorkspaceFileGroup.Validate(); err != nil {
		return fmt.Errorf("failed to validate workspace file group config: %w", err)
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

	if err := svc.DownloadServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate node service config: %w", err)
	}

	if err := svc.ExportServer.Validate(); err != nil {
		return fmt.Errorf("failed to validate export server config: %w", err)
	}

	if err := svc.Log.Validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if err := svc.MongoDB.Validate(); err != nil {
		return fmt.Errorf("failed to validate mongodb config: %w", err)
	}

	if err := svc.Etcd.Validate(); err != nil {
		return fmt.Errorf("failed to validate etcd config: %w", err)
	}

	if err := svc.GSE.Validate(); err != nil {
		return fmt.Errorf("failed to validate gse config: %w", err)
	}

	if err := svc.Repo.Validate(); err != nil {
		return fmt.Errorf("failed to validate repo config: %w", err)
	}

	if err := svc.Tracing.Validate(); err != nil {
		return fmt.Errorf("failed to validate tracing config: %w", err)
	}

	if err := svc.Profiling.Validate(); err != nil {
		return fmt.Errorf("failed to validate profiling config: %w", err)
	}

	if err := svc.Downloader.Validate(); err != nil {
		return fmt.Errorf("failed to validate downloader config: %w", err)
	}

	return nil
}

// FileGroup file group config.
type FileGroup struct {
	FullPath string `yaml:"fullPath" usage:"full path of agent file group"`
}

// Validate validates the config.
func (group *FileGroup) Validate() error {
	if group.FullPath == "" {
		return errors.New("fullPath is empty")
	}

	return nil
}
