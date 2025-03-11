/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package downloader this package provide a downloader to download files.
package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// Config the config for downloading.
type Config struct {
	URL          string
	Method       string
	Headers      map[string]string
	RequestBody  interface{}
	DestPath     string
	Timeout      time.Duration
	ProgressFunc func(int64, int64)
}

// Downloader downloader interface.
type Downloader interface {
	Download(ctx context.Context, config Config) error
}

// HTTPDownloader this is an implementation of Downloader.
type HTTPDownloader struct{}

// progressTracker tracing the progress of downloading.
type progressTracker struct {
	reader     io.Reader
	size       int64
	downloaded int64
	onProgress func(current int64, total int64)
}

// Read implements the io.Reader interface.
func (pt *progressTracker) Read(p []byte) (int, error) {
	n, err := pt.reader.Read(p)
	if n > 0 {
		pt.downloaded += int64(n)
		if pt.onProgress != nil {
			pt.onProgress(pt.downloaded, pt.size)
		}
	}

	return n, err
}

// Download download files.
func (d *HTTPDownloader) Download(ctx context.Context, config Config) error {
	// make sure the destination directory exists.
	destDir := filepath.Dir(config.DestPath)
	if err := utils.TryCreateDir(destDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// set up the HTTP client.
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          10,               // nolint: mnd
			IdleConnTimeout:       30 * time.Second, // nolint: mnd
			TLSHandshakeTimeout:   10 * time.Second, // nolint: mnd
			ExpectContinueTimeout: 1 * time.Second,  // nolint: mnd
		},
	}

	var reqBody io.Reader
	if config.RequestBody != nil {
		reqBodyBytes, err := json.Marshal(config.RequestBody)
		if err != nil {
			return fmt.Errorf("serialize request body failed: %v", err)
		}
		reqBody = bytes.NewBuffer(reqBodyBytes)
	}

	// create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, reqBody)
	if err != nil {
		return fmt.Errorf("create HTTP request failed: %v", err)
	}

	// set headers.
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// execute the HTTP request.
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute HTTP request failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// check the HTTP status code.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("HTTP request returned non-200 status code: %d, response: %s", resp.StatusCode, string(body))
	}

	// write the response body to a file.
	file, err := os.Create(config.DestPath)
	if err != nil {
		return fmt.Errorf("create destination file %s failed: %v", config.DestPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// track the progress of the download.
	progressReader := &progressTracker{
		reader:     resp.Body,
		size:       resp.ContentLength,
		onProgress: config.ProgressFunc,
	}

	_, err = io.Copy(file, progressReader)
	elapsed := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("write to destination file failed: %v", err)
	}

	// check if the download took too long.
	if elapsed > config.Timeout {
		return fmt.Errorf("download timed out, took %.2f seconds", elapsed.Seconds())
	}

	return nil
}
