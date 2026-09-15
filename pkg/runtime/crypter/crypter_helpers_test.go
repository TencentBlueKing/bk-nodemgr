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

package crypter

import (
	"bytes"
	"testing"
)

// TestMD5Sum checks the checksum is a stable hex digest of the input.
func TestMD5Sum(t *testing.T) {
	if got := MD5Sum("bk-nodemgr"); len(got) != 32 {
		t.Errorf("MD5Sum() = %q, want a 32-char hex digest", got)
	}

	if MD5Sum("") != MD5Sum("") {
		t.Error("MD5Sum() should be deterministic")
	}

	if MD5Sum("a") == MD5Sum("b") {
		t.Error("MD5Sum() should differ for different inputs")
	}
}

// TestGetRSAPublicKeyPEMByPrivateKeyPEM checks the public key is extracted
// from a generated private key, and that a non-RSA key is rejected.
func TestGetRSAPublicKeyPEMByPrivateKeyPEM(t *testing.T) {
	privPEM, _, err := GenerateRSAKeyPairPEM(RSAKeySize2048)
	if err != nil {
		t.Fatal(err)
	}

	pubPEM, err := GetRSAPublicKeyPEMByPrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatalf("GetRSAPublicKeyPEMByPrivateKeyPEM() error = %v", err)
	}

	if !bytes.Contains(pubPEM, []byte("PUBLIC KEY")) {
		t.Errorf("public key PEM = %q, want a PKIX public key block", pubPEM)
	}

	if _, err := GetRSAPublicKeyPEMByPrivateKeyPEM([]byte("not a pem")); err == nil {
		t.Error("GetRSAPublicKeyPEMByPrivateKeyPEM(invalid pem) should fail")
	}
}

// TestParseRSAPrivateKeyFromPEM checks the PKCS#8 fallback: a non-RSA key in
// a PKCS#8 wrapper is reported instead of being accepted.
func TestParseRSAPrivateKeyFromPEM(t *testing.T) {
	if _, err := parseRSAPrivateKeyFromPEM([]byte("not a pem")); err == nil {
		t.Error("parseRSAPrivateKeyFromPEM(non pem) should fail")
	}

	// an SM2 private key is valid PKCS#8 but not RSA.
	sm2PrivPEM, _, err := GenerateSM2KeyPairPEM()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parseRSAPrivateKeyFromPEM(sm2PrivPEM); err == nil {
		t.Error("parseRSAPrivateKeyFromPEM(sm2 key) should fail")
	}
}

// TestCrypterOptions checks the salt/label options are applied: the derived
// key changes with the salt (so the same plaintext yields a different
// ciphertext), while a custom label still round-trips through the OAEP path.
func TestCrypterOptions(t *testing.T) {
	key := []byte("unit-test-key")

	aesDefault, err := NewAESCrypter(key)
	if err != nil {
		t.Fatal(err)
	}

	aesSalted, err := NewAESCrypter(key, WithSalt([]byte("another-salt")))
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("option-check")

	defaultCt, err := aesDefault.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	saltedCt, err := aesSalted.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(defaultCt, saltedCt) {
		t.Error("WithSalt should change the derived key and thus the ciphertext")
	}

	privPEM, _, err := GenerateRSAKeyPairPEM(RSAKeySize2048)
	if err != nil {
		t.Fatal(err)
	}

	rsaCrypter, err := NewRSACrypterFromPrivateKey(privPEM, WithRSALabel([]byte("custom.label")))
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := rsaCrypter.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := rsaCrypter.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() with a custom label error = %v", err)
	}

	if !bytes.Equal(decoded, plaintext) {
		t.Errorf("Decrypt() got = %q, want %q", decoded, plaintext)
	}
}
