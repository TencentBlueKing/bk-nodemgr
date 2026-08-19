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

package metrics

import (
	"github.com/bits-and-blooms/bitset"
)

const (
	defaultSize = 2 << 24
	funcsSize   = 6
)

func bloomSeedsDefault(i int) uint {
	seeds := [funcsSize]uint{7, 11, 13, 31, 37, 61}

	return seeds[i]
}

type bloomFilter struct {
	Set   *bitset.BitSet
	Funcs []simpleHash
}

func newBloomFilter() *bloomFilter {
	bf := new(bloomFilter)
	bf.Funcs = make([]simpleHash, funcsSize)

	for i := 0; i < len(bf.Funcs); i++ {
		bf.Funcs[i] = simpleHash{defaultSize, bloomSeedsDefault(i)}
	}
	bf.Set = bitset.New(defaultSize)

	return bf
}

func (bf *bloomFilter) add(value string) {
	for _, f := range bf.Funcs {
		bf.Set.Set(f.hash(value))
	}
}

func (bf *bloomFilter) contains(value string) bool {
	if value == "" {
		return false
	}

	for _, f := range bf.Funcs {
		if !bf.Set.Test(f.hash(value)) {
			return false
		}
	}

	return true
}

type simpleHash struct {
	Cap  uint
	Seed uint
}

func (s *simpleHash) hash(value string) uint {
	var result uint
	for i := 0; i < len(value); i++ {
		result = result*s.Seed + uint(value[i])
	}

	return (s.Cap - 1) & result
}
