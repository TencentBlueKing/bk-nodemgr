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

// Package crypter ...
package crypter

import (
	"bytes"
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

func mustNewAESGCMCrypter(t *testing.T, key, aad []byte) Crypter {
	t.Helper()

	cry, err := NewAESGCMCrypter(key, aad)
	if err != nil {
		t.Fatalf("NewAESGCMCrypter() error = %v", err)
	}

	return cry
}

// TestNewAESGCMCrypter verifies constructor validation.
func TestNewAESGCMCrypter(t *testing.T) {
	if _, err := NewAESGCMCrypter(nil, []byte("test-purpose")); err == nil {
		t.Fatal("NewAESGCMCrypter() error = nil, want error")
	}
}

// TestAESGCMCrypter verifies authenticated round trips and randomized nonces.
func TestAESGCMCrypter(t *testing.T) {
	cry := mustNewAESGCMCrypter(t, []byte("test-secret"), []byte("test-purpose"))
	plaintext := []byte("export download token payload")

	firstCiphertext, err := cry.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	secondCiphertext, err := cry.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if bytes.Equal(firstCiphertext, secondCiphertext) {
		t.Fatal("Encrypt() produced identical ciphertexts, want unique nonces")
	}
	if firstCiphertext[0] != aesGCMVersion {
		t.Fatalf("Encrypt() version = %d, want %d", firstCiphertext[0], aesGCMVersion)
	}
	const gcmNonceAndTagSize = 12 + 16
	if len(firstCiphertext) != 1+gcmNonceAndTagSize+len(plaintext) {
		t.Fatalf("Encrypt() ciphertext length = %d, want %d", len(firstCiphertext), 1+gcmNonceAndTagSize+len(plaintext))
	}

	decrypted, err := cry.Decrypt(firstCiphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

// TestAESGCMDecryptRejectsInvalidCiphertext verifies authenticated failure paths.
func TestAESGCMDecryptRejectsInvalidCiphertext(t *testing.T) {
	key := []byte("test-secret")
	aad := []byte("test-purpose")
	cry := mustNewAESGCMCrypter(t, key, aad)
	ciphertext, err := cry.Encrypt([]byte("payload"))
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	tamperedCiphertext := bytes.Clone(ciphertext)
	tamperedCiphertext[len(tamperedCiphertext)-1] ^= 1
	unsupportedVersion := bytes.Clone(ciphertext)
	unsupportedVersion[0]++

	tests := []struct {
		name       string
		decrypter  Crypter
		ciphertext []byte
	}{
		{
			name:       "tampered authentication tag",
			decrypter:  cry,
			ciphertext: tamperedCiphertext,
		},
		{
			name:       "unsupported version",
			decrypter:  cry,
			ciphertext: unsupportedVersion,
		},
		{
			name:       "truncated ciphertext",
			decrypter:  cry,
			ciphertext: ciphertext[:len(ciphertext)-1],
		},
		{
			name:       "wrong key",
			decrypter:  mustNewAESGCMCrypter(t, []byte("different-secret"), aad),
			ciphertext: ciphertext,
		},
		{
			name:       "wrong additional data",
			decrypter:  mustNewAESGCMCrypter(t, key, []byte("different-purpose")),
			ciphertext: ciphertext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.decrypter.Decrypt(tt.ciphertext); err == nil {
				t.Fatal("Decrypt() error = nil, want error")
			}
		})
	}
}
