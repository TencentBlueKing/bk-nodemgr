/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package metrics

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metricType int

const (
	none metricType = iota
	counter
	gauge
	histogram
	summary

	defaultMetricPath = "/metrics"
	defaultSlowTime   = 1 * time.Second
)

func defaultExcludePaths() []string {
	return []string{}
}

func defaultDurationMSBuckets() []float64 {
	return []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}
}

type metricKey struct {
	requestTotal    string
	requestUVTotal  string
	uriRequestTotal string
	requestBody     string
	responseBody    string
	requestDuration string
	slowRequest     string
}

// Monitor is an object that uses to set gin server monitor.
type Monitor struct {
	slowTime          time.Duration
	metricPath        string
	excludePaths      []string
	durationMSBuckets []float64
	metrics           map[string]*metric
	bloomFilter       *bloomFilter
	typeHandler       map[metricType]func(metric *metric) error
	metricKey         *metricKey
}

// NewMonitor return a new monitor.
func NewMonitor(name string) *Monitor {
	monitor := &Monitor{
		metricPath:        defaultMetricPath,
		slowTime:          defaultSlowTime,
		excludePaths:      defaultExcludePaths(),
		durationMSBuckets: defaultDurationMSBuckets(),
		metrics:           make(map[string]*metric),
		bloomFilter:       newBloomFilter(),
		typeHandler: map[metricType]func(metric *metric) error{
			counter:   counterHandler,
			gauge:     gaugeHandler,
			histogram: histogramHandler,
			summary:   summaryHandler,
		},
		metricKey: &metricKey{
			requestTotal:    "request_total",
			requestUVTotal:  "request_uv_total",
			uriRequestTotal: "uri_request_total",
			requestBody:     "request_body",
			responseBody:    "response_body",
			requestDuration: "request_duration",
			slowRequest:     "slow_request",
		},
	}
	monitor.setMetricPrefix(strings.ReplaceAll(name, "-", "_") + "_")

	return monitor
}

// GetMetric used to get metric object by metric_name.
func (monitor *Monitor) getMetric(name string) (*metric, error) {
	if metric, ok := monitor.metrics[name]; ok {
		return metric, nil
	}

	return nil, fmt.Errorf("metric not found. name(%s)", name)
}

// WithMetricPath set metricPath property. metricPath is used for Prometheus
// to get gin server monitoring data.
func (monitor *Monitor) WithMetricPath(path string) *Monitor {
	monitor.metricPath = path

	return monitor
}

// WithExcludePaths set exclude paths which should not be reported (e.g. /ping /healthz...)
func (monitor *Monitor) WithExcludePaths(paths []string) *Monitor {
	monitor.excludePaths = paths

	return monitor
}

// WithSlowTime set slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (monitor *Monitor) WithSlowTime(slowTime time.Duration) *Monitor {
	monitor.slowTime = slowTime

	return monitor
}

// WithDurationMSBuckets set duration metric buckets in milliseconds. if not set then use default value.
func (monitor *Monitor) WithDurationMSBuckets(durationMSBuckets []float64) *Monitor {
	if len(durationMSBuckets) > 0 {
		monitor.durationMSBuckets = durationMSBuckets
	}

	return monitor
}

// SetMetricPrefix set metric prefix.
func (monitor *Monitor) setMetricPrefix(prefix string) {
	monitor.metricKey.requestTotal = prefix + monitor.metricKey.requestTotal
	monitor.metricKey.requestUVTotal = prefix + monitor.metricKey.requestUVTotal
	monitor.metricKey.uriRequestTotal = prefix + monitor.metricKey.uriRequestTotal
	monitor.metricKey.requestBody = prefix + monitor.metricKey.requestBody
	monitor.metricKey.responseBody = prefix + monitor.metricKey.responseBody
	monitor.metricKey.requestDuration = prefix + monitor.metricKey.requestDuration
	monitor.metricKey.slowRequest = prefix + monitor.metricKey.slowRequest
}

// AddMetric add custom monitor metric.
func (monitor *Monitor) AddMetric(metric *metric) error {
	if _, ok := monitor.metrics[metric.Name]; ok {
		return fmt.Errorf("metric already exists. name(%s)", metric.Name)
	}

	if metric.Name == "" {
		return errors.New("metric name cannot be empty")
	}
	if f, ok := monitor.typeHandler[metric.Type]; ok {
		if err := f(metric); err == nil {
			promErr := prometheus.Register(metric.vec)
			if promErr != nil {
				return promErr
			}

			monitor.metrics[metric.Name] = metric

			return nil
		}
	}

	return generateErrorTypeNotSupport(metric.Type, metric.Name)
}

// nolint:unparam
// for matching with func(metric *metric) error.
func counterHandler(metric *metric) error {
	metric.vec = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: metric.Name, Help: metric.Description},
		metric.Labels,
	)

	return nil
}

// nolint:unparam
// for matching with func(metric *metric) error.
func gaugeHandler(metric *metric) error {
	metric.vec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: metric.Name, Help: metric.Description},
		metric.Labels,
	)

	return nil
}

func histogramHandler(metric *metric) error {
	if len(metric.Buckets) == 0 {
		return fmt.Errorf("histogram type must specify buckets. name(%s)", metric.Name)
	}

	metric.vec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    metric.Name,
			Help:    metric.Description,
			Buckets: metric.Buckets,
		},
		metric.Labels,
	)

	return nil
}

func summaryHandler(metric *metric) error {
	if len(metric.Objectives) == 0 {
		return fmt.Errorf("summary type must specify objectives. name(%s)", metric.Name)
	}

	metric.vec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       metric.Name,
			Help:       metric.Description,
			Objectives: metric.Objectives,
		},
		metric.Labels,
	)

	return nil
}
