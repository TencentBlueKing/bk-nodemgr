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
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildAggregateDistinctFields(t *testing.T) {
	got := buildAggregateDistinctFields(types.NewHostDistinctRequestAllSet())
	want := []base.AggregateDistinctField{
		{ResultKey: distinctResultKeyBizID, FieldPath: FieldKeyStaticBizID, BSONType: bson.TypeInt64},
		{ResultKey: distinctResultKeyNodeRole, FieldPath: FieldKeyDynamicNodeRole, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyNodeStatus, FieldPath: FieldKeyDynamicNodeStatus, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyNodeVersion, FieldPath: FieldKeyDynamicNodeVersion, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyDeptName, FieldPath: FieldKeyStaticDeptName, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyOSType, FieldPath: FieldKeyStaticOSType, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyArch, FieldPath: FieldKeyStaticArch, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyAddressing, FieldPath: FieldKeyStaticAddressing, BSONType: bson.TypeString},
		{ResultKey: distinctResultKeyNetworkAreaID, FieldPath: FieldKeyStaticNetworkAreaID, BSONType: bson.TypeInt64},
		{ResultKey: distinctResultKeyNetworkUnitID, FieldPath: FieldKeyDynamicNetworkUnitID, BSONType: bson.TypeInt64},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildAggregateDistinctFields() = %#v, want %#v", got, want)
	}
}

func TestConvertAggregateDistinctResult(t *testing.T) {
	request := types.HostDistinctRequest{
		BizID:       true,
		NodeRole:    true,
		NodeStatus:  true,
		NodeVersion: true,
		OSType:      true,
	}
	result := base.AggregateDistinctResult{
		distinctResultKeyBizID:       {rawValue(t, int64(0)), rawValue(t, int64(42))},
		distinctResultKeyNodeRole:    {rawValue(t, ""), rawValue(t, string(types.NodeRoleAgent))},
		distinctResultKeyNodeStatus:  {rawValue(t, ""), rawValue(t, string(types.NodeStatusRunning))},
		distinctResultKeyNodeVersion: {rawValue(t, "")},
		distinctResultKeyOSType:      make([]bson.RawValue, 0),
	}

	got := convertAggregateDistinctResult(request, result)
	if !reflect.DeepEqual(got.BizID, []int64{0, 42}) {
		t.Fatalf("BizID = %#v, want [0 42]", got.BizID)
	}
	if !reflect.DeepEqual(got.NodeRole, []types.NodeRole{"", types.NodeRoleAgent}) {
		t.Fatalf("NodeRole = %#v", got.NodeRole)
	}
	if !reflect.DeepEqual(got.NodeStatus, []types.NodeStatus{"", types.NodeStatusRunning}) {
		t.Fatalf("NodeStatus = %#v", got.NodeStatus)
	}
	if !reflect.DeepEqual(got.NodeVersion, []string{""}) {
		t.Fatalf("NodeVersion = %#v, want empty string", got.NodeVersion)
	}
	if got.OSType == nil || len(got.OSType) != 0 {
		t.Fatalf("OSType = %#v, want non-nil empty slice", got.OSType)
	}
	if got.DeptName != nil || got.Arch != nil || got.Addressing != nil ||
		got.NetworkAreaID != nil || got.NetworkUnitID != nil {
		t.Fatalf("unselected fields must remain nil: %#v", got)
	}
}

func TestConvertAggregateDistinctResultSelectedEmpty(t *testing.T) {
	request := types.HostDistinctRequest{NetworkAreaID: true}
	result := base.AggregateDistinctResult{
		distinctResultKeyNetworkAreaID: make([]bson.RawValue, 0),
	}

	got := convertAggregateDistinctResult(request, result)
	if got.NetworkAreaID == nil || len(got.NetworkAreaID) != 0 {
		t.Fatalf("NetworkAreaID = %#v, want non-nil empty slice", got.NetworkAreaID)
	}
	if got.BizID != nil || got.NodeRole != nil || got.NodeStatus != nil || got.NodeVersion != nil ||
		got.DeptName != nil || got.OSType != nil || got.Arch != nil || got.Addressing != nil || got.NetworkUnitID != nil {
		t.Fatalf("unselected fields must remain nil: %#v", got)
	}
}

func rawValue(t *testing.T, value any) bson.RawValue {
	t.Helper()

	valueType, data, err := bson.MarshalValue(value)
	if err != nil {
		t.Fatalf("bson.MarshalValue() error = %v", err)
	}

	return bson.RawValue{Type: valueType, Value: data}
}
