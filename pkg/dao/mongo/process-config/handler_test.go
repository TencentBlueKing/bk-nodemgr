/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package processconfig

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func Test_convProcessConfigCustomConfigContext(t *testing.T) {
	want := map[string]any{
		"env": "prod",
		"nested": map[string]any{
			"enabled": true,
		},
	}

	config := convProcessConfigFromTypes(&types.ProcessConfig{
		Name:                "test-process-config-name",
		ProcessName:         "test-process-name",
		HostID:              2,
		CustomConfigContext: want,
	})
	if !reflect.DeepEqual(config.CustomConfigContext, want) {
		t.Fatalf("convProcessConfigFromTypes() CustomConfigContext = %v, want %v", config.CustomConfigContext, want)
	}

	got := convProcessConfigToTypes(config)
	if !reflect.DeepEqual(got.CustomConfigContext, want) {
		t.Fatalf("convProcessConfigToTypes() CustomConfigContext = %v, want %v", got.CustomConfigContext, want)
	}
}
