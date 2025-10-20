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

// Validate check body.
func (x *PluginInstallReq) Validate() error {
	process := x.GetProcess()
	if len(process) == 0 {
		return errors.New("plugin can not be empty")
	}

	for idx := range process {
		if err := process[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *PluginInstallReq_Process) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id can not be zero")
	}

	if x.GetPluginId() == "" {
		return errors.New("plugin_id can not be empty")
	}

	if x.GetVersion() == "" {
		return errors.New("version can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PluginInstallReq) AutoConvert() {
	process := x.GetProcess()
	for idx := range process {
		process[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *PluginInstallReq_Process) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}
