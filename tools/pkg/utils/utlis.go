/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils ...
package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// PathExists check if a path exists.
func PathExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}

		return false, fmt.Errorf("%s is not a directory", path)
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// DirExists check if a dir exists.
func DirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}

		return false, fmt.Errorf("%s is not a directory", path)
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	absSrcPath, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	absDstPath, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	// make sure the parent directory exists.
	if err := TryCreateDir(filepath.Dir(absDstPath)); err != nil {
		return err
	}

	// nolint: gosec
	source, err := os.Open(absSrcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	// nolint: gosec
	destination, err := os.Create(absDstPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = destination.Close()
	}()

	if _, err = io.Copy(destination, source); err != nil {
		return err
	}

	// copy file mode if possible.
	sourceInfo, err := os.Stat(absSrcPath)
	if err != nil {
		return err
	}

	return os.Chmod(absDstPath, sourceInfo.Mode())
}

// CopyDir copies a directory from src to dst.
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err := CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err := CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CheckWritePermission check write permission.
func CheckWritePermission(path string) error {
	// if path does not exist, use parent directory as dir.
	dir := path
	if exists, _ := PathExists(path); !exists {
		path, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path, path(%s): %w", path, err)
		}

		dir = filepath.Dir(path)
	}

	// create a temporary file in the directory to check write permission.
	tmpFile := filepath.Join(dir, ".write_test_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	// nolint: mnd
	if err := os.WriteFile(tmpFile, []byte(""), 0600); err != nil {
		return fmt.Errorf("no write permission: %w", err)
	}

	if err := os.Remove(tmpFile); err != nil {
		return fmt.Errorf("failed to remove file, file-name(%s): %w", tmpFile, err)
	}

	return nil
}

// BackupExistingDir backup existing files.
func BackupExistingDir(dirPath string) (string, error) {
	// get all existing entries.
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	// if there are no existing entries, return.
	if len(entries) == 0 {
		return "", nil
	}

	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(filepath.Dir(dirPath), "backup",
		filepath.Base(dirPath)+timestamp)

	// range existing entries and backup them.
	for _, entry := range entries {
		srcPath := filepath.Join(dirPath, entry.Name())
		dstPath := filepath.Join(backupDir, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return backupDir, fmt.Errorf("failed to backup directory %s: %w", entry.Name(), err)
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return backupDir, fmt.Errorf("failed to backup file %s: %w", entry.Name(), err)
			}
		}
	}

	return backupDir, nil
}

// BackupExistingDirIgnoreErr backup existing files and ignore error.
func BackupExistingDirIgnoreErr(dirPath string) (string, []error) {
	// get all existing entries.
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", []error{err}
	}

	// if there are no existing entries, return.
	if len(entries) == 0 {
		return "", nil
	}

	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(filepath.Dir(dirPath), "backup",
		filepath.Base(dirPath)+timestamp)

	var errs []error
	// range existing entries and backup them.
	for _, entry := range entries {
		srcPath := filepath.Join(dirPath, entry.Name())
		dstPath := filepath.Join(backupDir, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				errs = append(errs, fmt.Errorf("failed to backup directory %s: %w", entry.Name(), err))
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				errs = append(errs, fmt.Errorf("failed to backup file %s: %w", entry.Name(), err))
			}
		}
	}

	return backupDir, errs
}

// MakeExecutable grants executable permissions to a binary file.
// This function works cross-platform, handling the differences between Unix-like
// systems (Linux, macOS) and Windows.
func MakeExecutable(filePath string) error {
	if runtime.GOOS == "windows" {
		// Windows doesn't use the same permission system as Unix-like OSType
		// File extensions (.exe, .bat, etc.) determine executability
		// No need to change permissions, but we need to check if the file exists.
		_, err := os.Stat(filePath)
		return err
	}

	// For Unix-like systems (Linux, macOS, etc.)
	// Get current file info to preserve existing permissions
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// Get current permissions
	mode := info.Mode()

	// Add executable bit for user, group and others
	// This is equivalent to chmod +x
	// nolint: mnd
	newMode := mode | 0111

	// Apply new permissions
	if err := os.Chmod(filePath, newMode); err != nil {
		return fmt.Errorf("failed to set executable permission on %s: %w", filePath, err)
	}

	return nil
}

// TryCreateDir ...
func TryCreateDir(dirPath string) error {
	// 1. check dirPath is valid.
	if dirPath = strings.TrimSpace(dirPath); dirPath == "" {
		return errors.New("dirPath path cannot be empty")
	}

	// 2. check if dirPath exists.
	exists, err := DirExists(dirPath)
	if err != nil {
		return fmt.Errorf("failed to check path existence: %w", err)
	}
	// nolint: mnd
	if !exists {
		if err := os.MkdirAll(dirPath, 0750); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}
	}

	return nil
}

// CountDiskFreeSpace count disk free space.
func CountDiskFreeSpace(dirPath string) (uint64, error) {
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return 0, err
	}

	existingPath := absDirPath
	for {
		_, err = os.Stat(existingPath)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return 0, err
		}

		parentPath := filepath.Dir(existingPath)
		if parentPath == existingPath {
			return 0, errors.New("no existing parent directory found")
		}

		existingPath = parentPath
	}

	freeSpace, err := countFreeSpace(existingPath)
	if err != nil {
		return 0, fmt.Errorf("failed to count free space: %w", err)
	}

	return freeSpace, nil
}

// CheckNetTCPOpen check the remote server is open tcp port.
func CheckNetTCPOpen(host string, port uint64, timeout time.Duration) (bool, error) {
	conn, err := net.DialTimeout(NetTCP.String(), net.JoinHostPort(host, strconv.FormatUint(port, 10)), timeout)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = conn.Close()
	}()

	return true, nil
}

// CheckTCPPortIdle check tcp port is in use or not.
func CheckTCPPortIdle(_ context.Context, network, host string, port uint64) (bool, error) {
	listener, err := net.ListenTCP(network, &net.TCPAddr{IP: net.ParseIP(host), Port: int(port)})
	if err != nil {
		return false, err
	}

	_ = listener.Close()

	return true, nil
}

// CheckUDPPortIdle check udp port is in use or not.
func CheckUDPPortIdle(_ context.Context, network, host string, port uint64) (bool, error) {
	listener, err := net.ListenUDP(network, &net.UDPAddr{IP: net.ParseIP(host), Port: int(port)})
	if err != nil {
		return false, err
	}

	_ = listener.Close()

	return true, nil
}

// ListFiles list all files in a directory.
func ListFiles(dir string) ([]string, error) {
	if dir == "" {
		return nil, errors.New("invalid directory path")
	}

	// validate directory existence and permission.
	dirInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist, dir(%w)", err)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied, dir(%s)", dir)
		}

		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	files := make([]string, 0)
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return files, nil
}

// CollectPIDFiles returns all .pid file paths found in pidDir.
// If pidDir is empty, a nil slice is returned without error.
// A relative pidDir is resolved against absRootDir.
func CollectPIDFiles(absRootDir, pidDir string) ([]string, error) {
	pidDir = strings.TrimSpace(pidDir)
	if pidDir == "" {
		return nil, nil
	}

	pidDirAbs := filepath.Clean(pidDir)
	if !filepath.IsAbs(pidDirAbs) {
		pidDirAbs = filepath.Clean(filepath.Join(absRootDir, pidDirAbs))
	}

	if err := CheckDirPathSafe(pidDirAbs); err != nil {
		return nil, fmt.Errorf("invalid pid dir(%s): %w", pidDirAbs, err)
	}

	entries, err := os.ReadDir(pidDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read pid dir(%s): %w", pidDirAbs, err)
	}

	pidFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".pid") {
			continue
		}

		pidFiles = append(pidFiles, filepath.Join(pidDirAbs, entry.Name()))
	}

	return pidFiles, nil
}

// ReadPIDFromFile reads and validates a PID from the given file.
func ReadPIDFromFile(pidFile string) (int, error) {
	content, err := os.ReadFile(filepath.Clean(pidFile))
	if err != nil {
		return 0, fmt.Errorf("failed to read pid file(%s): %w", pidFile, err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid content(%s): %w", string(content), err)
	}
	if pid <= 1 {
		return 0, fmt.Errorf("dangerous pid(%d)", pid)
	}

	return pid, nil
}
