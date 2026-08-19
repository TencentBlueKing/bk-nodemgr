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

package compatibility

import (
	"errors"
	"fmt"
)

// Validate checks whether the policy is structurally valid after normalization.
func (p Policy) Validate() error {
	if p.EnabledPlugins == nil {
		return errors.New("enabled_plugins is missing")
	}
	if p.DisabledBiz == nil {
		return errors.New("disabled_biz is missing")
	}

	policy := p.Normalize()

	for _, pluginName := range policy.EnabledPlugins {
		if pluginName == "" {
			return errors.New("enabled_plugins item is empty")
		}
	}

	for _, item := range policy.DisabledBiz {
		if item.TenantID == "" {
			return errors.New("disabled_biz tenant_id is empty")
		}
		if item.BKBizID <= 0 {
			return fmt.Errorf("disabled_biz bk_biz_id is invalid: %d", item.BKBizID)
		}
	}

	return nil
}

// ValidateForWrite strictly validates a policy before persisting it.
func ValidateForWrite(policy Policy) error {
	return policy.Validate()
}
