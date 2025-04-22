/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/iface"
	"io"
)

// File is the file object.
type File struct {
	infoFn    func() (iface.FileInfo, error)
	contentFn func() (io.ReadCloser, error)
	fullPath  string
}

// FileObject return the file object.
func (f File) FileObject() iface.FileObject {
	return iface.RemoteFile
}

// Content return the content of the file.
func (f File) Content() (io.ReadCloser, error) {
	return f.contentFn()
}

// Info return the info of the file.
func (f File) Info() (iface.FileInfo, error) {
	return f.infoFn()
}
