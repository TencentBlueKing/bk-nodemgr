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
	"encoding/base64"
	"reflect"
	"testing"
)

// testRSACrypter creates an RSA-based crypter for testing.
func testRSACrypter(t *testing.T) Crypter {
	t.Helper()

	priv, _, err := GenerateRSAKeyPairPEM(RSAKeySize2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair() error = %v", err)
	}

	cry, err := NewRSACrypterFromPrivateKey(priv)
	if err != nil {
		t.Fatalf("NewRSACrypterFromPrivateKey() error = %v", err)
	}

	return cry
}

// TestRSA_Crypter ...
func TestRSA_Crypter(t *testing.T) {
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
				plaintext: []byte("example_test"),
			},
			want:    []byte("example_test"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cry := testRSACrypter(t)

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
			t.Logf("ciphertext: %x\n", ciphertext)
			t.Logf("decodeText: %s\n", decodeText)

			if !reflect.DeepEqual(decodeText, tt.args.plaintext) {
				t.Errorf("Decrypt() got = %v, want %v", decodeText, tt.args.plaintext)
			}
		})
	}
}

// TestDecryptRSABase64Ciphertext tests DecryptRSABase64Ciphertext with various scenarios.
func TestDecryptRSABase64Ciphertext(t *testing.T) {
	cry := testRSACrypter(t)

	// Prepare valid base64 ciphertext for normal path
	plaintext := []byte("secret-password")
	ciphertext, err := cry.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	validBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	tests := []struct {
		name    string
		cry     Crypter
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "normal decryption path",
			cry:     cry,
			input:   validBase64,
			want:    string(plaintext),
			wantErr: false,
		},
		{
			name:    "nil Crypter",
			cry:     nil,
			input:   validBase64,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid base64 input",
			cry:     cry,
			input:   "not-valid-base64",
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid base64 but RSA decryption failed",
			cry:     cry,
			input:   base64.StdEncoding.EncodeToString([]byte{RSAVersion, 0x01, 0x02}),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptRSABase64Ciphertext(tt.cry, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptRSABase64Ciphertext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecryptRSABase64Ciphertext() got = %q, want %q", got, tt.want)
			}
		})
	}
}
