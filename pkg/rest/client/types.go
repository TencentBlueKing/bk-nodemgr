/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package client

import (
	"time"

	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
)

const (
	// ToleranceLatencyTimeDefault default tolerance latency time.
	ToleranceLatencyTimeDefault = 500 * time.Millisecond
)

// Capability http request limit.
type Capability struct {
	// HTTPClient name for logging and metrics.
	Name string

	// HTTPClient http client.
	HTTPClient HTTPClient

	// Discover get request address.
	Discover restdiscovery.Interface

	// the max tolerance api request latency time, if exceeded this time, then
	// this request will be logged and warned.
	ToleranceLatencyTime time.Duration

	// MetricOpts metric option.
	MetricOpts MetricOption

	// TraceSvc trace service.
	TraceSvc tracing.IService
}

// MetricOption metrics options.
type MetricOption struct {
	// if not set, use default buckets value.
	// in milliseconds.
	DurationMSBuckets []float64
}

// Logger is the logger interface.
// nolint: interfacebloat
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugw(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorw(args ...interface{})
}
