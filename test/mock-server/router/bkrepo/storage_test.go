/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bkrepo"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testBaseDir   string
	keepTestDir   bool
	testProjectID string
	testRepoName  string
)

func initParams(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	testBaseDir = os.Getenv("BKREPO_TEST_BASE_DIR")
	keepTestDir = os.Getenv("BKREPO_TEST_KEEP_DIR") == "true"
	testProjectID = os.Getenv("BKREPO_TEST_PROJECT_ID")
	testRepoName = os.Getenv("BKREPO_TEST_REPO_NAME")
}

// newTestStorage creates a storage instance with a temporary directory.
func newTestStorage(t *testing.T) (*storage, string) {
	initParams(t)

	tmpDir, err := os.MkdirTemp(testBaseDir, "bkrepo-mock-test-*")
	require.NoError(t, err)

	conf := &Config{
		BaseDir: tmpDir,
	}

	store := newStorage(conf)

	return store, tmpDir
}

func cleanupTestDir(path string) {
	if !keepTestDir {
		os.RemoveAll(path)
	}
}

// TestStorage_CreateDir tests CreateDir.
func TestStorage_CreateDir(t *testing.T) {
	store, tmpDir := newTestStorage(t)
	defer cleanupTestDir(tmpDir)

	tests := []struct {
		name     string
		repoPath string
		wantErr  bool
		checkFn  func(t *testing.T, absPath string)
	}{
		{
			name:     "normal_test_create_dir",
			repoPath: "test/path",
			wantErr:  false,
			checkFn: func(t *testing.T, absPath string) {
				info, err := os.Stat(absPath)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name:     "normal_test_create_root_dir",
			repoPath: "/",
			wantErr:  false,
			checkFn: func(t *testing.T, absPath string) {
				info, err := os.Stat(absPath)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name:     "normal_test_create_existing_dir",
			repoPath: "existing",
			wantErr:  false,
			checkFn: func(t *testing.T, absPath string) {
				// Create again should succeed
				_, err := store.CreateDir(testProjectID, testRepoName, "existing")
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.CreateDir(testProjectID, testRepoName, tt.repoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFn != nil {
				cleanedPath := cleanRepoPath(tt.repoPath)
				absPath := getAbsPath(tmpDir, testProjectID, testRepoName, cleanedPath)
				tt.checkFn(t, absPath)
			}
		})
	}
}

// TestStorage_QueryNodeInfo tests QueryNodeInfo.
func TestStorage_QueryNodeInfo(t *testing.T) {
	store, tmpDir := newTestStorage(t)
	defer cleanupTestDir(tmpDir)

	// Setup: create a directory and a file
	testDir := "test-dir"
	testFile := "test-file.txt"
	testContent := []byte("test content")

	// Create directory
	_, err := store.CreateDir(testProjectID, testRepoName, testDir)
	require.NoError(t, err)

	// Create file
	filePath := filepath.Join(tmpDir, testProjectID, testRepoName, testDir, testFile)
	err = os.WriteFile(filePath, testContent, 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		repoPath string
		wantErr  bool
		checkFn  func(t *testing.T, nodeInfo *bkrepo.NodeInfo)
	}{
		{
			name:     "normal_test_query_dir",
			repoPath: testDir,
			wantErr:  false,
			checkFn: func(t *testing.T, nodeInfo *bkrepo.NodeInfo) {
				assert.True(t, nodeInfo.Folder)
				assert.Equal(t, testProjectID, nodeInfo.ProjectID)
				assert.Equal(t, testRepoName, nodeInfo.RepoName)
				assert.Equal(t, testDir, nodeInfo.Name)
			},
		},
		{
			name:     "normal_test_query_file",
			repoPath: testDir + "/" + testFile,
			wantErr:  false,
			checkFn: func(t *testing.T, nodeInfo *bkrepo.NodeInfo) {
				assert.False(t, nodeInfo.Folder)
				assert.Equal(t, testFile, nodeInfo.Name)
				assert.Equal(t, len(testContent), nodeInfo.Size)
			},
		},
		{
			name:     "normal_test_query_nonexistent",
			repoPath: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.QueryNodeInfo(testProjectID, testRepoName, tt.repoPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryNodeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, &resp.NodeInfo)
			}
		})
	}
}

// TestStorage_UploadFile tests UploadFile.
func TestStorage_UploadFile(t *testing.T) {
	store, tmpDir := newTestStorage(t)
	defer cleanupTestDir(tmpDir)

	tests := []struct {
		name      string
		repoPath  string
		content   string
		overwrite bool
		setup     func(t *testing.T)
		wantErr   bool
		checkFn   func(t *testing.T, absPath string)
	}{
		{
			name:      "upload_new_file",
			repoPath:  "upload/test.txt",
			content:   "hello world",
			overwrite: false,
			wantErr:   false,
			checkFn: func(t *testing.T, absPath string) {
				data, err := os.ReadFile(absPath)
				require.NoError(t, err)
				assert.Equal(t, "hello world", string(data))
			},
		},
		{
			name:      "upload_overwrite_existing",
			repoPath:  "overwrite/test.txt",
			content:   "new content",
			overwrite: true,
			setup: func(t *testing.T) {
				// Create existing file
				dir := filepath.Join(tmpDir, testProjectID, testRepoName, "overwrite")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "test.txt"), []byte("old content"), 0644))
			},
			wantErr: false,
			checkFn: func(t *testing.T, absPath string) {
				data, err := os.ReadFile(absPath)
				require.NoError(t, err)
				assert.Equal(t, "new content", string(data))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			reader := strings.NewReader(tt.content)
			resp, err := store.UploadFile(testProjectID, testRepoName, tt.repoPath, reader, tt.overwrite)

			if (err != nil) != tt.wantErr {
				t.Errorf("UploadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, testProjectID, resp.ProjectID)
				assert.Equal(t, testRepoName, resp.RepoName)
				assert.False(t, resp.Folder)
				assert.Equal(t, int64(len(tt.content)), resp.Size)

				if tt.checkFn != nil {
					cleanedPath := cleanRepoPath(tt.repoPath)
					absPath := getAbsPath(tmpDir, testProjectID, testRepoName, cleanedPath)
					tt.checkFn(t, absPath)
				}
			}
		})
	}
}

// TestStorage_DownloadFile tests DownloadFile.
func TestStorage_DownloadFile(t *testing.T) {
	store, tmpDir := newTestStorage(t)
	defer cleanupTestDir(tmpDir)

	projectID := testProjectID
	repoName := testRepoName
	testContent := []byte("download test content")

	// Setup: create a file and a directory
	fileDir := filepath.Join(tmpDir, projectID, repoName, "download")
	require.NoError(t, os.MkdirAll(fileDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fileDir, "test.txt"), testContent, 0644))

	tests := []struct {
		name     string
		repoPath string
		wantErr  bool
		checkFn  func(t *testing.T, reader io.ReadCloser, info os.FileInfo)
	}{
		{
			name:     "download_existing_file",
			repoPath: "download/test.txt",
			wantErr:  false,
			checkFn: func(t *testing.T, reader io.ReadCloser, info os.FileInfo) {
				defer reader.Close()
				data, err := io.ReadAll(reader)
				require.NoError(t, err)
				assert.Equal(t, testContent, data)
				assert.Equal(t, int64(len(testContent)), info.Size())
			},
		},
		{
			name:     "download_nonexistent_file",
			repoPath: "download/nonexistent.txt",
			wantErr:  true,
		},
		{
			name:     "download_directory_should_fail",
			repoPath: "download",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, info, err := store.DownloadFile(projectID, repoName, tt.repoPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, reader, info)
			}
		})
	}
}

// TestStorage_ListNode tests ListNode.
func TestStorage_ListNode(t *testing.T) {
	store, tmpDir := newTestStorage(t)
	defer cleanupTestDir(tmpDir)

	projectID := testProjectID
	repoName := testRepoName

	// Setup: create test directory structure
	listDir := filepath.Join(tmpDir, projectID, repoName, "list")
	require.NoError(t, os.MkdirAll(listDir, 0755))

	// Create some files and subdirectories
	for i := 0; i < 5; i++ {
		require.NoError(t, os.WriteFile(filepath.Join(listDir, fmt.Sprintf("file%d.txt", i)), []byte("content"), 0644))
	}
	require.NoError(t, os.MkdirAll(filepath.Join(listDir, "subdir1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(listDir, "subdir2"), 0755))

	tests := []struct {
		name       string
		repoPath   string
		pageNumber int
		pageSize   int
		wantErr    bool
		checkFn    func(t *testing.T, resp *bkrepo.ListNodeResp)
	}{
		{
			name:       "list_first_page",
			repoPath:   "list",
			pageNumber: 1,
			pageSize:   3,
			wantErr:    false,
			checkFn: func(t *testing.T, resp *bkrepo.ListNodeResp) {
				assert.Equal(t, 1, resp.PageNumber)
				assert.Equal(t, 3, resp.PageSize)
				assert.Len(t, resp.Records, 3)
			},
		},
		{
			name:       "list_second_page",
			repoPath:   "list",
			pageNumber: 2,
			pageSize:   3,
			wantErr:    false,
			checkFn: func(t *testing.T, resp *bkrepo.ListNodeResp) {
				assert.Equal(t, 2, resp.PageNumber)
				assert.Equal(t, 3, resp.PageSize)
				assert.Len(t, resp.Records, 3)
			},
		},
		{
			name:       "list_nonexistent_dir",
			repoPath:   "nonexistent",
			pageNumber: 1,
			pageSize:   10,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.ListNode(projectID, repoName, tt.repoPath, tt.pageNumber, tt.pageSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, resp)
			}
		})
	}
}
