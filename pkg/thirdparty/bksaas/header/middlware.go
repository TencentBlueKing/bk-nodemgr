/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package header define the header of bksaas.
package header

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

var _ restserver.IRequestIDSetter = &RequestIDSetter{}

// RequestIDSetter this is a midleware for setting request id.
type RequestIDSetter struct {
}

// SetRequestID ...
func (setter *RequestIDSetter) SetRequestID(r restserver.IRequest) error {
	// note: for thread safety you need to reset it here.
	r.Data().SetRequestID(BKRIDGetter(r.GetRequest()))

	if r.Data().GetRequestID() == "" {
		r.Data().SetRequestID(identifier.GenRequestID())
	}

	return nil
}

// NewBKSaasRequestIDSetter ...
func NewBKSaasRequestIDSetter() *RequestIDSetter {
	return &RequestIDSetter{}
}
