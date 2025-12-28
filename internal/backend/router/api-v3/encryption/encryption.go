/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package encryption provides the encryption related API handlers.
package encryption

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/encryption/rsa"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/asymmetricencryption"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg                      *gin.RouterGroup
	daoAsymmetricEncryption asymmetricencryption.IStorage
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		rg:                      rg.Group("/encryption"),
		daoAsymmetricEncryption: capability.StorageAsymmetricEncryption,
	}
}

// Load loads deploy policy handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	rsa.Load(h.rg, capability)
}
