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
	"context"
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/iface"
)

// File is the file object.
type File struct {
	info    NodeInfo
	absDirs []string

	handler *Handler
}

// FileObject return the file object.
func (f File) FileObject() iface.FileObject {
	return iface.RemoteFile
}

// Content return the content of the file.
func (f File) Content(ctx context.Context) (io.ReadCloser, error) {
	return f.handler.getFileContent(ctx, f.info.FullPath)
}

// Info return the info of the file.
func (f File) Info() iface.FileInfo {
	var desc string
	descValue, ok := f.info.Metadata["description"]
	if ok {
		desc, _ = descValue.(string)
	}

	return iface.FileInfo{
		Name:        f.info.Name,
		Size:        int64(f.info.Size),
		MD5:         f.info.Md5,
		Description: desc,
	}
}

// AbsDirs the func will return the abs dirs of file.
func (f File) AbsDirs() []string {
	return f.absDirs
}
