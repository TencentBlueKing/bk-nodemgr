//go:build linux

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"path/filepath"
	"sync"
)

const baseNameData = "gse_data"

// nolint: gochecknoglobals
var gseDataPath = struct {
	sync.Once
	filePath string
}{}

// GetGseDataProxyPath get gse data path.
func GetGseDataProxyPath() string {
	gseDataPath.Do(func() {
		gseDataPath.filePath = filepath.Join(GetSetupDir(), "bin", baseNameData)
	})

	return gseDataPath.filePath
}

// nolint: gochecknoglobals
var gseFilePath = struct {
	sync.Once
	filePath string
}{}

const baseNameFile = "gse_file"

// GetGseFileProxyPath get gse file path.
func GetGseFileProxyPath() string {
	gseFilePath.Do(func() {
		gseFilePath.filePath = filepath.Join(GetSetupDir(), "bin", baseNameFile)
	})

	return gseFilePath.filePath
}

// nolint: gochecknoglobals
var gseDataProxyConf = struct {
	sync.Once
	filePath string
	fileName string
}{
	Once:     sync.Once{},
	filePath: "",
	fileName: "gse_data_proxy.conf",
}

// GetGseDataProxyConfPath get gse data proxy config file path.
func GetGseDataProxyConfPath() string {
	gseDataProxyConf.Do(func() {
		gseDataProxyConf.filePath = filepath.Join(GetGseConfDir(), gseDataProxyConf.fileName)
	})

	return gseDataProxyConf.filePath
}

// nolint: gochecknoglobals
var gseFileProxyConf = struct {
	sync.Once
	filePath string
	fileName string
}{
	Once:     sync.Once{},
	filePath: "",
	fileName: "gse_file_proxy.conf",
}

// GetGseFileProxyConfPath get gse file proxy config file path.
func GetGseFileProxyConfPath() string {
	gseFileProxyConf.Do(func() {
		gseFileProxyConf.filePath = filepath.Join(GetGseConfDir(), gseFileProxyConf.fileName)
	})

	return gseFileProxyConf.filePath
}
