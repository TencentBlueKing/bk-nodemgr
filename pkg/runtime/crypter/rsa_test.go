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

// TestRSA_GenerateKeyPairPEMAndParse verifies PEM generation and parsing helpers.
func TestRSA_GenerateKeyPairPEMAndParse(t *testing.T) {
	privPEM, pubPEM, err := GenerateRSAKeyPairPEM(RSAKeySize2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairPEM() error = %v", err)
	}

	if len(privPEM) == 0 || len(pubPEM) == 0 {
		t.Fatalf("generated PEM data should not be empty")
	}

	priv, err := parseRSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		t.Fatalf("ParseRSAPrivateKeyFromPEM() error = %v", err)
	}

	pub, err := parseRSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		t.Fatalf("ParseRSAPublicKeyFromPEM() error = %v", err)
	}

	if priv.PublicKey.N.Cmp(pub.N) != 0 || priv.PublicKey.E != pub.E {
		t.Fatalf("parsed public key does not match private key's public part")
	}
}
