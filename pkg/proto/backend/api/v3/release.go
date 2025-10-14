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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *PackageReleaseListReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseListReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseListReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

func convertReleaseExactConditionsToTypes(exactCond *PackageReleaseExactConditions) *types.ReleaseExactFields {
	if exactCond == nil {
		return &types.ReleaseExactFields{}
	}

	plats := make([]platform.Platform, 0)
	for _, plat := range exactCond.GetPlatform() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	return &types.ReleaseExactFields{
		Platform:  plats,
		Version:   exactCond.GetVersion(),
		AsDefault: exactCond.GetAsDefault(),
		Enabled:   exactCond.GetEnabled(),
	}
}

func convertReleaseConditionsFromTypes(conditions *types.ReleaseCondition) (*PackageReleaseExactConditions, error) {
	if conditions == nil {
		return nil, nil
	}

	var exactCond *PackageReleaseExactConditions

	if conditions.ExactInclude != nil {
		exactCond = new(PackageReleaseExactConditions)
		exactCond.Platform = make([]*Platform, 0)
		for _, plat := range conditions.ExactInclude.Platform {
			exactCond.Platform = append(exactCond.Platform, ConvertPlatformFromTypes(plat))
		}
		exactCond.Version = conditions.ExactInclude.Version
	}

	if conditions.FuzzyInclude != nil || conditions.ExactExclude != nil || conditions.FuzzyExclude != nil {
		return nil, errors.New("fuzzy-include, exact-exclude, fuzzy-exclude are not supported")
	}

	return exactCond, nil
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseListResp) ConvertReleasesFromTypes(total int64, releases []*types.Release) {
	items := make([]*Release, len(releases))
	for idx, release := range releases {
		item := newEmptyRelease()
		*item.Name = release.Name
		*item.Generation = int64(release.Generation)
		*item.ReleaseType = string(release.Type)
		*item.OsType = string(release.Platform.OS)
		*item.CpuArch = string(release.Platform.Arch)
		*item.Version = release.Version
		*item.FileName = release.FileName
		item.Labels = release.Labels
		*item.Enabled = release.Enabled
		*item.AsDefault = release.AsDefault
		*item.Md5 = release.MD5
		*item.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Operator = release.Operator

		items[idx] = item
	}

	x.Data = &PackageReleaseListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertReleasesToTypes convert releases to types.
func (x *PackageReleaseListResp) ConvertReleasesToTypes() (int64, []*types.Release) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.Release, len(items))
	for idx, item := range items {
		release := &types.Release{
			Name:       item.GetName(),
			Generation: types.Generation(item.GetGeneration()),
			Type:       types.ReleaseType(item.GetReleaseType()),
			Version:    item.GetVersion(),
			Platform: platform.Platform{
				OS:   criteria.OSType(item.GetOsType()),
				Arch: criteria.CPUArch(item.GetCpuArch()),
			},
			Labels:    item.GetLabels(),
			FileName:  item.GetFileName(),
			MD5:       item.GetMd5(),
			Enabled:   item.GetEnabled(),
			AsDefault: item.GetAsDefault(),
			UpdatedAt: time.UnixMilli(int64(item.GetUpdatedAt())).Local(),
			Operator:  item.GetOperator(),
		}

		result[idx] = release
	}

	return data.GetTotal(), result
}

// Validate validates the request.
func (x *PackageReleaseDistinctReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert automatically converts the request to types.
func (x *PackageReleaseDistinctReq) AutoConvert() {
}

// ConvertExactIncludeConditionsToTypes converts the request to types.
func (x *PackageReleaseDistinctReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes converts the request to types.
func (x *PackageReleaseDistinctReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertResultFromTypes converts the result from types.
func (x *PackageReleaseDistinctResp) ConvertResultFromTypes(result *types.ReleaseDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PackageReleaseDistinctResp_Data{
		OsType:  formatRespSlice(result.OSType),
		CpuArch: formatRespSlice(result.CPUArch),
	}
}

// ConvertResultToTypes converts the response to types.
func (x *PackageReleaseDistinctResp) ConvertResultToTypes() *types.ReleaseDistinctResult {
	if x.GetData() == nil {
		return &types.ReleaseDistinctResult{}
	}

	data := x.GetData()

	return &types.ReleaseDistinctResult{
		OSType:  data.GetOsType(),
		CPUArch: data.GetCpuArch(),
	}
}

// Validate check body.
func (x *PackageReleaseSetLabelsReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseSetLabelsReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseSetLabelsReq) GetIdentifier() (types.Generation, types.ReleaseType, platform.Platform, string) {
	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleaseSetLabelsReq) SetIdentifer(
	gen types.Generation, rt types.ReleaseType, plat platform.Platform, ver string) {

	x.Generation = int64(gen)
	x.ReleaseType = string(rt)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleaseEnableReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseEnableReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseEnableReq) GetIdentifier() (types.Generation, types.ReleaseType, platform.Platform, string) {
	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleaseEnableReq) SetIdentifer(
	gen types.Generation, rt types.ReleaseType, plat platform.Platform, ver string) {

	x.Generation = int64(gen)
	x.ReleaseType = string(rt)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleaseDisableReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseDisableReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseDisableReq) GetIdentifier() (types.Generation, types.ReleaseType, platform.Platform, string) {
	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleaseDisableReq) SetIdentifer(
	gen types.Generation, rt types.ReleaseType, plat platform.Platform, ver string) {

	x.Generation = int64(gen)
	x.ReleaseType = string(rt)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleaseSetAsDefaultReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseSetAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseSetAsDefaultReq) GetIdentifier() (
	types.Generation, types.ReleaseType, platform.Platform, string) {

	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleaseSetAsDefaultReq) SetIdentifer(
	gen types.Generation, rt types.ReleaseType, plat platform.Platform, ver string) {

	x.Generation = int64(gen)
	x.ReleaseType = string(rt)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleaseCancelAsDefaultReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCancelAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseCancelAsDefaultReq) GetIdentifier() (
	types.Generation, types.ReleaseType, platform.Platform, string) {

	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleaseCancelAsDefaultReq) SetIdentifer(
	gen types.Generation, rt types.ReleaseType, plat platform.Platform, ver string) {

	x.Generation = int64(gen)
	x.ReleaseType = string(rt)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleaseDeleteReq) Validate() error {
	if err := types.ReleaseType(x.GetReleaseType()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseDeleteReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseDeleteReq) GetIdentifier() (
	types.Generation, types.ReleaseType, platform.Platform, string) {

	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleaseDeleteReq) SetIdentifer(
	gen types.Generation, rt types.ReleaseType, plat platform.Platform, ver string) {

	x.Generation = int64(gen)
	x.ReleaseType = string(rt)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

func newEmptyRelease() *Release {
	return &Release{
		Name:        new(string),
		Generation:  new(int64),
		ReleaseType: new(string),
		OsType:      new(string),
		CpuArch:     new(string),
		Version:     new(string),
		FileName:    new(string),
		Labels:      make([]string, 0),
		Enabled:     new(bool),
		AsDefault:   new(bool),
		Md5:         new(string),
		UpdatedAt:   new(uint64),
		Operator:    new(string),
	}
}
