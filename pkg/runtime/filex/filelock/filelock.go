/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package filelock ...
package filelock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileLock this is a file lock.
type FileLock struct {
	lockPath string
	file     *os.File
}

// NewFileLock new a file lock.
func NewFileLock(path string) *FileLock {
	lockFilePath := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".lock")
	return &FileLock{
		lockPath: lockFilePath,
	}
}

// TryLock try lock.
func (l *FileLock) TryLock() error {
	file, err := os.OpenFile(l.lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("lock already exists")
		}
		return fmt.Errorf("failed to create lock: %w", err)
	}
	l.file = file
	return nil
}

// Unlock un lock.
func (l *FileLock) Unlock() error {
	if l.file != nil {
		l.file.Close()
		return os.Remove(l.lockPath)
	}
	return nil
}
