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
	"crypto/rand"
	"testing"

	"github.com/emmansun/gmsm/sm2"
)

// newTestSM2CrypterFromGeneratedKey ...
func newTestSM2CrypterFromGeneratedKey(t *testing.T) Crypter {
	t.Helper()

	privPEM, _, err := GenerateSM2KeyPairPEM()
	if err != nil {
		t.Fatal(err)
	}

	cry, err := NewSM2CrypterFromPrivateKey(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	return cry
}

// TestSM2_Crypter ...
func TestSM2_Crypter(t *testing.T) {
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
				plaintext: []byte("国密SM2测试-凭据"),
			},
		},
		{
			name: "large-plaintext-no-hybrid-needed",
			args: args{
				plaintext: bytes.Repeat([]byte("x"), 4096),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cry := newTestSM2CrypterFromGeneratedKey(t)
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

// TestSM2_CiphertextFormat checks the version byte prefix.
func TestSM2_CiphertextFormat(t *testing.T) {
	cry := newTestSM2CrypterFromGeneratedKey(t)

	ciphertext, err := cry.Encrypt([]byte("sm2-format-check"))
	if err != nil {
		t.Fatal(err)
	}

	if ciphertext[0] != SM2Version {
		t.Errorf("version byte = %d, want %d", ciphertext[0], SM2Version)
	}

	// ASN.1 DER encoding starts with the SEQUENCE tag 0x30.
	if ciphertext[1] != 0x30 {
		t.Errorf("ciphertext body should be ASN.1 DER, got first byte 0x%02x", ciphertext[1])
	}
}

// TestSM2_RawC1C3C2Ciphertext verifies decryption also accepts raw
// C1C3C2 (uncompressed 04-prefix) ciphertexts, which common frontend
// SM2 libraries produce.
func TestSM2_RawC1C3C2Ciphertext(t *testing.T) {
	privPEM, _, err := GenerateSM2KeyPairPEM()
	if err != nil {
		t.Fatal(err)
	}

	cry, err := NewSM2CrypterFromPrivateKey(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	sm2Cry, ok := cry.(*SM2)
	if !ok {
		t.Fatal("unexpected crypter type")
	}

	plaintext := []byte("raw-c1c3c2-check")
	// nil opts produces raw C1C3C2 ciphertext.
	rawCiphertext, err := sm2.Encrypt(rand.Reader, sm2Cry.pub, plaintext, nil)
	if err != nil {
		t.Fatal(err)
	}

	enveloped := append([]byte{SM2Version}, rawCiphertext...)
	decoded, err := cry.Decrypt(enveloped)
	if err != nil {
		t.Fatalf("Decrypt(raw C1C3C2) failed: %v", err)
	}

	if !bytes.Equal(decoded, plaintext) {
		t.Errorf("Decrypt(raw) got = %q, want %q", decoded, plaintext)
	}
}

// TestSM2_WrongKey verifies authentication: a wrong key fails with an error.
func TestSM2_WrongKey(t *testing.T) {
	encrypter := newTestSM2CrypterFromGeneratedKey(t)

	decrypter := newTestSM2CrypterFromGeneratedKey(t)

	ciphertext, err := encrypter.Encrypt([]byte("wrong-key-check"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := decrypter.Decrypt(ciphertext); err == nil {
		t.Error("Decrypt with wrong key expected error, got nil")
	}
}

// TestSM2_TamperedCiphertext ...
func TestSM2_TamperedCiphertext(t *testing.T) {
	cry := newTestSM2CrypterFromGeneratedKey(t)

	ciphertext, err := cry.Encrypt([]byte("tamper-check"))
	if err != nil {
		t.Fatal(err)
	}

	ciphertext[len(ciphertext)-1] ^= 0xFF
	if _, err := cry.Decrypt(ciphertext); err == nil {
		t.Error("Decrypt(tampered) expected error, got nil")
	}
}

// TestSM2_InvalidCiphertext ...
func TestSM2_InvalidCiphertext(t *testing.T) {
	cry := newTestSM2CrypterFromGeneratedKey(t)

	if _, err := cry.Encrypt(nil); err == nil {
		t.Error("Encrypt(nil) expected error, got nil")
	}

	if _, err := cry.Decrypt(nil); err == nil {
		t.Error("Decrypt(nil) expected error, got nil")
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

// TestSM2_KeyPairPEMHelpers checks generation and public key extraction.
func TestSM2_KeyPairPEMHelpers(t *testing.T) {
	privPEM, pubPEM, err := GenerateSM2KeyPairPEM()
	if err != nil {
		t.Fatal(err)
	}

	extractedPubPEM, err := GetSM2PublicKeyPEMByPrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(pubPEM, extractedPubPEM) {
		t.Error("extracted public key PEM differs from the generated one")
	}
}

// TestSM2_InvalidPrivateKeyPEM ...
func TestSM2_InvalidPrivateKeyPEM(t *testing.T) {
	if _, err := NewSM2CrypterFromPrivateKey(nil); err == nil {
		t.Error("NewSM2CrypterFromPrivateKey(nil) expected error, got nil")
	}

	if _, err := NewSM2CrypterFromPrivateKey([]byte("not a pem")); err == nil {
		t.Error("NewSM2CrypterFromPrivateKey(invalid) expected error, got nil")
	}

	// an RSA private key is not an SM2 key
	rsaPrivPEM, _, err := GenerateRSAKeyPairPEM(RSAKeySize2048)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewSM2CrypterFromPrivateKey(rsaPrivPEM); err == nil {
		t.Error("NewSM2CrypterFromPrivateKey(RSA key) expected error, got nil")
	}
}

// TestSM2_RejectOtherSuite verifies the single-suite constraint: an SM2
// crypter rejects RSA ciphertexts and an RSA crypter rejects SM2 ciphertexts
// by their version byte, which is what install.go relies on under the
// globally enabled suite.
func TestSM2_RejectOtherSuite(t *testing.T) {
	rsaPrivPEM, _, err := GenerateRSAKeyPairPEM(RSAKeySize2048)
	if err != nil {
		t.Fatal(err)
	}

	rsaCrypter, err := NewRSACrypterFromPrivateKey(rsaPrivPEM)
	if err != nil {
		t.Fatal(err)
	}

	sm2PrivPEM, _, err := GenerateSM2KeyPairPEM()
	if err != nil {
		t.Fatal(err)
	}

	sm2Crypter, err := NewSM2CrypterFromPrivateKey(sm2PrivPEM)
	if err != nil {
		t.Fatal(err)
	}

	rsaV1, err := rsaCrypter.Encrypt([]byte("rsa-v1-credit"))
	if err != nil {
		t.Fatal(err)
	}

	sm2Ct, err := sm2Crypter.Encrypt([]byte("sm2-v3-credit"))
	if err != nil {
		t.Fatal(err)
	}

	// the SM2 suite rejects RSA ciphertexts.
	if _, err := sm2Crypter.Decrypt(rsaV1); err == nil {
		t.Error("sm2 crypter decrypting RSA ciphertext expected error, got nil")
	}

	// the RSA suite rejects SM2 ciphertexts.
	if _, err := rsaCrypter.Decrypt(sm2Ct); err == nil {
		t.Error("rsa crypter decrypting SM2 ciphertext expected error, got nil")
	}

	// same-suite round trips still work.
	if _, err := sm2Crypter.Decrypt(sm2Ct); err != nil {
		t.Errorf("sm2 round trip failed: %v", err)
	}
}
