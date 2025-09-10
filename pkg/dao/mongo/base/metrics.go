/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package base

import (
	"time"

	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
)

var (
	// nolint: gochecknoglobals
	monitor *daomongo.Monitor
)

type metricData struct {
	param daomongo.MetricParam

	startTime time.Time
}

func (m *metricData) start(operation daomongo.MetricOperation, requestDataLength int) *metricData {
	m.startTime = time.Now()
	m.param.Operation = operation
	m.param.RequestDataLength = requestDataLength

	return m
}

func (m *metricData) end(err error, responseDataLength int) {
	m.param.ProcessDuration = time.Since(m.startTime)
	m.param.Err = err
	m.param.ResponseDataLength = responseDataLength

	monitor.MetricHandle(&m.param)
}

func newMetricData(tableName string) *metricData {
	return &metricData{
		param: daomongo.MetricParam{
			TableName: tableName,
		},
	}
}

// nolint: gochecknoinits
func init() {
	monitor = daomongo.NewMonitor("mongodb").Enable()
}
