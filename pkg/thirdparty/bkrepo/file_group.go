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
	"path"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// FileGroup defines the file group.
type FileGroup struct {
	info    NodeInfo
	absDirs []string

	handler *Handler
}

// Name return the name of the file group.
func (group *FileGroup) Name() string {
	return group.info.Name
}

// SubGroups return the sub groups of the file group.
func (group *FileGroup) SubGroups(ctx context.Context) ([]iface.FileGroup, error) {
	return group.handler.listGroups(ctx, group.info.FullPath)
}

// GetFile return the file of the file group.
func (group *FileGroup) GetFile(ctx context.Context, name string) (iface.File, error) {
	return group.handler.GetFile(ctx, path.Join(group.info.FullPath, name))
}

// AllFiles return all files of the file group.
func (group *FileGroup) AllFiles(ctx context.Context) ([]iface.File, error) {
	return group.handler.listFiles(ctx, group.info.FullPath)
}

// Store store a file to the file group.
func (group *FileGroup) Store(ctx context.Context, info iface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	return group.handler.storeFile(ctx, group.info.FullPath, info, reader, overwrite)
}

// AbsDirs the func will return the abs dirs of file group.
func (group *FileGroup) AbsDirs() []string {
	return group.absDirs
}
