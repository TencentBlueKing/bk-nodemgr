/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for
 * the specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package cipher

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ensureDefaultCipher idempotently creates the default keypair of the globally
// enabled suite under DefaultCipherName. Only one suite is provisioned: the
// globally enabled crypto suite is a deployment decision, so the keypair of the
// disabled suite is never generated nor stored.
func (s *Storage) ensureDefaultCipher(nCtx contextx.IContext, keyType types.CipherKeyType) error {
	exist, err := s.existCipher(nCtx, types.DefaultCipherName, keyType)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	var priv, pub []byte

	switch keyType {
	case types.CipherKeyTypeRSA4096:
		priv, pub, err = crypter.GenerateRSAKeyPairPEM(crypter.RSAKeySize4096)
		if err != nil {
			return fmt.Errorf("failed to generate rsa cipher: %w", err)
		}
	case types.CipherKeyTypeSM2:
		priv, pub, err = crypter.GenerateSM2KeyPairPEM()
		if err != nil {
			return fmt.Errorf("failed to generate sm2 cipher: %w", err)
		}
	default:
		return fmt.Errorf("unsupported default cipher key type: %s", keyType)
	}

	if err := s.createCipher(nCtx, &types.Cipher{
		Name:        types.DefaultCipherName,
		KeyType:     keyType,
		Description: types.DefaultCipherDescription,
		PrivateKey:  priv,
		PublicKey:   pub,
	}); err != nil {
		exist, existErr := s.existCipher(nCtx, types.DefaultCipherName, keyType)
		if existErr != nil {
			return errors.Join(
				fmt.Errorf("failed to create %s cipher: %w", keyType, err),
				fmt.Errorf("failed to re-read existing %s cipher: %w", keyType, existErr),
			)
		}

		if exist {
			return nil
		}

		return fmt.Errorf("failed to create %s cipher: %w", keyType, err)
	}

	return nil
}
