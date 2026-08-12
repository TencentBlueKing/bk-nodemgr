/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package downloader

import (
	"io"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

func newLimitedReadCloser(rc io.ReadCloser, limit int64) io.ReadCloser {
	return &limitedReadCloser{
		Reader: io.LimitReader(rc, limit),
		close:  rc.Close,
	}
}

type limitedReadCloser struct {
	io.Reader
	close func() error
	once  sync.Once
	err   error
}

func (r *limitedReadCloser) Close() error {
	r.once.Do(func() {
		r.err = r.close()
	})

	return r.err
}

func newTemporaryFile(file fileiface.File, path string, cleanup func() error) fileiface.File {
	return &temporaryFile{
		File:    file,
		path:    path,
		cleanup: cleanup,
	}
}

type temporaryFile struct {
	fileiface.File
	path    string
	cleanup func() error
	once    sync.Once
}

func (f *temporaryFile) Content(nCtx contextx.IContext) (io.ReadCloser, error) {
	rc, err := f.File.Content(nCtx)
	if err != nil {
		return nil, err
	}

	return &temporaryFileReader{ReadCloser: rc, path: f.path, cleanup: f.cleanup, once: &f.once}, nil
}

type temporaryFileReader struct {
	io.ReadCloser
	path    string
	cleanup func() error
	once    *sync.Once
}

func (r *temporaryFileReader) Close() error {
	err := r.ReadCloser.Close()
	r.once.Do(func() {
		logTempFileCleanupError(r.path, r.cleanup())
	})

	return err
}

func logTempFileCleanupError(path string, err error) {
	if err != nil {
		logger.G.Sys().WithErr(err).With("file", path).Error("failed to clean up temporary file")
	}
}
