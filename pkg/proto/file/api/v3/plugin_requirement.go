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

// Validate checks the file name and applicability without decoding content.
func (x *PluginRequirementCreateReq) Validate() error {
	if err := types.ValidatePluginRequirementFileName(x.GetName()); err != nil {
		return err
	}

	return x.ConvertScopeToTypes().Validate()
}

// AutoConvert defaults omitted applicability dimensions to unrestricted.
func (x *PluginRequirementCreateReq) AutoConvert() {
	x.Os = conv.NonEmptyOr(x.GetOs(), types.PluginRequirementScopeAll)
	x.Arch = conv.NonEmptyOr(x.GetArch(), types.PluginRequirementScopeAll)
	x.Version = conv.NonEmptyOr(x.GetVersion(), types.PluginRequirementScopeAll)
}

// ConvertScopeToTypes converts applicability without interpreting all as a platform.
func (x *PluginRequirementCreateReq) ConvertScopeToTypes() types.PluginRequirementFileScope {
	return types.PluginRequirementFileScope{
		OS: criteria.OSType(x.GetOs()), Arch: criteria.CPUArch(x.GetArch()), Version: x.GetVersion(),
	}
}

// ConvertScopeFromTypes sets the applicability dimensions for an upload request.
func (x *PluginRequirementCreateReq) ConvertScopeFromTypes(scope types.PluginRequirementFileScope) {
	x.Os, x.Arch, x.Version = string(scope.OS), string(scope.Arch), scope.Version
}

// Validate checks the record ID and replacement file name.
func (x *PluginRequirementUpdateReq) Validate() error {
	if err := validatePluginRequirementID(x.GetId()); err != nil {
		return err
	}

	return types.ValidatePluginRequirementFileName(x.GetName())
}

// AutoConvert preserves the supplied record ID, name and Base64 content.
func (x *PluginRequirementUpdateReq) AutoConvert() {}

// Validate requires an explicit complete applicability replacement.
func (x *PluginRequirementUpdateScopeReq) Validate() error {
	if err := validatePluginRequirementID(x.GetId()); err != nil {
		return err
	}

	return x.ConvertScopeToTypes().Validate()
}

// AutoConvert preserves scope fields so an omission cannot accidentally broaden a scope.
func (x *PluginRequirementUpdateScopeReq) AutoConvert() {}

// ConvertScopeToTypes converts a complete applicability replacement.
func (x *PluginRequirementUpdateScopeReq) ConvertScopeToTypes() types.PluginRequirementFileScope {
	return types.PluginRequirementFileScope{
		OS: criteria.OSType(x.GetOs()), Arch: criteria.CPUArch(x.GetArch()), Version: x.GetVersion(),
	}
}

// ConvertScopeFromTypes sets the applicability replacement for a request.
func (x *PluginRequirementUpdateScopeReq) ConvertScopeFromTypes(scope types.PluginRequirementFileScope) {
	x.Os, x.Arch, x.Version = string(scope.OS), string(scope.Arch), scope.Version
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
	for _, id := range exact.GetId() {
		if err := validatePluginRequirementID(id); err != nil {
			return err
		}
	}
	for _, os := range exact.GetOs() {
		scope := types.PluginRequirementFileScope{OS: criteria.OSType(os), Arch: types.PluginRequirementScopeAll, Version: types.PluginRequirementScopeAll}
		if err := scope.Validate(); err != nil {
			return err
		}
	}
	for _, arch := range exact.GetArch() {
		scope := types.PluginRequirementFileScope{OS: types.PluginRequirementScopeAll, Arch: criteria.CPUArch(arch), Version: types.PluginRequirementScopeAll}
		if err := scope.Validate(); err != nil {
			return err
		}
	}
	for _, version := range exact.GetVersion() {
		scope := types.PluginRequirementFileScope{OS: types.PluginRequirementScopeAll, Arch: types.PluginRequirementScopeAll, Version: version}
		if err := scope.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AutoConvert preserves exact filters; an empty filter does not select all-only records.
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
		ID: exact.GetId(), Name: exact.GetName(),
		OS:      conv.SliceToSlice(exact.GetOs(), func(value string) criteria.OSType { return criteria.OSType(value) }),
		Arch:    conv.SliceToSlice(exact.GetArch(), func(value string) criteria.CPUArch { return criteria.CPUArch(value) }),
		Version: exact.GetVersion(),
	}}
}

// ConvertPluginRequirementFileConditionsFromTypes converts exact filters for File clients.
func ConvertPluginRequirementFileConditionsFromTypes(condition *types.PluginRequirementFileCondition) *PluginRequirementFileConditions {
	if condition == nil || condition.ExactInclude == nil {
		return nil
	}
	exact := condition.ExactInclude

	return &PluginRequirementFileConditions{ExactInclude: &PluginRequirementFileExactConditions{
		Id: exact.ID, Name: exact.Name,
		Os:      conv.SliceToSlice(exact.OS, func(value criteria.OSType) string { return string(value) }),
		Arch:    conv.SliceToSlice(exact.Arch, func(value criteria.CPUArch) string { return string(value) }),
		Version: exact.Version,
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
	return validatePluginRequirementID(x.GetId())
}

// AutoConvert preserves the supplied record ID.
func (x *PluginRequirementGetReq) AutoConvert() {}

// Validate checks the requested download record ID.
func (x *PluginRequirementDownloadReq) Validate() error {
	return validatePluginRequirementID(x.GetId())
}

// AutoConvert preserves the supplied download record ID.
func (x *PluginRequirementDownloadReq) AutoConvert() {}

// Validate checks the file record ID and transfer destination.
func (x *PluginRequirementTransferReq) Validate() error {
	if err := validatePluginRequirementID(x.GetId()); err != nil {
		return err
	}
	if err := types.ValidatePluginRequirementFileName(x.GetTargetName()); err != nil {
		return fmt.Errorf("invalid target_name: %w", err)
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
		return fmt.Errorf("invalid id: %w", err)
	}

	return nil
}

// ConvertPluginRequirementFileRecordFromTypes converts a stored file record.
func ConvertPluginRequirementFileRecordFromTypes(record *types.PluginRequirementFileRecord) *PluginRequirementFileRecord {
	if record == nil {
		return nil
	}

	return &PluginRequirementFileRecord{
		Id:        record.ID,
		Name:      record.Name,
		SavedName: record.SavedName,
		Md5:       record.MD5,
		Size:      record.Size,
		UpdatedAt: record.UpdatedAt.Unix(),
		Operator:  record.Operator,
		Os:        string(record.OS),
		Arch:      string(record.Arch),
		Version:   record.Version,
	}
}

// ConvertToTypes converts protocol file metadata to a domain record.
func (x *PluginRequirementFileRecord) ConvertToTypes() *types.PluginRequirementFileRecord {
	if x == nil {
		return nil
	}

	return &types.PluginRequirementFileRecord{
		ID:        x.GetId(),
		Name:      x.GetName(),
		SavedName: x.GetSavedName(),
		MD5:       x.GetMd5(),
		Size:      x.GetSize(),
		UpdatedAt: time.Unix(x.GetUpdatedAt(), 0),
		Operator:  x.GetOperator(),
		OS:        criteria.OSType(conv.NonEmptyOr(x.GetOs(), types.PluginRequirementScopeAll)),
		Arch:      criteria.CPUArch(conv.NonEmptyOr(x.GetArch(), types.PluginRequirementScopeAll)),
		Version:   conv.NonEmptyOr(x.GetVersion(), types.PluginRequirementScopeAll),
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
		TaskId:   new(string),
		FileName: new(string),
		FileMd5:  new(string),
		FileSize: new(int64),
	}

	*data.TaskId = tf.GetTaskID()
	info := tf.GetFileInfo()
	*data.FileName = info.Name
	*data.FileMd5 = info.MD5
	*data.FileSize = info.Size

	x.Data = data
}
