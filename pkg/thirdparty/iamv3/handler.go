/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iamv3

import (
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler the Handler of IAM v3.
// This is a skeleton MVP - methods will be added as needed.
type IHandler interface {
	// Future methods will be added here (e.g., IsAllowed, RegisterResources, etc.)
}

// Handler the Handler of IAM v3 (skeleton implementation).
type Handler struct {
	cli *cli
	// Future fields: scheduler, cache, logger, etc.
}

// Verify that Handler implements IHandler interface.
var _ IHandler = (*Handler)(nil)

// New initialize a new IAM v3 Handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{cli: cli}
	// Future: h.initCache(), h.initScheduler(), etc.

	return h, nil
}
