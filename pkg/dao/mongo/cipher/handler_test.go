/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cipher

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")))
}

var once = sync.Once{}

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) {
	testDatas := []*types.Cipher{
		{
			Name:        "test-key-1",
			KeyType:     "rsa",
			Description: "unit test key",
			PrivateKey:  []byte("test-private-key-content-1"),
			PublicKey:   []byte("test-public-key-content-1"),
		},
		{
			Name:        "test-key-2",
			KeyType:     "ecc",
			Description: "unit test key 2",
			PrivateKey:  []byte("test-private-key-content-2"),
			PublicKey:   []byte("test-public-key-content-2"),
		},
	}

	once.Do(func() {
		h := testClient(t)
		for _, data := range testDatas {
			err := h.Create(nCtx, data)
			if err != nil {
				t.Errorf("prepareData() error = %v", err)
			}
		}
	})
}

// Test_handler_Get tests the Get method of the handler.
func Test_handler_Get(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx    contextx.IContext
		name    string
		keyType types.CipherKeyType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *types.Cipher
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				name:    "test-key-1",
				keyType: "rsa",
			},
			wantErr: false,
			want: &types.Cipher{
				Name:        "test-key-1",
				KeyType:     "rsa",
				Description: "unit test key",
				PrivateKey:  []byte("test-private-key-content-1"),
				PublicKey:   []byte("test-public-key-content-1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Get(tt.args.nCtx, tt.args.name, tt.args.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", got)
		})
	}
}

// Test_handler_Exist tests the Exist method of the handler.
func Test_handler_Exist(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx    contextx.IContext
		name    string
		keyType types.CipherKeyType
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				name:    "test-key-1",
				keyType: "rsa",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "non exist",
			args: args{
				nCtx:    nCtx,
				name:    "non-exist-key",
				keyType: "rsa",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Exist(tt.args.nCtx, tt.args.name, tt.args.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_handler_List tests the List method of the handler.
func Test_handler_List(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx    contextx.IContext
		name    string
		keyType types.CipherKeyType
	}
	tests := []struct {
		name    string
		args    args
		want    []*types.Cipher
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				name:    "test-key-1",
				keyType: "rsa",
			},
			want: []*types.Cipher{
				{
					Name:        "test-key-1",
					KeyType:     "rsa",
					Description: "unit test key",
					PrivateKey:  []byte("test-private-key-content-1"),
					PublicKey:   []byte("test-public-key-content-1"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, _, err := h.List(nCtx, types.UnlimitedPage(), WithName(tt.args.name), WithKeyType(string(tt.args.keyType)))
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("item: %+v", item)
			}
		})
	}
}

// Test_handler_ListWithoutCount tests the ListWithoutCount method of the handler.
func Test_handler_ListWithoutCount(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx    contextx.IContext
		name    string
		keyType types.CipherKeyType
	}
	tests := []struct {
		name    string
		args    args
		want    []*types.Cipher
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				name:    "test-key-1",
				keyType: "rsa",
			},
			want: []*types.Cipher{
				{
					Name:        "test-key-1",
					KeyType:     "rsa",
					Description: "unit test key",
					PrivateKey:  []byte("test-private-key-content-1"),
					PublicKey:   []byte("test-public-key-content-1"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListWithoutCount(
				nCtx, types.UnlimitedPage(),
				WithName(tt.args.name), WithKeyType(string(tt.args.keyType)),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListWithoutCount() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("item: %+v", item)
			}
		})
	}
}

// Test_handler_Count tests the Count method of the handler.
func Test_handler_Count(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx    contextx.IContext
		name    string
		keyType types.CipherKeyType
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				name:    "test-key-1",
				keyType: "rsa",
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Count(nCtx, WithName(tt.args.name), WithKeyType(string(tt.args.keyType)))
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Count() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_handler_Create_Get_Exist_Delete tests basic CRUD flow.
func Test_handler_Create(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx contextx.IContext
		ae   *types.Cipher
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				ae: &types.Cipher{
					Name:        "test-key-3",
					KeyType:     "ed25519",
					Description: "unit test key 3",
					PrivateKey:  []byte("test-private-key-content-3"),
					PublicKey:   []byte("test-public-key-content-3"),
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate",
			args: args{
				nCtx: nCtx,
				ae: &types.Cipher{
					Name:        "test-key-1",
					KeyType:     "rsa",
					Description: "unit test key duplicate",
					PrivateKey:  []byte("test-private-key-content-duplicate"),
					PublicKey:   []byte("test-public-key-content-duplicate"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.Create(tt.args.nCtx, tt.args.ae)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_Delete tests the Delete method of the handler.
func Test_handler_Delete(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx    contextx.IContext
		name    string
		keyType types.CipherKeyType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				name:    "test-key-2",
				keyType: "rsa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.Delete(tt.args.nCtx, tt.args.name, tt.args.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
		})
	}
}
