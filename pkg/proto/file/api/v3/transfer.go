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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check request body.
func (x *TransferLaunchReleaseReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TransferLaunchReleaseReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *TransferLaunchReleaseReq) GetIdentifier() (
	types.Generation, types.ReleaseType, platform.Platform, string) {

	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		ConvertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// ConvertResult convert result.
func (x *TransferLaunchReleaseResp) ConvertResult(tf types.ISimpleTransferHandler) {
	data := &TransferLaunchReleaseResp_Data{
		TaskId:          new(string),
		ReleaseFileName: new(string),
		ReleaseFileSize: new(int64),
		ReleaseFileMd5:  new(string),
	}

	*data.TaskId = tf.GetTaskID()

	info := tf.GetFileInfo()
	*data.ReleaseFileName = info.Name
	*data.ReleaseFileSize = info.Size
	*data.ReleaseFileMd5 = info.MD5

	x.Data = data
}

// Validate check request body.
func (x *TransferLaunchInstallerReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TransferLaunchInstallerReq) AutoConvert() {
}

// ConvertResult convert result.
func (x *TransferLaunchInstallerResp) ConvertResult(tf types.ISimpleTransferHandler) {
	data := &TransferLaunchInstallerResp_Data{
		TaskId:            new(string),
		InstallerFileName: new(string),
		InstallerFileSize: new(int64),
		InstallerFileMd5:  new(string),
	}

	*data.TaskId = tf.GetTaskID()

	info := tf.GetFileInfo()
	*data.InstallerFileName = info.Name
	*data.InstallerFileSize = info.Size
	*data.InstallerFileMd5 = info.MD5

	x.Data = data
}

// Validate check request body.
func (x *TransferQueryReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TransferQueryReq) AutoConvert() {
}

// ConvertResult convert result.
func (x *TransferQueryResp) ConvertResult(upload, download *types.SimpleTransferResult) {
	data := &TransferQueryResp_Data{
		Upload:    ConvertSimpleTransferFromTypes(upload),
		Download:  ConvertSimpleTransferFromTypes(download),
		StartTime: new(int64),
		EndTime:   new(int64),
	}

	*data.StartTime = upload.StartTime.UnixMilli()
	*data.EndTime = download.EndTime.UnixMilli()

	x.Data = data
}

// ConvertSimpleTransferToTypes convert simple transfer to types.
func ConvertSimpleTransferToTypes(result *SimpleTransferResult) *types.SimpleTransferResult {
	return &types.SimpleTransferResult{
		ErrorCode:     int(result.GetErrorCode()),
		ErrorMessage:  result.GetErrorMessage(),
		Terminated:    result.GetTerminated(),
		Progress:      uint(result.GetProgress()),
		SpeedKBPerSec: uint64(result.GetSpeedKbytesPerSec()),
	}
}

// ConvertSimpleTransferFromTypes convert types to simple transfer.
func ConvertSimpleTransferFromTypes(result *types.SimpleTransferResult) *SimpleTransferResult {
	data := newEmptySimpleTransferResult()
	*data.ErrorCode = int64(result.ErrorCode)
	*data.ErrorMessage = result.ErrorMessage
	*data.Terminated = result.Terminated
	*data.Progress = int64(result.Progress)
	*data.SpeedKbytesPerSec = int64(result.SpeedKBPerSec)

	return data
}

func newEmptySimpleTransferResult() *SimpleTransferResult {
	return &SimpleTransferResult{
		ErrorCode:         new(int64),
		ErrorMessage:      new(string),
		Terminated:        new(bool),
		Progress:          new(int64),
		SpeedKbytesPerSec: new(int64),
	}
}
