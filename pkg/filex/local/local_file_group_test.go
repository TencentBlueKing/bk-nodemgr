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

// Package local tests the local implementation of file groups.
package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func TestNewLocalDir(t *testing.T) {
	dir := t.TempDir()
	group, err := NewLocalDir(dir)
	if err != nil {
		t.Fatalf("NewLocalDir() error = %v", err)
	}
	if group.Name() != filepath.Base(dir) {
		t.Errorf("Name() = %q, want %q", group.Name(), filepath.Base(dir))
	}
}

func TestLocalDir_Copy(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(t *testing.T, sourceRoot, destinationRoot string)
		srcPath     string
		destPath    string
		overwrite   bool
		wantErr     bool
		wantPath    string
		wantContent string
	}{
		{
			name: "file to file", srcPath: "source.txt", destPath: "target.txt", wantPath: "target.txt", wantContent: "source",
			prepare: func(t *testing.T, sourceRoot, _ string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source.txt"), "source")
			},
		},
		{
			name:        "file to existing directory",
			srcPath:     "source.txt",
			destPath:    "target",
			wantPath:    "target/source.txt",
			wantContent: "source",
			prepare: func(t *testing.T, sourceRoot, destinationRoot string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source.txt"), "source")
				makeTestDir(t, filepath.Join(destinationRoot, "target"))
			},
		},
		{
			name:        "file to destination group root",
			srcPath:     "source.txt",
			destPath:    ".",
			wantPath:    "source.txt",
			wantContent: "source",
			prepare: func(t *testing.T, sourceRoot, _ string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source.txt"), "source")
			},
		},
		{
			name:        "directory to existing directory copies source directory",
			srcPath:     "source",
			destPath:    "target",
			wantPath:    "target/source/nested/file.txt",
			wantContent: "source",
			prepare: func(t *testing.T, sourceRoot, destinationRoot string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source/nested/file.txt"), "source")
				makeTestDir(t, filepath.Join(destinationRoot, "target"))
			},
		},
		{
			name:        "directory to missing target copies contents",
			srcPath:     "source",
			destPath:    "target",
			wantPath:    "target/nested/file.txt",
			wantContent: "source",
			prepare: func(t *testing.T, sourceRoot, _ string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source/nested/file.txt"), "source")
			},
		},
		{
			name: "directory to file returns error", srcPath: "source", destPath: "target", wantErr: true,
			prepare: func(t *testing.T, sourceRoot, destinationRoot string) {
				makeTestDir(t, filepath.Join(sourceRoot, "source"))
				writeTestFile(t, filepath.Join(destinationRoot, "target"), "destination")
			},
		},
		{
			name: "existing file requires overwrite", srcPath: "source.txt", destPath: "target.txt", wantErr: true,
			prepare: func(t *testing.T, sourceRoot, destinationRoot string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source.txt"), "source")
				writeTestFile(t, filepath.Join(destinationRoot, "target.txt"), "destination")
			},
		},
		{
			name:        "existing file is replaced when overwrite enabled",
			srcPath:     "source.txt",
			destPath:    "target.txt",
			overwrite:   true,
			wantPath:    "target.txt",
			wantContent: "source",
			prepare: func(t *testing.T, sourceRoot, destinationRoot string) {
				writeTestFile(t, filepath.Join(sourceRoot, "source.txt"), "source")
				writeTestFile(t, filepath.Join(destinationRoot, "target.txt"), "destination")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceRoot, destinationRoot := t.TempDir(), t.TempDir()
			tt.prepare(t, sourceRoot, destinationRoot)
			sourceGroup := newTestLocalDir(t, sourceRoot)
			destinationGroup := newTestLocalDir(t, destinationRoot)

			err := sourceGroup.Copy(contextx.Background(), tt.srcPath, destinationGroup, tt.destPath, tt.overwrite)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantPath != "" {
				assertTestFile(t, filepath.Join(destinationRoot, tt.wantPath), tt.wantContent)
			}
		})
	}
}

func TestLocalDir_CopyRejectsTraversal(t *testing.T) {
	sourceRoot, destinationRoot := t.TempDir(), t.TempDir()
	writeTestFile(t, filepath.Join(sourceRoot, "source.txt"), "source")
	writeTestFile(t, filepath.Join(sourceRoot, "source-dir/file.txt"), "source")
	sourceGroup := newTestLocalDir(t, sourceRoot)
	destinationGroup := newTestLocalDir(t, destinationRoot)

	for _, tt := range []struct{ name, srcPath, destPath string }{
		{name: "source parent", srcPath: "../source.txt", destPath: "target.txt"},
		{name: "destination parent", srcPath: "source.txt", destPath: "../target.txt"},
		{name: "source absolute", srcPath: filepath.Join(sourceRoot, "source.txt"), destPath: "target.txt"},
		{name: "destination absolute", srcPath: "source.txt", destPath: filepath.Join(destinationRoot, "target.txt")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := sourceGroup.Copy(contextx.Background(), tt.srcPath, destinationGroup, tt.destPath, false); err == nil {
				t.Fatal("Copy() error = nil, want path validation error")
			}
		})
	}

	for _, destPath := range []string{"source-dir", "source-dir/subdir"} {
		if err := sourceGroup.Copy(contextx.Background(), "source-dir", sourceGroup, destPath, false); err == nil {
			t.Fatalf("Copy() to %q error = nil, want source directory containment error", destPath)
		}
	}
}

func TestLocalDir_Remove(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "directory/file.txt"), "content")
	group := newTestLocalDir(t, root)
	if err := group.Remove(contextx.Background(), "directory"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "directory")); !os.IsNotExist(err) {
		t.Fatalf("removed directory stat error = %v, want not exist", err)
	}
	if err := group.Remove(contextx.Background(), "../outside"); err == nil {
		t.Fatal("Remove() error = nil, want path validation error")
	}
}

func TestLocalDir_CopyAndRemoveRejectCanceledContext(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "source.txt"), "source")
	group := newTestLocalDir(t, root)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	nCtx := contextx.New(ctx)
	if err := group.Copy(nCtx, "source.txt", group, "target.txt", false); err == nil {
		t.Fatal("Copy() error = nil, want canceled context error")
	}
	if err := group.Remove(nCtx, "source.txt"); err == nil {
		t.Fatal("Remove() error = nil, want canceled context error")
	}
}

func newTestLocalDir(t *testing.T, path string) *LocalDir {
	t.Helper()
	group, err := NewLocalDir(path)
	if err != nil {
		t.Fatalf("NewLocalDir(%q) error = %v", path, err)
	}
	return group
}

func makeTestDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil { // nolint:mnd
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	makeTestDir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0644); err != nil { // nolint:mnd
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertTestFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(got) != want {
		t.Errorf("ReadFile(%q) = %q, want %q", path, got, want)
	}
}
