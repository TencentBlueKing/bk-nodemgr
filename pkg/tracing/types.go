/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tracing provides OpenTelemetry-based distributed tracing support for multiple services.
// This package is designed to support multiple services within the same program
// without using global state and supports multiple exporters (stdout, OTLP, Jaeger).
package tracing

import "fmt"

// ExporterType defines the type of exporter to use.
type ExporterType string

const (
	// ExporterTypeStdout exports traces to stdout.
	ExporterTypeStdout ExporterType = "stdout"
	// ExporterTypeOTLP exports traces using OpenTelemetry Protocol.
	ExporterTypeOTLP ExporterType = "otlp"
)

// Validate validates the exporter type.
func (et ExporterType) Validate() error {
	switch et {
	case ExporterTypeStdout, ExporterTypeOTLP:
		return nil
	default:
		return fmt.Errorf("invalid exporter type, exporter-type(%s)", et)
	}
}
