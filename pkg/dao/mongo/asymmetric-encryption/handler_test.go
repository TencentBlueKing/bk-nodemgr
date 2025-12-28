/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package asymmetricencryption

import (
	"context"
	"os"
	"reflect"
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
	testDatas := []*types.AsymmetricEncryption{
		{
			CipherType:  "rsa",
			KeyType:     "private",
			Description: "unit test key",
			Content:     []byte("test-content"),
		},
		{
			CipherType:  "rsa",
			KeyType:     "public",
			Description: "unit test key",
			Content:     []byte("test-content"),
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

// Test_handler_Create_Get_Exist_Delete tests basic CRUD flow.
func Test_handler_Get(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx       contextx.IContext
		cipherType types.AsymmetricCipherType
		keyType    types.AsymmetricKeyType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *types.AsymmetricEncryption
	}{
		{
			name: "normal",
			args: args{
				nCtx:       nCtx,
				cipherType: "rsa",
				keyType:    "private",
			},
			wantErr: false,
			want: &types.AsymmetricEncryption{
				CipherType:  "rsa",
				KeyType:     "private",
				Description: "unit test key",
				Content:     []byte("test-content"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Get(tt.args.nCtx, tt.args.keyType, tt.args.cipherType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}

			if got.CipherType != tt.want.CipherType || got.KeyType != tt.want.KeyType ||
				got.Description != tt.want.Description || !reflect.DeepEqual(got.Content, tt.want.Content) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
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
		ae   *types.AsymmetricEncryption
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
				ae: &types.AsymmetricEncryption{
					CipherType:  "ed25519",
					KeyType:     "private",
					Description: "unit test key",
					Content:     []byte("test-content"),
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate",
			args: args{
				nCtx: nCtx,
				ae: &types.AsymmetricEncryption{
					CipherType:  "rsa",
					KeyType:     "public",
					Description: "unit test key 2",
					Content:     []byte("test-content-2"),
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

func Test_handler_Exist(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx       contextx.IContext
		cipherType types.AsymmetricCipherType
		keyType    types.AsymmetricKeyType
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
				nCtx:       nCtx,
				cipherType: "rsa",
				keyType:    "private",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "non exist",
			args: args{
				nCtx:       nCtx,
				cipherType: "ecc",
				keyType:    "private",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Exist(tt.args.nCtx, tt.args.keyType, tt.args.cipherType)
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

// Test_handler_Upsert tests the Upsert method of the handler.
func Test_handler_Upsert(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)
	type args struct {
		nCtx                 contextx.IContext
		asymmetricEncryption *types.AsymmetricEncryption
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
				asymmetricEncryption: &types.AsymmetricEncryption{
					CipherType:  "rsa",
					KeyType:     "private",
					Description: "unit test key updated",
					Content:     []byte("test-content-updated"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.Upsert(tt.args.nCtx, tt.args.asymmetricEncryption)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr: %v", err, tt.wantErr)
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
		nCtx       contextx.IContext
		cipherType types.AsymmetricCipherType
		keyType    types.AsymmetricKeyType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:       nCtx,
				cipherType: "rsa",
				keyType:    "public",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.Delete(tt.args.nCtx, tt.args.keyType, tt.args.cipherType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr: %v", err, tt.wantErr)
				return
			}
		})
	}
}
