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
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
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
			return fmt.Errorf("failed to get absolute path, path(%s), err: %w", path, err)
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
		return fmt.Errorf("failed to remove file, file-name(%s), err: %w", tmpFile, err)
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
	exists, err := PathExists(dirPath)
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

	_, err = os.Stat(absDirPath)
	if err != nil {
		return 0, err
	}

	freeSpace, err := checkFreeSpace(absDirPath)
	if err != nil {
		return 0, fmt.Errorf("failed to count free space: %w", err)
	}

	return freeSpace, nil
}

// CheckNetTCPOpen check the remote server is open tcp port.
func CheckNetTCPOpen(host string, port int, timeout time.Duration) (bool, error) {
	target := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout(NetTCP, target, timeout)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = conn.Close()
	}()

	return true, nil
}

const checkPortIdleTimeout = 2 * time.Second

// CheckTCPPortIdle check tcp port is in use or not.
func CheckTCPPortIdle(ctx context.Context, port uint64) (bool, error) {
	address := fmt.Sprintf("127.0.0.1:%d", port)

	tCtx, cancel := context.WithTimeout(ctx, checkPortIdleTimeout)
	defer cancel()

	// try to connect tcp port.
	var dialer net.Dialer
	conn, _ := dialer.DialContext(tCtx, NetTCP, address)
	if conn != nil {
		_ = conn.Close()
		return false, errors.New("port is in use")
	}

	// port could not be connected, but could't be sure whether it is idle.
	var listener net.Listener
	listener, err := net.Listen(NetTCP, address)
	if err != nil {
		return false, err
	}

	_ = listener.Close()

	return true, nil
}

// CheckTCP6PortIdle check tcp6 port is in use or not.
func CheckTCP6PortIdle(ctx context.Context, port uint64) (bool, error) {
	address := fmt.Sprintf("[::1]:%d", port)

	tCtx, cancel := context.WithTimeout(ctx, checkPortIdleTimeout)
	defer cancel()

	// try to connect tcp port.
	var dialer net.Dialer
	conn, _ := dialer.DialContext(tCtx, NetTCP, address)
	if conn != nil {
		_ = conn.Close()

		return false, errors.New("port is in use")
	}

	// port could not be connected, but could't be sure whether it is idle.
	var listener net.Listener
	listener, err := net.Listen(NetTCP6, address)
	if err != nil {
		return false, err
	}

	_ = listener.Close()

	return true, nil
}

// CleanDirectory clean directory.
func CleanDirectory(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	paths := make([]string, len(entries))
	for idx, entry := range entries {
		paths[idx] = filepath.Join(path, entry.Name())
		if err := os.RemoveAll(paths[idx]); err != nil {
			return nil, fmt.Errorf("failed to remove %s: %w", paths[idx], err)
		}
	}

	return paths, nil
}
