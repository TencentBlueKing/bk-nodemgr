/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package mongo is the mongo package.
package mongo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/metrics"
)

const (
	// DefaultSlowTime is the default threshold for slow requests.
	DefaultSlowTime = 1 * time.Second
)

func defaultDurationMSBuckets() []float64 {
	return []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}
}

type metricSets struct {
	requestTotal       *metrics.Metric
	requestDataLength  *metrics.Metric
	responseDataLength *metrics.Metric
	requestDuration    *metrics.Metric
	slowRequest        *metrics.Metric
}

// Monitor is a wrapper for prometheus metrics.
type Monitor struct {
	slowTime          time.Duration
	durationMSBuckets []float64
	metricSets        metricSets
}

// NewMonitor return a new monitor.
func NewMonitor(name string, opts ...OptFunc) *Monitor {
	monitor := &Monitor{
		slowTime:          DefaultSlowTime,
		durationMSBuckets: defaultDurationMSBuckets(),
	}

	for _, opt := range opts {
		opt(monitor)
	}

	metricNamePrefix := strings.ReplaceAll(name, "-", "_") + "_"
	metricLabels := []string{"table", "operation", "success"}
	monitor.metricSets = metricSets{
		requestTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_total",
			Description: "the database received request num.",
			Labels:      metricLabels,
		},
		requestDataLength: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_data_length",
			Description: "the database received request data length.",
			Labels:      metricLabels,
		},
		responseDataLength: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "response_data_length",
			Description: "the database received response data length.",
			Labels:      metricLabels,
		},
		requestDuration: &metrics.Metric{
			Type:        metrics.MetricTypeHistogram,
			Name:        metricNamePrefix + "request_duration",
			Description: "the time database took to handle the request. unit ms.",
			Labels:      metricLabels,
			Buckets:     monitor.durationMSBuckets,
		},
		slowRequest: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "slow_request",
			Description: fmt.Sprintf("the database handled slow requests counter, t=%dms.", monitor.slowTime.Milliseconds()),
			Labels:      metricLabels,
		},
	}

	return monitor
}

// OptFunc is a function that modifies the monitor.
type OptFunc func(monitor *Monitor)

// WithSlowTime set slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func WithSlowTime(slowTime time.Duration) OptFunc {
	return func(monitor *Monitor) {
		monitor.slowTime = slowTime
	}
}

// WithDurationMSBuckets set duration metric buckets in milliseconds. if not set then use default value.
func WithDurationMSBuckets(durationMSBuckets []float64) OptFunc {
	return func(monitor *Monitor) {
		if len(durationMSBuckets) > 0 {
			monitor.durationMSBuckets = durationMSBuckets
		}
	}
}

// Enable enable metrics into prometheus handler.
func (monitor *Monitor) Enable() *Monitor {
	_ = monitor.metricSets.requestTotal.Enable()
	_ = monitor.metricSets.requestDataLength.Enable()
	_ = monitor.metricSets.responseDataLength.Enable()
	_ = monitor.metricSets.requestDuration.Enable()
	_ = monitor.metricSets.slowRequest.Enable()

	return monitor
}

// MetricHandle handle metrics.
// nolint:cyclop
func (monitor *Monitor) MetricHandle(param *MetricParam) {
	labels := []string{param.TableName, string(param.Operation), strconv.FormatBool(param.Err == nil)}

	// set request total
	_ = monitor.metricSets.requestTotal.Inc(labels)

	// set request data length
	if param.RequestDataLength > 0 {
		_ = monitor.metricSets.requestDataLength.Add(labels, float64(param.RequestDataLength))
	}

	// set response data length
	if param.ResponseDataLength > 0 {
		_ = monitor.metricSets.responseDataLength.Add(labels, float64(param.ResponseDataLength))
	}

	// set request duration
	_ = monitor.metricSets.requestDuration.Observe(labels, float64(param.ProcessDuration.Milliseconds()))

	// set slow request
	if param.ProcessDuration > monitor.slowTime {
		_ = monitor.metricSets.slowRequest.Inc(labels)
	}
}

// MetricOperation the operation to do.
type MetricOperation string

const (
	// MetricOperationCountDucuments the count documents operation.
	MetricOperationCountDucuments MetricOperation = "count_documents"

	// MetricOperationFindOne the find one operation.
	MetricOperationFindOne MetricOperation = "find_one"

	// MetricOperationFind the find operation.
	MetricOperationFind MetricOperation = "find"

	// MetricOperationInsertMany the insert many operation.
	MetricOperationInsertMany MetricOperation = "insert_many"

	// MetricOperationInsertOne the insert one operation.
	MetricOperationInsertOne MetricOperation = "insert_one"

	// MetricOperationUpdateMany the update many operation.
	MetricOperationUpdateMany MetricOperation = "update_many"

	// MetricOperationCreateIndexes the create indexes operation.
	MetricOperationCreateIndexes MetricOperation = "create_indexes"

	// MetricOperationDistinct the distinct operation.
	MetricOperationDistinct MetricOperation = "distinct"

	// MetricOperationBulkWrite the bulk write operation.
	MetricOperationBulkWrite MetricOperation = "bulk_write"
)

// MetricParam the metric param.
type MetricParam struct {
	// the table name to operate.
	TableName string

	// the operation to do.
	Operation MetricOperation

	// the request data length.
	RequestDataLength int

	// the response data length.
	ResponseDataLength int

	// the returned error.
	Err error

	ProcessDuration time.Duration
}
