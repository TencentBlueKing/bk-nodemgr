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

package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/metrics"
)

const (
	defaultSlowTime = 1 * time.Second
)

func defaultExcludePaths() []string {
	return []string{}
}

func defaultDurationMSBuckets() []float64 {
	return []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}
}

type metricSets struct {
	requestTotal    *metrics.Metric
	requestUVTotal  *metrics.Metric
	requestBody     *metrics.Metric
	responseBody    *metrics.Metric
	requestDuration *metrics.Metric
	slowRequest     *metrics.Metric
}

// Monitor is an object that uses to set gin server monitor.
type Monitor struct {
	excludePaths      []string
	slowTime          time.Duration
	durationMSBuckets []float64
	bloomFilter       *bloomFilter
	metricSets        metricSets
}

// NewMonitor return a new monitor.
func NewMonitor(name string, opts ...OptFunc) *Monitor {
	monitor := &Monitor{
		excludePaths:      defaultExcludePaths(),
		slowTime:          defaultSlowTime,
		durationMSBuckets: defaultDurationMSBuckets(),
		bloomFilter:       newBloomFilter(),
	}

	for _, opt := range opts {
		opt(monitor)
	}

	metricNamePrefix := strings.ReplaceAll(name, "-", "_") + "_"
	metricLabels := []string{"uri", "method", "code"}
	monitor.metricSets = metricSets{
		requestTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_total",
			Description: "the server received request num.",
			Labels:      metricLabels,
		},
		requestUVTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_uv_total",
			Description: "the server received ip num.",
			Labels:      metricLabels,
		},
		requestBody: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "request_body",
			Description: "the server received request body size, unit byte",
			Labels:      metricLabels,
		},
		responseBody: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "response_body",
			Description: "the server send response body size, unit byte",
			Labels:      metricLabels,
		},
		requestDuration: &metrics.Metric{
			Type:        metrics.MetricTypeHistogram,
			Name:        metricNamePrefix + "request_duration",
			Description: "the time server took to handle the request. unit ms.",
			Labels:      metricLabels,
			Buckets:     monitor.durationMSBuckets,
		},
		slowRequest: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "slow_request",
			Description: fmt.Sprintf("the server handled slow requests counter, t=%dms.", monitor.slowTime.Milliseconds()),
			Labels:      metricLabels,
		},
	}

	return monitor
}

// OptFunc is a function that modifies the Monitor.
type OptFunc func(monitor *Monitor)

// WithExcludePaths set excludePaths property. excludePaths is used to determine
// whether the request should not be recorded.
func WithExcludePaths(excludePaths []string) OptFunc {
	return func(monitor *Monitor) {
		monitor.excludePaths = excludePaths
	}
}

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
