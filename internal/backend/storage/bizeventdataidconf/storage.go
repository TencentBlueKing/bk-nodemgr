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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "biz_event_data_id_conf"

	metricOperationGetBizEventDataIDConf              = "get_biz_event_data_id_conf"
	metricOperationUpdateBizAgentBaseAlarmEventDataID = "update_biz_agent_base_alarm_event_data_id"
)

// NewStorage creates a new business event data-id config storage handler.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}

	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

// Storage provides a business event data-id config storage handler.
type Storage struct {
	basestorage.Storage

	daoBizEventDataIDConf bizEventDataIDConfDao
}

// GetBizEventDataIDConf gets business event data-id config by business ID.
func (s *Storage) GetBizEventDataIDConf(
	nCtx contextx.IContext,
	bkBizID int64,
) (types.BizEventDataIDConf, bool, error) {

	var conf types.BizEventDataIDConf
	var found bool
	err := s.WrapFn(nCtx, metricOperationGetBizEventDataIDConf, func(ctx contextx.IContext) error {
		var err error
		conf, found, err = s.getBizEventDataIDConf(ctx, bkBizID)

		return err
	})

	return conf, found, err
}

// UpdateBizAgentBaseAlarmEventDataID updates agent base alarm event data-id by business ID.
func (s *Storage) UpdateBizAgentBaseAlarmEventDataID(nCtx contextx.IContext, bkBizID, eventDataID int64) error {
	return s.WrapFn(nCtx, metricOperationUpdateBizAgentBaseAlarmEventDataID, func(ctx contextx.IContext) error {
		return s.updateBizAgentBaseAlarmEventDataID(ctx, bkBizID, eventDataID)
	})
}
