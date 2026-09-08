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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// realOperInstID mirrors what pkg/identifier produces: a tag, a colon, then a uuid without
// dashes. The colon is the reason a staging segment has to be sanitized.
const realOperInstID = "oper-inst:dfd9917fb77942e5ac061cc6a8077df2"

func TestSanitizePathSegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "colon in a real operation instance id",
			in:   realOperInstID,
			want: "oper-inst_dfd9917fb77942e5ac061cc6a8077df2",
		},
		{
			name: "path separators are neutralized",
			in:   "../../etc/passwd",
			want: "______etc_passwd",
		},
		{
			name: "windows separators are neutralized",
			in:   `..\..\windows`,
			want: "______windows",
		},
		{
			name: "allowed characters are preserved",
			in:   "abcXYZ-089_",
			want: "abcXYZ-089_",
		},
		{
			name: "null byte is neutralized",
			in:   "oper\x00inst",
			want: "oper_inst",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizePathSegment(tt.in)
			assert.Equal(t, tt.want, got)
			assert.NotContains(t, got, string(filepath.Separator), "a segment must never contain a separator")
		})
	}
}

// TestInstanceDir_IsDeterministic covers the invariant the whole design rests on: CheckPkgStats
// and StoragePkg arrive as two independent events and must derive the same directory from the
// operation instance id alone, with no shared state between them.
func TestInstanceDir_IsDeterministic(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(t.TempDir())

	first, err := fs.instanceDir(realOperInstID)
	require.NoError(t, err)

	second, err := fs.instanceDir(realOperInstID)
	require.NoError(t, err)

	assert.Equal(t, first, second, "the same operation instance must always map to the same dir")
}

func TestInstanceDir_IsolatesInstances(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(t.TempDir())

	first, err := fs.instanceDir("oper-inst:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, err)

	second, err := fs.instanceDir("oper-inst:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.NoError(t, err)

	assert.NotEqual(t, first, second, "different operation instances must not share a staging dir")
}

func TestInstanceDir_StaysInsideWorkspace(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	fs := newWorkspaceFS(rootDir)

	// Traversal attempts survive sanitization only as literal underscores, so they resolve to
	// a harmless child of the workspace rather than escaping it.
	for _, operInstID := range []string{"../../etc", "/absolute/path", `..\..\win`, ".."} {
		got, err := fs.instanceDir(operInstID)
		require.NoError(t, err, "oper-inst-id(%s) should sanitize to a usable segment", operInstID)
		assert.Equal(t, rootDir, filepath.Dir(got), "oper-inst-id(%s) escaped the workspace", operInstID)
	}
}

func TestInstanceDir_RejectsUnusableIDs(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(t.TempDir())

	// An empty id sanitizes to an empty segment, which would resolve to the workspace root
	// itself: removing it would wipe every concurrent installation's staging dir.
	_, err := fs.instanceDir("")
	require.Error(t, err, "an empty operation instance id must be rejected")
}

func TestEnsureInstanceDir(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(filepath.Join(t.TempDir(), "transfer-file"))

	// The workspace root does not exist yet: ensuring a staging dir must create the whole chain.
	dir, err := fs.ensureInstanceDir(realOperInstID)
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Ensuring twice must be a no-op rather than an error: a retried action replays the event.
	again, err := fs.ensureInstanceDir(realOperInstID)
	require.NoError(t, err)
	assert.Equal(t, dir, again)
}

// TestRemoveInstanceDir_LeavesSiblingsIntact is the regression test for the outage where one
// installation deleted the package another installation was still reading, surfacing as
// "failed to open source file: ... no such file or directory".
func TestRemoveInstanceDir_LeavesSiblingsIntact(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(t.TempDir())

	const (
		mineID  = "oper-inst:11111111111111111111111111111111"
		otherID = "oper-inst:22222222222222222222222222222222"
	)

	mine, err := fs.ensureInstanceDir(mineID)
	require.NoError(t, err)

	other, err := fs.ensureInstanceDir(otherID)
	require.NoError(t, err)

	// Both installations stage a package under the very same name.
	const pkgName = "installer_windows_amd64.exe"
	require.NoError(t, os.WriteFile(filepath.Join(mine, pkgName), []byte("mine"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(other, pkgName), []byte("other"), 0600))

	require.NoError(t, fs.removeInstanceDir(mineID))

	_, err = os.Stat(mine)
	assert.True(t, os.IsNotExist(err), "the owning instance's staging dir should be gone")

	content, err := os.ReadFile(filepath.Join(other, pkgName))
	require.NoError(t, err, "a concurrent instance's package must survive")
	assert.Equal(t, []byte("other"), content)
}

func TestRemoveInstanceDir_IsIdempotent(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(t.TempDir())

	// StoragePkg cleans up even when nothing was ever staged, so a missing dir is not an error.
	require.NoError(t, fs.removeInstanceDir(realOperInstID))
}

func TestListOrphanInstanceDirs(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	fs := newWorkspaceFS(rootDir)

	staleDir := filepath.Join(rootDir, "oper-inst_stale")
	require.NoError(t, os.MkdirAll(staleDir, storageTmpDirMode))

	freshDir := filepath.Join(rootDir, "oper-inst_fresh")
	require.NoError(t, os.MkdirAll(freshDir, storageTmpDirMode))

	// A loose file in the staging root is not an instance dir and must be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "loose.txt"), []byte("x"), 0600))

	staleTime := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(staleDir, staleTime, staleTime))

	orphans, err := fs.listOrphanInstanceDirs(time.Now().Add(-24 * time.Hour))
	require.NoError(t, err)

	assert.Equal(t, []string{staleDir}, orphans,
		"only directories untouched for longer than the cutoff are orphans")
}

func TestListOrphanInstanceDirs_MissingRoot(t *testing.T) {
	t.Parallel()

	// Before the first installation the staging root does not exist yet. The GC runs on
	// startup, so this must not be reported as a failure.
	fs := newWorkspaceFS(filepath.Join(t.TempDir(), "never-created"))

	orphans, err := fs.listOrphanInstanceDirs(time.Now())
	require.NoError(t, err)
	assert.Empty(t, orphans)
}

func TestValidateWorkspaceFilename(t *testing.T) {
	t.Parallel()

	valid := []string{
		"installer_windows_amd64.exe",
		"gse_agent-2-v2.1.6-beta.68-windows_amd64.tgz",
	}

	for _, name := range valid {
		assert.NoError(t, validateWorkspaceFilename(name), "filename(%s) should be accepted", name)
	}

	invalid := []string{
		"",
		".",
		"..",
		"/etc/passwd",
		"../escape.tgz",
		"sub/dir.tgz",
		`sub\dir.tgz`,
		"nul\x00byte.tgz",
	}

	for _, name := range invalid {
		assert.Error(t, validateWorkspaceFilename(name), "filename(%q) should be rejected", name)
	}
}

func TestRemoveAll_RefusesProtectedPaths(t *testing.T) {
	t.Parallel()

	fs := newWorkspaceFS(t.TempDir())

	for _, path := range []string{"", "/", "/dev/null", "/sys/kernel", "/proc/1"} {
		assert.Error(t, fs.removeAll(path), "path(%s) must be refused", path)
	}
}

func TestCheckPathInWorkspace(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	fs := newWorkspaceFS(rootDir)

	assert.NoError(t, fs.checkPathInWorkspace(filepath.Join(rootDir, "child")))
	assert.NoError(t, fs.checkPathInWorkspace(filepath.Join(rootDir, "child", "grandchild")))

	parent := filepath.Dir(strings.TrimSuffix(rootDir, string(filepath.Separator)))
	assert.Error(t, fs.checkPathInWorkspace(parent), "the workspace parent must be out of bounds")
	assert.Error(t, fs.checkPathInWorkspace("/etc/passwd"), "an unrelated absolute path must be out of bounds")
}
