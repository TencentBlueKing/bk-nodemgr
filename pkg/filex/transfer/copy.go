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

// Package transfer provides low-level streaming copy for file group backends.
// Business callers use FileGroup.Copy so backends can apply native copy policy.
package transfer

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// CopyStream streams a node between backends using group-relative slash paths.
// Callers must handle native dispatch, source/target identity and directory overlap,
// and tenant policy before invoking it. Business callers use FileGroup.Copy.
// Partial destination writes are not rolled back.
func CopyStream(ctx contextx.IContext, src fileiface.FileGroup, srcPath string, dest fileiface.FileGroup, destPath string, overwrite bool) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if src == nil || dest == nil {
		return errors.New("source and destination file groups cannot be nil")
	}

	isDir, err := src.IsDir(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("stat source failed, path(%s): %w", srcPath, err)
	}
	srcPath = path.Clean(srcPath)
	if srcPath == "." {
		return errors.New("source path must identify a node in the file group")
	}
	targetIsDir, err := dest.IsDir(ctx, destPath)
	targetMissing := errors.Is(err, fs.ErrNotExist)
	if err != nil && !targetMissing {
		return fmt.Errorf("stat destination failed, path(%s): %w", destPath, err)
	}
	destPath = path.Clean(destPath)
	if isDir && !targetMissing && !targetIsDir {
		return errors.New("cannot copy a directory to a file")
	}
	if targetIsDir {
		destPath = path.Join(destPath, path.Base(srcPath))
	}
	if !isDir {
		return copyFile(ctx, src, srcPath, dest, destPath, overwrite)
	}
	srcGroup, err := src.GetSubGroup(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("get source directory failed, path(%s): %w", srcPath, err)
	}
	dstGroup, err := dest.EnsureSubGroup(ctx, destPath)
	if err != nil {
		return fmt.Errorf("create destination directory failed, path(%s): %w", destPath, err)
	}

	return copyDir(ctx, srcGroup, dstGroup, overwrite)
}

func validateChildName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return fmt.Errorf("invalid child name, name(%s)", name)
	}

	return nil
}

func copyDir(ctx contextx.IContext, src fileiface.FileGroup, dst fileiface.FileGroup, overwrite bool) error {
	subgroups, err := src.SubGroups(ctx)
	if err != nil {
		return fmt.Errorf("list source directories failed: %w", err)
	}
	for _, subgroup := range subgroups {
		if err := validateChildName(subgroup.Name()); err != nil {
			return err
		}
		child, err := dst.EnsureSubGroup(ctx, subgroup.Name())
		if err != nil {
			return fmt.Errorf("create destination directory failed, name(%s): %w", subgroup.Name(), err)
		}
		if err := copyDir(ctx, subgroup, child, overwrite); err != nil {
			return err
		}
	}

	files, err := src.AllFiles(ctx)
	if err != nil {
		return fmt.Errorf("list source files failed: %w", err)
	}
	for _, file := range files {
		info := file.Info()
		if err := validateChildName(info.Name); err != nil {
			return err
		}
		// Enumeration may include nodes that the backend does not support copying.
		isDir, err := src.IsDir(ctx, info.Name)
		if err != nil {
			return fmt.Errorf("stat source file failed, name(%s): %w", info.Name, err)
		}
		if isDir {
			return fmt.Errorf("source file is now a directory, name(%s)", info.Name)
		}
		if err := storeFile(ctx, file, dst, info, overwrite); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(ctx contextx.IContext, src fileiface.FileGroup, srcPath string, dst fileiface.FileGroup, dstPath string, overwrite bool) error {
	parent, err := dst.EnsureSubGroup(ctx, path.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("create destination parent failed, path(%s): %w", dstPath, err)
	}
	file, err := src.GetFile(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("get source file failed, path(%s): %w", srcPath, err)
	}
	info := file.Info()
	info.Name = path.Base(dstPath)

	return storeFile(ctx, file, parent, info, overwrite)
}

func storeFile(ctx contextx.IContext, file fileiface.File, dst fileiface.FileGroup, info fileiface.FileInfo, overwrite bool) (retErr error) {
	isDir, err := dst.IsDir(ctx, info.Name)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat destination file failed, name(%s): %w", info.Name, err)
	}
	if isDir {
		return fmt.Errorf("destination is a directory, name(%s)", info.Name)
	}

	reader, err := file.Content(ctx)
	if err != nil {
		return fmt.Errorf("open source content failed, name(%s): %w", info.Name, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close source content failed, name(%s): %w", info.Name, err))
		}
	}()
	if err := dst.Store(ctx, info, reader, overwrite); err != nil {
		return fmt.Errorf("copy file failed, name(%s): %w", info.Name, err)
	}

	return nil
}
