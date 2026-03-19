/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	valueGroupKeyFileMandatoryTCP = "VALUE_GROUP_FILE_MANDATORY_TCP"
	valueGroupKeyProxyLoggerLevel = "VALUE_GROUP_PROXY_LOGGER_LEVEL"
	valueGroupKeyProxyLoggerPath  = "VALUE_GROUP_PROXY_LOGGER_PATH"
)

// nolint: gochecknoglobals
var valueGroupKeyMap = map[string]struct{}{
	valueGroupKeyFileMandatoryTCP: {},
	valueGroupKeyProxyLoggerLevel: {},
	valueGroupKeyProxyLoggerPath:  {},
}

func isValueGroupKey(key string) bool {
	_, ok := valueGroupKeyMap[key]
	return ok
}

func findTemplateItem(blocks []types.ConfigPolicyTemplateBlock, blockID, itemID string) (
	types.ConfigPolicyTemplateItem, bool) {

	for _, block := range blocks {
		if block.ID == blockID {
			for _, item := range block.Items {
				if item.ID == itemID {
					return item, true
				}
			}
		}
	}

	return types.ConfigPolicyTemplateItem{}, false
}
