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

package utils

import (
	"context"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
)

func TestBuildDownloadServerURLs(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("tenant-a"))
	endpoints := []discover.Endpoint{
		{IPV4: "127.0.0.1", IPV6: "::1", Port: 8080},
		{IPV4: "127.0.0.2", Port: 8081},
	}

	got, err := BuildDownloadServerURLs(nCtx, endpoints...)
	if err != nil {
		t.Fatalf("failed to build download server urls: %v", err)
	}

	addrs := strings.Split(got, installer.ServerAddrSeparator)
	if len(addrs) != 3 {
		t.Fatalf("expected 3 addrs, got %d: %q", len(addrs), got)
	}

	want := map[string]bool{
		"http://127.0.0.1:8080/tenant-a": false,
		"http://[::1]:8080/tenant-a":     false,
		"http://127.0.0.2:8081/tenant-a": false,
	}
	for _, addr := range addrs {
		if _, ok := want[addr]; !ok {
			t.Fatalf("unexpected addr %q from %q", addr, got)
		}

		want[addr] = true
	}
	for addr, seen := range want {
		if !seen {
			t.Fatalf("expected addr %q from %q", addr, got)
		}
	}
}

func TestBuildDownloadServerURLs_RejectsMissingTenant(t *testing.T) {
	nCtx := contextx.New(context.Background())
	_, err := BuildDownloadServerURLs(nCtx, discover.Endpoint{IPV4: "127.0.0.1", Port: 8080})
	if err != nil {
		return
	}

	t.Fatal("expected error")
}

func TestSelectOneDownloadServerV4URL(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("tenant-a"))
	got, err := SelectOneDownloadServerV4URL(nCtx, []discover.Endpoint{{IPV4: "127.0.0.1", Port: 8080}})
	if err != nil {
		t.Fatalf("failed to select download server url: %v", err)
	}
	if got != "http://127.0.0.1:8080/tenant-a" {
		t.Fatalf("expected tenant-prefixed url, got %q", got)
	}
}
