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

const gseAgentConfFileName = "gse_agent.conf"
const gseFileProxyConfName = "gse_file_proxy.conf"
const gseDataProxyConfName = "gse_data_proxy.conf"

// nolint: gochecknoglobals
var gseAgentConf = struct {
	sync.Once
	filePath string
}{
	Once:     sync.Once{},
	filePath: "",
}

// GetGseAgentConfPath get gse agent config file path.
func GetGseAgentConfPath() string {
	gseAgentConf.Do(func() {
		gseAgentConf.filePath = filepath.Join(GetGseConfDir(), gseAgentConfFileName)
	})

	return gseAgentConf.filePath
}

// GetGseConfDir get gse agent config dir.
func GetGseConfDir() string {
	dir := filepath.Join(GetSetupDir(), "etc")

	return dir
}

// nolint: gochecknoglobals
var tmpAgentConf = struct {
	sync.Once
	filePath string
}{
	Once:     sync.Once{},
	filePath: "",
}

// GetTmpAgentConfPath get tmp gse agent config file path.
func GetTmpAgentConfPath() string {
	tmpAgentConf.Do(func() {
		tmpAgentConf.filePath = filepath.Join(GetTmpConfigDir(), gseAgentConfFileName)
	})

	return tmpAgentConf.filePath
}

// nolint: gochecknoglobals
var tmpDataProxyConf = struct {
	sync.Once
	filePath string
}{
	Once:     sync.Once{},
	filePath: "",
}

// GetTmpDataProxyConfPath get tmp gse data proxy config file path.
func GetTmpDataProxyConfPath() string {
	tmpDataProxyConf.Do(func() {
		tmpDataProxyConf.filePath = filepath.Join(GetTmpConfigDir(), gseDataProxyConfName)
	})

	return tmpDataProxyConf.filePath
}

// nolint: gochecknoglobals
var tmpFileProxyConf = struct {
	sync.Once
	filePath string
}{
	Once:     sync.Once{},
	filePath: "",
}

// GetTmpFileProxyConfPath get tmp gse file proxy config file path.
func GetTmpFileProxyConfPath() string {
	tmpFileProxyConf.Do(func() {
		tmpFileProxyConf.filePath = filepath.Join(GetTmpConfigDir(), gseFileProxyConfName)
	})

	return tmpFileProxyConf.filePath
}
