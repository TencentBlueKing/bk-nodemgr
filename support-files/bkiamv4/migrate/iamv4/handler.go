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

// Package iamv4 encapsulates IAM V4 model APIs for the migration tool.
package iamv4

import (
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
)

// IHandler combines all model registration operations without HTTP details.
type IHandler interface {
	IHandlerSystem
	IHandlerResourceType
	IHandlerAction
	IHandlerRole
}

// Handler adapts model registration operations to the IAM V4 client.
type Handler struct {
	cli *cli
}

var _ IHandler = (*Handler)(nil)

// New initializes the migration tool's IAM V4 handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	client, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: client}, nil
}
