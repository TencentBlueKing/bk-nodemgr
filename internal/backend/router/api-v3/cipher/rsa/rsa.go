/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rsa provides the RSA encryption related API handlers.
package rsa

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg        *gin.RouterGroup
	daoCipher cipher.IStorage
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:        rg.Group("/rsa"),
		daoCipher: capability.StorageCipher,
	}
}

// Load loads deploy policy handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/get_public_key", restserver.Handler(h.GetRSAPublicKey))
}

// GetRSAPublicKey get rsa public key.
func (h *handler) GetRSAPublicKey(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.GetRSAPublicKeyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get rsa public key, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cipher, err := h.daoCipher.GetCipher(rCtx, types.DefaultCipherName, types.CipherKeyTypeRSA4096)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get rsa public key, failed to get public key from storage")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := &protoBackend.GetRSAPublicKeyResp_Data{
		PublicKey: string(cipher.PublicKey),
	}

	return resp, nil
}
