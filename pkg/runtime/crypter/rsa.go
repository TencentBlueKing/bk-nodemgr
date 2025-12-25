/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package crypter provides encryption utilities based on RSA-OAEP
package crypter

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

const (
	// RSADefaultLabel is the default label used for RSA-OAEP.
	RSADefaultLabel = "com.example.crypto.rsa.v1"
	// RSAVersion is the version prefix for RSA ciphertext.
	RSAVersion = 1

	// PEMBlockTypeRSAPrivateKeyType is the PEM block type for RSA private keys.
	PEMBlockTypeRSAPrivateKeyType string = "RSA PRIVATE KEY"

	// PEMBlockTypeRSAPublicKeyType is the PEM block type for RSA public keys.
	PEMBlockTypeRSAPublicKeyType string = "RSA PUBLIC KEY"

	// PEMBlockTypePrivateKeyType is the PEM block type for generic private keys.
	PEMBlockTypePrivateKeyType string = "PRIVATE KEY"

	// PEMBlockTypePublicKeyType is the PEM block type for generic public keys.
	PEMBlockTypePublicKeyType string = "PUBLIC KEY"
)

// RSAKeySize defines the RSA key size.
type RSAKeySize int

const (
	// RSAKeySize2048 represents a 2048-bit RSA key size.
	RSAKeySize2048 RSAKeySize = 2048
	// RSAKeySize3072 represents a 3072-bit RSA key size.
	RSAKeySize3072 RSAKeySize = 3072
	// RSAKeySize4096 represents a 4096-bit RSA key size.
	RSAKeySize4096 RSAKeySize = 4096
)

// Validate checks if the RSAKeySize is valid.
func (r RSAKeySize) Validate() error {
	switch r {
	case RSAKeySize2048, RSAKeySize3072, RSAKeySize4096:
		return nil
	default:
		return fmt.Errorf("invalid RSA key size: %d", r)
	}
}

// ToInt converts RSAKeySize to int.
func (r RSAKeySize) ToInt() int {
	return int(r)
}

// RSA defines the crypter instance based on RSA-OAEP.
type RSA struct {
	pub   *rsa.PublicKey
	priv  *rsa.PrivateKey
	label []byte
}

// RSAOption is an option for the RSA crypter.
type RSAOption func(*RSA)

// WithRSALabel sets the label for RSA-OAEP.
func WithRSALabel(label []byte) RSAOption {
	return func(r *RSA) {
		r.label = label
	}
}

// NewRSACrypterFromPrivateKey creates a new RSA crypter from a private key.
// It enables both encryption and decryption.
func NewRSACrypterFromPrivateKey(pemBytes []byte, opts ...RSAOption) (Crypter, error) {
	if pemBytes == nil {
		return nil, errors.New("rsa pem private key cannot be nil")
	}

	priv, err := parseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa private key: %w", err)
	}

	r := &RSA{
		priv:  priv,
		pub:   &priv.PublicKey,
		label: []byte(RSADefaultLabel),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// NewRSACrypterFromPublicKey creates a new RSA crypter from a public key.
// It only supports encryption.
func NewRSACrypterFromPublicKey(pemBytes []byte, opts ...RSAOption) (Crypter, error) {
	if pemBytes == nil {
		return nil, errors.New("rsa pem public key cannot be nil")
	}

	pub, err := parseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa public key: %w", err)
	}

	r := &RSA{
		pub:   pub,
		label: []byte(RSADefaultLabel),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// Encrypt encrypts the plaintext using RSA-OAEP.
// The output layout is: RSAVersion(1) + raw RSA ciphertext.
func (r *RSA) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	if r.pub == nil {
		return nil, errors.New("rsa public key is not set")
	}

	h := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(h, rand.Reader, r.pub, plaintext, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with rsa-oaep: %w", err)
	}

	// prepend version byte
	result := make([]byte, 1+len(ciphertext))
	result[0] = RSAVersion
	copy(result[1:], ciphertext)

	return result, nil
}

// Decrypt decrypts the ciphertext using RSA-OAEP.
// It expects the input layout: RSAVersion(1) + raw RSA ciphertext.
func (r *RSA) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) <= 1 {
		return nil, errors.New("invalid ciphertext length")
	}

	if r.priv == nil {
		return nil, errors.New("rsa private key is not set")
	}

	if ciphertext[0] != RSAVersion {
		return nil, fmt.Errorf("unsupported rsa version: %d", ciphertext[0])
	}

	h := sha256.New()
	actualCiphertext := ciphertext[1:]

	plaintext, err := rsa.DecryptOAEP(h, rand.Reader, r.priv, actualCiphertext, r.label)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with rsa-oaep: %w", err)
	}

	return plaintext, nil
}

// parseRSAPrivateKeyFromPEM parses an RSA private key from a PEM-encoded block.
// It supports PKCS#1 and PKCS#8 formats.
func parseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid pem data for rsa private key")
	}

	// PKCS1 first if possible
	if priv, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return priv, nil
	}

	// then try PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa private key: %w", err)
	}

	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("parsed private key is not RSA")
	}

	return priv, nil
}

// parseRSAPublicKeyFromPEM parses an RSA public key from a PEM-encoded block.
// It supports PKCS#1 and PKIX public key formats.
func parseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid pem data for rsa public key")
	}

	// PKCS1 first if possible
	if pub, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return pub, nil
	}

	// then try PKIX
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa public key: %w", err)
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("parsed public key is not RSA")
	}

	return pub, nil
}

// GenerateRSAKeyPairPEM generates an RSA key pair and returns the PEM-encoded private key and public key.
func GenerateRSAKeyPairPEM(bits RSAKeySize) ([]byte, []byte, error) {
	priv, err := generateRSAKeyPair(bits)
	if err != nil {
		return nil, nil, err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypeRSAPrivateKeyType, Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal rsa public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: PEMBlockTypePrivateKeyType, Bytes: pubBytes})

	return privPEM, pubPEM, nil
}

// generateRSAKeyPair generates an RSA private/public key pair with the given bit size.
func generateRSAKeyPair(bits RSAKeySize) (*rsa.PrivateKey, error) {
	if err := bits.Validate(); err != nil {
		return nil, fmt.Errorf("validate rsa key size failed, bits(%d): %w", bits, err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, bits.ToInt())
	if err != nil {
		return nil, fmt.Errorf("failed to generate rsa key: %w", err)
	}

	return priv, nil
}
