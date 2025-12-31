/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// 30 days for max operate time range.
	maxOperateTimeRangeDuration = 365 * 24 * time.Hour
)

// Validate check body.
func (x *TopoEventListReq) Validate() error {
	if err := validatePage(x.GetPage()); err != nil {
		return err
	}

	if err := validateTimeRange(x.GetOperateTimeRange(), maxOperateTimeRangeDuration); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *TopoEventListReq) AutoConvert() {
	if x.GetOperateTimeRange() == nil {
		x.OperateTimeRange = &TimeRange{
			StartTimestampSec: time.Now().Add(-1 * maxOperateTimeRangeDuration).Unix(),
			EndTimestampSec:   time.Now().Unix(),
		}
	}
}

// ConvertPageToTypes convert page to types.
func (x *TopoEventListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoEventListReq) ConvertConditionsToTypes() (*types.TopoEventCondition, error) {
	return convertTopoEventConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *TopoEventListReq) ConvertConditionsFromTypes(condition *types.TopoEventCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertTopoEventConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.OperateTimeRange = timeRange
	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertTopoEventsToTypes convert topo events to types.
func (x *TopoEventListResp) ConvertTopoEventsToTypes() (int64, []*types.TopoEvent) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.TopoEvent, len(items))
	for idx, item := range items {
		result[idx] = &types.TopoEvent{
			TenantID:        item.GetTenantId(),
			Type:            types.TopoEventType(item.GetType()),
			NetworkAreaID:   item.GetBkNetworkareaId(),
			NetworkAreaName: item.GetBkNetworkareaName(),
			NetworkUnitID:   item.GetBkNetworkunitId(),
			NetworkUnitName: item.GetBkNetworkunitName(),
			AccessPointID:   item.GetAccesspointId(),
			AccessPointName: item.GetAccesspointName(),
			OperateTime:     time.UnixMilli(item.GetOperateTime()),
			Operator:        item.GetOperator(),
		}
	}

	return data.GetTotal(), result
}

// ConvertTopoEventsFromTypes convert types to topo events.
func (x *TopoEventListResp) ConvertTopoEventsFromTypes(total int64, events []*types.TopoEvent) {
	items := make([]*TopoEvent, len(events))
	for idx, event := range events {
		item := newEmptyTopoEvent()
		*item.TenantId = event.TenantID
		*item.Type = string(event.Type)
		*item.BkNetworkareaId = event.NetworkAreaID
		*item.BkNetworkareaName = event.NetworkAreaName
		*item.BkNetworkunitId = event.NetworkUnitID
		*item.BkNetworkunitName = event.NetworkUnitName
		*item.AccesspointId = event.AccessPointID
		*item.AccesspointName = event.AccessPointName
		*item.OperateTime = event.OperateTime.UnixMilli()
		*item.Operator = event.Operator

		items[idx] = item
	}

	x.Data = &TopoEventListResp_Data{
		Total: total,
		Items: items,
	}
}

// Validate check body.
func (x *TopoEventDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoEventDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoEventDistinctReq) ConvertConditionsToTypes() (*types.TopoEventCondition, error) {
	return convertTopoEventConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *TopoEventDistinctReq) ConvertConditionsFromTypes(condition *types.TopoEventCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertTopoEventConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.OperateTimeRange = timeRange
	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertResultFromTypes convert result from types.
func (x *TopoEventDistinctResp) ConvertResultFromTypes(result *types.TopoEventDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &TopoEventDistinctResp_Data{
		Type:            formatRespSlice(types.TopoEventTypeListToStringList(result.Type)),
		BkNetworkareaId: formatRespSlice(result.NetworkAreaID),
		BkNetworkunitId: formatRespSlice(result.NetworkUnitID),
		AccesspointId:   formatRespSlice(result.AccessPointID),
		Operator:        formatRespSlice(result.Operator),
	}
}

func newEmptyTopoEvent() *TopoEvent {
	return &TopoEvent{
		TenantId:          new(string),
		Type:              new(string),
		BkNetworkareaId:   new(int64),
		BkNetworkareaName: new(string),
		BkNetworkunitId:   new(int64),
		BkNetworkunitName: new(string),
		AccesspointId:     new(int64),
		AccesspointName:   new(string),
		OperateTime:       new(int64),
		Operator:          new(string),
	}
}

func convertTopoEventConditionsToTypes(
	exactCond *TopoEventExactConditions,
	fuzzyCond *TopoEventFuzzyConditions,
	timeRange *TimeRange) (*types.TopoEventCondition, error) {

	condition := &types.TopoEventCondition{}

	if timeRange != nil {
		operateTimeRange, err := convertTimeRangeToTypes(timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert operate time range: %w", err)
		}
		condition.OperateTimeRange = operateTimeRange
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.TopoEventExactFields{
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			AccessPointID: exactCond.GetAccesspointId(),
			Type:          types.StringListToTopoEventTypeList(exactCond.GetType()),
			Operator:      exactCond.GetOperator(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.TopoEventFuzzyFields{
			NetworkAreaName: fuzzyCond.GetBkNetworkareaName(),
			NetworkUnitName: fuzzyCond.GetBkNetworkunitName(),
		}
	}

	return condition, nil
}

func convertTopoEventConditionsFromTypes(condition *types.TopoEventCondition) (
	*TopoEventExactConditions, *TopoEventFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var timeRange *TimeRange
	var exactCond *TopoEventExactConditions
	var fuzzyCond *TopoEventFuzzyConditions

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	if condition.ExactInclude != nil {
		exactCond = &TopoEventExactConditions{
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
			BkNetworkunitId: condition.ExactInclude.NetworkUnitID,
			AccesspointId:   condition.ExactInclude.AccessPointID,
			Type:            types.TopoEventTypeListToStringList(condition.ExactInclude.Type),
			Operator:        condition.ExactInclude.Operator,
		}
	}

	if condition.FuzzyInclude != nil {
		fuzzyCond = &TopoEventFuzzyConditions{
			BkNetworkareaName: condition.FuzzyInclude.NetworkAreaName,
			BkNetworkunitName: condition.FuzzyInclude.NetworkUnitName,
		}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, nil, errors.New("exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
}

// Validate check body.
func (x *PackageEventListReq) Validate() error {
	if err := validatePage(x.GetPage()); err != nil {
		return err
	}

	if err := validateTimeRange(x.GetOperateTimeRange(), maxOperateTimeRangeDuration); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageEventListReq) AutoConvert() {
	if x.GetOperateTimeRange() == nil {
		x.OperateTimeRange = &TimeRange{
			StartTimestampSec: time.Now().Add(-1 * maxOperateTimeRangeDuration).Unix(),
			EndTimestampSec:   time.Now().Unix(),
		}
	}
}

// ConvertPageToTypes convert page to types.
func (x *PackageEventListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageEventListReq) ConvertConditionsToTypes() (*types.PackageEventCondition, error) {
	return convertPackageEventConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *PackageEventListReq) ConvertConditionsFromTypes(condition *types.PackageEventCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertPackageEventConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.OperateTimeRange = timeRange
	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertPackageEventsToTypes convert package events to types.
func (x *TopoEventListResp) ConvertPackageEventsToTypes() (int64, []*types.TopoEvent) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.TopoEvent, len(items))
	for idx, item := range items {
		result[idx] = &types.TopoEvent{
			TenantID:        item.GetTenantId(),
			Type:            types.TopoEventType(item.GetType()),
			NetworkAreaID:   item.GetBkNetworkareaId(),
			NetworkAreaName: item.GetBkNetworkareaName(),
			NetworkUnitID:   item.GetBkNetworkunitId(),
			NetworkUnitName: item.GetBkNetworkunitName(),
			AccessPointID:   item.GetAccesspointId(),
			AccessPointName: item.GetAccesspointName(),
			OperateTime:     time.UnixMilli(item.GetOperateTime()),
			Operator:        item.GetOperator(),
		}
	}

	return data.GetTotal(), result
}

// ConvertPackageEventsFromTypes convert types to topo events.
func (x *PackageEventListResp) ConvertPackageEventsFromTypes(total int64, events []*types.PackageEvent) {
	items := make([]*PackageEvent, len(events))
	for idx, event := range events {
		item := newEmptyPackageEvent()
		*item.Name = event.Name
		*item.EventType = string(event.EventType)
		*item.ReleaseType = string(event.ReleaseType)
		*item.Generation = int64(event.Generation)
		*item.CpuArch = string(event.CPUArch)
		*item.OsType = string(event.OSType)
		*item.Version = string(event.Version)
		*item.OperateTime = event.OperateTime.UnixMilli()
		*item.Operator = event.Operator

		items[idx] = item
	}

	x.Data = &PackageEventListResp_Data{
		Total: total,
		Items: items,
	}
}

func convertPackageEventConditionsToTypes(
	exactCond *PackageEventExactConditions,
	fuzzyCond *PackageEventFuzzyConditions,
	timeRange *TimeRange) (*types.PackageEventCondition, error) {

	condition := &types.PackageEventCondition{}

	if timeRange != nil {
		operateTimeRange, err := convertTimeRangeToTypes(timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert operate time range: %w", err)
		}
		condition.OperateTimeRange = operateTimeRange
	}

	osTypeList, err := criteria.StringListToOSTypeList(exactCond.GetOsType())
	if err != nil {
		return nil, fmt.Errorf("failed to convert os type list: %w", err)
	}

	cpuArchList, err := criteria.StringListToCPUArchList(exactCond.GetCpuArch())
	if err != nil {
		return nil, fmt.Errorf("failed to convert cpu arch list: %w", err)
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.PackageEventExactFields{
			EventType:   types.StringListToPackageEventTypeList(exactCond.GetEventType()),
			ReleaseType: types.StringListToReleaseTypeList(exactCond.GetReleaseType()),
			Version:     exactCond.GetVersion(),
			OSType:      osTypeList,
			CPUArch:     cpuArchList,
			Generation:  types.Int64ListToGenerationList(exactCond.GetGeneration()),
			Operator:    exactCond.GetOperator(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.PackageEventFuzzyFields{}
	}

	return condition, nil
}

func convertPackageEventConditionsFromTypes(condition *types.PackageEventCondition) (
	*PackageEventExactConditions, *PackageEventFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var timeRange *TimeRange
	var exactCond *PackageEventExactConditions
	var fuzzyCond *PackageEventFuzzyConditions

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	if condition.ExactInclude != nil {
		exactCond = &PackageEventExactConditions{
			EventType:   types.PackageEventTypeListToStringList(condition.ExactInclude.EventType),
			ReleaseType: types.ReleaseTypeListToStringList(condition.ExactInclude.ReleaseType),
			Version:     condition.ExactInclude.Version,
			OsType:      criteria.OSTypeListToStringList(condition.ExactInclude.OSType),
			CpuArch:     criteria.CPUArchListToStringList(condition.ExactInclude.CPUArch),
			Generation:  types.GenerationListToInt64List(condition.ExactInclude.Generation),
			Operator:    condition.ExactInclude.Operator,
		}
	}

	if condition.FuzzyInclude != nil {
		fuzzyCond = &PackageEventFuzzyConditions{}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, nil, errors.New("exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
}

// Validate check body.
func (x *PackageEventDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageEventDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageEventDistinctReq) ConvertConditionsToTypes() (*types.PackageEventCondition, error) {
	return convertPackageEventConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertResultToTypes convert result to types.
func (x *PackageEventDistinctResp) ConvertResultToTypes() (*types.PackageEventDistinctResult, error) {
	if x.GetData() == nil {
		return &types.PackageEventDistinctResult{}, nil
	}

	data := x.GetData()

	osTypeList, err := criteria.StringListToOSTypeList(data.GetOsType())
	if err != nil {
		return nil, fmt.Errorf("failed to convert os type list: %w", err)
	}

	cpuArchList, err := criteria.StringListToCPUArchList(data.GetCpuArch())
	if err != nil {
		return nil, fmt.Errorf("failed to convert cpu arch list: %w", err)
	}

	return &types.PackageEventDistinctResult{
		ReleaseType: types.StringListToReleaseTypeList(data.GetReleaseType()),
		EventType:   types.StringListToPackageEventTypeList(data.GetEventType()),
		OSType:      osTypeList,
		CPUArch:     cpuArchList,
		Version:     data.GetVersion(),
		Operator:    data.GetOperator(),
	}, nil
}

// ConvertResultFromTypes convert result from types.
func (x *PackageEventDistinctResp) ConvertResultFromTypes(result *types.PackageEventDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PackageEventDistinctResp_Data{
		EventType:   formatRespSlice(types.PackageEventTypeListToStringList(result.EventType)),
		ReleaseType: formatRespSlice(types.ReleaseTypeListToStringList(result.ReleaseType)),
		OsType:      formatRespSlice(criteria.OSTypeListToStringList(result.OSType)),
		CpuArch:     formatRespSlice(criteria.CPUArchListToStringList(result.CPUArch)),
		Version:     formatRespSlice(result.Version),
		Operator:    formatRespSlice(result.Operator),
	}
}

func newEmptyPackageEvent() *PackageEvent {
	return &PackageEvent{
		Name:        new(string),
		EventType:   new(string),
		ReleaseType: new(string),
		Generation:  new(int64),
		OsType:      new(string),
		CpuArch:     new(string),
		Version:     new(string),
		OperateTime: new(int64),
		Operator:    new(string),
	}
}

// Validate check body.
func (x *ConfigPolicyEventListReq) Validate() error {
	if err := validatePage(x.GetPage()); err != nil {
		return err
	}

	if err := validateTimeRange(x.GetOperateTimeRange(), maxOperateTimeRangeDuration); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyEventListReq) AutoConvert() {
	if x.GetOperateTimeRange() == nil {
		x.OperateTimeRange = &TimeRange{
			StartTimestampSec: time.Now().Add(-1 * maxOperateTimeRangeDuration).Unix(),
			EndTimestampSec:   time.Now().Unix(),
		}
	}
}

// ConvertPageToTypes convert page to types.
func (x *ConfigPolicyEventListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *ConfigPolicyEventListReq) ConvertConditionsToTypes() (*types.ConfigPolicyEventCondition, error) {
	return convertConfigPolicyEventConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertConditionsFromTypes convert types to conditions.
func (x *ConfigPolicyEventListReq) ConvertConditionsFromTypes(condition *types.ConfigPolicyEventCondition) error {
	exactCond, fuzzyCond, timeRange, err := convertConfigPolicyEventConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.OperateTimeRange = timeRange
	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertConfigPolicyEventsToTypes convert policy events to types.
func (x *TopoEventListResp) ConvertConfigPolicyEventsToTypes() (int64, []*types.TopoEvent) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.TopoEvent, len(items))
	for idx, item := range items {
		result[idx] = &types.TopoEvent{
			TenantID:        item.GetTenantId(),
			Type:            types.TopoEventType(item.GetType()),
			NetworkAreaID:   item.GetBkNetworkareaId(),
			NetworkAreaName: item.GetBkNetworkareaName(),
			NetworkUnitID:   item.GetBkNetworkunitId(),
			NetworkUnitName: item.GetBkNetworkunitName(),
			AccessPointID:   item.GetAccesspointId(),
			AccessPointName: item.GetAccesspointName(),
			OperateTime:     time.UnixMilli(item.GetOperateTime()),
			Operator:        item.GetOperator(),
		}
	}

	return data.GetTotal(), result
}

// ConvertConfigPolicyEventsFromTypes convert types to topo events.
func (x *ConfigPolicyEventListResp) ConvertConfigPolicyEventsFromTypes(total int64, events []*types.ConfigPolicyEvent) {
	items := make([]*ConfigPolicyEvent, len(events))
	for idx, event := range events {
		item := newEmptyConfigPolicyEvent()
		*item.TenantId = event.TenantID
		*item.Type = string(event.Type)
		*item.ConfigpolicyType = string(event.ConfigPolicyType)
		*item.ConfigpolicyId = event.ConfigPolicyID
		*item.ConfigpolicyName = event.ConfigPolicyName
		*item.Version = event.Version
		*item.OperateTime = event.OperateTime.UnixMilli()
		*item.Operator = event.Operator

		items[idx] = item
	}

	x.Data = &ConfigPolicyEventListResp_Data{
		Total: total,
		Items: items,
	}
}

func convertConfigPolicyEventConditionsToTypes(
	exactCond *ConfigPolicyEventExactConditions,
	fuzzyCond *ConfigPolicyEventFuzzyConditions,
	timeRange *TimeRange) (*types.ConfigPolicyEventCondition, error) {

	condition := &types.ConfigPolicyEventCondition{}

	if timeRange != nil {
		operateTimeRange, err := convertTimeRangeToTypes(timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to convert operate time range: %w", err)
		}
		condition.OperateTimeRange = operateTimeRange
	}

	configPolicyTypeList, err := types.StringListToConfigPolicyTypeList(exactCond.GetConfigpolicyType())
	if err != nil {
		return nil, fmt.Errorf("failed to convert config policy type list: %w", err)
	}

	eventTypeList, err := types.StringListToConfigPolicyEventTypeList(exactCond.GetType())
	if err != nil {
		return nil, fmt.Errorf("failed to convert config policy event type list: %w", err)
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ConfigPolicyEventExactFields{
			Type:             eventTypeList,
			ConfigPolicyType: configPolicyTypeList,
			Version:          exactCond.GetVersion(),
			ConfigPolicyID:   exactCond.GetConfigpolicyId(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.ConfigPolicyEventFuzzyFields{
			ConfigPolicyName: fuzzyCond.GetConfigpolicyName(),
			Operator:         fuzzyCond.GetOperator(),
		}
	}

	return condition, nil
}

func convertConfigPolicyEventConditionsFromTypes(condition *types.ConfigPolicyEventCondition) (
	*ConfigPolicyEventExactConditions, *ConfigPolicyEventFuzzyConditions, *TimeRange, error) {

	if condition == nil {
		return nil, nil, nil, nil
	}

	var timeRange *TimeRange
	var exactCond *ConfigPolicyEventExactConditions
	var fuzzyCond *ConfigPolicyEventFuzzyConditions

	if condition.OperateTimeRange != nil {
		timeRange = &TimeRange{
			StartTimestampSec: condition.OperateTimeRange.StartTime.Unix(),
			EndTimestampSec:   condition.OperateTimeRange.EndTime.Unix(),
		}
	}

	if condition.ExactInclude != nil {
		exactCond = &ConfigPolicyEventExactConditions{
			Type:             types.ConfigPolicyEventTypeListToStringList(condition.ExactInclude.Type),
			ConfigpolicyType: types.ConfigPolicyTypeListToStringList(condition.ExactInclude.ConfigPolicyType),
			Version:          condition.ExactInclude.Version,
			ConfigpolicyId:   condition.ExactInclude.ConfigPolicyID,
		}
	}

	if condition.FuzzyInclude != nil {
		fuzzyCond = &ConfigPolicyEventFuzzyConditions{
			ConfigpolicyName: condition.FuzzyInclude.ConfigPolicyName,
			Operator:         condition.FuzzyInclude.Operator,
		}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, nil, errors.New("exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, timeRange, nil
}

// Validate check body.
func (x *ConfigPolicyEventDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ConfigPolicyEventDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *ConfigPolicyEventDistinctReq) ConvertConditionsToTypes() (*types.ConfigPolicyEventCondition, error) {
	return convertConfigPolicyEventConditionsToTypes(
		x.GetExactIncludeConditions(),
		x.GetFuzzyIncludeConditions(),
		x.GetOperateTimeRange())
}

// ConvertResultToTypes convert result to types.
func (x *ConfigPolicyEventDistinctResp) ConvertResultToTypes() (*types.ConfigPolicyEventDistinctResult, error) {
	if x.GetData() == nil {
		return &types.ConfigPolicyEventDistinctResult{}, nil
	}

	data := x.GetData()

	eventTypeList, err := types.StringListToConfigPolicyEventTypeList(data.GetType())
	if err != nil {
		return &types.ConfigPolicyEventDistinctResult{}, fmt.Errorf("failed to convert config policy event type list: %w", err)
	}

	configPolicyTypeList, err := types.StringListToConfigPolicyTypeList(data.GetConfigpolicyType())
	if err != nil {
		return &types.ConfigPolicyEventDistinctResult{}, fmt.Errorf("failed to convert config policy type list: %w", err)
	}

	return &types.ConfigPolicyEventDistinctResult{
		Type:             eventTypeList,
		ConfigPolicyID:   data.GetConfigpolicyId(),
		ConfigPolicyName: data.GetConfigpolicyName(),
		ConfigPolicyType: configPolicyTypeList,
		Version:          data.GetVersion(),
		Operator:         data.GetOperator(),
	}, nil
}

// ConvertResultFromTypes convert result from types.
func (x *ConfigPolicyEventDistinctResp) ConvertResultFromTypes(result *types.ConfigPolicyEventDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &ConfigPolicyEventDistinctResp_Data{
		Type:             formatRespSlice(types.ConfigPolicyEventTypeListToStringList(result.Type)),
		ConfigpolicyType: formatRespSlice(types.ConfigPolicyTypeListToStringList(result.ConfigPolicyType)),
		Version:          formatRespSlice(result.Version),
		Operator:         formatRespSlice(result.Operator),
		ConfigpolicyId:   formatRespSlice(result.ConfigPolicyID),
		ConfigpolicyName: formatRespSlice(result.ConfigPolicyName),
	}
}

func newEmptyConfigPolicyEvent() *ConfigPolicyEvent {
	return &ConfigPolicyEvent{
		TenantId:         new(string),
		Type:             new(string),
		ConfigpolicyType: new(string),
		ConfigpolicyId:   new(int64),
		ConfigpolicyName: new(string),
		Version:          new(int64),
		OperateTime:      new(int64),
		Operator:         new(string),
	}
}
