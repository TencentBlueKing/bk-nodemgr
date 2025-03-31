/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import "github.com/TencentBlueKing/bk-nodemgr/pkg/types"

// Validate check body.
func (x *TopoConstantGetReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoConstantGetReq) AutoConvert() {
}

// ConvertFieldsFromTypes convert fields from types to proto.
func (x *TopoConstantGetReq) ConvertFieldsFromTypes(fields types.TopoConstantFields) error {
	x.BkCloudVendor = fields.CloudVendor
	x.BkOsType = fields.OSType

	return nil
}

// ConvertConstantToTypes convert constant to types.
func (x *TopoConstantGetResp) ConvertConstantToTypes() *types.TopoConstant {
	if x.GetData() == nil {
		return &types.TopoConstant{}
	}

	return &types.TopoConstant{
		CloudVendor: x.GetData().GetBkCloudVendor(),
		OSType:      x.GetData().GetBkOsType(),
	}
}
