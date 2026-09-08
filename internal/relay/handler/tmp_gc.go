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

package handler

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
)

const (
	// stagingGCInterval controls how often abandoned staging directories are reclaimed.
	stagingGCInterval = 1 * time.Hour

	// stagingOrphanAge is how long a staging directory may stay untouched before it is
	// considered abandoned. A single installation writes its packages within minutes, so a
	// day of inactivity means the flow that owned the directory never came back to clean up.
	stagingOrphanAge = 24 * time.Hour
)

// StartStagingGC reclaims staging directories left behind by installations that died between
// CheckPkgStats, which creates the directory, and StoragePkg, which removes it. Without this
// every crashed installation would leak its transferred packages onto the relay disk.
func (h *handler) StartStagingGC(nCtx contextx.IContext) {
	h.collectOrphanStagingDirs(nCtx)

	go func() {
		ticker := time.NewTicker(stagingGCInterval)
		defer ticker.Stop()

		for {
			select {
			case <-nCtx.Done():
				logger.G.Sys().Info("stopping staging dir gc")

				return
			case <-ticker.C:
				h.collectOrphanStagingDirs(nCtx)
			}
		}
	}()
}

func (h *handler) collectOrphanStagingDirs(nCtx contextx.IContext) {
	orphans, err := h.storageFS.listOrphanInstanceDirs(time.Now().Add(-stagingOrphanAge))
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to list orphan staging dirs")

		return
	}

	for _, dir := range orphans {
		if err := h.storageFS.removeAll(dir); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("staging-dir", dir).Error("failed to remove orphan staging dir")

			continue
		}

		logger.G.Biz(nCtx).With("staging-dir", dir).Info("removed orphan staging dir")
	}
}
