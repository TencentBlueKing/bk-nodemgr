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

package v3

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate validates a package event list request.
func (x *PackageEventListReq) Validate() error {
	if x.GetOnlyCount() {
		return nil
	}
	return x.GetPage().Validate()
}

// AutoConvert applies package event list defaults.
func (x *PackageEventListReq) AutoConvert() {
	// No defaults are required for package event queries.
}

// Validate validates a package event distinct request.
func (x *PackageEventDistinctReq) Validate() error {
	return nil
}

// AutoConvert applies package event distinct defaults.
func (x *PackageEventDistinctReq) AutoConvert() {
	// No defaults are required for package event queries.
}

// ConvertPageToTypes converts a package event query page.
func (x *PackageEventListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts request conditions to domain conditions.
func (x *PackageEventListReq) ConvertConditionsToTypes() ([]*types.PackageEventCondition, error) {
	return convertPackageEventConditionsToTypes(x.GetConditions())
}

// ConvertConditionsToTypes converts request conditions to domain conditions.
func (x *PackageEventDistinctReq) ConvertConditionsToTypes() ([]*types.PackageEventCondition, error) {
	return convertPackageEventConditionsToTypes(x.GetConditions())
}

// ConvertConditionsFromTypes converts domain conditions to a list request.
func (x *PackageEventListReq) ConvertConditionsFromTypes(conditions ...*types.PackageEventCondition) {
	x.Conditions = convertPackageEventConditionsFromTypes(conditions...)
}

// ConvertConditionsFromTypes converts domain conditions to a distinct request.
func (x *PackageEventDistinctReq) ConvertConditionsFromTypes(conditions ...*types.PackageEventCondition) {
	x.Conditions = convertPackageEventConditionsFromTypes(conditions...)
}

// ConvertPackageEventsFromTypes converts domain package events to a response.
func (x *PackageEventListResp) ConvertPackageEventsFromTypes(total int64, events []*types.PackageEvent) {
	items := make([]*PackageEventInfo, len(events))
	for i, event := range events {
		if event == nil {
			continue
		}
		items[i] = &PackageEventInfo{TenantId: event.TenantID, Name: event.Name, EventType: string(event.EventType),
			ReleaseType: string(event.ReleaseType), Generation: int64(event.Generation), OsType: string(event.OSType),
			CpuArch: string(event.CPUArch), Version: event.Version, OperateTime: event.OperateTime.UnixMilli(), Operator: event.Operator}
	}
	x.Data = &PackageEventListResp_Data{Total: total, Items: items}
}

// ConvertPackageEventsToTypes converts a response to domain package events.
func (x *PackageEventListResp_Data) ConvertPackageEventsToTypes() (int64, []*types.PackageEvent) {
	if x == nil {
		return 0, nil
	}
	items := x.GetItems()
	events := make([]*types.PackageEvent, len(items))
	for i, item := range items {
		if item == nil {
			continue
		}
		events[i] = &types.PackageEvent{TenantID: item.GetTenantId(), Name: item.GetName(),
			EventType: types.PackageEventType(item.GetEventType()), ReleaseType: types.ReleaseType(item.GetReleaseType()),
			Generation: types.Generation(item.GetGeneration()), OSType: criteria.OSType(item.GetOsType()),
			CPUArch: criteria.CPUArch(item.GetCpuArch()), Version: item.GetVersion(),
			OperateTime: time.UnixMilli(item.GetOperateTime()), Operator: item.GetOperator()}
	}
	return x.GetTotal(), events
}

// ConvertResultFromTypes converts a domain distinct result to a response.
func (x *PackageEventDistinctResp) ConvertResultFromTypes(result *types.PackageEventDistinctResult) {
	if result == nil {
		return
	}

	osTypes := conv.SliceToSlice(result.OSType, func(os criteria.OSType) string { return string(os) })
	cpuArchs := conv.SliceToSlice(result.CPUArch, func(arch criteria.CPUArch) string { return string(arch) })
	x.Data = &PackageEventDistinctResp_Data{ReleaseType: types.ReleaseTypeListToStringList(result.ReleaseType),
		EventType: types.PackageEventTypeListToStringList(result.EventType),
		OsType:    normalizePackageEventPlatformStrings(osTypes, string(criteria.OSUnknown)),
		CpuArch:   normalizePackageEventPlatformStrings(cpuArchs, string(criteria.CPUArchUnknown)),
		Version:   result.Version, Operator: result.Operator}
}

// ConvertResultToTypes converts a distinct response to a domain result.
func (x *PackageEventDistinctResp_Data) ConvertResultToTypes() (*types.PackageEventDistinctResult, error) {
	if x == nil {
		return &types.PackageEventDistinctResult{}, nil
	}
	osTypes, err := conv.SliceToSliceWithError(normalizePackageEventPlatformStrings(x.GetOsType(), string(criteria.OSUnknown)),
		func(os string) (criteria.OSType, error) {
			return criteria.OSType(os), criteria.OSType(os).Validate()
		})
	if err != nil {
		return nil, fmt.Errorf("failed to convert os type list: %w", err)
	}
	cpuArchs, err := conv.SliceToSliceWithError(normalizePackageEventPlatformStrings(x.GetCpuArch(), string(criteria.CPUArchUnknown)),
		func(arch string) (criteria.CPUArch, error) {
			return criteria.CPUArch(arch), criteria.CPUArch(arch).Validate()
		})
	if err != nil {
		return nil, fmt.Errorf("failed to convert cpu arch list: %w", err)
	}
	return &types.PackageEventDistinctResult{ReleaseType: types.StringListToReleaseTypeList(x.GetReleaseType()),
		EventType: types.StringListToPackageEventTypeList(x.GetEventType()), OSType: osTypes, CPUArch: cpuArchs,
		Version: x.GetVersion(), Operator: x.GetOperator()}, nil
}

func normalizePackageEventPlatformStrings(values []string, unknown string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			value = unknown
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func convertPackageEventConditionsToTypes(conditions []*PackageEventQueryConditions) ([]*types.PackageEventCondition, error) {
	result := make([]*types.PackageEventCondition, 0, len(conditions))
	for _, condition := range conditions {
		if condition == nil {
			continue
		}
		include, err := convertPackageEventExactFieldsToTypes(condition.GetExactInclude())
		if err != nil {
			return nil, err
		}
		exclude, err := convertPackageEventExactFieldsToTypes(condition.GetExactExclude())
		if err != nil {
			return nil, err
		}
		item := &types.PackageEventCondition{ExactInclude: include, ExactExclude: exclude}
		if condition.GetFuzzyInclude() != nil {
			item.FuzzyInclude = &types.PackageEventFuzzyFields{}
		}
		if condition.GetFuzzyExclude() != nil {
			item.FuzzyExclude = &types.PackageEventFuzzyFields{}
		}
		if value := condition.GetOperateTimeRange(); value != nil {
			item.OperateTimeRange = &types.TimeRange{StartTime: time.Unix(value.GetStartTimestampSec(), 0), EndTime: time.Unix(value.GetEndTimestampSec(), 0)}
		}
		result = append(result, item)
	}
	return result, nil
}
func convertPackageEventExactFieldsToTypes(fields *PackageEventExactConditions) (*types.PackageEventExactFields, error) {
	if fields == nil {
		return nil, nil
	}
	osTypes, err := criteria.StringListToOSTypeList(fields.GetOsType())
	if err != nil {
		return nil, err
	}
	cpuArchs, err := criteria.StringListToCPUArchList(fields.GetCpuArch())
	if err != nil {
		return nil, err
	}
	return &types.PackageEventExactFields{EventType: types.StringListToPackageEventTypeList(fields.GetEventType()),
		ReleaseType: types.StringListToReleaseTypeList(fields.GetReleaseType()), Version: fields.GetVersion(),
		OSType: osTypes, CPUArch: cpuArchs, Generation: types.Int64ListToGenerationList(fields.GetGeneration()),
		Operator: fields.GetOperator()}, nil
}
func convertPackageEventConditionsFromTypes(conditions ...*types.PackageEventCondition) []*PackageEventQueryConditions {
	result := make([]*PackageEventQueryConditions, 0, len(conditions))
	for _, condition := range conditions {
		if condition == nil {
			continue
		}
		item := &PackageEventQueryConditions{ExactInclude: convertPackageEventExactFieldsFromTypes(condition.ExactInclude),
			ExactExclude: convertPackageEventExactFieldsFromTypes(condition.ExactExclude)}
		if condition.FuzzyInclude != nil {
			item.FuzzyInclude = &PackageEventFuzzyConditions{}
		}
		if condition.FuzzyExclude != nil {
			item.FuzzyExclude = &PackageEventFuzzyConditions{}
		}
		if value := condition.OperateTimeRange; value != nil {
			item.OperateTimeRange = &PackageEventQueryTimeRange{StartTimestampSec: value.StartTime.Unix(), EndTimestampSec: value.EndTime.Unix()}
		}
		result = append(result, item)
	}
	return result
}
func convertPackageEventExactFieldsFromTypes(fields *types.PackageEventExactFields) *PackageEventExactConditions {
	if fields == nil {
		return nil
	}
	return &PackageEventExactConditions{Generation: types.GenerationListToInt64List(fields.Generation),
		OsType: criteria.OSTypeListToStringList(fields.OSType), CpuArch: criteria.CPUArchListToStringList(fields.CPUArch),
		ReleaseType: types.ReleaseTypeListToStringList(fields.ReleaseType), Operator: fields.Operator,
		EventType: types.PackageEventTypeListToStringList(fields.EventType), Version: fields.Version}
}

// ConvertFromTypes converts a package event page and filters to a list request.
func (x *PackageEventListReq) ConvertFromTypes(page types.Page, conditions ...*types.PackageEventCondition) error {
	x.Page = new(Page)
	x.Page.ConvertFromTypes(page)
	x.OnlyCount = false
	x.ConvertConditionsFromTypes(conditions...)
	return x.Validate()
}

// ConvertCountFromTypes converts filters to a count-only request.
func (x *PackageEventListReq) ConvertCountFromTypes(conditions ...*types.PackageEventCondition) {
	x.Page = nil
	x.OnlyCount = true
	x.ConvertConditionsFromTypes(conditions...)
}

// ConvertFromTypes converts distinct fields and filters to a request.
func (x *PackageEventDistinctReq) ConvertFromTypes(request types.PackageEventDistinctRequest, conditions ...*types.PackageEventCondition) {
	x.Fields = &PackageEventDistinctFields{
		ReleaseType: request.ReleaseType,
		EventType:   request.EventType,
		OsType:      request.OSType,
		CpuArch:     request.CPUArch,
		Version:     request.Version,
		Operator:    request.Operator,
	}
	x.ConvertConditionsFromTypes(conditions...)
}

// ConvertDistinctRequestToTypes converts the selected event fields.
func (x *PackageEventDistinctReq) ConvertDistinctRequestToTypes() types.PackageEventDistinctRequest {
	fields := x.GetFields()
	return types.PackageEventDistinctRequest{
		ReleaseType: fields.GetReleaseType(),
		EventType:   fields.GetEventType(),
		OSType:      fields.GetOsType(),
		CPUArch:     fields.GetCpuArch(),
		Version:     fields.GetVersion(),
		Operator:    fields.GetOperator(),
	}
}
