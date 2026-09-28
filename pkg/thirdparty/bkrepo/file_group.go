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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/transfer"
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

// IsDir reports whether a node is a directory.
func (group *FileGroup) IsDir(nCtx contextx.IContext, relativePath string) (bool, error) {
	if nCtx == nil {
		return false, errInvalidContext
	}

	fullPath, err := group.directoryPath(relativePath)
	if err != nil {
		return false, err
	}

	response, err := group.handler.cli.QueryNodeInfo(nCtx, &QueryNodeInfoReq{Path: fullPath})
	if errors.Is(err, errNodeNotFound) {
		return false, fmt.Errorf("stat node failed, path(%s): %w", relativePath, errors.Join(fs.ErrNotExist, err))
	}
	if err != nil {
		return false, fmt.Errorf("stat node failed, path(%s): %w", relativePath, err)
	}

	return response.NodeInfo.Folder, nil
}

// GetSubGroup returns a directory relative to this file group.
func (group *FileGroup) GetSubGroup(nCtx contextx.IContext, relativePath string) (fileiface.FileGroup, error) {
	fullPath, err := group.directoryPath(relativePath)
	if err != nil {
		return nil, err
	}
	if nCtx == nil {
		return nil, errInvalidContext
	}

	subGroup, err := group.handler.GetFileGroup(nCtx, fullPath)
	if errors.Is(err, errNodeNotFound) {
		return nil, fmt.Errorf("get subgroup failed, path(%s): %w", relativePath, errors.Join(fs.ErrNotExist, err))
	}
	if err != nil {
		return nil, err
	}

	return subGroup, nil
}

// EnsureSubGroup creates a directory and its missing parents relative to this group.
func (group *FileGroup) EnsureSubGroup(nCtx contextx.IContext, relativePath string) (fileiface.FileGroup, error) {
	fullPath, err := group.directoryPath(relativePath)
	if err != nil {
		return nil, err
	}
	if nCtx == nil {
		return nil, errInvalidContext
	}

	if fullPath == "/" {
		return group.handler.EnsureFileGroup(nCtx, fullPath)
	}

	existing, err := group.handler.GetFileGroup(nCtx, fullPath)
	if !errors.Is(err, errNodeNotFound) {
		return existing, err
	}

	current := "/"
	var result fileiface.FileGroup
	for part := range strings.SplitSeq(strings.TrimPrefix(fullPath, "/"), "/") {
		current = path.Join(current, part)
		result, err = group.handler.EnsureFileGroup(nCtx, current)
		if err != nil {
			return nil, fmt.Errorf("ensure file group failed, path(%s): %w", current, err)
		}
	}

	return result, nil
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
func (group *FileGroup) Store(nCtx contextx.IContext, info fileiface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	return group.handler.storeFile(nCtx, group.info.FullPath, info, reader, overwrite)
}

// Copy copies a node to a supported file group.
func (group *FileGroup) Copy(nCtx contextx.IContext, srcPath string, destGroup fileiface.FileGroup, destPath string, overwrite bool) error {
	if group == nil {
		return errors.New("source file group cannot be nil")
	}
	destBKRepoGroup, ok := destGroup.(*FileGroup)
	if !ok {
		return transfer.CopyStream(nCtx, group, srcPath, destGroup, destPath, overwrite)
	}
	if destBKRepoGroup == nil {
		return errors.New("destination file group cannot be nil")
	}

	return group.handler.copyNode(nCtx, group, srcPath, destBKRepoGroup, destPath, overwrite)
}

// Remove deletes a node in this group.
func (group *FileGroup) Remove(nCtx contextx.IContext, path string) error {
	return group.handler.deleteNode(nCtx, group, path)
}

// AbsDirs the func will return the abs dirs of file group.
func (group *FileGroup) AbsDirs() []string {
	return group.absDirs
}

func (group *FileGroup) directoryPath(relativePath string) (string, error) {
	if group == nil {
		return "", errors.New("file group cannot be nil")
	}
	if strings.ContainsAny(relativePath, "\\\x00") {
		return "", fmt.Errorf("directory path must use slash separators, path(%s)", relativePath)
	}

	return resolveNodePath(group, relativePath, true)
}
