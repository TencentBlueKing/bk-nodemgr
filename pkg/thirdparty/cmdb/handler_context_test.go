/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func TestHandlerContextWithVirtualUser(t *testing.T) {
	const (
		tenantID    = "tenant-1"
		callerUser  = "caller"
		virtualUser = "bk-nodemgr"
		loginName   = "caller-login"
		messageID   = "message-1"
		valueKey    = "request-source"
		value       = "unit-test"
	)

	parentCtx, cancel := context.WithCancel(context.Background())
	nCtx := contextx.New(parentCtx,
		contextx.WithTenantID(tenantID),
		contextx.WithBKUsername(callerUser),
		contextx.WithLoginName(loginName),
		contextx.WithMessageID(messageID),
		contextx.WithValues(map[string]any{valueKey: value}),
	)
	h := &Handler{cli: &cli{config: &Config{VirtualUser: virtualUser}}}

	cmdbCtx := h.contextWithVirtualUser(nCtx)

	if cmdbCtx.BKUsername() != virtualUser {
		t.Fatalf("BKUsername() = %q, want %q", cmdbCtx.BKUsername(), virtualUser)
	}
	if nCtx.BKUsername() != callerUser {
		t.Fatalf("source BKUsername() = %q, want %q", nCtx.BKUsername(), callerUser)
	}
	if cmdbCtx.TenantID() != tenantID {
		t.Fatalf("TenantID() = %q, want %q", cmdbCtx.TenantID(), tenantID)
	}
	if cmdbCtx.LoginName() != loginName {
		t.Fatalf("LoginName() = %q, want %q", cmdbCtx.LoginName(), loginName)
	}
	if cmdbCtx.MessageID() != messageID {
		t.Fatalf("MessageID() = %q, want %q", cmdbCtx.MessageID(), messageID)
	}
	if got, ok := cmdbCtx.Value(valueKey).(string); !ok || got != value {
		t.Fatalf("Value(%q) = %q, %t, want %q, true", valueKey, got, ok, value)
	}

	cancel()
	select {
	case <-cmdbCtx.Done():
	default:
		t.Fatal("derived context does not observe parent cancellation")
	}
}
