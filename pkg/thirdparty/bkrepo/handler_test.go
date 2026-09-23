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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	fileiface "github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/joho/godotenv"
)

type testHeaderSetter struct{}

// GetAuthHeader ...
func (testHeaderSetter) GetAuthHeader() string {
	return os.Getenv("BK_REPO_AUTHHEADER")
}

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("bkrepo", []string{os.Getenv("BK_REPO_ENDPOINT")}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             bkrepoTestTraceService{},
	}

	h, err := New(clientCap, &Config{
		RepoName:  os.Getenv("BK_REPO_REPONAME"),
		ProjectID: os.Getenv("BK_REPO_PROJECTID"),
		Username:  os.Getenv("BK_REPO_USERNAME"),
		Password:  os.Getenv("BK_REPO_PASSWORD"),
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_EnsureFileGroup tests EnsureFileGroup.
func Test_EnsureFileGroup(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))
	client := testClient(t)

	_, err := client.EnsureFileGroup(nCtx, "/unittest")
	if err != nil {
		t.Fatal(err)
	}
}

// Test_Store tests Store.
func Test_Store(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))
	client := testClient(t)

	type args struct {
		nCtx          contextx.IContext
		fileGroupName string
		fileName      string
		fileContent   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-0",
			args: args{
				nCtx:          nil,
				fileGroupName: "/unittest",
				fileName:      "test.txt",
				fileContent:   "test",
			},
			wantErr: true,
		},
		{
			name: "test-1",
			args: args{
				nCtx:          nCtx,
				fileGroupName: "/unittest",
				fileName:      "test-1.txt",
				fileContent:   "test",
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				nCtx:          nCtx,
				fileGroupName: "/unittest",
				fileName:      "test-2.txt",
				fileContent:   "test",
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				nCtx:          nCtx,
				fileGroupName: "/unittest",
				fileName:      "test-3.txt",
				fileContent:   "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := client.EnsureFileGroup(tt.args.nCtx, tt.args.fileGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureFileGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			err = group.Store(tt.args.nCtx, fileiface.FileInfo{
				Name: tt.args.fileName,
			}, io.NopCloser(bytes.NewReader([]byte(tt.args.fileContent))), true)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Test_CopyAndRemove tests Copy and Remove.
func Test_CopyAndRemove(t *testing.T) {
	nCtx := contextx.New(t.Context(), contextx.WithTenantID("single"))
	client := testClient(t)

	rootGroup, err := client.EnsureFileGroup(nCtx, "/unittest")
	if err != nil {
		t.Fatal(err)
	}

	runDir := fmt.Sprintf("copy-remove-%d", time.Now().UnixNano())
	removed := false
	t.Cleanup(func() {
		if removed {
			return
		}

		cleanupCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))
		if err := rootGroup.Remove(cleanupCtx, runDir); err != nil {
			t.Errorf("failed to clean up BKRepo integration test directory %q: %v", runDir, err)
		}
	})

	sourceGroup, err := client.EnsureFileGroup(nCtx, "/unittest/"+runDir+"/system/release/plugin")
	if err != nil {
		t.Fatal(err)
	}
	destinationGroup, err := client.EnsureFileGroup(nCtx, "/unittest/"+runDir+"/tenant-a/release/plugin")
	if err != nil {
		t.Fatal(err)
	}

	const fileName = "source.tgz"
	const content = "copy-remove-integration-test"
	err = sourceGroup.Store(
		nCtx,
		fileiface.FileInfo{Name: fileName},
		io.NopCloser(bytes.NewReader([]byte(content))),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	sourceFile, err := sourceGroup.GetFile(nCtx, fileName)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("source file: info=%+v, absDirs=%v", sourceFile.Info(), sourceFile.AbsDirs())

	if err := sourceGroup.Copy(nCtx, fileName, destinationGroup, ".", false); err != nil {
		t.Fatal(err)
	}

	copiedFile, err := destinationGroup.GetFile(nCtx, fileName)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("copied file: info=%+v, absDirs=%v", copiedFile.Info(), copiedFile.AbsDirs())

	reader, err := copiedFile.Content(nCtx)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Fatalf("copied file content = %q, want %q", got, content)
	}

	if err := rootGroup.Remove(nCtx, runDir); err != nil {
		t.Fatal(err)
	}
	removed = true
}

// Test_List tests List.
func Test_List(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))
	client := testClient(t)

	type args struct {
		nCtx          contextx.IContext
		fileGroupName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				nCtx:          nCtx,
				fileGroupName: "/unittest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := client.GetFileGroup(tt.args.nCtx, tt.args.fileGroupName)
			if err != nil {
				t.Fatal(err)
			}

			files, err := group.AllFiles(tt.args.nCtx)
			if err != nil {
				t.Fatal(err)
			}

			for _, file := range files {
				t.Log(file.Info().Name)
			}
		})
	}
}

func Test_Get(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))
	client := testClient(t)

	type args struct {
		nCtx          contextx.IContext
		fileGroupName string
		fileName      string
		fileContent   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				nCtx:          nCtx,
				fileGroupName: "/unittest",
				fileName:      "test-1.txt",
				fileContent:   "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := client.GetFileGroup(tt.args.nCtx, tt.args.fileGroupName)
			if err != nil {
				t.Fatal(err)
			}

			file, err := group.GetFile(tt.args.nCtx, tt.args.fileName)
			if err != nil {
				t.Fatal(err)
			}

			reader, err := file.Content(tt.args.nCtx)
			if err != nil {
				t.Fatal(err)
			}

			data, readErr := io.ReadAll(reader)
			closeErr := reader.Close()
			if readErr != nil {
				t.Fatal(readErr)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}

			if string(data) != tt.args.fileContent {
				t.Fatal("not equal")
			}
		})
	}
}
