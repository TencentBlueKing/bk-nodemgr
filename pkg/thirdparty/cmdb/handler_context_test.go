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

package cmdb

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

func TestContextWithTenantIDPreservesCallerIdentity(t *testing.T) {
	const (
		sourceTenantID = "tenant-1"
		targetTenantID = "tenant-2"
		bkUsername     = "caller-tenant-user"
		loginName      = "caller-login"
		messageID      = "message-1"
		valueKey       = "request-source"
		value          = "unit-test"
	)

	parentCtx, cancel := context.WithCancel(context.Background())
	nCtx := contextx.New(parentCtx,
		contextx.WithTenantID(sourceTenantID),
		contextx.WithBKUsername(bkUsername),
		contextx.WithLoginName(loginName),
		contextx.WithMessageID(messageID),
		contextx.WithValues(map[string]any{valueKey: value}),
	)

	cmdbCtx := contextx.From(nCtx, contextx.WithTenantID(targetTenantID))

	if cmdbCtx.TenantID() != targetTenantID {
		t.Fatalf("TenantID() = %q, want %q", cmdbCtx.TenantID(), targetTenantID)
	}
	if nCtx.TenantID() != sourceTenantID {
		t.Fatalf("source TenantID() = %q, want %q", nCtx.TenantID(), sourceTenantID)
	}
	if cmdbCtx.BKUsername() != bkUsername {
		t.Fatalf("BKUsername() = %q, want %q", cmdbCtx.BKUsername(), bkUsername)
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
