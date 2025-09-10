/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package metric provides metric for workflow.
package metric

import (
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/metrics"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	defaultSlowTime = 1 * time.Second
)

func defaultDurationMSBuckets() []float64 {
	return []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}
}

type metricSets struct {
	operationInstanceLaunchedTotal     *metrics.Metric
	operationInstanceLaunchedDuration  *metrics.Metric
	operationInstanceProcessedTotal    *metrics.Metric
	operationInstanceProcessedDuration *metrics.Metric
	actionReceivedTotal                *metrics.Metric
	actionProcessedTotal               *metrics.Metric
	actionProcessedDuration            *metrics.Metric
	actionNotResigteredTotal           *metrics.Metric
	actionDataNotFoundTotal            *metrics.Metric
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
	monitor.metricSets = metricSets{
		operationInstanceLaunchedTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "operation_instance_launched_total",
			Description: "the workflow operation instance launched num.",
			Labels:      []string{"operation", "success"},
		},
		operationInstanceLaunchedDuration: &metrics.Metric{
			Type:        metrics.MetricTypeHistogram,
			Name:        metricNamePrefix + "operation_instance_launched_duration",
			Description: "the workflow operation instance launched duration. unit ms.",
			Labels:      []string{"operation", "success"},
			Buckets:     monitor.durationMSBuckets,
		},
		operationInstanceProcessedTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "operation_instance_processed_total",
			Description: "the workflow operation instance processed num.",
			Labels:      []string{"operation", "state", "success"},
		},
		operationInstanceProcessedDuration: &metrics.Metric{
			Type:        metrics.MetricTypeHistogram,
			Name:        metricNamePrefix + "operation_instance_processed_duration",
			Description: "the workflow operation instance processed duration. unit ms.",
			Labels:      []string{"operation", "state", "success"},
			Buckets:     monitor.durationMSBuckets,
		},
		actionReceivedTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "action_received_total",
			Description: "the workflow action received num.",
			Labels:      []string{"operation", "action", "success"},
		},
		actionProcessedTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "action_processed_total",
			Description: "the workflow action processed num.",
			Labels:      []string{"operation", "action", "state", "success"},
		},
		actionProcessedDuration: &metrics.Metric{
			Type:        metrics.MetricTypeHistogram,
			Name:        metricNamePrefix + "action_processed_duration",
			Description: "the workflow action processed duration. unit ms.",
			Labels:      []string{"operation", "action", "state", "success"},
			Buckets:     monitor.durationMSBuckets,
		},
		actionNotResigteredTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "action_not_resigtered_total",
			Description: "the workflow action not registered num.",
			Labels:      []string{"action"},
		},
		actionDataNotFoundTotal: &metrics.Metric{
			Type:        metrics.MetricTypeCounter,
			Name:        metricNamePrefix + "action_data_not_found_total",
			Description: "the workflow action data not found num.",
			Labels:      []string{"action"},
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
	_ = monitor.metricSets.operationInstanceLaunchedTotal.Enable()
	_ = monitor.metricSets.operationInstanceLaunchedDuration.Enable()
	_ = monitor.metricSets.operationInstanceProcessedTotal.Enable()
	_ = monitor.metricSets.operationInstanceProcessedDuration.Enable()
	_ = monitor.metricSets.actionReceivedTotal.Enable()
	_ = monitor.metricSets.actionProcessedTotal.Enable()
	_ = monitor.metricSets.actionProcessedDuration.Enable()

	return monitor
}

// RecordOperationInstanceLaunched record operation instance launched.
func (monitor *Monitor) RecordOperationInstanceLaunched(param Param) {
	labels := []string{param.OperationName, strconv.FormatBool(param.Err == nil)}

	// set operation instance launched total.
	_ = monitor.metricSets.operationInstanceLaunchedTotal.Inc(labels)

	// set operation instance launched duration.
	_ = monitor.metricSets.operationInstanceLaunchedDuration.Observe(labels, float64(param.ProcessDuration.Milliseconds()))
}

// RecordOperationInstanceProcessed record operation instance processed.
func (monitor *Monitor) RecordOperationInstanceProcessed(param Param) {
	labels := []string{
		param.OperationName,
		string(param.OperationState),
		strconv.FormatBool(param.OperationState == operation.StateSuccess),
	}

	// set operation instance processed total.
	_ = monitor.metricSets.operationInstanceProcessedTotal.Inc(labels)

	// set operation instance processed duration.
	_ = monitor.metricSets.operationInstanceProcessedDuration.Observe(labels, float64(param.ProcessDuration.Milliseconds()))
}

// RecordActionReceived record action received.
func (monitor *Monitor) RecordActionReceived(param Param) {
	labels := []string{param.OperationName, param.ActionName, strconv.FormatBool(param.Err == nil)}

	// set action received total.
	_ = monitor.metricSets.actionReceivedTotal.Inc(labels)
}

// RecordActionProcessed record action processed.
func (monitor *Monitor) RecordActionProcessed(param Param) {
	labels := []string{
		param.OperationName,
		param.ActionName,
		string(param.ActionState),
		strconv.FormatBool(param.ActionState == action.StateSuccess),
	}

	// set action processed total.
	_ = monitor.metricSets.actionProcessedTotal.Inc(labels)

	// set action processed duration.
	_ = monitor.metricSets.actionProcessedDuration.Observe(labels, float64(param.ProcessDuration.Milliseconds()))
}

// RecordActionNotResigter record action not registered.
func (monitor *Monitor) RecordActionNotResigter(param Param) {
	// set action not registered total.
	_ = monitor.metricSets.actionNotResigteredTotal.Inc([]string{param.ActionName})
}

// RecordActionDataNotFound record action data not found.
func (monitor *Monitor) RecordActionDataNotFound(param Param) {
	// set action data not found total.
	_ = monitor.metricSets.actionDataNotFoundTotal.Inc([]string{param.ActionName})
}

// Param the metric param.
type Param struct {
	OperationName string
	ActionName    string

	OperationState operation.State
	ActionState    action.State

	// the returned error.
	Err error

	ProcessDuration time.Duration
}
