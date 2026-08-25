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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

const (
	propertyTypeInt        = "int"
	propertyTypeSingleChar = "singlechar"
)

// GetOpsFieldDefs returns the predefined ops (out-of-band) field definitions to be created in CMDB for a business.
func GetOpsFieldDefs() []ObjectAttributeInfo {
	return []ObjectAttributeInfo{
		{
			BKObjID:        TopoNodeObjIDHost,
			BKPropertyID:   string(ccFieldOpsConsoleHostID),
			BKPropertyName: "管控机HostID",
			BKPropertyType: propertyTypeInt,
			Editable:       true,
			// CMDB API document does not mention that the option must be set for int and float types
			Option: NumberOption{
				Min: "",
				Max: "",
			},
		},
		{
			BKObjID:        TopoNodeObjIDHost,
			BKPropertyID:   string(ccFieldOpsOutBandType),
			BKPropertyName: "带外设备类型",
			BKPropertyType: propertyTypeSingleChar,
			Editable:       true,
		},
		{
			BKObjID:        TopoNodeObjIDHost,
			BKPropertyID:   string(ccFieldOpsOutBandProtocol),
			BKPropertyName: "带外设备协议",
			BKPropertyType: propertyTypeSingleChar,
			Editable:       true,
		},
		{
			BKObjID:        TopoNodeObjIDHost,
			BKPropertyID:   string(ccFieldOpsBMCIP),
			BKPropertyName: "带外管理IP",
			BKPropertyType: propertyTypeSingleChar,
			Editable:       true,
		},
		{
			BKObjID:        TopoNodeObjIDHost,
			BKPropertyID:   string(ccFieldOpsBMCPort),
			BKPropertyName: "带外管理端口",
			BKPropertyType: propertyTypeInt,
			Editable:       true,
			// CMDB API document does not mention that the option must be set for int and float types
			// Port range is 0-65535, so set min to 0 and max to 65535
			Option: NumberOption{
				Min: "0",
				Max: "65535",
			},
		},
	}
}

// IObjectAttribute defines the interface for object attribute operations in CMDB.
type IObjectAttribute interface {
	// EnsureBizOpsCustomFields ensures the ops (out-of-band) custom fields exist for a business in CMDB.
	// It queries existing attributes and only creates the ones that are missing.
	EnsureBizOpsCustomFields(ctx contextx.IContext, bizID int64) error
}

// EnsureBizOpsCustomFields ensures the ops (out-of-band) custom fields exist in CMDB for the given business.
// It queries existing attributes for the business and only creates missing ones.
func (h *Handler) EnsureBizOpsCustomFields(nCtx contextx.IContext, bizID int64) error {
	resp, err := h.cli.searchObjectAttribute(nCtx, &SearchObjectAttributeReq{
		BKBizID: bizID,
		BKObjID: TopoNodeObjIDHost,
	})
	if err != nil {
		return fmt.Errorf("failed to search object attributes: %w", err)
	}

	existingMap := make(map[string]struct{}, len(*resp))
	for _, attr := range *resp {
		existingMap[attr.BKPropertyID] = struct{}{}
	}

	opsFields := GetOpsFieldDefs()
	for _, field := range opsFields {
		if _, ok := existingMap[field.BKPropertyID]; ok {
			continue
		}

		req := &CreateBizCustomFieldReq{
			BKBizID:        bizID,
			BKObjID:        field.BKObjID,
			BKPropertyID:   field.BKPropertyID,
			BKPropertyName: field.BKPropertyName,
			BKPropertyType: field.BKPropertyType,
			Editable:       field.Editable,
			Option:         field.Option,
		}
		if _, err := h.cli.createBizCustomField(nCtx, req); err != nil {
			return fmt.Errorf("failed to create biz custom field %s: %w", field.BKPropertyID, err)
		}
	}

	return nil
}
