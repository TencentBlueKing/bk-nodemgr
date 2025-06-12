/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package precheck ...
package precheck

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// MBSize this is MB.
const MBSize = 1024 * 1024

// CheckDiskFreeSpace ...
func CheckDiskFreeSpace(requires []DiskRequire) error {
	gp := gopool.NewPool()
	for idx := range requires {
		req := requires[idx]
		gp.Go(func() error {
			absDirPath, err := filepath.Abs(req.DirPath)
			if err != nil {
				return fmt.Errorf("get abs path failed, err: %v", err)
			}

			freeSpace, err := utils.CountDiskFreeSpace(absDirPath)
			if err != nil {
				return fmt.Errorf("CountDiskFreeSpace failed, err: %v", err)
			}

			freeMB := freeSpace / MBSize

			if freeMB < req.DemandMB {
				return fmt.Errorf("the free disk space is not enough, "+
					"demand-space(%dMB) ,dir-path(%s) , free-space(%dMB)",
					req.DemandMB, req.DirPath, freeMB)
			}

			logger.Infof(constant.StepPreCheck,
				"dir has enough free space, dir(%s), free-space(%dMB)",
				req.DirPath, freeMB)

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("CheckDiskFreeSpace failed, err: %v", err)
	}

	return nil
}

// DiskRequire ...
type DiskRequire struct {
	DemandMB uint64 `json:"demand_mb"`
	DirPath  string `json:"dir_path"`
}

// Validate validate DiskRequire.
func (require *DiskRequire) Validate() error {
	if require.DemandMB == 0 {
		return errors.New("DiskRequire.DemandMB is required")
	}

	if require.DirPath == "" {
		return errors.New("DiskRequire.DirPath is required")
	}

	return nil
}
