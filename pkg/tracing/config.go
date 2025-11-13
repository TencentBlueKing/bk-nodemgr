/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tracing

import "fmt"

// Config is the configuration for the tracer manager.
type Config struct {
	Exporter    ExporterConfig
	Environment string
}

// Validate validates the configuration.
func (conf Config) Validate() error {
	if err := conf.Exporter.Validate(); err != nil {
		return fmt.Errorf("failed to validate exporter config: %w", err)
	}

	if conf.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	return nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Exporter: ExporterConfig{
			ExporterType: ExporterTypeStdout,
		},
		Environment: "dev",
	}
}

// ServiceConfig holds configuration for service tracer.
type ServiceConfig struct {
	// service name
	ServiceName string
	SampleRate  float64
}

// Validate validates ServiceConfig.
func (conf *ServiceConfig) Validate() error {
	if conf.SampleRate < 0 || conf.SampleRate > 1 {
		return fmt.Errorf("sample rate must be between 0.0 and 1.0")
	}

	if conf.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	return nil
}

// ExporterConfig holds configuration for exporters.
type ExporterConfig struct {
	ExporterType ExporterType

	// Exporter-specific configurations
	OTLPConfig *OTLPConfig
}

// Validate validates ExporterConfig.
func (conf *ExporterConfig) Validate() error {
	if err := conf.ExporterType.Validate(); err != nil {
		return fmt.Errorf("exporter type is required")
	}

	if conf.ExporterType == ExporterTypeOTLP && conf.OTLPConfig == nil {
		return fmt.Errorf("OTLP conf is required")
	}

	return nil
}

// OTLPConfig holds configuration for OTLP exporter.
type OTLPConfig struct {
	// OTLP endpoint, e.g., "localhost:4317"
	Endpoint string

	// whether to use insecure connection
	Insecure bool

	// additional headers
	Headers map[string]string
}
