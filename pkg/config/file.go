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
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"gopkg.in/yaml.v2"
)

const (
	// file service config default values.
	defaultFileRunMode      = RunModeRelease
	defaultFileTenantMode   = tenant.ModeSingle
	defaultFileHTTPBindIP   = "127.0.0.1"
	defaultFileHTTPPort     = 6000
	defaultFileAdminBindIP  = "127.0.0.1"
	defaultFileAdminPort    = 6001
	defaultFileLogDir       = "/bk-nodeman/log/"
	defaultFileLogMaxNum    = 10
	defaultFileLogMaxSizeMB = 200
	defaultFileLogLevel     = "INFO"
)

// NewFileService generates a new FileService with default values.
func NewFileService() *FileService {
	return &FileService{
		RunMode:    defaultFileRunMode,
		TenantMode: defaultFileTenantMode,
		HTTPServer: HTTPServer{
			BindIP: defaultFileHTTPBindIP,
			Port:   defaultFileHTTPPort,
		},
		AdminServer: HTTPServer{
			BindIP: defaultFileAdminBindIP,
			Port:   defaultFileAdminPort,
		},
		Log: Log{
			Dir:       defaultFileLogDir,
			MaxSizeMB: defaultFileLogMaxSizeMB,
			MaxNum:    defaultFileLogMaxNum,
			Level:     defaultFileLogLevel,
		},
	}
}

// FileService the config of file service.
type FileService struct {
	RunMode        RunMode     `yaml:"runMode" usage:"run mode of service"`
	TenantMode     tenant.Mode `yaml:"tenantMode" usage:"tenant mode of service"`
	InContainer    bool        `yaml:"inContainer" usage:"whether in container"`
	Etcd           Etcd        `yaml:"etcd" usage:"etcd config of file service"`
	GSE            GSE         `yaml:"gse" usage:"gse config of backend service"`
	HTTPServer     HTTPServer  `yaml:"httpServer" usage:"http server config of file service"`
	AdminServer    HTTPServer  `yaml:"adminServer" usage:"admin server config of file service"`
	TempFileGroup  FileGroup   `yaml:"tempFileGroup" usage:"temp file group config of file service"`
	LocalFileGroup FileGroup   `yaml:"localFileGroup" usage:"local file group config of file service"`
	Repo           Repo        `yaml:"repo" usage:"repo config of file service"`
	MongoDB        MongoDB     `yaml:"mongodb" usage:"mongodb config of file service"`
	Log            Log         `yaml:"log" usage:"log config of file service"`
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
	if err := svc.TempFileGroup.Validate(); err != nil {
		return err
	}

	if err := svc.LocalFileGroup.Validate(); err != nil {
		return err
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
