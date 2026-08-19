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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	distinctResultKeyBizID         = "biz_id"
	distinctResultKeyNodeRole      = "node_role"
	distinctResultKeyNodeStatus    = "node_status"
	distinctResultKeyNodeVersion   = "node_version"
	distinctResultKeyDeptName      = "dept_name"
	distinctResultKeyOSType        = "os_type"
	distinctResultKeyArch          = "arch"
	distinctResultKeyAddressing    = "addressing"
	distinctResultKeyNetworkAreaID = "networkarea_id"
	distinctResultKeyNetworkUnitID = "networkunit_id"
)

// DistinctFields returns distinct values for the requested host fields in one aggregation.
func (h *handler) DistinctFields(
	nCtx contextx.IContext, request types.HostDistinctRequest, opts ...OptFn,
) (*types.HostDistinctResult, error) {

	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	fields := buildAggregateDistinctFields(request)
	if len(fields) == 0 {
		return new(types.HostDistinctResult), nil
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	result, err := base.AggregateDistinct(nCtx, h.tenantDao(nCtx.TenantID()), filter, fields)
	if err != nil {
		return nil, err
	}

	return convertAggregateDistinctResult(request, result), nil
}

func buildAggregateDistinctFields(request types.HostDistinctRequest) []base.AggregateDistinctField {
	fields := make([]base.AggregateDistinctField, 0)
	if request.BizID {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyBizID,
			FieldPath: FieldKeyStaticBizID,
			BSONType:  bson.TypeInt64,
		})
	}
	if request.NodeRole {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyNodeRole,
			FieldPath: FieldKeyDynamicNodeRole,
			BSONType:  bson.TypeString,
		})
	}
	if request.NodeStatus {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyNodeStatus,
			FieldPath: FieldKeyDynamicNodeStatus,
			BSONType:  bson.TypeString,
		})
	}
	if request.NodeVersion {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyNodeVersion,
			FieldPath: FieldKeyDynamicNodeVersion,
			BSONType:  bson.TypeString,
		})
	}
	if request.DeptName {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyDeptName,
			FieldPath: FieldKeyStaticDeptName,
			BSONType:  bson.TypeString,
		})
	}
	if request.OSType {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyOSType,
			FieldPath: FieldKeyStaticOSType,
			BSONType:  bson.TypeString,
		})
	}
	if request.Arch {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyArch,
			FieldPath: FieldKeyStaticArch,
			BSONType:  bson.TypeString,
		})
	}
	if request.Addressing {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyAddressing,
			FieldPath: FieldKeyStaticAddressing,
			BSONType:  bson.TypeString,
		})
	}
	if request.NetworkAreaID {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyNetworkAreaID,
			FieldPath: FieldKeyStaticNetworkAreaID,
			BSONType:  bson.TypeInt64,
		})
	}
	if request.NetworkUnitID {
		fields = append(fields, base.AggregateDistinctField{
			ResultKey: distinctResultKeyNetworkUnitID,
			FieldPath: FieldKeyDynamicNetworkUnitID,
			BSONType:  bson.TypeInt64,
		})
	}

	return fields
}

func convertAggregateDistinctResult(
	request types.HostDistinctRequest, result base.AggregateDistinctResult,
) *types.HostDistinctResult {

	data := new(types.HostDistinctResult)
	if request.BizID {
		data.BizID = rawInt64Values(result[distinctResultKeyBizID])
	}
	if request.NodeRole {
		data.NodeRole = types.StringListToNodeRoleList(rawStringValues(result[distinctResultKeyNodeRole]))
	}
	if request.NodeStatus {
		data.NodeStatus = types.StringListToNodeStatusList(rawStringValues(result[distinctResultKeyNodeStatus]))
	}
	if request.NodeVersion {
		data.NodeVersion = rawStringValues(result[distinctResultKeyNodeVersion])
	}
	if request.DeptName {
		data.DeptName = rawStringValues(result[distinctResultKeyDeptName])
	}
	if request.OSType {
		data.OSType = rawStringValues(result[distinctResultKeyOSType])
	}
	if request.Arch {
		data.Arch = rawStringValues(result[distinctResultKeyArch])
	}
	if request.Addressing {
		data.Addressing = rawStringValues(result[distinctResultKeyAddressing])
	}
	if request.NetworkAreaID {
		data.NetworkAreaID = rawInt64Values(result[distinctResultKeyNetworkAreaID])
	}
	if request.NetworkUnitID {
		data.NetworkUnitID = rawInt64Values(result[distinctResultKeyNetworkUnitID])
	}

	return data
}

func rawStringValues(values []bson.RawValue) []string {
	result := make([]string, len(values))
	for idx, value := range values {
		result[idx] = value.StringValue()
	}

	return result
}

func rawInt64Values(values []bson.RawValue) []int64 {
	result := make([]int64, len(values))
	for idx, value := range values {
		result[idx] = value.Int64()
	}

	return result
}
