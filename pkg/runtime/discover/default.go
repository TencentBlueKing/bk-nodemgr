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
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
)

// ProviderDefault this defines the default service discovery provider.
type ProviderDefault struct {
	mutex    sync.RWMutex
	services map[ServiceName]map[string]Instance
}

// NewProviderDefault creates a new default provider.
func NewProviderDefault() *ProviderDefault {
	return &ProviderDefault{
		services: make(map[ServiceName]map[string]Instance),
	}
}

// GetAllService get all specific service instances.
func (p *ProviderDefault) GetAllService(serviceName ServiceName) ([]Instance, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	instanceMap, ok := p.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound()
	}

	instances := conv.MapValueToSlice(instanceMap)

	return instances, nil
}

// GetAllEndpoint get all specific service endpoints.
func (p *ProviderDefault) GetAllEndpoint(serviceName ServiceName, endpointName EndpointName) (
	[]Endpoint, error) {

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	instanceMap, ok := p.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound()
	}

	endpoints := make([]Endpoint, 0, len(instanceMap))

	for _, instance := range instanceMap {
		if endpoint, exists := instance.Endpoints[endpointName]; exists {
			endpoints = append(endpoints, endpoint)
		}
	}

	if len(endpoints) == 0 {
		return nil, ErrEndpointNotFound()
	}

	return endpoints, nil
}

// GetEndpoint get a specific service endpoint.
func (p *ProviderDefault) GetEndpoint(
	serviceName ServiceName, endpointName EndpointName, selector Selector) (Endpoint, error) {

	endpoints, err := p.GetAllEndpoint(serviceName, endpointName)
	if err != nil {
		return Endpoint{}, err
	}

	endpoint, err := selector.Select(endpoints)
	if err != nil {
		return Endpoint{}, err
	}

	return endpoint, nil
}

// Register registers a service instance.
func (p *ProviderDefault) Register(serviceName ServiceName, instances ...Instance) error {
	for _, instance := range instances {
		if err := instance.Validate(); err != nil {
			return err
		}
	}

	p.mutex.Lock()

	if _, exists := p.services[serviceName]; !exists {
		p.services[serviceName] = make(map[string]Instance)
	}

	for _, instance := range instances {
		p.services[serviceName][instance.ID] = instance
	}

	p.mutex.Unlock()

	return nil
}

// Update updates a service instance.
func (p *ProviderDefault) Update(serviceName ServiceName, instance Instance) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.services[serviceName]; !exists {
		return ErrNotRegistered()
	}

	p.services[serviceName][instance.ID] = instance

	return nil
}

// Deregister deregister a service instance.
func (p *ProviderDefault) Deregister(serviceName ServiceName, instanceID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.services[serviceName]; !exists {
		return ErrNotRegistered()
	}

	if _, exists := p.services[serviceName][instanceID]; !exists {
		return ErrNotRegistered()
	}

	delete(p.services[serviceName], instanceID)

	return nil
}

// Start starts the provider.
func (p *ProviderDefault) Start(_ context.Context) error {
	return nil
}

// Stop stops the provider.
func (p *ProviderDefault) Stop() error {
	return nil
}
