/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tmp provides the ability to control the temporary directory used by the runtime.
package tmp

import (
	"fmt"
	"os"
)

// Clean cleans the file and removes it from the filesystem.
func Clean() error {
	tmpDir, err := GetTmpDir()
	if err != nil {
		return fmt.Errorf("failed to get tmp dir: %w", err)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to remove tmp dir: %w", err)
	}

	return nil
}
