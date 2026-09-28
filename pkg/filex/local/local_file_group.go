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

package local

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/transfer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/spf13/afero"
)

const defaultBufferSize = 32 * 1024 // 32KB usually has better performance.

// NewLocalDir creates a new LocalDir.
func NewLocalDir(fullPath string) (*LocalDir, error) {
	exists, err := afero.Exists(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("file does not exist, fullPath(%s)", fullPath)
	}

	// check if path is a dir.
	isDir, err := afero.IsDir(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if path is dir: %w", err)
	}

	if !isDir {
		return nil, fmt.Errorf("path is not a dir, fullPath(%s)", fullPath)
	}

	group := &LocalDir{
		name:     filepath.Base(fullPath),
		fullPath: fullPath,
		absDirs:  fileiface.ConvertAbsPathToAbsDirs(fullPath),
	}

	return group, nil
}

// LocalDir local file group.
// nolint: revive
type LocalDir struct {
	name     string
	fullPath string
	absDirs  []string
}

// Name the name of file group.
func (group *LocalDir) Name() string {
	return group.name
}

// SubGroups the sub groups of file group.
func (group *LocalDir) SubGroups(_ contextx.IContext) ([]fileiface.FileGroup, error) {
	entries, err := afero.ReadDir(rFs(), group.fullPath)
	if err != nil {
		return nil, fmt.Errorf("read dir failed: %w", err)
	}

	subGroups := make([]fileiface.FileGroup, 0)
	for _, entry := range entries {
		fullPath := filepath.Join(group.fullPath, entry.Name())

		if entry.IsDir() {
			subDir, err := NewLocalDir(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create local file group, subgroup(%s): %w", fullPath, err)
			}

			subGroups = append(subGroups, subDir)
		}
	}

	return subGroups, nil
}

// IsDir reports whether the group-relative node is a directory.
func (group *LocalDir) IsDir(ctx contextx.IContext, relativePath string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context cannot be nil")
	}
	fullPath, err := group.resolveDirectoryPath(relativePath)
	if err != nil {
		return false, err
	}
	info, err := os.Lstat(fullPath)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

// GetSubGroup returns an existing group-relative directory.
func (group *LocalDir) GetSubGroup(ctx contextx.IContext, relativePath string) (fileiface.FileGroup, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	fullPath, err := group.resolveDirectoryPath(relativePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("inspect directory failed, path(%s): %w", fullPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory, path(%s)", fullPath)
	}

	return NewLocalDir(fullPath)
}

// EnsureSubGroup creates a group-relative directory and its missing parents.
func (group *LocalDir) EnsureSubGroup(ctx contextx.IContext, relativePath string) (fileiface.FileGroup, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	fullPath, err := group.resolveDirectoryPath(relativePath)
	if err != nil {
		return nil, err
	}
	if err := wFs().MkdirAll(fullPath, 0755); err != nil { // nolint:mnd
		return nil, fmt.Errorf("create directory failed, path(%s): %w", fullPath, err)
	}

	return NewLocalDir(fullPath)
}

// AllFiles the files of file group.
func (group *LocalDir) AllFiles(_ contextx.IContext) ([]fileiface.File, error) {
	entries, err := afero.ReadDir(rFs(), group.fullPath)
	if err != nil {
		return nil, fmt.Errorf("read dir failed: %w", err)
	}

	files := make([]fileiface.File, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(group.fullPath, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, fmt.Errorf("inspect file entry failed, path(%s): %w", fullPath, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("unsupported file type, path(%s)", fullPath)
		}

		file, err := NewLocalFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create local file. file(%s): %w", fullPath, err)
		}

		files = append(files, file)
	}

	return files, nil
}

// GetFile the func will get a file from the file group.
func (group *LocalDir) GetFile(_ contextx.IContext, name string) (fileiface.File, error) {
	fullPath := filepath.Join(group.fullPath, name)
	file, err := NewLocalFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file. file(%s): %w", fullPath, err)
	}

	return file, nil
}

// Store the func will store a file into the file group.
func (group *LocalDir) Store(nCtx contextx.IContext, info fileiface.FileInfo, reader io.ReadCloser, overwrite bool) (retErr error) {
	if reader == nil {
		return errors.New("file cannot be nil")
	}
	if nCtx == nil {
		return errors.New("context cannot be nil")
	}
	if err := nCtx.Err(); err != nil {
		return fmt.Errorf("context is done: %w", err)
	}

	// check dir exist or not.
	exists, err := afero.DirExists(wFs(), group.fullPath)
	if err != nil {
		return fmt.Errorf("check directory existence failed: %w", err)
	}

	if !exists {
		err = wFs().MkdirAll(group.fullPath, 0755) // nolint:mnd
		if err != nil {
			return fmt.Errorf("create dir failed: %w", err)
		}

		logger.G.Biz(nCtx).With("path", group.fullPath).Info("successfully create dir")
	}

	fileFullPath := filepath.Join(group.fullPath, info.Name)
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	lfile, err := wFs().OpenFile(fileFullPath, flags, 0666) // nolint:mnd // Preserve Create permissions, subject to umask.
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}

	defer func() {
		if err := lfile.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close destination file failed: %w", err))
		}
	}()

	if err := group.writeDataToFile(nCtx, lfile, reader); err != nil {
		return fmt.Errorf("write file content failed: %w", err)
	}

	return nil
}

// writeDataToFile write data to local file.
func (group *LocalDir) writeDataToFile(nCtx contextx.IContext, lfile afero.File, reader io.ReadCloser) (retErr error) {
	// use bufio.NewWriter to improve performance.
	writer := bufio.NewWriter(lfile)
	defer func() {
		if flushErr := writer.Flush(); flushErr != nil && retErr == nil {
			retErr = fmt.Errorf("flush buffer failed: %w", flushErr)
		}
	}()

	buf := make([]byte, defaultBufferSize)

	for {
		if err := nCtx.Err(); err != nil {
			return err
		}

		done, err := writeChunk(writer, buf, reader)
		if err != nil {
			return err
		}

		if done {
			return nil
		}
	}
}

// writeChunk reads one buffer-sized chunk from reader and writes it to writer.
// It returns (true, nil) on EOF, (false, nil) after a successful partial write,
// and (false, err) on any read or write error.
func writeChunk(writer *bufio.Writer, buf []byte, reader io.Reader) (bool, error) {
	n, err := reader.Read(buf)
	if err != nil {
		if err != io.EOF {
			return false, fmt.Errorf("read content failed: %w", err)
		}

		if n > 0 {
			if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
				return false, fmt.Errorf("write final buffer failed: %w", writeErr)
			}
		}

		return true, nil
	}

	if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
		return false, fmt.Errorf("failed to write file content: %w", writeErr)
	}

	return false, nil
}

// AbsDirs the func will return the abs dirs of file group.
func (group *LocalDir) AbsDirs() []string {
	return group.absDirs
}

// Copy copies a file or directory to a supported file group.
func (group *LocalDir) Copy(nCtx contextx.IContext, srcPath string, destGroup fileiface.FileGroup, destPath string, overwrite bool) error {
	if nCtx == nil {
		return errors.New("context cannot be nil")
	}
	if group == nil {
		return errors.New("source file group cannot be nil")
	}

	destLocalDir, ok := destGroup.(*LocalDir)
	if !ok {
		return transfer.CopyStream(nCtx, group, srcPath, destGroup, destPath, overwrite)
	}
	if destLocalDir == nil {
		return errors.New("destination file group cannot be nil")
	}

	srcFullPath, err := group.resolvePath(srcPath)
	if err != nil {
		return fmt.Errorf("resolve source path failed: %w", err)
	}

	destFullPath, err := destLocalDir.resolveDestinationPath(destPath)
	if err != nil {
		return fmt.Errorf("resolve destination path failed: %w", err)
	}

	srcInfo, err := rFs().Stat(srcFullPath)
	if err != nil {
		return fmt.Errorf("stat source path failed: %w", err)
	}
	if !srcInfo.IsDir() && !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("unsupported source file type, path(%s)", srcFullPath)
	}

	destInfo, err := wFs().Stat(destFullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat destination path failed: %w", err)
	}

	if srcInfo.IsDir() {
		return group.copyDirectory(nCtx, srcFullPath, destFullPath, srcInfo, destInfo, overwrite)
	}

	targetPath := destFullPath
	if destInfo != nil && destInfo.IsDir() {
		targetPath = filepath.Join(destFullPath, filepath.Base(srcFullPath))
	}

	return group.copyFile(nCtx, srcFullPath, targetPath, srcInfo, overwrite)
}

func (group *LocalDir) copyDirectory(
	nCtx contextx.IContext, srcPath string, destPath string, srcInfo os.FileInfo, destInfo os.FileInfo, overwrite bool) error {

	if destInfo != nil && !destInfo.IsDir() {
		return errors.New("cannot copy a directory to a file")
	}

	targetPath := destPath
	if destInfo != nil {
		targetPath = filepath.Join(destPath, filepath.Base(srcPath))
	}

	resolvedSource, err := filepath.Abs(srcPath)
	if err != nil {
		return fmt.Errorf("check directory overlap failed: %w", err)
	}
	resolvedSource, err = filepath.EvalSymlinks(resolvedSource)
	if err != nil {
		return fmt.Errorf("check directory overlap failed: %w", err)
	}
	resolvedTarget, err := resolveTargetPath(targetPath)
	if err != nil {
		return fmt.Errorf("check directory overlap failed: %w", err)
	}
	if isSameOrSubPath(resolvedSource, resolvedTarget) || isSameOrSubPath(resolvedTarget, resolvedSource) {
		return errors.New("cannot copy a directory to itself or its subdirectory")
	}

	if err := wFs().MkdirAll(targetPath, srcInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("create destination directory failed: %w", err)
	}

	copyEntry := func(path string, info os.FileInfo, walkErr error) error {
		return group.copyDirectoryEntry(nCtx, srcPath, targetPath, path, info, walkErr, overwrite)
	}

	if err := afero.Walk(rFs(), srcPath, copyEntry); err != nil {
		return fmt.Errorf("copy directory failed: %w", err)
	}

	return nil
}

func (group *LocalDir) copyDirectoryEntry(
	nCtx contextx.IContext, srcRoot string, destRoot string, srcPath string, srcInfo os.FileInfo, walkErr error, overwrite bool) error {

	if walkErr != nil {
		return fmt.Errorf("walk source path failed: %w", walkErr)
	}

	if err := nCtx.Err(); err != nil {
		return fmt.Errorf("context canceled: %w", err)
	}

	relPath, err := filepath.Rel(srcRoot, srcPath)
	if err != nil {
		return fmt.Errorf("resolve relative path failed: %w", err)
	}

	if relPath == "." {
		return nil
	}

	destPath := filepath.Join(destRoot, relPath)
	if srcInfo.IsDir() {
		if err := wFs().MkdirAll(destPath, srcInfo.Mode().Perm()); err != nil {
			return fmt.Errorf("create destination subdirectory failed: %w", err)
		}

		return nil
	}

	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("unsupported source file type, path(%s)", srcPath)
	}

	if err := group.copyFile(nCtx, srcPath, destPath, srcInfo, overwrite); err != nil {
		return fmt.Errorf("copy source file failed: %w", err)
	}

	return nil
}

func (group *LocalDir) copyFile(nCtx contextx.IContext, srcPath string, destPath string, srcInfo os.FileInfo, overwrite bool) (retErr error) {
	if err := nCtx.Err(); err != nil {
		return fmt.Errorf("context canceled: %w", err)
	}

	if err := wFs().MkdirAll(filepath.Dir(destPath), 0755); err != nil { // nolint:mnd
		return fmt.Errorf("create destination parent directory failed: %w", err)
	}

	destInfo, err := wFs().Stat(destPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat destination file failed: %w", err)
	}

	if destInfo != nil && os.SameFile(srcInfo, destInfo) {
		return errors.New("source and destination files must differ")
	}

	if destInfo != nil && !overwrite {
		return fmt.Errorf("destination file already exists, path(%s)", destPath)
	}

	srcFile, err := rFs().Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}

	defer func() {
		if err := srcFile.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("close source file failed: %w", err)
		}
	}()

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	destFile, err := wFs().OpenFile(destPath, flags, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("open destination file failed: %w", err)
	}

	defer func() {
		if err := destFile.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("close destination file failed: %w", err)
		}
	}()

	if err := group.writeDataToFile(nCtx, destFile, srcFile); err != nil {
		return fmt.Errorf("write destination file failed: %w", err)
	}

	return nil
}

// Remove deletes a file or directory in this file group.
func (group *LocalDir) Remove(nCtx contextx.IContext, path string) error {
	if nCtx == nil {
		return errors.New("context cannot be nil")
	}
	if err := nCtx.Err(); err != nil {
		return fmt.Errorf("context canceled: %w", err)
	}

	fullPath, err := group.resolvePath(path)
	if err != nil {
		return fmt.Errorf("resolve file path failed: %w", err)
	}
	if _, err := wFs().Stat(fullPath); err != nil {
		return fmt.Errorf("stat path failed: %w", err)
	}
	if err := wFs().RemoveAll(fullPath); err != nil {
		return fmt.Errorf("remove path failed: %w", err)
	}

	return nil
}

func (group *LocalDir) resolvePath(relativePath string) (string, error) {
	return group.resolvePathWithRoot(relativePath, false)
}

func (group *LocalDir) resolveDestinationPath(relativePath string) (string, error) {
	return group.resolvePathWithRoot(relativePath, true)
}

func (group *LocalDir) resolvePathWithRoot(relativePath string, allowGroupRoot bool) (string, error) {
	if relativePath == "" {
		return "", errors.New("path cannot be empty")
	}
	if filepath.IsAbs(relativePath) || filepath.VolumeName(relativePath) != "" {
		return "", errors.New("path must be relative")
	}

	cleanedPath := filepath.Clean(relativePath)
	if cleanedPath == "." {
		if allowGroupRoot {
			return filepath.Clean(group.fullPath), nil
		}

		return "", errors.New("path must identify a node in the file group")
	}
	if cleanedPath == ".." || strings.HasPrefix(cleanedPath, ".."+string(filepath.Separator)) {
		return "", errors.New("path must not escape the file group")
	}

	return filepath.Join(group.fullPath, cleanedPath), nil
}

func isSameOrSubPath(parentPath string, childPath string) bool {
	parentPath, err := filepath.Abs(parentPath)
	if err != nil {
		return false
	}
	childPath, err = filepath.Abs(childPath)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(parentPath, childPath)
	if err != nil {
		return false
	}

	return relPath == "." || (relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)))
}

func resolveTargetPath(targetPath string) (string, error) {
	current, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}
	var missing []string
	for {
		_, err := os.Lstat(current)
		if os.IsNotExist(err) {
			parent := filepath.Dir(current)
			if parent == current {
				return "", fmt.Errorf("no existing ancestor for target path(%s)", targetPath)
			}
			missing = append(missing, filepath.Base(current))
			current = parent

			continue
		}
		if err != nil {
			return "", err
		}

		resolved, err := filepath.EvalSymlinks(current)
		if err != nil {
			return "", err
		}
		for _, part := range slices.Backward(missing) {
			resolved = filepath.Join(resolved, part)
		}

		return resolved, nil
	}
}

// Path checks assume no concurrent changes to the directory tree.
func (group *LocalDir) resolveDirectoryPath(relativePath string) (string, error) {
	if group == nil {
		return "", errors.New("file group cannot be nil")
	}
	if relativePath == "" || strings.HasPrefix(relativePath, "/") || strings.ContainsAny(relativePath, "\\\x00") {
		return "", fmt.Errorf("path must be a relative slash path, path(%s)", relativePath)
	}
	cleaned := path.Clean(relativePath)
	fullPath, err := group.resolveDestinationPath(filepath.FromSlash(cleaned))
	if err != nil {
		return "", err
	}
	current := group.fullPath
	for part := range strings.SplitSeq("./"+cleaned, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return fullPath, nil
		}
		if err != nil {
			return "", fmt.Errorf("inspect path failed, path(%s): %w", current, err)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return "", fmt.Errorf("unsupported file type, path(%s)", current)
		}
	}

	return fullPath, nil
}
