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

// Package metrics provides common metrics handler.
package metrics

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricType defines the metric type.
type MetricType string

const (
	// MetricTypeNone none type.
	MetricTypeNone MetricType = "none"

	// MetricTypeCounter counter type.
	MetricTypeCounter MetricType = "counter"

	// MetricTypeGauge gauge type.
	MetricTypeGauge MetricType = "gauge"

	// MetricTypeHistogram histogram type.
	MetricTypeHistogram MetricType = "histogram"

	// MetricTypeSummary summary type.
	MetricTypeSummary MetricType = "summary"
)

// Metric defines a metric object. Users can use it to save
// metric data. Every metric should be globally unique by name.
type Metric struct {
	Type        MetricType
	Name        string
	Description string
	Labels      []string
	Buckets     []float64
	Objectives  map[float64]float64

	vec prometheus.Collector
}

// Enable enables the metric and register it into prometheus.
func (mc *Metric) Enable() error {
	if mc.Name == "" {
		return errors.New("metric name cannot be empty")
	}

	switch mc.Type {
	case MetricTypeCounter:
		mc.vec = prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: mc.Name, Help: mc.Description},
			mc.Labels,
		)

	case MetricTypeGauge:
		mc.vec = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: mc.Name, Help: mc.Description},
			mc.Labels,
		)

	case MetricTypeHistogram:
		if len(mc.Buckets) == 0 {
			return fmt.Errorf("histogram type must specify buckets. name(%s)", mc.Name)
		}

		mc.vec = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    mc.Name,
				Help:    mc.Description,
				Buckets: mc.Buckets,
			},
			mc.Labels,
		)

	case MetricTypeSummary:
		if len(mc.Objectives) == 0 {
			return fmt.Errorf("summary type must specify objectives. name(%s)", mc.Name)
		}

		mc.vec = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       mc.Name,
				Help:       mc.Description,
				Objectives: mc.Objectives,
			},
			mc.Labels,
		)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return prometheus.Register(mc.vec)
}

// SetGaugeValue set data for Gauge type Metric.
func (mc *Metric) SetGaugeValue(labelValues []string, value float64) error {
	switch mc.Type {
	case MetricTypeNone:
		return generateErrorDoesNotExist(mc.Name)

	case MetricTypeGauge:
		vec, ok := mc.vec.(*prometheus.GaugeVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeGauge, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Set(value)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

// Inc increases value for Counter/Gauge type metric, increments
// the counter by 1.
func (mc *Metric) Inc(labelValues []string) error {
	switch mc.Type {
	case MetricTypeNone:
		return generateErrorDoesNotExist(mc.Name)

	case MetricTypeGauge:
		vec, ok := mc.vec.(*prometheus.GaugeVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeGauge, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Inc()

	case MetricTypeCounter:
		vec, ok := mc.vec.(*prometheus.CounterVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeCounter, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Inc()

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

// Add adds the given value to the Metric object. Only
// for Counter/Gauge type metric.
func (mc *Metric) Add(labelValues []string, value float64) error {
	switch mc.Type {
	case MetricTypeNone:
		return generateErrorDoesNotExist(mc.Name)

	case MetricTypeGauge:
		vec, ok := mc.vec.(*prometheus.GaugeVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeGauge, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Add(value)

	case MetricTypeCounter:
		vec, ok := mc.vec.(*prometheus.CounterVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeCounter, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Add(value)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

// Observe is used by Histogram and Summary type metric to
// add observations.
func (mc *Metric) Observe(labelValues []string, value float64) error {
	switch mc.Type {
	case MetricTypeNone:
		return generateErrorDoesNotExist(mc.Name)

	case MetricTypeHistogram:
		vec, ok := mc.vec.(*prometheus.HistogramVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeHistogram, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Observe(value)

	case MetricTypeSummary:
		vec, ok := mc.vec.(*prometheus.SummaryVec)
		if !ok {
			return generateErrorNotSpecificType(MetricTypeSummary, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Observe(value)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

func generateErrorDoesNotExist(name string) error {
	return fmt.Errorf("metric does not exist. name(%s)", name)
}

func generateErrorTypeNotSupport(t MetricType, name string) error {
	return fmt.Errorf("metric not support type. name(%s), type(%s)", name, string(t))
}

func generateErrorNotSpecificType(t MetricType, name string) error {
	switch t {
	case MetricTypeNone, MetricTypeGauge, MetricTypeCounter, MetricTypeHistogram, MetricTypeSummary:
		return fmt.Errorf("metric type is not specific. name(%s), type(%s)", name, string(t))
	default:
		return fmt.Errorf("metric type is invalid. name(%s), type(%s)", name, string(t))
	}
}
