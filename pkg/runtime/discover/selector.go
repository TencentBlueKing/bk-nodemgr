/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package discover ...
package discover

import "math/rand"

// RandomSelector this defines the random selector.
type RandomSelector struct{}

// Select select one instance from the instances.
func (r *RandomSelector) Select(instances []Instance) (Instance, error) {
	length := len(instances)

	if length == 0 {
		return Instance{}, ErrServiceNotFound()
	}

	// this just is a simple random selector.
	// nolint: gosec
	idx := rand.Intn(length)

	return instances[idx], nil
}
