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

package rediscache

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/metrics"
)

const (
	defaultSlowTime = 1 * time.Second
)

func defaultDurationMSBuckets() []float64 {
	return []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}
}

type metricSets struct {
	requestTotal     *metrics.Metric
	requestDataSize  *metrics.Metric
	responseDataSize *metrics.Metric
	requestDuration  *metrics.Metric
	slowRequest      *metrics.Metric
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
		slowTime:          defaultSlowTime,
		durationMSBuckets: defaultDurationMSBuckets(),
	}

	for _, opt := range opts {
		opt(monitor)
	}

	metricNamePrefix := strings.ReplaceAll(name, "-", "_") + "_"
	metricLabels := []string{"operation", "success"}
	monitor.metricSets = metricSets{
		requestTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_total",
			Description: "the redis received request num.",
			Labels:      metricLabels,
		},
		requestDataSize: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_data_size",
			Description: "the redis received request data size. in bytes.",
			Labels:      metricLabels,
		},
		responseDataSize: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "response_data_size",
			Description: "the redis received response data size. in bytes.",
			Labels:      metricLabels,
		},
		requestDuration: &metrics.Metric{
			Type:        metrics.MetricTypeHistogram,
			Name:        metricNamePrefix + "request_duration",
			Description: "the time redis took to handle the request. unit ms.",
			Labels:      metricLabels,
			Buckets:     monitor.durationMSBuckets,
		},
		slowRequest: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "slow_request",
			Description: fmt.Sprintf("the redis handled slow requests counter, t=%dms.", monitor.slowTime.Milliseconds()),
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
	_ = monitor.metricSets.requestDataSize.Enable()
	_ = monitor.metricSets.responseDataSize.Enable()
	_ = monitor.metricSets.requestDuration.Enable()
	_ = monitor.metricSets.slowRequest.Enable()

	return monitor
}

// MetricHandle handle metrics.
// nolint:cyclop
func (monitor *Monitor) MetricHandle(param *MetricParam) {
	labels := []string{string(param.Operation), strconv.FormatBool(param.Err == nil)}

	// set request total
	_ = monitor.metricSets.requestTotal.Inc(labels)

	// set request data size
	if param.RequestDataSize > 0 {
		_ = monitor.metricSets.requestDataSize.Add(labels, float64(param.RequestDataSize))
	}

	// set response data size
	if param.ResponseDataSize > 0 {
		_ = monitor.metricSets.responseDataSize.Add(labels, float64(param.ResponseDataSize))
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
	// MetricOperationGet get a value.
	MetricOperationGet MetricOperation = "get"

	// MetricOperationSet set a value.
	MetricOperationSet MetricOperation = "set"

	// MetricOperationSetNX set a value if not exist.
	MetricOperationSetNX MetricOperation = "setnx"

	// MetricOperationExists check if a key exists.
	MetricOperationExists MetricOperation = "exists"

	// MetricOperationDelete delete a key.
	MetricOperationDelete MetricOperation = "delete"
)

// MetricParam the metric param.
type MetricParam struct {
	// the operation to do.
	Operation MetricOperation

	// the request data size.
	RequestDataSize int

	// the response data size.
	ResponseDataSize int

	// the returned error.
	Err error

	ProcessDuration time.Duration
}
