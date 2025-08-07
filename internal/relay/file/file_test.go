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

// Setup a test file manager without creating test file
func setupTestFileManager(t *testing.T) (IFileManager, string, func()) {
	baseDir := "/data/home/coonnerlian/workspace/tmp/"
	filedir := "/data/home/coonnerlian/workspace/"
	testFileName := "gse_agent_inner-v2.1.6-beta.59.tgz"
	testFilePath := filepath.Join(filedir, testFileName)

	// Create a file manager
	ctx, cancel := context.WithCancel(context.Background())
	fm := NewFileManager(ctx, baseDir, logger.LoggerDefault{})

	return fm, testFilePath, func() {
		cancel()
	}
}

// CalculateFileMD5 computes the MD5 checksum for a file
func calculateFileMD5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Tests
func TestFileStorageLifecycle(t *testing.T) {
	fm, srcPath, cleanup := setupTestFileManager(t)
	defer cleanup()

	t.Run("StoreFile", func(t *testing.T) {
		info, err := fm.StoreFile(context.Background(), srcPath, testFilename)
		assert.NoError(t, err, "file store failed")
		assert.Equal(t, testFilename, info.Name, "file name not match")
		assert.Greater(t, info.Size, int64(0), "file size is zero")
	})

	time.Sleep(10 * time.Second)

	t.Run("StoreFile", func(t *testing.T) {
		info, err := fm.StoreFile(context.Background(), srcPath, testFilename)
		assert.NoError(t, err, "file store failed")
		assert.Equal(t, testFilename, info.Name, "file name not match")
		assert.Greater(t, info.Size, int64(0), "file size is zero")
	})

	t.Run("DownloadFile", func(t *testing.T) {
		time.Sleep(10 * time.Second)
		file, err := fm.GetFile(context.Background(), testFilename)
		assert.NoError(t, err, "failed to download file")
		reader, err := file.Content(context.Background())
		assert.NoError(t, err, "failed to get file content")
		defer reader.Close()

		data, err := io.ReadAll(reader)
		assert.NoError(t, err, "read file failed")
		assert.Equal(t, []byte("hello"), data, "file content not match")
	})

	t.Run("FileExists", func(t *testing.T) {
		exists := fm.FileExists(context.Background(), testFilename, calculateFileMD5(srcPath))
		assert.True(t, exists, "file not exists")
	})
	t.Log("wait GC...")
	time.Sleep(5 * time.Second)
}
