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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"gopkg.in/yaml.v2"
)

const (
	// file service config default values.
	defaultFileRunMode              = RunModeRelease
	defaultFileTenantMode           = tenant.ModeSingle
	defaultFileInfoBindIP           = "127.0.0.1"
	defaultFileInfoPort             = 28200
	defaultFileInfoTraceServiceName = "file-server-info"
	defaultFileInfoIdentity         = AuthIdentityNone

	defaultFileAdminBindIP           = "127.0.0.1"
	defaultFileAdminPort             = 28201
	defaultFileAdminTraceServiceName = "file-server-admin"
	defaultFileAdminIdentity         = AuthIdentityRestServer

	defaultFileBasicBindIP           = "127.0.0.1"
	defaultFileBasicPort             = 28202
	defaultFileBasicTraceServiceName = "file-server-basic"
	defaultFileBasicIdentity         = AuthIdentityNone

	defaultFileDownloadBindIP           = "127.0.0.1"
	defaultFileDownloadPort             = 28203
	defaultFileDownloadTraceServiceName = "file-server-download"
	defaultFileDownloadIdentity         = AuthIdentityNone

	defaultFileLogDir        = "/bk-nodeman/log/"
	defaultFileLogMaxNum     = 10
	defaultFileLogMaxSizeMB  = 200
	defaultFileLogLevel      = "INFO"
	defaultFileAdvertiseIPv4 = "127.0.0.1"
	defaultFileAdvertiseIPv6 = "::1"

	defaultFileMongoDBAppName          = "bk_nodemgr_file"
	defaultFileMongoDBTraceServiceName = "bk_nodemgr_mongo"

	defaultFileEtcdUsername = "root"
	defaultFileEtcdPassword = ""

	defaultFileWorkspaceGroupFullPath = "/bk-nodeman/file/"

	defaultFileTracingExporterType = "stdout"

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
			BindIP:        defaultFileInfoBindIP,
			Port:          defaultFileInfoPort,
			AuthIdentity:  defaultFileInfoIdentity,
			AdvertiseIPV4: defaultFileAdvertiseIPv4,
			AdvertiseIPV6: defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileInfoTraceServiceName,
			},
		},
		AdminServer: HTTPServer{
			BindIP:        defaultFileAdminBindIP,
			Port:          defaultFileAdminPort,
			AuthIdentity:  defaultFileAdminIdentity,
			AdvertiseIPV4: defaultFileAdvertiseIPv4,
			AdvertiseIPV6: defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileAdminTraceServiceName,
			},
		},
		BasicServer: HTTPServer{
			BindIP:        defaultFileBasicBindIP,
			Port:          defaultFileBasicPort,
			AuthIdentity:  defaultFileBasicIdentity,
			AdvertiseIPV4: defaultFileAdvertiseIPv4,
			AdvertiseIPV6: defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileBasicTraceServiceName,
			},
		},
		DownloadServer: HTTPServer{
			BindIP:        defaultFileDownloadBindIP,
			Port:          defaultFileDownloadPort,
			AuthIdentity:  defaultFileDownloadIdentity,
			AdvertiseIPV4: defaultFileAdvertiseIPv4,
			AdvertiseIPV6: defaultFileAdvertiseIPv6,
			TraceService: TraceService{
				TraceServiceName: defaultFileDownloadTraceServiceName,
			},
		},
		WorkspaceFileGroup: FileGroup{
			FullPath: defaultFileWorkspaceGroupFullPath,
		},
		Log: Log{
			Dir:       defaultFileLogDir,
			MaxSizeMB: defaultFileLogMaxSizeMB,
			MaxNum:    defaultFileLogMaxNum,
			Level:     defaultFileLogLevel,
		},
		Tracing: Tracing{
			ExporterType: defaultFileTracingExporterType,
		},
		MongoDB: MongoDB{
			AppName: defaultFileMongoDBAppName,
			TraceService: TraceService{
				TraceServiceName: defaultFileMongoDBTraceServiceName,
			},
		},
	}
}

// FileService the config of file service.
type FileService struct {
	RunMode            RunMode     `yaml:"runMode" usage:"run mode of service"`
	TenantMode         tenant.Mode `yaml:"tenantMode" usage:"tenant mode of service"`
	Etcd               Etcd        `yaml:"etcd" usage:"etcd config of file service"`
	GSE                GSE         `yaml:"gse" usage:"gse config of file service"`
	InfoServer         HTTPServer  `yaml:"infoServer" usage:"info server config of file service"`
	AdminServer        HTTPServer  `yaml:"adminServer" usage:"admin server config of file service"`
	BasicServer        HTTPServer  `yaml:"basicServer" usage:"basic server config of file service"`
	DownloadServer     HTTPServer  `yaml:"downloadServer" usage:"download server config of file service"`
	WorkspaceFileGroup FileGroup   `yaml:"workspaceFileGroup" usage:"workspace file group config of file service"`
	MountHostDir       string      `yaml:"mountHostDir" usage:"mount host dir of file service"`
	Repo               Repo        `yaml:"repo" usage:"repo config of file service"`
	MongoDB            MongoDB     `yaml:"mongodb" usage:"mongodb config of file service"`
	Log                Log         `yaml:"log" usage:"log config of file service"`
	Tracing            Tracing     `yaml:"tracing" usage:"tracing config of file service"`
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
