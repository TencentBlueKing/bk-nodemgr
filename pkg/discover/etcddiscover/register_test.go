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

package etcddiscover

import (
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
)

func TestProviderEtcdRegisterPrevalidatesAllInstances(t *testing.T) {
	provider := NewProviderEtcd(&config.Etcd{})
	validInstance := discover.Instance{
		ID:   "valid-instance",
		Name: "valid instance",
	}
	invalidInstance := discover.Instance{
		Name: "missing id",
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Register() panicked before validating all instances: %v", recovered)
		}
	}()

	if err := provider.Register(discover.ServiceNameBackend, validInstance, invalidInstance); err != nil {
		if !strings.Contains(err.Error(), "invalid instance ID") {
			t.Fatalf("Register() error = %v, want invalid instance ID", err)
		}

		return
	}

	t.Fatal("Register() error is nil, want invalid instance error")
}
