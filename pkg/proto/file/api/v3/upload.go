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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check request body.
func (x *UploadAgentReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetCpuArch() == "" {
		return errors.New("cpu_arch is required")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *UploadAgentReq) AutoConvert() {
}

// Validate check request body.
func (x *UploadProxyReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	if x.GetCpuArch() == "" {
		return errors.New("cpu_arch is required")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type is required")
	}

	if x.GetVersion() == "" {
		return errors.New("version is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *UploadProxyReq) AutoConvert() {
}

// Validate check request body.
func (x *UploadOriginAgentReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginAgentReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginAgentResp) ConvertResultFromTypes(generated bool, detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginAgentResp_Data{
		UploadId:    new(string),
		Existed:     new(bool),
		Generated:   new(bool),
		Name:        new(string),
		Size:        new(int64),
		Md5:         new(string),
		Version:     new(string),
		ChangelogEn: new(string),
		ChangelogZh: new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.ChangelogEn = detail.ChangeLogEN
	*data.ChangelogZh = detail.ChangeLogZH
	data.Platforms = plats

	x.Data = data
}

// ConvertPlatformsToTypes convert platforms to types.
func (x *UploadOriginAgentResp_Data) ConvertPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetPlatforms() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	return plats
}

// Validate check request body.
func (x *UploadOriginServerReq) Validate() error {
	if x.GetGeneration() == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginServerReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginServerResp) ConvertResultFromTypes(generated bool, detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginServerResp_Data{
		UploadId:  new(string),
		Existed:   new(bool),
		Generated: new(bool),
		Name:      new(string),
		Size:      new(int64),
		Md5:       new(string),
		Version:   new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	data.Platforms = plats

	x.Data = data
}

// ConvertPlatformsToTypes convert platforms to types.
func (x *UploadOriginServerResp_Data) ConvertPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetPlatforms() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	return plats
}

// Validate check request body.
func (x *UploadOriginCertReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginCertReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginCertResp) ConvertResultFromTypes(generated bool, detail *types.OriginCertPkgDetail) {
	if detail == nil {
		return
	}

	data := &UploadOriginCertResp_Data{
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
func (x *UploadOriginBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginBinToolReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginBinToolResp) ConvertResultFromTypes(generated bool, detail *types.OriginBinToolPkgDetail) {
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

	data := &UploadOriginBinToolResp_Data{
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

// ConvertAgentPlatformsToTypes convert agent platforms to types.
func (x *UploadOriginBinToolResp_Data) ConvertAgentPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetAgentPlatforms() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	return plats
}

// ConvertProxyPlatformsToTypes convert proxy platforms to types.
func (x *UploadOriginBinToolResp_Data) ConvertProxyPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetProxyPlatforms() {
		plats = append(plats, ConvertPlatformToTypes(plat))
	}

	return plats
}

// Validate check request body.
func (x *UploadOriginPluginBinToolReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginPluginBinToolReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginPluginBinToolResp) ConvertResultFromTypes(generated bool, detail *types.OriginPluginBinToolPkgDetail) {
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

	data := &UploadOriginPluginBinToolResp_Data{
		UploadId:  new(string),
		Existed:   new(bool),
		Generated: new(bool),
		Name:      new(string),
		Size:      new(int64),
		Md5:       new(string),
		V2:        new(UploadOriginPluginBinToolResp_Data_V2Info),
		V3:        new(UploadOriginPluginBinToolResp_Data_V3Info),
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
func (x *UploadOriginPluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginPluginV2Req) AutoConvert() {}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginPluginV2Resp) ConvertResultFromTypes(generated bool, detail *types.OriginPluginV2PkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginPluginV2Resp_Data{
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

	data.Platforms = plats

	x.Data = data
}

// ConvertPlatformsToTypes convert platforms to types.
func (x *UploadOriginPluginV2Resp_Data) ConvertPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return plats
}

// Validate check request body.
func (x *UploadOriginExternalPluginV2Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginExternalPluginV2Req) AutoConvert() {}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginExternalPluginV2Resp) ConvertResultFromTypes(generated bool, detail *types.OriginExternalPluginV2PkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginExternalPluginV2Resp_Data{
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

	data.Platforms = plats

	x.Data = data
}

// ConvertPlatformsToTypes convert platforms to types.
func (x *UploadOriginExternalPluginV2Resp_Data) ConvertPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return plats
}

// Validate check request body.
func (x *UploadOriginPluginV3Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginPluginV3Req) AutoConvert() {}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginPluginV3Resp) ConvertResultFromTypes(generated bool, detail *types.OriginPluginV3PkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginPluginV3Resp_Data{
		UploadId:         new(string),
		Existed:          new(bool),
		Generated:        new(bool),
		Name:             new(string),
		Size:             new(int64),
		Md5:              new(string),
		PluginPkgName:    new(string),
		Version:          new(string),
		Description:      new(string),
		DescriptionEn:    new(string),
		Scenario:         new(string),
		ScenarioEn:       new(string),
		LaunchNode:       new(string),
		TemplateRenderer: new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.FileInfo.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.PluginPkgName = detail.PluginPkgName
	*data.Version = detail.Version
	*data.Description = detail.Description
	*data.DescriptionEn = detail.DescriptionEn
	*data.Scenario = detail.Scenario
	*data.ScenarioEn = detail.ScenarioEn
	*data.LaunchNode = detail.LaunchNode
	*data.TemplateRenderer = string(detail.TemplateRenderer)

	data.Platforms = plats

	x.Data = data
}

// ConvertPlatformsToTypes convert platforms to types.
func (x *UploadOriginPluginV3Resp_Data) ConvertPlatformsToTypes() []platform.Platform {
	plats := make([]platform.Platform, 0)
	for _, plat := range x.GetPlatforms() {
		plats = append(plats, platform.Platform{
			OS:   criteria.OSType(plat.GetOsType()),
			Arch: criteria.CPUArch(plat.GetCpuArch()),
		})
	}

	return plats
}
