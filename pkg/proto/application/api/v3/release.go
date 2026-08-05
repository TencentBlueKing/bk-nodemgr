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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"google.golang.org/protobuf/types/known/structpb"
)

// ==================== Helper Functions ====================

func convertReleaseConditionsToTypes(exactCond *PackageReleaseExactConditions) *types.ReleaseCondition {
	condition := types.ReleaseCondition{}

	plats := make([]platfmt.Platform, 0)
	for _, plat := range exactCond.GetPlatform() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ReleaseExactFields{
			Platform:  plats,
			Version:   exactCond.GetVersion(),
			AsDefault: exactCond.GetAsDefault(),
			Enabled:   exactCond.GetEnabled(),
			IsVisible: exactCond.GetIsVisible(),
			Name:      exactCond.GetName(),
			FileName:  exactCond.GetFileName(),
		}
	}

	return &condition
}

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
		IsVisible: exactCond.GetIsVisible(),
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
		exactCond.AsDefault = conditions.ExactInclude.AsDefault
		exactCond.Enabled = conditions.ExactInclude.Enabled
		exactCond.IsVisible = conditions.ExactInclude.IsVisible
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
		IsVisible:   new(bool),
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

func newEmptyReleaseCert() *ReleaseCert {
	return &ReleaseCert{
		Release: newEmptyRelease(),
	}
}

func newEmptyReleaseBinTool() *ReleaseBinTool {
	return &ReleaseBinTool{
		Release: newEmptyRelease(),
	}
}

func newEmptyReleasePluginBinTool() *ReleasePluginBinTool {
	return &ReleasePluginBinTool{
		Release: newEmptyRelease(),
	}
}

// ==================== PackageReleaseAgent ====================

// Validate check body.
func (x *PackageReleaseAgentListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseAgentListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentListReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
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
		*item.Release.IsVisible = release.IsVisible
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
				IsVisible: item.GetRelease().GetIsVisible(),
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
func (x *PackageReleaseAgentListBriefReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentListBriefReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseAgentListBriefReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentListBriefReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseAgentListBriefReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseAgentListBriefResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseAgent) {
	items := make([]*ReleaseAgentBrief, len(releases))
	for idx, release := range releases {
		items[idx] = convertReleaseAgentBriefFromTypes(&release.Release)
	}

	x.Data = &PackageReleaseAgentListBriefResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertReleasesToTypes convert releases to types.
func (x *PackageReleaseAgentListBriefResp) ConvertReleasesToTypes() (int64, []*types.ReleaseAgent) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.ReleaseAgent, len(items))
	for idx, item := range items {
		result[idx] = &types.ReleaseAgent{Release: *convertReleaseAgentBriefToTypes(item)}
	}

	return data.GetTotal(), result
}

func convertReleaseAgentBriefFromTypes(release *types.Release) *ReleaseAgentBrief {
	data := &ReleaseAgentBrief{}
	data.Generation = new(int64)
	data.OsType = new(string)
	data.CpuArch = new(string)
	data.Version = new(string)
	data.Enabled = new(bool)
	data.IsVisible = new(bool)
	data.AsDefault = new(bool)

	*data.Generation = int64(release.Generation)
	*data.OsType = string(release.Platform.OS)
	*data.CpuArch = string(release.Platform.Arch)
	*data.Version = release.Version
	*data.Enabled = release.Enabled
	*data.IsVisible = release.IsVisible
	*data.AsDefault = release.AsDefault

	return data
}

func convertReleaseAgentBriefToTypes(brief *ReleaseAgentBrief) *types.Release {
	return &types.Release{
		Generation: types.Generation(brief.GetGeneration()),
		Platform: platfmt.Platform{
			OS:   criteria.OSType(brief.GetOsType()),
			Arch: criteria.CPUArch(brief.GetCpuArch()),
		},
		Version:   brief.GetVersion(),
		Enabled:   brief.GetEnabled(),
		IsVisible: brief.GetIsVisible(),
		AsDefault: brief.GetAsDefault(),
	}
}

func convertReleaseProxyBriefFromTypes(release *types.Release, changeLogEN, changeLogZH string) *ReleaseProxyBrief {
	data := &ReleaseProxyBrief{}
	data.Generation = new(int64)
	data.OsType = new(string)
	data.CpuArch = new(string)
	data.Version = new(string)
	data.Enabled = new(bool)
	data.IsVisible = new(bool)
	data.AsDefault = new(bool)
	data.ChangeLogEn = new(string)
	data.ChangeLogZh = new(string)

	*data.Generation = int64(release.Generation)
	*data.OsType = string(release.Platform.OS)
	*data.CpuArch = string(release.Platform.Arch)
	*data.Version = release.Version
	*data.Enabled = release.Enabled
	*data.IsVisible = release.IsVisible
	*data.AsDefault = release.AsDefault
	*data.ChangeLogEn = changeLogEN
	*data.ChangeLogZh = changeLogZH

	return data
}

func convertReleasePluginBriefFromTypes(release *types.Release) *ReleasePluginBrief {
	data := &ReleasePluginBrief{}
	data.Generation = new(int64)
	data.OsType = new(string)
	data.CpuArch = new(string)
	data.Version = new(string)
	data.Enabled = new(bool)
	data.IsVisible = new(bool)
	data.AsDefault = new(bool)

	*data.Generation = int64(release.Generation)
	*data.OsType = string(release.Platform.OS)
	*data.CpuArch = string(release.Platform.Arch)
	*data.Version = release.Version
	*data.Enabled = release.Enabled
	*data.IsVisible = release.IsVisible
	*data.AsDefault = release.AsDefault

	return data
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

// ConvertDistinctFieldToTypes convert distinct field to types.
func (x *PackageReleaseAgentDistinctReq) ConvertDistinctFieldToTypes() types.ReleaseDistinctField {
	field := x.GetDistinctField()
	if field == nil {
		return types.ReleaseDistinctField{}
	}
	return types.ReleaseDistinctField{
		OSType:  field.GetOsType(),
		CPUArch: field.GetCpuArch(),
	}
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentDistinctReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertResultFromTypes convert result from types.
func (x *PackageReleaseAgentDistinctResp) ConvertResultFromTypes(result *types.ReleaseDistinctResult) {
	if result == nil {
		return
	}
	x.Data = &PackageReleaseDistinctData{
		OsType:  result.OSType,
		CpuArch: result.CPUArch,
	}
}

// Validate check body.
func (x *PackageReleaseAgentSetLabelsManyReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentSetLabelsManyReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentSetLabelsManyReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
}

// Validate check body.
func (x *PackageReleaseAgentSetLabelsManyResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentSetLabelsManyResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentEnableReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentEnableReq) AutoConvert() {
}

// GetIdentifier get identifier for agent release (generation, platform, version).
func (x *PackageReleaseAgentEnableReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseAgentDisableReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDisableReq) AutoConvert() {
}

// GetIdentifier get identifier for agent release (generation, platform, version).
func (x *PackageReleaseAgentDisableReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseAgentVisibleReq) Validate() error { return nil }

// AutoConvert auto convert.
func (x *PackageReleaseAgentVisibleReq) AutoConvert() {}

// Validate check body.
func (x *PackageReleaseAgentUnvisibleReq) Validate() error { return nil }

// AutoConvert auto convert.
func (x *PackageReleaseAgentUnvisibleReq) AutoConvert() {}

// Validate check body.
func (x *PackageReleaseAgentSetAsDefaultReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentSetAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier for agent release (generation, platform, version).
func (x *PackageReleaseAgentSetAsDefaultReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseAgentCancelAsDefaultReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentCancelAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier for agent release (generation, platform, version).
func (x *PackageReleaseAgentCancelAsDefaultReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseAgentCancelAsDefaultResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentCancelAsDefaultResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentDeleteReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDeleteReq) AutoConvert() {
}

// GetIdentifier get identifier for agent release (generation, platform, version).
func (x *PackageReleaseAgentDeleteReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseAgentDeleteResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDeleteResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseAgentCountDeployedReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentCountDeployedReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseAgentCountDeployedReq) ConvertConditionsToTypes() ([]*types.HostCondition, error) {
	items := x.GetItems()
	if len(items) == 0 {
		return make([]*types.HostCondition, 0), nil
	}

	conditions := make([]*types.HostCondition, len(items))
	for idx, item := range items {
		if item.GetPlatform() == nil {
			return nil, errors.New("got invalid platform")
		}

		conditions[idx] = &types.HostCondition{
			DynamicExactInclude: &types.HostDynamicExactFields{
				NodeRole:       []types.NodeRole{types.NodeRoleAgent},
				NodeGeneration: []int64{item.GetGeneration()},
				OSType:         []string{item.GetPlatform().GetOsType()},
				Arch:           []string{item.GetPlatform().GetCpuArch()},
				NodeVersion:    []string{item.GetVersion()},
			},
		}
	}

	return conditions, nil
}

// ConvertResultFromTypes convert result from types.
func (x *PackageReleaseAgentCountDeployedResp) ConvertResultFromTypes(results []int64) {
	x.Data = &PackageReleaseAgentCountDeployedResp_Data{
		Counts: results,
	}
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

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseProxyListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyListReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyListReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
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
		*item.Release.IsVisible = release.IsVisible
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
				IsVisible: item.GetRelease().GetIsVisible(),
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
func (x *PackageReleaseProxyListBriefReq) Validate() error {
	if err := types.Generation(x.GetGeneration()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyListBriefReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *PackageReleaseProxyListBriefReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertExactIncludeConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyListBriefReq) ConvertExactIncludeConditionsToTypes() *types.ReleaseExactFields {
	return convertReleaseExactConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyListBriefReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertConditionsFromTypes convert conditions from types.
func (x *PackageReleaseProxyListBriefReq) ConvertConditionsFromTypes(condition *types.ReleaseCondition) error {
	exactCond, err := convertReleaseConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond

	return nil
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseProxyListBriefResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseProxy) {
	items := make([]*ReleaseProxyBrief, len(releases))
	for idx, release := range releases {
		items[idx] = convertReleaseProxyBriefFromTypes(&release.Release, release.ChangeLogEN, release.ChangeLogZH)
	}

	x.Data = &PackageReleaseProxyListBriefResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertReleasesToTypes convert releases to types.
func (x *PackageReleaseProxyListBriefResp) ConvertReleasesToTypes() (int64, []*types.ReleaseProxy) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.ReleaseProxy, len(items))
	for idx, item := range items {
		releaseProxy := &types.ReleaseProxy{
			Release: types.Release{
				Generation: types.Generation(item.GetGeneration()),
				Platform: platfmt.Platform{
					OS:   criteria.OSType(item.GetOsType()),
					Arch: criteria.CPUArch(item.GetCpuArch()),
				},
				Version:   item.GetVersion(),
				Enabled:   item.GetEnabled(),
				IsVisible: item.GetIsVisible(),
				AsDefault: item.GetAsDefault(),
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
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDistinctReq) AutoConvert() {
}

// ConvertDistinctFieldToTypes convert distinct field to types.
func (x *PackageReleaseProxyDistinctReq) ConvertDistinctFieldToTypes() types.ReleaseDistinctField {
	field := x.GetDistinctField()
	if field == nil {
		return types.ReleaseDistinctField{}
	}
	return types.ReleaseDistinctField{
		OSType:  field.GetOsType(),
		CPUArch: field.GetCpuArch(),
	}
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyDistinctReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
}

// ConvertResultFromTypes convert result from types.
func (x *PackageReleaseProxyDistinctResp) ConvertResultFromTypes(result *types.ReleaseDistinctResult) {
	if result == nil {
		return
	}
	x.Data = &PackageReleaseDistinctData{
		OsType:  result.OSType,
		CpuArch: result.CPUArch,
	}
}

// Validate check body.
func (x *PackageReleaseProxySetLabelsManyReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxySetLabelsManyReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxySetLabelsManyReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
}

// Validate check body.
func (x *PackageReleaseProxySetLabelsManyResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxySetLabelsManyResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyEnableReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyEnableReq) AutoConvert() {
}

// GetIdentifier get identifier for proxy release (generation, platform, version).
func (x *PackageReleaseProxyEnableReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseProxyEnableResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyEnableResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyDisableReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDisableReq) AutoConvert() {
}

// GetIdentifier get identifier for proxy release (generation, platform, version).
func (x *PackageReleaseProxyDisableReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseProxyVisibleReq) Validate() error { return nil }

// AutoConvert auto convert.
func (x *PackageReleaseProxyVisibleReq) AutoConvert() {}

// Validate check body.
func (x *PackageReleaseProxyUnvisibleReq) Validate() error { return nil }

// AutoConvert auto convert.
func (x *PackageReleaseProxyUnvisibleReq) AutoConvert() {}

// Validate check body.
func (x *PackageReleaseProxyDisableResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDisableResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxySetAsDefaultReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxySetAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier for proxy release (generation, platform, version).
func (x *PackageReleaseProxySetAsDefaultReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseProxySetAsDefaultResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxySetAsDefaultResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyCancelAsDefaultReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyCancelAsDefaultReq) AutoConvert() {
}

// GetIdentifier get identifier for proxy release (generation, platform, version).
func (x *PackageReleaseProxyCancelAsDefaultReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseProxyCancelAsDefaultResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyCancelAsDefaultResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyDeleteReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDeleteReq) AutoConvert() {
}

// GetIdentifier get identifier for proxy release (generation, platform, version).
func (x *PackageReleaseProxyDeleteReq) GetIdentifier() (types.Generation, platfmt.Platform, string) {
	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check body.
func (x *PackageReleaseProxyDeleteResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDeleteResp) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseProxyCountDeployedReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyCountDeployedReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleaseProxyCountDeployedReq) ConvertConditionsToTypes() ([]*types.HostCondition, error) {
	items := x.GetItems()
	if len(items) == 0 {
		return make([]*types.HostCondition, 0), nil
	}

	conditions := make([]*types.HostCondition, len(items))
	for idx, item := range items {
		if item.GetPlatform() == nil {
			return nil, errors.New("got invalid platform")
		}

		conditions[idx] = &types.HostCondition{
			DynamicExactInclude: &types.HostDynamicExactFields{
				NodeRole:       []types.NodeRole{types.NodeRoleProxy},
				NodeGeneration: []int64{item.GetGeneration()},
				OSType:         []string{item.GetPlatform().GetOsType()},
				Arch:           []string{item.GetPlatform().GetCpuArch()},
				NodeVersion:    []string{item.GetVersion()},
			},
		}
	}

	return conditions, nil
}

// ConvertResultFromTypes convert result from types.
func (x *PackageReleaseProxyCountDeployedResp) ConvertResultFromTypes(results []int64) {
	x.Data = &PackageReleaseProxyCountDeployedResp_Data{
		Counts: results,
	}
}

// ==================== PackageReleasePlugin ====================

// Validate check body.
func (x *PackageReleasePluginGetConfigVariablesReq) Validate() error {
	if len(x.GetPlatforms()) == 0 {
		return fmt.Errorf("platforms cannot be empty")
	}

	for _, plat := range x.GetPlatforms() {
		if !ConvertPlatformToTypes(plat).Validate() {
			return fmt.Errorf("failed to validate platform, plat(%+v)", plat)
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginGetConfigVariablesReq) AutoConvert() {
}

// GetIdentifier gets identifier for plugin release.
func (x *PackageReleasePluginGetConfigVariablesReq) GetIdentifier() (string, string, types.Generation, []platfmt.Platform) {
	plats := conv.SliceToSlice(x.GetPlatforms(), ConvertPlatformToTypes)

	return x.GetName(), x.GetVersion(), types.Generation(x.GetGeneration()), plats
}

// ConvertConfigVariablesFromTypes converts config variables from types.
func (x *PackageReleasePluginGetConfigVariablesResp) ConvertConfigVariablesFromTypes(configTemplates map[string][]*types.PluginPkgConfigTemplate) {
	if len(configTemplates) == 0 {
		return
	}

	variables := make(map[string]*PackageReleasePluginGetConfigVariablesResp_ConfigVariablesList, len(configTemplates))
	for plat, configTemplate := range configTemplates {
		variables[plat] = &PackageReleasePluginGetConfigVariablesResp_ConfigVariablesList{
			Items: conv.SliceToSlice(configTemplate, func(template *types.PluginPkgConfigTemplate) *ConfigVariables {
				result := &ConfigVariables{
					Name:          template.Name,
					FilePath:      template.FilePath,
					SourcePath:    template.SourcePath,
					IsMainConfig:  template.IsMainConfig,
					SourceContent: template.SourceContent,
					Variables:     make(map[string]*ConfigVariables_Property),
				}

				for key, property := range template.Variables {
					result.Variables[key] = convertPluginPkgConfigTemplatePropertyFromTypes(property)
				}

				return result
			}),
		}
	}

	x.Data = &PackageReleasePluginGetConfigVariablesResp_Data{
		ConfigVariables: variables,
	}
}

func convertPluginPkgConfigTemplatePropertyFromTypes(property *types.PluginPkgConfigTemplateProperty,
) *ConfigVariables_Property {

	if property == nil {
		return nil
	}

	result := &ConfigVariables_Property{
		Title:         property.Title,
		Type:          property.Type,
		Required:      property.Required,
		Description:   property.Description,
		DescriptionEn: property.DescriptionEn,
		Properties:    make(map[string]*ConfigVariables_Property),
	}

	for key, child := range property.Properties {
		result.Properties[key] = convertPluginPkgConfigTemplatePropertyFromTypes(child)
	}

	if property.Default != nil {
		defaultValue, err := structpb.NewValue(property.Default)
		if err == nil {
			result.Default = defaultValue
		}
	}

	return result
}

// ==================== PackageReleaseCert ====================

// Validate check body.
func (x *PackageReleaseCertListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCertListReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseCertDeleteReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCertDeleteReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseCertDeleteResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCertDeleteResp) AutoConvert() {
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseCertListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseCert) {
	items := make([]*ReleaseCert, len(releases))
	for idx, release := range releases {
		item := newEmptyReleaseCert()
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
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseBinToolListReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseBinToolDeleteReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseBinToolDeleteReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleaseBinToolDeleteResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseBinToolDeleteResp) AutoConvert() {
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleaseBinToolListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleaseBinTool) {
	items := make([]*ReleaseBinTool, len(releases))
	for idx, release := range releases {
		item := newEmptyReleaseBinTool()
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
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginBinToolListReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleasePluginBinToolDeleteReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginBinToolDeleteReq) AutoConvert() {
}

// Validate check body.
func (x *PackageReleasePluginBinToolDeleteResp) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginBinToolDeleteResp) AutoConvert() {
}

// ConvertReleasesFromTypes convert releases from types.
func (x *PackageReleasePluginBinToolListResp) ConvertReleasesFromTypes(total int64, releases []*types.ReleasePluginBinTool) {
	items := make([]*ReleasePluginBinTool, len(releases))
	for idx, release := range releases {
		item := newEmptyReleasePluginBinTool()
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
