/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package iface define the interface of filex.
package iface

import (
	"context"
	"io"
)

// FileGroup directory interface.
type FileGroup interface {
	// Name the name of file group.
	Name() string

	// SubGroups the sub groups of file group.
	SubGroups(ctx context.Context) ([]FileGroup, error)

	// GetFile get a file by name.
	GetFile(ctx context.Context, name string) (File, error)

	// AllFiles the all files of file group.
	AllFiles(ctx context.Context) ([]File, error)

	// Store the func will store a file into the file group.
	Store(ctx context.Context, info FileInfo, file io.ReadCloser, overwrite bool) error
}

// File file interface.
type File interface {
	// FileObject the file object.
	FileObject() FileObject

	FileContent

	// Info the info of file.
	Info(ctx context.Context) (FileInfo, error)
}

// FileInfo file info.
type FileInfo struct {
	// Name the name of file.
	Name string

	// Size the size of file.
	Size int64

	// Md5 the md5 of file.
	Md5 string

	// Description the description of file.
	Description string

	// ExtendFields the extend fields of file.
	ExtendFields map[string]string
}

// FileContent the content of file.
type FileContent interface {
	Content(ctx context.Context) (io.ReadCloser, error)
}

// FileObject define this file is local file or remote file.
type FileObject string

const (
	// LocalFile means this file is local file.
	LocalFile FileObject = "local"

	// RemoteFile means this file is get from remote.
	RemoteFile FileObject = "remote"
)
