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

// Package filex provides shared file stream primitives.
package filex

import (
	"io"
	"sync"
)

// OnceReadCloser wraps a non-nil reader so Close runs once and returns the same error on later calls.
// Wrap at the resource boundary and share the returned stream; the owner still must defer Close.
func OnceReadCloser(reader io.ReadCloser) io.ReadCloser {
	return &readCloser{ReadCloser: reader, close: sync.OnceValue(reader.Close)}
}

type readCloser struct {
	io.ReadCloser
	close func() error
}

func (reader *readCloser) Close() error {
	return reader.close()
}

// OnceReadWriteCloser wraps a non-nil file so Close runs once and returns the same error on later calls.
// Wrap at the resource boundary and share the returned stream; the owner still must defer Close.
func OnceReadWriteCloser(file io.ReadWriteCloser) io.ReadWriteCloser {
	return &readWriteCloser{ReadWriteCloser: file, close: sync.OnceValue(file.Close)}
}

type readWriteCloser struct {
	io.ReadWriteCloser
	close func() error
}

func (file *readWriteCloser) Close() error {
	return file.close()
}
