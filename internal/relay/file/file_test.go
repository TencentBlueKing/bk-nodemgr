/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package file

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/stretchr/testify/assert"
)

const (
	baseDir      = ""
	testFilename = "gse_agent_inner-v2.1.6-beta.59.tgz"
)

// testClient ...
func setupTestFileManager(t *testing.T) (IFileManager, string, func()) {
	testFilePath := filepath.Join(baseDir, "test.go")

	if err := os.WriteFile(testFilePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	fm := NewFileManager(ctx, baseDir, logger.LoggerDefault{})

	return fm, testFilePath, func() {
		cancel()
		time.Sleep(100 * time.Millisecond)
	}
}

func TestFileStorageLifecycle(t *testing.T) {
	fm, srcPath, cleanup := setupTestFileManager(t)
	defer cleanup()
	const testFilename = "test.go"

	t.Run("StoreFile", func(t *testing.T) {
		info, err := fm.StoreFile(context.Background(), srcPath, testFilename)
		assert.NoError(t, err, "file store failed")
		assert.Equal(t, testFilename, info.Name, "file name not match")
		assert.Greater(t, info.Size, int64(0), "file size is zero")
	})

	t.Run("GetFileInfo", func(t *testing.T) {
		info, err := fm.GetFileInfo(context.Background(), testFilename)
		assert.NoError(t, err, "file info get failed")
		assert.Equal(t, testFilename, info.Name, "file name not match")
		assert.FileExists(t, info.FullPath, "file not exists")
	})

	t.Run("DownloadFile", func(t *testing.T) {
		content, info, err := fm.DownloadFile(context.Background(), testFilename)
		assert.NoError(t, err, "download file failed")
		reader, err := content.Content(context.Background())
		assert.NoError(t, err, "reder file failed")
		defer reader.Close()

		data, err := io.ReadAll(reader)
		fmt.Println(string(data))
		assert.NoError(t, err, "read file failed")
		assert.Equal(t, []byte("hello"), data, "file content not match")
		assert.Equal(t, testFilename, info.Name, "file name not match")
	})

	t.Run("FileExists", func(t *testing.T) {
		exists := fm.FileExists(context.Background(), testFilename)
		assert.True(t, exists, "file not exists")
	})

	t.Log("wait GC...")
	time.Sleep(30 * time.Second)
}

func TestLargeFileStorageLifecycle(t *testing.T) {
	fm, srcPath, cleanup := setupTestFileManager(t)
	defer cleanup()

	expectedMD5 := calculateFileMD5(srcPath)

	t.Run("StoreFile", func(t *testing.T) {
		info, err := fm.StoreFile(context.Background(), srcPath, testFilename)
		assert.NoError(t, err, "failed to store file")
		assert.Equal(t, testFilename, info.Name, "file name mismatch")
		assert.Greater(t, info.Size, int64(0), "file size is zero")
	})

	t.Run("GetFileInfo", func(t *testing.T) {
		info, err := fm.GetFileInfo(context.Background(), testFilename)
		assert.NoError(t, err, "failed to get file info")
		assert.Equal(t, testFilename, info.Name, "file name mismatch")
		assert.FileExists(t, info.FullPath, "file does not exist")
		assert.Equal(t, expectedMD5, info.MD5, "MD5 hash mismatch")
	})

	t.Run("DownloadFile", func(t *testing.T) {
		content, info, err := fm.DownloadFile(context.Background(), testFilename)
		assert.NoError(t, err, "failed to download file")
		assert.Equal(t, testFilename, info.Name, "file name mismatch")

		reader, err := content.Content(context.Background())
		assert.NoError(t, err, "failed to get file content")
		defer reader.Close()

		dataMD5 := calculateReaderMD5(reader)
		assert.Equal(t, expectedMD5, dataMD5, "MD5 hash mismatch")
	})

	t.Run("FileExists", func(t *testing.T) {
		exists := fm.FileExists(context.Background(), testFilename)
		assert.True(t, exists, "file should exist")
	})
}

// calculateFileMD5 computes MD5 checksum for a file.
func calculateFileMD5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	return calculateReaderMD5(file)
}

// calculateReaderMD5 computes MD5 checksum for an io.Reader.
func calculateReaderMD5(reader io.Reader) string {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
