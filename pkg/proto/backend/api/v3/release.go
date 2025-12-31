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

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// release list max limit
	maxReleaseLimit = 1000
)

// ==================== Helper Functions ====================

func convertReleaseExactConditionsToTypes(exactCond *PackageReleaseExactConditions) *types.ReleaseExactFields {
	if exactCond == nil {
		return &types.ReleaseExactFields{}
	}

	plats := make([]platfmt.Platform, 0)
	for _, plat := range exactCond.GetPlatform() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	return &types.ReleaseExactFields{
		Platform:  plats,
		Version:   exactCond.GetVersion(),
		AsDefault: exactCond.GetAsDefault(),
		Enabled:   exactCond.GetEnabled(),
		Name:      exactCond.GetName(),
		FileName:  exactCond.GetFileName(),
	}
}

func convertReleaseConditionsFromTypes(conditions *types.ReleaseCondition) (*PackageReleaseExactConditions, error) {
	if conditions == nil {
		return &PackageReleaseExactConditions{}, nil
	}

	var exactCond *PackageReleaseExactConditions

	if conditions.ExactInclude != nil {
		exactCond = new(PackageReleaseExactConditions)
		exactCond.Platform = make([]*Platform, 0)
		for _, plat := range conditions.ExactInclude.Platform {
			exactCond.Platform = append(exactCond.Platform, ConvertPlatformFromTypes(plat))
		}
		exactCond.Version = conditions.ExactInclude.Version
		exactCond.AsDefault = conditions.ExactInclude.AsDefault
		exactCond.Enabled = conditions.ExactInclude.Enabled
		exactCond.Name = conditions.ExactInclude.Name
		exactCond.FileName = conditions.ExactInclude.FileName
	}

	if conditions.FuzzyInclude != nil || conditions.ExactExclude != nil || conditions.FuzzyExclude != nil {
		return nil, errors.New("fuzzy-include, exact-exclude, fuzzy-exclude are not supported")
	}

	return exactCond, nil
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

func newEmptyReleaseAgent() *ReleaseAgent {
	return &ReleaseAgent{
		Release:     newEmptyRelease(),
		ChangeLogEn: new(string),
		ChangeLogZh: new(string),
	}
}

func newEmptyReleaseProxy() *ReleaseProxy {
	return &ReleaseProxy{
		Release:     newEmptyRelease(),
		ChangeLogEn: new(string),
		ChangeLogZh: new(string),
	}
}

func newEmptyReleasePlugin() *ReleasePlugin {
	plugin := &ReleasePlugin{
		Release: newEmptyRelease(),
	}

	return plugin
}

// ==================== PackageReleaseAgent ====================

// Validate check body.
func (x *PackageReleaseAgentListReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentListReq) AutoConvert() {
}

// PageLimit return page limit.
func (x *PackageReleaseAgentListReq) PageLimit() int {
	return maxReleaseLimit
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseAgentListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentListReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseAgentListReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseAgentListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseAgent) {
	items := make([]*ReleaseAgent, len(releases))
	for idx, release := range releases {
		item := newEmptyReleaseAgent()
		*item.Release.Name = release.Name
		*item.Release.Generation = int64(release.Generation)
		*item.Release.ReleaseType = string(release.Type)
		*item.Release.OsType = string(release.Platform.OS)
		*item.Release.CpuArch = string(release.Platform.Arch)
		*item.Release.Version = release.Version
		*item.Release.FileName = release.FileName
		item.Release.Labels = release.Labels
		*item.Release.Enabled = release.Enabled
		*item.Release.AsDefault = release.AsDefault
		*item.Release.Md5 = release.MD5
		*item.Release.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Release.Operator = release.Operator

		*item.ChangeLogEn = release.ChangeLogEN
		*item.ChangeLogZh = release.ChangeLogZH
		items[idx] = item
	}

	x.Data = &PackageReleaseAgentListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertReleasesToTypes convert releases to types.
func (x *PackageReleaseAgentListResp) ConvertReleasesToTypes() (int64, []*types.ReleaseAgent) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.ReleaseAgent, len(items))
	for idx, item := range items {
		releaseAgent := &types.ReleaseAgent{
			Release: types.Release{
				Name:       item.GetRelease().GetName(),
				Generation: types.Generation(item.GetRelease().GetGeneration()),
				Type:       types.ReleaseType(item.GetRelease().GetReleaseType()),
				Version:    item.GetRelease().GetVersion(),
				Platform: platfmt.Platform{
					OS:   criteria.OSType(item.GetRelease().GetOsType()),
					Arch: criteria.CPUArch(item.GetRelease().GetCpuArch()),
				},
				Labels:    item.GetRelease().GetLabels(),
				FileName:  item.GetRelease().GetFileName(),
				MD5:       item.GetRelease().GetMd5(),
				Enabled:   item.GetRelease().GetEnabled(),
				AsDefault: item.GetRelease().GetAsDefault(),
				UpdatedAt: time.UnixMilli(int64(item.GetRelease().GetUpdatedAt())).Local(),
				Operator:  item.GetRelease().GetOperator(),
			},
			ReleaseAdditionInfoAgent: types.ReleaseAdditionInfoAgent{
				ChangeLogEN: item.GetChangeLogEn(),
				ChangeLogZH: item.GetChangeLogZh(),
			},
		}
		result[idx] = releaseAgent
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *PackageReleaseAgentDistinctReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDistinctReq) AutoConvert() {
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentDistinctReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseAgentDistinctReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertResultFromTypes converts the result from types.
func (x *PackageReleaseAgentDistinctResp) ConvertResultFromTypes(result *types.ReleaseDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PackageReleaseDistinctData{
		OsType:  formatRespSlice(result.OSType),
		CpuArch: formatRespSlice(result.CPUArch),
	}
}

// ConvertResultToTypes converts the response to types.
func (x *PackageReleaseAgentDistinctResp) ConvertResultToTypes() *types.ReleaseDistinctResult {
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
func (x *PackageReleaseAgentSetLabelsManyReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentSetLabelsManyReq) AutoConvert() {
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentSetLabelsManyReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseAgentSetLabelsManyReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// Validate check body.
func (x *PackageReleaseAgentEnableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentEnableReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentDisableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDisableReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentSetAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentSetAsDefaultReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentCancelAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentCancelAsDefaultReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDeleteReq) AutoConvert() {
}

// ==================== PackageReleaseProxy ====================

// Validate check body.
func (x *PackageReleaseProxyListReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyListReq) AutoConvert() {
}

// PageLimit return page limit.
func (x *PackageReleaseProxyListReq) PageLimit() int {
	return maxReleaseLimit
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseProxyListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyListReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseProxyListReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseProxyListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseProxy) {
	items := make([]*ReleaseProxy, len(releases))
	for idx, release := range releases {
		item := newEmptyReleaseProxy()
		*item.Release.Name = release.Name
		*item.Release.Generation = int64(release.Generation)
		*item.Release.ReleaseType = string(release.Type)
		*item.Release.OsType = string(release.Platform.OS)
		*item.Release.CpuArch = string(release.Platform.Arch)
		*item.Release.Version = release.Version
		*item.Release.FileName = release.FileName
		item.Release.Labels = release.Labels
		*item.Release.Enabled = release.Enabled
		*item.Release.AsDefault = release.AsDefault
		*item.Release.Md5 = release.MD5
		*item.Release.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Release.Operator = release.Operator

		*item.ChangeLogEn = release.ChangeLogEN
		*item.ChangeLogZh = release.ChangeLogZH
		items[idx] = item
	}

	x.Data = &PackageReleaseProxyListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertReleasesToTypes convert releases to types.
func (x *PackageReleaseProxyListResp) ConvertReleasesToTypes() (int64, []*types.ReleaseProxy) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.ReleaseProxy, len(items))
	for idx, item := range items {
		releaseProxy := &types.ReleaseProxy{
			Release: types.Release{
				Name:       item.GetRelease().GetName(),
				Generation: types.Generation(item.GetRelease().GetGeneration()),
				Type:       types.ReleaseType(item.GetRelease().GetReleaseType()),
				Version:    item.GetRelease().GetVersion(),
				Platform: platfmt.Platform{
					OS:   criteria.OSType(item.GetRelease().GetOsType()),
					Arch: criteria.CPUArch(item.GetRelease().GetCpuArch()),
				},
				Labels:    item.GetRelease().GetLabels(),
				FileName:  item.GetRelease().GetFileName(),
				MD5:       item.GetRelease().GetMd5(),
				Enabled:   item.GetRelease().GetEnabled(),
				AsDefault: item.GetRelease().GetAsDefault(),
				UpdatedAt: time.UnixMilli(int64(item.GetRelease().GetUpdatedAt())).Local(),
				Operator:  item.GetRelease().GetOperator(),
			},
			ReleaseAdditionInfoProxy: types.ReleaseAdditionInfoProxy{
				ChangeLogEN: item.GetChangeLogEn(),
				ChangeLogZH: item.GetChangeLogZh(),
			},
		}
		result[idx] = releaseProxy
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *PackageReleaseProxyDistinctReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDistinctReq) AutoConvert() {
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyDistinctReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseProxyDistinctReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertResultFromTypes converts the result from types.
func (x *PackageReleaseProxyDistinctResp) ConvertResultFromTypes(result *types.ReleaseDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &PackageReleaseDistinctData{
		OsType:  formatRespSlice(result.OSType),
		CpuArch: formatRespSlice(result.CPUArch),
	}
}

// ConvertResultToTypes converts the response to types.
func (x *PackageReleaseProxyDistinctResp) ConvertResultToTypes() *types.ReleaseDistinctResult {
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
func (x *PackageReleaseProxySetLabelsManyReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxySetLabelsManyReq) AutoConvert() {
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxySetLabelsManyReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseProxySetLabelsManyReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// Validate check body.
func (x *PackageReleaseProxyEnableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyEnableReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyDisableReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDisableReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxySetAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxySetAsDefaultReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyCancelAsDefaultReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyCancelAsDefaultReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDeleteReq) AutoConvert() {
}

// ==================== PackageReleasePlugin ====================

// Validate check body.
func (x *PackageReleasePluginListReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginListReq) AutoConvert() {
}

// PageLimit return page limit.
func (x *PackageReleasePluginListReq) PageLimit() int {
	return maxReleaseLimit
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleasePluginListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage(), x.PageLimit())
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleasePluginListReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleasePluginListReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertReleasePluginsFromTypes convert releases from types.
func (x *PackageReleasePluginListResp) ConvertReleasePluginsFromTypes(total int64, releasePlugins []*types.ReleasePlugin) {
	items := make([]*ReleasePlugin, len(releasePlugins))
	for idx, release := range releasePlugins {
		item := newEmptyReleasePlugin()
		*item.Release.Name = release.Name
		*item.Release.Generation = int64(release.Generation)
		*item.Release.ReleaseType = string(release.Type)
		*item.Release.OsType = string(release.Platform.OS)
		*item.Release.CpuArch = string(release.Platform.Arch)
		*item.Release.Version = release.Version
		*item.Release.FileName = release.FileName
		item.Release.Labels = release.Labels
		*item.Release.Enabled = release.Enabled
		*item.Release.AsDefault = release.AsDefault
		*item.Release.Md5 = release.MD5
		*item.Release.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Release.Operator = release.Operator

		items[idx] = item
	}

	x.Data = &PackageReleasePluginListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertReleasePluginsToTypes convert releases to types.
func (x *PackageReleasePluginListResp) ConvertReleasePluginsToTypes() (int64, []*types.ReleasePlugin) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.ReleasePlugin, len(items))
	for idx, item := range items {
		releasePlugin := &types.ReleasePlugin{
			Release: types.Release{
				Name:       item.GetRelease().GetName(),
				Generation: types.Generation(item.GetRelease().GetGeneration()),
				Type:       types.ReleaseType(item.GetRelease().GetReleaseType()),
				Version:    item.GetRelease().GetVersion(),
				Platform: platfmt.Platform{
					OS:   criteria.OSType(item.GetRelease().GetOsType()),
					Arch: criteria.CPUArch(item.GetRelease().GetCpuArch()),
				},
				Labels:    item.GetRelease().GetLabels(),
				FileName:  item.GetRelease().GetFileName(),
				MD5:       item.GetRelease().GetMd5(),
				Enabled:   item.GetRelease().GetEnabled(),
				AsDefault: item.GetRelease().GetAsDefault(),
				UpdatedAt: time.UnixMilli(int64(item.GetRelease().GetUpdatedAt())).Local(),
				Operator:  item.GetRelease().GetOperator(),
			},
		}
		result[idx] = releasePlugin
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *PackageReleasePluginEnableReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginEnableReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleasePluginEnableReq) GetIdentifier() (string, types.Generation, platfmt.Platform, string) {
	return x.GetName(), types.Generation(x.GetGeneration()), ConvertPlatformToTypes(x.GetPlatform()), x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleasePluginEnableReq) SetIdentifer(name string, gen types.Generation, plat platfmt.Platform, ver string) {
	x.Name = name
	x.Generation = int64(gen)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleasePluginDisableReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginDisableReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleasePluginDisableReq) GetIdentifier() (string, types.Generation, platfmt.Platform, string) {
	return x.GetName(), types.Generation(x.GetGeneration()), ConvertPlatformToTypes(x.GetPlatform()), x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleasePluginDisableReq) SetIdentifer(name string, gen types.Generation, plat platfmt.Platform, ver string) {
	x.Name = name
	x.Generation = int64(gen)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleasePluginSetAsDefaultReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginSetAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleasePluginSetAsDefaultReq) GetIdentifier() (string, types.Generation, platfmt.Platform, string) {
	return x.GetName(), types.Generation(x.GetGeneration()), ConvertPlatformToTypes(x.GetPlatform()), x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleasePluginSetAsDefaultReq) SetIdentifer(name string, gen types.Generation, plat platfmt.Platform, ver string) {
	x.Name = name
	x.Generation = int64(gen)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleasePluginCancelAsDefaultReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginCancelAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleasePluginCancelAsDefaultReq) GetIdentifier() (string, types.Generation, platfmt.Platform, string) {
	return x.GetName(), types.Generation(x.GetGeneration()), ConvertPlatformToTypes(x.GetPlatform()), x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleasePluginCancelAsDefaultReq) SetIdentifer(name string, gen types.Generation, plat platfmt.Platform, ver string) {
	x.Name = name
	x.Generation = int64(gen)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// Validate check body.
func (x *PackageReleasePluginDeleteReq) Validate() error {
	if !ConvertPlatformToTypes(x.GetPlatform()).Validate() {
		return fmt.Errorf("failed to validate platform, plat(%+v)", x.GetPlatform())
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginDeleteReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleasePluginDeleteReq) GetIdentifier() (string, types.Generation, platfmt.Platform, string) {
	return x.GetName(), types.Generation(x.GetGeneration()), ConvertPlatformToTypes(x.GetPlatform()), x.GetVersion()
}

// SetIdentifer set identifier.
func (x *PackageReleasePluginDeleteReq) SetIdentifer(name string, gen types.Generation, plat platfmt.Platform, ver string) {
	x.Name = name
	x.Generation = int64(gen)
	x.Platform = ConvertPlatformFromTypes(plat)
	x.Version = ver
}

// ==================== PackageReleaseCert ====================

// Validate check body.
func (x *PackageReleaseCertListReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCertListReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseCertDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCertDeleteReq) AutoConvert() {
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseCertListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseCert) {
	items := make([]*ReleaseCert, len(releases))
	for idx, release := range releases {
		item := &ReleaseCert{
			Release: newEmptyRelease(),
		}
		*item.Release.Name = release.Name
		*item.Release.Generation = int64(release.Generation)
		*item.Release.ReleaseType = string(release.Type)
		*item.Release.OsType = string(release.Platform.OS)
		*item.Release.CpuArch = string(release.Platform.Arch)
		*item.Release.Version = release.Version
		*item.Release.FileName = release.FileName
		item.Release.Labels = release.Labels
		*item.Release.Enabled = release.Enabled
		*item.Release.AsDefault = release.AsDefault
		*item.Release.Md5 = release.MD5
		*item.Release.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Release.Operator = release.Operator
		items[idx] = item
	}

	x.Data = &PackageReleaseCertListResp_Data{
		Total: total,
		Items: items,
	}
}

// ==================== PackageReleaseBinTool ====================

// Validate check body.
func (x *PackageReleaseBinToolListReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseBinToolListReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseBinToolDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseBinToolDeleteReq) AutoConvert() {
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseBinToolListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseBinTool) {
	items := make([]*ReleaseBinTool, len(releases))
	for idx, release := range releases {
		item := &ReleaseBinTool{
			Release: newEmptyRelease(),
		}
		*item.Release.Name = release.Name
		*item.Release.Generation = int64(release.Generation)
		*item.Release.ReleaseType = string(release.Type)
		*item.Release.OsType = string(release.Platform.OS)
		*item.Release.CpuArch = string(release.Platform.Arch)
		*item.Release.Version = release.Version
		*item.Release.FileName = release.FileName
		item.Release.Labels = release.Labels
		*item.Release.Enabled = release.Enabled
		*item.Release.AsDefault = release.AsDefault
		*item.Release.Md5 = release.MD5
		*item.Release.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Release.Operator = release.Operator
		items[idx] = item
	}

	x.Data = &PackageReleaseBinToolListResp_Data{
		Total: total,
		Items: items,
	}
}

// ==================== PackageReleasePluginBinTool ====================

// Validate check body.
func (x *PackageReleasePluginBinToolListReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginBinToolListReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleasePluginBinToolDeleteReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginBinToolDeleteReq) AutoConvert() {
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleasePluginBinToolListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleasePluginBinTool) {
	items := make([]*ReleasePluginBinTool, len(releases))
	for idx, release := range releases {
		item := &ReleasePluginBinTool{
			Release: newEmptyRelease(),
		}
		*item.Release.Name = release.Name
		*item.Release.Generation = int64(release.Generation)
		*item.Release.ReleaseType = string(release.Type)
		*item.Release.OsType = string(release.Platform.OS)
		*item.Release.CpuArch = string(release.Platform.Arch)
		*item.Release.Version = release.Version
		*item.Release.FileName = release.FileName
		item.Release.Labels = release.Labels
		*item.Release.Enabled = release.Enabled
		*item.Release.AsDefault = release.AsDefault
		*item.Release.Md5 = release.MD5
		*item.Release.UpdatedAt = uint64(release.UpdatedAt.UnixMilli())
		*item.Release.Operator = release.Operator
		items[idx] = item
	}

	x.Data = &PackageReleasePluginBinToolListResp_Data{
		Total: total,
		Items: items,
	}
}
