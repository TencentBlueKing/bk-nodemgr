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

package cipher

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) ensureDefaultCipher(nCtx contextx.IContext) error {
	exist, err := s.existCipher(nCtx, types.DefaultCipherName, types.CipherKeyTypeRSA4096)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	priv, pub, err := crypter.GenerateRSAKeyPairPEM(crypter.RSAKeySize4096)
	if err != nil {
		return fmt.Errorf("failed to generate rsa cipher: %w", err)
	}

	if err := s.createCipher(nCtx, &types.Cipher{
		Name:        types.DefaultCipherName,
		KeyType:     types.CipherKeyTypeRSA4096,
		Description: types.DefaultCipherDescription,
		PrivateKey:  priv,
		PublicKey:   pub,
	}); err != nil {
		exist, existErr := s.existCipher(nCtx, types.DefaultCipherName, types.CipherKeyTypeRSA4096)
		if existErr != nil {
			return errors.Join(
				fmt.Errorf("failed to create rsa cipher: %w", err),
				fmt.Errorf("failed to re-read existing rsa cipher: %w", existErr),
			)
		}

		if exist {
			return nil
		}

		return fmt.Errorf("failed to create rsa cipher: %w", err)
	}

	return nil
}
