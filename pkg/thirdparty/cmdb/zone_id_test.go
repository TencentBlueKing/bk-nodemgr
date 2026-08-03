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
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

func TestHostEventInfoUnmarshalNumericBKIDCAreaID(t *testing.T) {
	event := map[string]any{
		"bk_cursor":     "cursor-1",
		"bk_resource":   "host",
		"bk_event_type": "update",
		"bk_detail": map[string]any{
			"bk_host_id":      float64(10001),
			"bk_cloud_id":     float64(0),
			"bk_idc_area_id":  float64(42),
			"bk_host_innerip": "127.0.0.1",
		},
	}

	var got HostEventInfo
	if err := conv.MapToStruct(event, &got); err != nil {
		t.Fatalf("MapToStruct() error = %v", err)
	}

	if got.BKDetail.BKIDCAreaID != 42 {
		t.Fatalf("BKIDCAreaID = %d, want 42", got.BKDetail.BKIDCAreaID)
	}
}
