/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package bizeventdataidconf

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/bizeventdataidconf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

type bizEventDataIDConfDao = bizeventdataidconf.IHandler

func (s *Storage) initDao() error {
	s.daoBizEventDataIDConf = bizeventdataidconf.New(s.Database)
	return nil
}

func (s *Storage) check() error {
	if s.daoBizEventDataIDConf == nil {
		return errors.New("dao biz event data-id conf is nil")
	}

	return nil
}

func (s *Storage) getBizEventDataIDConf(
	nCtx contextx.IContext,
	bkBizID int64,
) (types.BizEventDataIDConf, bool, error) {

	conf, err := s.daoBizEventDataIDConf.Get(nCtx, bkBizID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return types.BizEventDataIDConf{}, false, nil
	}
	if err != nil {
		return types.BizEventDataIDConf{}, false, fmt.Errorf("failed to get biz event data-id conf: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return types.BizEventDataIDConf{}, false, fmt.Errorf("invalid biz event data-id conf: %w", err)
	}

	return *conf, true, nil
}

func (s *Storage) updateBizAgentBaseAlarmEventDataID(nCtx contextx.IContext, bkBizID, eventDataID int64) error {
	if err := s.daoBizEventDataIDConf.UpdateAgentBaseAlarmEventDataID(nCtx, bkBizID, eventDataID); err != nil {
		return fmt.Errorf("failed to update biz agent base alarm event data-id: %w", err)
	}

	return nil
}
