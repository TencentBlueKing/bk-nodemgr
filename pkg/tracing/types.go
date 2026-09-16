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

package tracing

import "fmt"

// ExporterType defines the type of exporter to use.
type ExporterType string

// OTLPProtocol defines the transport protocol used by OTLP exporter.
type OTLPProtocol string

const (
	// ExporterTypeStdout exports traces to stdout.
	ExporterTypeStdout ExporterType = "stdout"
	// ExporterTypeOTLP exports traces using OpenTelemetry Protocol.
	ExporterTypeOTLP ExporterType = "otlp"

	// OTLPProtocolGRPC exports OTLP traces over gRPC.
	OTLPProtocolGRPC OTLPProtocol = "grpc"
	// OTLPProtocolHTTP exports OTLP traces over HTTP.
	OTLPProtocolHTTP OTLPProtocol = "http"
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

// Validate validates the OTLP protocol.
func (p OTLPProtocol) Validate() error {
	switch p {
	case "", OTLPProtocolGRPC, OTLPProtocolHTTP:
		return nil
	default:
		return fmt.Errorf("invalid OTLP protocol, protocol(%s)", p)
	}
}
