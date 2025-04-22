/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bkrepo

import (
	"context"
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/iface"
	"golang.org/x/sync/singleflight"
	"io"
	"time"
)

// CheckIntervalFile check interval for file.
const CheckIntervalFile = time.Second * 30

// CheckIntervalSubGroup check interval for sub group.
const CheckIntervalSubGroup = time.Minute

type FileGroup struct {
	lastCheckTimeFile     time.Time
	lastCheckTimeSubGroup time.Time

	name        string
	subGroupMap map[string]iface.FileGroup
	fileMap     map[string]iface.File

	sg                   singleflight.Group
	refreshSubGroupMapFn func() error
	refreshFileMapFn     func() error

	storeFn func(ctx context.Context, info iface.FileInfo, reader io.ReadCloser, overwrite bool) error
}

// Name return the name of the file group.
func (group *FileGroup) Name() string {
	return group.name
}

// SubGroups return the sub groups of the file group.
func (group *FileGroup) SubGroups() ([]iface.FileGroup, error) {
	if time.Since(group.lastCheckTimeSubGroup) > CheckIntervalSubGroup {
		if err := group.refreshSubGroupMapFn(); err != nil {
			return nil, fmt.Errorf("refresh sub group failed, err: %w", err)
		}
	}

	return conv.MapToSlice(group.subGroupMap), nil
}

// GetFile return the file of the file group.
func (group *FileGroup) GetFile(name string) (iface.File, error) {
	if time.Since(group.lastCheckTimeFile) > CheckIntervalFile {
		if err := group.refreshFileMapFn(); err != nil {
			return nil, fmt.Errorf("refresh file failed, err: %w", err)
		}
	}

	file, ok := group.fileMap[name]
	if !ok {
		return nil, fmt.Errorf("file not found, name(%s)", name)
	}

	return file, nil
}

// AllFiles return all files of the file group.
func (group *FileGroup) AllFiles() ([]iface.File, error) {
	if time.Since(group.lastCheckTimeFile) > CheckIntervalFile {
		if err := group.refreshFileMapFn(); err != nil {
			return nil, fmt.Errorf("refresh file failed, err: %w", err)
		}
	}

	return conv.MapToSlice(group.fileMap), nil
}

// Store store a file to the file group.
func (group *FileGroup) Store(ctx context.Context, info iface.FileInfo, reader io.ReadCloser, overwrite bool) error {
	return group.storeFn(ctx, info, reader, overwrite)
}
