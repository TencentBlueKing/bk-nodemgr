/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package winapi

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// AddRegistryCurrentUserStartRun Adds a startup item for the current user.
func AddRegistryCurrentUserStartRun(cmdName, cmdStr string) error {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS,
	)
	if err != nil {
		return fmt.Errorf("open registry key failed, err: %w", err)
	}
	defer key.Close()

	err = key.SetStringValue(cmdName, cmdStr)
	if err != nil {
		return fmt.Errorf("set registry cmdName failed, err: %w", err)
	}

	return nil
}

// DelRegistryCurrentUserStartRun Deletes the startup item for the current user.
func DelRegistryCurrentUserStartRun(cmdName string) error {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS,
	)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("open registry key failed, err: %w", err)
	}
	defer key.Close()

	err = key.DeleteValue(cmdName)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("delete registry cmdName failed, err: %w", err)
	}

	return nil
}
