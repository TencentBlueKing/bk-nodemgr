/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package crypter ...
package crypter

import (
	"bytes"
	"testing"
)

// newTestSM4Crypter ...
func newTestSM4Crypter(t *testing.T) Crypter {
	t.Helper()

	sm4, err := NewSM4Crypter([]byte("1234567890123456"))
	if err != nil {
		t.Fatal(err)
	}

	return sm4
}

// TestSM4_Crypter ...
func TestSM4_Crypter(t *testing.T) {
	type args struct {
		plaintext []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal",
			args: args{
				plaintext: []byte("cQtyA*3862zrGk"),
			},
		},
		{
			name: "chinese",
			args: args{
				plaintext: []byte("国密SM4测试-凭据"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cry := newTestSM4Crypter(t)
			ciphertext, err := cry.Encrypt(tt.args.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decodeText, err := cry.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decodeText, tt.args.plaintext) {
				t.Errorf("Decrypt() got = %v, want %v", decodeText, tt.args.plaintext)
			}
		})
	}
}

// TestSM4_CiphertextFormat checks the version byte and envelope layout:
// SM4Version(1) + Nonce(12) + ciphertext with tag(16).
func TestSM4_CiphertextFormat(t *testing.T) {
	cry := newTestSM4Crypter(t)

	plaintext := []byte("sm4-format-check")
	ciphertext, err := cry.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	if ciphertext[0] != SM4Version {
		t.Errorf("version byte = %d, want %d", ciphertext[0], SM4Version)
	}

	wantLen := 1 + sm4GCMNonceLen + len(plaintext) + sm4GCMTagLen
	if len(ciphertext) != wantLen {
		t.Errorf("ciphertext length = %d, want %d", len(ciphertext), wantLen)
	}
}

// TestSM4_RandomNonceCiphertext checks the random nonce design: encrypting
// the same plaintext twice yields different ciphertexts. No consumer relies
// on deterministic ciphertext (host credits are upserted by creditID).
func TestSM4_RandomNonceCiphertext(t *testing.T) {
	cry := newTestSM4Crypter(t)

	plaintext := []byte("random-nonce-check")
	first, err := cry.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	second, err := cry.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(first, second) {
		t.Error("encrypting the same plaintext twice produced identical ciphertexts, nonce is not random")
	}
}

// TestSM4_EmptyPlaintext ...
func TestSM4_EmptyPlaintext(t *testing.T) {
	cry := newTestSM4Crypter(t)

	if _, err := cry.Encrypt(nil); err == nil {
		t.Error("Encrypt(nil) expected error, got nil")
	}
}

// TestSM4_InvalidCiphertext ...
func TestSM4_InvalidCiphertext(t *testing.T) {
	cry := newTestSM4Crypter(t)

	// too short
	if _, err := cry.Decrypt([]byte{SM4Version}); err == nil {
		t.Error("Decrypt(too short) expected error, got nil")
	}

	// unsupported version byte
	ciphertext, err := cry.Encrypt([]byte("version-check"))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[0] = 99
	if _, err := cry.Decrypt(ciphertext); err == nil {
		t.Error("Decrypt(unsupported version) expected error, got nil")
	}
}

// TestSM4_TamperedCiphertext verifies GCM authenticity: any tampered byte
// must fail decryption.
func TestSM4_TamperedCiphertext(t *testing.T) {
	cry := newTestSM4Crypter(t)

	ciphertext, err := cry.Encrypt([]byte("tamper-check"))
	if err != nil {
		t.Fatal(err)
	}

	// flip a bit in the last ciphertext block.
	ciphertext[len(ciphertext)-1] ^= 0xFF
	if _, err := cry.Decrypt(ciphertext); err == nil {
		t.Error("Decrypt(tampered) expected error, got nil")
	}
}

// TestSM4_WrongKey verifies GCM authenticity: a wrong key fails decryption
// with an error (unlike the previous unauthenticated SM4-CTR design).
func TestSM4_WrongKey(t *testing.T) {
	encrypter := newTestSM4Crypter(t)

	decrypter, err := NewSM4Crypter([]byte("another-key-1234"))
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := encrypter.Encrypt([]byte("wrong-key-check"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := decrypter.Decrypt(ciphertext); err == nil {
		t.Error("Decrypt with wrong key expected error, got nil")
	}
}

// TestSM4_CustomSalt ...
func TestSM4_CustomSalt(t *testing.T) {
	plaintext := []byte("custom-salt-check")

	defaultCrypter := newTestSM4Crypter(t)
	saltedCrypter, err := NewSM4Crypter([]byte("1234567890123456"), WithSM4Salt([]byte("custom-salt")))
	if err != nil {
		t.Fatal(err)
	}

	defaultCt, err := defaultCrypter.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	saltedCt, err := saltedCrypter.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(defaultCt, saltedCt) {
		t.Error("different salts should derive different keys and produce different ciphertexts")
	}
}
