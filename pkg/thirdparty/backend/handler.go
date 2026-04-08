/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler is interface for nodeman backend Handler.
// nolint: interfacebloat
type IHandler interface {
	IHandlerAuth
	IHandlerHost
	IHandlerNodeAgent
	IHandlerNodeProxy
	IHandlerNodeWorkflow
	IHandlerNodeConstant
	IHandlerRelease
	IHandlerConfigPolicy
	IHandlerPlugin
	IHandlerPluginWorkflow
	IHandlerProcess
	IHandlerEvent
	IHandlerTopo
	IHandlerCipher
}

var _ IHandler = &Handler{}

// Handler defines the backend handler.
type Handler struct {
	cli *cli
}

// New initialize a new nodeman backend Handler.
func New(c *restclient.Capability, conf Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: cli}, nil
}
