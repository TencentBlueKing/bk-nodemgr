/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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
	updates := generateHostStaticUpdates(types.HostStaticFields{ZoneID: true}, &types.Host{
		Static: &types.HostStatic{ZoneID: 42},
	})

	if updates[FieldKeyStaticZoneID] != int64(42) {
		t.Fatalf("ZoneID update = %v, want 42", updates[FieldKeyStaticZoneID])
	}
}
