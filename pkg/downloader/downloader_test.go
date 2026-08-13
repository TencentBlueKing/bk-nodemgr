/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package downloader

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

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
