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
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestConvertHostFieldSelectionToFields_DynamicAgentState(t *testing.T) {
	fields := convertHostFieldSelectionToFields(&types.HostFieldSelection{
		HostID:         true,
		AgentID:        true,
		NodeRole:       true,
		NodeStatus:     true,
		NodeVersion:    true,
		NodeGeneration: true,
	})
	expected := []string{
		FieldKeyHostID,
		FieldKeyDynamicNodeRole,
		FieldKeyDynamicNodeStatus,
		FieldKeyDynamicNodeVersion,
		FieldKeyDynamicNodeGeneration,
		FieldKeyDynamicAgentID,
	}

	if !reflect.DeepEqual(fields, expected) {
		t.Fatalf("fields = %#v, want %#v", fields, expected)
	}
}
