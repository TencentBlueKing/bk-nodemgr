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
	"io"
	"os"
	"testing"

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

			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}

			if string(data) != tt.args.fileContent {
				t.Fatal("not equal")
			}
		})
	}
}
