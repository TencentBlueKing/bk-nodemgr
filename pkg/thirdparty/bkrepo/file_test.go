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

package bkrepo

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
)

// TestParseNodeTime pins which node timestamps parse; unknown formats stay zero.
func TestParseNodeTime(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "zone-less-millis", value: "2026-09-23T19:10:53.113", want: "2026-09-23T19:10:53.113Z"},
		{name: "zone-less-seconds", value: "2026-09-23T19:10:53", want: "2026-09-23T19:10:53Z"},
		{name: "empty", value: "", want: ""},
		// BKRepo never sends a timezone offset, so even a qualified timestamp is unknown.
		{name: "zone-ful", value: "2026-05-06T12:00:00Z", want: ""},
		{name: "space-separated", value: "2026-05-06 12:00:00", want: ""},
		{name: "date-only", value: "2026-05-06", want: ""},
		{name: "epoch-millis", value: "1756123456789", want: ""},
		{name: "garbage", value: "not-a-time", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNodeTime(tt.value)
			if tt.want == "" {
				if !got.IsZero() {
					t.Fatalf("parseNodeTime(%q) = %s, want zero time", tt.value, got)
				}

				return
			}

			want, err := time.Parse(time.RFC3339Nano, tt.want)
			if err != nil {
				t.Fatalf("invalid expectation %q: %v", tt.want, err)
			}
			if !got.Equal(want) {
				t.Fatalf("parseNodeTime(%q) = %s, want %s", tt.value, got, want)
			}
		})
	}
}

func TestFileInfoModTime(t *testing.T) {
	file := File{info: NodeInfo{Name: "agent.tgz", LastModifiedDate: "2026-09-23T19:10:53.113"}}
	want := time.Date(2026, 9, 23, 19, 10, 53, 113000000, time.UTC)
	if got := file.Info().ModTime; !got.Equal(want) {
		t.Fatalf("Info().ModTime = %s, want %s", got, want)
	}

	// A payload without a timestamp must not look like an old artifact.
	missing := File{info: NodeInfo{Name: "agent.tgz"}}
	if got := missing.Info().ModTime; !got.IsZero() {
		t.Fatalf("Info().ModTime without lastModifiedDate = %s, want zero time", got)
	}
}

// Test_FileModTime checks real BKRepo payloads on the paths cleanup reads.
func Test_FileModTime(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))
	client := testClient(t)

	group, err := client.EnsureFileGroup(nCtx, "/unittest")
	if err != nil {
		t.Fatal(err)
	}

	name := fmt.Sprintf("modtime-%d.txt", time.Now().UnixNano())
	if err = group.Store(
		nCtx, fileiface.FileInfo{Name: name},
		io.NopCloser(strings.NewReader("modtime")), true,
	); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := group.Remove(nCtx, name); err != nil {
			t.Logf("failed to remove %s: %v", name, err)
		}
	}()

	check := func(scene, raw string, info fileiface.FileInfo) {
		if info.ModTime.IsZero() {
			t.Fatalf("%s: ModTime is zero for lastModifiedDate(%q), parseNodeTime misses the real format", scene, raw)
		}
		// Zone-less payloads leave room for the source UTC offset.
		if delta := time.Since(info.ModTime); delta < -26*time.Hour || delta > 26*time.Hour {
			t.Fatalf("%s: ModTime %s is not on the current time scale (delta %s)", scene, info.ModTime, delta)
		}
	}

	file, err := group.GetFile(nCtx, name)
	if err != nil {
		t.Fatal(err)
	}

	stored, ok := file.(*File)
	if !ok {
		t.Fatalf("unexpected file type %T", file)
	}
	t.Logf("bkrepo lastModifiedDate=%q createdDate=%q", stored.info.LastModifiedDate, stored.info.CreatedDate)
	check("GetFile", stored.info.LastModifiedDate, file.Info())

	files, err := group.AllFiles(nCtx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range files {
		listed, ok := item.(*File)
		if !ok || listed.info.Name != name {
			continue
		}
		check("AllFiles", listed.info.LastModifiedDate, item.Info())

		return
	}
	t.Fatalf("stored file %s is missing from AllFiles", name)
}
