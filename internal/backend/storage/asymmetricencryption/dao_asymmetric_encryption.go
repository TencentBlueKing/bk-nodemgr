/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package asymmetricencryption defines the interface for asymmetric encryption storage handlers.
package asymmetricencryption

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) getAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (
	*types.AsymmetricEncryption, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	ae, err := s.daoAsymmetricEncryption.Get(nCtx, keyType, cipherType)
	if err != nil {
		return nil, err
	}

	return ae, nil
}

func (s *Storage) createAsymmetricEncryption(nCtx contextx.IContext, encryption ...*types.AsymmetricEncryption) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	return s.daoAsymmetricEncryption.Create(nCtx, encryption...)
}

func (s *Storage) existAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) (
	bool, error) {

	if nCtx == nil {
		return false, basestorage.ErrNilContent()
	}

	return s.daoAsymmetricEncryption.Exist(nCtx, keyType, cipherType)
}

func (s *Storage) upsertAsymmetricEncryption(nCtx contextx.IContext, encryption *types.AsymmetricEncryption) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	return s.daoAsymmetricEncryption.Upsert(nCtx, encryption)
}

func (s *Storage) deleteAsymmetricEncryption(nCtx contextx.IContext, keyType types.AsymmetricKeyType, cipherType types.AsymmetricCipherType) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	return s.daoAsymmetricEncryption.Delete(nCtx, keyType, cipherType)
}
