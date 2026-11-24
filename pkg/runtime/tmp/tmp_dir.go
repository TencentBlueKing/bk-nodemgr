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

package tmp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GetTmpDir creates and returns a unique temporary directory.
func GetTmpDir() (string, error) {
	dir := filepath.Join(
		os.TempDir(),
		prefixName+"_"+time.Now().Format("20060102150405")+"_"+strings.ReplaceAll(uuid.New().String(), "-", ""))

	dirInfo, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil { // nolint:mnd
			return "", fmt.Errorf("failed to create tmp dir: %w", err)
		}

		return dir, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat tmp dir: %w", err)
	}

	if !dirInfo.IsDir() {
		return "", fmt.Errorf("tmp dir is not a dir, dir(%s) info(%+v)", dir, dirInfo)
	}

	return dir, nil
}
