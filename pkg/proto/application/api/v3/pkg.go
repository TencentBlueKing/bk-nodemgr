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

// Validate check request body.
func (x *PackageUploadOriginAgentReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginAgentReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginAgentResp) ConvertResultFromTypes(generated bool, detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &PackageUploadOriginAgentResp_Data{
		UploadId:    new(string),
		Existed:     new(bool),
		Generated:   new(bool),
		Name:        new(string),
		Size:        new(int64),
		Md5:         new(string),
		Version:     new(string),
		ChangeLogEn: new(string),
		ChangeLogZh: new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.ChangeLogEn = detail.ChangeLogEN
	*data.ChangeLogZh = detail.ChangeLogZH
	data.Platforms = plats

	x.Data = data
}

// Validate check request body.
func (x *PackageUploadOriginServerReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginServerReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginServerResp) ConvertResultFromTypes(generated bool, detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &PackageUploadOriginServerResp_Data{
		UploadId:    new(string),
		Existed:     new(bool),
		Generated:   new(bool),
		Name:        new(string),
		Size:        new(int64),
		Md5:         new(string),
		Version:     new(string),
		ChangeLogEn: new(string),
		ChangeLogZh: new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.ChangeLogEn = detail.ChangeLogEN
	*data.ChangeLogZh = detail.ChangeLogZH
	data.Platforms = plats

	x.Data = data
}

// Validate check request body.
func (x *PackageUploadOriginProxyReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginProxyReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginProxyResp) ConvertResultFromTypes(generated bool, detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &PackageUploadOriginProxyResp_Data{
		UploadId:    new(string),
		Existed:     new(bool),
		Generated:   new(bool),
		Name:        new(string),
		Size:        new(int64),
		Md5:         new(string),
		Version:     new(string),
		ChangeLogEn: new(string),
		ChangeLogZh: new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.ChangeLogEn = detail.ChangeLogEN
	*data.ChangeLogZh = detail.ChangeLogZH
	data.Platforms = plats

	x.Data = data
}

// Validate check request body.
func (x *PackageUploadOriginCertReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginCertReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginCertResp) ConvertResultFromTypes(generated bool, detail *types.OriginCertPkgDetail) {
	if detail == nil {
		return
	}

	data := &PackageUploadOriginCertResp_Data{
		UploadId:  new(string),
		Existed:   new(bool),
		Generated: new(bool),
		Name:      new(string),
		Size:      new(int64),
		Md5:       new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	data.CertFiles = detail.CertFiles

	x.Data = data
}

// Validate check request body.
func (x *PackageUploadOriginBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginBinToolReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginBinToolResp) ConvertResultFromTypes(generated bool, detail *types.OriginBinToolPkgDetail) {
	if detail == nil {
		return
	}

	agentPlats := make([]*Platform, 0)
	for _, plat := range detail.AgentPlatforms {
		agentPlats = append(agentPlats, ConvertPlatformFromTypes(plat))
	}
	proxyPlats := make([]*Platform, 0)
	for _, plat := range detail.ProxyPlatforms {
		proxyPlats = append(proxyPlats, ConvertPlatformFromTypes(plat))
	}

	data := &PackageUploadOriginBinToolResp_Data{
		UploadId:       new(string),
		Existed:        new(bool),
		Generated:      new(bool),
		Name:           new(string),
		Size:           new(int64),
		Md5:            new(string),
		AgentPlatforms: agentPlats,
		ProxyPlatforms: proxyPlats,
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	data.AgentPlatforms = agentPlats
	data.ProxyPlatforms = proxyPlats

	x.Data = data
}

// Validate check request body.
func (x *PackageUploadOriginPluginBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginPluginBinToolReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginPluginBinToolResp) ConvertResultFromTypes(generated bool, detail *types.OriginPluginBinToolPkgDetail) {
	if detail == nil {
		return
	}

	platsV2 := make([]*Platform, 0)
	for _, plat := range detail.V2.Platforms {
		platsV2 = append(platsV2, ConvertPlatformFromTypes(plat))
	}
	platsV3 := make([]*Platform, 0)
	for _, plat := range detail.V3.Platforms {
		platsV3 = append(platsV3, ConvertPlatformFromTypes(plat))
	}

	data := &PackageUploadOriginPluginBinToolResp_Data{
		UploadId:  new(string),
		Existed:   new(bool),
		Generated: new(bool),
		Name:      new(string),
		Size:      new(int64),
		Md5:       new(string),
		V2:        new(PackageUploadOriginPluginBinToolResp_Data_V2Info),
		V3:        new(PackageUploadOriginPluginBinToolResp_Data_V3Info),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	data.V2.Platforms = platsV2
	data.V3.Platforms = platsV3

	x.Data = data
}

// Validate check request body.
func (x *PackagePublishReleaseAgentReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleaseAgentReq) AutoConvert() {
}

// Validate check request body.
func (x *PackagePublishReleaseProxyReq) Validate() error {
	if x.GetUploadOriginPkgType() == "" {
		return errors.New("upload_origin_pkg_type is required")
	}

	switch types.UploadCategory(x.GetUploadOriginPkgType()) {
	case types.UploadCategoryOriginProxy, types.UploadCategoryOriginServer:
		return nil
	default:
		return fmt.Errorf("unsupported upload origin pkg type: %s", x.GetUploadOriginPkgType())
	}
}

// AutoConvert auto convert.
func (x *PackagePublishReleaseProxyReq) AutoConvert() {
}

// Validate check request body.
func (x *PackagePublishReleaseCertReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleaseCertReq) AutoConvert() {
}

// Validate check request body.
func (x *PackagePublishReleaseBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleaseBinToolReq) AutoConvert() {
}

// Validate check request body.
func (x *PackagePublishReleasePluginBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleasePluginBinToolReq) AutoConvert() {
}

// Validate check request body.
func (x *PackageReleaseAgentDownloadReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetPlatform() == nil {
		return errors.New("platform is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseAgentDownloadReq) AutoConvert() {
}

// Validate check request body.
func (x *PackageReleaseProxyDownloadReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetPlatform() == nil {
		return errors.New("platform is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseProxyDownloadReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *PackageReleaseAgentDownloadReq) GetIdentifier() (
	types.Generation, platfmt.Platform, string) {

	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// GetIdentifier get identifier.
func (x *PackageReleaseProxyDownloadReq) GetIdentifier() (
	types.Generation, platfmt.Platform, string) {

	return types.Generation(x.GetGeneration()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// GetIdentifier get identifier.
func (x *PackageReleasePluginDownloadReq) GetIdentifier() (
	string, platfmt.Platform, string) {

	return x.GetName(),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// Validate check request body.
func (x *PackageReleasePluginDownloadReq) Validate() error {
	if x.GetName() == "" {
		return errors.New("name is required")
	}

	if x.GetPlatform() == nil {
		return errors.New("platform is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginDownloadReq) AutoConvert() {
}

// Validate check request body.
func (x *PackageReleaseCertDownloadReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseCertDownloadReq) AutoConvert() {
}

// Validate check request body.
func (x *PackageReleaseBinToolDownloadReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleaseBinToolDownloadReq) AutoConvert() {
}

// Validate check request body.
func (x *PackageReleasePluginBinToolDownloadReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetName() == "" {
		return errors.New("name is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *PackageReleasePluginBinToolDownloadReq) AutoConvert() {
}

// Validate check request body.
func (x *PackageUploadOriginPluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginPluginV2Req) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginPluginV2Resp) ConvertResultFromTypes(generated bool, detail *types.OriginPluginV2PkgDetail) {
	if detail == nil {
		return
	}
	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}
	data := &PackageUploadOriginPluginV2Resp_Data{
		UploadId:      new(string),
		Existed:       new(bool),
		Generated:     new(bool),
		Name:          new(string),
		Size:          new(int64),
		Md5:           new(string),
		Version:       new(string),
		Description:   new(string),
		DescriptionEn: new(string),
		Scenario:      new(string),
		ScenarioEn:    new(string),
		ConfigFile:    new(string),
		ConfigFormat:  new(string),
		LaunchNode:    new(string),
		Platforms:     plats,
	}
	*x = PackageUploadOriginPluginV2Resp{
		Code:      0,
		Message:   "success",
		RequestId: "",
		Error:     nil,
		Data:      data,
	}
	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.Description = detail.Description
	*data.DescriptionEn = detail.DescriptionEn
	*data.Scenario = detail.Scenario
	*data.ScenarioEn = detail.ScenarioEn
	*data.ConfigFile = detail.ConfigFile
	*data.ConfigFormat = detail.ConfigFormat
	*data.LaunchNode = detail.LaunchNode
}

// Validate check request body.
func (x *PackageUploadOriginExternalPluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginExternalPluginV2Req) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginExternalPluginV2Resp) ConvertResultFromTypes(generated bool, detail *types.OriginExternalPluginV2PkgDetail) {
	if detail == nil {
		return
	}
	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}
	data := &PackageUploadOriginExternalPluginV2Resp_Data{
		UploadId:      new(string),
		Existed:       new(bool),
		Generated:     new(bool),
		Name:          new(string),
		Size:          new(int64),
		Md5:           new(string),
		Version:       new(string),
		Description:   new(string),
		DescriptionEn: new(string),
		Scenario:      new(string),
		ScenarioEn:    new(string),
		ConfigFile:    new(string),
		ConfigFormat:  new(string),
		LaunchNode:    new(string),
		Platforms:     plats,
	}
	*x = PackageUploadOriginExternalPluginV2Resp{
		Code:      0,
		Message:   "success",
		RequestId: "",
		Error:     nil,
		Data:      data,
	}
	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.Description = detail.Description
	*data.DescriptionEn = detail.DescriptionEn
	*data.Scenario = detail.Scenario
	*data.ScenarioEn = detail.ScenarioEn
	*data.ConfigFile = detail.ConfigFile
	*data.ConfigFormat = detail.ConfigFormat
	*data.LaunchNode = detail.LaunchNode
}

// Validate check request body.
func (x *PackageUploadOriginPluginV3Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackageUploadOriginPluginV3Req) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *PackageUploadOriginPluginV3Resp) ConvertResultFromTypes(generated bool, detail *types.OriginPluginV3PkgDetail) {
	if detail == nil {
		return
	}
	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}
	data := &PackageUploadOriginPluginV3Resp_Data{
		UploadId:         new(string),
		Existed:          new(bool),
		Generated:        new(bool),
		Name:             new(string),
		Size:             new(int64),
		Md5:              new(string),
		Version:          new(string),
		Description:      new(string),
		DescriptionEn:    new(string),
		Scenario:         new(string),
		ScenarioEn:       new(string),
		LaunchNode:       new(string),
		TemplateRenderer: new(string),
		Platforms:        plats,
	}
	*x = PackageUploadOriginPluginV3Resp{
		Code:      0,
		Message:   "success",
		RequestId: "",
		Error:     nil,
		Data:      data,
	}
	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.Description = detail.Description
	*data.DescriptionEn = detail.DescriptionEn
	*data.Scenario = detail.Scenario
	*data.ScenarioEn = detail.ScenarioEn
	*data.LaunchNode = detail.LaunchNode
	*data.TemplateRenderer = string(detail.TemplateRenderer)
}

// Validate check request body.
func (x *PackagePublishReleasePluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleasePluginV2Req) AutoConvert() {
}

// Validate check request body.
func (x *PackagePublishReleaseExternalPluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleaseExternalPluginV2Req) AutoConvert() {
}

// Validate check request body.
func (x *PackagePublishReleasePluginV3Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *PackagePublishReleasePluginV3Req) AutoConvert() {
}

func newEmptyReleasePlugin() *ReleasePlugin {
	plugin := &ReleasePlugin{
		Release: newEmptyRelease(),
	}

	return plugin
}

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

// ConvertPageToTypes convert page to types.
func (x *PackageReleasePluginListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
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

// ConvertConditionsToTypes convert conditions to types.
func (x *PackageReleasePluginListReq) ConvertConditionsToTypes() *types.ReleaseCondition {
	return convertReleaseConditionsToTypes(x.GetExactIncludeConditions())
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
	if x.GetPlatform() == nil {
		x.Platform = &Platform{}
	}

	if x.GetPlatform().GetOsType() == "" {
		x.Platform.OsType = string(criteria.OSUnknown)
	}

	if x.GetPlatform().GetCpuArch() == "" {
		x.Platform.CpuArch = string(criteria.CPUArchUnknown)
	}
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
