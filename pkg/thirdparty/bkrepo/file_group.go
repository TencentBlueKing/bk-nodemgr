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

package bkrepo

import (
	"io"
	"path"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
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
func (group *FileGroup) SubGroups(nCtx contextx.IContext) ([]fileiface.FileGroup, error) {
	return group.handler.listGroups(nCtx, group.info.FullPath)
}

// GetFile return the file of the file group.
func (group *FileGroup) GetFile(nCtx contextx.IContext, name string) (fileiface.File, error) {
	return group.handler.GetFile(nCtx, path.Join(group.info.FullPath, name))
}

// AllFiles return all files of the file group.
func (group *FileGroup) AllFiles(nCtx contextx.IContext) ([]fileiface.File, error) {
	return group.handler.listFiles(nCtx, group.info.FullPath)
}

// Store store a file to the file group.
func (group *FileGroup) Store(
	nCtx contextx.IContext,
	info fileiface.FileInfo,
	reader io.ReadCloser,
	overwrite bool,
) error {

	return group.handler.storeFile(nCtx, group.info.FullPath, info, reader, overwrite)
}

// AbsDirs the func will return the abs dirs of file group.
func (group *FileGroup) AbsDirs() []string {
	return group.absDirs
}
