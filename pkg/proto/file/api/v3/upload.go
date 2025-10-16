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
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.ChangelogEn = detail.ChangeLogEN
	*data.ChangelogZh = detail.ChangeLogZH
	data.Platforms = plats

	x.Data = data
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
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	data.Platforms = plats

	x.Data = data
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
	*data.Name = detail.Name
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
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	data.AgentPlatforms = agentPlats
	data.ProxyPlatforms = proxyPlats

	x.Data = data
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

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginPluginBinToolResp_Data{
		UploadId:  new(string),
		Existed:   new(bool),
		Generated: new(bool),
		Name:      new(string),
		Size:      new(int64),
		Md5:       new(string),
		Platforms: plats,
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	data.Platforms = plats

	x.Data = data
}

// Validate check request body.
func (x *UploadOriginOfficialPluginReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginOfficialPluginReq) AutoConvert() {}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginOfficialPluginResp) ConvertResultFromTypes(generated bool, detail *types.OriginOfficialPluginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginOfficialPluginResp_Data{
		UploadId:     new(string),
		Existed:      new(bool),
		Generated:    new(bool),
		Name:         new(string),
		Size:         new(int64),
		Md5:          new(string),
		Version:      new(string),
		Description:  new(string),
		Scenario:     new(string),
		ConfigFile:   new(string),
		ConfigFormat: new(string),
		LaunchNode:   new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.Description = detail.Description
	*data.Scenario = detail.Scenario
	*data.ConfigFile = detail.ConfigFile
	*data.ConfigFormat = detail.ConfigFormat
	*data.LaunchNode = detail.LaunchNode

	data.Platforms = plats

	x.Data = data
}

// Validate check request body.
func (x *UploadOriginExternalPluginReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginExternalPluginReq) AutoConvert() {}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginExternalPluginResp) ConvertResultFromTypes(generated bool, detail *types.OriginExternalPluginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, ConvertPlatformFromTypes(plat))
	}

	data := &UploadOriginExternalPluginResp_Data{
		UploadId:     new(string),
		Existed:      new(bool),
		Generated:    new(bool),
		Name:         new(string),
		Size:         new(int64),
		Md5:          new(string),
		Version:      new(string),
		Description:  new(string),
		Scenario:     new(string),
		ConfigFile:   new(string),
		ConfigFormat: new(string),
		LaunchNode:   new(string),
	}

	*data.UploadId = detail.UploadID
	*data.Existed = detail.Existed
	*data.Generated = generated
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	*data.Version = detail.Version
	*data.Description = detail.Description
	*data.Scenario = detail.Scenario
	*data.ConfigFile = detail.ConfigFile
	*data.ConfigFormat = detail.ConfigFormat
	*data.LaunchNode = detail.LaunchMode

	data.Platforms = plats

	x.Data = data
}
