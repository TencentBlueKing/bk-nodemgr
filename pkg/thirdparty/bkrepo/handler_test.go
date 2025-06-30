package bkrepo

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/joho/godotenv"
)

type testHeaderSetter struct{}

// GetAuthHeader ...
func (testHeaderSetter) GetAuthHeader() (string, error) {
	return os.Getenv("BK_REPO_AUTHHEADER"), nil
}

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery("bkrepo", []string{os.Getenv("BK_REPO_ENDPOINT")}),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               logger.LoggerDefault{},
	}

	h, err := New(clientCap, &Config{
		RepoName:  os.Getenv("BK_REPO_REPONAME"),
		ProjectID: os.Getenv("BK_REPO_PROJECTID"),
		Username:  os.Getenv("BK_REPO_USERNAME"),
		Password:  os.Getenv("BK_REPO_PASSWORD"),
	}, WithLogger(logger.LoggerDefault{}))
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_EnsureFileGroup tests EnsureFileGroup.
func Test_EnsureFileGroup(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")
	client := testClient(t)

	_, err := client.EnsureFileGroup(ctx, "/unittest")
	if err != nil {
		t.Fatal(err)
	}
}

// Test_Store tests Store.
func Test_Store(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")
	client := testClient(t)

	type args struct {
		ctx           context.Context
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
				ctx:           nil,
				fileGroupName: "/unittest",
				fileName:      "test.txt",
				fileContent:   "test",
			},
			wantErr: true,
		},
		{
			name: "test-1",
			args: args{
				ctx:           ctx,
				fileGroupName: "/unittest",
				fileName:      "test-1.txt",
				fileContent:   "test",
			},
			wantErr: false,
		},
		{
			name: "test-2",
			args: args{
				ctx:           ctx,
				fileGroupName: "/unittest",
				fileName:      "test-2.txt",
				fileContent:   "test",
			},
			wantErr: false,
		},
		{
			name: "test-3",
			args: args{
				ctx:           ctx,
				fileGroupName: "/unittest",
				fileName:      "test-3.txt",
				fileContent:   "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := client.EnsureFileGroup(tt.args.ctx, tt.args.fileGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureFileGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			err = group.Store(tt.args.ctx, iface.FileInfo{
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
	ctx, _ := tenant.SetID(context.Background(), "single")
	client := testClient(t)

	type args struct {
		ctx           context.Context
		fileGroupName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				ctx:           ctx,
				fileGroupName: "/unittest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := client.GetFileGroup(tt.args.ctx, tt.args.fileGroupName)
			if err != nil {
				t.Fatal(err)
			}

			files, err := group.AllFiles(tt.args.ctx)
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
	ctx, _ := tenant.SetID(context.Background(), "single")
	client := testClient(t)

	type args struct {
		ctx           context.Context
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
				ctx:           ctx,
				fileGroupName: "/unittest",
				fileName:      "test-1.txt",
				fileContent:   "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := client.GetFileGroup(tt.args.ctx, tt.args.fileGroupName)
			if err != nil {
				t.Fatal(err)
			}

			file, err := group.GetFile(tt.args.ctx, tt.args.fileName)
			if err != nil {
				t.Fatal(err)
			}

			reader, err := file.Content(tt.args.ctx)
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
