/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this LocalFile except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package local

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/spf13/afero"
)

// NewLocalFile creates a new LocalFile.
func NewLocalFile(fullPath string) (*LocalFile, error) {
	exists, err := afero.Exists(rFs(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists, err: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("file does not exist, fullPath(%s)", fullPath)
	}

	// check if it's a directory
	info, err := rFs().Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info, err: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("invalid path, path is a directory, fullPath(%s)", fullPath)
	}

	fileName := filepath.Base(fullPath)
	md5Str, err := MD5SumWithBuffer(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file md5, err: %w", err)
	}

	stat, err := rFs().Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info, err: %w", err)
	}

	file := &LocalFile{
		info: iface.FileInfo{
			Name:         fileName,
			Size:         stat.Size(),
			Md5:          md5Str,
			Description:  "",
			ExtendFields: nil,
		},
		fullPath: fullPath,
	}

	return file, nil
}

// MD5SumWithBuffer returns the MD5 sum of a file.
func MD5SumWithBuffer(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file failed, err: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 32*1024)
	hash := md5.New()
	reader := bufio.NewReader(file)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if _, err := hash.Write(buffer[:n]); err != nil {
				return "", fmt.Errorf("write to hash failed, err: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read file failed, err: %w", err)
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// LocalFile represents a local file.
type LocalFile struct {
	info     iface.FileInfo
	fullPath string
}

// Content returns LocalFile content.
func (f *LocalFile) Content(_ context.Context) (io.ReadCloser, error) {
	absPath, err := filepath.Abs(f.fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve absolute path: %w", err)
	}

	// 打开文件
	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local file not found: %w", err)
		}

		return nil, fmt.Errorf("failed to open local file: %w", err)
	}

	return file, nil
}

// Info returns LocalFile info.
func (f *LocalFile) Info(_ context.Context) (iface.FileInfo, error) {
	return f.info, nil
}

// FileObject returns LocalFile file object.
func (f *LocalFile) FileObject() iface.FileObject {
	return iface.LocalFile
}

// Name returns LocalFile name.
func (f *LocalFile) Name() string {
	return f.info.Name
}
