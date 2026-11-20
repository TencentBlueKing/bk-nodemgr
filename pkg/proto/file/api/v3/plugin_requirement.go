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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// Validate checks the file name without decoding the Base64 content.
func (x *PluginRequirementCreateReq) Validate() error {
	return types.ValidatePluginRequirementFileName(x.GetName())
}

// AutoConvert preserves the supplied name and Base64 content.
func (x *PluginRequirementCreateReq) AutoConvert() {}

// Validate checks the record ID and replacement file name.
func (x *PluginRequirementUpdateReq) Validate() error {
	if err := validatePluginRequirementID(x.GetId()); err != nil {
		return err
	}

	return types.ValidatePluginRequirementFileName(x.GetName())
}

// AutoConvert preserves the supplied record ID, name and Base64 content.
func (x *PluginRequirementUpdateReq) AutoConvert() {}

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
