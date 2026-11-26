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
	"errors"
	"fmt"
	"math"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// Validate checks the file identity, filename, and applicability.
func (x *PluginRequirementCreateReq) Validate() error {
	if err := x.ConvertKeyToTypes().Validate(); err != nil {
		return err
	}
	if err := types.ValidatePluginRequirementFileName(x.GetTargetFileName()); err != nil {
		return err
	}

	return x.ConvertScopeToTypes().Validate()
}

// AutoConvert preserves empty applicability dimensions as unrestricted.
func (x *PluginRequirementCreateReq) AutoConvert() {}

// ConvertKeyToTypes converts the owning plugin and dependency name.
func (x *PluginRequirementCreateReq) ConvertKeyToTypes() types.PluginRequirementFileKey {
	return types.PluginRequirementFileKey{PluginName: x.GetPluginName(), Name: x.GetName()}
}

// ConvertScopeToTypes converts the file's applicability.
func (x *PluginRequirementCreateReq) ConvertScopeToTypes() types.PluginRequirementFileScope {
	return convertPluginRequirementScope(x.GetOs(), x.GetArch(), x.GetVersion())
}

// Validate checks the record ID and replacement names.
func (x *PluginRequirementUpdateReq) Validate() error {
	if err := validatePluginRequirementID(x.GetRequirementFileId()); err != nil {
		return err
	}
	if err := types.ValidatePluginRequirementName(x.GetName()); err != nil {
		return err
	}

	return types.ValidatePluginRequirementFileName(x.GetTargetFileName())
}

// AutoConvert preserves the supplied values.
func (x *PluginRequirementUpdateReq) AutoConvert() {}

// Validate checks the record ID and replacement applicability.
func (x *PluginRequirementUpdateScopeReq) Validate() error {
	if err := validatePluginRequirementID(x.GetRequirementFileId()); err != nil {
		return err
	}

	return x.ConvertScopeToTypes().Validate()
}

// AutoConvert preserves empty applicability dimensions as unrestricted.
func (x *PluginRequirementUpdateScopeReq) AutoConvert() {}

// ConvertScopeToTypes converts the complete applicability replacement.
func (x *PluginRequirementUpdateScopeReq) ConvertScopeToTypes() types.PluginRequirementFileScope {
	return convertPluginRequirementScope(x.GetOs(), x.GetArch(), x.GetVersion())
}

func convertPluginRequirementScope(os, arch, version []string) types.PluginRequirementFileScope {
	return types.PluginRequirementFileScope{
		OS:      conv.SliceToSlice(os, func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:    conv.SliceToSlice(arch, func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version: version,
	}
}

// Validate checks the requested file record ID.
func (x *PluginRequirementGetReq) Validate() error {
	return validatePluginRequirementID(x.GetRequirementFileId())
}

// AutoConvert preserves the supplied record ID.
func (x *PluginRequirementGetReq) AutoConvert() {}

// Validate checks the requested file record ID.
func (x *PluginRequirementDownloadReq) Validate() error {
	return validatePluginRequirementID(x.GetRequirementFileId())
}

func validatePluginRequirementID(id string) error {
	if err := uuid.Validate(id); err != nil {
		return fmt.Errorf("invalid requirement_file_id: %w", err)
	}

	return nil
}

// Validate checks paging and exact filter values.
func (x *PluginRequirementListReq) Validate() error {
	if _, err := x.ConvertPageToTypes(); err != nil {
		return err
	}

	exact := x.GetConditions().GetExactInclude()
	for _, id := range exact.GetRequirementFileId() {
		if err := validatePluginRequirementID(id); err != nil {
			return err
		}
	}
	for _, name := range exact.GetName() {
		if err := types.ValidatePluginRequirementName(name); err != nil {
			return err
		}
	}
	for _, filename := range exact.GetTargetFileName() {
		if err := types.ValidatePluginRequirementFileName(filename); err != nil {
			return err
		}
	}

	return convertPluginRequirementScope(exact.GetOs(), exact.GetArch(), exact.GetVersion()).Validate()
}

// AutoConvert preserves the supplied exact filters.
func (x *PluginRequirementListReq) AutoConvert() {}

// ConvertPageToTypes converts the required paging settings.
func (x *PluginRequirementListReq) ConvertPageToTypes() (types.Page, error) {
	page := x.GetPage()
	if page == nil {
		return types.Page{}, errors.New("page is required")
	}
	if page.GetOffset() < 0 || page.GetOffset() > math.MaxInt {
		return types.Page{}, errors.New("page offset is out of range")
	}
	if page.GetLimit() <= 0 || page.GetLimit() > math.MaxInt {
		return types.Page{}, errors.New("page limit is out of range")
	}

	return types.Page{Offset: int(page.GetOffset()), Limit: int(page.GetLimit()), Sort: page.GetSort()}, nil
}

// ConvertConditionsToTypes converts the exact filters.
func (x *PluginRequirementListReq) ConvertConditionsToTypes() *types.PluginRequirementFileCondition {
	exact := x.GetConditions().GetExactInclude()
	scope := convertPluginRequirementScope(exact.GetOs(), exact.GetArch(), exact.GetVersion())

	return &types.PluginRequirementFileCondition{ExactInclude: &types.PluginRequirementFileExactFields{
		RequirementFileID: exact.GetRequirementFileId(),
		PluginName:        exact.GetPluginName(),
		Name:              exact.GetName(),
		TargetFileName:    exact.GetTargetFileName(),
		OS:                scope.OS,
		Arch:              scope.Arch,
		Version:           scope.Version,
	}}
}

// ConvertPluginRequirementFileFromTypes converts domain metadata to the API response.
func ConvertPluginRequirementFileFromTypes(record *types.PluginRequirementFileRecord) *PluginRequirementFile {
	if record == nil {
		return nil
	}

	return &PluginRequirementFile{
		RequirementFileId: record.RequirementFileID,
		PluginName:        record.PluginName,
		Name:              record.Name,
		TargetFileName:    record.TargetFileName,
		FileName:          record.FileName,
		Md5:               record.MD5,
		Size:              record.Size,
		UpdatedAt:         record.UpdatedAt.Unix(),
		Operator:          record.Operator,
		Os:                conv.SliceToSlice(record.OS, func(value criteria.OSType) string { return string(value) }),
		Arch:              conv.SliceToSlice(record.Arch, func(value criteria.CPUArch) string { return string(value) }),
		Version:           conv.NonEmptyOr(record.Version, []string{}),
	}
}

// ConvertResultFromTypes converts the created file record to a response.
func (x *PluginRequirementCreateResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementCreateResp_Data{File: ConvertPluginRequirementFileFromTypes(record)}
}

// ConvertResultFromTypes converts the updated file record to a response.
func (x *PluginRequirementUpdateResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementUpdateResp_Data{File: ConvertPluginRequirementFileFromTypes(record)}
}

// ConvertResultFromTypes converts the scope-updated file record to a response.
func (x *PluginRequirementUpdateScopeResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementUpdateScopeResp_Data{File: ConvertPluginRequirementFileFromTypes(record)}
}

// ConvertResultFromTypes converts the queried file record to a response.
func (x *PluginRequirementGetResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementGetResp_Data{File: ConvertPluginRequirementFileFromTypes(record)}
}

// ConvertResultFromTypes converts a page of file records.
func (x *PluginRequirementListResp) ConvertResultFromTypes(records []*types.PluginRequirementFileRecord, total int64) {
	x.Data = &PluginRequirementListResp_Data{
		Total: total,
		Items: conv.SliceToSlice(records, ConvertPluginRequirementFileFromTypes),
	}
}
