/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package crypter ...
package crypter

import (
	"reflect"
	"testing"
)

// testCrypter ...
func testCrypter(t *testing.T) Crypter {
	aes, err := NewAESCrypter([]byte("1234567890123456"))
	if err != nil {
		t.Fatal(err)
	}

	return aes
}

// TestAES_Crypter ...
func TestAES_Crypter(t *testing.T) {
	type args struct {
		plaintext []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				plaintext: []byte("cQtyA*3862zrGk"),
			},
			want:    []byte("cQtyA*3862zrGk"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cry := testCrypter(t)
			ciphertext, err := cry.Encrypt(tt.args.plaintext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			decodeText, err := cry.Decrypt(ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("plaintext: %s\n", tt.args.plaintext)
			t.Logf("ciphertext: %s\n", ciphertext)
			t.Logf("decodeText: %s\n", decodeText)

			if !reflect.DeepEqual(decodeText, tt.args.plaintext) {
				t.Errorf("Decrypt() got = %v, want %v", decodeText, tt.args.plaintext)
			}
		})
	}
}
