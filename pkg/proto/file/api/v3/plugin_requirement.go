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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// Validate checks the business key, downlink file name and applicability without decoding content.
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

// ConvertKeyToTypes converts the owning plugin and business dependency name.
func (x *PluginRequirementCreateReq) ConvertKeyToTypes() types.PluginRequirementFileKey {
	return types.PluginRequirementFileKey{PluginName: x.GetPluginName(), Name: x.GetName()}
}

// ConvertKeyFromTypes sets the owning plugin and business dependency name.
func (x *PluginRequirementCreateReq) ConvertKeyFromTypes(key types.PluginRequirementFileKey) {
	x.PluginName, x.Name = key.PluginName, key.Name
}

// ConvertScopeToTypes converts applicability dimensions, with empty slices meaning unrestricted.
func (x *PluginRequirementCreateReq) ConvertScopeToTypes() types.PluginRequirementFileScope {
	return types.PluginRequirementFileScope{
		OS:      conv.SliceToSlice(x.GetOs(), func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:    conv.SliceToSlice(x.GetArch(), func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version: conv.NonEmptyOr(x.GetVersion(), []string{}),
	}
}

// ConvertScopeFromTypes sets the applicability dimensions for an upload request.
func (x *PluginRequirementCreateReq) ConvertScopeFromTypes(scope types.PluginRequirementFileScope) {
	x.Os = conv.SliceToSlice(scope.OS, func(value criteria.OSType) string { return string(value) })
	x.Arch = conv.SliceToSlice(scope.Arch, func(value criteria.CPUArch) string { return string(value) })
	x.Version = conv.NonEmptyOr(scope.Version, []string{})
}

// Validate checks the record ID, business dependency name and replacement file name.
func (x *PluginRequirementUpdateReq) Validate() error {
	if err := validatePluginRequirementID(x.GetRequirementFileId()); err != nil {
		return err
	}
	if err := types.ValidatePluginRequirementName(x.GetName()); err != nil {
		return err
	}

	return types.ValidatePluginRequirementFileName(x.GetTargetFileName())
}

// AutoConvert preserves the supplied record ID, name and Base64 content.
func (x *PluginRequirementUpdateReq) AutoConvert() {}

// Validate checks the record ID and replacement applicability dimensions.
func (x *PluginRequirementUpdateScopeReq) Validate() error {
	if err := validatePluginRequirementID(x.GetRequirementFileId()); err != nil {
		return err
	}

	return x.ConvertScopeToTypes().Validate()
}

// AutoConvert preserves empty applicability dimensions as unrestricted.
func (x *PluginRequirementUpdateScopeReq) AutoConvert() {}

// ConvertScopeToTypes converts a complete applicability replacement.
func (x *PluginRequirementUpdateScopeReq) ConvertScopeToTypes() types.PluginRequirementFileScope {
	return types.PluginRequirementFileScope{
		OS:      conv.SliceToSlice(x.GetOs(), func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:    conv.SliceToSlice(x.GetArch(), func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version: conv.NonEmptyOr(x.GetVersion(), []string{}),
	}
}

// ConvertScopeFromTypes sets the applicability replacement for a request.
func (x *PluginRequirementUpdateScopeReq) ConvertScopeFromTypes(scope types.PluginRequirementFileScope) {
	x.Os = conv.SliceToSlice(scope.OS, func(value criteria.OSType) string { return string(value) })
	x.Arch = conv.SliceToSlice(scope.Arch, func(value criteria.CPUArch) string { return string(value) })
	x.Version = conv.NonEmptyOr(scope.Version, []string{})
}

// ConvertResultFromTypes converts a metadata-only update result.
func (x *PluginRequirementUpdateScopeResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementUpdateScopeResp_Data{File: ConvertPluginRequirementFileRecordFromTypes(record)}
}

// ConvertFileToTypes converts a metadata-only update response.
func (x *PluginRequirementUpdateScopeResp_Data) ConvertFileToTypes() *types.PluginRequirementFileRecord {
	return x.GetFile().ConvertToTypes()
}

// Validate checks pagination and exact filter values without applying wildcard matching.
func (x *PluginRequirementListReq) Validate() error {
	if err := x.GetPage().Validate(); err != nil {
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
	for _, fileName := range exact.GetTargetFileName() {
		if err := types.ValidatePluginRequirementFileName(fileName); err != nil {
			return err
		}
	}
	scope := types.PluginRequirementFileScope{
		OS:      conv.SliceToSlice(exact.GetOs(), func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:    conv.SliceToSlice(exact.GetArch(), func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version: exact.GetVersion(),
	}

	return scope.Validate()
}

// AutoConvert preserves exact filters; an empty filter does not restrict that field.
func (x *PluginRequirementListReq) AutoConvert() {}

// ConvertPageToTypes converts the shared pagination contract.
func (x *PluginRequirementListReq) ConvertPageToTypes() (types.Page, error) {
	return x.GetPage().ConvertToTypes()
}

// ConvertConditionsToTypes converts exact administrative filters.
func (x *PluginRequirementListReq) ConvertConditionsToTypes() *types.PluginRequirementFileCondition {
	exact := x.GetConditions().GetExactInclude()
	if exact == nil {
		return nil
	}

	return &types.PluginRequirementFileCondition{ExactInclude: &types.PluginRequirementFileExactFields{
		RequirementFileID: exact.GetRequirementFileId(),
		PluginName:        exact.GetPluginName(),
		Name:              exact.GetName(),
		TargetFileName:    exact.GetTargetFileName(),
		OS:                conv.SliceToSlice(exact.GetOs(), func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:              conv.SliceToSlice(exact.GetArch(), func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version:           exact.GetVersion(),
	}}
}

// ConvertPluginRequirementFileConditionsFromTypes converts exact filters for File clients.
func ConvertPluginRequirementFileConditionsFromTypes(condition *types.PluginRequirementFileCondition) *PluginRequirementFileConditions {
	if condition == nil || condition.ExactInclude == nil {
		return nil
	}
	exact := condition.ExactInclude

	return &PluginRequirementFileConditions{ExactInclude: &PluginRequirementFileExactConditions{
		RequirementFileId: conv.NonEmptyOr(exact.RequirementFileID, []string{}),
		PluginName:        conv.NonEmptyOr(exact.PluginName, []string{}),
		Name:              conv.NonEmptyOr(exact.Name, []string{}),
		TargetFileName:    conv.NonEmptyOr(exact.TargetFileName, []string{}),
		Os:                conv.SliceToSlice(exact.OS, func(value criteria.OSType) string { return string(value) }),
		Arch:              conv.SliceToSlice(exact.Arch, func(value criteria.CPUArch) string { return string(value) }),
		Version:           conv.NonEmptyOr(exact.Version, []string{}),
	}}
}

// ConvertResultFromTypes converts paged file records without file contents.
func (x *PluginRequirementListResp) ConvertResultFromTypes(records []*types.PluginRequirementFileRecord, total int64) {
	x.Data = &PluginRequirementListResp_Data{
		Total: total,
		Items: conv.SliceToSlice(records, ConvertPluginRequirementFileRecordFromTypes),
	}
}

// ConvertFilesToTypes converts a file list response for File clients.
func (x *PluginRequirementListResp_Data) ConvertFilesToTypes() []*types.PluginRequirementFileRecord {
	return conv.SliceToSlice(x.GetItems(), func(item *PluginRequirementFileRecord) *types.PluginRequirementFileRecord {
		return item.ConvertToTypes()
	})
}

// Validate checks the requested file record ID.
func (x *PluginRequirementGetReq) Validate() error {
	return validatePluginRequirementID(x.GetRequirementFileId())
}

// AutoConvert preserves the supplied record ID.
func (x *PluginRequirementGetReq) AutoConvert() {}

// Validate checks the requested download record ID.
func (x *PluginRequirementDownloadReq) Validate() error {
	return validatePluginRequirementID(x.GetRequirementFileId())
}

// AutoConvert preserves the supplied download record ID.
func (x *PluginRequirementDownloadReq) AutoConvert() {}

// Validate checks the file record ID and transfer destination.
func (x *PluginRequirementTransferReq) Validate() error {
	if err := validatePluginRequirementID(x.GetRequirementFileId()); err != nil {
		return err
	}
	if x.GetTargetDir() == "" {
		return errors.New("target_dir is required")
	}
	if x.GetTargetHostId() < 0 {
		return errors.New("target_host_id must be >= 0")
	}

	return nil
}

// AutoConvert preserves the supplied transfer destination.
func (x *PluginRequirementTransferReq) AutoConvert() {}

func validatePluginRequirementID(id string) error {
	if err := uuid.Validate(id); err != nil {
		return fmt.Errorf("invalid requirement_file_id: %w", err)
	}

	return nil
}

// ConvertPluginRequirementFileRecordFromTypes converts a stored file record.
func ConvertPluginRequirementFileRecordFromTypes(record *types.PluginRequirementFileRecord) *PluginRequirementFileRecord {
	if record == nil {
		return nil
	}

	return &PluginRequirementFileRecord{
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

// ConvertToTypes converts protocol file metadata to a domain record.
func (x *PluginRequirementFileRecord) ConvertToTypes() *types.PluginRequirementFileRecord {
	if x == nil {
		return nil
	}

	return &types.PluginRequirementFileRecord{
		RequirementFileID: x.GetRequirementFileId(),
		PluginName:        x.GetPluginName(),
		Name:              x.GetName(),
		TargetFileName:    x.GetTargetFileName(),
		FileName:          x.GetFileName(),
		MD5:               x.GetMd5(),
		Size:              x.GetSize(),
		UpdatedAt:         time.Unix(x.GetUpdatedAt(), 0),
		Operator:          x.GetOperator(),
		OS:                conv.SliceToSlice(x.GetOs(), func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:              conv.SliceToSlice(x.GetArch(), func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version:           conv.NonEmptyOr(x.GetVersion(), []string{}),
	}
}

// ConvertResultFromTypes converts the created file record to a response.
func (x *PluginRequirementCreateResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementCreateResp_Data{File: ConvertPluginRequirementFileRecordFromTypes(record)}
}

// ConvertFileToTypes converts the created file record to domain metadata.
func (x *PluginRequirementCreateResp_Data) ConvertFileToTypes() *types.PluginRequirementFileRecord {
	return x.GetFile().ConvertToTypes()
}

// ConvertResultFromTypes converts the updated file record to a response.
func (x *PluginRequirementUpdateResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementUpdateResp_Data{File: ConvertPluginRequirementFileRecordFromTypes(record)}
}

// ConvertFileToTypes converts the updated file record to domain metadata.
func (x *PluginRequirementUpdateResp_Data) ConvertFileToTypes() *types.PluginRequirementFileRecord {
	return x.GetFile().ConvertToTypes()
}

// ConvertResultFromTypes converts the queried file record to a response.
func (x *PluginRequirementGetResp) ConvertResultFromTypes(record *types.PluginRequirementFileRecord) {
	x.Data = &PluginRequirementGetResp_Data{File: ConvertPluginRequirementFileRecordFromTypes(record)}
}

// ConvertFileToTypes converts the queried file record to domain metadata.
func (x *PluginRequirementGetResp_Data) ConvertFileToTypes() *types.PluginRequirementFileRecord {
	return x.GetFile().ConvertToTypes()
}

// ConvertResult converts a launched file transfer to its response metadata.
func (x *PluginRequirementTransferResp) ConvertResult(tf types.ISimpleTransferHandler) {
	data := &PluginRequirementTransferResp_Data{
		TaskId:         new(string),
		TargetFileName: new(string),
		FileMd5:        new(string),
		FileSize:       new(int64),
	}

	*data.TaskId = tf.GetTaskID()
	info := tf.GetFileInfo()
	*data.TargetFileName = info.Name
	*data.FileMd5 = info.MD5
	*data.FileSize = info.Size

	x.Data = data
}
