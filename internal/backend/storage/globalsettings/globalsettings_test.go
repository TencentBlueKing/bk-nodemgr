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

package globalsettings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestValidateNetworkUnitSegmentRuleSettings(t *testing.T) {
	t.Run("ignore unrelated setting", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: "other_setting",
			Value:       "not-json-at-all",
		}})
		require.NoError(t, err)
	})

	t.Run("reject invalid json for target setting", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: types.GlobalSettingNameNetworkUnitSegmentRules,
			Value:       "{",
		}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), types.GlobalSettingNameNetworkUnitSegmentRules)
	})

	t.Run("accept valid json shape for target setting", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: types.GlobalSettingNameNetworkUnitSegmentRules,
			Value:       `{"1001":{"rules":[{"cidrs":["10.0.0.0/24"],"bk_networkunit_id":200101}]}}`,
		}})
		require.NoError(t, err)
	})

	t.Run("reject empty cidrs", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: types.GlobalSettingNameNetworkUnitSegmentRules,
			Value:       `{"1001":{"rules":[{"cidrs":[],"bk_networkunit_id":200101}]}}`,
		}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cidrs")
	})

	t.Run("reject missing networkunit id", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: types.GlobalSettingNameNetworkUnitSegmentRules,
			Value:       `{"1001":{"rules":[{"cidrs":["10.0.0.0/24"],"bk_networkunit_id":-1}]}}`,
		}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bk_networkunit_id")
	})

	t.Run("reject ipv6 cidr", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: types.GlobalSettingNameNetworkUnitSegmentRules,
			Value:       `{"1001":{"rules":[{"cidrs":["2001:db8::/64"],"bk_networkunit_id":200101}]}}`,
		}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "IPv6")
	})

	t.Run("reject fallback before specific rule", func(t *testing.T) {
		err := validateNetworkUnitSegmentRuleSettings([]*types.GlobalSettings{{
			SettingName: types.GlobalSettingNameNetworkUnitSegmentRules,
			Value:       `{"1001":{"rules":[{"cidrs":["0.0.0.0/0"],"bk_networkunit_id":200199},{"cidrs":["10.0.0.0/24"],"bk_networkunit_id":200101}]}}`,
		}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fallback")
	})
}
