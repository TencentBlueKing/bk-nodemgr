/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"errors"
	"fmt"
)

// Generation represents a node generation.
type Generation int64

const (
	// GenerationAll means it is for all generation.
	GenerationAll Generation = 0

	// Generation1 means this node is the first generation.
	Generation1 Generation = 1

	// Generation2 means this node is the second generation.
	Generation2 Generation = 2
)

// Validate validates the node generation.
func (gen Generation) Validate() error {
	switch gen {
	case Generation1:
		return fmt.Errorf("generation 1 now no support")
	case Generation2:
		return nil
	default:
		return errors.New("invalid generation")
	}
}

// GenerationListToInt64List converts a generation list to a int64 list.
func GenerationListToInt64List(genList []Generation) []int64 {
	data := make([]int64, len(genList))
	for idx, gen := range genList {
		data[idx] = int64(gen)
	}

	return data
}

// Int64ListToGenerationList converts a int64 list to a generation list.
func Int64ListToGenerationList(genList []int64) []Generation {
	data := make([]Generation, len(genList))
	for idx, gen := range genList {
		data[idx] = Generation(gen)
	}

	return data
}
