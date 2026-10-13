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

package host

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestConvertHostFromTypesWritesNumericZoneID(t *testing.T) {
	host := convertHostFromTypes(&types.Host{
		HostID:   10001,
		TenantID: "tenant-a",
		Static: &types.HostStatic{
			ZoneID: 42,
		},
	})

	if host.Static.ZoneID != 42 {
		t.Fatalf("ZoneID = %d, want 42", host.Static.ZoneID)
	}
}

func TestConvertHostToTypesReadsNumericZoneID(t *testing.T) {
	host := convertHostToTypes(&Host{
		Static: &HostStatic{
			ZoneID: 42,
		},
	})

	if host.Static.ZoneID != 42 {
		t.Fatalf("ZoneID = %d, want 42", host.Static.ZoneID)
	}
}

func TestGenerateHostStaticUpdatesWritesNumericZoneID(t *testing.T) {
	updates := generateHostStaticUpdates(types.HostStaticFields{ZoneID: true}, &Host{
		Static: &HostStatic{ZoneID: 42},
	})

	if updates[FieldKeyStaticZoneID] != int64(42) {
		t.Fatalf("ZoneID update = %v, want 42", updates[FieldKeyStaticZoneID])
	}
}

func TestConvertHostToTypesReadsLegacyOpsFields(t *testing.T) {
	host := convertHostToTypes(&Host{
		Static: &HostStatic{
			LegacySyncedOpsOutBandIP:   "127.0.0.1",
			LegacySyncedOpsOutBandPort: 623,
		},
		Dynamic: &HostDynamic{
			LegacyOpsOutBandIP:   "127.0.0.2",
			LegacyOpsOutBandPort: 624,
		},
	})

	if host.Static.SyncedOpsOutBandIP != "127.0.0.1" {
		t.Fatalf("SyncedOpsOutBandIP = %q, want %q", host.Static.SyncedOpsOutBandIP, "127.0.0.1")
	}
	if host.Static.SyncedOpsOutBandPort != 623 {
		t.Fatalf("SyncedOpsOutBandPort = %d, want 623", host.Static.SyncedOpsOutBandPort)
	}
	if host.Dynamic.OpsOutBandIP != "127.0.0.2" {
		t.Fatalf("OpsOutBandIP = %q, want %q", host.Dynamic.OpsOutBandIP, "127.0.0.2")
	}
	if host.Dynamic.OpsOutBandPort != 624 {
		t.Fatalf("OpsOutBandPort = %d, want 624", host.Dynamic.OpsOutBandPort)
	}
}
