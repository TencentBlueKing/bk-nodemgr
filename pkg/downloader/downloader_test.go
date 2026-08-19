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

package downloader

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func TestDownloadPreservesEscapedPath(t *testing.T) {
	t.Parallel()

	content := []byte("agent package")
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.RequestURI() != "/download/a%2Fb" {
			http.Error(rw, "unexpected request URI", http.StatusBadRequest)
			return
		}

		_, _ = rw.Write(content)
	}))
	defer server.Close()

	downloader, err := New(config.Downloader{
		TraceService: config.TraceService{TraceServiceName: "downloader-test"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	nCtx := contextx.New(context.Background())
	file, err := downloader.Download(nCtx, server.URL+"/download/a%2Fb", DownloadOptions{
		Filename: "agent.tgz",
		Checksum: Checksum{
			Algorithm: ChecksumAlgorithmMD5,
			Value:     fmt.Sprintf("%x", md5.Sum(content)),
		},
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	reader, err := file.Content(nCtx)
	if err != nil {
		t.Fatalf("Content: %v", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("content = %q, want %q", string(got), string(content))
	}
}

func TestDownloadRejectsUnsafeFilename(t *testing.T) {
	t.Parallel()

	downloader, err := New(config.Downloader{
		TraceService: config.TraceService{TraceServiceName: "downloader-test"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tests := []struct {
		name     string
		filename string
	}{
		{name: "parent traversal", filename: "../../../etc/passwd"},
		{name: "absolute path", filename: "/tmp/passwd"},
		{name: "windows separator", filename: `dir\passwd`},
		{name: "dot", filename: "."},
		{name: "dot dot", filename: ".."},
	}

	nCtx := contextx.New(context.Background())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := downloader.Download(nCtx, "http://example.com/file", DownloadOptions{
				Filename: tt.filename,
				Checksum: Checksum{Algorithm: ChecksumAlgorithmMD5, Value: "d41d8cd98f00b204e9800998ecf8427e"},
			})
			if err != nil {
				if !errors.Is(err, ErrDownloadFailed) {
					t.Fatalf("Download error = %v, want ErrDownloadFailed", err)
				}

				return
			}

			t.Fatal("Download returned nil error for unsafe filename")
		})
	}
}

func TestValidateDownloadFilename(t *testing.T) {
	t.Parallel()

	filename, err := validateDownloadFilename("agent.tar.gz")
	if err != nil {
		t.Fatalf("validateDownloadFilename: %v", err)
	}

	if filename != "agent.tar.gz" {
		t.Fatalf("filename = %q, want agent.tar.gz", filename)
	}
}

func TestValidateDownloadURLRejectsIPv6Zone(t *testing.T) {
	t.Parallel()

	_, err := validateDownloadURL("https://[fe80::1%25eth0]/agent.tgz", nil, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "IPv6 zones") {
			t.Fatalf("validateDownloadURL error = %v, want IPv6 zone error", err)
		}

		return
	}

	t.Fatal("validateDownloadURL returned nil error for IPv6 zone")
}
