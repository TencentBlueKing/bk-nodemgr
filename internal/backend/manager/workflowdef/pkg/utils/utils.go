/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"fmt"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ComparePlatforms compares the expected and actual package platforms.
func ComparePlatforms(expected []platfmt.Platform, releases []*types.ReleasePlugin) error {
	expectedSet := make(map[platfmt.Platform]struct{}, len(expected))
	for _, platform := range expected {
		expectedSet[platform] = struct{}{}
	}

	actual := make([]platfmt.Platform, 0, len(releases))
	actualSet := make(map[platfmt.Platform]struct{}, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}

		actual = append(actual, release.Platform)
		actualSet[release.Platform] = struct{}{}
	}

	if platformSetsEqual(expectedSet, actualSet) {
		return nil
	}

	return fmt.Errorf("compare package platform set mismatch, expected(%v), actual(%v)", expected, actual)
}

func platformSetsEqual(left, right map[platfmt.Platform]struct{}) bool {
	if len(left) != len(right) {
		return false
	}

	for platform := range left {
		if _, ok := right[platform]; !ok {
			return false
		}
	}

	return true
}
