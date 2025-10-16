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
	*data.Name = detail.Name
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
	*data.Name = detail.Name
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
	*data.Name = detail.Name
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
	*data.Name = detail.Name
	*data.Size = detail.Size
	*data.Md5 = detail.MD5
	data.AgentPlatforms = agentPlats
	data.ProxyPlatforms = proxyPlats

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
	return nil
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
