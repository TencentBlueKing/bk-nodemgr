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
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check request body.
func (x *TransferReleaseLaunchReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TransferReleaseLaunchReq) AutoConvert() {
}

// GetIdentifier get identifier.
func (x *TransferReleaseLaunchReq) GetIdentifier() (
	types.Generation, types.ReleaseType, platform.Platform, string) {

	return types.Generation(x.GetGeneration()),
		types.ReleaseType(x.GetReleaseType()),
		convertPlatformToTypes(x.GetPlatform()),
		x.GetVersion()
}

// ConvertResult convert result.
func (x *TransferReleaseLaunchResp) ConvertResult(tf manager.ITransfer) {
	data := &TransferReleaseLaunchResp_Data{
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
func (x *TransferReleaseQueryReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TransferReleaseQueryReq) AutoConvert() {
}

// ConvertResult convert result.
func (x *TransferReleaseQueryResp) ConvertResult(upload, download *types.TransferResult) {
	uploadData := newEmptyTransferProgress()
	*uploadData.ErrorCode = int64(upload.ErrorCode)
	*uploadData.ErrorMessage = upload.ErrorMessage
	*uploadData.Progress = int64(upload.Progress)
	*uploadData.SpeedKbytesPerSec = int64(upload.SpeedKBPerSec)
	*uploadData.Terminated = upload.StatusCode == types.TransferStatusEndUploading

	downloadData := newEmptyTransferProgress()
	*downloadData.ErrorCode = int64(download.ErrorCode)
	*downloadData.ErrorMessage = download.ErrorMessage
	*downloadData.Progress = int64(download.Progress)
	*downloadData.SpeedKbytesPerSec = int64(download.SpeedKBPerSec)
	*downloadData.Terminated = download.StatusCode == types.TransferStatusEndDownloading

	data := &TransferReleaseQueryResp_Data{
		Upload:    uploadData,
		Download:  downloadData,
		StartTime: new(int64),
		EndTime:   new(int64),
	}

	*data.StartTime = upload.StartTime.UnixMilli()
	*data.EndTime = download.EndTime.UnixMilli()

	x.Data = data
}

func newEmptyTransferProgress() *TransferProgress {
	return &TransferProgress{
		ErrorCode:         new(int64),
		ErrorMessage:      new(string),
		Terminated:        new(bool),
		Progress:          new(int64),
		SpeedKbytesPerSec: new(int64),
	}
}
