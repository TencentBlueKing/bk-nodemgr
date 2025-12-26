/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"
)

// Validate check request body.
func (x *DownloadAgentReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetCpuArch() == "" {
		return errors.New("cpu_arch is required")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *DownloadAgentReq) AutoConvert() {
}

// Validate check request body.
func (x *DownloadProxyReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetCpuArch() == "" {
		return errors.New("cpu_arch is required")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *DownloadProxyReq) AutoConvert() {
}

// Validate check request body.
func (x *DownloadPluginReq) Validate() error {
	if x.GetPluginPkgName() == "" {
		return errors.New("plugin_pkg_name is required")
	}

	if x.GetCpuArch() == "" {
		return errors.New("cpu_arch is required")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *DownloadPluginReq) AutoConvert() {
}

// Validate check request body.
func (x *DownloadInstallerReq) Validate() error {
	if x.GetCpuArch() == "" {
		return errors.New("cpu_arch is required")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *DownloadInstallerReq) AutoConvert() {
}
