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

const (
	// DefaultCipherName is the default name for cipher keys.
	DefaultCipherName = "DEFAULT"

	// DefaultCipherDescription is the default description for cipher keys.
	DefaultCipherDescription = "Default cipher key"
)

// CipherKeyType defines the type of asymmetric key.
type CipherKeyType string

const (
	// CipherKeyTypeAES256 defines the AES 256 key type.
	CipherKeyTypeAES256 CipherKeyType = "AES256"

	// CipherKeyTypeRSA2048 defines the RSA 2048 key type.
	CipherKeyTypeRSA2048 CipherKeyType = "RSA2048"

	// CipherKeyTypeRSA3072 defines the RSA 3072 key type.
	CipherKeyTypeRSA3072 CipherKeyType = "RSA3072"

	// CipherKeyTypeRSA4096 defines the RSA 4096 key type.
	CipherKeyTypeRSA4096 CipherKeyType = "RSA4096"

	// CipherKeyTypeSM2 defines the SM2 key type.
	CipherKeyTypeSM2 CipherKeyType = "SM2"

	// CipherKeyTypeSM4 defines the SM4 key type.
	CipherKeyTypeSM4 CipherKeyType = "SM4"
)

// Validate validates the CipherKeyType.
func (kt CipherKeyType) Validate() error {
	switch kt {
	case CipherKeyTypeAES256,
		CipherKeyTypeRSA2048,
		CipherKeyTypeRSA3072,
		CipherKeyTypeRSA4096,
		CipherKeyTypeSM2,
		CipherKeyTypeSM4:
		return nil
	default:
		return fmt.Errorf("invalid CipherKeyType: %s", kt)
	}
}

// Cipher represents cipher information.
type Cipher struct {
	Name        string
	KeyType     CipherKeyType
	Description string
	PrivateKey  []byte
	PublicKey   []byte
}

// Validate validates the Cipher.
func (ae *Cipher) Validate() error {
	if err := ae.KeyType.Validate(); err != nil {
		return err
	}

	if len(ae.PrivateKey) == 0 {
		return fmt.Errorf("private key cannot be empty")
	}

	return nil
}
