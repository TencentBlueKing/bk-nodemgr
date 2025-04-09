/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package identifier provides identifier definition for runtime.
package identifier

import (
	"context"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/contextvalues"
	"github.com/google/uuid"
)

const (
	tagRequestID           = "rid"
	tagTriggerID           = "trig"
	tagOperationID         = "oper"
	tagOperationInstanceID = "oper-inst"
	tagActionInstanceID    = "act-inst"
	tagServiceID           = "svc"
)

func generateID(tag string) string {
	return tag + ":" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenRequestID generates a request id.
func GenRequestID() string {
	return generateID(tagRequestID)
}

// GenTriggerID generates a trigger id.
func GenTriggerID() string {
	return generateID(tagTriggerID)
}

// GenOperationID generates a operation id.
func GenOperationID() string {
	return generateID(tagOperationID)
}

// GenOperationInstanceID generates a operation instance id.
func GenOperationInstanceID() string {
	return generateID(tagOperationInstanceID)
}

// GenActionInstanceID generates a action instance id.
func GenActionInstanceID() string {
	return generateID(tagActionInstanceID)
}

// GenServiceID generates a service id.
func GenServiceID() string {
	return generateID(tagServiceID)
}

// GetRequestID gets the request id from context.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	rid, err := contextvalues.Get(ctx, contextvalues.KeyRequestID)
	if err != nil {
		return ""
	}

	return rid
}

// SetRequestID sets the request id to context.
func SetRequestID(ctx context.Context, rid string) (context.Context, error) {
	return contextvalues.Set(ctx, contextvalues.KeyRequestID, rid)
}
