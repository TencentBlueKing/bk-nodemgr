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

// Package crypter provides encryption utilities based on SM2 public-key encryption.
package crypter

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

const (
	// SM2Version is the version prefix for SM2 ciphertext, sharing the
	// frontend credential ciphertext space with RSA v1 (OAEP) / v2 (hybrid)
	// so decryption can dispatch by the leading version byte.
	SM2Version = 3
)

// SM2 defines the crypter instance based on SM2 public-key encryption,
// backed by emmansun/gmsm (pure Go).
type SM2 struct {
	priv *sm2.PrivateKey
	pub  *ecdsa.PublicKey
}

// NewSM2CrypterFromPrivateKey creates a new SM2 crypter from a PKCS#8 PEM
// private key. It enables both encryption and decryption.
func NewSM2CrypterFromPrivateKey(pemBytes []byte) (Crypter, error) {
	if pemBytes == nil {
		return nil, errors.New("sm2 pem private key cannot be nil")
	}

	priv, err := parseSM2PrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sm2 private key: %w", err)
	}

	return &SM2{
		priv: priv,
		pub:  &priv.PublicKey,
	}, nil
}

// Encrypt encrypts the plaintext using SM2 with ASN.1 (C1C3C2) DER encoding,
// compatible with BouncyCastle and jsrsasign style implementations.
// The ciphertext layout: SM2Version(1) + SM2 ciphertext.
func (crypt *SM2) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	if crypt.pub == nil {
		return nil, errors.New("sm2 public key is not set")
	}

	ciphertext, err := sm2.EncryptASN1(rand.Reader, crypt.pub, plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with sm2: %w", err)
	}

	result := make([]byte, 1+len(ciphertext))
	result[0] = SM2Version
	copy(result[1:], ciphertext)

	return result, nil
}

// Decrypt decrypts the SM2 ciphertext. Both ASN.1 DER and raw C1C3C2
// (uncompressed 04-prefixed) SM2 ciphertexts are accepted, so frontend
// implementations may use either encoding. A wrong key or tampered
// ciphertext fails with an error (SM3 digest check).
func (crypt *SM2) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) <= 1 {
		return nil, errors.New("invalid ciphertext length")
	}

	if crypt.priv == nil {
		return nil, errors.New("sm2 private key is not set")
	}

	// verification version.
	if ciphertext[0] != SM2Version {
		return nil, fmt.Errorf("unsupported version: %d", ciphertext[0])
	}

	plaintext, err := sm2.Decrypt(crypt.priv, ciphertext[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with sm2: %w", err)
	}

	return plaintext, nil
}

// GenerateSM2KeyPairPEM generates an SM2 key pair and returns the
// PKCS#8 PEM-encoded private key and PKIX PEM-encoded public key.
func GenerateSM2KeyPairPEM() ([]byte, []byte, error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate sm2 key: %w", err)
	}

	privBytes, err := smx509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal sm2 private key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypePrivateKeyType, Bytes: privBytes})

	pubBytes, err := smx509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal sm2 public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypePublicKeyType, Bytes: pubBytes})

	return privPEM, pubPEM, nil
}

// GetSM2PublicKeyPEMByPrivateKeyPEM extracts the public key PEM from the
// given SM2 private key PEM.
func GetSM2PublicKeyPEMByPrivateKeyPEM(privPEM []byte) ([]byte, error) {
	priv, err := parseSM2PrivateKeyFromPEM(privPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sm2 private key: %w", err)
	}

	pubBytes, err := smx509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sm2 public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypePublicKeyType, Bytes: pubBytes})

	return pubPEM, nil
}

// parseSM2PrivateKeyFromPEM parses an SM2 private key from a PKCS#8
// PEM-encoded block.
func parseSM2PrivateKeyFromPEM(pemBytes []byte) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid pem data for sm2 private key")
	}

	key, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pkcs8 private key: %w", err)
	}

	priv, ok := key.(*sm2.PrivateKey)
	if !ok {
		return nil, errors.New("parsed private key is not SM2")
	}

	return priv, nil
}
