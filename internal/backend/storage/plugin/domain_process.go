/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package plugin

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	daoProcess "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/process"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) markExpiredProcessesUnknown(nCtx contextx.IContext, hostIDs []int64, deadline time.Time) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(hostIDs) == 0 {
		return nil
	}

	err := s.daoProcess.UpdateInfoStatus(nCtx, types.ProcessStatusUnknown,
		daoProcess.WithHostID(hostIDs...), daoProcess.WithInfoLastSyncAtBefore(deadline))
	if err != nil {
		return fmt.Errorf("failed to mark expired processes unknown: %w", err)
	}

	return nil
}
