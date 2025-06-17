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
	if x.Generation == 0 {
		return errors.New("generation is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *UploadOriginAgentReq) AutoConvert() {
}

// ConvertResultFromTypes convert result from types.
func (x *UploadOriginAgentResp) ConvertResultFromTypes(detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, convertPlatformFromTypes(plat))
	}

	x.Data = &UploadOriginAgentResp_Data{
		Version:     detail.Version,
		Name:        detail.Name,
		Size:        detail.Size,
		Md5:         detail.MD5,
		ChangelogEn: detail.ChangeLogEN,
		ChangelogZh: detail.ChangeLogZH,
		Platforms:   plats,
	}
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
func (x *UploadOriginServerResp) ConvertResultFromTypes(detail *types.OriginPkgDetail) {
	if detail == nil {
		return
	}

	plats := make([]*Platform, 0)
	for _, plat := range detail.Platforms {
		plats = append(plats, convertPlatformFromTypes(plat))
	}

	x.Data = &UploadOriginServerResp_Data{
		Version:   detail.Version,
		Name:      detail.Name,
		Size:      detail.Size,
		Md5:       detail.MD5,
		Platforms: plats,
	}
}
