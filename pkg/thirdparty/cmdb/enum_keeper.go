/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"context"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
)

type iEnumResourceKeeper interface {
	getValue(key string) string
	getKey(value string) string
	update(ctx context.Context) error
	values() []string
}

func newCloudVendorKeeper(cli *cli) *enumCloudVendorKeeper {
	return &enumCloudVendorKeeper{
		enumBasicKeeper: enumBasicKeeper{
			cli:         cli,
			objectID:    "plat",
			attributeID: "bk_cloud_vendor",
			mapping:     make(map[string]string),
		}}
}

type enumCloudVendorKeeper struct {
	enumBasicKeeper
}

func newCPUArchKeeper(cli *cli) *enumCPUArchKeeper {
	return &enumCPUArchKeeper{
		enumBasicKeeper: enumBasicKeeper{
			cli:          cli,
			objectID:     "host",
			attributeID:  "bk_cpu_architecture",
			mapping:      make(map[string]string),
			unknownValue: criteria.CPUArchUnknown,
		},
		cpuArchMapping: make(map[string]string),
	}
}

type enumCPUArchKeeper struct {
	enumBasicKeeper

	// cpuArchMapping is a map of platform-normalized cpu arch to cmdb cpu arch name.
	cpuArchMapping      map[string]string
	cpuArchMappingMutex sync.RWMutex
}

func (keeper *enumCPUArchKeeper) getValue(key string) string {
	value, err := platform.NormalizeArch(keeper.enumBasicKeeper.getValue(key))
	if err != nil {
		return keeper.enumBasicKeeper.unknownValue
	}

	return value
}

func (keeper *enumCPUArchKeeper) getKey(value string) string {
	keeper.cpuArchMappingMutex.RLock()
	cpuArchName, ok := keeper.cpuArchMapping[value]
	if !ok {
		keeper.cpuArchMappingMutex.RUnlock()
		return keeper.unknownKey
	}
	keeper.cpuArchMappingMutex.RUnlock()

	return keeper.enumBasicKeeper.getKey(cpuArchName)
}

func (keeper *enumCPUArchKeeper) update(ctx context.Context) error {
	if err := keeper.enumBasicKeeper.update(ctx); err != nil {
		return err
	}

	keeper.enumBasicKeeper.mutex.RLock()
	values := conv.MapToSlice(keeper.enumBasicKeeper.mapping)
	keeper.enumBasicKeeper.mutex.RUnlock()

	keeper.cpuArchMappingMutex.Lock()
	for _, v := range values {
		normalizeArch, err := platform.NormalizeArch(v)
		if err != nil {
			continue
		}

		keeper.cpuArchMapping[normalizeArch] = v
	}
	keeper.cpuArchMappingMutex.Unlock()

	return nil
}

func newOSTypeKeeper(cli *cli) *enumOSTypeKeeper {
	return &enumOSTypeKeeper{
		enumBasicKeeper: enumBasicKeeper{
			cli:          cli,
			objectID:     "host",
			attributeID:  "bk_os_type",
			mapping:      make(map[string]string),
			unknownValue: string(criteria.OSUnknown),
		},
		osTypeMapping: make(map[string]string),
	}
}

type enumOSTypeKeeper struct {
	enumBasicKeeper

	// osTypeMapping is a map of platform-normalized OS to cmdb os type name.
	osTypeMapping      map[string]string
	osTypeMappingMutex sync.RWMutex
}

func (keeper *enumOSTypeKeeper) getValue(key string) string {
	value, err := platform.NormalizeOS(keeper.enumBasicKeeper.getValue(key))
	if err != nil {
		return keeper.enumBasicKeeper.unknownValue
	}

	return string(value)
}

func (keeper *enumOSTypeKeeper) getKey(value string) string {
	keeper.osTypeMappingMutex.RLock()
	osTypeName, ok := keeper.osTypeMapping[value]
	if !ok {
		keeper.osTypeMappingMutex.RUnlock()
		return keeper.unknownKey
	}
	keeper.osTypeMappingMutex.RUnlock()

	return keeper.enumBasicKeeper.getKey(osTypeName)
}

func (keeper *enumOSTypeKeeper) update(ctx context.Context) error {
	if err := keeper.enumBasicKeeper.update(ctx); err != nil {
		return err
	}

	keeper.enumBasicKeeper.mutex.RLock()
	values := conv.MapToSlice(keeper.enumBasicKeeper.mapping)
	keeper.enumBasicKeeper.mutex.RUnlock()

	keeper.osTypeMappingMutex.Lock()
	for _, v := range values {
		normalizedOS, err := platform.NormalizeOS(v)
		if err != nil {
			continue
		}

		keeper.osTypeMapping[string(normalizedOS)] = v
	}
	keeper.osTypeMappingMutex.Unlock()

	return nil
}

// enumBasicKeeper provides cmdb enum resource in object attributes auto query and keep.
type enumBasicKeeper struct {
	cli          *cli
	objectID     string
	attributeID  string
	unknownKey   string
	unknownValue string

	mutex   sync.RWMutex
	mapping map[string]string
}

func (keeper *enumBasicKeeper) getValue(key string) string {
	keeper.mutex.RLock()
	defer keeper.mutex.RUnlock()

	value, ok := keeper.mapping[key]
	if ok {
		return value
	}

	return keeper.unknownValue
}

func (keeper *enumBasicKeeper) getKey(value string) string {
	keeper.mutex.RLock()
	defer keeper.mutex.RUnlock()

	for k, v := range keeper.mapping {
		if v == value {
			return k
		}
	}

	return keeper.unknownKey
}

func (keeper *enumBasicKeeper) update(ctx context.Context) error {
	result, err := keeper.searchObjectAttributeEnumOption(
		ctx, keeper.objectID, CCNoBusinessID, keeper.attributeID)
	if err != nil {
		return err
	}

	keeper.mutex.Lock()
	defer keeper.mutex.Unlock()

	keeper.mapping = make(map[string]string)
	for _, option := range result {
		keeper.mapping[option.Key] = option.Value
	}

	return nil
}

func (keeper *enumBasicKeeper) values() []string {
	keeper.mutex.RLock()
	defer keeper.mutex.RUnlock()

	return conv.MapToSlice(keeper.mapping)
}

// searchObjectAttributeEnumOption search cmdb object attribute's option, like bk_cloud_vendor and bk_os_type.
func (keeper *enumBasicKeeper) searchObjectAttributeEnumOption(
	ctx context.Context, objID string, bizID int64, objAttrID string) ([]*EnumOption, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	req := &SearchObjectAttributeReq{
		TenantID: tenantID,
		BKObjID:  objID,
		BKBizID:  bizID,
	}

	resp, err := keeper.cli.searchObjectAttribute(ctx, req)
	if err != nil {
		return nil, err
	}

	var result []*EnumOption
	for _, objAttribute := range *resp {
		if objAttribute.BKPropertyID != objAttrID {
			continue
		}

		options, ok := objAttribute.Option.([]any)
		if !ok {
			return nil, fmt.Errorf("try to convert type to []any failed, bk_property_id(%s), option(%v)", objAttrID,
				objAttribute.Option)
		}

		result = make([]*EnumOption, len(options))
		for index, option := range options {
			mapOption, ok := option.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("try to convert type to map[string]any failed, bk_property_id(%s), option(%v)",
					objAttrID, option)
			}

			key, ok := mapOption["id"].(string)
			if !ok {
				return nil, fmt.Errorf("try to convert type to string failed, bk_property_id(%s), option[id](%v)",
					objAttrID, mapOption["id"])
			}

			name, ok := mapOption["name"].(string)
			if !ok {
				return nil, fmt.Errorf("try to convert type to string failed, bk_property_id(%s), option[name](%v)",
					objAttrID, mapOption["name"])
			}

			result[index] = &EnumOption{
				Key:   key,
				Value: name,
			}
		}
	}

	return result, nil
}
