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

package iamv4

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestToResourcePreservesAttributes(t *testing.T) {
	attrs := map[string]interface{}{
		"_bk_iam_path_": "/networkarea,0/",
	}

	resource := toResource(types.IAMResource{ID: "0", Attributes: attrs})
	if resource.ID != "0" {
		t.Fatalf("toResource() ID = %q; want 0", resource.ID)
	}
	if !reflect.DeepEqual(resource.Attributes, attrs) {
		t.Fatalf("toResource() Attributes = %#v; want %#v", resource.Attributes, attrs)
	}
}

func TestFirstResourcePreservesAttributes(t *testing.T) {
	attrs := map[string]interface{}{
		"_bk_iam_path_": "/networkarea,0/",
	}

	resource := firstResource([]types.IAMResource{{ID: "0", Attributes: attrs}})
	if resource == nil {
		t.Fatal("firstResource() = nil; want resource")
	}
	if !reflect.DeepEqual(resource.Attributes, attrs) {
		t.Fatalf("firstResource() Attributes = %#v; want %#v", resource.Attributes, attrs)
	}
}

func TestFirstResourceEmpty(t *testing.T) {
	if resource := firstResource(nil); resource != nil {
		t.Fatalf("firstResource(nil) = %#v; want nil", resource)
	}
}
