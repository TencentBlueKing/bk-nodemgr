/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package base

import (
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SortSeparator is the separator for sort keys.
const SortSeparator = ","

// ParsePage parses the page information and sets it to the FindOptions.
func ParsePage(page types.Page) *options.FindOptions {
	findOpt := new(options.FindOptions)
	if page.Offset > 0 {
		findOpt.SetSkip(int64(page.Offset))
	}
	if page.Limit > 0 {
		findOpt.SetLimit(int64(page.Limit))
	}

	if page.Sort != "" {
		sortKeys := strings.Split(page.Sort, SortSeparator)
		sortOpt := make(bson.D, len(sortKeys))
		for i, key := range sortKeys {
			if strings.HasPrefix(key, "-") {
				sortOpt[i] = bson.E{Key: strings.TrimPrefix(key, "-"), Value: -1}
			} else {
				sortOpt[i] = bson.E{Key: key, Value: 1}
			}
		}
	}

	return findOpt
}
