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

package cipher

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetCurrentPublicKey get the public key of the globally enabled credential
// encryption suite: CLASSIC returns the RSA4096 key, SHANGMI returns the SM2
// key. The key_type field tells the caller which algorithm to encrypt with.
func (h *handler) GetCurrentPublicKey(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.GetCurrentPublicKeyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get current public key, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	keyType := h.cryptoType.CredentialKeyType()

	if err := h.daoCipher.EnsureDefaultCipher(rCtx, keyType); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get current public key, failed to ensure default cipher")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	cipher, err := h.daoCipher.GetCipher(rCtx, types.DefaultCipherName, keyType)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get current public key, failed to get public key from storage")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &protoBackend.GetCurrentPublicKeyResp_Data{
		KeyType:   string(keyType),
		PublicKey: string(cipher.PublicKey),
	}

	return resp, nil
}
