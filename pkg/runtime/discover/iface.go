/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package discover ...
package discover

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
)

// Provider this defines the interface of a complete service discovery provider.
type Provider interface {
	// Discover provides discovering interface.
	Discover

	// Registry provides registry interface.
	Registry

	// Start starts the provider.
	Start(ctx context.Context) error

	// Stop stops the provider.
	Stop() error
}

// EndpointName the name of endpoint.
type EndpointName string

const (
	// EndpointNameBackendInfo the backend info endpoint name.
	EndpointNameBackendInfo EndpointName = "backend-info"

	// EndpointNameBackendAdmin the backend admin endpoint name.
	EndpointNameBackendAdmin EndpointName = "backend-admin"

	// EndpointNameBackendBasic the backend endpoint name.
	EndpointNameBackendBasic EndpointName = "backend-basic"

	// EndpointNameBackendCallback the backend callback endpoint name.
	EndpointNameBackendCallback EndpointName = "backend-callback"

	// EndpointNameBackendPorxy the backend proxy endpoint name.
	EndpointNameBackendPorxy EndpointName = "backend-proxy"

	// EndpointNameApplicationInfo the application info endpoint name.
	EndpointNameApplicationInfo EndpointName = "application-info"

	// EndpointNameApplicationAdmin the application admin endpoint name.
	EndpointNameApplicationAdmin EndpointName = "application-admin"

	// EndpointNameApplicationBasic the application basic endpoint name.
	EndpointNameApplicationBasic EndpointName = "application-basic"

	// EndpointNameFileInfo the file info endpoint name.
	EndpointNameFileInfo EndpointName = "file-info"

	// EndpointNameFileAdmin the file admin endpoint name.
	EndpointNameFileAdmin EndpointName = "file-admin"

	// EndpointNameFileBasic the file basic endpoint name.
	EndpointNameFileBasic EndpointName = "file-basic"

	// EndpointNameFileDownload the file download endpoint name.
	EndpointNameFileDownload EndpointName = "file-download"

	// EndpointNameRelayCallback the relay callback endpoint name.
	EndpointNameRelayCallback EndpointName = "relay-callback"

	// EndpointNameRelayFile the relay file endpoint name.
	EndpointNameRelayFile EndpointName = "relay-file"

	// EndpointNameRelayInfo the relay info endpoint name.
	EndpointNameRelayInfo EndpointName = "relay-info"
)

// Endpoint defines the exported endpoint of service on discover.
type Endpoint struct {
	IPV4 string            `json:"ipv4"`
	IPV6 string            `json:"ipv6"`
	Port int               `json:"port"`
	Nice int               `json:"nice"`
	Meta map[string]string `json:"meta"`
}

// Validate validates the endpoint.
func (ep *Endpoint) Validate() error {
	if ep.IPV4 == "" && ep.IPV6 == "" {
		return ErrInvalidServiceIP()
	}
	if ep.Port == 0 {
		return ErrInvalidServiceIP()
	}

	return nil
}

// GetIPV4Address returns the ipv4 address of the service.
func (ep *Endpoint) GetIPV4Address() string {
	return fmt.Sprintf("%s:%d", ep.IPV4, ep.Port)
}

// GetIPV6Address returns the ipv6 address of the service.
func (ep *Endpoint) GetIPV6Address() string {
	return fmt.Sprintf("[%s]:%d", ep.IPV6, ep.Port)
}

// ServiceName the name of service to register.
type ServiceName string

const (
	// ServiceNameBackend the backend service name.
	ServiceNameBackend ServiceName = "backend"
	// ServiceNameApplication the application service name.
	ServiceNameApplication ServiceName = "application"
	// ServiceNameFile the file service name.
	ServiceNameFile ServiceName = "file"
	// ServiceNameRelay the relay service name.
	ServiceNameRelay ServiceName = "relay"
)

// NewInstance creates a new instance.
func NewInstance(name string, meta map[string]string) Instance {
	inst := Instance{
		ID:        identifier.GenServiceID(),
		Name:      name,
		Endpoints: make(map[EndpointName]Endpoint),
		Meta:      meta,
	}

	return inst
}

// Instance this defines a service instance.
type Instance struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	Endpoints map[EndpointName]Endpoint `json:"endpoints"`
	Meta      map[string]string         `json:"meta"`
}

// Validate validates the instance.
func (instance *Instance) Validate() error {
	if instance.ID == "" {
		return ErrInvalidInstanceID()
	}
	if instance.Name == "" {
		return ErrInvalidInstanceName()
	}

	return nil
}

// Update updates the instance.
func (instance *Instance) Update(endpointName EndpointName, endpoint Endpoint) *Instance {
	if instance.Endpoints == nil {
		instance.Endpoints = make(map[EndpointName]Endpoint)
	}

	instance.Endpoints[endpointName] = endpoint

	return instance
}

// Discover this defines the interface of service discovery.
type Discover interface {
	// GetAllService get all specific service instances.
	GetAllService(serviceName ServiceName) ([]Instance, error)

	// GetAllEndpoint get all specific service endpoints.
	GetAllEndpoint(serviceName ServiceName, endpointName EndpointName) ([]Endpoint, error)

	// GetEndpoint get a specific service endpoint.
	GetEndpoint(serviceName ServiceName, endpointName EndpointName, selector Selector) (
		Endpoint, error)
}

// Registry this defines the interface of service registry.
type Registry interface {
	// Register registers a service instance.
	Register(serviceName ServiceName, instance Instance) error

	// Update updates a service instance.
	Update(serviceName ServiceName, instance Instance) error

	// Deregister deregisters a service instance.
	Deregister(serviceName ServiceName, instanceID string) error
}

// Selector this defines the interface of a service endpoint selector.
type Selector interface {
	// Select selects a service endpoint from the given endpoints.
	Select(endpoints []Endpoint) (Endpoint, error)
}
