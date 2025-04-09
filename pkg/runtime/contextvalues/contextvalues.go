/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package contextvalues defines the runtime context value manager.
package contextvalues

import (
	"context"
	"errors"
)

// Key defines the context key type.
type Key int

const (
	// KeyTenantID defines the tenant id key type.
	KeyTenantID Key = iota

	// KeyRequestID defines the request id key type.
	KeyRequestID
)

// Get get the value of the key.
func Get(ctx context.Context, key Key) (string, error) {
	if ctx == nil {
		return "", errors.New("ctx is nil")
	}

	value, ok := ctx.Value(key).(string)
	if !ok {
		return "", errors.New("key not found in context")
	}

	return value, nil
}

// Set set the value of the key.
func Set(ctx context.Context, key Key, value string) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	return context.WithValue(ctx, key, value), nil
}
