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

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerCipher cipher handler interface.
type IHandlerCipher interface {
	// GetRSAPublicKey get rsa public key.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @return the rsa public key string and error.
	GetRSAPublicKey(nCtx contextx.IContext) (string, error)

	// GetCurrentPublicKey get the public key of the globally enabled
	// credential encryption suite.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @return the key type (RSA4096 or SM2), the public key string and error.
	GetCurrentPublicKey(nCtx contextx.IContext) (types.CipherKeyType, string, error)
}

// GetRSAPublicKey get rsa public key.
func (h *Handler) GetRSAPublicKey(nCtx contextx.IContext) (string, error) {
	req := &protoBackend.GetRSAPublicKeyReq{}

	resp, err := h.cli.getRSAPublicKey(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetData().GetPublicKey(), nil
}

// GetCurrentPublicKey get the public key of the globally enabled credential
// encryption suite.
func (h *Handler) GetCurrentPublicKey(nCtx contextx.IContext) (types.CipherKeyType, string, error) {
	req := &protoBackend.GetCurrentPublicKeyReq{}

	resp, err := h.cli.getCurrentPublicKey(nCtx, req)
	if err != nil {
		return "", "", err
	}

	return types.CipherKeyType(resp.GetData().GetKeyType()), resp.GetData().GetPublicKey(), nil
}
