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

package counter

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func testClient(t *testing.T) Handler {
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

// Test_handler_Generate generate sequence.
func Test_handler_Generate(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "test-1",
			key:     "key-1",
			wantErr: false,
		},
		{
			name:    "test-2",
			key:     "key-1",
			wantErr: false,
		},
		{
			name:    "test-3",
			key:     "key-1",
			wantErr: false,
		},
		{
			name:    "test-4",
			key:     "key-1",
			wantErr: false,
		},
		{
			name:    "test-5",
			key:     "key-1",
			wantErr: false,
		},
		{
			name:    "test-6",
			key:     "key-2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Generate(nCtx, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Generate() got = %v", got)
		})
	}
}

// Test_handler_GenerateN reserves n consecutive sequence values in one call.
func Test_handler_GenerateN(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	tests := []struct {
		name    string
		key     string
		n       int64
		wantErr bool
	}{
		{
			name:    "reserve 5 consecutive values",
			key:     "key-generate-n",
			n:       5,
			wantErr: false,
		},
		{
			name:    "reserve 1 is equivalent to Generate",
			key:     "key-generate-n",
			n:       1,
			wantErr: false,
		},
		{
			name:    "n=0 should fail",
			key:     "key-generate-n",
			n:       0,
			wantErr: true,
		},
		{
			name:    "negative n should fail",
			key:     "key-generate-n",
			n:       -1,
			wantErr: true,
		},
		{
			name:    "empty key should fail",
			key:     "",
			n:       3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GenerateN(nCtx, tt.key, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("GenerateN(n=%d) got = %v", tt.n, got)
		})
	}
}
