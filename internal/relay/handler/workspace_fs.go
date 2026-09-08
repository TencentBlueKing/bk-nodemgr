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

package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type workspaceFS struct {
	rootDir string
}

func newWorkspaceFS(rootDir string) workspaceFS {
	return workspaceFS{rootDir: rootDir}
}

// instanceDir returns the staging directory dedicated to one operation instance.
//
// Every concurrent installation gets its own directory so that transfers, reads and cleanup
// of the same package name cannot interfere: sharing a single path let one flow delete or
// overwrite the file another flow was still using.
func (fs workspaceFS) instanceDir(operInstID string) (string, error) {
	segment := sanitizePathSegment(operInstID)
	if err := validateWorkspaceFilename(segment); err != nil {
		return "", fmt.Errorf("invalid operation instance id for staging dir. oper-inst-id(%s): %w", operInstID, err)
	}

	absPath := filepath.Clean(filepath.Join(fs.rootDir, segment))
	if err := fs.checkPathInWorkspace(absPath); err != nil {
		return "", fmt.Errorf("will not resolve staging dir outside workspace. root(%s), oper-inst-id(%s): %w",
			fs.rootDir, operInstID, err)
	}

	return absPath, nil
}

// ensureInstanceDir creates the staging directory of one operation instance.
func (fs workspaceFS) ensureInstanceDir(operInstID string) (string, error) {
	absPath, err := fs.instanceDir(operInstID)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(absPath, storageTmpDirMode); err != nil {
		return "", fmt.Errorf("failed to ensure staging dir. path(%s): %w", absPath, err)
	}

	return absPath, nil
}

// removeInstanceDir drops the whole staging directory of one operation instance.
func (fs workspaceFS) removeInstanceDir(operInstID string) error {
	absPath, err := fs.instanceDir(operInstID)
	if err != nil {
		return err
	}

	return fs.removeAll(absPath)
}

// sanitizePathSegment maps everything outside an allow-list to '_' so an identifier is safe
// to use as a single path segment. Operation instance ids look like "oper-inst:<hex>", and
// the colon is not portable across the platforms a relay runs on.
func sanitizePathSegment(segment string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, segment)
}

// listOrphanInstanceDirs returns the staging directories not modified since cutoff.
//
// Age is the criterion rather than liveness: a directory is created before its transfer
// starts, so a freshly created empty one is normal and must not be reclaimed. Adding or
// removing files inside a directory bumps its mtime, which keeps active ones out of range.
func (fs workspaceFS) listOrphanInstanceDirs(cutoff time.Time) ([]string, error) {
	entries, err := os.ReadDir(fs.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read staging root. dir(%s): %w", fs.rootDir, err)
	}

	orphans := make([]string, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			continue
		}

		orphans = append(orphans, filepath.Join(fs.rootDir, entry.Name()))
	}

	return orphans, nil
}

func validateWorkspaceFilename(filename string) error {
	if filename == "" || filename == "." || filename == ".." {
		return fmt.Errorf("invalid workspace filename. filename(%s)", filename)
	}
	if filepath.IsAbs(filename) {
		return fmt.Errorf("invalid workspace filename, got absolute path. filename(%s)", filename)
	}
	if strings.ContainsAny(filename, "/\\") || strings.ContainsRune(filename, 0) || filepath.Base(filename) != filename {
		return fmt.Errorf("invalid workspace filename, expected basename only. filename(%s)", filename)
	}

	return nil
}

func (fs workspaceFS) checkPathInWorkspace(absPath string) error {
	if fs.isProtectedPath(absPath) {
		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	rootDir := filepath.Clean(fs.rootDir)
	relPath, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		return fmt.Errorf("failed to check if path in workspace. path(%s): %w", absPath, err)
	}

	if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." || filepath.IsAbs(relPath) {
		return fmt.Errorf("path not in workspace. root(%s), path(%s)", rootDir, absPath)
	}

	return nil
}

func (fs workspaceFS) removeAll(absPath string) error {
	if fs.isProtectedPath(absPath) {
		return fmt.Errorf("failed to remove all, got invalid path. path(%s)", absPath)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to remove all. path(%s): %w", absPath, err)
	}

	return nil
}

func (fs workspaceFS) isProtectedPath(absPath string) bool {
	return absPath == "" ||
		absPath == "/" ||
		strings.HasPrefix(absPath, "/dev/") ||
		strings.HasPrefix(absPath, "/sys/") ||
		strings.HasPrefix(absPath, "/proc/")
}
