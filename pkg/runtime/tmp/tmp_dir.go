/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tmp

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// tmpDirInstance is the instance of tmpDir.
// nolint: gochecknoglobals
var tmpDirInstance struct {
	mu   sync.Mutex
	once sync.Once
	dir  string
	err  error
}

// GetTmpDir returns the system's temporary directory path.
func GetTmpDir() (string, error) {
	tmpDirInstance.mu.Lock()
	defer tmpDirInstance.mu.Unlock()

	tmpDirInstance.once.Do(func() {
		tmpDirInstance.dir, tmpDirInstance.err = createRootTmpDir()
	})
	if tmpDirInstance.err != nil {
		return "", tmpDirInstance.err
	}

	dirInfo, err := os.Stat(tmpDirInstance.dir)
	if err != nil {
		return "", fmt.Errorf("failed to stat tmp dir: %w", err)
	}

	if !dirInfo.IsDir() {
		return "", fmt.Errorf("tmp dir is not a dir, dir(%s) info(%+v)", tmpDirInstance.dir, dirInfo)
	}

	return tmpDirInstance.dir, nil
}

func resetTmpDir() {
	tmpDirInstance.once = sync.Once{}
	tmpDirInstance.dir = ""
	tmpDirInstance.err = nil
}

func newTempDir(name string) (string, error) {
	tmpDir, err := GetTmpDir()
	if err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp(tmpDir, filepath.Base(name)+"-*")
	if err != nil {
		return "", fmt.Errorf("failed to create tmp sub dir: %w", err)
	}

	return dir, nil
}
