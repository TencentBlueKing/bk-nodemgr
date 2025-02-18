/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/joho/godotenv"
)

// testAction ...
func testAction(t *testing.T) (crypter.Crypter, operengine.ActionDef) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	crypt, err := crypter.NewAESCrypter([]byte(os.Getenv("ENCRYPT_KEY")))
	if err != nil {
		t.Fatal(err)
	}

	return crypt, NewActionSshHostExecCmd(crypt, logger.LoggerDefault{})
}

// Test_sshHostExecCmd_Do ...
func Test_sshHostExecCmd_Do(t *testing.T) {
	crypt, action := testAction(t)
	ciphertext, err := crypt.Encrypt([]byte("cQtyA*3862zrGk"))
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx *operengine.ActionInstContext
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: context.Background(),
					Data: &operengine.ActionInstData{
						Content: map[string]any{
							"ip":      "127.0.0.1",
							"port":    36000,
							"user":    "root",
							"passwd":  ciphertext,
							"command": "uname -a",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = action.Do(tt.args.ctx)
			if err != nil {
				t.Logf("action.Do() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
