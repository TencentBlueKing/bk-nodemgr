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
	"sync"
	"time"

	"github.com/google/uuid"
)

// tmpDirInstance is the instance of tmpDir.
// nolint: gochecknoglobals
var tmpDirInstance struct {
	once sync.Once
	dir  string
}

// GetTmpDir returns the system's temporary directory path.
func GetTmpDir() (string, error) {
	tmpDirInstance.dir = filepath.Join(
		os.TempDir(),
		prefixName+"_"+time.Now().Format("20060102150405")+"_"+strings.ReplaceAll(uuid.New().String(), "-", ""))

	dirInfo, err := os.Stat(tmpDirInstance.dir)
	if err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(tmpDirInstance.dir, 0700); err != nil { // nolint:mnd
			return "", fmt.Errorf("failed to create tmp dir: %w", err)
		}

		return tmpDirInstance.dir, nil
	}

	if !dirInfo.IsDir() {
		return "", fmt.Errorf("tmp dir is not a dir, dir(%s) info(%+v)", tmpDirInstance.dir, dirInfo)
	}

	return tmpDirInstance.dir, nil
}
