/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import "fmt"

// AsymmetricCipherType defines the type of asymmetric cipher.
type AsymmetricCipherType string

const (
	// AsymmetricCipherTypePublic defines the public key asymmetric cipher type.
	AsymmetricCipherTypePublic AsymmetricCipherType = "public_key"
	// AsymmetricCipherTypePrivate defines the private key asymmetric cipher type.
	AsymmetricCipherTypePrivate AsymmetricCipherType = "private_key"
)

// Validate validates the AsymmetricCipherType.
func (ct AsymmetricCipherType) Validate() error {
	switch ct {
	case AsymmetricCipherTypePublic, AsymmetricCipherTypePrivate:
		return nil
	default:
		return fmt.Errorf("invalid AsymmetricCipherType: %s", ct)
	}
}

// AsymmetricKeyType defines the type of asymmetric key.
type AsymmetricKeyType string

const (
	// AsymmetricKeyTypeRSA defines the RSA key type.
	AsymmetricKeyTypeRSA AsymmetricKeyType = "rsa"
)

// Validate validates the AsymmetricKeyType.
func (kt AsymmetricKeyType) Validate() error {
	switch kt {
	case AsymmetricKeyTypeRSA:
		return nil
	default:
		return fmt.Errorf("invalid AsymmetricKeyType: %s", kt)
	}
}

// AsymmetricEncryption represents asymmetric encryption information.
type AsymmetricEncryption struct {
	CipherType  AsymmetricCipherType
	KeyType     AsymmetricKeyType
	Description string
	Content     []byte
}

// Validate validates the AsymmetricEncryption.
func (ae *AsymmetricEncryption) Validate() error {
	if err := ae.CipherType.Validate(); err != nil {
		return err
	}

	if err := ae.KeyType.Validate(); err != nil {
		return err
	}

	return nil
}

const (
	// DefaultAsymmetricEncryptionDescription is the default description for asymmetric encryption keys.
	DefaultAsymmetricEncryptionDescription = "Default asymmetric encryption key"
)

// NewDefaultRSAPublicEncryption creates a new default RSA public key asymmetric encryption.
func NewDefaultRSAPublicEncryption(content []byte) *AsymmetricEncryption {
	return &AsymmetricEncryption{
		CipherType:  AsymmetricCipherTypePublic,
		KeyType:     AsymmetricKeyTypeRSA,
		Description: DefaultAsymmetricEncryptionDescription,
		Content:     content,
	}
}

// NewDefaultRSAPrivateEncryption creates a new default RSA private key asymmetric encryption.
func NewDefaultRSAPrivateEncryption(content []byte) *AsymmetricEncryption {
	return &AsymmetricEncryption{
		CipherType:  AsymmetricCipherTypePrivate,
		KeyType:     AsymmetricKeyTypeRSA,
		Description: DefaultAsymmetricEncryptionDescription,
		Content:     content,
	}
}
