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

package processconfig

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
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
		TemplateName:        "test-template-name",
		ProcessName:         "test-process-name",
		HostID:              2,
		Set:                 "deploy_policy",
		CustomConfigContext: want,
	})
	if config.TemplateName != "test-template-name" {
		t.Fatalf("convProcessConfigFromTypes() TemplateName = %q, want %q", config.TemplateName, "test-template-name")
	}
	if config.Set != "deploy_policy" {
		t.Fatalf("convProcessConfigFromTypes() Set = %v, want %v", config.Set, "deploy_policy")
	}
	if !reflect.DeepEqual(config.CustomConfigContext, want) {
		t.Fatalf("convProcessConfigFromTypes() CustomConfigContext = %v, want %v", config.CustomConfigContext, want)
	}

	got := convProcessConfigToTypes(config)
	if got.TemplateName != "test-template-name" {
		t.Fatalf("convProcessConfigToTypes() TemplateName = %q, want %q", got.TemplateName, "test-template-name")
	}
	if got.Set != "deploy_policy" {
		t.Fatalf("convProcessConfigToTypes() Set = %v, want %v", got.Set, "deploy_policy")
	}
	if !reflect.DeepEqual(got.CustomConfigContext, want) {
		t.Fatalf("convProcessConfigToTypes() CustomConfigContext = %v, want %v", got.CustomConfigContext, want)
	}
}

func TestWithTemplateName(t *testing.T) {
	filter := WithTemplateName("template.conf")(base.AliveFilter())
	if len(filter) != 2 {
		t.Fatalf("WithTemplateName() filter len = %d, want %d", len(filter), 2)
	}
	if filter[1].Key != FieldKeyTemplateName {
		t.Fatalf("WithTemplateName() key = %q, want %q", filter[1].Key, FieldKeyTemplateName)
	}
}
