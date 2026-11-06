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

package processv2config

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/structpb"
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
		TemplateName:        "test-template-name",
		ProcessName:         "test-process-name",
		HostID:              2,
		Set:                 "deploy_policy",
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

func Test_convProcessConfigToTypesCustomConfigContextWithBSONContainers(t *testing.T) {
	config := &ProcessConfig{
		Name:        "test-process-config-name",
		ProcessName: "test-process-name",
		HostID:      2,
		CustomConfigContext: map[string]any{
			"paths": primitive.A{"/tmp/a", "/tmp/b"},
			"items": primitive.A{
				primitive.M{"enabled": true},
				primitive.A{int32(80), int64(443)},
			},
			"nested": map[string]any{
				"listen": []any{"127.0.0.1"},
			},
		},
	}

	got := convProcessConfigToTypes(config)

	want := map[string]any{
		"paths": []any{"/tmp/a", "/tmp/b"},
		"items": []any{
			map[string]any{"enabled": true},
			[]any{int32(80), int64(443)},
		},
		"nested": map[string]any{
			"listen": []any{"127.0.0.1"},
		},
	}
	if !reflect.DeepEqual(got.CustomConfigContext, want) {
		t.Fatalf("CustomConfigContext = %#v, want %#v", got.CustomConfigContext, want)
	}

	// the plugin config file list APIs pass this map to structpb.NewStruct.
	if _, err := structpb.NewStruct(got.CustomConfigContext); err != nil {
		t.Fatalf("structpb.NewStruct() error = %v", err)
	}
}
