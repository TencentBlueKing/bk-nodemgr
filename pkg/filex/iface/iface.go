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

// Package iface define the interface of filex.
package iface

import (
	"io"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// FileGroup directory interface.
// nolint:interfacebloat
type FileGroup interface {
	// Name the name of file group.
	// It is named since this file group is init.
	Name() string

	// AbsDirs the func will return the abs dirs of file group.
	AbsDirs() []string

	/**
	 * The following methods will cause actual action, such as DiskIO, NetworkIO, etc.
	 */

	// SubGroups the sub groups of file group.
	// It should lists all sub groups with actual action.
	SubGroups(nCtx contextx.IContext) ([]FileGroup, error)

	// IsDir reports whether a group-relative slash path identifies a directory.
	// "." identifies this group. Missing nodes must return an error matching fs.ErrNotExist.
	// Unsupported node types return an error rather than false.
	IsDir(nCtx contextx.IContext, relativePath string) (bool, error)

	// GetSubGroup returns an existing directory at a group-relative slash path.
	// "." queries this directory. Missing directories match fs.ErrNotExist; files are errors.
	GetSubGroup(nCtx contextx.IContext, relativePath string) (FileGroup, error)

	// EnsureSubGroup returns a directory, creating it and missing parents if necessary.
	// "." ensures this directory exists. A file at the requested path is an error.
	EnsureSubGroup(nCtx contextx.IContext, relativePath string) (FileGroup, error)

	// AllFiles the all files of file group.
	// It should lists all files with actual action.
	AllFiles(nCtx contextx.IContext) ([]File, error)

	// GetFile get a file by name.
	// It should get file with actual action.
	GetFile(nCtx contextx.IContext, name string) (File, error)

	// Store borrows file to store its content; it never closes the input stream.
	// The caller owns file and must close it after Store returns.
	Store(nCtx contextx.IContext, info FileInfo, file io.ReadCloser, overwrite bool) error

	// Copy copies a file or subgroup to another file group using the backend's copy policy.
	// Cross-backend paths are group-relative slash paths; native copies retain backend syntax.
	// A source must identify a node; "." is allowed as a destination group root.
	Copy(nCtx contextx.IContext, srcPath string, destGroup FileGroup, destPath string, overwrite bool) error

	// Remove deletes a file or subgroup from the file group.
	Remove(nCtx contextx.IContext, path string) error
}

// File file interface.
type File interface {
	// FileObject the file object.
	FileObject() FileObject

	// Info the info of file.
	// It is a cache read since the file is got from the file group.
	// Re-get the file if you want it fresh.
	Info() FileInfo

	// AbsDir the func will return the abs dirs of file.
	AbsDirs() []string

	FileContent
}

// FileInfo file info.
type FileInfo struct {
	// Name the name of file.
	Name string

	// Size the size of file.
	Size int64

	// Md5 the md5 of file.
	MD5 string

	// ModTime the mod time of file.
	ModTime time.Time

	// Description the description of file.
	Description string

	// ExtendFields the extend fields of file.
	ExtendFields map[string]string
}

// FileContent the content of file.
type FileContent interface {
	// Content opens a stream and transfers ownership to the caller, which must defer Close.
	Content(nCtx contextx.IContext) (io.ReadCloser, error)
}

// FileObject define this file is local file or remote file.
type FileObject string

const (
	// LocalFile means this file is local file.
	LocalFile FileObject = "local"

	// RemoteFile means this file is get from remote.
	RemoteFile FileObject = "remote"
)

// ConvertAbsPathToAbsDirs convert abs path to abs dirs.
func ConvertAbsPathToAbsDirs(path string) []string {
	dirs := make([]string, 0)
	for _, item := range strings.Split(strings.ReplaceAll(path, "\\", "/"), "/") {
		if item != "" {
			dirs = append(dirs, item)
		}
	}

	return dirs
}
