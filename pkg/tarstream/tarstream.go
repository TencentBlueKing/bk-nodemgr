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

// Package tarstream provides utilities for streaming files into tar archives.
package tarstream

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// AddStreamFileToTar writes a streamed file (with known Content-Length from HTTP headers)
// into a tar archive. The pathPrefix is prepended to fileName in the tar entry name.
func AddStreamFileToTar(
	tw *tar.Writer,
	pathPrefix string,
	fileName string,
	reader io.Reader,
	headers http.Header,
	mode int64,
) error {

	contentLength, err := ParseContentLength(headers)
	if err != nil {
		return fmt.Errorf("failed to parse file content length from headers: %w", err)
	}

	if contentLength <= 0 {
		return fmt.Errorf(
			"missing or zero Content-Length for streamed file %s; "+
				"tar format requires a known size — ensure upstream returns Content-Length header",
			fileName,
		)
	}

	if err := tw.WriteHeader(&tar.Header{
		Name:    fmt.Sprintf("%s/%s", pathPrefix, fileName),
		Mode:    mode,
		Size:    contentLength,
		ModTime: time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", fileName, err)
	}

	if _, err := io.CopyN(tw, reader, contentLength); err != nil {
		return fmt.Errorf("failed to stream %s into tar: %w", fileName, err)
	}

	return nil
}

// AddTextFileToTar writes an in-memory text content as a file into a tar archive.
func AddTextFileToTar(tw *tar.Writer, pathPrefix string, fileName string, content []byte, mode int64) error {
	if err := tw.WriteHeader(&tar.Header{
		Name:    fmt.Sprintf("%s/%s", pathPrefix, fileName),
		Mode:    mode,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", fileName, err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write content for %s: %w", fileName, err)
	}

	return nil
}

// ParseContentLength extracts and validates the Content-Length header value.
func ParseContentLength(headers http.Header) (int64, error) {
	if headers == nil {
		return 0, nil
	}
	contentLen := headers.Get("Content-Length")
	if contentLen == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseInt(contentLen, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid content length(%s): %w", contentLen, err)
	}

	return parsed, nil
}
