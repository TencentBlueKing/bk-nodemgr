/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package file defines the file manager interface.
package file

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
)

type fileContent struct {
	path string
	fm   *fileManagerImpl
}

func (fc *fileContent) Content(ctx context.Context) (io.ReadCloser, error) {
	fc.fm.refLock.Lock()
	fc.fm.refs[fc.path]++
	fc.fm.refLock.Unlock()

	dirPath := filepath.Dir(fc.path)
	fileName := filepath.Base(fc.path)

	fileGroup, err := local.NewLocalDir(dirPath, fc.fm.logger)
	if err != nil {
		fc.fm.decrementRef(fc.path)
		return nil, fmt.Errorf("failed to create file group: %w", err)
	}

	file, err := fileGroup.GetFile(ctx, fileName)
	if err != nil {
		fc.fm.decrementRef(fc.path)
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	reader, err := file.Content(ctx)
	if err != nil {
		fc.fm.decrementRef(fc.path)
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return &refCountReader{
		ReadCloser: reader,
		onClose:    func() { fc.fm.decrementRef(fc.path) },
	}, nil
}

type refCountReader struct {
	io.ReadCloser
	onClose func()
	closed  bool
}

func (r *refCountReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	err := r.ReadCloser.Close()
	r.onClose()

	return err
}
